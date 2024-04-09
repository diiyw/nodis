package nodis

import (
	"errors"
	"path/filepath"
	"sync"
	"time"

	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/pb"
)

type Key struct {
	sync.RWMutex
	Expiration int64
	lastUse    int64
	Type       ds.DataType
	changed    bool
}

func newKey(typ ds.DataType) *Key {
	k := &Key{Type: typ}
	k.changed = true
	return k
}

func (k *Key) expired() bool {
	if k == nil {
		return false
	}
	return k.Expiration != 0 && k.Expiration <= time.Now().UnixMilli()
}

// Del a key
func (n *Nodis) Del(keys ...string) int64 {
	var c int64 = 0
	for _, key := range keys {
		tx := n.writeKey(key, nil)
		if !tx.isOk() {
			continue
		}
		n.delKey(key)
		c++
		tx.commit()
	}
	return c
}

func (n *Nodis) Exists(keys ...string) int64 {
	var num int64
	for _, key := range keys {
		tx := n.readKey(key)
		if tx.isOk() {
			num++
		}
		tx.commit()
	}
	return num
}

// Expire the keys
func (n *Nodis) Expire(key string, seconds int64) int64 {
	if seconds == 0 {
		return n.Del(key)
	}
	tx := n.writeKey(key, nil)
	if !tx.isOk() {
		return 0
	}
	if tx.key.Expiration == 0 {
		tx.key.Expiration = time.Now().UnixMilli()
	}
	tx.key.Expiration += seconds * 1000
	tx.commit()
	n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(tx.key.Expiration))
	return 1
}

// ExpirePX the keys in milliseconds
func (n *Nodis) ExpirePX(key string, milliseconds int64) int64 {
	if milliseconds == 0 {
		return n.Del(key)
	}
	tx := n.writeKey(key, nil)
	if !tx.isOk() {
		return 0
	}
	if tx.key.Expiration == 0 {
		tx.key.Expiration = time.Now().UnixMilli()
	}
	tx.key.Expiration += milliseconds
	tx.commit()
	n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(tx.key.Expiration))
	return 1
}

// ExpireNX the keys only when the key has no expiry
func (n *Nodis) ExpireNX(key string, seconds int64) int64 {
	tx := n.writeKey(key, nil)
	if !tx.isOk() {
		return 0
	}
	if tx.key.Expiration != 0 {
		return 0
	}
	tx.key.Expiration = time.Now().UnixMilli() + seconds*1000
	tx.commit()
	n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(tx.key.Expiration))
	return 1
}

// ExpireXX the keys only when the key has an existing expiry
func (n *Nodis) ExpireXX(key string, seconds int64) int64 {
	tx := n.writeKey(key, nil)
	if !tx.isOk() {
		return 0
	}
	if tx.key.Expiration == 0 {
		tx.commit()
		return 0
	}
	tx.key.Expiration += seconds * 1000
	tx.commit()
	n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(tx.key.Expiration))
	return 1
}

// ExpireLT the keys only when the new expiry is less than current one
func (n *Nodis) ExpireLT(key string, seconds int64) int64 {
	tx := n.writeKey(key, nil)
	if !tx.isOk() {
		return 0
	}
	if tx.key.Expiration == 0 {
		tx.commit()
		return 0
	}
	ms := seconds * 1000
	if tx.key.Expiration > time.Now().UnixMilli()-ms {
		tx.key.Expiration -= ms
		tx.commit()
		n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(tx.key.Expiration))
		return 1
	}
	tx.commit()
	return 0
}

// ExpireGT the keys only when the new expiry is greater than current one
func (n *Nodis) ExpireGT(key string, seconds int64) int64 {
	tx := n.writeKey(key, nil)
	if !tx.isOk() {
		return 0
	}
	now := time.Now().UnixMilli()
	if tx.key.Expiration == 0 {
		tx.key.Expiration = now
	}
	ms := seconds * 1000
	if tx.key.Expiration < now+ms {
		tx.key.Expiration += ms
		tx.commit()
		n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(tx.key.Expiration))
		return 1
	}
	tx.commit()
	return 0
}

