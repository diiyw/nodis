# Execution, persistence, and synchronization

## Execution contract

A database owns one execution RWMutex. Writes, persistence, GC, WATCH changes, and
complete commands use the exclusive lock; hot `Get` and `Exists` reads use a shared
lock. Cold `Get` reads fall back to the exclusive path to load storage values.
There are no per-key locks or lock upgrades. Nested operations borrow the same
`Tx`; a complete Redis command and every command inside `EXEC` use one execution
context. This provides isolation, including compound SET options and multi-key
operations. Runtime command errors do not roll back preceding commands in EXEC,
matching its execution model.

Blocking list commands register their wake hints atomically with checking the
lists, release the execution lock while waiting, and recheck after every wake.
Inside EXEC they return immediately when no element is available. Closing the
database wakes blocked callers and closes its Redis and synchronization servers.

WATCH state uses the execution mutex too. Deletion, replacement, expiration,
rename, and clear invalidate watches. EXEC checks expiry and the watch flags in
its own execution context; EXEC, DISCARD, UNWATCH, and disconnect remove registrations.

`OpenWithError` reports initialization failures; `Open` retains the convenient
pointer-only API and panics on failure. Options are copied at open. Supply a
separate Storage and Channel instance for each database; explicitly supplying the
same backend is not a shared-database concurrency mechanism.

Byte slices returned by embedded APIs are borrowed views, preserving zero-allocation
hot reads. Callers must not mutate them or retain them across concurrent in-place
mutations such as SETBIT/SETRANGE; copy when ownership is needed. Input byte slices
transferred to the database must likewise not be mutated by callers. Redis replies
and replication snapshots are built while execution is isolated. RESP parser
buffers used by writes are retained as immutable argument storage; small read-only
requests can reuse their buffers. Queued commands always own their arguments.

## Durability contract

A successful ordinary write commits to memory, not disk. Call `Flush()` when the
application needs a persistence barrier. `Snapshot()` flushes first and then makes
a backend checkpoint. `Close()` stops workers and networking, flushes, and closes
the backend only after that flush succeeds. Failed Flush/Close calls return errors;
dirty values, tombstones, and pending clears remain available for retry. A failed
Close stops networking but leaves the store available for recovery and another Close.

`GCDuration` periodically flushes pending changes before evicting cold values.
No value with an unsuccessful write is marked clean or evicted. A process crash
before a successful persistence barrier may lose the unflushed changes. Flush
visits dirty keys and tombstones only; each GC pass examines at most 256 indexed
keys, rotating through the full key set over subsequent passes. Shared hot reads
record an atomic access bit that the next GC pass uses to preserve recently used
values for one cycle. Repeated access in one cycle does not accumulate eviction
credit. Expired entries
are removed from memory in that bounded pass and their durable tombstones are
flushed on the next pass. Storage I/O remains under the execution mutex.

Storage identity is the logical key name. Expiration is record metadata. Memory
owns serialized snapshots rather than references to live mutable structures.
Pebble uses an atomic synchronous batch for replacements, deletions, and clear.
The persisted key format is `0x01 || name`; a reserved `0x00` key stores the format
version. Values contain `NDR1`, an eight-byte little-endian signed Unix millisecond
expiration, a one-byte value type, and the encoded value. Zero expiration means
persistent. Checkpoints are written beside the live directory, under
`<path>-snapshots/<UnixNano>`.

Custom backends implement `storage.Storage`, including
`ScanKeys(func(*ds.Key) bool) error`. Implement `storage.BatchStorage.Apply` to
provide atomic Flush. Without it, Flush retries the entire change set after any
failure, but cannot promise atomic recovery across a crash during a partial batch.

Only the current Pebble format is supported. A nonempty database without the
current format marker, or with an unknown marker, is rejected. There is no legacy
reader or migration utility; existing data must be exported by the application
before switching formats.

## NDS1 stream

The project synchronization layer has no protobuf imports, generated files, or
runtime. Pebble's Prometheus metrics and CockroachDB error dependencies still
bring protobuf modules into the transitive dependency graph.

`NewStream()` (the default Channel) carries frames over TCP. Pass `host:port` to
Broadcast/Subscribe. `NewWebsocket()` uses exactly the same frames in binary
WebSocket messages, with `ws://host:port` for subscription. Browser WebSocket
messages use ArrayBuffer directly to preserve delivery order. These formats are
incompatible with the previous protobuf stream; upgrade both ends together.

| Offset | Size | Content |
| --- | --- | --- |
| 0 | 4 | ASCII `NDS1` (magic and protocol version) |
| 4 | 4 | Total frame size, unsigned big-endian, including checksum |
| 8 | 8 | Per-connection sequence, unsigned big-endian, starting at 1 |
| 16 | 1 | Operation code |
| 17 | Variable | Operation fields, in their declared order |
| Last 4 | 4 | CRC32C of every preceding frame byte, unsigned big-endian |

Strings/bytes use an unsigned Varint length followed by bytes. Signed integers
use Go's signed Varint encoding. Doubles are IEEE-754 binary64 little-endian;
booleans are 0/1 bytes. Arrays start with an unsigned Varint element count.
`patch/patch.go` defines the operation codes and exact field order. The decoder
validates lengths, counts, types, checksum, frame boundaries, and contiguous
connection sequences. Frames are capped at 64 MiB. `AppendEncode` and `Encoder`
reuse caller/encoder buffers; decoded payloads own their data.

A connection registers its subscription and captures its initial snapshot in the
same execution context. The first batch resets matching replica keys and restores
the current matching state. Subsequent batches contain final state replacements
and tombstones from each completed command/EXEC, including absolute expiration.
This handles destination/source filtering, repeated keys, rename, list moves,
random removals, and clear without replaying nondeterministic operations.
A replica validates the complete state batch before applying it under one execution
lock. Applying a remote batch notifies local watchers but does not echo it to other
replication connections. This is source-to-replica synchronization, not multi-writer
conflict resolution. Do not write independently to replicas that must mirror a source.

Each subscription has one FIFO worker; callbacks run outside database locks.
Subscription IDs never move. Replication queues are bounded by 1024 batches and
64 MiB of pending frames. A slow or oversized connection is closed instead of
silently dropping committed updates. An initial snapshot or atomic transaction
must fit in one frame; otherwise the stream closes explicitly. Reconnection starts
a full snapshot rather than resuming from a durable log. Local WatchKey callbacks
have an ordered queue and should return promptly; UnWatchKey cancels pending calls.

## Verification

`go test ./...`, `go test -race ./...`, `go vet ./...`, and
`GOOS=js GOARCH=wasm go build ./...` cover the native and browser build paths.
Regression tests exercise crash/reopen behavior, failure retry, concurrent first
writes and cold reads, WATCH/EXEC, blocking pops, malformed RESP, owned arguments,
all operation codecs, FIFO delivery, atomic state batches, TCP snapshot/incremental
sync and storage format validation. Fuzz targets cover RESP and NDS1 input.

Steady-state embedded GET, SET of an existing key, and EXISTS, plus buffered GET
RESP parsing and preallocated NDS1 encoding, have allocation regression checks
and/or benchmarks. Creating data, returning collections, and owning persisted or
replicated snapshots necessarily allocate; the hot execution lock adds no allocations.
