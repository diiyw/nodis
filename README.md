# Nodis

![GitHub top language](https://img.shields.io/github/languages/top/diiyw/nodis) ![GitHub Release](https://img.shields.io/github/v/release/diiyw/nodis)

<div class="column" align="left">
  <a href="https://godoc.org/github.com/diiyw/nodis"><img src="https://godoc.org/github.com/diiyw/nodis?status.svg" /></a>
  <a href="https://goreportcard.com/report/github.com/diiyw/nodis"><img src="https://goreportcard.com/badge/github.com/diiyw/nodis" /></a>
  <a href="https://codecov.io/gh/diiyw/nodis"><img src="https://codecov.io/gh/diiyw/nodis/branch/main/graph/badge.svg?token=CupujOXpbe"/></a>
</div>

English | [简体中文](https://github.com/diiyw/nodis/blob/main/README_zh-cn.md)

Nodis is a Redis implementation using the Golang programming language. This implementation provides a simple way to embed Redis functionality directly into your application or run it as a standalone server. The supported commands are compatible with the original Redis protocol, allowing you to use existing Redis clients like goredis for testing and integration.

## Supported Data Types

- Bitmap
- String
- List
- Hash
- Set
- Sorted Set

## Key Features

- **Fast and Embeddable**: The Golang-based implementation is designed to be fast and easily embeddable within your applications.
- **Low Memory Usage**: The system only stores hot data in memory, minimizing the overall memory footprint.
- **Custom Data Storage Backends**: You can integrate custom data storage backends, such as Amazon S3,sqlite, browser-based storage, and more.
- **Browser Support with WebAssembly**: Starting from version 1.2.0, this Redis implementation can run directly in the browser using WebAssembly.
- **Remote Change Monitoring**: From version 1.2.0 onwards, the system supports watching for changes from remote sources.
- **Redis Protocol Compatibility**: As of version 1.5.0, this Redis implementation fully supports the original Redis protocol, ensuring seamless integration with existing Redis clients.

## Supported Commands

| **Client Handling** | **Configuration** | **Key Commands** | **String Commands** | **Set Commands** | **Hash Commands** | **List Commands** | **Sorted Set Commands** | **Geo Commands** |
| ------------------- | ----------------- | ---------------- | ------------------- | ---------------- | ----------------- | ----------------- |-------------------------| ---------------- |
| CLIENT              | FLUSHALL          | DEL              | GET                 | SADD             | HSET              | LPUSH             | ZADD                    | GEOADD		   |
| PING                | FLUSHDB           | EXISTS           | SET                 | SSCAN            | HGET              | RPUSH             | ZCARD                   | GEOPOS		   |
| QUIT                | SAVE              | EXPIRE           | INCR                | SCARD            | HDEL              | LPOP              | ZRANK                   | GEOHASH		   |
| ECHO                | INFO              | EXPIREAT         | DECR                | SPOP             | HLEN              | RPOP              | ZREVRANK                | GEODISH		   |
| DBSIZE              |                   | KEYS             | SETBIT              | SDIFF            | HKEYS             | LLEN              | ZSCORE                  | GEORADIUS		   |
| MULTI               |                   | TTL              | GETBIT              | SINTER           | HEXISTS           | LINDEX            | ZINCRBY                 | GEORADIUSBYMEMBER|
| DISCARD             |                   | RENAME           | INCR                | SISMEMBER        | HGETALL           | LINSERT           | ZRANGE                  |				   |
| EXEC                |                   | TYPE             | DESR                | SMEMBERS         | HINCRBY           | LPUSHX            | ZREVRANGE               |				   |
|                     |                   | SCAN             | SETEX               | SREM             | HICRBYFLOAT       | RPUSHX            | ZRANGEBYSCORE           |				   |
|                     |                   | RANDOMKEY        | INCRBY              | SMOVE            | HSETNX            | LREM              | ZREVRANGEBYSCORE        |				   |
|                     |                   | RENAMEEX         | DECRBY              | SRANDMEMBER      | HMGET             | LSET              | ZREM                    |				   |
|                     |                   | PERSIST          | SETNX               | SINTERSTORE      | HMSET             | LRANGE            | ZREMRANGEBYRANK         |				   |
|                     |                   | PTTL             | INCRBYFLOAT         | SUNIONSTORE      | HCLEAR            | LPOPRPUSH         | ZREMRANGEBYSCORE        |				   |
|                     |                   | UNLINK           | APPEND              |                  | HSCAN             | RPOPLPUSH         | ZCLEAR                  |				   |
|                     |                   |                  | GETRANGE            |                  | HVALS             | BLPOP             | ZEXISTS                 |				   |
|                     |                   |                  | STRLEN              |                  | HSTRLEN           | BRPOP             | ZUNIONSTORE             |				   |
|                     |                   |                  | SETRANGE            |                  |                   |                   | ZINTERSTORE             |				   |
|                     |                   |                  |                     |                  |                   |                   | ZSCAN                   |       |

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
	n.Set("key", []byte("value"),false)
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
	"github.com/diiyw/nodis/patch"
	"time"
)

func main() {
	var opt = nodis.DefaultOptions
	n := nodis.Open(opt)
	opt.Synchronizer = nodis.NewWebsocket()
	n.WatchKey([]string{"*"}, func(op patch.Op) {
		fmt.Println("Server:", op.Data.GetKey(), op.Data.(*patch.OpSet).Value)
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
	"github.com/diiyw/nodis/patch"
)

func main() {
	var opt = nodis.DefaultOptions
	opt.Synchronizer = nodis.NewWebsocket()
	n := nodis.Open(opt)
	n.WatchKey([]string{"*"}, func(op patch.Op) {
		fmt.Println("Subscribe: ", op.Data.GetKey())
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

Memory mode.

<details>
	<summary>Embed benchmark</summary>

Windows 11: 12C/32G

```bash
goos: windows
goarch: amd64
pkg: github.com/diiyw/nodis/bench
cpu: 12th Gen Intel(R) Core(TM) i5-12490F
BenchmarkSet
BenchmarkSet-12         	 2159343	       514.7 ns/op	     302 B/op	       8 allocs/op
BenchmarkGet
BenchmarkGet-12         	 6421864	       183.8 ns/op	     166 B/op	       3 allocs/op
BenchmarkLPush
BenchmarkLPush-12       	 2166828	       566.3 ns/op	     358 B/op	      10 allocs/op
BenchmarkLPop
BenchmarkLPop-12        	13069830	        80.41 ns/op	     159 B/op	       3 allocs/op
BenchmarkSAdd
BenchmarkSAdd-12        	 2007924	       592.6 ns/op	     406 B/op	      11 allocs/op
BenchmarkSMembers
BenchmarkSMembers-12    	 6303288	       179.8 ns/op	     166 B/op	       3 allocs/op
BenchmarkZAdd
BenchmarkZAdd-12        	 1580179	       832.6 ns/op	     302 B/op	      10 allocs/op
BenchmarkZRank
BenchmarkZRank-12       	 6011108	       186.7 ns/op	     165 B/op	       3 allocs/op
BenchmarkHSet
BenchmarkHSet-12        	 1997553	       654.3 ns/op	     486 B/op	      11 allocs/op
BenchmarkHGet
BenchmarkHGet-12        	 5895134	       193.3 ns/op	     165 B/op	       3 allocs/op
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

Windows 11: 12C/32G

```bash
redis-benchmark -p 6380 -t set,get,lpush,lpop,sadd,smembers,zadd,zrank,hset,hget -n 100000 -q
```

```
SET: 116144.02 requests per second
GET: 125156.45 requests per second
LPUSH: 121951.22 requests per second
LPOP: 126103.41 requests per second
SADD: 121951.22 requests per second
HSET: 122850.12 requests per second
```

</details>

## Note

If you want to persist data, please make sure to call the `Close()` method when your application exits.
