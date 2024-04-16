package nodis

import (
	"time"

	"github.com/diiyw/nodis/pb"

	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/ds/hash"
)

func (n *Nodis) newHash() ds.DataStruct {
	return hash.NewHashMap()
}

func (n *Nodis) HSet(key string, field string, value []byte) {
	meta := n.store.writeKey(key, n.newHash)
	meta.ds.(*hash.HashMap).HSet(field, value)
	meta.commit()
	n.notify(pb.NewOp(pb.OpType_HSet, key).Fields(field).Value(value))
}

func (n *Nodis) HGet(key string, field string) []byte {
	meta := n.store.readKey(key)
	defer meta.commit()
	if !meta.isOk() {
		return nil
	}
	return meta.ds.(*hash.HashMap).HGet(field)
}

func (n *Nodis) HDel(key string, fields ...string) int64 {
	meta := n.store.writeKey(key, nil)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	meta.ds.(*hash.HashMap).HDel(fields...)
	if meta.ds.(*hash.HashMap).HLen() == 0 {
		n.store.delKey(key)
	}
	meta.commit()
	n.notify(pb.NewOp(pb.OpType_HDel, key).Fields(fields...))
	return int64(len(fields))
}

func (n *Nodis) HLen(key string) int64 {
	meta := n.store.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	v := meta.ds.(*hash.HashMap).HLen()
	meta.commit()
	return v
}

func (n *Nodis) HKeys(key string) []string {
	meta := n.store.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return nil
	}
	v := meta.ds.(*hash.HashMap).HKeys()
	meta.commit()
	return v
}

func (n *Nodis) HExists(key string, field string) bool {
	meta := n.store.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return false
	}
	v := meta.ds.(*hash.HashMap).HExists(field)
	meta.commit()
	return v
}

func (n *Nodis) HGetAll(key string) map[string][]byte {
	meta := n.store.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return nil
	}
	v := meta.ds.(*hash.HashMap).HGetAll()
	meta.commit()
	return v
}

func (n *Nodis) HIncrBy(key string, field string, value int64) int64 {
	meta := n.store.writeKey(key, n.newHash)
	v := meta.ds.(*hash.HashMap).HIncrBy(field, value)
	meta.commit()
	n.notify(pb.NewOp(pb.OpType_HIncrBy, key).Fields(field).IncrInt(value))
	return v
}

func (n *Nodis) HIncrByFloat(key string, field string, value float64) float64 {
	meta := n.store.writeKey(key, n.newHash)
	v := meta.ds.(*hash.HashMap).HIncrByFloat(field, value)
	meta.commit()
	n.notify(pb.NewOp(pb.OpType_HIncrByFloat, key).Fields(field).IncrFloat(value))
	return v
}

func (n *Nodis) HSetNX(key string, field string, value []byte) bool {
	meta := n.store.writeKey(key, nil)
	if meta.isOk() && meta.ds.(*hash.HashMap).HExists(field) {
		meta.commit()
		return false
	}
	h := n.newHash()
	n.store.values.Set(key, h)
	k := newKey()
	k.lastUse = time.Now().Unix()
	n.store.keys.Set(key, k)
	h.(*hash.HashMap).HSet(field, value)
	meta.commit()
	return true
}

func (n *Nodis) HMSet(key string, fields map[string][]byte) {
	meta := n.store.writeKey(key, n.newHash)
	var ops = make([]*pb.Op, 0, len(fields))
	for field, value := range fields {
		meta.ds.(*hash.HashMap).HSet(field, value)
		ops = append(ops, pb.NewOp(pb.OpType_HSet, key).Fields(field).Value(value))
	}
	n.notify(ops...)
	meta.ds.(*hash.HashMap).HMSet(fields)
	meta.commit()
}

func (n *Nodis) HMGet(key string, fields ...string) [][]byte {
	meta := n.store.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return nil
	}
	v := meta.ds.(*hash.HashMap).HMGet(fields...)
	meta.commit()
	return v
}

func (n *Nodis) HClear(key string) {
	n.Del(key)
}

func (n *Nodis) HScan(key string, cursor int64, match string, count int64) (int64, map[string][]byte) {
	meta := n.store.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return 0, nil
	}
	c, v := meta.ds.(*hash.HashMap).HScan(cursor, match, count)
	meta.commit()
	return c, v
}

func (n *Nodis) HVals(key string) [][]byte {
	meta := n.store.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return nil
	}
	v := meta.ds.(*hash.HashMap).HVals()
	meta.commit()
	return v
}
