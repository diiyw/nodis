// Package server exposes a Nodis database through the native NDC1 byte protocol.
package server

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/diiyw/nodis"
	"github.com/diiyw/nodis/protocol"
)

var ErrAuthentication = errors.New("nodis server: authentication failed")

type Authenticator interface {
	Authenticate(context.Context, []byte) (nodis.Principal, error)
}

type AuthenticateFunc func(context.Context, []byte) (nodis.Principal, error)

func (fn AuthenticateFunc) Authenticate(ctx context.Context, credential []byte) (nodis.Principal, error) {
	return fn(ctx, credential)
}

type StaticTokens map[string]nodis.Principal

func (tokens StaticTokens) Authenticate(_ context.Context, credential []byte) (nodis.Principal, error) {
	principal, ok := tokens[string(credential)]
	if !ok {
		return nodis.Principal{}, ErrAuthentication
	}
	return principal, nil
}

type Server struct {
	DB            *nodis.DB
	Authenticator Authenticator
	mu            sync.Mutex
	listener      net.Listener
	connections   map[net.Conn]struct{}
	closed        bool
}

func New(db *nodis.DB, authenticator Authenticator) *Server {
	return &Server{DB: db, Authenticator: authenticator, connections: make(map[net.Conn]struct{})}
}

func (s *Server) ListenAndServe(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	return s.Serve(listener)
}

func (s *Server) Serve(listener net.Listener) error {
	if s.DB == nil || s.Authenticator == nil {
		_ = listener.Close()
		return ErrAuthentication
	}
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		_ = listener.Close()
		return net.ErrClosed
	}
	s.listener = listener
	s.mu.Unlock()
	for {
		conn, err := listener.Accept()
		if err != nil {
			s.mu.Lock()
			closed := s.closed
			s.mu.Unlock()
			if closed {
				return nil
			}
			return err
		}
		s.mu.Lock()
		if s.closed {
			s.mu.Unlock()
			_ = conn.Close()
			return nil
		}
		s.connections[conn] = struct{}{}
		s.mu.Unlock()
		go s.handle(conn)
	}
}

func (s *Server) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	var err error
	if s.listener != nil {
		err = s.listener.Close()
	}
	for conn := range s.connections {
		_ = conn.Close()
	}
	s.mu.Unlock()
	return err
}

type serverWire struct {
	conn          net.Conn
	reader        protocol.CommandReader[net.Conn]
	writer        protocol.CommandWriter[net.Conn]
	queryDecoder  protocol.QueryDecoder
	queryBuffer   nodis.QueryBuffer
	queryResponse []byte
}

func (w *serverWire) respond(request protocol.Command, status protocol.Status, payload []byte) error {
	return w.writer.Write(protocol.Command{RequestID: request.RequestID, Op: request.Op, Status: status, Payload: payload})
}

func (s *Server) handle(conn net.Conn) {
	wire := serverWire{conn: conn, reader: protocol.CommandReader[net.Conn]{Reader: conn}, writer: protocol.CommandWriter[net.Conn]{Writer: conn}}

	defer func() { _ = conn.Close(); s.mu.Lock(); delete(s.connections, conn); s.mu.Unlock() }()
	// Handshake and authentication must not leave idle sockets open forever.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
	if encrypted, ok := conn.(*tls.Conn); ok {
		if err := encrypted.HandshakeContext(ctx); err != nil {
			return
		}
	}
	first, err := wire.reader.Read()
	if err != nil {
		return
	}
	if first.Op != protocol.OpAuthenticate {
		_ = wire.respond(first, protocol.StatusUnauthorized, nil)
		return
	}
	principal, err := s.Authenticator.Authenticate(ctx, first.Payload)
	if err != nil || principal.ID == "" {
		_ = wire.respond(first, protocol.StatusUnauthorized, nil)
		return
	}
	if err := wire.respond(first, protocol.StatusOK, nil); err != nil {
		return
	}
	cancel()
	_ = conn.SetDeadline(time.Time{})
	for {
		command, err := wire.reader.Read()
		if err != nil {
			return
		}
		if command.Status != protocol.StatusOK {
			return
		}
		if command.Op == protocol.OpSubscribe {
			s.subscribe(&wire, principal, command)
			return
		}
		var status protocol.Status
		var payload []byte
		if command.Op == protocol.OpFind {
			status, payload = wire.find(s.DB, principal, command.Payload)
		} else {
			status, payload = s.execute(principal, command)
		}
		if err := wire.respond(command, status, payload); err != nil {
			return
		}
	}
}

