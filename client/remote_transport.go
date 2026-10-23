package client

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

// ErrOutcomeUnknown means a mutation may have committed before the connection
// failed. Reconnects never replay a request automatically.
var ErrOutcomeUnknown = errors.New("nodis client: server execution outcome unknown")

type OutcomeUnknownError struct {
	Op  protocol.Op
	Err error
}

func (e *OutcomeUnknownError) Error() string        { return ErrOutcomeUnknown.Error() + ": " + e.Err.Error() }
func (e *OutcomeUnknownError) Unwrap() error        { return e.Err }
func (e *OutcomeUnknownError) Is(target error) bool { return target == ErrOutcomeUnknown }

type remoteWire struct {
	conn   net.Conn
	reader protocol.CommandReader[net.Conn]
	writer protocol.CommandWriter[net.Conn]
}

func newRemoteWire(conn net.Conn) *remoteWire {
	return &remoteWire{conn: conn, reader: protocol.CommandReader[net.Conn]{Reader: conn}, writer: protocol.CommandWriter[net.Conn]{Writer: conn}}
}

func (w *remoteWire) close() {
	// Interrupt pending TLS writes as well as ordinary socket reads.
	_ = w.conn.SetDeadline(time.Now())
	_ = w.conn.Close()
}

type Remote struct {
	address     string
	credential  []byte
	tlsConfig   *tls.Config
	scope       nodis.SyncScope
	ctx         context.Context
	cancel      context.CancelFunc
	requestGate chan struct{}
	connMu      sync.Mutex
	conn        *remoteWire
	syncConn    *remoteWire
	nextID      uint64
	cache       viewCache
	closeOnce   sync.Once
	done        chan struct{}
}

type DialOptions struct {
	Scope     nodis.SyncScope
	TLSConfig *tls.Config
	// DisableSync uses only authoritative requests and leaves Cached empty.
	DisableSync bool
}

func Dial(ctx context.Context, address string, credential []byte, scope nodis.SyncScope) (*Remote, error) {
	return DialWithOptions(ctx, address, credential, DialOptions{Scope: scope})
}

func DialTLS(ctx context.Context, address string, credential []byte, scope nodis.SyncScope, config *tls.Config) (*Remote, error) {
	if config == nil {
		return nil, errors.New("nodis client: TLS config is required")
	}
	return DialWithOptions(ctx, address, credential, DialOptions{Scope: scope, TLSConfig: config})
}

func DialWithOptions(ctx context.Context, address string, credential []byte, options DialOptions) (*Remote, error) {
	life, cancel := context.WithCancel(context.Background())
	r := &Remote{
		address: address, credential: append([]byte(nil), credential...),
		scope: nodis.SyncScope{Documents: append([]nodis.DocumentID(nil), options.Scope.Documents...)},
		ctx:   life, cancel: cancel, requestGate: make(chan struct{}, 1), done: make(chan struct{}),
	}
	r.requestGate <- struct{}{}
	r.cache.init(!options.DisableSync)
	if options.TLSConfig != nil {
		r.tlsConfig = options.TLSConfig.Clone()
		if r.tlsConfig.MinVersion < tls.VersionTLS13 {
			r.tlsConfig.MinVersion = tls.VersionTLS13
		}
	}
	conn, err := r.dialAuthenticated(ctx)
	if err != nil {
		cancel()
		return nil, err
	}
	r.conn = conn
	if options.DisableSync {
		close(r.done)
		return r, nil
	}
	syncConn, err := r.dialAuthenticated(ctx)
	if err == nil {
		err = r.startSubscription(ctx, syncConn)
	}
	if err != nil {
		cancel()
		conn.close()
		if syncConn != nil {
			syncConn.close()
		}
		return nil, err
	}
	r.syncConn = syncConn
	go r.syncLoop(syncConn)
	return r, nil
}

