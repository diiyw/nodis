package nodis

import "github.com/diiyw/nodis/ds/set"

func (n *Nodis) newSet() *set.Set {
	return set.NewSet()
}

// Set a key with a value and a TTL
func (n *Nodis) Set(key string, value []byte, ttl int64) {
	s := n.getDs(key)
	if s == nil {
		s = n.newSet()
	}
	s.(*set.Set).Set(value)
	n.saveDs(key, s, ttl)
}

// Get a key
func (n *Nodis) Get(key string) []byte {
	s := n.getDs(key)
	if s == nil {
		return nil
	}
	return s.(*set.Set).Get()
}
