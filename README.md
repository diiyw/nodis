# Nodis
![GitHub top language](https://img.shields.io/github/languages/top/diiyw/nodis) ![GitHub Release](https://img.shields.io/github/v/release/diiyw/nodis)
<div class="column" align="left">
  <a href="https://godoc.org/github.com/diiyw/nodis"><img src="https://godoc.org/github.com/diiyw/nodis?status.svg" /></a>
  <a href="https://goreportcard.com/report/github.com/diiyw/nodis"><img src="https://goreportcard.com/badge/github.com/diiyw/nodis" /></a>
  <a href="https://codecov.io/gh/diiyw/nodis"><img src="https://codecov.io/gh/diiyw/nodis/branch/main/graph/badge.svg?token=CupujOXpbe"/></a>
</div>


English | [简体中文](https://github.com/diiyw/nodis/blob/main/README_zh-cn.md)

A Golang implemented Redis data structure. 
It is a simple and easy to embed in your application.

## Supported Data Types

- Bitmap
- String
- List
- Hash
- Set
- Sorted Set

## Features

- Fast and embeddable
- Low memory used, only hot data stored in memory
- Snapshot and WAL for data storage.
- Support custom data storage as backend.(e.g. S3, Browser, etc.)
- Runing on browser with WebAssembly. (^v1.2.0)
- Support watch changes from remote. (^v1.2.0)

## Get Started
```bash
 go get github.com/diiyw/nodis@v1.2.0
```
```go
package main

import "github.com/diiyw/nodis"

func main() {
	// Create a new Nodis instance
	opt := nodis.DefaultOptions
	n := nodis.Open(opt)
	defer n.Close()
	// Set a key-value pair
	n.Set("key", []byte("value"))
	n.LPush("list", []byte("value1"))
}
```
Watch changes from remote

Server 
```go
package main

import (
	"fmt"
	"github.com/diiyw/nodis"
	"github.com/diiyw/nodis/pb"
	"github.com/diiyw/nodis/sync"
	"time"
)

func main() {
	var opt = nodis.DefaultOptions
	n := nodis.Open(opt)
	opt.Synchronizer = sync.NewWebsocket()
	n.Watch([]string{"*"}, func(op *pb.Operation) {
		fmt.Println("Server:", op.Key, string(op.Value))
	})
	go func() {
		for {
			time.Sleep(time.Second)
			n.Set("test", []byte(time.Now().Format("2006-01-02 15:04:05")))
		}
	}()
	err := n.Publish("127.0.0.1:6380", []string{"*"})
	if err != nil {
		panic(err)
	}
}
```
Browser client built with WebAssembly

```bash
GOOS=js GOARCH=wasm go build -o test.wasm
```
```go
package main

import (
	"fmt"
	"github.com/diiyw/nodis"
	"github.com/diiyw/nodis/fs"
	"github.com/diiyw/nodis/pb"
	"github.com/diiyw/nodis/sync"
)

func main() {
	var opt = nodis.DefaultOptions
	opt.Filesystem = &fs.Memory{}
	opt.Synchronizer = sync.NewWebsocket()
	n := nodis.Open(opt)
	n.Watch([]string{"*"}, func(op *pb.Operation) {
		fmt.Println("Subscribe: ", op.Key)
	})
	err := n.Subscribe("ws://127.0.0.1:6380")
	if err != nil {
		panic(err)
	}
	select {}
}
```
## Benchmark
Windows 11: 12C/32G
```bash
goos: windows
goarch: amd64
pkg: github.com/diiyw/nodis/bench
BenchmarkSet-12             1247017               844.3 ns/op           223 B/op          4 allocs/op
BenchmarkGet-12      		7624095               144.2 ns/op             7 B/op          0 allocs/op
BenchmarkLPush-12       	1331316               884.9 ns/op           271 B/op          5 allocs/op
BenchmarkLPop-12    		15884398              70.02 ns/op             8 B/op          1 allocs/op
BenchmarkSAdd-12    		1204911                1032 ns/op           335 B/op          6 allocs/op
BenchmarkSMembers-12      	7263865               142.0 ns/op             8 B/op          1 allocs/op
BenchmarkZAdd-12      		1311826               845.4 ns/op           214 B/op          7 allocs/op
BenchmarkZRank-12   		6371636               160.2 ns/op             7 B/op          0 allocs/op
BenchmarkHSet-12   		1000000                1079 ns/op           399 B/op          7 allocs/op
BenchmarkHGet-12    		6938287               183.0 ns/op             7 B/op          0 allocs/op
```
Linux VM: 2C/8GB
```bash
goos: linux
goarch: amd64
pkg: github.com/diiyw/nodis/bench             
BenchmarkSet-2        	 1000000	      1359 ns/op	     223 B/op	       4 allocs/op
BenchmarkGet-2        	 4724623	     214.9 ns/op	       7 B/op	       0 allocs/op
BenchmarkLPush-2      	 1000000	      1422 ns/op	     271 B/op	       5 allocs/op
BenchmarkLPop-2       	17787996	     71.42 ns/op	       8 B/op	       1 allocs/op
BenchmarkSAdd-2       	 1000000	      1669 ns/op	     335 B/op	       6 allocs/op
BenchmarkSMembers-2   	 5861822	     178.0 ns/op	       8 B/op	       1 allocs/op
BenchmarkZAdd-2       	 1000000	      1625 ns/op	     214 B/op	       7 allocs/op
BenchmarkZRank-2      	 5033864	     207.4 ns/op	       7 B/op	       0 allocs/op
BenchmarkHSet-2       	  939238	      1782 ns/op	     399 B/op	       7 allocs/op
BenchmarkHGet-2       	 6019508	     197.3 ns/op	       7 B/op	       0 allocs/op
```

## Note
If you want to persist data, please make sure to call the `Close()` method when your application exits.