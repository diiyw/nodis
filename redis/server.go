package redis

import (
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
)

const (
	MultiNone    uint8 = 0
	MultiPrepare uint8 = 1
	MultiCommit  uint8 = 2
	MultiError   uint8 = 4
)

var nextClientID atomic.Int64
var (
	ClientLocker sync.RWMutex
	Clients      = make(map[int]*Conn)
)

type HandlerFunc func(*Conn, Command)
type Conn struct {
	Fd   int
	Name string
	*Reader
	*Writer
	Client    net.Conn
	Commands  []Command
	Current   Command
	OnClose   func()
	State     uint8
	WatchKeys map[string]bool
}
type Server struct {
	mu          sync.Mutex
	listener    net.Listener
	connections map[net.Conn]struct{}
	closed      bool
	handler     HandlerFunc
}

func NewServer(handler HandlerFunc) *Server {
	return &Server{connections: make(map[net.Conn]struct{}), handler: handler}
}
func Serve(addr string, handler HandlerFunc) error { return NewServer(handler).Serve(addr) }
func (s *Server) Serve(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer l.Close()
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return net.ErrClosed
	}
	s.listener = l
	s.mu.Unlock()
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		s.mu.Lock()
		if s.closed {
			s.mu.Unlock()
			_ = conn.Close()
			return net.ErrClosed
		}
		s.connections[conn] = struct{}{}
		s.mu.Unlock()
		go func() {
			defer func() { s.mu.Lock(); delete(s.connections, conn); s.mu.Unlock() }()
			handleConn(conn, s.handler)
		}()
	}
}
func (s *Server) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	if s.listener != nil {
		_ = s.listener.Close()
	}
	for c := range s.connections {
		_ = c.Close()
	}
	return nil
}
func handleConn(conn net.Conn, handler HandlerFunc) {
	c := &Conn{Fd: int(nextClientID.Add(1)), Reader: NewReader(conn), Writer: NewWriter(conn), Client: conn, WatchKeys: make(map[string]bool)}
	ClientLocker.Lock()
	Clients[c.Fd] = c
	ClientLocker.Unlock()
	defer func() {
		if p := recover(); p != nil {
			c.WriteError("ERR " + fmt.Sprint(p))
			_ = c.Push()
		}
		_ = conn.Close()
		if c.OnClose != nil {
			c.OnClose()
		}
		ClientLocker.Lock()
		delete(Clients, c.Fd)
		ClientLocker.Unlock()
	}()
	for {
		if err := c.ReadCommand(); err != nil {
			if err != io.EOF {
				c.WriteError(err.Error())
				_ = c.Push()
			}
			return
		}
		handler(c, c.cmd)
		if err := c.Push(); err != nil {
			return
		}
	}
}
