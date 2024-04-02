package redis

import (
	"log"
	"net"
	"strings"
)

func Serve(addr string, handler func(cmd string, args []Value) Value) error {
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

func handleConn(conn net.Conn, handler func(cmd string, args []Value) Value) {
	defer conn.Close()
	for {
		resp := NewResp(conn)
		value, err := resp.Read()
		if err != nil {
			return
		}

		if value.typ != "array" {
			log.Println("Invalid request, expected array")
			continue
		}

		if len(value.Array) == 0 {
			log.Println("Invalid request, expected array length > 0")
			continue
		}
		command := strings.ToUpper(value.Array[0].Bulk)
		args := value.Array[1:]

		writer := NewWriter(conn)
		result := handler(command, args)
		writer.Write(result)
	}
}
