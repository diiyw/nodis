package nodis

import (
	"encoding/binary"
	"errors"
	"path/filepath"
	"time"

	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/pb"
)

type Key struct {
	expiration int64
	lastUse    int64
	offset     int64
	size       uint32
	fileId     uint16
	changed    bool
}

func newKey() *Key {
	k := &Key{}
	k.changed = true
	return k
}

func (k *Key) expired(now int64) bool {
	if k == nil {
		return false
	}
	return k.expiration != 0 && k.expiration <= now
}

// marshal index to bytes
func (k *Key) marshal() []byte {
	var b [22]byte
	binary.LittleEndian.PutUint64(b[0:8], uint64(k.offset))
	binary.LittleEndian.PutUint64(b[8:16], uint64(k.expiration))
	binary.LittleEndian.PutUint32(b[16:20], k.size)
	binary.LittleEndian.PutUint16(b[20:22], k.fileId)
	return b[:]
}

// unmarshal bytes to index
func (k *Key) unmarshal(b []byte) {
	k.offset = int64(binary.LittleEndian.Uint64(b[0:8]))
	k.expiration = int64(binary.LittleEndian.Uint64(b[8:16]))
	k.size = binary.LittleEndian.Uint32(b[16:20])
	k.fileId = binary.LittleEndian.Uint16(b[20:22])
}

// Del a key
func (n *Nodis) Del(keys ...string) int64 {
	var c int64 = 0
	for _, key := range keys {
		meta := n.store.writeKey(key, nil)
		if !meta.isOk() {
			meta.commit()
			continue
		}
		n.store.delKey(key)
		c++
		meta.commit()
	}
	return c
}

func (n *Nodis) Exists(keys ...string) int64 {
	var num int64
	for _, key := range keys {
		meta := n.store.readKey(key)
		if meta.isOk() {
			num++
		}
		meta.commit()
	}
	return num
}

// Expire the keys
func (n *Nodis) Expire(key string, seconds int64) int64 {
	if seconds == 0 {
		return n.Del(key)
	}
	meta := n.store.writeKey(key, nil)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	if meta.key.expiration == 0 {
		meta.key.expiration = time.Now().UnixMilli()
	}
	meta.key.expiration += seconds * 1000
	n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(meta.key.expiration))
	meta.commit()
	return 1
}

// ExpirePX the keys in milliseconds
func (n *Nodis) ExpirePX(key string, milliseconds int64) int64 {
	if milliseconds == 0 {
		return n.Del(key)
	}
	meta := n.store.writeKey(key, nil)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	if meta.key.expiration == 0 {
		meta.key.expiration = time.Now().UnixMilli()
	}
	meta.key.expiration += milliseconds
	n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(meta.key.expiration))
	meta.commit()
	return 1
}

// ExpireNX the keys only when the key has no expiry
func (n *Nodis) ExpireNX(key string, seconds int64) int64 {
	meta := n.store.writeKey(key, nil)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	if meta.key.expiration != 0 {
		return 0
	}
	meta.key.expiration = time.Now().UnixMilli() + seconds*1000
	n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(meta.key.expiration))
	meta.commit()
	return 1
}

// ExpireXX the keys only when the key has an existing expiry
func (n *Nodis) ExpireXX(key string, seconds int64) int64 {
	meta := n.store.writeKey(key, nil)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	if meta.key.expiration == 0 {
		meta.commit()
		return 0
	}
	meta.key.expiration += seconds * 1000
	n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(meta.key.expiration))
	meta.commit()
	return 1
}

// ExpireLT the keys only when the new expiry is less than current one
func (n *Nodis) ExpireLT(key string, seconds int64) int64 {
	meta := n.store.writeKey(key, nil)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	if meta.key.expiration == 0 {
		meta.commit()
		return 0
	}
	ms := seconds * 1000
	if meta.key.expiration > time.Now().UnixMilli()-ms {
		meta.key.expiration -= ms
		meta.commit()
		n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(meta.key.expiration))
		return 1
	}
	meta.commit()
	return 0
}

// ExpireGT the keys only when the new expiry is greater than current one
func (n *Nodis) ExpireGT(key string, seconds int64) int64 {
	meta := n.store.writeKey(key, nil)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	now := time.Now().UnixMilli()
	if meta.key.expiration == 0 {
		meta.key.expiration = now
	}
	ms := seconds * 1000
	if meta.key.expiration < now+ms {
		meta.key.expiration += ms
		meta.commit()
		n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(meta.key.expiration))
		return 1
	}
	meta.commit()
	return 0
}

