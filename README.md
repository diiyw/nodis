# Nodis
![GitHub top language](https://img.shields.io/github/languages/top/diiyw/nodis) ![GitHub Release](https://img.shields.io/github/v/release/diiyw/nodis)
<div class="column" align="left">
  <a href="https://godoc.org/github.com/diiyw/nodis"><img src="https://godoc.org/github.com/diiyw/nodis?status.svg" /></a>
  <a href="https://goreportcard.com/report/github.com/diiyw/nodis"><img src="https://goreportcard.com/badge/github.com/diiyw/nodis" /></a>
  <a href="https://codecov.io/gh/diiyw/nodis"><img src="https://codecov.io/gh/diiyw/nodis/branch/main/graph/badge.svg?token=CupujOXpbe"/></a>
</div>


English | [简体中文](https://github.com/diiyw/nodis/blob/main/README_zh-cn.md)

Redis re-implemented using golang. 
Simple way to embed in your application.

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
- Running on browser with WebAssembly. (^v1.2.0)
- Support watch changes from remote. (^v1.2.0)
- Support redis protocol. (^v1.3.0)

## Get Started
```bash
 go get github.com/diiyw/nodis@v1.2.0
```
Or use test version
```bash
 go get github.com/diiyw/nodis@main
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
- Watch changes from remote `Server`
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
- Browser client built with WebAssembly

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
BenchmarkSet-12         	 1469863	        715.9 ns/op	     543 B/op	       7 allocs/op
BenchmarkGet-12         	12480278	        96.47 ns/op	       7 B/op	       0 allocs/op
BenchmarkLPush-12       	 1484466	        786.2 ns/op	     615 B/op	       9 allocs/op
BenchmarkLPop-12        	77275986	        15.10 ns/op	       0 B/op	       0 allocs/op
BenchmarkSAdd-12        	 1542252	        831.9 ns/op	     663 B/op	      10 allocs/op
BenchmarkSMembers-12    	12739020	        95.18 ns/op	       8 B/op	       1 allocs/op
BenchmarkZAdd-12        	 1000000	        1177 ns/op	     550 B/op	      10 allocs/op
BenchmarkZRank-12       	11430135	        104.1 ns/op	       7 B/op	       0 allocs/op
BenchmarkHSet-12        	 1341817	        863.5 ns/op	     743 B/op	      11 allocs/op
BenchmarkHGet-12        	 9801158	        105.9 ns/op	       7 B/op	       0 allocs/op
```
Linux VM: 2C/8GB
```bash
goos: linux
goarch: amd64
pkg: github.com/diiyw/nodis/bench             
BenchmarkSet-2        	  750900	      1828 ns/op	     591 B/op	       8 allocs/op
BenchmarkGet-2        	 4765485	       247.9 ns/op	      55 B/op	       1 allocs/op
BenchmarkLPush-2      	  851473	      1866 ns/op	     663 B/op	      10 allocs/op
BenchmarkLPop-2       	18313623	        56.78 ns/op	      49 B/op	       1 allocs/op
BenchmarkSAdd-2       	  857107	      2231 ns/op	     710 B/op	      11 allocs/op
BenchmarkSMembers-2   	 4297828	       306.2 ns/op	      56 B/op	       2 allocs/op
BenchmarkZAdd-2       	  788445	      2082 ns/op	     598 B/op	      11 allocs/op
BenchmarkZRank-2      	 3196694	       329.8 ns/op	      55 B/op	       1 allocs/op
BenchmarkHSet-2       	  823741	      2200 ns/op	     790 B/op	      12 allocs/op
BenchmarkHGet-2       	 4493481	       290.2 ns/op	      55 B/op	       1 allocs/op
```

## Note
If you want to persist data, please make sure to call the `Close()` method when your application exits.