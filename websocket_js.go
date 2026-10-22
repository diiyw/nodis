package nodis

import (
	"errors"
	"github.com/diiyw/nodis/patch"
	"syscall/js"
)

type Websocket struct {
	connection js.Value
	message    js.Func
	closed     bool
}

func NewWebsocket() *Websocket { return &Websocket{} }
func (ws *Websocket) Publish(string, func(ChannelConn)) error {
	return errors.New("nodis: WebSocket publishing is unavailable in browsers")
}
func (ws *Websocket) Subscribe(addr string, fn func(patch.Op) error) error {
	if ws.connection.Truthy() {
		return errors.New("nodis: WebSocket already connected")
	}
	connection := js.Global().Get("WebSocket").New(addr)
	connection.Set("binaryType", "arraybuffer")
	ws.connection = connection
	var sequence uint64
	ws.message = js.FuncOf(func(_ js.Value, args []js.Value) any {
		if ws.closed {
			return nil
		}
		array := js.Global().Get("Uint8Array").New(args[0].Get("data"))
		data := make([]byte, array.Get("byteLength").Int())
		js.CopyBytesToGo(data, array)
		op, err := patch.DecodeOp(data)
		if err == nil && op.Sequence != sequence+1 {
			err = patch.ErrSequence
		}
		if err == nil {
			sequence = op.Sequence
			err = fn(op)
		}
		if err != nil {
			println("nodis subscribe:", err.Error())
			_ = ws.Close()
		}
		return nil
	})
	connection.Call("addEventListener", "message", ws.message)
	return nil
}
func (ws *Websocket) Close() error {
	if ws.closed {
		return nil
	}
	ws.closed = true
	if ws.connection.Truthy() {
		ws.connection.Call("removeEventListener", "message", ws.message)
		ws.connection.Call("close")
		ws.message.Release()
	}
	return nil
}
