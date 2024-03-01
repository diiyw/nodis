# Nodis
A Golang implemented Redis data structure. 
It is a simple and easy to use in-memory key-value store.

## Supported Data Types

- String
- List
- Hash
- Sorted Set

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
}

```