// ExpireAt the keys
func (n *Nodis) ExpireAt(key string, timestamp time.Time) int64 {
	tx := n.writeKey(key, nil)
	if !tx.isOk() {
		return 0
	}
	tx.key.Expiration = timestamp.UnixMilli()
	tx.commit()
	n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(tx.key.Expiration))
	return 1
}

// ExpireAtNX the keys only when the key has no expiry
func (n *Nodis) ExpireAtNX(key string, timestamp time.Time) int64 {
	tx := n.writeKey(key, nil)
	if !tx.isOk() {
		return 0
	}
	if tx.key.Expiration != 0 {
		tx.commit()
		return 0
	}
	tx.key.Expiration = timestamp.UnixMilli()
	tx.commit()
	n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(tx.key.Expiration))
	return 1
}

// ExpireAtXX the keys only when the key has an existing expiry
func (n *Nodis) ExpireAtXX(key string, timestamp time.Time) int64 {
	tx := n.writeKey(key, nil)
	if !tx.isOk() {
		return 0
	}
	if tx.key.Expiration != 0 {
		tx.commit()
		return 0
	}
	tx.key.Expiration = timestamp.UnixMilli()
	tx.commit()
	n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(tx.key.Expiration))
	return 1
}

// ExpireAtLT the keys only when the new expiry is less than current one
func (n *Nodis) ExpireAtLT(key string, timestamp time.Time) int64 {
	tx := n.writeKey(key, nil)
	if !tx.isOk() {
		return 0
	}
	if tx.key.Expiration != 0 {
		tx.commit()
		return 0
	}
	unix := timestamp.UnixMilli()
	if tx.key.Expiration > unix {
		tx.key.Expiration = unix
		tx.commit()
		n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(tx.key.Expiration))
		return 1
	}
	return 0
}

// ExpireAtGT the keys only when the new expiry is greater than current one
func (n *Nodis) ExpireAtGT(key string, timestamp time.Time) int64 {
	tx := n.writeKey(key, nil)
	if !tx.isOk() {
		return 0
	}
	unix := timestamp.UnixMilli()
	if tx.key.Expiration == 0 {
		tx.key.Expiration = unix
	}
	if tx.key.Expiration < unix {
		tx.key.Expiration = unix
		n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(tx.key.Expiration))
		return 1
	}
	return 0
}

// Keys gets the keys
func (n *Nodis) Keys(pattern string) []string {
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
	var keys []string
	for key := range keyMap {
		keys = append(keys, key)
	}
	return keys
}

// TTL gets the TTL
func (n *Nodis) TTL(key string) time.Duration {
	tx := n.writeKey(key, nil)
	if !tx.isOk() {
		return 0
	}
	if tx.key.Expiration == 0 {
		tx.commit()
		return 0
	}
	s := tx.key.Expiration / 1000
	ns := (tx.key.Expiration - s*1000) * 1000 * 1000
	tx.commit()
	return time.Until(time.Unix(s, ns))
}

// Rename a key
func (n *Nodis) Rename(key, key2 string) error {
	tx := n.writeKey(key, nil)
	if !tx.isOk() {
		return errors.New("newKey exists")
	}
	v, ok := n.dataStructs.Get(key)
	if !ok {
		return errors.New("key does not exist")
	}
	tx2 := n.writeKey(key2, nil)
	n.delKey(key)
	n.dataStructs.Set(key2, v)
	n.keys.Set(key2, newKey(v.Type()))
	tx.commit()
	tx2.commit()
	n.notify(pb.NewOp(pb.OpType_Rename, key).DstKey(key2))
	return nil
}

// Type gets the type of key
func (n *Nodis) Type(key string) string {
	tx := n.readKey(key)
	if !tx.isOk() {
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
	return ds.DataTypeMap[tx.ds.Type()]
}

// Scan the keys
func (n *Nodis) Scan(cursor int64, match string, count int64) (int64, []string) {
	keys := make([]string, 0, n.keys.Len())
	n.keys.Scan(func(key string, k *Key) bool {
		matched, _ := filepath.Match(match, key)
		if matched && !k.expired() {
			keys = append(keys, key)
		}
		return true
	})
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
