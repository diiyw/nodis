package nodis

import (
	"errors"
	"log"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/pb"
)

type Key struct {
	Expiration int64
	lastUse    atomic.Int64
	Type       ds.DataType
	changed    atomic.Bool
}

func newKey(typ ds.DataType, ms int64) *Key {
	k := &Key{Type: typ}
	if ms != 0 {
		k.Expiration = ms + time.Now().UnixMilli()
	}
	k.changed.Store(true)
	return k
}

func (k *Key) expired() bool {
	if k == nil {
		return false
	}
	return k.Expiration != 0 && k.Expiration <= time.Now().UnixMilli()
}

func (n *Nodis) getKey(key string) (*Key, bool) {
	k, ok := n.keys.Get(key)
	if k.expired() {
		n.keys.Delete(key)
		n.dataStructs.Delete(key)
		n.store.remove(key)
		ok = false
	}
	return k, ok
}

// Del a key
func (n *Nodis) Del(keys ...string) int64 {
	var c int64 = 0
	for _, key := range keys {
		_, d := n.getDs(key, nil, 0)
		if d == nil {
			continue
		}
		d.Lock()
		n.Lock()
		n.dataStructs.Delete(key)
		n.keys.Delete(key)
		n.Unlock()
		n.store.remove(key)
		n.notify(pb.NewOp(pb.OpType_Del, key))
		d.Unlock()
		c++
	}
	return c
}

func (n *Nodis) Exists(keys ...string) int64 {
	n.RLock()
	var num int64
	for _, key := range keys {
		_, ok := n.exists(key)
		if ok {
			num++
		}
	}
	n.RUnlock()
	return num
}

// exists checks if a key exists
func (n *Nodis) exists(key string) (k *Key, ok bool) {
	k, ok = n.getKey(key)
	if !ok {
		// try get from store
		v, err := n.store.get(key)
		if err == nil && len(v) > 0 {
			key, d, ttl, err := n.parseDs(v)
			if err != nil {
				log.Println("Parse DataStruct:", err)
				return
			}
			if d != nil {
				n.dataStructs.Set(key, d)
				k = newKey(d.Type(), ttl)
				k.changed.Store(false)
				ok = true
				n.keys.Set(key, k)
				return
			}
		}
	}
	return
}

// Expire the keys
func (n *Nodis) Expire(key string, seconds int64) int64 {
	if seconds == 0 {
		return n.Del(key)
	}
	n.Lock()
	k, ok := n.getKey(key)
	if !ok {
		n.Unlock()
		return 0
	}
	if k.Expiration == 0 {
		k.Expiration = time.Now().UnixMilli()
	}
	k.Expiration += seconds * 1000
	k.changed.Store(true)
	n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(k.Expiration))
	n.Unlock()
	return 1
}

// ExpirePX the keys in milliseconds
func (n *Nodis) ExpirePX(key string, milliseconds int64) int64 {
	return n.Expire(key, milliseconds/1000)
}

// ExpireNX the keys only when the key has no expiry
func (n *Nodis) ExpireNX(key string, seconds int64) int64 {
	n.Lock()
	k, ok := n.getKey(key)
	if !ok {
		n.Unlock()
		return 0
	}
	if k.Expiration != 0 {
		n.Unlock()
		return 0
	}
	k.Expiration = time.Now().UnixMilli() + seconds*1000
	k.changed.Store(true)
	n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(k.Expiration))
	n.Unlock()
	return 1
}

// ExpireXX the keys only when the key has an existing expiry
func (n *Nodis) ExpireXX(key string, seconds int64) int64 {
	n.Lock()
	k, ok := n.getKey(key)
	if !ok {
		n.Unlock()
		return 0
	}
	if k.Expiration == 0 {
		n.Unlock()
		return 0
	}
	k.Expiration += seconds * 1000
	k.changed.Store(true)
	n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(k.Expiration))
	n.Unlock()
	return 1
}

// ExpireLT the keys only when the new expiry is less than current one
func (n *Nodis) ExpireLT(key string, seconds int64) int64 {
	n.Lock()
	k, ok := n.getKey(key)
	if !ok {
		n.Unlock()
		return 0
	}
	if k.Expiration == 0 {
		n.Unlock()
		return 0
	}
	ms := seconds * 1000
	if k.Expiration > time.Now().UnixMilli()-ms {
		k.Expiration -= ms
		k.changed.Store(true)
		n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(k.Expiration))
		n.Unlock()
		return 1
	}
	n.Unlock()
	return 0
}

// ExpireGT the keys only when the new expiry is greater than current one
func (n *Nodis) ExpireGT(key string, seconds int64) int64 {
	n.Lock()
	k, ok := n.getKey(key)
	if !ok {
		n.Unlock()
		return 0
	}
	now := time.Now().UnixMilli()
	if k.Expiration == 0 {
		k.Expiration = now
	}
	ms := seconds * 1000
	if k.Expiration < now+ms {
		k.Expiration += ms
		k.changed.Store(true)
		n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(k.Expiration))
		n.Unlock()
		return 1
	}
	n.Unlock()
	return 0
}

// ExpireAt the keys
func (n *Nodis) ExpireAt(key string, timestamp time.Time) int64 {
	n.Lock()
	k, ok := n.getKey(key)
	if !ok {
		n.Unlock()
		return 0
	}
	k.Expiration = timestamp.UnixMilli()
	k.changed.Store(true)
	n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(k.Expiration))
	n.Unlock()
	return 1
}