func (r *Remote) dialAuthenticated(ctx context.Context) (*remoteWire, error) {
	dialCtx, cancel := context.WithCancel(ctx)
	stopLife := context.AfterFunc(r.ctx, cancel)
	defer stopLife()
	defer cancel()
	var conn net.Conn
	var err error
	if r.tlsConfig != nil {
		conn, err = (&tls.Dialer{Config: r.tlsConfig}).DialContext(dialCtx, "tcp", r.address)
	} else {
		conn, err = (&net.Dialer{}).DialContext(dialCtx, "tcp", r.address)
	}
	if err != nil {
		return nil, err
	}
	wire := newRemoteWire(conn)
	stop := context.AfterFunc(dialCtx, wire.close)
	defer stop()
	command := protocol.Command{RequestID: 1, Op: protocol.OpAuthenticate, Payload: r.credential}
	err = wire.writer.Write(command)
	if err == nil {
		var response protocol.Command
		response, err = wire.reader.Read()
		if err == nil {
			err = responseError(command, response)
		}
	}
	if dialCtx.Err() != nil {
		err = dialCtx.Err()
	}
	if err != nil {
		wire.close()
		return nil, err
	}
	return wire, nil
}

func responseError(request, response protocol.Command) error {
	if request.RequestID != response.RequestID || request.Op != response.Op || response.Status > protocol.StatusTooLarge {
		return protocol.ErrInvalidCommand
	}
	if response.Status != protocol.StatusOK {
		return statusError(response.Status)
	}
	return nil
}

func (r *Remote) startSubscription(ctx context.Context, wire *remoteWire) error {
	stop := context.AfterFunc(ctx, wire.close)
	defer stop()
	payload := binary.AppendUvarint(nil, uint64(len(r.scope.Documents)))
	for _, document := range r.scope.Documents {
		payload = binary.LittleEndian.AppendUint64(payload, uint64(document))
	}
	request := protocol.Command{RequestID: 2, Op: protocol.OpSubscribe, Payload: payload}
	if err := wire.writer.Write(request); err != nil {
		return err
	}
	response, err := wire.reader.Read()
	if err != nil {
		return err
	}
	if err := responseError(request, response); err != nil {
		return err
	}
	for {
		event, err := readSyncEvent(wire)
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := r.cache.stage(event); err != nil {
			return err
		}
		if event.Type == nodis.EventSnapshotDone {
			return nil
		}
	}
}

func readSyncEvent(wire *remoteWire) (nodis.Event, error) {
	message, err := wire.reader.Read()
	if err != nil {
		return nodis.Event{}, err
	}
	if message.RequestID != 2 || message.Op != protocol.OpEvent || message.Status != protocol.StatusOK {
		return nodis.Event{}, protocol.ErrInvalidCommand
	}
	return protocol.DecodeEvent(message.Payload)
}

func (r *Remote) syncLoop(wire *remoteWire) {
	defer close(r.done)
	defer func() { r.dropSync(wire) }()
	for {
		event, err := readSyncEvent(wire)
		if err == nil && event.Type != nodis.EventReset {
			err = r.cache.apply(event)
		}
		if err == nil && event.Type != nodis.EventReset {
			continue
		}
		r.cache.invalidate(err)
		r.dropSync(wire)
		var ok bool
		wire, ok = r.reconnectSync()
		if !ok {
			return
		}
	}
}

func (r *Remote) dropSync(wire *remoteWire) {
	if wire == nil {
		return
	}
	r.connMu.Lock()
	if r.syncConn == wire {
		r.syncConn = nil
	}
	r.connMu.Unlock()
	wire.close()
}

func (r *Remote) reconnectSync() (*remoteWire, bool) {
	delay := 100 * time.Millisecond
	for {
		timer := time.NewTimer(delay)
		select {
		case <-r.ctx.Done():
			timer.Stop()
			return nil, false
		case <-timer.C:
		}
		ctx, cancel := context.WithTimeout(r.ctx, 5*time.Second)
		wire, err := r.dialAuthenticated(ctx)
		if err == nil {
			r.connMu.Lock()
			if r.ctx.Err() == nil {
				r.syncConn = wire
			} else {
				err = nodis.ErrClosed
			}
			r.connMu.Unlock()
			if err == nil {
				err = r.startSubscription(ctx, wire)
			}
		}
		cancel()
		if err == nil {
			return wire, true
		}
		r.cache.invalidate(err)
		r.dropSync(wire)
		if delay < 5*time.Second {
			delay = min(delay*2, 5*time.Second)
		}
	}
}