func (s *Server) execute(principal nodis.Principal, command protocol.Command) (protocol.Status, []byte) {
	switch command.Op {
	case protocol.OpGetDocument:
		name := string(command.Payload)
		doc, err := s.DB.Document(principal, name)
		if err != nil {
			return errorStatus(err), nil
		}
		payload := binary.LittleEndian.AppendUint64(nil, uint64(doc.ID()))
		payload = append(payload, doc.Name()...)
		return protocol.StatusOK, payload
	case protocol.OpCreateDocument:
		doc, err := s.DB.CreateDocument(principal, string(command.Payload))
		if err != nil {
			return errorStatus(err), nil
		}
		return protocol.StatusOK, binary.LittleEndian.AppendUint64(nil, uint64(doc.ID()))
	case protocol.OpRenameDocument:
		if len(command.Payload) < 9 {
			return protocol.StatusInvalid, nil
		}
		id := nodis.DocumentID(binary.LittleEndian.Uint64(command.Payload))
		if err := s.DB.RenameDocument(principal, id, string(command.Payload[8:])); err != nil {
			return errorStatus(err), nil
		}
		return protocol.StatusOK, nil
	case protocol.OpDeleteDocument:
		if len(command.Payload) != 8 {
			return protocol.StatusInvalid, nil
		}
		if err := s.DB.DeleteDocument(principal, nodis.DocumentID(binary.LittleEndian.Uint64(command.Payload))); err != nil {
			return errorStatus(err), nil
		}
		return protocol.StatusOK, nil
	case protocol.OpGet:
		if len(command.Payload) != 16 {
			return protocol.StatusInvalid, nil
		}
		view, err := s.DB.Get(principal, nodis.DocumentID(binary.LittleEndian.Uint64(command.Payload)), nodis.ObjectID(binary.LittleEndian.Uint64(command.Payload[8:])))
		if err != nil {
			return errorStatus(err), nil
		}
		payload, err := view.Portable()
		if err != nil {
			return errorStatus(err), nil
		}
		return protocol.StatusOK, payload
	case protocol.OpCreate:
		document, _, fields, err := decodeMutation(command.Payload, false)
		if err != nil {
			return protocol.StatusInvalid, nil
		}
		tx := s.DB.Begin(principal)
		draft, err := tx.Create(document)
		if err == nil {
			err = setFields(draft, fields)
		}
		var id nodis.ObjectID
		if err == nil {
			id, err = draft.Save()
		}
		if err == nil {
			err = tx.Commit()
		} else {
			tx.Rollback()
		}
		if err != nil {
			return errorStatus(err), nil
		}
		return protocol.StatusOK, binary.LittleEndian.AppendUint64(nil, uint64(id))
	case protocol.OpUpdate:
		document, object, fields, err := decodeMutation(command.Payload, true)
		if err != nil {
			return protocol.StatusInvalid, nil
		}
		revision := binary.LittleEndian.Uint64(command.Payload[16:24])
		tx := s.DB.Begin(principal)
		draft, err := tx.Update(document, object, revision)
		if err == nil {
			err = setFields(draft, fields)
		}
		if err == nil {
			_, err = draft.Save()
		}
		if err == nil {
			err = tx.Commit()
		} else {
			tx.Rollback()
		}
		if err != nil {
			return errorStatus(err), nil
		}
		return protocol.StatusOK, nil
	case protocol.OpDelete:
		if len(command.Payload) != 24 {
			return protocol.StatusInvalid, nil
		}
		document := nodis.DocumentID(binary.LittleEndian.Uint64(command.Payload))
		object := nodis.ObjectID(binary.LittleEndian.Uint64(command.Payload[8:]))
		revision := binary.LittleEndian.Uint64(command.Payload[16:])
		tx := s.DB.Begin(principal)
		err := tx.Delete(document, object, revision)
		if err == nil {
			err = tx.Commit()
		} else {
			tx.Rollback()
		}
		if err != nil {
			return errorStatus(err), nil
		}
		return protocol.StatusOK, nil
	default:
		return protocol.StatusInvalid, nil
	}
}

