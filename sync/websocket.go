//go:build !js

package sync

import (
	"net/http"

	"github.com/diiyw/nodis/pb"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

type Websocket struct {
	upgrader websocket.Upgrader
}

func NewWebsocket() *Websocket {
	return &Websocket{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (ws *Websocket) Publish(addr string, fn func(Conn)) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c, err := ws.upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		fn(&WsConn{c: c})
	})
	return http.ListenAndServe(addr, nil)
}

func (ws *Websocket) Subscribe(addr string, fn func(*pb.Op)) error {
	c, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		return err
	}
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			return err
		}
		var op = &pb.Op{Operation: &pb.Operation{}}
		err = proto.Unmarshal(message, op.Operation)
		if err != nil {
			return err
		}
		fn(op)
	}
}

type WsConn struct {
	c *websocket.Conn
}

func (w *WsConn) Send(op *pb.Op) error {
	p, err := proto.Marshal(op)
	if err != nil {
		return err
	}
	err = w.c.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return err
	}
	return nil
}

func (w *WsConn) Wait() error {
	for {
		_, _, err := w.c.ReadMessage()
		if err != nil {
			return err
		}
	}
}
