package redis

import (
	"net"

	"github.com/tidwall/btree"
)

const (
	MultiNone    uint8 = 0
	MultiPrepare uint8 = 1
	MultiCommit  uint8 = 2
	MultiError   uint8 = 4
)

type HandlerFunc func(c *Conn, cmd Command)

type Conn struct {
	*Reader
	*Writer
	Network   net.Conn
	Commands  []func()
	State     uint8
	WatchKeys btree.Map[string, bool]
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
	c := &Conn{
		Reader:   NewReader(conn),
		Writer:   NewWriter(conn),
		Network:  conn,
		Commands: make([]func(), 0),
	}
	for {
		err := c.Reader.ReadCommand()
		if err != nil {
			c.WriteError(err.Error())
			_ = c.Flush()
			break
		}
		handler(c, c.cmd)
		_ = c.Flush()
	}
}
