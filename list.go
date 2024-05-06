package nodis

import (
	"time"

	"github.com/diiyw/nodis/pb"

	"github.com/diiyw/nodis/ds"

	"github.com/diiyw/nodis/ds/list"
)

// newList creates a new list
func (n *Nodis) newList() ds.DataStruct {
	return list.NewDoublyLinkedList()
}

func (n *Nodis) LPush(key string, values ...[]byte) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, n.newList)
		meta.ds.(*list.DoublyLinkedList).LPush(values...)
		v = meta.ds.(*list.DoublyLinkedList).LLen()
		n.notify(pb.NewOp(pb.OpType_LPush, key).Values(values))
		return nil
	})
	return v
}

func (n *Nodis) RPush(key string, values ...[]byte) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, n.newList)
		meta.ds.(*list.DoublyLinkedList).RPush(values...)
		v = meta.ds.(*list.DoublyLinkedList).LLen()
		n.notify(pb.NewOp(pb.OpType_RPush, key).Values(values))
		return nil
	})
	return v
}

func (n *Nodis) LPop(key string, count int64) [][]byte {
	var v [][]byte
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, nil)
		if !meta.isOk() {
			return nil
		}
		v = meta.ds.(*list.DoublyLinkedList).LPop(count)
		if meta.ds.(*list.DoublyLinkedList).LLen() == 0 {
			n.store.delKey(key)
		}
		n.notify(pb.NewOp(pb.OpType_LPop, key).Count(count))
		return nil
	})
	return v
}

func (n *Nodis) RPop(key string, count int64) [][]byte {
	var v [][]byte
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, nil)
		if !meta.isOk() {
			return nil
		}
		v = meta.ds.(*list.DoublyLinkedList).RPop(count)
		if meta.ds.(*list.DoublyLinkedList).LLen() == 0 {
			n.store.delKey(key)
		}
		n.notify(pb.NewOp(pb.OpType_RPop, key).Count(count))
		return nil
	})
	return v
}

func (n *Nodis) LLen(key string) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		d, ok := meta.ds.(*list.DoublyLinkedList)
		if !ok {
			v = -1
			return nil
		}
		v = d.LLen()
		return nil
	})
	return v
}

func (n *Nodis) LIndex(key string, index int64) []byte {
	var v []byte
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		v = meta.ds.(*list.DoublyLinkedList).LIndex(index)
		return nil
	})
	return v
}

func (n *Nodis) LInsert(key string, pivot, data []byte, before bool) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, nil)
		if !meta.isOk() {
			return nil
		}
		v = meta.ds.(*list.DoublyLinkedList).LInsert(pivot, data, before)
		n.notify(pb.NewOp(pb.OpType_LInsert, key).Value(data).Pivot(pivot).Before(before))
		return nil
	})
	return v
}

func (n *Nodis) LPushX(key string, data []byte) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, nil)
		if !meta.isOk() {
			return nil
		}
		v = meta.ds.(*list.DoublyLinkedList).LPushX(data)
		n.notify(pb.NewOp(pb.OpType_LPushX, key).Value(data))
		return nil
	})
	return v
}

func (n *Nodis) RPushX(key string, data []byte) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, nil)
		if !meta.isOk() {
			return nil
		}
		v = meta.ds.(*list.DoublyLinkedList).RPushX(data)
		n.notify(pb.NewOp(pb.OpType_RPushX, key).Value(data))
		return nil
	})
	return v
}

func (n *Nodis) LRem(key string, data []byte, count int64) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, nil)
		if !meta.isOk() {
			return nil
		}
		v = meta.ds.(*list.DoublyLinkedList).LRem(count, data)
		n.notify(pb.NewOp(pb.OpType_LRem, key).Value(data).Count(count))
		return nil
	})
	return v
}

func (n *Nodis) LSet(key string, index int64, data []byte) bool {
	var v bool
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, n.newList)
		n.notify(pb.NewOp(pb.OpType_LSet, key).Value(data).Index(index))
		v = meta.ds.(*list.DoublyLinkedList).LSet(index, data)
		return nil
	})
	return v
}

func (n *Nodis) LTrim(key string, start, stop int64) {
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, nil)
		if !meta.isOk() {
			return nil
		}
		meta.ds.(*list.DoublyLinkedList).LTrim(start, stop)
		n.notify(pb.NewOp(pb.OpType_LTrim, key).Start(start).Stop(stop))
		return nil
	})
}

func (n *Nodis) LRange(key string, start, stop int64) [][]byte {
	var v [][]byte
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		v = meta.ds.(*list.DoublyLinkedList).LRange(start, stop)
		return nil
	})
	return v
}

func (n *Nodis) LPopRPush(source, destination string) []byte {
	var v [][]byte
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(source, nil)
		if !meta.isOk() {

			return nil
		}
		v = meta.ds.(*list.DoublyLinkedList).LPop(1)
		if v == nil {
			return nil
		}
		if meta.ds.(*list.DoublyLinkedList).LLen() == 0 {
			n.store.delKey(source)
		}
		dst := tx.writeKey(destination, n.newList)
		dst.ds.(*list.DoublyLinkedList).RPush(v...)
		n.notify(pb.NewOp(pb.OpType_LPopRPush, source).DstKey(destination))
		return nil
	})
	return v[0]
}

func (n *Nodis) RPopLPush(source, destination string) []byte {
	var v [][]byte
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(source, nil)
		if !meta.isOk() {
			return nil
		}
		v = meta.ds.(*list.DoublyLinkedList).RPop(1)
		if v == nil {
			return nil
		}
		if meta.ds.(*list.DoublyLinkedList).LLen() == 0 {
			n.store.delKey(source)
		}
		dst := tx.writeKey(destination, n.newList)
		dst.ds.(*list.DoublyLinkedList).LPush(v...)
		n.notify(pb.NewOp(pb.OpType_RPopLPush, source).DstKey(destination))
		return nil
	})
	return v[0]
}

func (n *Nodis) BLPop(timeout time.Duration, keys ...string) (string, []byte) {
	for _, key := range keys {
		results := n.LPop(key, 1)
		if results != nil {
			n.notify(pb.NewOp(pb.OpType_LPop, key))
			return key, results[0]
		}
	}
	time.Sleep(timeout)
	for _, key := range keys {
		results := n.LPop(key, 1)
		if results != nil {
			n.notify(pb.NewOp(pb.OpType_LPop, key))
			return key, results[0]
		}
	}
	return "", nil
}

func (n *Nodis) BRPop(timeout time.Duration, keys ...string) (string, []byte) {
	for _, key := range keys {
		results := n.RPop(key, 1)
		if results != nil {
			n.notify(pb.NewOp(pb.OpType_RPop, key))
			return key, results[0]
		}
	}
	time.Sleep(timeout)
	for _, key := range keys {
		results := n.RPop(key, 1)
		if results != nil {
			n.notify(pb.NewOp(pb.OpType_RPop, key))
			return key, results[0]
		}
	}
	return "", nil
}
