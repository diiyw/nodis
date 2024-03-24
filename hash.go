package nodis

import (
	"github.com/diiyw/nodis/pb"
	"time"

	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/ds/hash"
)

func (n *Nodis) newHash() ds.DataStruct {
	return hash.NewHashMap()
}

func (n *Nodis) HSet(key string, field string, value []byte) {
	k, h := n.getDs(key, n.newHash, 0)
	k.changed.Store(true)
	h.(*hash.HashMap).HSet(field, value)
	n.notify(pb.NewOp(pb.OpType_HSet, key).Field(field).Value(value))
}

func (n *Nodis) HGet(key string, field string) []byte {
	_, h := n.getDs(key, nil, 0)
	if h == nil {
		return nil
	}
	return h.(*hash.HashMap).HGet(field)
}

func (n *Nodis) HDel(key string, field string) {
	k, h := n.getDs(key, nil, 0)
	if h == nil {
		return
	}
	k.changed.Store(true)
	h.(*hash.HashMap).HDel(field)
	if h.(*hash.HashMap).HLen() == 0 {
		n.dataStructs.Delete(key)
		n.keys.Delete(key)
	}
	n.notify(pb.NewOp(pb.OpType_HDel, key).Field(field))
}

func (n *Nodis) HLen(key string) int {
	_, h := n.getDs(key, nil, 0)
	if h == nil {
		return 0
	}
	return h.(*hash.HashMap).HLen()
}

func (n *Nodis) HKeys(key string) []string {
	_, h := n.getDs(key, nil, 0)
	if h == nil {
		return nil
	}
	return h.(*hash.HashMap).HKeys()
}

func (n *Nodis) HExists(key string, field string) bool {
	_, h := n.getDs(key, nil, 0)
	if h == nil {
		return false
	}
	return h.(*hash.HashMap).HExists(field)
}

func (n *Nodis) HGetAll(key string) map[string][]byte {
	_, h := n.getDs(key, nil, 0)
	if h == nil {
		return nil
	}
	return h.(*hash.HashMap).HGetAll()
}

func (n *Nodis) HIncrBy(key string, field string, value int64) int64 {
	k, h := n.getDs(key, n.newHash, 0)
	k.changed.Store(true)
	n.notify(pb.NewOp(pb.OpType_HIncrBy, key).Field(field).IncrInt(value))
	return h.(*hash.HashMap).HIncrBy(field, value)
}

func (n *Nodis) HIncrByFloat(key string, field string, value float64) float64 {
	k, h := n.getDs(key, n.newHash, 0)
	k.changed.Store(true)
	n.notify(pb.NewOp(pb.OpType_HIncrByFloat, key).Field(field).IncrFloat(value))
	return h.(*hash.HashMap).HIncrByFloat(field, value)
}

func (n *Nodis) HSetNX(key string, field string, value []byte) bool {
	_, h := n.getDs(key, nil, 0)
	if h != nil {
		return false
	}
	h = n.newHash()
	n.dataStructs.Set(key, h)
	k := newKey(h.Type(), 0)
	k.lastUse.Store(uint32(time.Now().Unix()))
	n.keys.Set(key, k)
	k.changed.Store(true)
	n.HSet(key, field, value)
	n.notify(pb.NewOp(pb.OpType_HSetNX, key).Field(field).Value(value))
	return true
}

func (n *Nodis) HMSet(key string, fields map[string][]byte) {
	k, h := n.getDs(key, n.newHash, 0)
	k.changed.Store(true)
	var ops = make([]*pb.Op, 0, len(fields))
	for field, value := range fields {
		h.(*hash.HashMap).HSet(field, value)
		ops = append(ops, pb.NewOp(pb.OpType_HSet, key).Field(field).Value(value))
	}
	n.notify(ops...)
	h.(*hash.HashMap).HMSet(fields)
}

func (n *Nodis) HMGet(key string, fields ...string) [][]byte {
	_, h := n.getDs(key, nil, 0)
	if h == nil {
		return nil
	}
	return h.(*hash.HashMap).HMGet(fields...)
}

func (n *Nodis) HClear(key string) {
	n.Del(key)
}

func (n *Nodis) HScan(key string, cursor int, match string, count int) (int, map[string][]byte) {
	_, h := n.getDs(key, nil, 0)
	if h == nil {
		return 0, nil
	}
	return h.(*hash.HashMap).HScan(cursor, match, count)
}

func (n *Nodis) HVals(key string) [][]byte {
	_, h := n.getDs(key, nil, 0)
	if h == nil {
		return nil
	}
	return h.(*hash.HashMap).HVals()
}
