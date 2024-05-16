package nodis

import (
	"errors"
	"math/rand"
	"path/filepath"
	"sync"
	"time"

	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/ds/list"
	"github.com/diiyw/nodis/pb"
	"github.com/diiyw/nodis/redis"
)

type Key struct {
	offset int64
	size   uint32
	fileId uint16
}

func newKey() *Key {
	return &Key{}
}

// Del a key
func (n *Nodis) Del(keys ...string) int64 {
	var c int64 = 0
	_ = n.exec(func(tx *Tx) error {
		for _, key := range keys {
			meta := tx.writeKey(key, nil)
			if !meta.isOk() {
				continue
			}
			tx.delKey(key)
			c++
		}
		return nil
	})
	return c
}

func (n *Nodis) Exists(keys ...string) int64 {
	var num int64
	_ = n.exec(func(tx *Tx) error {
		for _, key := range keys {
			meta := tx.readKey(key)
			if meta.isOk() {
				num++
			}
		}
		return nil
	})
	return num
}

// Expire the keys
func (n *Nodis) Expire(key string, seconds int64) int64 {
	if seconds == 0 {
		return n.Del(key)
	}
	var v int64 = 1
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, nil)
		if !meta.isOk() {
			v = 0
			return nil
		}
		meta.expiration = time.Now().UnixMilli()
		meta.expiration += seconds * 1000
		n.signalModifiedKey(key, meta)
		n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(meta.expiration))
		return nil
	})
	return v
}

// ExpirePX the keys in milliseconds
func (n *Nodis) ExpirePX(key string, milliseconds int64) int64 {
	if milliseconds == 0 {
		return n.Del(key)
	}
	var v int64 = 1
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, nil)
		if !meta.isOk() {
			v = 0
			return nil
		}
		if meta.expiration == 0 {
			meta.expiration = time.Now().UnixMilli()
		}
		meta.expiration += milliseconds
		n.signalModifiedKey(key, meta)
		n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(meta.expiration))
		return nil
	})
	return v
}

// ExpireNX the keys only when the key has no expiry
func (n *Nodis) ExpireNX(key string, seconds int64) int64 {
	var v int64 = 1
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, nil)
		if !meta.isOk() {
			v = 0
			return nil
		}
		if meta.expiration != 0 {
			v = 0
			return nil
		}
		meta.expiration = time.Now().UnixMilli() + seconds*1000
		n.signalModifiedKey(key, meta)
		n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(meta.expiration))
		return nil
	})
	return v
}

// ExpireXX the keys only when the key has an existing expiry
func (n *Nodis) ExpireXX(key string, seconds int64) int64 {
	var v int64 = 1
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, nil)
		if !meta.isOk() {
			v = 0
			return nil
		}
		if meta.expiration == 0 {
			v = 0
			return nil
		}
		meta.expiration += seconds * 1000
		n.signalModifiedKey(key, meta)
		n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(meta.expiration))
		return nil
	})
	return v
}

// ExpireLT the keys only when the new expiry is less than current one
func (n *Nodis) ExpireLT(key string, seconds int64) int64 {
	var v int64 = 1
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, nil)
		if !meta.isOk() {
			v = 0
			return nil
		}
		if meta.expiration == 0 {
			v = 0
			return nil
		}
		ms := seconds * 1000
		if meta.expiration > time.Now().UnixMilli()-ms {
			meta.expiration -= ms
			n.signalModifiedKey(key, meta)
			n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(meta.expiration))
		}
		return nil
	})
	return v
}

// ExpireGT the keys only when the new expiry is greater than current one
func (n *Nodis) ExpireGT(key string, seconds int64) int64 {
	var v int64 = 0
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, nil)
		if !meta.isOk() {
			return nil
		}
		now := time.Now().UnixMilli()
		if meta.expiration == 0 {
			meta.expiration = now
		}
		ms := seconds * 1000
		if meta.expiration < now+ms {
			meta.expiration += ms
			n.signalModifiedKey(key, meta)
			n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(meta.expiration))
			v = 1
		}
		return nil
	})
	return v
}

// ExpireAt the keys
func (n *Nodis) ExpireAt(key string, timestamp time.Time) int64 {
	var v int64 = 1
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, nil)
		if !meta.isOk() {
			v = 0
			return nil
		}
		meta.expiration = timestamp.UnixMilli()
		n.signalModifiedKey(key, meta)
		n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(meta.expiration))
		return nil
	})
	return v
}

