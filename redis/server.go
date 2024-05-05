package redis

import (
	"net"
)

type HandlerFunc func(c *Conn, cmd *Command)

type Conn struct {
	*Reader
	*Writer
	Network  net.Conn
	Multi    bool
	Commands []*Command
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
		Commands: make([]*Command, 0),
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
