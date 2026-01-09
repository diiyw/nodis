package nodis

import (
	"errors"
	"math/rand"
	"path/filepath"
	"runtime"
	"time"

	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/ds/list"
	"github.com/diiyw/nodis/redis"

	"github.com/diiyw/nodis/patch"
)

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

func (n *Nodis) Unlink(keys ...string) int64 {
	v := n.Del(keys...)
	runtime.GC()
	return v
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
		meta.key.Expiration = time.Now().UnixMilli() + seconds*1000
		n.signalModifiedKey(key, meta)
		n.notify(func() []patch.Op {
			return []patch.Op{{Type: patch.OpTypeExpire, Data: &patch.OpExpire{Key: key, Expiration: meta.key.Expiration}}}
		})
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
		if meta.key.Expiration == 0 {
			meta.key.Expiration = time.Now().UnixMilli()
		}
		meta.key.Expiration += milliseconds
		n.signalModifiedKey(key, meta)
		n.notify(func() []patch.Op {
			return []patch.Op{{Type: patch.OpTypeExpire, Data: &patch.OpExpire{Key: key, Expiration: meta.key.Expiration}}}
		})
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
		if meta.key.Expiration != 0 {
			v = 0
			return nil
		}
		meta.key.Expiration = time.Now().UnixMilli() + seconds*1000
		n.signalModifiedKey(key, meta)
		n.notify(func() []patch.Op {
			return []patch.Op{{Type: patch.OpTypeExpire, Data: &patch.OpExpire{Key: key, Expiration: meta.key.Expiration}}}
		})
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
		if meta.key.Expiration == 0 {
			v = 0
			return nil
		}
		meta.key.Expiration += seconds * 1000
		n.signalModifiedKey(key, meta)
		n.notify(func() []patch.Op {
			return []patch.Op{{Type: patch.OpTypeExpire, Data: &patch.OpExpire{Key: key, Expiration: meta.key.Expiration}}}
		})
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
		if meta.key.Expiration == 0 {
			v = 0
			return nil
		}
		ms := seconds * 1000
		if meta.key.Expiration > time.Now().UnixMilli()-ms {
			meta.key.Expiration -= ms
			n.signalModifiedKey(key, meta)
			n.notify(func() []patch.Op {
				return []patch.Op{{Type: patch.OpTypeExpire, Data: &patch.OpExpire{Key: key, Expiration: meta.key.Expiration}}}
			})
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
		if meta.key.Expiration == 0 {
			meta.key.Expiration = now
		}
		ms := seconds * 1000
		if meta.key.Expiration < now+ms {
			meta.key.Expiration += ms
			n.signalModifiedKey(key, meta)
			n.notify(func() []patch.Op {
				return []patch.Op{{Type: patch.OpTypeExpire, Data: &patch.OpExpire{Key: key, Expiration: meta.key.Expiration}}}
			})
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
		meta.key.Expiration = timestamp.UnixMilli()
		n.signalModifiedKey(key, meta)
		n.notify(func() []patch.Op {
			return []patch.Op{{Type: patch.OpTypeExpire, Data: &patch.OpExpire{Key: key, Expiration: meta.key.Expiration}}}
		})
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
		if meta.key.Expiration != 0 {
			v = 0
			return nil
		}
		meta.key.Expiration = timestamp.UnixMilli()
		n.signalModifiedKey(key, meta)
		n.notify(func() []patch.Op {
			return []patch.Op{{Type: patch.OpTypeExpire, Data: &patch.OpExpire{Key: key, Expiration: meta.key.Expiration}}}
		})
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
		if meta.key.Expiration == 0 {
			v = 0
			return nil
		}
		meta.key.Expiration = timestamp.UnixMilli()
		n.signalModifiedKey(key, meta)
		n.notify(func() []patch.Op {
			return []patch.Op{{Type: patch.OpTypeExpire, Data: &patch.OpExpire{Key: key, Expiration: meta.key.Expiration}}}
		})
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
		if meta.key.Expiration == 0 {
			v = 0
			return nil
		}
		unix := timestamp.UnixMilli()
		if unix < meta.key.Expiration {
			meta.key.Expiration = unix
			n.signalModifiedKey(key, meta)
			n.notify(func() []patch.Op {
				return []patch.Op{{Type: patch.OpTypeExpire, Data: &patch.OpExpire{Key: key, Expiration: meta.key.Expiration}}}
			})
		} else {
			v = 0
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
		if meta.key.Expiration == 0 {
			meta.key.Expiration = unix
		}
		if meta.key.Expiration < unix {
			meta.key.Expiration = unix
			n.signalModifiedKey(key, meta)
			n.notify(func() []patch.Op {
				return []patch.Op{{Type: patch.OpTypeExpire, Data: &patch.OpExpire{Key: key, Expiration: meta.key.Expiration}}}
			})
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
	for key, m := range n.store.metadata {
		matched, _ := filepath.Match(pattern, key)
		if matched && !m.expired(now) {
			keys = append(keys, key)
		}
	}
	n.store.mu.RUnlock()
	return keys
}

func (n *Nodis) Watch(rn *redis.Conn, keys ...string) {
	n.store.watchMu.Lock()
	for _, key := range keys {
		if _, ok := rn.WatchKeys[key]; !ok {
			rn.WatchKeys[key] = false
		}
		clients, ok := n.store.watchedKeys[key]
		if !ok {
			l := list.NewLinkedListG[*redis.Conn]()
			l.LPush(rn)
			n.store.watchedKeys[key] = l
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
		clients, ok := n.store.watchedKeys[key]
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
	clear(rn.WatchKeys)
	n.store.watchMu.Unlock()
}

// Keyspace gets the keyspace
func (n *Nodis) Keyspace() (keys int64, expires int64, avgTTL int64) {
	n.store.mu.Lock()
	now := time.Now().UnixMilli()
	for _, m := range n.store.metadata {
		if m.expired(now) {
			expires++
		}
		if m.key.Expiration != 0 {
			avgTTL += m.key.Expiration - now
		}
		keys++
	}
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
		if meta.key.Expiration == 0 {
			v = -1
			return nil
		}
		s := meta.key.Expiration / 1000
		ns := (meta.key.Expiration - s*1000) * 1000 * 1000
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
		if meta.key.Expiration == 0 {
			v = -1
			return nil
		}
		v = meta.key.Expiration - time.Now().UnixMilli()
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
			tx.lockMeta(dstMeta)
			dstMeta.key.Expiration = meta.key.Expiration
			dstMeta.state = meta.state
		}
		dstMeta.setValue(meta.value)
		tx.storeMeta(dstMeta)
		n.signalModifiedKey(key, meta)
		n.signalModifiedKey(key, dstMeta)
		n.notify(func() []patch.Op {
			return []patch.Op{{Type: patch.OpTypeRename, Data: &patch.OpRename{Key: key, DstKey: dstKey}}}
		})
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
		tx.lockMeta(dstMeta)
		dstMeta.key.Expiration = meta.key.Expiration
		dstMeta.state = meta.state
		dstMeta.setValue(meta.value)
		tx.storeMeta(dstMeta)
		n.signalModifiedKey(key, meta)
		n.signalModifiedKey(key, dstMeta)
		n.notify(func() []patch.Op {
			return []patch.Op{{Type: patch.OpTypeRename, Data: &patch.OpRename{Key: key, DstKey: dstKey}}}
		})
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
	keyLen := int64(len(n.store.metadata))
	if keyLen == 0 {
		return 0, nil
	}
	if cursor >= keyLen {
		return 0, nil
	}
	keys := make([]string, 0)
	now := time.Now().UnixMilli()
	tx := newTx(n.store)
	defer tx.commit()
	var iterCursor int64 = 0
	for key, m := range n.store.metadata {
		iterCursor++
		if cursor--; cursor > 0 {
			continue
		}
		if iterCursor > keyLen {
			iterCursor = 0
			break
		}
		if count == 0 {
			break
		}
		count--
		_ = tx.rLockKey(key)
		matched, _ := filepath.Match(match, key)
		if matched && !m.expired(now) {
			if typ != 0 && m.valueType != typ {
				continue
			}
			keys = append(keys, key)
		}
	}
	return iterCursor, keys
}

// RandomKey gets a random key
func (n *Nodis) RandomKey() string {
	now := time.Now().UnixMilli()
	var keys []string
	n.store.mu.RLock()
	for key, m := range n.store.metadata {
		if !m.expired(now) {
			keys = append(keys, key)
		}
	}
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
		if meta.key.Expiration == 0 {
			return nil
		}
		v = 1
		meta.key.Expiration = 0
		n.signalModifiedKey(key, meta)
		n.notify(func() []patch.Op {
			return []patch.Op{{Type: patch.OpTypePersist, Data: &patch.OpPersist{Key: key}}}
		})
		return nil
	})
	return v
}

func (n *Nodis) signalModifiedKey(key string, meta *metadata) {
	meta.state |= KeyStateModified
	n.store.watchMu.RLock()
	clients, ok := n.store.watchedKeys[key]
	if ok {
		clients.ForRange(func(c *redis.Conn) bool {
			c.WatchKeys[key] = true
			return true
		})
	}
	n.store.watchMu.RUnlock()
}
