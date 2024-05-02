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
	meta := n.store.writeKey(key, n.newList)
	meta.ds.(*list.DoublyLinkedList).LPush(values...)
	v := meta.ds.(*list.DoublyLinkedList).LLen()
	meta.commit()
	n.notify(pb.NewOp(pb.OpType_LPush, key).Values(values))
	return v
}

func (n *Nodis) RPush(key string, values ...[]byte) int64 {
	meta := n.store.writeKey(key, n.newList)
	meta.ds.(*list.DoublyLinkedList).RPush(values...)
	v := meta.ds.(*list.DoublyLinkedList).LLen()
	meta.commit()
	n.notify(pb.NewOp(pb.OpType_RPush, key).Values(values))
	return v
}

func (n *Nodis) LPop(key string, count int64) [][]byte {
	meta := n.store.writeKey(key, nil)
	if !meta.isOk() {
		meta.commit()
		return nil
	}
	v := meta.ds.(*list.DoublyLinkedList).LPop(count)
	if meta.ds.(*list.DoublyLinkedList).LLen() == 0 {
		n.store.delKey(key)
	}
	meta.commit()
	n.notify(pb.NewOp(pb.OpType_LPop, key).Count(count))
	return v
}

func (n *Nodis) RPop(key string, count int64) [][]byte {
	meta := n.store.writeKey(key, nil)
	if !meta.isOk() {
		meta.commit()
		return nil
	}
	v := meta.ds.(*list.DoublyLinkedList).RPop(count)
	if meta.ds.(*list.DoublyLinkedList).LLen() == 0 {
		n.store.delKey(key)
	}
	meta.commit()
	n.notify(pb.NewOp(pb.OpType_RPop, key).Count(count))
	return v
}

func (n *Nodis) LLen(key string) int64 {
	meta := n.store.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	ds, ok := meta.ds.(*list.DoublyLinkedList)
	if !ok {
		meta.commit()
		return -1
	}
	v := ds.LLen()
	meta.commit()
	return v
}

func (n *Nodis) LIndex(key string, index int64) []byte {
	meta := n.store.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return nil
	}
	v := meta.ds.(*list.DoublyLinkedList).LIndex(index)
	meta.commit()
	return v
}

func (n *Nodis) LInsert(key string, pivot, data []byte, before bool) int64 {
	meta := n.store.writeKey(key, nil)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	v := meta.ds.(*list.DoublyLinkedList).LInsert(pivot, data, before)
	meta.commit()
	n.notify(pb.NewOp(pb.OpType_LInsert, key).Value(data).Pivot(pivot).Before(before))
	return v
}

func (n *Nodis) LPushX(key string, data []byte) int64 {
	meta := n.store.writeKey(key, nil)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	v := meta.ds.(*list.DoublyLinkedList).LPushX(data)
	meta.commit()
	n.notify(pb.NewOp(pb.OpType_LPushX, key).Value(data))
	return v
}

func (n *Nodis) RPushX(key string, data []byte) int64 {
	meta := n.store.writeKey(key, nil)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	v := meta.ds.(*list.DoublyLinkedList).RPushX(data)
	meta.commit()
	n.notify(pb.NewOp(pb.OpType_RPushX, key).Value(data))
	return v
}

func (n *Nodis) LRem(key string, data []byte, count int64) int64 {
	meta := n.store.writeKey(key, nil)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	v := meta.ds.(*list.DoublyLinkedList).LRem(count, data)
	meta.commit()
	n.notify(pb.NewOp(pb.OpType_LRem, key).Value(data).Count(count))
	return v
}

func (n *Nodis) LSet(key string, index int64, data []byte) bool {
	meta := n.store.writeKey(key, n.newList)
	n.notify(pb.NewOp(pb.OpType_LSet, key).Value(data).Index(index))
	v := meta.ds.(*list.DoublyLinkedList).LSet(index, data)
	meta.commit()
	return v
}

func (n *Nodis) LTrim(key string, start, stop int64) {
	meta := n.store.writeKey(key, nil)
	if !meta.isOk() {
		meta.commit()
		return
	}
	meta.ds.(*list.DoublyLinkedList).LTrim(start, stop)
	meta.commit()
}

func (n *Nodis) LRange(key string, start, stop int64) [][]byte {
	meta := n.store.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return nil
	}
	v := meta.ds.(*list.DoublyLinkedList).LRange(start, stop)
	meta.commit()
	return v
}

func (n *Nodis) LPopRPush(source, destination string) []byte {
	meta := n.store.writeKey(source, nil)
	if !meta.isOk() {
		meta.commit()
		return nil
	}
	v := meta.ds.(*list.DoublyLinkedList).LPop(1)
	if v == nil {
		return nil
	}
	if meta.ds.(*list.DoublyLinkedList).LLen() == 0 {
		n.store.delKey(source)
	}
	tx2 := n.store.writeKey(destination, n.newList)
	tx2.ds.(*list.DoublyLinkedList).RPush(v...)
	tx2.commit()
	meta.commit()
	n.notify(pb.NewOp(pb.OpType_LPopRPush, source).DstKey(destination))
	return v[0]
}

func (n *Nodis) RPopLPush(source, destination string) []byte {
	meta := n.store.writeKey(source, nil)
	if !meta.isOk() {
		meta.commit()
		return nil
	}
	v := meta.ds.(*list.DoublyLinkedList).RPop(1)
	if v == nil {
		return nil
	}
	if meta.ds.(*list.DoublyLinkedList).LLen() == 0 {
		n.store.delKey(source)
	}
	tx2 := n.store.writeKey(destination, n.newList)
	tx2.ds.(*list.DoublyLinkedList).LPush(v...)
	tx2.commit()
	meta.commit()
	n.notify(pb.NewOp(pb.OpType_RPopLPush, source).DstKey(destination))
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
