# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Nodis is a Redis implementation in Go that can be embedded directly into applications or run as a standalone server. It supports the Redis protocol, allowing standard Redis clients (like redis-cli or go-redis) to connect.

## Common Commands

### Build and Run
```bash
# Run tests
go test

# Run tests with coverage
go test -coverprofile=coverage.out -covermode=atomic

# Run benchmarks
go test -bench=. ./bench/

# Build the standalone server
go build -o nodis-server ./cmd/nodis-server

# Run the server (default port 6380, memory storage)
./nodis-server

# Run with Pebble storage
./nodis-server :6380 pebble

# Build for WebAssembly (browser support)
GOOS=js GOARCH=wasm go build -o nodis.wasm
```

### Testing with Redis CLI
```bash
redis-cli -p 6380
```

## Architecture

### Core Components

- **nodis.go** - Main entry point. `Nodis` struct manages the store, listeners for key watching, and blocking operations. `Open()` initializes the database with options for GC and snapshot intervals.

- **store.go** - Internal key-value store with metadata tracking. Handles garbage collection (`gc()`), flushing changes to storage (`flush()`), and manages watched keys for transactions.

- **metadata.go** - Key metadata including expiration, modification state, and reference counting for memory management.

- **tx.go** - Transaction support with key locking for MULTI/EXEC commands.

### Data Structures (`ds/` package)

- `ds/str/` - String type (also used for bitmaps)
- `ds/list/` - Doubly linked list with generic variant
- `ds/hash/` - Hash map
- `ds/set/` - Set using map
- `ds/zset/` - Sorted set using skiplist

### Storage Layer (`storage/` package)

Implements the `Storage` interface:
- `memory.go` - In-memory storage (default)
- `pebble.go` - Persistent storage using CockroachDB's Pebble

### Redis Protocol (`redis/` package)

- `server.go` - TCP server handling connections
- `resp.go` - RESP protocol reader
- `cmd.go` - Command parsing with options

### Command Handlers

- **handler.go** - Maps Redis command names to handler functions via `GetCommand()`. All Redis commands are implemented here.
- Type-specific files: `str.go`, `list.go`, `hash.go`, `set.go`, `zset.go`, `geo.go`, `key.go`

### Pub/Sub and Replication

- **websocket.go** / **websocket_js.go** - WebSocket-based change broadcasting for replication
- **patch/** - Operation definitions (protobuf) for replicating changes between nodes
- `WatchKey()` / `Broadcast()` / `Subscribe()` - Key change notification system

## Key Patterns

### Adding a New Redis Command
1. Add handler function in `handler.go` following the pattern of existing commands
2. Register it in the `GetCommand()` switch statement
3. Implement the actual logic in the appropriate type file (e.g., `str.go` for string commands)

### Storage Backend
Implement the `storage.Storage` interface to add new storage backends:
```go
type Storage interface {
    Init() error
    Get(key *ds.Key) (ds.Value, error)
    Set(key *ds.Key, value ds.Value) error
    Delete(key *ds.Key) error
    Clear() error
    Close() error
    Snapshot() error
    ScanKeys(func(*ds.Key) bool)
}
```

## Supported Data Types

String, List, Hash, Set, Sorted Set, Bitmap, Geo
