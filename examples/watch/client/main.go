package main

import (
	"fmt"

	"github.com/diiyw/nodis"
	"github.com/diiyw/nodis/patch"
)

func main() {
	opt := *nodis.DefaultOptions
	opt.Channel = nodis.NewWebsocket()
	n := nodis.Open(&opt)
	n.WatchKey([]string{"*"}, func(op patch.Op) {
		fmt.Println("Subscribe: ", op.Data.GetKey(), op.Type)
	})
	err := n.Subscribe("ws://127.0.0.1:6380")
	if err != nil {
		panic(err)
	}
	select {}
}
