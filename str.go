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
	k, s := n.getDs(key, n.newStr, 0)
	k.changed.Store(true)
	n.notify(pb.NewOp(pb.OpType_Set, key).Value(value))
	s.(*str.String).Set(value)
}

// SetEX set a key with specified expire time, in seconds (a positive integer).
func (n *Nodis) SetEX(key string, value []byte, seconds int64) {
	k, s := n.getDs(key, n.newStr, seconds)
	k.changed.Store(true)
	if k.Expiration == 0 {
		k.Expiration = time.Now().UnixMilli()
	}
	k.Expiration += seconds * 1000
	n.notify(pb.NewOp(pb.OpType_Set, key).Value(value).Expiration(k.Expiration))
	s.(*str.String).Set(value)
}

// SetPX set a key with specified expire time, in milliseconds (a positive integer).
func (n *Nodis) SetPX(key string, value []byte, milliseconds int64) {
	k, s := n.getDs(key, n.newStr, milliseconds/1000)
	k.changed.Store(true)
	if k.Expiration == 0 {
		k.Expiration = time.Now().UnixMilli()
	}
	k.Expiration += milliseconds
	n.notify(pb.NewOp(pb.OpType_Set, key).Value(value).Expiration(k.Expiration))
	s.(*str.String).Set(value)
}

// SetNX set a key with a value if it does not exist
func (n *Nodis) SetNX(key string, value []byte) bool {
	_, s := n.getDs(key, nil, 0)
	if s != nil {
		return false
	}
	k, s := n.getDs(key, n.newStr, 0)
	k.changed.Store(true)
	n.notify(pb.NewOp(pb.OpType_Set, key).Value(value))
	s.(*str.String).Set(value)
	return true
}

// SetXX set a key with a value if it exists
func (n *Nodis) SetXX(key string, value []byte) bool {
	_, s := n.getDs(key, nil, 0)
	if s == nil {
		return false
	}
	k, s := n.getDs(key, n.newStr, 0)
	k.changed.Store(true)
	n.notify(pb.NewOp(pb.OpType_Set, key).Value(value))
	s.(*str.String).Set(value)
	return true
}

// Get a key
func (n *Nodis) Get(key string) []byte {
	_, s := n.getDs(key, nil, 0)
	if s == nil {
		return nil
	}
	return s.(*str.String).Get()
}

// Incr increment the integer value of a key by one
func (n *Nodis) Incr(key string) int64 {
	k, s := n.getDs(key, n.newStr, 0)
	k.changed.Store(true)
	v := s.(*str.String).Incr()
	m := make([]byte, 8)
	binary.LittleEndian.PutUint64(m, uint64(v))
	n.notify(pb.NewOp(pb.OpType_Set, key).Value(m))
	return v
}

// Decr decrement the integer value of a key by one
func (n *Nodis) Decr(key string) int64 {
	k, s := n.getDs(key, n.newStr, 0)
	k.changed.Store(true)
	v := s.(*str.String).Decr()
	m := make([]byte, 8)
	binary.LittleEndian.PutUint64(m, uint64(v))
	n.notify(pb.NewOp(pb.OpType_Set, key).Value(m))
	return v
}

// SetBit set a bit in a key
func (n *Nodis) SetBit(key string, offset int64, value bool) int {
	_, s := n.getDs(key, n.newStr, 0)
	k := s.(*str.String)
	return k.SetBit(offset, value)
}

// GetBit get a bit in a key
func (n *Nodis) GetBit(key string, offset int64) int64 {
	_, s := n.getDs(key, nil, 0)
	if s == nil {
		return 0
	}
	return s.(*str.String).GetBit(offset)
}

// BitCount returns the number of bits set to 1
func (n *Nodis) BitCount(key string, start, end int64) int64 {
	_, s := n.getDs(key, nil, 0)
	if s == nil {
		return 0
	}
	return s.(*str.String).BitCount(start, end)
}
