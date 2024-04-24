package nodis

import (
	"encoding/binary"
	"time"

	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/ds/str"
	"github.com/diiyw/nodis/pb"
)

func (n *Nodis) newStr() ds.DataStruct {
	return str.NewString()
}

// Set a key with a value and a TTL
func (n *Nodis) Set(key string, value []byte) {
	meta := n.store.writeKey(key, n.newStr)
	n.notify(pb.NewOp(pb.OpType_Set, key).Value(value))
	meta.ds.(*str.String).Set(value)
	meta.commit()
}

// SetEX set a key with specified expire time, in seconds (a positive integer).
func (n *Nodis) SetEX(key string, value []byte, seconds int64) {
	meta := n.store.writeKey(key, n.newStr)
	if meta.key.expiration == 0 {
		meta.key.expiration = time.Now().UnixMilli()
	}
	meta.key.expiration += seconds * 1000
	meta.ds.(*str.String).Set(value)
	n.notify(pb.NewOp(pb.OpType_Set, key).Value(value).Expiration(meta.key.expiration))
	meta.commit()
}

// SetPX set a key with specified expire time, in milliseconds (a positive integer).
func (n *Nodis) SetPX(key string, value []byte, milliseconds int64) {
	meta := n.store.writeKey(key, n.newStr)
	if meta.key.expiration == 0 {
		meta.key.expiration = time.Now().UnixMilli()
	}
	meta.key.expiration += milliseconds
	n.notify(pb.NewOp(pb.OpType_Set, key).Value(value).Expiration(meta.key.expiration))
	meta.ds.(*str.String).Set(value)
	meta.commit()
}

// SetNX set a key with a value if it does not exist
func (n *Nodis) SetNX(key string, value []byte) bool {
	meta := n.store.writeKey(key, nil)
	if meta.isOk() {
		meta.commit()
		return false
	}
	meta = n.store.writeKey(key, nil)
	n.notify(pb.NewOp(pb.OpType_Set, key).Value(value))
	meta.ds.(*str.String).Set(value)
	meta.commit()
	return true
}

// SetXX set a key with a value if it exists
func (n *Nodis) SetXX(key string, value []byte) bool {
	meta := n.store.writeKey(key, nil)
	if !meta.isOk() {
		meta.commit()
		return false
	}
	n.notify(pb.NewOp(pb.OpType_Set, key).Value(value))
	meta.ds.(*str.String).Set(value)
	meta.commit()
	return true
}

// Get a key
func (n *Nodis) Get(key string) []byte {
	meta := n.store.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return nil
	}
	v := meta.ds.(*str.String).Get()
	meta.commit()
	return v
}

// Incr increment the integer value of a key by one
func (n *Nodis) Incr(key string) int64 {
	meta := n.store.writeKey(key, n.newStr)
	v := meta.ds.(*str.String).Incr(1)
	m := make([]byte, 8)
	binary.LittleEndian.PutUint64(m, uint64(v))
	n.notify(pb.NewOp(pb.OpType_Set, key).Value(m))
	meta.commit()
	return v
}

func (n *Nodis) IncrBy(key string, increment int64) int64 {
	meta := n.store.writeKey(key, n.newStr)
	v := meta.ds.(*str.String).Incr(increment)
	m := make([]byte, 8)
	binary.LittleEndian.PutUint64(m, uint64(v))
	n.notify(pb.NewOp(pb.OpType_Set, key).Value(m))
	meta.commit()
	return v
}

// Decr decrement the integer value of a key by one
func (n *Nodis) Decr(key string) int64 {
	meta := n.store.writeKey(key, n.newStr)
	v := meta.ds.(*str.String).Decr(1)
	m := make([]byte, 8)
	binary.LittleEndian.PutUint64(m, uint64(v))
	n.notify(pb.NewOp(pb.OpType_Set, key).Value(m))
	meta.commit()
	return v
}

func (n *Nodis) DecrBy(key string, decrement int64) int64 {
	meta := n.store.writeKey(key, n.newStr)
	v := meta.ds.(*str.String).Decr(decrement)
	m := make([]byte, 8)
	binary.LittleEndian.PutUint64(m, uint64(v))
	n.notify(pb.NewOp(pb.OpType_Set, key).Value(m))
	meta.commit()
	return v
}

// SetBit set a bit in a key
func (n *Nodis) SetBit(key string, offset int64, value bool) int {
	meta := n.store.writeKey(key, n.newStr)
	k := meta.ds.(*str.String)
	v := k.SetBit(offset, value)
	meta.commit()
	return v
}

// GetBit get a bit in a key
func (n *Nodis) GetBit(key string, offset int64) int64 {
	meta := n.store.writeKey(key, nil)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	v := meta.ds.(*str.String).GetBit(offset)
	meta.commit()
	return v
}

// BitCount returns the number of bits set to 1
func (n *Nodis) BitCount(key string, start, end int64) int64 {
	meta := n.store.writeKey(key, nil)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	v := meta.ds.(*str.String).BitCount(start, end)
	meta.commit()
	return v
}
