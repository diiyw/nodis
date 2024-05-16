# Nodis

![GitHub top language](https://img.shields.io/github/languages/top/diiyw/nodis) ![GitHub Release](https://img.shields.io/github/v/release/diiyw/nodis)

<div class="column" align="left">
  <a href="https://godoc.org/github.com/diiyw/nodis"><img src="https://godoc.org/github.com/diiyw/nodis?status.svg" /></a>
  <a href="https://goreportcard.com/report/github.com/diiyw/nodis"><img src="https://goreportcard.com/badge/github.com/diiyw/nodis" /></a>
  <a href="https://codecov.io/gh/diiyw/nodis"><img src="https://codecov.io/gh/diiyw/nodis/branch/main/graph/badge.svg?token=CupujOXpbe"/></a>
</div>

English | [简体中文](https://github.com/diiyw/nodis/blob/main/README_zh-cn.md)

Nodis 是一个使用 Golang 编程语言实现的 Redis。这个实现提供了一种将 Redis 功能直接嵌入到应用程序中或作为独立服务器运行的简单方法。支持的命令与原始 Redis 协议兼容,允许您使用现有的 Redis 客户端(如 goredis)进行测试和集成。

## 支持的数据类型

Bitmap
String
List
Hash
Set
Sorted Set

## 主要特性

- **快速和可嵌入**: Golang 实现的设计目标是快速和易于嵌入到您的应用程序中。
- **低内存使用**: 该系统只在内存中存储热数据,将整体内存占用降到最低。
- **快照和 WAL 用于数据存储**: 这个 Redis 实现支持快照和预写日志(WAL)机制,用于可靠的数据存储。
- **自定义数据存储后端**: 您可以集成自定义的数据存储后端,如 Amazon S3、浏览器存储等。
  使用 WebAssembly 支持浏览器: 从 1.2.0 版本开始,这个 Redis 实现可以直接在浏览器中使用 WebAssembly 运行。
- **远程变更监控**: 从 1.2.0 版本开始,该系统支持监视来自远程源的变更。
- **Redis 协议兼容性**: 从 1.5.0 版本开始,这个 Redis 实现完全支持原始的 Redis 协议,确保与现有的 Redis 客户端无缝集成。

## 支持的 Redis 命令

| **Client Handling** | **Configuration** | **Key Commands** | **String Commands** | **Set Commands** | **Hash Commands** | **List Commands** | **Sorted Set Commands** |
| ------------------- | ----------------- | ---------------- | ------------------- | ---------------- | ----------------- | ----------------- | ----------------------- |
| CLIENT              | FLUSHALL          | DEL              | GET                 | SADD             | HSET              | LPUSH             | ZADD                    |
| PING                | FLUSHDB           | EXISTS           | SET                 | SSCAN            | HGET              | RPUSH             | ZCARD                   |
| QUIT                | SAVE              | EXPIRE           | INCR                | SCARD            | HDEL              | LPOP              | ZRANK                   |
| ECHO                | INFO              | EXPIREAT         | DECR                | SPOP             | HLEN              | RPOP              | ZREVRANK                |
| DBSIZE              |                   | KEYS             | SETBIT              | SDIFF            | HKEYS             | LLEN              | ZSCORE                  |
| MULTI               |                   | TTL              | GETBIT              | SINTER           | HEXISTS           | LINDEX            | ZINCRBY                 |
| DISCARD             |                   | RENAME           | INCR                | SISMEMBER        | HGETALL           | LINSERT           | ZRANGE                  |
| EXEC                |                   | TYPE             | DESR                | SMEMBERS         | HINCRBY           | LPUSHX            | ZREVRANGE               |
|                     |                   | SCAN             | SETEX               | SREM             | HICRBYFLOAT       | RPUSHX            | ZRANGEBYSCORE           |
|                     |                   | RANDOMKEY        | INCRBY              | SMOVE            | HSETNX            | LREM              | ZREVRANGEBYSCORE        |
|                     |                   | RENAMEEX         | DECRBY              | SRANDMEMBER      | HMGET             | LSET              | ZREM                    |
|                     |                   | PERSIST          | SETNX               | SINTERSTORE      | HMSET             | LRANGE            | ZREMRANGEBYRANK         |
|                     |                   |                  | INCRBYFLOAT         | SUNIONSTORE      | HCLEAR            | LPOPRPUSH         | ZREMRANGEBYSCORE        |
|                     |                   |                  | APPEND              |                  | HSCAN             | RPOPLPUSH         | ZCLEAR                  |
|                     |                   |                  | GETRANGE            |                  | HVALS             | BLPOP             | ZEXISTS                 |
|                     |                   |                  | STRLEN              |                  | HSTRLEN           | BRPOP             | ZUNIONSTORE             |
|                     |                   |                  | SETRANGE            |                  |                   |                   | ZINTERSTORE             |

## Get Started

```bash
 go get github.com/diiyw/nodis@v1.5.0
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
	n.Set("key", []byte("value"), false)
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
	n.Stick([]string{"*"}, func(op *pb.Operation) {
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

## Note

如果你想持久化请保证，在你的应用退出时调用`Close()`方法。
