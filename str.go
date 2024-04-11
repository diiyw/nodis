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
	tx := n.writeKey(key, n.newStr)
	n.notify(pb.NewOp(pb.OpType_Set, key).Value(value))
	tx.ds.(*str.String).Set(value)
	tx.commit()
}

// SetEX set a key with specified expire time, in seconds (a positive integer).
func (n *Nodis) SetEX(key string, value []byte, seconds int64) {
	tx := n.writeKey(key, n.newStr)
	if tx.key.expiration == 0 {
		tx.key.expiration = time.Now().UnixMilli()
	}
	tx.key.expiration += seconds * 1000
	tx.ds.(*str.String).Set(value)
	tx.commit()
	n.notify(pb.NewOp(pb.OpType_Set, key).Value(value).Expiration(tx.key.expiration))
}

// SetPX set a key with specified expire time, in milliseconds (a positive integer).
func (n *Nodis) SetPX(key string, value []byte, milliseconds int64) {
	tx := n.writeKey(key, n.newStr)
	if tx.key.expiration == 0 {
		tx.key.expiration = time.Now().UnixMilli()
	}
	tx.key.expiration += milliseconds
	n.notify(pb.NewOp(pb.OpType_Set, key).Value(value).Expiration(tx.key.expiration))
	tx.ds.(*str.String).Set(value)
	tx.commit()
}

// SetNX set a key with a value if it does not exist
func (n *Nodis) SetNX(key string, value []byte) bool {
	tx := n.writeKey(key, nil)
	if tx.isOk() {
		tx.commit()
		return false
	}
	tx = n.writeKey(key, nil)
	n.notify(pb.NewOp(pb.OpType_Set, key).Value(value))
	tx.ds.(*str.String).Set(value)
	tx.commit()
	return true
}

// SetXX set a key with a value if it exists
func (n *Nodis) SetXX(key string, value []byte) bool {
	tx := n.writeKey(key, nil)
	if !tx.isOk() {
		tx.commit()
		return false
	}
	n.notify(pb.NewOp(pb.OpType_Set, key).Value(value))
	tx.ds.(*str.String).Set(value)
	tx.commit()
	return true
}

// Get a key
func (n *Nodis) Get(key string) []byte {
	tx := n.writeKey(key, nil)
	if !tx.isOk() {
		tx.commit()
		return nil
	}
	v := tx.ds.(*str.String).Get()
	tx.commit()
	return v
}

// Incr increment the integer value of a key by one
func (n *Nodis) Incr(key string) int64 {
	tx := n.writeKey(key, n.newStr)
	v := tx.ds.(*str.String).Incr()
	m := make([]byte, 8)
	binary.LittleEndian.PutUint64(m, uint64(v))
	n.notify(pb.NewOp(pb.OpType_Set, key).Value(m))
	tx.commit()
	return v
}

// Decr decrement the integer value of a key by one
func (n *Nodis) Decr(key string) int64 {
	tx := n.writeKey(key, n.newStr)
	v := tx.ds.(*str.String).Decr()
	m := make([]byte, 8)
	binary.LittleEndian.PutUint64(m, uint64(v))
	n.notify(pb.NewOp(pb.OpType_Set, key).Value(m))
	tx.commit()
	return v
}

// SetBit set a bit in a key
func (n *Nodis) SetBit(key string, offset int64, value bool) int {
	tx := n.writeKey(key, n.newStr)
	k := tx.ds.(*str.String)
	v := k.SetBit(offset, value)
	tx.commit()
	return v
}

// GetBit get a bit in a key
func (n *Nodis) GetBit(key string, offset int64) int64 {
	tx := n.writeKey(key, nil)
	if !tx.isOk() {
		tx.commit()
		return 0
	}
	v := tx.ds.(*str.String).GetBit(offset)
	tx.commit()
	return v
}

// BitCount returns the number of bits set to 1
func (n *Nodis) BitCount(key string, start, end int64) int64 {
	tx := n.writeKey(key, nil)
	if !tx.isOk() {
		tx.commit()
		return 0
	}
	v := tx.ds.(*str.String).BitCount(start, end)
	tx.commit()
	return v
}
