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
	tx := n.writeKey(key, n.newHash)
	tx.ds.(*hash.HashMap).HSet(field, value)
	tx.commit()
	n.notify(pb.NewOp(pb.OpType_HSet, key).Fields(field).Value(value))
}

func (n *Nodis) HGet(key string, field string) []byte {
	tx := n.readKey(key)
	if !tx.isOk() {
		return nil
	}
	defer tx.commit()
	return tx.ds.(*hash.HashMap).HGet(field)
}

func (n *Nodis) HDel(key string, fields ...string) int64 {
	tx := n.writeKey(key, nil)
	if !tx.isOk() {
		return 0
	}
	tx.ds.(*hash.HashMap).HDel(fields...)
	if tx.ds.(*hash.HashMap).HLen() == 0 {
		n.delKey(key)
	}
	tx.commit()
	n.notify(pb.NewOp(pb.OpType_HDel, key).Fields(fields...))
	return int64(len(fields))
}

func (n *Nodis) HLen(key string) int64 {
	tx := n.readKey(key)
	if !tx.isOk() {
		return 0
	}
	v := tx.ds.(*hash.HashMap).HLen()
	tx.commit()
	return v
}

func (n *Nodis) HKeys(key string) []string {
	tx := n.readKey(key)
	if !tx.isOk() {
		return nil
	}
	v := tx.ds.(*hash.HashMap).HKeys()
	tx.commit()
	return v
}

func (n *Nodis) HExists(key string, field string) bool {
	tx := n.readKey(key)
	if !tx.isOk() {
		return false
	}
	v := tx.ds.(*hash.HashMap).HExists(field)
	tx.commit()
	return v
}

func (n *Nodis) HGetAll(key string) map[string][]byte {
	tx := n.readKey(key)
	if !tx.isOk() {
		return nil
	}
	v := tx.ds.(*hash.HashMap).HGetAll()
	tx.commit()
	return v
}

func (n *Nodis) HIncrBy(key string, field string, value int64) int64 {
	tx := n.writeKey(key, n.newHash)
	v := tx.ds.(*hash.HashMap).HIncrBy(field, value)
	tx.commit()
	n.notify(pb.NewOp(pb.OpType_HIncrBy, key).Fields(field).IncrInt(value))
	return v
}

func (n *Nodis) HIncrByFloat(key string, field string, value float64) float64 {
	tx := n.writeKey(key, n.newHash)
	v := tx.ds.(*hash.HashMap).HIncrByFloat(field, value)
	tx.commit()
	n.notify(pb.NewOp(pb.OpType_HIncrByFloat, key).Fields(field).IncrFloat(value))
	return v
}

func (n *Nodis) HSetNX(key string, field string, value []byte) bool {
	tx := n.writeKey(key, nil)
	if tx.isOk() && tx.ds.(*hash.HashMap).HExists(field) {
		tx.commit()
		return false
	}
	h := n.newHash()
	n.dataStructs.Set(key, h)
	k := newKey(h.Type())
	k.lastUse = time.Now().Unix()
	n.keys.Set(key, k)
	n.HSet(key, field, value)
	tx.commit()
	return true
}

func (n *Nodis) HMSet(key string, fields map[string][]byte) {
	tx := n.writeKey(key, n.newHash)
	var ops = make([]*pb.Op, 0, len(fields))
	for field, value := range fields {
		tx.ds.(*hash.HashMap).HSet(field, value)
		ops = append(ops, pb.NewOp(pb.OpType_HSet, key).Fields(field).Value(value))
	}
	n.notify(ops...)
	tx.ds.(*hash.HashMap).HMSet(fields)
	tx.commit()
}

func (n *Nodis) HMGet(key string, fields ...string) [][]byte {
	tx := n.readKey(key)
	if !tx.isOk() {
		return nil
	}
	v := tx.ds.(*hash.HashMap).HMGet(fields...)
	tx.commit()
	return v
}

func (n *Nodis) HClear(key string) {
	n.Del(key)
}

func (n *Nodis) HScan(key string, cursor int64, match string, count int64) (int64, map[string][]byte) {
	tx := n.readKey(key)
	if !tx.isOk() {
		return 0, nil
	}
	c, v := tx.ds.(*hash.HashMap).HScan(cursor, match, count)
	tx.commit()
	return c, v
}

func (n *Nodis) HVals(key string) [][]byte {
	tx := n.readKey(key)
	if !tx.isOk() {
		return nil
	}
	v := tx.ds.(*hash.HashMap).HVals()
	tx.commit()
	return v
}
