package nodis

import (
	"github.com/diiyw/nodis/pb"

	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/ds/hash"
)

func (n *Nodis) newHash() ds.DataStruct {
	return hash.NewHashMap()
}

func (n *Nodis) HSet(key string, field string, value []byte) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, n.newHash)
		v = meta.ds.(*hash.HashMap).HSet(field, value)
		meta.signalModifiedKey()
		n.notify(pb.NewOp(pb.OpType_HSet, key).Fields(field).Value(value))
		return nil
	})
	return v
}

func (n *Nodis) HGet(key string, field string) []byte {
	var v []byte
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		v = meta.ds.(*hash.HashMap).HGet(field)
		return nil
	})
	return v
}

func (n *Nodis) HDel(key string, fields ...string) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, nil)
		if !meta.isOk() {
			return nil
		}
		v = meta.ds.(*hash.HashMap).HDel(fields...)
		if meta.ds.(*hash.HashMap).HLen() == 0 {
			tx.delKey(key)
		}
		meta.signalModifiedKey()
		n.notify(pb.NewOp(pb.OpType_HDel, key).Fields(fields...))
		return nil
	})
	return v
}

func (n *Nodis) HLen(key string) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		v = meta.ds.(*hash.HashMap).HLen()
		return nil
	})
	return v
}

func (n *Nodis) HKeys(key string) []string {
	var v []string
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		v = meta.ds.(*hash.HashMap).HKeys()
		return nil
	})
	return v
}

func (n *Nodis) HExists(key string, field string) bool {
	var v bool
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		v = meta.ds.(*hash.HashMap).HExists(field)
		return nil
	})
	return v
}

func (n *Nodis) HGetAll(key string) map[string][]byte {
	var v map[string][]byte
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {

			return nil
		}
		v = meta.ds.(*hash.HashMap).HGetAll()
		return nil
	})
	return v
}

func (n *Nodis) HIncrBy(key string, field string, value int64) (int64, error) {
	var v int64
	var err error
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, n.newHash)
		v, err = meta.ds.(*hash.HashMap).HIncrBy(field, value)
		meta.signalModifiedKey()
		n.notify(pb.NewOp(pb.OpType_HIncrBy, key).Fields(field).IncrInt(value))
		return nil
	})
	return v, err
}

func (n *Nodis) HIncrByFloat(key string, field string, value float64) (float64, error) {
	var v float64
	var err error
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, n.newHash)
		v, err = meta.ds.(*hash.HashMap).HIncrByFloat(field, value)
		meta.signalModifiedKey()
		n.notify(pb.NewOp(pb.OpType_HIncrByFloat, key).Fields(field).IncrFloat(value))
		return err
	})
	return v, err
}

// HSetNX Sets field in the hash stored at key to value, only if field does not yet exist.
// If key does not exist, a new key holding a hash is created.
// If field already exists, this operation has no effect.
func (n *Nodis) HSetNX(key string, field string, value []byte) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, nil)
		if meta.isOk() && meta.ds.(*hash.HashMap).HExists(field) {
			return nil
		}
		if !meta.isOk() {
			meta.ds = n.newHash()
			meta.ok = true
			n.store.values.Set(key, meta.ds)
			k := newKey()
			n.store.keys.Set(key, k)
			meta.key = k
		}
		v = meta.ds.(*hash.HashMap).HSet(field, value)
		meta.signalModifiedKey()
		n.notify(pb.NewOp(pb.OpType_HSet, key).Fields(field).Value(value))
		return nil
	})
	return v
}

func (n *Nodis) HMSet(key string, fields map[string][]byte) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, n.newHash)
		var ops = make([]*pb.Op, 0, len(fields))
		var v int64 = 0
		for field, value := range fields {
			v += meta.ds.(*hash.HashMap).HSet(field, value)
			ops = append(ops, pb.NewOp(pb.OpType_HSet, key).Fields(field).Value(value))
		}
		meta.signalModifiedKey()
		n.notify(ops...)
		meta.ds.(*hash.HashMap).HMSet(fields)
		return nil
	})
	return v
}

func (n *Nodis) HMGet(key string, fields ...string) [][]byte {
	var v [][]byte
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		v = meta.ds.(*hash.HashMap).HMGet(fields...)
		return nil
	})
	return v
}

func (n *Nodis) HClear(key string) {
	n.Del(key)
}

func (n *Nodis) HScan(key string, cursor int64, match string, count int64) (int64, map[string][]byte) {
	var c int64
	var v map[string][]byte
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		c, v = meta.ds.(*hash.HashMap).HScan(cursor, match, count)
		return nil
	})
	return c, v
}

func (n *Nodis) HVals(key string) [][]byte {
	var v [][]byte
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {

			return nil
		}
		v = meta.ds.(*hash.HashMap).HVals()
		return nil
	})
	return v
}

func (n *Nodis) HStrLen(key, field string) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		v = meta.ds.(*hash.HashMap).HStrLen(field)
		return nil
	})
	return v
}
