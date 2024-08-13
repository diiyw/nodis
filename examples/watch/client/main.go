package main

import (
	"fmt"

	"github.com/diiyw/nodis"
	"github.com/diiyw/nodis/patch"
)

func main() {
	var opt = nodis.DefaultOptions
	opt.Synchronizer = nodis.NewWebsocket()
	n := nodis.Open(opt)
	n.WatchKey([]string{"*"}, func(op patch.Op) {
		fmt.Println("Subscribe: ", op.Data.GetKey(), string(op.Data.(*patch.OpSet).Value))
	})
	err := n.Subscribe("ws://127.0.0.1:6380")
	if err != nil {
		panic(err)
	}
	select {}
}
