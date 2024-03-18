# Nodis
![GitHub top language](https://img.shields.io/github/languages/top/diiyw/nodis) ![GitHub Release](https://img.shields.io/github/v/release/diiyw/nodis)
<div class="column" align="left">
  <a href="https://godoc.org/github.com/diiyw/nodis"><img src="https://godoc.org/github.com/diiyw/nodis?status.svg" /></a>
  <a href="https://goreportcard.com/report/github.com/diiyw/nodis"><img src="https://goreportcard.com/badge/github.com/diiyw/nodis" /></a>
  <a href="https://goreportcard.com/report/github.com/diiyw/nodis"><img src="https://github.com/diiyw/nodis/workflows/Go/badge.svg?branch=main"/></a>
  <a href="https://codecov.io/gh/diiyw/nodis"><img src="https://codecov.io/gh/diiyw/nodis/branch/main/graph/badge.svg?token=CupujOXpbe"/></a>
</div>


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
 go get github.com/diiyw/nodis@v1.1.0-beta.10
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

## Benchmark
```bash
goos: windows
goarch: amd64
pkg: github.com/diiyw/nodis/bench
cpu: 12th Gen Intel(R) Core(TM) i5-12490F
BenchmarkSet-12               2095956                   538.0 ns/op            269B/op            3 allocs/op
BenchmarkGet-12               15941229                  68.03 ns/op            7B/op              0 allocs/op
BenchmarkLPush-12             2238814                   608.0 ns/op            306B/op            4 allocs/op
BenchmarkLPop-12              18975078                  63.12 ns/op            7B/op              0 allocs/op
BenchmarkSAdd-12              1000000                   1029 ns/op             1285B/op           6 allocs/op
BenchmarkSMembers-12          17402097                  66.90 ns/op            8B/op              1 allocs/op
BenchmarkZAdd-12              1847750                   704.0 ns/op            245B/op            7 allocs/op
BenchmarkZRank-12             16226691                  74.38 ns/op            7B/op              0 allocs/op
BenchmarkHSet-12              1000000                   1704 ns/op             2453B/op           7 allocs/op
BenchmarkHGet-12              16762840                  73.54 ns/op            7B/op              0 allocs/op
```

## Note
Nodis is done by following the Redis data structure. It is not a complete Redis server. It is a simple and easy to embed in your application.
