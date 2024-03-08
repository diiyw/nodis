# Nodis
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
- Low memory used, only hot key stored in memory
- Snapshot and WAL for data storage.

## Get Started
```bash
 go get github.com/diiyw/nodis 
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