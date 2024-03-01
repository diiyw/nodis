package nodis

import (
	"github.com/diiyw/nodis/ds/hash"
)

func (n *Nodis) newHash(key string) *hash.HashMap {
	h := hash.NewHashMap()
	n.store.Put(key, h)
	n.keys.Put(key, newKey("hash", 0))
	return h
}

func (n *Nodis) HSet(key string, field string, value []byte) {
	h := n.getDs(key)
	if h == nil {
		h = n.newHash(key)
	}
	h.(*hash.HashMap).HSet(field, value)
}

func (n *Nodis) HGet(key string, field string) ([]byte, bool) {
	h := n.getDs(key)
	if h == nil {
		return nil, false
	}
	return h.(*hash.HashMap).HGet(field)
}

func (n *Nodis) HDel(key string, field string) {
	h := n.getDs(key)
	if h == nil {
		return
	}
	h.(*hash.HashMap).HDel(field)
	if h.(*hash.HashMap).HLen() == 0 {
		n.store.Delete(key)
		n.keys.Delete(key)
	}
}

func (n *Nodis) HLen(key string) int {
	h := n.getDs(key)
	if h == nil {
		return 0
	}
	return h.(*hash.HashMap).HLen()
}

func (n *Nodis) HKeys(key string) []string {
	h := n.getDs(key)
	if h == nil {
		return nil
	}
	return h.(*hash.HashMap).HKeys()
}

func (n *Nodis) HExists(key string, field string) bool {
	h := n.getDs(key)
	if h == nil {
		return false
	}
	return h.(*hash.HashMap).HExists(field)
}

func (n *Nodis) HGetAll(key string) map[string][]byte {
	h := n.getDs(key)
	if h == nil {
		return nil
	}
	return h.(*hash.HashMap).HGetAll()
}

func (n *Nodis) HIncrBy(key string, field string, value int64) int64 {
	h := n.getDs(key)
	if h == nil {
		h = n.newHash(key)
	}
	return h.(*hash.HashMap).HIncrBy(field, value)
}

func (n *Nodis) HIncrByFloat(key string, field string, value float64) float64 {
	h := n.getDs(key)
	if h == nil {
		h = n.newHash(key)
	}
	return h.(*hash.HashMap).HIncrByFloat(field, value)
}

func (n *Nodis) HSetNX(key string, field string, value []byte) bool {
	h := n.getDs(key)
	if h != nil {
		return false
	}
	n.HSet(key, field, value)
	return true
}

func (n *Nodis) HMSet(key string, fields map[string][]byte) {
	n.Lock()
	defer n.Unlock()
	n.exists(key)
	h, ok := n.store.Get(key)
	if !ok {
		h = n.newHash(key)
	}
	h.(*hash.HashMap).HMSet(fields)
}

func (n *Nodis) HMGet(key string, fields ...string) [][]byte {
	h := n.getDs(key)
	if h == nil {
		return nil
	}
	return h.(*hash.HashMap).HMGet(fields...)
}

func (n *Nodis) HClear(key string) {
	n.Lock()
	defer n.Unlock()
	if !n.exists(key) {
		return
	}
	n.store.Delete(key)
	n.keys.Delete(key)
}

func (n *Nodis) HScan(key string, cursor int, match string, count int) (int, map[string][]byte) {
	h := n.getDs(key)
	if h == nil {
		return 0, nil
	}
	return h.(*hash.HashMap).HScan(cursor, match, count)
}

func (n *Nodis) HVals(key string) [][]byte {
	h := n.getDs(key)
	if h == nil {
		return nil
	}
	return h.(*hash.HashMap).HVals()
}
