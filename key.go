package nodis

import (
	"encoding/binary"
	"errors"
	"math/rand"
	"path/filepath"
	"time"

	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/pb"
)

const (
	KeyStateNone                uint8 = 0
	KeyStateModified            uint8 = 1
	KeyStateWatching            uint8 = 2
	KeyStateWatchBeforeModified uint8 = 4
	KeyStateWatchAfterModified  uint8 = 8
)

type Key struct {
	expiration   int64
	modifiedTime int64
	offset       int64
	size         uint32
	fileId       uint16
	dataType     ds.DataType
	state        uint8
}

func newKey() *Key {
	k := &Key{}
	k.state |= KeyStateModified
	return k
}

func (k *Key) expired(now int64) bool {
	if k == nil {
		return false
	}
	return k.expiration != 0 && k.expiration <= now
}

// modified return if the key is modified
func (k *Key) modified() bool {
	return k.state&KeyStateModified == KeyStateModified
}

func (k *Key) resetModified() {
	if k.state&KeyStateModified == KeyStateModified {
		k.state ^= KeyStateModified
		k.modifiedTime = 0
	}
}

// marshal index to bytes
func (k *Key) marshal() []byte {
	var b [23]byte
	binary.LittleEndian.PutUint64(b[0:8], uint64(k.offset))
	binary.LittleEndian.PutUint64(b[8:16], uint64(k.expiration))
	binary.LittleEndian.PutUint32(b[16:20], k.size)
	binary.LittleEndian.PutUint16(b[20:22], k.fileId)
	b[22] = uint8(k.dataType)
	return b[:]
}

// unmarshal bytes to index
func (k *Key) unmarshal(b []byte) {
	k.offset = int64(binary.LittleEndian.Uint64(b[0:8]))
	k.expiration = int64(binary.LittleEndian.Uint64(b[8:16]))
	k.size = binary.LittleEndian.Uint32(b[16:20])
	k.fileId = binary.LittleEndian.Uint16(b[20:22])
	k.dataType = ds.DataType(b[22])
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
		if meta.key.expiration == 0 {
			meta.key.expiration = time.Now().UnixMilli()
		}
		meta.key.expiration += seconds * 1000
		meta.signalModifiedKey()
		n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(meta.key.expiration))
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
		if meta.key.expiration == 0 {
			meta.key.expiration = time.Now().UnixMilli()
		}
		meta.key.expiration += milliseconds
		meta.signalModifiedKey()
		n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(meta.key.expiration))
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
		if meta.key.expiration != 0 {
			v = 0
			return nil
		}
		meta.key.expiration = time.Now().UnixMilli() + seconds*1000
		meta.signalModifiedKey()
		n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(meta.key.expiration))
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
		if meta.key.expiration == 0 {
			v = 0
			return nil
		}
		meta.key.expiration += seconds * 1000
		meta.signalModifiedKey()
		n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(meta.key.expiration))
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
		if meta.key.expiration == 0 {
			v = 0
			return nil
		}
		ms := seconds * 1000
		if meta.key.expiration > time.Now().UnixMilli()-ms {
			meta.key.expiration -= ms
			meta.signalModifiedKey()
			n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(meta.key.expiration))
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
		if meta.key.expiration == 0 {
			meta.key.expiration = now
		}
		ms := seconds * 1000
		if meta.key.expiration < now+ms {
			meta.key.expiration += ms
			meta.signalModifiedKey()
			n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(meta.key.expiration))
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
		meta.key.expiration = timestamp.UnixMilli()
		meta.signalModifiedKey()
		n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(meta.key.expiration))
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
		if meta.key.expiration != 0 {
			v = 0
			return nil
		}
		meta.key.expiration = timestamp.UnixMilli()
		meta.signalModifiedKey()
		n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(meta.key.expiration))
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
		if meta.key.expiration != 0 {
			v = 0
			return nil
		}
		meta.key.expiration = timestamp.UnixMilli()
		meta.signalModifiedKey()
		n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(meta.key.expiration))
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
		if meta.key.expiration != 0 {
			v = 0
			return nil
		}
		unix := timestamp.UnixMilli()
		if meta.key.expiration > unix {
			meta.key.expiration = unix
			meta.signalModifiedKey()
			n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(meta.key.expiration))
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
		if meta.key.expiration == 0 {
			meta.key.expiration = unix
		}
		if meta.key.expiration < unix {
			meta.key.expiration = unix
			meta.signalModifiedKey()
			n.notify(pb.NewOp(pb.OpType_Expire, key).Expiration(meta.key.expiration))
		}
		return nil
	})
	return v
}

// Keys gets the keys
func (n *Nodis) Keys(pattern string) []string {
	var keys []string
	now := time.Now().UnixMilli()
	n.store.mu.Lock()
	n.store.keys.Scan(func(key string, k *Key) bool {
		matched, _ := filepath.Match(pattern, key)
		if matched && !k.expired(now) {
			keys = append(keys, key)
		}
		return true
	})
	n.store.mu.Unlock()
	return keys
}

// Keys gets the keys
func (n *Nodis) Keyspace() (keys int64, expires int64, avgTTL int64) {
	n.store.mu.Lock()
	keys = int64(n.store.keys.Len())
	now := time.Now().UnixMilli()
	n.store.keys.Scan(func(key string, k *Key) bool {
		if k.expired(now) {
			expires++
		}
		if k.expiration != 0 {
			avgTTL += k.expiration - now
		}
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
		if meta.key.expiration == 0 {
			v = -1
			return nil
		}
		s := meta.key.expiration / 1000
		ns := (meta.key.expiration - s*1000) * 1000 * 1000
		v = time.Until(time.Unix(s, ns)).Round(time.Second)
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
		v, ok := n.store.values.Get(key)
		if !ok {
			return errors.New("key does not exist")
		}
		_ = tx.writeKey(dstKey, nil)
		tx.delKey(key)
		n.store.values.Set(dstKey, v)
		n.store.keys.Set(dstKey, newKey())
		meta.signalModifiedKey()
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
		v, ok := n.store.values.Get(key)
		if !ok {
			return errors.New("key does not exist")
		}
		tx.delKey(key)
		n.store.values.Set(dstKey, v)
		n.store.keys.Set(dstKey, newKey())
		meta.signalModifiedKey()
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
		v = meta.ds.Type().String()
		return nil
	})
	return v
}

// Scan the keys
func (n *Nodis) Scan(cursor int64, match string, count int64, typ ds.DataType) (int64, []string) {
	keyLen := int64(n.store.keys.Len())
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
	n.store.keys.Scan(func(key string, k *Key) bool {
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
		if matched && !k.expired(now) {
			if typ != 0 && k.dataType != typ {
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
	n.store.keys.Scan(func(key string, k *Key) bool {
		if !k.expired(now) {
			keys = append(keys, key)
		}
		return true
	})
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
		if meta.key.expiration == 0 {
			return nil
		}
		v = 1
		meta.key.expiration = 0
		meta.signalModifiedKey()
		n.notify(pb.NewOp(pb.OpType_Persist, key))
		return nil
	})
	return v
}
