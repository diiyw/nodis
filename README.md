# Nodis

Nodis is a schema-free document database for embedded and client/server use. It provides compact binary objects, transactional object pipelines, field-level authorization, a native WAL with durable snapshots, and permission-filtered client synchronization over TCP or TLS 1.3 (including mutual TLS).

Use `DB.Checkpoint()` for manual WAL compaction or `Options.CheckpointInterval` for periodic checkpoints. Recovery preserves field IDs, ACLs, pipeline bindings, and object bytes. Checkpoints currently hold the database write lock; live objects remain in memory.

Queries support generic scalar filters, field projections, and ascending ObjectID keyset pagination through `DB.Find` / `FindInto` and client `Find` / `FindAsync`. Predicate evaluation and projection enforce field permissions. A reusable query buffer, codecs, and server execution allocate zero objects after warmup. A derived ordered primary index is rebuilt on recovery; field filters currently scan candidates under a read lock, and pages do not share a retained snapshot.

`client.Dial` automatically synchronizes authorized data. `client.DialWithOptions` can disable synchronization, and `client.DialTLS` encrypts requests and synchronization. Successful synchronous results and generic `Future[T]` completions confirm server execution; local cached views update independently.

Clients invalidate their cache on detected synchronization failure and publish replacement snapshots atomically. `SyncStatus()` exposes readiness, the received-event watermark, permission revision, and errors. Reconnection rebuilds authorized data, including after permission restoration. Interrupted requests discard the command connection; the next request reconnects without replaying the failed one. Interrupted mutations can return `client.ErrOutcomeUnknown`: the server may have committed despite the missing response.

Generic storage and network frames reuse buffers. Regression tests enforce zero allocations for prepared scalar/cache lookups, sync status reads, protocol encoding, warmed frame reads/writes (including real TCP), and WAL appends; connection setup and full CRUD operations may allocate.

See [README_zh-cn.md](README_zh-cn.md) and [DESIGN.md](DESIGN.md).
