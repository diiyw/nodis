//go:build !js

package nodis

import (
	"net/http"

	"github.com/diiyw/nodis/patch"
	"github.com/gorilla/websocket"
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

func (ws *Websocket) Publish(addr string, fn func(ChannelConn)) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c, err := ws.upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		fn(&WsConn{c: c})
	})
	return http.ListenAndServe(addr, nil)
}

func (ws *Websocket) Subscribe(addr string, fn func(patch.Op)) error {
	c, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		return err
	}
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			return err
		}
		var op patch.Op
		op, err = patch.DecodeOp(message)
		if err != nil {
			return err
		}
		fn(op)
	}
}

type WsConn struct {
	c *websocket.Conn
}

func (w *WsConn) Send(op patch.Op) error {
	err := w.c.WriteMessage(websocket.BinaryMessage, op.Encode())
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
