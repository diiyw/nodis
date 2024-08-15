package nodis

import (
	"errors"
	"syscall/js"

	"github.com/diiyw/nodis/patch"
)

type Websocket struct {
}

func NewWebsocket() *Websocket {
	return &Websocket{}
}

func (ws *Websocket) Publish(addr string, fn func(c ChannelConn)) error {
	return errors.New("Websocket publish not implemented in js")
}

func (ws *Websocket) Subscribe(addr string, fn func(patch.Op)) error {
	jsWs := js.Global().Get("WebSocket").New(addr)
	jsWs.Call("addEventListener", "open", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return nil
	}))
	jsWs.Call("addEventListener", "message", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		blob := args[0].Get("data")
		blob.Call("arrayBuffer").Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
			uint8array := js.Global().Get("Uint8Array").New(args[0])
			var data = make([]byte, uint8array.Get("length").Int())
			n := js.CopyBytesToGo(data, uint8array)
			if n > 0 {
				op, err := patch.DecodeOp(data)
				if err != nil {
					println("Subscribe:", err.Error())
				}
				fn(op)
			}
			return nil
		}))
		return nil
	}))
	return nil
}
