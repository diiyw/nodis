package main

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/diiyw/nodis"
	"github.com/diiyw/nodis/storage"
)

var CLI struct {
	Addr    string `arg:"" default:":6380" usage:"nodis server address"`
	Storage string `arg:"" default:"memory" usage:"select storage: memory, pebble"`
}

func main() {
	_ = kong.Parse(&CLI)
	opt := nodis.DefaultOptions
	if CLI.Storage == "pebble" {
		opt.Storage = storage.NewPebble("data", nil)
	}
	n := nodis.Open(opt)
	if err := n.Serve(CLI.Addr); err != nil {
		fmt.Printf("Serve() = %v", err)
	}
}
