package nodis

import (
	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/ds/set"
)

func (n *Nodis) newSet() ds.DataStruct {
	return set.NewSet()
}

// Set a key with a value and a TTL
func (n *Nodis) Set(key string, value []byte, ttl int64) {
	s := n.getDs(key, n.newSet, ttl)
	s.(*set.Set).Set(value)
}

// Get a key
func (n *Nodis) Get(key string) []byte {
	s := n.getDs(key, nil, 0)
	if s == nil {
		return nil
	}
	return s.(*set.Set).Get()
}
