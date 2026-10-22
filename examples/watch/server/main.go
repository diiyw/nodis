package main

import (
	"fmt"
	"time"

	"github.com/diiyw/nodis"
	"github.com/diiyw/nodis/patch"
)

func main() {
	opt := *nodis.DefaultOptions
	opt.Channel = nodis.NewWebsocket()
	n := nodis.Open(&opt)
	n.WatchKey([]string{"*"}, func(op patch.Op) {
		fmt.Println("Server:", op.Data.GetKey(), op.Type)
	})
	go func() {
		for {
			time.Sleep(time.Second)
			n.Set("test", []byte(time.Now().Format("2006-01-02 15:04:05")), false)
		}
	}()
	err := n.Broadcast("127.0.0.1:6380", []string{"*"})
	if err != nil {
		panic(err)
	}
}