// ExpireAtNX the keys only when the key has no expiry
func (n *Nodis) ExpireAtNX(key string, timestamp time.Time) int64 {
	n.Lock()
	k, ok := n.getKey(key)
	if !ok {
		n.Unlock()
		return 0
	}
	if k.Expiration != 0 {
		n.Unlock()
		return 0
	}
	k.Expiration = timestamp.UnixMilli()
	k.changed.Store(true)
	n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(k.Expiration))
	n.Unlock()
	return 1
}

// ExpireAtXX the keys only when the key has an existing expiry
func (n *Nodis) ExpireAtXX(key string, timestamp time.Time) int64 {
	n.Lock()
	k, ok := n.getKey(key)
	if !ok {
		n.Unlock()
		return 0
	}
	if k.Expiration == 0 {
		n.Unlock()
		return 0
	}
	k.Expiration = timestamp.UnixMilli()
	k.changed.Store(true)
	n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(k.Expiration))
	n.Unlock()
	return 1
}

// ExpireAtLT the keys only when the new expiry is less than current one
func (n *Nodis) ExpireAtLT(key string, timestamp time.Time) int64 {
	n.Lock()
	k, ok := n.getKey(key)
	if !ok {
		n.Unlock()
		return 0
	}
	if k.Expiration == 0 {
		n.Unlock()
		return 0
	}
	unix := timestamp.UnixMilli()
	if k.Expiration > unix {
		k.Expiration = unix
		k.changed.Store(true)
		n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(k.Expiration))
		n.Unlock()
		return 1
	}
	n.Unlock()
	return 0
}

// ExpireAtGT the keys only when the new expiry is greater than current one
func (n *Nodis) ExpireAtGT(key string, timestamp time.Time) int64 {
	n.Lock()
	k, ok := n.getKey(key)
	if !ok {
		n.Unlock()
		return 0
	}
	unix := timestamp.UnixMilli()
	if k.Expiration == 0 {
		k.Expiration = unix
	}
	if k.Expiration < unix {
		k.Expiration = unix
		k.changed.Store(true)
		n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(k.Expiration))
		n.Unlock()
		return 1
	}
	n.Unlock()
	return 0
}

// Keys gets the keys
func (n *Nodis) Keys(pattern string) []string {
	n.RLock()
	keyMap := make(map[string]struct{})
	n.keys.Scan(func(key string, k *Key) bool {
		matched, _ := filepath.Match(pattern, key)
		if matched && !k.expired() {
			keyMap[key] = struct{}{}
		}
		return true
	})
	n.store.RLock()
	n.store.index.Scan(func(key string, _ *index) bool {
		matched, _ := filepath.Match(pattern, key)
		if matched {
			keyMap[key] = struct{}{}
		}
		return true
	})
	n.store.RUnlock()
	n.RUnlock()
	var keys []string
	for key := range keyMap {
		keys = append(keys, key)
	}
	return keys
}

// TTL gets the TTL
func (n *Nodis) TTL(key string) time.Duration {
	n.RLock()
	k, ok := n.getKey(key)
	if !ok {
		n.RUnlock()
		return 0
	}
	n.RUnlock()
	if k.Expiration == 0 {
		return 0
	}
	s := k.Expiration / 1000
	ns := (k.Expiration - s*1000) * 1000 * 1000
	return time.Until(time.Unix(k.Expiration/1000, ns))
}

// Rename a key
func (n *Nodis) Rename(key, key2 string) error {
	n.Lock()
	_, ok := n.getKey(key2)
	if ok {
		n.Unlock()
		return errors.New("newKey exists")
	}
	v, ok := n.dataStructs.Get(key)
	if !ok {
		n.Unlock()
		return errors.New("key does not exist")
	}
	n.dataStructs.Delete(key)
	n.dataStructs.Set(key2, v)
	n.keys.Delete(key)
	n.keys.Set(key2, newKey(v.Type(), 0))
	n.store.remove(key)
	n.notify(
		pb.NewOp(pb.OpType_Rename, key).DstKey(key2),
	)
	n.Unlock()
	return nil
}

// Type gets the type of key
func (n *Nodis) Type(key string) string {
	n.RLock()
	k, ok := n.getKey(key)
	if !ok {
		n.RUnlock()
		n.store.RLock()
		v, err := n.store.get(key)
		if err != nil || len(v) == 0 {
			n.store.RUnlock()
			return "none"
		}
		_, d, _, err := n.parseDs(v)
		if err != nil {
			n.store.RUnlock()
			return "none"
		}
		n.store.RUnlock()
		return ds.DataTypeMap[d.Type()]
	}
	n.RUnlock()
	return ds.DataTypeMap[k.Type]
}

// Scan the keys
func (n *Nodis) Scan(cursor int64, match string, count int64) (int64, []string) {
	n.RLock()
	keys := make([]string, 0, n.keys.Len())
	n.keys.Scan(func(key string, k *Key) bool {
		matched, _ := filepath.Match(match, key)
		if matched && !k.expired() {
			keys = append(keys, key)
		}
		return true
	})
	n.RUnlock()
	if len(keys) == 0 {
		return 0, nil
	}
	lk := int64(len(keys))
	if cursor >= lk {
		return 0, nil
	}
	if count > lk {
		count = lk
	}
	if cursor+count > lk {
		count = lk - cursor
	}
	if count == 0 {
		count = lk
	}
	return cursor + count, keys[cursor : cursor+count]
}
