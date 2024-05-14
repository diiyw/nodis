# Nodis
![GitHub top language](https://img.shields.io/github/languages/top/diiyw/nodis) ![GitHub Release](https://img.shields.io/github/v/release/diiyw/nodis)
<div class="column" align="left">
  <a href="https://godoc.org/github.com/diiyw/nodis"><img src="https://godoc.org/github.com/diiyw/nodis?status.svg" /></a>
  <a href="https://goreportcard.com/report/github.com/diiyw/nodis"><img src="https://goreportcard.com/badge/github.com/diiyw/nodis" /></a>
  <a href="https://codecov.io/gh/diiyw/nodis"><img src="https://codecov.io/gh/diiyw/nodis/branch/main/graph/badge.svg?token=CupujOXpbe"/></a>
</div>


English | [简体中文](https://github.com/diiyw/nodis/blob/main/README_zh-cn.md)

Redis re-implemented using golang. 
Simple way to embed in your application or as a standalone server.
Supported commands are compatible with Redis. testing code using [phpredis](https://github.com/phpredis/phpredis/tree/develop/tests).

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
- Support redis protocol. (^v1.5.0)

## Supported Commands
| **Client Handling** | **Configuration** | **Key Commands** | **String Commands** | **Set Commands** | **Hash Commands** | **List Commands** | **Sorted Set Commands** |
|---------------------|-----------------|-----------------|---------------------|-----------------|-----------------|------------------|----------------|
| CLIENT              | FLUSHALL       	| DEL             | GET                 | SADD            | HSET            | LPUSH            | ZADD                  |
| PING                | FLUSHDB     	| EXISTS          | SET                 | SSCAN           | HGET            | RPUSH            | ZCARD                 |
| QUIT                | SAVE       		| EXPIRE          | INCR                | SCARD           | HDEL            | LPOP             | ZRANK                 |
| ECHO                | INFO          	| EXPIREAT        | DECR                | SPOP            | HLEN            | RPOP             | ZREVRANK              |
| DBSIZE              |             	| KEYS            | SETBIT              | SDIFF           | HKEYS           | LLEN             | ZSCORE                |
| MULTI               |                 | TTL             | GETBIT              | SINTER          | HEXISTS         | LINDEX           | ZINCRBY               |
| DISCARD             |                 | RENAME          | INCR              	| SISMEMBER       | HGETALL         | LINSERT          | ZRANGE                |
| EXEC                |                 | TYPE            | DESR                | SMEMBERS        | HINCRBY         | LPUSHX           | ZREVRANGE             |
|                     |                 | SCAN            | SETEX               | SREM            | HICRBYFLOAT    	| RPUSHX           | ZRANGEBYSCORE         |
|                     |                 | RANDOMKEY       | INCRBY              | SMOVE           | HSETNX          | LREM             | ZREVRANGEBYSCORE      |
|                     |                 | RENAMEEX        | DECRBY              | SRANDMEMBER     | HMGET           | LSET             | ZREM                  |
|                     |                 | PERSIST         | SETNX               | SINTERSTORE     | HMSET           | LRANGE           | ZREMRANGEBYRANK       |
|                     |                 |                 | INCRBYFLOAT         | SUNIONSTORE     | HCLEAR          | LPOPRPUSH        | ZREMRANGEBYSCORE      |
|                     |                 |                 | APPEND              |                 | HSCAN           | RPOPLPUSH        | ZCLEAR                |
|                     |                 |                 | GETRANGE            |                 | HVALS           | BLPOP            | ZEXISTS               |
|                     |                 |                 | STRLEN              |                 | HSTRLEN         | BRPOP            | ZUNIONSTORE           |
|                     |                 |                 | SETRANGE            |                 |                 |                  | ZINTERSTORE		   |
## Get Started
```bash
 go get github.com/diiyw/nodis@latest
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
## Examples

<details>
	<summary> Watch changes</summary>

Server: 
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
</details>
<details>
	<summary> Simple Redis Server</summary>

```go	
package main

import (
	"fmt"
	"net/http"

	"github.com/diiyw/nodis"
)

func main() {
	opt := nodis.DefaultOptions
	n := nodis.Open(opt)
	if err := n.Serve(":6380"); err != nil {
		fmt.Printf("Serve() = %v", err)
	}
}
```
You can use redis-cli to connect to the server.

```bash
redis-cli -p 6380
> set key value
```

</details>

## Benchmark
<details>
	<summary>Embed benchmark</summary>

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

Linux VM: 4C/8GB

```bash
goos: linux
goarch: amd64
pkg: github.com/diiyw/nodis/bench             
BenchmarkSet-4        	  806912	      1658 ns/op	     543 B/op	       7 allocs/op
BenchmarkGet-4        	 5941904	       190.6 ns/op	       7 B/op	       0 allocs/op
BenchmarkLPush-4      	  852932	      1757 ns/op	     615 B/op	       9 allocs/op
BenchmarkLPop-4       	40668902	        27.22 ns/op	       0 B/op	       0 allocs/op
BenchmarkSAdd-4       	  706376	      1913 ns/op	     662 B/op	      10 allocs/op
BenchmarkSMembers-4   	 4819993	       208.1 ns/op	       8 B/op	       1 allocs/op
BenchmarkZAdd-4       	  729039	      2013 ns/op	     550 B/op	      10 allocs/op
BenchmarkZRank-4      	 4959448	       246.4 ns/op	       7 B/op	       0 allocs/op
BenchmarkHSet-4       	  735676	      1971 ns/op	     742 B/op	      11 allocs/op
BenchmarkHGet-4       	 4442625	       243.4 ns/op	       7 B/op	       0 allocs/op
```

</details>
<details>
	<summary>Redis benchmark tool</summary>

```bash
redis-benchmark -p 6380 -t set,get,lpush,lpop,sadd,smembers,zadd,zrank,hset,hget -n 100000 -q   
```

```
SET: 89126.56 requests per second
GET: 90415.91 requests per second
LPUSH: 91491.30 requests per second
LPOP: 92165.90 requests per second
SADD: 91911.76 requests per second
HSET: 93023.25 requests per second
```
</details>

## Note
If you want to persist data, please make sure to call the `Close()` method when your application exits.
