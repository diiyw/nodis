package main

import (
	"fmt"
	"github.com/diiyw/nodis/fs"
	"net"
	"os"

	"github.com/diiyw/nodis"
)

func main() {
	addr := ":6380"
	if len(os.Args) > 1 {
		addr = os.Args[1]
		ip := net.ParseIP(addr)
		if ip == nil {
			fmt.Printf("invalid ip address: %s", addr)
			os.Exit(0)
		}
	}
	opt := nodis.DefaultOptions
	opt.Filesystem = &fs.Disk{}
	n := nodis.Open(opt)
	if err := n.Serve(addr); err != nil {
		fmt.Printf("Serve() = %v", err)
	}
}