// ExpireAtNX the keys only when the key has no expiry
func (n *Nodis) ExpireAtNX(key string, timestamp time.Time) int64 {
	var v int64 = 1
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, nil)
		if !meta.isOk() {
			v = 0
			return nil
		}
		if meta.expiration != 0 {
			v = 0
			return nil
		}
		meta.expiration = timestamp.UnixMilli()
		n.signalModifiedKey(key, meta)
		n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(meta.expiration))
		return nil
	})
	return v
}

// ExpireAtXX the keys only when the key has an existing expiry
func (n *Nodis) ExpireAtXX(key string, timestamp time.Time) int64 {
	var v int64 = 1
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, nil)
		if !meta.isOk() {
			v = 0
			return nil
		}
		if meta.expiration != 0 {
			v = 0
			return nil
		}
		meta.expiration = timestamp.UnixMilli()
		n.signalModifiedKey(key, meta)
		n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(meta.expiration))
		return nil
	})
	return v
}

// ExpireAtLT the keys only when the new expiry is less than current one
func (n *Nodis) ExpireAtLT(key string, timestamp time.Time) int64 {
	var v int64 = 1
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, nil)
		if !meta.isOk() {
			v = 0
			return nil
		}
		if meta.expiration != 0 {
			v = 0
			return nil
		}
		unix := timestamp.UnixMilli()
		if meta.expiration > unix {
			meta.expiration = unix
			n.signalModifiedKey(key, meta)
			n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(meta.expiration))
		}
		return nil
	})
	return v
}

// ExpireAtGT the keys only when the new expiry is greater than current one
func (n *Nodis) ExpireAtGT(key string, timestamp time.Time) int64 {
	var v int64 = 1
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, nil)
		if !meta.isOk() {
			v = 0
			return nil
		}
		unix := timestamp.UnixMilli()
		if meta.expiration == 0 {
			meta.expiration = unix
		}
		if meta.expiration < unix {
			meta.expiration = unix
			n.signalModifiedKey(key, meta)
			n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(meta.expiration))
		}
		return nil
	})
	return v
}

// Keys gets the keys
func (n *Nodis) Keys(pattern string) []string {
	var keys []string
	now := time.Now().UnixMilli()
	n.store.mu.RLock()
	n.store.metadata.Scan(func(key string, m *metadata) bool {
		matched, _ := filepath.Match(pattern, key)
		if matched && !m.expired(now) {
			keys = append(keys, key)
		}
		return true
	})
	n.store.mu.RUnlock()
	return keys
}

func (n *Nodis) Watch(rn *redis.Conn, keys ...string) {
	n.store.watchMu.Lock()
	for _, key := range keys {
		if _, ok := rn.WatchKeys.Get(key); !ok {
			rn.WatchKeys.Set(key, false)
		}
		clients, ok := n.store.watchedKeys.Get(key)
		if !ok {
			l := list.NewLinkedListG[*redis.Conn]()
			l.LPush(rn)
			n.store.watchedKeys.Set(key, l)
			continue
		}
		var found bool
		clients.ForRange(func(c *redis.Conn) bool {
			if c == rn {
				found = true
				return false
			}
			return true
		})
		if !found {
			clients.LPush(rn)
		}
	}
	n.store.watchMu.Unlock()
}

func (n *Nodis) UnWatch(rn *redis.Conn, keys ...string) {
	n.store.watchMu.Lock()
	for _, key := range keys {
		clients, ok := n.store.watchedKeys.Get(key)
		if !ok {
			continue
		}
		clients.ForRangeNode(func(ng *list.NodeG[*redis.Conn]) bool {
			if ng.Value() == rn {
				clients.RemoveNode(ng)
				return false
			}
			return true
		})
	}
	rn.WatchKeys.Clear()
	n.store.watchMu.Unlock()
}

// Keyspace gets the keyspace
func (n *Nodis) Keyspace() (keys int64, expires int64, avgTTL int64) {
	n.store.mu.Lock()
	now := time.Now().UnixMilli()
	n.store.metadata.Scan(func(key string, k *metadata) bool {
		if k.expired(now) {
			expires++
		}
		if k.expiration != 0 {
			avgTTL += k.expiration - now
		}
		keys++
		return true
	})
	if keys > 0 {
		avgTTL = avgTTL / keys / 1000
	}
	n.store.mu.Unlock()
	return
}

// TTL gets the TTL
func (n *Nodis) TTL(key string) time.Duration {
	var v time.Duration
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			v = -2
			return nil
		}
		if meta.expiration == 0 {
			v = -1
			return nil
		}
		s := meta.expiration / 1000
		ns := (meta.expiration - s*1000) * 1000 * 1000
		v = time.Until(time.Unix(s, ns)).Round(time.Second)
		return nil
	})
	return v
}