// ExpireAt the keys
func (n *Nodis) ExpireAt(key string, timestamp time.Time) int64 {
	meta := n.store.writeKey(key, nil)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	meta.key.expiration = timestamp.UnixMilli()
	n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(meta.key.expiration))
	meta.commit()
	return 1
}

// ExpireAtNX the keys only when the key has no expiry
func (n *Nodis) ExpireAtNX(key string, timestamp time.Time) int64 {
	meta := n.store.writeKey(key, nil)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	if meta.key.expiration != 0 {
		meta.commit()
		return 0
	}
	meta.key.expiration = timestamp.UnixMilli()
	n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(meta.key.expiration))
	meta.commit()
	return 1
}

// ExpireAtXX the keys only when the key has an existing expiry
func (n *Nodis) ExpireAtXX(key string, timestamp time.Time) int64 {
	meta := n.store.writeKey(key, nil)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	if meta.key.expiration != 0 {
		meta.commit()
		return 0
	}
	meta.key.expiration = timestamp.UnixMilli()
	n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(meta.key.expiration))
	meta.commit()
	return 1
}

// ExpireAtLT the keys only when the new expiry is less than current one
func (n *Nodis) ExpireAtLT(key string, timestamp time.Time) int64 {
	meta := n.store.writeKey(key, nil)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	if meta.key.expiration != 0 {
		meta.commit()
		return 0
	}
	unix := timestamp.UnixMilli()
	if meta.key.expiration > unix {
		meta.key.expiration = unix
		meta.commit()
		n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(meta.key.expiration))
		return 1
	}
	return 0
}

// ExpireAtGT the keys only when the new expiry is greater than current one
func (n *Nodis) ExpireAtGT(key string, timestamp time.Time) int64 {
	meta := n.store.writeKey(key, nil)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	unix := timestamp.UnixMilli()
	if meta.key.expiration == 0 {
		meta.key.expiration = unix
	}
	if meta.key.expiration < unix {
		meta.key.expiration = unix
		n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(meta.key.expiration))
		return 1
	}
	return 0
}

// Keys gets the keys
func (n *Nodis) Keys(pattern string) []string {
	var keys []string
	now := time.Now().UnixMilli()
	n.store.RLock()
	n.store.keys.Scan(func(key string, k *Key) bool {
		matched, _ := filepath.Match(pattern, key)
		if matched && !k.expired(now) {
			keys = append(keys, key)
		}
		return true
	})
	n.store.RUnlock()
	return keys
}

// TTL gets the TTL
func (n *Nodis) TTL(key string) time.Duration {
	meta := n.store.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	if meta.key.expiration == 0 {
		meta.commit()
		return 0
	}
	s := meta.key.expiration / 1000
	ns := (meta.key.expiration - s*1000) * 1000 * 1000
	meta.commit()
	return time.Until(time.Unix(s, ns))
}

// Rename a key
func (n *Nodis) Rename(key, key2 string) error {
	meta := n.store.writeKey(key, nil)
	if !meta.isOk() {
		meta.commit()
		return errors.New("newKey exists")
	}
	v, ok := n.store.values.Get(key)
	if !ok {
		return errors.New("key does not exist")
	}
	tx2 := n.store.writeKey(key2, nil)
	n.store.delKey(key)
	n.store.values.Set(key2, v)
	n.store.keys.Set(key2, newKey())
	meta.commit()
	tx2.commit()
	n.notify(pb.NewOp(pb.OpType_Rename, key).DstKey(key2))
	return nil
}

// Type gets the type of key
func (n *Nodis) Type(key string) string {
	meta := n.store.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return "none"
	}
	v := ds.DataTypeMap[meta.ds.Type()]
	meta.commit()
	return v
}

// Scan the keys
func (n *Nodis) Scan(cursor int64, match string, count int64) (int64, []string) {
	keys := make([]string, 0)
	now := time.Now().UnixMilli()
	n.store.keys.Scan(func(key string, k *Key) bool {
		matched, _ := filepath.Match(match, key)
		if matched && !k.expired(now) {
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
