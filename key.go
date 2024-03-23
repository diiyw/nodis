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
	lastUse    atomic.Uint32
	Type       ds.DataType
	changed    atomic.Bool
}

func newKey(typ ds.DataType, seconds int64) *Key {
	k := &Key{Type: typ}
	if seconds != 0 {
		k.Expiration = seconds + time.Now().Unix()
	}
	k.changed.Store(true)
	return k
}

func (k *Key) expired() bool {
	if k == nil {
		return false
	}
	return k.Expiration != 0 && k.Expiration <= time.Now().Unix()
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
func (n *Nodis) Del(key string) {
	_, d := n.getDs(key, nil, 0)
	d.Lock()
	n.Lock()
	n.dataStructs.Delete(key)
	n.keys.Delete(key)
	n.store.remove(key)
	n.notify(pb.NewOp(pb.OpType_Del, key))
	n.Unlock()
	d.Unlock()
}

func (n *Nodis) Exists(key string) bool {
	n.RLock()
	_, ok := n.exists(key)
	n.RUnlock()
	return ok
}

// exists checks if a key exists
func (n *Nodis) exists(key string) (k *Key, ok bool) {
	k, ok = n.getKey(key)
	if !ok {
		// try get from store
		v, err := n.store.get(key)
		if err == nil && len(v) > 0 {
			key, d, ttl, err := parseDs(v)
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
func (n *Nodis) Expire(key string, seconds int64) {
	n.Lock()
	k, ok := n.getKey(key)
	if !ok {
		n.Unlock()
		return
	}
	k.Expiration += seconds
	k.changed.Store(true)
	n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(k.Expiration))
	n.Unlock()
}

// ExpireAt the keys
func (n *Nodis) ExpireAt(key string, timestamp time.Time) {
	n.Lock()
	k, ok := n.getKey(key)
	if !ok {
		n.Unlock()
		return
	}
	k.Expiration = timestamp.Unix()
	k.changed.Store(true)
	n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(k.Expiration))
	n.Unlock()
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
		return -1
	}
	n.RUnlock()
	return time.Until(time.Unix(k.Expiration, 0))
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
		_, d, _, err := parseDs(v)
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
func (n *Nodis) Scan(cursor int, match string, count int) (int, []string) {
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
	if cursor >= len(keys) {
		return 0, nil
	}
	if count > len(keys) {
		count = len(keys)
	}
	if cursor+count > len(keys) {
		count = len(keys) - cursor
	}
	return cursor + count, keys[cursor : cursor+count]
}
