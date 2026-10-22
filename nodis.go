package nodis

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/internal/listener"
	"github.com/diiyw/nodis/patch"
	"github.com/diiyw/nodis/redis"
	"github.com/diiyw/nodis/storage"
)

var ErrUnknownOperation = errors.New("unknown operation")

type instance struct {
	serverMu          sync.Mutex
	server            *redis.Server
	store             *store
	listenersMu       sync.Mutex
	listeners         map[int]*listener.Listener
	listenerCount     atomic.Int64
	nextListenerID    int
	blockingKeysMutex sync.Mutex
	blockingKeys      map[string]map[chan struct{}]struct{}
	options           Options
	done              chan struct{}
	stopOnce          sync.Once
	workers           sync.WaitGroup
}
type Nodis struct {
	*instance
	tx *Tx
}

func Open(opt *Options) *Nodis {
	n, err := OpenWithError(opt)
	if err != nil {
		panic(err)
	}
	return n
}
func OpenWithError(opt *Options) (*Nodis, error) {
	opts := *DefaultOptions
	if opt != nil {
		opts = *opt
	}
	if opts.Storage == nil {
		opts.Storage = storage.NewMemory()
	}
	if opts.Channel == nil {
		opts.Channel = NewStream()
	}
	s, err := newStore(opts.Storage)
	if err != nil {
		return nil, err
	}
	n := &Nodis{instance: &instance{store: s, options: opts, listeners: make(map[int]*listener.Listener), blockingKeys: make(map[string]map[chan struct{}]struct{}), done: make(chan struct{})}}
	n.periodic(opts.GCDuration, func() error { return s.gc() })
	n.periodic(opts.SnapshotDuration, n.Snapshot)
	return n, nil
}
func (n *Nodis) periodic(interval time.Duration, fn func() error) {
	if interval <= 0 {
		return
	}
	n.workers.Go(func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-n.done:
				return
			case <-ticker.C:
				if err := fn(); err != nil {
					log.Println("nodis background:", err)
				}
			}
		}
	})
}

// Flush persists all prior memory commits. A failed Flush is safe to retry.
func (n *Nodis) Flush() error { return n.exec(func(tx *Tx) error { return n.store.flushLocked() }) }
func (n *Nodis) Snapshot() error {
	return n.exec(func(tx *Tx) error {
		if err := n.store.flushLocked(); err != nil {
			return err
		}
		return n.store.ss.Snapshot()
	})
}
func (n *Nodis) Close() error {
	n.stopOnce.Do(func() { close(n.done) })
	n.serverMu.Lock()
	if n.server != nil {
		_ = n.server.Close()
	}
	n.serverMu.Unlock()
	if closer, ok := n.options.Channel.(interface{ Close() error }); ok {
		_ = closer.Close()
	}
	n.workers.Wait()
	n.store.mu.Lock()
	defer n.store.mu.Unlock()
	if n.store.closed {
		return nil
	}
	if err := n.store.flushLocked(); err != nil {
		return err
	}
	if err := n.store.ss.Close(); err != nil {
		return err
	}
	n.store.closed = true
	clear(n.store.watchedKeys)
	n.listenersMu.Lock()
	n.listenerCount.Store(0)
	for id, w := range n.listeners {
		w.Stop()
		delete(n.listeners, id)
	}
	n.listenersMu.Unlock()
	return nil
}
func (n *Nodis) Clear() {
	_ = n.exec(func(tx *Tx) error {
		n := n.inTx(tx)
		for key := range n.store.watchedKeys {
			n.store.invalidate(key)
		}
		clear(n.store.metadata)
		clear(n.store.dirty)
		clear(n.store.deleted)
		clear(n.store.gcPositions)
		clear(n.store.gcKeys)
		n.store.gcKeys = n.store.gcKeys[:0]
		n.store.gcCursor = 0
		n.store.clearPending = true
		tx.cleared = true
		clear(n.store.changed)
		n.notify(func() []patch.Op { return []patch.Op{{Type: patch.OpTypeClear, Data: &patch.OpClear{}}} })
		return nil
	})
}

