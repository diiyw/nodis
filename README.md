# Nodis
![GitHub top language](https://img.shields.io/github/languages/top/diiyw/nodis) ![GitHub Release](https://img.shields.io/github/v/release/diiyw/nodis)


English | [简体中文](https://github.com/diiyw/nodis/blob/main/README_zh-cn.md)

A Golang implemented Redis data structure. 
It is a simple and easy to embed in your application.

## Supported Data Types

- String
- List
- Hash
- Set
- Sorted Set

## Features

- Fast and embeddable
- Low memory used, only hot data stored in memory
- Snapshot and WAL for data storage.

## Get Started
```bash
 go get github.com/diiyw/nodis@v1.1.0-beta.3
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
Nodis is done by following the Redis data structure. It is not a complete Redis server. It is a simple and easy to embed in your application.