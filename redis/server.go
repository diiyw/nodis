package redis

import (
	"io"
	"net"
	"sync"
)

const (
	MultiNone    uint8 = 0
	MultiPrepare uint8 = 1
	MultiCommit  uint8 = 2
	MultiError   uint8 = 4
)

var (
	ClientLocker sync.RWMutex
	Clients      = make(map[int]*Conn)
)

type HandlerFunc func(c *Conn, cmd Command)

type Conn struct {
	Fd   int
	Name string
	*Reader
	*Writer
	Client    net.Conn
	Commands  []func()
	State     uint8
	WatchKeys map[string]bool
}

func Serve(addr string, handler HandlerFunc) error {
	// Create a new server
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	for {
		// Listen for connections
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go handleConn(conn, handler)
	}
}

func handleConn(conn net.Conn, handler HandlerFunc) {
	ClientLocker.Lock()
	c := &Conn{
		Fd:        len(Clients) + 1,
		Reader:    NewReader(conn),
		Writer:    NewWriter(conn),
		Client:    conn,
		Commands:  make([]func(), 0),
		WatchKeys: make(map[string]bool),
	}
	Clients[c.Fd] = c
	ClientLocker.Unlock()
	for {
		err := c.Reader.ReadCommand()
		if err != nil {
			c.Writer.Reset()
			if _, ok := err.(*net.OpError); ok || err == io.EOF {
				break
			}
			c.WriteError(err.Error())
			_ = c.Push()
			break
		}
		handler(c, c.cmd)
		_ = c.Push()
	}
	_ = conn.Close()
	ClientLocker.Lock()
	delete(Clients, c.Fd)
	ClientLocker.Unlock()
}