// notify snapshots payload ownership under the execution lock. Callbacks run only
// after the entire command/EXEC finishes, in commit order, outside that lock.
func (n *Nodis) notify(f func() []patch.Op) {
	if n.listenerCount.Load() == 0 {
		return
	}
	n.listenersMu.Lock()
	hasListeners := false
	for _, w := range n.listeners {
		if !w.Batch {
			hasListeners = true
			break
		}
	}
	n.listenersMu.Unlock()
	if !hasListeners {
		return
	}
	for _, op := range f() {
		data, err := op.AppendEncode(nil)
		if err != nil {
			panic(err)
		}
		owned, err := patch.DecodeOp(data)
		if err != nil {
			panic(err)
		}
		n.store.events = append(n.store.events, owned)
	}
}
func (n *Nodis) WatchKey(pattern []string, fn func(patch.Op)) int {
	n.listenersMu.Lock()
	defer n.listenersMu.Unlock()
	select {
	case <-n.done:
		return -1
	default:
	}
	id := n.nextListenerID
	n.nextListenerID++
	n.listeners[id] = listener.New(pattern, fn)
	n.listenerCount.Add(1)
	return id
}
func (n *Nodis) UnWatchKey(id int) {
	n.listenersMu.Lock()
	defer n.listenersMu.Unlock()
	if w := n.listeners[id]; w != nil {
		w.Stop()
		delete(n.listeners, id)
		n.listenerCount.Add(-1)
	}
}
func (n *Nodis) ApplyPatch(ops ...patch.Op) (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("nodis: apply patch: %v", p)
		}
	}()
	return n.exec(func(tx *Tx) error {
		scoped := n.inTx(tx)
		tx.suppressEvents = true
		for _, op := range ops {
			if err := scoped.applyPatch(op); err != nil {
				return err
			}
		}
		return nil
	})
}
func (n *Nodis) applyPatch(p patch.Op) error {
	switch op := p.Data.(type) {
	case *patch.OpBatch:
		if len(op.Frames) > 1<<20 {
			return patch.ErrInvalidFrame
		}
		decoded := make([]patch.Op, len(op.Frames))
		values := make([]ds.Value, len(op.Frames))
		for i, frame := range op.Frames {
			var err error
			decoded[i], err = patch.DecodeOp(frame)
			if err != nil {
				return err
			}
			switch child := decoded[i].Data.(type) {
			case *patch.OpRestore:
				values[i], err = storage.DecodeValue(child.Value)
				if err != nil {
					return err
				}
			case *patch.OpDel, *patch.OpReset:
			default:
				return patch.ErrInvalidFrame
			}
		}
		for i, child := range decoded {
			if restore, ok := child.Data.(*patch.OpRestore); ok {
				m := newMetadata(ds.NewKey(restore.Key, restore.Expiration), true)
				m.setValue(values[i])
				n.tx.storeMeta(m)
				n.signalModifiedKey(restore.Key, m)
				n.notify(func() []patch.Op { return []patch.Op{{Type: patch.OpTypeRestore, Data: restore}} })
			} else if err := n.applyPatch(child); err != nil {
				return err
			}
		}
	case *patch.OpRestore:
		value, err := storage.DecodeValue(op.Value)
		if err != nil {
			return err
		}
		m := newMetadata(ds.NewKey(op.Key, op.Expiration), true)
		m.setValue(value)
		n.tx.storeMeta(m)
		n.signalModifiedKey(op.Key, m)
		n.notify(func() []patch.Op { return []patch.Op{{Type: patch.OpTypeRestore, Data: op}} })
	case *patch.OpReset:
		for _, pattern := range op.Patterns {
			for _, key := range n.Keys(pattern) {
				n.Del(key)
			}
		}

	case *patch.OpClear:
		n.Clear()
	case *patch.OpDel:
		n.Del(op.Key)
	case *patch.OpExpire:
		n.ExpireAt(op.Key, time.UnixMilli(op.Expiration))
	case *patch.OpExpireAt:
		n.ExpireAt(op.Key, time.UnixMilli(op.Expiration))
	case *patch.OpHClear:
		n.HClear(op.Key)
	case *patch.OpHDel:
		n.HDel(op.Key, op.Fields...)
	case *patch.OpHIncrBy:
		_, err := n.HIncrBy(op.Key, op.Field, op.IncrInt)
		return err
	case *patch.OpHIncrByFloat:
		_, err := n.HIncrByFloat(op.Key, op.Field, op.IncrFloat)
		return err
	case *patch.OpHMSet:
		if len(op.Fields) != len(op.Values) {
			return patch.ErrInvalidFrame
		}
		for i, field := range op.Fields {
			n.HSet(op.Key, field, op.Values[i])
		}
	case *patch.OpPersist:
		n.Persist(op.Key)
	case *patch.OpZUnionStore:
		n.ZUnionStore(op.Key, op.Keys, op.Weights, op.Aggregate)
	case *patch.OpZInterStore:
		n.ZInterStore(op.Key, op.Keys, op.Weights, op.Aggregate)
	case *patch.OpRenameNX:
		return n.RenameNX(op.Key, op.DstKey)
	case *patch.OpHSet:
		n.HSet(op.Key, op.Field, op.Value)
	case *patch.OpLInsert:
		n.LInsert(op.Key, op.Pivot, op.Value, op.Before)
	case *patch.OpLPop:
		n.LPop(op.Key, op.Count)
	case *patch.OpLPopRPush:
		n.LPopRPush(op.Key, op.DstKey)
	case *patch.OpLPush:
		n.LPush(op.Key, op.Values...)
	case *patch.OpLPushX:
		n.LPushX(op.Key, op.Value)
	case *patch.OpLRem:
		n.LRem(op.Key, op.Value, op.Count)
	case *patch.OpLSet:
		n.LSet(op.Key, op.Index, op.Value)
	case *patch.OpLTrim:
		n.LTrim(op.Key, op.Start, op.Stop)
	case *patch.OpRPop:
		n.RPop(op.Key, op.Count)
	case *patch.OpRPopLPush:
		n.RPopLPush(op.Key, op.DstKey)
	case *patch.OpRPush:
		n.RPush(op.Key, op.Values...)
	case *patch.OpRPushX:
		n.RPushX(op.Key, op.Value)
	case *patch.OpSAdd:
		n.SAdd(op.Key, op.Members...)
	case *patch.OpSRem:
		n.SRem(op.Key, op.Members...)
	case *patch.OpSet:
		n.Set(op.Key, op.Value, op.KeepTTL)
		if op.Expiration != 0 {
			n.ExpireAt(op.Key, time.UnixMilli(op.Expiration))
		}
	case *patch.OpZAdd:
		n.ZAdd(op.Key, op.Member, op.Score)
	case *patch.OpZClear:
		n.ZClear(op.Key)
	case *patch.OpZIncrBy:
		n.ZIncrBy(op.Key, op.Member, op.Score)
	case *patch.OpZRem:
		if len(op.Members) > 0 {
			n.ZRem(op.Key, op.Members...)
		} else {
			n.ZRem(op.Key, op.Member)
		}
	case *patch.OpZRemRangeByRank:
		n.ZRemRangeByRank(op.Key, op.Start, op.Stop)
	case *patch.OpZRemRangeByScore:
		n.ZRemRangeByScore(op.Key, op.Min, op.Max, int(op.Mode))
	case *patch.OpRename:
		return n.Rename(op.Key, op.DstKey)
	default:
		return ErrUnknownOperation
	}
	return nil
}

