package main

import (
	"fmt"

	"github.com/diiyw/nodis"
)

func main() {
	opt := nodis.DefaultOptions
	n := nodis.Open(opt)
	if err := n.Serve(":6380"); err != nil {
		fmt.Printf("Serve() = %v", err)
	}
}
