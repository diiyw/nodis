package nodis

import (
	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/ds/str"
	"github.com/diiyw/nodis/pb"
)

func (n *Nodis) newStr() ds.DataStruct {
	return str.NewString()
}

// Set a key with a value and a TTL
func (n *Nodis) Set(key string, value []byte, ttl int64) {
	k, s := n.getDs(key, n.newStr, ttl)
	k.changed.Store(true)
	n.notify(pb.NewOp(pb.OpType_Set, key).Value(value))
	s.(*str.String).Set(value)
}

// Get a key
func (n *Nodis) Get(key string) []byte {
	_, s := n.getDs(key, nil, 0)
	if s == nil {
		return nil
	}
	return s.(*str.String).Get()
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