type namedValue struct {
	name  string
	value nodis.Value
}

func decodeMutation(payload []byte, update bool) (nodis.DocumentID, nodis.ObjectID, []namedValue, error) {
	header := 8
	if update {
		header = 24
	}
	if len(payload) < header {
		return 0, 0, nil, protocol.ErrInvalidCommand
	}
	document := nodis.DocumentID(binary.LittleEndian.Uint64(payload))
	var object nodis.ObjectID
	if update {
		object = nodis.ObjectID(binary.LittleEndian.Uint64(payload[8:]))
	}
	count, width := binary.Uvarint(payload[header:])
	if width <= 0 || count > uint64(len(payload)-header-width)/3 {
		return 0, 0, nil, protocol.ErrInvalidCommand
	}
	pos := header + width
	fields := make([]namedValue, 0, int(count))
	for range count {
		length, n := binary.Uvarint(payload[pos:])
		if n <= 0 || length > uint64(len(payload)-pos-n) {
			return 0, 0, nil, protocol.ErrInvalidCommand
		}
		pos += n
		name := string(payload[pos : pos+int(length)])
		pos += int(length)
		value, used, err := nodis.DecodeValue(payload[pos:])
		if err != nil {
			return 0, 0, nil, err
		}
		pos += used
		fields = append(fields, namedValue{name: name, value: value})
	}
	if pos != len(payload) {
		return 0, 0, nil, protocol.ErrInvalidCommand
	}
	return document, object, fields, nil
}

func setFields(draft *nodis.Draft, fields []namedValue) error {
	for i := range fields {
		if err := draft.Set(fields[i].name, fields[i].value); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) subscribe(wire *serverWire, principal nodis.Principal, command protocol.Command) {
	conn := wire.conn
	count, width := binary.Uvarint(command.Payload)
	if width <= 0 || (len(command.Payload)-width)%8 != 0 || count != uint64((len(command.Payload)-width)/8) {
		_ = wire.respond(command, protocol.StatusInvalid, nil)
		return
	}
	scope := nodis.SyncScope{Documents: make([]nodis.DocumentID, int(count))}
	for i := range scope.Documents {
		scope.Documents[i] = nodis.DocumentID(binary.LittleEndian.Uint64(command.Payload[width+i*8:]))
	}
	sub, err := s.DB.Subscribe(principal, scope)
	if err != nil {
		_ = wire.respond(command, errorStatus(err), nil)
		return
	}
	defer sub.Close()
	// A subscription connection only receives events after the initial request.
	// Detect peer shutdown even when the database is idle, releasing the listener.
	go func() {
		var unexpected [1]byte
		_, _ = conn.Read(unexpected[:])
		sub.Close()
	}()
	if err := wire.respond(command, protocol.StatusOK, nil); err != nil {
		return
	}
	var encoded []byte
	for event := range sub.C {
		encoded = protocol.AppendEvent(encoded[:0], event)
		if err := wire.writer.Write(protocol.Command{RequestID: command.RequestID, Op: protocol.OpEvent, Payload: encoded}); err != nil {
			return
		}
	}
}

func errorStatus(err error) protocol.Status {
	switch {
	case errors.Is(err, nodis.ErrPermission):
		return protocol.StatusForbidden
	case errors.Is(err, nodis.ErrNotFound):
		return protocol.StatusNotFound
	case errors.Is(err, nodis.ErrConflict):
		return protocol.StatusConflict
	case errors.Is(err, nodis.ErrQueryTooLarge):
		return protocol.StatusTooLarge
	case errors.Is(err, nodis.ErrInvalidName), errors.Is(err, nodis.ErrInvalidValue), errors.Is(err, nodis.ErrInvalidState), errors.Is(err, nodis.ErrInvalidQuery):
		return protocol.StatusInvalid
	default:
		return protocol.StatusInternal
	}
}
