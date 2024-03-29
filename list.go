package nodis

import (
	"github.com/diiyw/nodis/pb"
	"time"

	"github.com/diiyw/nodis/ds"

	"github.com/diiyw/nodis/ds/list"
)

// newList creates a new list
func (n *Nodis) newList() ds.DataStruct {
	return list.NewDoublyLinkedList()
}

func (n *Nodis) LPush(key string, values ...[]byte) {
	k, l := n.getDs(key, n.newList, 0)
	k.changed.Store(true)
	l.(*list.DoublyLinkedList).LPush(values...)
	n.notify(pb.NewOp(pb.OpType_LPush, key).Values(values))
}

func (n *Nodis) RPush(key string, values ...[]byte) {
	k, l := n.getDs(key, n.newList, 0)
	k.changed.Store(true)
	for _, v := range values {
		l.(*list.DoublyLinkedList).RPush(v)
	}
	n.notify(pb.NewOp(pb.OpType_RPush, key).Values(values))
}

func (n *Nodis) LPop(key string) []byte {
	k, l := n.getDs(key, nil, 0)
	if l == nil {
		return nil
	}
	k.changed.Store(true)
	v := l.(*list.DoublyLinkedList).LPop()
	if l.(*list.DoublyLinkedList).LLen() == 0 {
		n.Del(key)
	}
	n.notify(pb.NewOp(pb.OpType_LPop, key))
	return v
}

func (n *Nodis) RPop(key string) []byte {
	k, l := n.getDs(key, nil, 0)
	if l == nil {
		return nil
	}
	k.changed.Store(true)
	v := l.(*list.DoublyLinkedList).RPop()
	if l.(*list.DoublyLinkedList).LLen() == 0 {
		n.Del(key)
	}
	n.notify(pb.NewOp(pb.OpType_RPop, key))
	return v
}

func (n *Nodis) LLen(key string) int {
	_, l := n.getDs(key, nil, 0)
	if l == nil {
		return 0
	}
	return l.(*list.DoublyLinkedList).LLen()
}

func (n *Nodis) LIndex(key string, index int) []byte {
	_, l := n.getDs(key, nil, 0)
	if l == nil {
		return nil
	}
	return l.(*list.DoublyLinkedList).LIndex(index)
}

func (n *Nodis) LInsert(key string, pivot, data []byte, before bool) int {
	k, l := n.getDs(key, n.newList, 0)
	k.changed.Store(true)
	n.notify(pb.NewOp(pb.OpType_LInsert, key).Value(data).Pivot(pivot).Before(before))
	return l.(*list.DoublyLinkedList).LInsert(pivot, data, before)
}

func (n *Nodis) LPushX(key string, data []byte) int {
	k, l := n.getDs(key, n.newList, 0)
	k.changed.Store(true)
	n.notify(pb.NewOp(pb.OpType_LPushX, key).Value(data))
	return l.(*list.DoublyLinkedList).LPushX(data)
}

func (n *Nodis) RPushX(key string, data []byte) int {
	k, l := n.getDs(key, n.newList, 0)
	k.changed.Store(true)
	n.notify(pb.NewOp(pb.OpType_RPushX, key).Value(data))
	return l.(*list.DoublyLinkedList).RPushX(data)
}

func (n *Nodis) LRem(key string, count int64, data []byte) int64 {
	k, l := n.getDs(key, nil, 0)
	if l == nil {
		return 0
	}
	k.changed.Store(true)
	n.notify(pb.NewOp(pb.OpType_LRem, key).Value(data).Count(count))
	return l.(*list.DoublyLinkedList).LRem(count, data)
}

func (n *Nodis) LSet(key string, index int64, data []byte) bool {
	k, l := n.getDs(key, n.newList, 0)
	k.changed.Store(true)
	n.notify(pb.NewOp(pb.OpType_LSet, key).Value(data).Index(index))
	return l.(*list.DoublyLinkedList).LSet(index, data)
}

func (n *Nodis) LTrim(key string, start, stop int64) {
	k, l := n.getDs(key, nil, 0)
	if l == nil {
		return
	}
	k.changed.Store(true)
	l.(*list.DoublyLinkedList).LTrim(start, stop)
}

func (n *Nodis) LRange(key string, start, stop int) [][]byte {
	_, l := n.getDs(key, nil, 0)
	if l == nil {
		return nil
	}
	return l.(*list.DoublyLinkedList).LRange(start, stop)
}

func (n *Nodis) LPopRPush(source, destination string) []byte {
	k, l := n.getDs(source, nil, 0)
	if l == nil {
		return nil
	}
	k.changed.Store(true)
	v := l.(*list.DoublyLinkedList).LPop()
	if l.(*list.DoublyLinkedList).LLen() == 0 {
		n.Del(source)
	}
	destinationKey, ok := n.exists(destination)
	if !ok {
		n.dataStructs.Set(destination, list.NewDoublyLinkedList())
		destinationKey = newKey(ds.List, 0)
		n.keys.Set(destination, destinationKey)
	}
	destinationKey.changed.Store(true)
	l, _ = n.dataStructs.Get(destination)
	l.(*list.DoublyLinkedList).RPush(v)
	n.notify(pb.NewOp(pb.OpType_LPopRPush, source).DstKey(destination))
	return v
}

func (n *Nodis) RPopLPush(source, destination string) []byte {
	k, l := n.getDs(source, nil, 0)
	if l == nil {
		return nil
	}
	k.changed.Store(true)
	v := l.(*list.DoublyLinkedList).RPop()
	if l.(*list.DoublyLinkedList).LLen() == 0 {
		n.Del(source)
	}
	destinationKey, ok := n.exists(destination)
	if !ok {
		n.dataStructs.Set(destination, list.NewDoublyLinkedList())
		destinationKey = newKey(ds.List, 0)
		n.keys.Set(destination, destinationKey)
	}
	destinationKey.changed.Store(true)
	l, _ = n.dataStructs.Get(destination)
	l.(*list.DoublyLinkedList).LPush(v)
	n.notify(pb.NewOp(pb.OpType_RPopLPush, source).DstKey(destination))
	return v
}

func (n *Nodis) BLPop(key string, timeout time.Duration) []byte {
	k, l := n.getDs(key, n.newList, 0)
	v := l.(*list.DoublyLinkedList).BLPop(timeout)
	k.changed.Store(true)
	if l.(*list.DoublyLinkedList).LLen() == 0 {
		n.Del(key)
	}
	n.notify(pb.NewOp(pb.OpType_LPop, key))
	return v
}

func (n *Nodis) BRPop(key string, timeout time.Duration) []byte {
	k, l := n.getDs(key, n.newList, 0)
	v := l.(*list.DoublyLinkedList).BRPop(timeout)
	k.changed.Store(true)
	if l.(*list.DoublyLinkedList).LLen() == 0 {
		n.Del(key)
	}
	n.notify(pb.NewOp(pb.OpType_RPop, key))
	return v
}
