package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"github.com/diiyw/nodis"
)

func main() {
	opt := nodis.DefaultOptions
	n := nodis.Open(opt)
	go func() {
		_ = http.ListenAndServe("0.0.0.0:6060", nil)
	}()
	if err := n.Serve(":6380"); err != nil {
		fmt.Printf("Serve() = %v", err)
	}
}
