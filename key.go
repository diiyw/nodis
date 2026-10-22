package nodis

import (
	"errors"
	"github.com/diiyw/nodis/internal/glob"
	"math/rand"
	"sort"
	"time"

	"github.com/diiyw/nodis/ds"
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
			n.notify(func() []patch.Op { return []patch.Op{{Type: patch.OpTypeDel, Data: &patch.OpDel{Key: key}}} })
			c++
		}
		return nil
	})
	return c
}

func (n *Nodis) Unlink(keys ...string) int64 { return n.Del(keys...) }

func (n *Nodis) Exists(keys ...string) int64 {
	var num int64
	if n.tx == nil {
		s := n.store
		s.mu.RLock()
		if s.closed {
			s.mu.RUnlock()
			return 0
		}
		now := time.Now().UnixMilli()
		for _, key := range keys {
			m := s.metadata[key]
			if m.isOk() && !m.expired(now) {
				m.markAccessed()
				num++
			}
		}
		s.mu.RUnlock()
		return num
	}
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

// All expiration changes use an absolute Unix millisecond timestamp internally.
func (n *Nodis) expireAt(key string, expiration int64, mode string) int64 {
	var result int64
	_ = n.exec(func(tx *Tx) error {
		n := n.inTx(tx)
		m := tx.writeKey(key, nil)
		if !m.isOk() {
			return nil
		}
		old := m.key.Expiration
		if mode == "NX" && old != 0 || mode == "XX" && old == 0 || mode == "GT" && (old == 0 || expiration <= old) || mode == "LT" && old != 0 && expiration >= old {
			return nil
		}
		if expiration <= time.Now().UnixMilli() {
			returnValue := n.Del(key)
			result = returnValue
			return nil
		}
		m.key.Expiration = expiration
		n.signalModifiedKey(key, m)
		n.notify(func() []patch.Op {
			return []patch.Op{{Type: patch.OpTypeExpire, Data: &patch.OpExpire{Key: key, Expiration: expiration}}}
		})
		result = 1
		return nil
	})
	return result
}
func (n *Nodis) Expire(key string, seconds int64) int64 {
	return n.expireAt(key, time.Now().UnixMilli()+seconds*1000, "")
}
func (n *Nodis) ExpirePX(key string, ms int64) int64 {
	return n.expireAt(key, time.Now().UnixMilli()+ms, "")
}
func (n *Nodis) ExpireNX(key string, seconds int64) int64 {
	return n.expireAt(key, time.Now().UnixMilli()+seconds*1000, "NX")
}
func (n *Nodis) ExpireXX(key string, seconds int64) int64 {
	return n.expireAt(key, time.Now().UnixMilli()+seconds*1000, "XX")
}
func (n *Nodis) ExpireGT(key string, seconds int64) int64 {
	return n.expireAt(key, time.Now().UnixMilli()+seconds*1000, "GT")
}
func (n *Nodis) ExpireLT(key string, seconds int64) int64 {
	return n.expireAt(key, time.Now().UnixMilli()+seconds*1000, "LT")
}
func (n *Nodis) ExpireAt(key string, t time.Time) int64 { return n.expireAt(key, t.UnixMilli(), "") }
func (n *Nodis) ExpireAtNX(key string, t time.Time) int64 {
	return n.expireAt(key, t.UnixMilli(), "NX")
}
func (n *Nodis) ExpireAtXX(key string, t time.Time) int64 {
	return n.expireAt(key, t.UnixMilli(), "XX")
}
func (n *Nodis) ExpireAtGT(key string, t time.Time) int64 {
	return n.expireAt(key, t.UnixMilli(), "GT")
}
func (n *Nodis) ExpireAtLT(key string, t time.Time) int64 {
	return n.expireAt(key, t.UnixMilli(), "LT")
}

// Keys gets the keys
func (n *Nodis) Keys(pattern string) []string {
	var keys []string
	_ = n.exec(func(tx *Tx) error {
		now := time.Now().UnixMilli()
		for key, m := range n.store.metadata {
			if glob.Match(pattern, key) && !m.expired(now) {
				keys = append(keys, key)
			}
		}
		return nil
	})
	return keys
}
func (n *Nodis) Watch(c *redis.Conn, keys ...string) {
	_ = n.exec(func(tx *Tx) error {
		if c.WatchKeys == nil {
			c.WatchKeys = make(map[string]bool)
		}
		for _, key := range keys {
			if m := n.store.metadata[key]; m != nil && m.expired(time.Now().UnixMilli()) {
				tx.delKey(key)
			}
			if _, ok := c.WatchKeys[key]; !ok {
				c.WatchKeys[key] = false
			}
			clients := n.store.watchedKeys[key]
			if clients == nil {
				clients = make(map[*redis.Conn]struct{})
				n.store.watchedKeys[key] = clients
			}
			clients[c] = struct{}{}
		}
		return nil
	})
}
func (n *Nodis) UnWatch(c *redis.Conn, keys ...string) {
	_ = n.exec(func(tx *Tx) error {
		if len(keys) == 0 {
			for key := range c.WatchKeys {
				delete(n.store.watchedKeys[key], c)
				if len(n.store.watchedKeys[key]) == 0 {
					delete(n.store.watchedKeys, key)
				}
			}
			clear(c.WatchKeys)
		} else {
			for _, key := range keys {
				delete(n.store.watchedKeys[key], c)
				if len(n.store.watchedKeys[key]) == 0 {
					delete(n.store.watchedKeys, key)
				}
				delete(c.WatchKeys, key)
			}
		}
		return nil
	})
}
func (n *Nodis) Keyspace() (keys int64, expires int64, avgTTL int64) {
	_ = n.exec(func(tx *Tx) error {
		now := time.Now().UnixMilli()
		for _, m := range n.store.metadata {
			if m.expired(now) {
				continue
			}
			keys++
			if m.key.Expiration != 0 {
				expires++
				avgTTL += m.key.Expiration - now
			}
		}
		if expires > 0 {
			avgTTL = avgTTL / expires / 1000
		}
		return nil
	})
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
		v = time.Until(time.Unix(s, ns)).Truncate(time.Second)
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
func (n *Nodis) Rename(key, dstKey string) error   { return n.rename(key, dstKey, false) }
func (n *Nodis) RenameNX(key, dstKey string) error { return n.rename(key, dstKey, true) }
func (n *Nodis) rename(key, dstKey string, nx bool) error {
	return n.exec(func(tx *Tx) error {
		n := n.inTx(tx)
		src := tx.writeKey(key, nil)
		if !src.isOk() {
			return errors.New("key does not exist")
		}
		if key == dstKey {
			if nx {
				return errors.New("newKey exists")
			}
			return nil
		}
		dst := tx.writeKey(dstKey, nil)
		if nx && dst.isOk() {
			return errors.New("newKey exists")
		}
		// Validate/load both sides before changing either map entry.
		dst.key.Expiration = src.key.Expiration
		dst.setValue(src.value)
		dst.state |= KeyStateModified
		tx.delKey(key)
		tx.storeMeta(dst)
		n.signalModifiedKey(dstKey, dst)
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
func (n *Nodis) Scan(cursor int64, match string, count int64, typ ds.ValueType) (next int64, keys []string) {
	_ = n.exec(func(tx *Tx) error {
		if cursor < 0 || count <= 0 {
			return nil
		}
		names := make([]string, 0, len(n.store.metadata))
		for key := range n.store.metadata {
			names = append(names, key)
		}
		sort.Strings(names)
		if cursor >= int64(len(names)) {
			return nil
		}
		end := cursor + count
		if end < cursor || end > int64(len(names)) {
			end = int64(len(names))
		}
		for _, key := range names[cursor:end] {
			m := tx.readKey(key)
			if !m.isOk() {
				continue
			}
			if glob.Match(match, key) && (typ == 0 || m.valueType == typ) {
				keys = append(keys, key)
			}
		}
		if end < int64(len(names)) {
			next = end
		}
		return nil
	})
	return
}
func (n *Nodis) RandomKey() string {
	var result string
	_ = n.exec(func(tx *Tx) error {
		now := time.Now().UnixMilli()
		seen := 0
		for key, m := range n.store.metadata {
			if !m.expired(now) {
				seen++
				if rand.Intn(seen) == 0 {
					result = key
				}
			}
		}
		return nil
	})
	return result
}

// Persist the key
func (n *Nodis) Persist(key string) int64 {
	var v int64 = 0
	_ = n.exec(func(tx *Tx) error {
		n := n.inTx(tx)
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

func (n *Nodis) signalModifiedKey(key string, m *metadata) {
	m.state |= KeyStateModified
	n.store.dirty[key] = struct{}{}
	n.store.changed[key] = struct{}{}
	n.store.invalidate(key)
}