func (r *Remote) Cached(document nodis.DocumentID, object nodis.ObjectID) (*nodis.PortableObject, bool) {
	return r.cache.cached(document, object)
}

func (r *Remote) SyncStatus() SyncStatus { return r.cache.syncStatus() }

func (r *Remote) commandConnection(ctx context.Context) (*remoteWire, error) {
	r.connMu.Lock()
	wire := r.conn
	r.connMu.Unlock()
	if wire != nil {
		return wire, nil
	}
	wire, err := r.dialAuthenticated(ctx)
	if err != nil {
		return nil, err
	}
	r.connMu.Lock()
	defer r.connMu.Unlock()
	if r.ctx.Err() != nil {
		wire.close()
		return nil, nodis.ErrClosed
	}
	r.conn = wire
	return wire, nil
}

func (r *Remote) dropCommand(wire *remoteWire) {
	r.connMu.Lock()
	if r.conn == wire {
		r.conn = nil
	}
	r.connMu.Unlock()
	wire.close()
}

func (r *Remote) request(ctx context.Context, op protocol.Op, payload []byte) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-r.ctx.Done():
		return nil, nodis.ErrClosed
	case <-r.requestGate:
	}
	defer func() { r.requestGate <- struct{}{} }()
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if r.ctx.Err() != nil {
		return nil, nodis.ErrClosed
	}
	if len(payload) > protocol.MaxFrameSize-protocol.CommandHeaderSize-4 {
		return nil, protocol.ErrInvalidCommand
	}
	wire, err := r.commandConnection(ctx)
	if err != nil {
		return nil, r.requestError(ctx, op, false, err)
	}
	var stop func() bool
	if ctx.Done() != nil {
		stop = context.AfterFunc(ctx, wire.close)
	}
	defer func() {
		if stop != nil {
			stop()
		}
		if ctx.Err() != nil {
			r.dropCommand(wire)
		}
	}()
	r.nextID++
	request := protocol.Command{RequestID: r.nextID + 1, Op: op, Payload: payload}
	if err := wire.writer.Write(request); err != nil {
		r.dropCommand(wire)
		return nil, r.requestError(ctx, op, true, err)
	}
	response, err := wire.reader.Read()
	if err != nil {
		r.dropCommand(wire)
		return nil, r.requestError(ctx, op, true, err)
	}
	if response.RequestID != request.RequestID || response.Op != op || response.Status > protocol.StatusTooLarge {
		r.dropCommand(wire)
		return nil, r.requestError(ctx, op, true, protocol.ErrInvalidCommand)
	}
	if response.Status != protocol.StatusOK {
		return nil, statusError(response.Status)
	}
	return append([]byte(nil), response.Payload...), nil
}

func (r *Remote) requestError(ctx context.Context, op protocol.Op, sent bool, err error) error {
	if ctx.Err() != nil {
		err = ctx.Err()
	} else if r.ctx.Err() != nil {
		err = nodis.ErrClosed
	}
	if sent {
		switch op {
		case protocol.OpCreate, protocol.OpUpdate, protocol.OpDelete, protocol.OpCreateDocument, protocol.OpRenameDocument, protocol.OpDeleteDocument:
			return &OutcomeUnknownError{Op: op, Err: err}
		}
	}
	return err
}

func (r *Remote) Close() {
	r.closeOnce.Do(func() {
		r.cancel()
		r.connMu.Lock()
		command, subscription := r.conn, r.syncConn
		r.conn, r.syncConn = nil, nil
		r.connMu.Unlock()
		if command != nil {
			command.close()
		}
		if subscription != nil {
			subscription.close()
		}
		r.cache.close()
		<-r.done
	})
}