// PTTL gets the TTL in milliseconds
func (n *Nodis) PTTL(key string) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			v = -2
			return nil
		}
		if meta.expiration == 0 {
			v = -1
			return nil
		}
		v = meta.expiration - time.Now().UnixMilli()
		return nil
	})
	return v
}

// Rename a key
func (n *Nodis) Rename(key, dstKey string) error {
	return n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, nil)
		if !meta.isOk() {
			return errors.New("key not exists")
		}
		dstMeta := tx.writeKey(dstKey, nil)
		tx.delKey(key)
		if !dstMeta.isOk() {
			dstMeta.key = newKey()
			dstMeta.RWMutex = new(sync.RWMutex)
			dstMeta.expiration = meta.expiration
			n.store.mu.Lock()
			n.store.metadata.Set(dstKey, dstMeta)
			n.store.mu.Unlock()
		}
		dstMeta.setValue(meta.value)
		n.signalModifiedKey(key, meta)
		n.signalModifiedKey(key, dstMeta)
		n.notify(pb.NewOp(pb.OpType_Rename, key).DstKey(dstKey))
		return nil
	})
}

// RenameNX a key
func (n *Nodis) RenameNX(key, dstKey string) error {
	return n.exec(func(tx *Tx) error {
		dstMeta := tx.writeKey(dstKey, nil)
		if dstMeta.isOk() {
			return errors.New("newKey exists")
		}
		meta := tx.writeKey(key, nil)
		if !meta.isOk() {
			return errors.New("key does not exist")
		}
		tx.delKey(key)
		dstMeta.key = newKey()
		dstMeta.RWMutex = new(sync.RWMutex)
		dstMeta.expiration = meta.expiration
		dstMeta.setValue(meta.value)
		n.store.mu.Lock()
		n.store.metadata.Set(dstKey, dstMeta)
		n.store.mu.Unlock()
		n.signalModifiedKey(key, meta)
		n.signalModifiedKey(key, dstMeta)
		n.notify(pb.NewOp(pb.OpType_Rename, key).DstKey(dstKey))
		return nil
	})
}

// Type gets the type of key
func (n *Nodis) Type(key string) string {
	var v string
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			v = "none"
			return nil
		}
		v = meta.value.Type().String()
		return nil
	})
	return v
}

// Scan the keys
func (n *Nodis) Scan(cursor int64, match string, count int64, typ ds.ValueType) (int64, []string) {
	keyLen := int64(n.store.metadata.Len())
	if keyLen == 0 {
		return 0, nil
	}
	if cursor >= keyLen {
		return 0, nil
	}
	keys := make([]string, 0)
	now := time.Now().UnixMilli()
	tx := newTx(n.store)
	var iterCursor int64 = 0
	n.store.metadata.Scan(func(key string, m *metadata) bool {
		iterCursor++
		if cursor--; cursor > 0 {
			return true
		}
		if iterCursor > keyLen {
			iterCursor = 0
			return false
		}
		if count == 0 {
			return false
		}
		count--
		meta := tx.rLockKey(key)
		defer meta.commit()
		matched, _ := filepath.Match(match, key)
		if matched && !m.expired(now) {
			if typ != 0 && m.valueType != typ {
				return true
			}
			keys = append(keys, key)
		}
		return true
	})
	return iterCursor, keys
}

// RandomKey gets a random key
func (n *Nodis) RandomKey() string {
	now := time.Now().UnixMilli()
	var keys []string
	n.store.mu.RLock()
	n.store.metadata.Scan(func(key string, m *metadata) bool {
		if !m.expired(now) {
			keys = append(keys, key)
		}
		return true
	})
	n.store.mu.RUnlock()
	if len(keys) == 0 {
		return ""
	}
	return keys[rand.Intn(len(keys))]
}

// Persist the key
func (n *Nodis) Persist(key string) int64 {
	var v int64 = 0
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, nil)
		if !meta.isOk() {
			return nil
		}
		if meta.expiration == 0 {
			return nil
		}
		v = 1
		meta.expiration = 0
		n.signalModifiedKey(key, meta)
		n.notify(pb.NewOp(pb.OpType_Persist, key))
		return nil
	})
	return v
}

func (n *Nodis) signalModifiedKey(key string, meta *metadata) {
	meta.state |= KeyStateModified
	n.store.watchMu.RLock()
	clients, ok := n.store.watchedKeys.Get(key)
	if ok {
		clients.ForRange(func(c *redis.Conn) bool {
			c.WatchKeys.Set(key, true)
			return true
		})
	}
	n.store.watchMu.RUnlock()
}
