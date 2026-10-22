package nodis

import (
	"errors"
	"github.com/diiyw/nodis/patch"
	"net"
	"sync"
)

// Stream transports NDS1 frames over TCP. Reconnection starts a full sync.
type Stream struct {
	mu          sync.Mutex
	listener    net.Listener
	connections map[net.Conn]struct{}
	closed      bool
}

func NewStream() *Stream { return &Stream{connections: make(map[net.Conn]struct{})} }
func (s *Stream) track(c net.Conn) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		_ = c.Close()
		return ErrClosed
	}
	s.connections[c] = struct{}{}
	return nil
}
func (s *Stream) untrack(c net.Conn) {
	_ = c.Close()
	s.mu.Lock()
	delete(s.connections, c)
	s.mu.Unlock()
}
func (s *Stream) Publish(addr string, fn func(ChannelConn)) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer l.Close()
	s.mu.Lock()
	if s.closed || s.listener != nil {
		s.mu.Unlock()
		return errors.New("nodis: stream already listening or closed")
	}
	s.listener = l
	s.mu.Unlock()
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		if err = s.track(conn); err != nil {
			return err
		}
		go func() { defer s.untrack(conn); fn(&streamConn{Conn: conn, encoder: patch.NewEncoder(conn)}) }()
	}
}
func (s *Stream) Subscribe(addr string, fn func(patch.Op) error) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	if err = s.track(conn); err != nil {
		return err
	}
	defer s.untrack(conn)
	decoder := patch.NewDecoder(conn)
	for {
		op, err := decoder.Read()
		if err != nil {
			return err
		}
		if err = fn(op); err != nil {
			return err
		}
	}
}
func (s *Stream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	if s.listener != nil {
		_ = s.listener.Close()
	}
	for conn := range s.connections {
		_ = conn.Close()
	}
	return nil
}

type streamConn struct {
	net.Conn
	mu      sync.Mutex
	encoder *patch.Encoder
}

func (c *streamConn) Send(op patch.Op) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.encoder.Write(op); err != nil {
		_ = c.Conn.Close()
		return err
	}
	return nil
}
func (c *streamConn) Wait() error { var b [1]byte; _, err := c.Conn.Read(b[:]); return err }