// Broadcast subscribes and snapshots under the same execution lock. A reconnect
// replaces matching replica keys before receiving subsequent committed batches.
func (n *Nodis) Broadcast(addr string, pattern []string) error {
	return n.options.Channel.Publish(addr, func(c ChannelConn) {
		id := -1
		defer func() {
			n.UnWatchKey(id)
			_ = c.Close()
			if p := recover(); p != nil {
				log.Println("nodis stream snapshot:", p)
			}
		}()
		err := n.exec(func(tx *Tx) error {
			n.listenersMu.Lock()
			defer n.listenersMu.Unlock()
			id = n.nextListenerID
			n.nextListenerID++
			w := listener.New(pattern, func(op patch.Op) {
				if err := c.Send(op); err != nil {
					_ = c.Close()
				}
			})
			w.Batch = true
			w.OnStop = c.Close
			registered := false
			defer func() {
				if !registered {
					w.Stop()
				}
			}()
			frames := [][]byte{(patch.Op{Type: patch.OpTypeReset, Data: &patch.OpReset{Patterns: pattern}}).Encode()}
			for key := range n.store.metadata {
				if !w.Matched(key) {
					continue
				}
				m := tx.readKey(key)
				if !m.isOk() {
					continue
				}
				frames = append(frames, (patch.Op{Type: patch.OpTypeRestore, Data: &patch.OpRestore{Key: key, Expiration: m.key.Expiration, Value: storage.EncodeValue(m.value)}}).Encode())
			}
			w.Push(patch.Op{Type: patch.OpTypeBatch, Data: &patch.OpBatch{Frames: frames}})
			n.listeners[id] = w
			n.listenerCount.Add(1)
			registered = true
			return nil
		})
		if err != nil {
			_ = c.Close()
			return
		}
		defer n.UnWatchKey(id)
		_ = c.Wait()
	})
}
func (n *Nodis) Subscribe(addr string) error {
	return n.options.Channel.Subscribe(addr, func(op patch.Op) error { return n.ApplyPatch(op) })
}
func (n *Nodis) Serve(addr string) error {
	server := redis.NewServer(func(conn *redis.Conn, cmd redis.Command) {
		if conn.OnClose == nil {
			conn.OnClose = func() { n.UnWatch(conn) }
		}
		GetCommand(cmd.Name)(n, conn, cmd)
	})
	n.serverMu.Lock()
	select {
	case <-n.done:
		n.serverMu.Unlock()
		return ErrClosed
	default:
	}
	if n.server != nil {
		n.serverMu.Unlock()
		return errors.New("nodis: Redis server already started")
	}
	n.server = server
	n.serverMu.Unlock()
	return server.Serve(addr)
}
func (n *Nodis) inTx(tx *Tx) Nodis { return Nodis{instance: n.instance, tx: tx} }
func (n *Nodis) exec(fn func(*Tx) error) (err error) {
	if n.tx != nil {
		return fn(n.tx)
	}
	s := n.store
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return ErrClosed
	}
	tx := &s.tx
	tx.err = nil
	tx.suppressEvents = false
	tx.cleared = false
	defer func() { defer func() { clear(s.changed); clear(s.events); s.events = s.events[:0] }(); n.publish(tx) }()
	err = fn(tx)
	if err == nil {
		err = tx.err
	}
	return err
}

