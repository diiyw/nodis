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
	tx := n.writeKey(key, n.newList)
	tx.ds.(*list.DoublyLinkedList).LPush(values...)
	tx.commit()
	n.notify(pb.NewOp(pb.OpType_LPush, key).Values(values))
	return int64(len(values))
}

func (n *Nodis) RPush(key string, values ...[]byte) int64 {
	tx := n.writeKey(key, n.newList)
	for _, v := range values {
		tx.ds.(*list.DoublyLinkedList).RPush(v)
	}
	tx.commit()
	n.notify(pb.NewOp(pb.OpType_RPush, key).Values(values))
	return int64(len(values))
}

func (n *Nodis) LPop(key string, count int64) [][]byte {
	tx := n.writeKey(key, nil)
	if !tx.isOk() {
		return nil
	}
	v := tx.ds.(*list.DoublyLinkedList).LPop(count)
	if tx.ds.(*list.DoublyLinkedList).LLen() == 0 {
		n.delKey(key)
	}
	tx.commit()
	n.notify(pb.NewOp(pb.OpType_LPop, key).Count(count))
	return v
}

func (n *Nodis) RPop(key string, count int64) [][]byte {
	tx := n.writeKey(key, nil)
	if !tx.isOk() {
		return nil
	}
	v := tx.ds.(*list.DoublyLinkedList).RPop(count)
	if tx.ds.(*list.DoublyLinkedList).LLen() == 0 {
		n.delKey(key)
	}
	tx.commit()
	n.notify(pb.NewOp(pb.OpType_RPop, key).Count(count))
	return v
}

func (n *Nodis) LLen(key string) int64 {
	tx := n.readKey(key)
	if !tx.isOk() {
		return 0
	}
	v := tx.ds.(*list.DoublyLinkedList).LLen()
	tx.commit()
	return v
}

func (n *Nodis) LIndex(key string, index int64) []byte {
	tx := n.readKey(key)
	if !tx.isOk() {
		return nil
	}
	v := tx.ds.(*list.DoublyLinkedList).LIndex(index)
	tx.commit()
	return v
}

func (n *Nodis) LInsert(key string, pivot, data []byte, before bool) int64 {
	tx := n.writeKey(key, n.newList)
	v := tx.ds.(*list.DoublyLinkedList).LInsert(pivot, data, before)
	tx.commit()
	n.notify(pb.NewOp(pb.OpType_LInsert, key).Value(data).Pivot(pivot).Before(before))
	return v
}

func (n *Nodis) LPushX(key string, data []byte) int64 {
	tx := n.writeKey(key, n.newList)
	v := tx.ds.(*list.DoublyLinkedList).LPushX(data)
	tx.commit()
	n.notify(pb.NewOp(pb.OpType_LPushX, key).Value(data))
	return v
}

func (n *Nodis) RPushX(key string, data []byte) int64 {
	tx := n.writeKey(key, n.newList)
	v := tx.ds.(*list.DoublyLinkedList).RPushX(data)
	tx.commit()
	n.notify(pb.NewOp(pb.OpType_RPushX, key).Value(data))
	return v
}

func (n *Nodis) LRem(key string, count int64, data []byte) int64 {
	tx := n.writeKey(key, nil)
	if !tx.isOk() {
		return 0
	}
	v := tx.ds.(*list.DoublyLinkedList).LRem(count, data)
	tx.commit()
	n.notify(pb.NewOp(pb.OpType_LRem, key).Value(data).Count(count))
	return v
}

func (n *Nodis) LSet(key string, index int64, data []byte) bool {
	tx := n.writeKey(key, n.newList)
	n.notify(pb.NewOp(pb.OpType_LSet, key).Value(data).Index(index))
	v := tx.ds.(*list.DoublyLinkedList).LSet(index, data)
	tx.commit()
	return v
}

func (n *Nodis) LTrim(key string, start, stop int64) {
	tx := n.writeKey(key, nil)
	if !tx.isOk() {
		return
	}
	tx.ds.(*list.DoublyLinkedList).LTrim(start, stop)
	tx.commit()
}

func (n *Nodis) LRange(key string, start, stop int64) [][]byte {
	tx := n.readKey(key)
	if !tx.isOk() {
		return nil
	}
	v := tx.ds.(*list.DoublyLinkedList).LRange(start, stop)
	tx.commit()
	return v
}

func (n *Nodis) LPopRPush(source, destination string) []byte {
	tx := n.writeKey(source, nil)
	if !tx.isOk() {
		return nil
	}
	v := tx.ds.(*list.DoublyLinkedList).LPop(1)
	if v == nil {
		return nil
	}
	if tx.ds.(*list.DoublyLinkedList).LLen() == 0 {
		n.delKey(source)
	}
	tx2 := n.writeKey(destination, n.newList)
	tx2.ds.(*list.DoublyLinkedList).RPush(v...)
	tx2.commit()
	tx.commit()
	n.notify(pb.NewOp(pb.OpType_LPopRPush, source).DstKey(destination))
	return v[0]
}

func (n *Nodis) RPopLPush(source, destination string) []byte {
	tx := n.writeKey(source, nil)
	if !tx.isOk() {
		return nil
	}
	v := tx.ds.(*list.DoublyLinkedList).RPop(1)
	if v == nil {
		return nil
	}
	if tx.ds.(*list.DoublyLinkedList).LLen() == 0 {
		n.delKey(source)
	}
	tx2 := n.writeKey(destination, n.newList)
	tx2.ds.(*list.DoublyLinkedList).LPush(v...)
	tx2.commit()
	tx.commit()
	n.notify(pb.NewOp(pb.OpType_RPopLPush, source).DstKey(destination))
	return v[0]
}

func (n *Nodis) BLPop(key string, timeout time.Duration) []byte {
	tx := n.writeKey(key, n.newList)
	v := tx.ds.(*list.DoublyLinkedList).BLPop(timeout)
	if tx.ds.(*list.DoublyLinkedList).LLen() == 0 {
		n.delKey(key)
	}
	tx.commit()
	n.notify(pb.NewOp(pb.OpType_LPop, key))
	return v
}

func (n *Nodis) BRPop(key string, timeout time.Duration) []byte {
	tx := n.writeKey(key, n.newList)
	v := tx.ds.(*list.DoublyLinkedList).BRPop(timeout)
	if tx.ds.(*list.DoublyLinkedList).LLen() == 0 {
		n.delKey(key)
	}
	tx.commit()
	n.notify(pb.NewOp(pb.OpType_RPop, key))
	return v
}
