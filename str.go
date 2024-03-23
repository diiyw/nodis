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
