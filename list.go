package nodis

import (
	"time"

	"github.com/diiyw/nodis/ds"

	"github.com/diiyw/nodis/ds/list"
)

// newList creates a new list
func (n *Nodis) newList() ds.DataStruct {
	return list.NewDoublyLinkedList()
}

func (n *Nodis) LPush(key string, values ...[]byte) {
	l := n.getDs(key, n.newList, 0)
	l.(*list.DoublyLinkedList).LPush(values...)
}

func (n *Nodis) RPush(key string, values ...[]byte) {
	l := n.getDs(key, n.newList, 0)
	for _, v := range values {
		l.(*list.DoublyLinkedList).RPush(v)
	}
}

func (n *Nodis) LPop(key string) []byte {
	l := n.getDs(key, nil, 0)
	if l == nil {
		return nil
	}
	v := l.(*list.DoublyLinkedList).LPop()
	if l.(*list.DoublyLinkedList).LLen() == 0 {
		n.Del(key)
	}
	return v
}

func (n *Nodis) RPop(key string) []byte {
	l := n.getDs(key, nil, 0)
	if l == nil {
		return nil
	}
	v := l.(*list.DoublyLinkedList).RPop()
	if l.(*list.DoublyLinkedList).LLen() == 0 {
		n.Del(key)
	}
	return v
}

func (n *Nodis) LLen(key string) int {
	l := n.getDs(key, nil, 0)
	if l == nil {
		return 0
	}
	return l.(*list.DoublyLinkedList).LLen()
}

func (n *Nodis) LIndex(key string, index int) ([]byte, bool) {
	l := n.getDs(key, nil, 0)
	if l == nil {
		return nil, false
	}
	return l.(*list.DoublyLinkedList).LIndex(index)
}

func (n *Nodis) LInsert(key string, pivot, data []byte, before bool) int {
	l := n.getDs(key, n.newList, 0)
	return l.(*list.DoublyLinkedList).LInsert(pivot, data, before)
}

func (n *Nodis) LPushX(key string, data []byte) int {
	l := n.getDs(key, n.newList, 0)
	return l.(*list.DoublyLinkedList).LPushX(data)
}

func (n *Nodis) RPushX(key string, data []byte) int {
	l := n.getDs(key, n.newList, 0)
	return l.(*list.DoublyLinkedList).RPushX(data)
}

func (n *Nodis) LRem(key string, count int, data []byte) int {
	l := n.getDs(key, nil, 0)
	if l == nil {
		return 0
	}
	return l.(*list.DoublyLinkedList).LRem(count, data)
}

func (n *Nodis) LSet(key string, index int, data []byte) bool {
	l := n.getDs(key, n.newList, 0)
	return l.(*list.DoublyLinkedList).LSet(index, data)
}

func (n *Nodis) LTrim(key string, start, stop int) {
	l := n.getDs(key, nil, 0)
	if l == nil {
		return
	}
	l.(*list.DoublyLinkedList).LTrim(start, stop)
}

func (n *Nodis) LRange(key string, start, stop int) [][]byte {
	l := n.getDs(key, nil, 0)
	if l == nil {
		return nil
	}
	return l.(*list.DoublyLinkedList).LRange(start, stop)
}

func (n *Nodis) LPopRPush(source, destination string) []byte {
	l := n.getDs(source, nil, 0)
	if l == nil {
		return nil
	}
	v := l.(*list.DoublyLinkedList).LPop()
	if l.(*list.DoublyLinkedList).LLen() == 0 {
		n.store.Delete(source)
		n.keys.Delete(source)
	}
	if !n.exists(destination) {
		n.store.Put(destination, list.NewDoublyLinkedList())
		n.keys.Put(destination, newKey(ds.List, 0))
	}
	l, _ = n.store.Get(destination)
	l.(*list.DoublyLinkedList).RPush(v)
	return v
}

func (n *Nodis) RPopLPush(source, destination string) []byte {
	l := n.getDs(source, nil, 0)
	if l == nil {
		return nil
	}
	v := l.(*list.DoublyLinkedList).RPop()
	if l.(*list.DoublyLinkedList).LLen() == 0 {
		n.store.Delete(source)
		n.keys.Delete(source)
	}
	if !n.exists(destination) {
		n.store.Put(destination, list.NewDoublyLinkedList())
		n.keys.Put(destination, newKey(ds.List, 0))
	}
	l, _ = n.store.Get(destination)
	l.(*list.DoublyLinkedList).LPush(v)
	return v
}

func (n *Nodis) BLPop(key string, timeout time.Duration) []byte {
	l := n.getDs(key, nil, 0)
	if l == nil {
		return nil
	}
	v := l.(*list.DoublyLinkedList).BLPop(timeout)
	if l.(*list.DoublyLinkedList).LLen() == 0 {
		n.store.Delete(key)
		n.keys.Delete(key)
	}
	return v
}

func (n *Nodis) BRPop(key string, timeout time.Duration) []byte {
	l := n.getDs(key, nil, 0)
	if l == nil {
		return nil
	}
	v := l.(*list.DoublyLinkedList).BRPop(timeout)
	if l.(*list.DoublyLinkedList).LLen() == 0 {
		n.store.Delete(key)
		n.keys.Delete(key)
	}
	return v
}