func (n *Nodis) publish(tx *Tx) {
	if n.listenerCount.Load() == 0 {
		return
	}
	s := n.store
	n.listenersMu.Lock()
	defer n.listenersMu.Unlock()
	for id, w := range n.listeners {
		func() {
			defer func() {
				if p := recover(); p != nil {
					log.Println("nodis stream serialization:", p)
					w.Stop()
					delete(n.listeners, id)
					n.listenerCount.Add(-1)
				}
			}()
			if w.Batch {
				if tx.suppressEvents {
					return
				}
				var frames [][]byte
				if tx.cleared {
					frames = append(frames, (patch.Op{Type: patch.OpTypeReset, Data: &patch.OpReset{Patterns: w.Patterns()}}).Encode())
				}
				for key := range s.changed {
					if !w.Matched(key) {
						continue
					}
					m := s.metadata[key]
					if !m.isOk() || m.expired(time.Now().UnixMilli()) {
						frames = append(frames, (patch.Op{Type: patch.OpTypeDel, Data: &patch.OpDel{Key: key}}).Encode())
					} else {
						frames = append(frames, (patch.Op{Type: patch.OpTypeRestore, Data: &patch.OpRestore{Key: key, Expiration: m.key.Expiration, Value: storage.EncodeValue(m.value)}}).Encode())
					}
				}
				if len(frames) > 0 {
					w.Push(patch.Op{Type: patch.OpTypeBatch, Data: &patch.OpBatch{Frames: frames}})
				}
			} else {
				for _, op := range s.events {
					if op.Type == patch.OpTypeClear || w.Matched(op.Data.GetKey()) {
						w.Push(op)
					}
				}
			}
		}()
	}
}
