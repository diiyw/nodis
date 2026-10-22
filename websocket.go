//go:build !js

package nodis

import (
	"bytes"
	"github.com/diiyw/nodis/patch"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"sync"
)

// Websocket carries the same NDS1 frames as Stream, with one frame per message.
type Websocket struct {
	upgrader    websocket.Upgrader
	mu          sync.Mutex
	server      *http.Server
	connections map[*websocket.Conn]struct{}
	closed      bool
}

func NewWebsocket() *Websocket {
	return &Websocket{connections: make(map[*websocket.Conn]struct{}), upgrader: websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}}
}
func (ws *Websocket) track(c *websocket.Conn) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	if ws.closed {
		_ = c.Close()
		return ErrClosed
	}
	ws.connections[c] = struct{}{}
	return nil
}
func (ws *Websocket) untrack(c *websocket.Conn) {
	_ = c.Close()
	ws.mu.Lock()
	delete(ws.connections, c)
	ws.mu.Unlock()
}
func (ws *Websocket) Publish(addr string, fn func(ChannelConn)) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c, err := ws.upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		if err = ws.track(c); err != nil {
			return
		}
		defer ws.untrack(c)
		c.SetReadLimit(patch.MaxFrameSize)
		fn(&WsConn{c: c})
	})
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer listener.Close()
	server := &http.Server{Addr: listener.Addr().String(), Handler: mux}
	ws.mu.Lock()
	if ws.closed || ws.server != nil {
		ws.mu.Unlock()
		return ErrClosed
	}
	ws.server = server
	ws.mu.Unlock()
	return server.Serve(listener)
}
func (ws *Websocket) Subscribe(addr string, fn func(patch.Op) error) error {
	c, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		return err
	}
	if err = ws.track(c); err != nil {
		return err
	}
	defer ws.untrack(c)
	c.SetReadLimit(patch.MaxFrameSize)
	var sequence uint64
	for {
		kind, message, err := c.ReadMessage()
		if err != nil {
			return err
		}
		if kind != websocket.BinaryMessage {
			return patch.ErrInvalidFrame
		}
		op, err := patch.DecodeOp(message)
		if err != nil {
			return err
		}
		if op.Sequence != sequence+1 {
			return patch.ErrSequence
		}
		sequence = op.Sequence
		if err = fn(op); err != nil {
			return err
		}
	}
}
func (ws *Websocket) Close() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	if ws.closed {
		return nil
	}
	ws.closed = true
	if ws.server != nil {
		_ = ws.server.Close()
	}
	for c := range ws.connections {
		_ = c.Close()
	}
	return nil
}

type WsConn struct {
	c       *websocket.Conn
	mu      sync.Mutex
	buffer  bytes.Buffer
	encoder *patch.Encoder
}

func (w *WsConn) Send(op patch.Op) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buffer.Reset()
	if w.encoder == nil {
		w.encoder = patch.NewEncoder(&w.buffer)
	}
	if err := w.encoder.Write(op); err != nil {
		return err
	}
	err := w.c.WriteMessage(websocket.BinaryMessage, w.buffer.Bytes())
	if err != nil {
		_ = w.c.Close()
	}
	return err
}
func (w *WsConn) Close() error { return w.c.Close() }
func (w *WsConn) Wait() error {
	for {
		_, _, err := w.c.ReadMessage()
		if err != nil {
			return err
		}
	}
}
