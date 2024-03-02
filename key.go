package nodis

import (
	"errors"
	"path/filepath"
	"time"
)

type Key struct {
	Type string
	TTL  int64
}

func newKey(typ string, ttl int64) *Key {
	k := &Key{Type: typ}
	if ttl != 0 {
		k.TTL = ttl + time.Now().Unix()
	}
	return k
}

func (k *Key) expired() bool {
	if k == nil {
		return false
	}
	return k.TTL != 0 && k.TTL <= time.Now().Unix()
}

// Del a key
func (n *Nodis) Del(key string) {
	n.Lock()
	ds := n.getDs(key, nil, 0)
	ds.Lock()
	n.store.Delete(key)
	n.keys.Delete(key)
	ds.Unlock()
	n.Unlock()
}

func (n *Nodis) Exists(key string) bool {
	n.RLock()
	ok := n.exists(key)
	n.RUnlock()
	return ok
}

// exists checks if a key exists
func (n *Nodis) exists(key string) bool {
	_, ok := n.getKey(key)
	return ok
}

// Expire the keys
func (n *Nodis) Expire(key string, seconds int64) {
	n.Lock()
	k, ok := n.getKey(key)
	if !ok {
		n.Unlock()
		return
	}
	k.TTL += seconds
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
	k.TTL = timestamp.Unix()
	n.Unlock()
}

// Keys gets the keys
func (n *Nodis) Keys(pattern string) []string {
	n.RLock()
	keys := make([]string, 0, n.keys.Count())
	n.keys.Iter(func(key string, k *Key) bool {
		matched, _ := filepath.Match(pattern, key)
		if matched && !k.expired() {
			keys = append(keys, key)
		}
		return true
	})
	n.RUnlock()
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
	return time.Duration(k.TTL - time.Now().Unix())
}

// Rename a key
func (n *Nodis) Rename(key, newkey string) error {
	n.Lock()
	k, ok := n.getKey(newkey)
	if ok {
		n.Unlock()
		return errors.New("newkey exists")
	}
	v, ok := n.store.Get(key)
	if !ok {
		n.Unlock()
		return errors.New("key does not exist")
	}
	n.store.Delete(key)
	n.store.Put(newkey, v)
	n.keys.Delete(key)
	n.keys.Put(newkey, k)
	n.Unlock()
	return nil
}

// Type gets the type of a key
func (n *Nodis) Type(key string) string {
	n.RLock()
	k, ok := n.getKey(key)
	if !ok {
		n.RUnlock()
		return "none"
	}
	n.RUnlock()
	return k.Type
}

func (n *Nodis) getKey(key string) (*Key, bool) {
	k, ok := n.keys.Get(key)
	if !ok {
		n.store.Delete(key)
	}
	if k.expired() {
		n.keys.Delete(key)
		n.store.Delete(key)
		ok = false
	}
	return k, ok
}
