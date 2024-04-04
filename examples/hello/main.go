package main

import (
	"fmt"

	"github.com/diiyw/nodis"
)

func main() {
	// Create a new Nodis instance
	opt := nodis.DefaultOptions
	n := nodis.Open(opt)
	defer n.Close()
	// Set a key-value pair
	n.Set("echo", []byte("hello world"))
	fmt.Println("Set key-value pair: ", string(n.Get("echo")))
}
