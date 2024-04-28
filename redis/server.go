package redis

import (
	"net"
)

type HandlerFunc func(w *Writer, cmd *Command)

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
	reader, writer := NewReader(conn), NewWriter(conn)
	defer func() {
		// if r := recover(); r != nil {
		// 	writer.WriteError(r.(error).Error())
		// 	writer.Flush()
		// }
		conn.Close()
	}()
	for {
		err := reader.ReadCommand()
		if err != nil {
			writer.WriteError(err.Error())
			_ = writer.Flush()
			break
		}
		handler(writer, reader.cmd)
		_ = writer.Flush()
	}
}
