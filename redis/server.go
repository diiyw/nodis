package redis

import (
	"log"
	"net"
)

func Serve(addr string, handler func(cmd Value, args []Value) Value) error {
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

func handleConn(conn net.Conn, handler func(cmd Value, args []Value) Value) {
	writer := NewWriter(conn)
	defer func() {
		// if r := recover(); r != nil {
		// 	_ = writer.Write(ErrorValue(r.(error).Error()))
		// }
		conn.Close()
	}()
	resp := NewResp(conn)
	for {
		value, err := resp.Read()
		if err != nil {
			return
		}

		if value.typ != ArrayType {
			log.Println("Invalid request, expected array")
			continue
		}

		if len(value.Array) == 0 {
			log.Println("Invalid request, expected array length > 0")
			continue
		}
		cmd := value.Array[0]
		cmd.Options = value.Options
		cmd.Args = value.Args
		args := value.Array[1:]

		result := handler(cmd, args)
		_ = writer.Write(result)
	}
}
