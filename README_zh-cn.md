# Nodis
![GitHub top language](https://img.shields.io/github/languages/top/diiyw/nodis) ![GitHub Release](https://img.shields.io/github/v/release/diiyw/nodis)
<div class="column" align="left">
  <a href="https://godoc.org/github.com/diiyw/nodis"><img src="https://godoc.org/github.com/diiyw/nodis?status.svg" /></a>
  <a href="https://goreportcard.com/report/github.com/diiyw/nodis"><img src="https://goreportcard.com/badge/github.com/diiyw/nodis" /></a>
  <a href="https://codecov.io/gh/diiyw/nodis"><img src="https://codecov.io/gh/diiyw/nodis/branch/master/graph/badge.svg?token=CupujOXpbe"/></a>
</div>


[English](https://github.com/diiyw/nodis/blob/main/README.md) | 简体中文

Nodis 是一个简单可嵌入到应用中内存数据库，实现Redis的数据结构。

## 支持的类型

- String
- List
- Hash
- Set
- Sorted Set

## 特点

- 快速可嵌入的
- 低内存使用，只有热数据才在内存中
- 快照和WAL存储的支持

## Get Started

```bash
go get github.com/diiyw/nodis@v1.1.0-beta.11
```

```go
package main

import "github.com/diiyw/nodis"

func main() {
	// Create a new Nodis instance
	opt := nodis.DefaultOptions
	n := nodis.Open(opt)

	// Set a key-value pair
	n.Set("key", []byte("value"), 0)
	n.LPush("list", []byte("value1"))
}

```

## Note

Nodis 实现了Redis的数据结构. 但是并不是完整的Redis Server服务，它只是可以方便的切入到各自的应用使用
