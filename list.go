package nodis

import (
	"time"

	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/patch"

	"github.com/diiyw/nodis/ds/list"
)

// newList creates a new list
func (n *Nodis) newList() ds.Value {
	return list.NewLinkedList()
}

func (n *Nodis) LPush(key string, values ...[]byte) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, n.newList)
		meta.value.(*list.LinkedList).LPush(values...)
		v = meta.value.(*list.LinkedList).LLen()
		n.notifyBlockingKey(key)
		n.signalModifiedKey(key, meta)
		n.notify(func() []patch.Op {
			return []patch.Op{{Type: patch.OpTypeLPush, Data: &patch.OpLPush{Key: key, Values: values}}}
		})
		return nil
	})
	return v
}

func (n *Nodis) RPush(key string, values ...[]byte) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, n.newList)
		meta.value.(*list.LinkedList).RPush(values...)
		v = meta.value.(*list.LinkedList).LLen()
		n.notifyBlockingKey(key)
		n.signalModifiedKey(key, meta)
		n.notify(func() []patch.Op {
			return []patch.Op{{Type: patch.OpTypeRPush, Data: &patch.OpRPush{Key: key, Values: values}}}
		})
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
		v = meta.value.(*list.LinkedList).LPop(count)
		if meta.value.(*list.LinkedList).LLen() == 0 {
			tx.delKey(key)
		}
		n.signalModifiedKey(key, meta)
		n.notify(func() []patch.Op {
			return []patch.Op{{Type: patch.OpTypeLPop, Data: &patch.OpLPop{Key: key, Count: count}}}
		})
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
		v = meta.value.(*list.LinkedList).RPop(count)
		if meta.value.(*list.LinkedList).LLen() == 0 {
			tx.delKey(key)
		}
		n.signalModifiedKey(key, meta)
		n.notify(func() []patch.Op {
			return []patch.Op{{Type: patch.OpTypeRPop, Data: &patch.OpRPop{Key: key, Count: count}}}
		})
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
		d, ok := meta.value.(*list.LinkedList)
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
		v = meta.value.(*list.LinkedList).LIndex(index)
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
		v = meta.value.(*list.LinkedList).LInsert(pivot, data, before)
		n.signalModifiedKey(key, meta)
		n.notify(func() []patch.Op {
			return []patch.Op{{Type: patch.OpTypeLInsert, Data: &patch.OpLInsert{Key: key, Value: data, Pivot: pivot, Before: before}}}
		})
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
		meta.value.(*list.LinkedList).LPush(data)
		v = meta.value.(*list.LinkedList).LLen()
		n.signalModifiedKey(key, meta)
		n.notify(func() []patch.Op {
			return []patch.Op{{Type: patch.OpTypeLPushX, Data: &patch.OpLPushX{Key: key, Value: data}}}
		})
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
		meta.value.(*list.LinkedList).RPush(data)
		v = meta.value.(*list.LinkedList).LLen()
		n.signalModifiedKey(key, meta)
		n.notify(func() []patch.Op {
			return []patch.Op{{Type: patch.OpTypeRPushX, Data: &patch.OpRPushX{Key: key, Value: data}}}
		})
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
		dv := meta.value.(*list.LinkedList)
		v = dv.LRem(count, data)
		if dv.LLen() == 0 {
			tx.delKey(key)
		}
		n.signalModifiedKey(key, meta)
		n.notify(func() []patch.Op {
			return []patch.Op{{Type: patch.OpTypeLRem, Data: &patch.OpLRem{Key: key, Value: data, Count: count}}}
		})
		return nil
	})
	return v
}

func (n *Nodis) LSet(key string, index int64, data []byte) bool {
	var v bool
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, n.newList)
		n.signalModifiedKey(key, meta)
		n.notify(func() []patch.Op {
			return []patch.Op{{Type: patch.OpTypeLSet, Data: &patch.OpLSet{Key: key, Value: data, Index: index}}}
		})
		v = meta.value.(*list.LinkedList).LSet(index, data)
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
		meta.value.(*list.LinkedList).LTrim(start, stop)
		n.signalModifiedKey(key, meta)
		n.notify(func() []patch.Op {
			return []patch.Op{{Type: patch.OpTypeLTrim, Data: &patch.OpLTrim{Key: key, Start: start, Stop: stop}}}
		})
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
		v = meta.value.(*list.LinkedList).LRange(start, stop)
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
		v = meta.value.(*list.LinkedList).LPop(1)
		if v == nil {
			return nil
		}
		if meta.value.(*list.LinkedList).LLen() == 0 {
			tx.delKey(source)
		}
		n.signalModifiedKey(source, meta)
		dst := tx.writeKey(destination, n.newList)
		dst.value.(*list.LinkedList).RPush(v...)
		n.notifyBlockingKey(destination)
		n.signalModifiedKey(destination, dst)
		n.notify(func() []patch.Op {
			return []patch.Op{{Type: patch.OpTypeLPopRPush, Data: &patch.OpLPopRPush{Key: source, DstKey: destination}}}
		})
		return nil
	})
	if len(v) == 0 {
		return nil
	}
	return v[0]
}

func (n *Nodis) RPopLPush(source, destination string) []byte {
	var v [][]byte
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(source, nil)
		if !meta.isOk() {
			return nil
		}
		v = meta.value.(*list.LinkedList).RPop(1)
		if v == nil {
			return nil
		}
		if meta.value.(*list.LinkedList).LLen() == 0 {
			tx.delKey(source)
		}
		n.signalModifiedKey(source, meta)
		dst := tx.writeKey(destination, n.newList)
		dst.value.(*list.LinkedList).LPush(v...)
		n.notifyBlockingKey(destination)
		n.signalModifiedKey(destination, dst)
		n.notify(func() []patch.Op {
			return []patch.Op{{Type: patch.OpTypeRPopLPush, Data: &patch.OpRPopLPush{Key: source, DstKey: destination}}}
		})
		return nil
	})
	if len(v) == 0 {
		return nil
	}
	return v[0]
}

func (n *Nodis) addBlockKey(key string, c chan string) {
	n.blockingKeysMutex.Lock()
	cList, ok := n.blockingKeys[key]
	if !ok {
		cList = list.NewLinkedListG[chan string]()
		cList.LPush(c)
		n.blockingKeys[key] = cList
	} else {
		cList.LPush(c)
	}
	n.blockingKeysMutex.Unlock()
}

func (n *Nodis) notifyBlockingKey(key string) {
	n.blockingKeysMutex.RLock()
	cList, ok := n.blockingKeys[key]
	n.blockingKeysMutex.RUnlock()
	if !ok {
		return
	}
	cList.ForRange(func(c chan string) bool {
		c <- key
		return true
	})
}

func (n *Nodis) removeBlockingKeys(rc chan string, keys ...string) {
	n.blockingKeysMutex.Lock()
	for _, key := range keys {
		cList, ok := n.blockingKeys[key]
		if !ok {
			n.blockingKeysMutex.Unlock()
			return
		}
		cList.ForRangeNode(func(node *list.NodeG[chan string]) bool {
			if node.Value() == rc {
				cList.RemoveNode(node)
				return false
			}
			return true
		})
	}
	close(rc)
	n.blockingKeysMutex.Unlock()
}

func (n *Nodis) BLPop(timeout time.Duration, keys ...string) (string, []byte) {
	var c = make(chan string)
	defer n.removeBlockingKeys(c, keys...)
	for _, key := range keys {
		results := n.LPop(key, 1)
		if results != nil {
			n.notify(func() []patch.Op {
				return []patch.Op{{Type: patch.OpTypeLPop, Data: &patch.OpLPop{Key: key}}}
			})
			return key, results[0]
		}
		n.addBlockKey(key, c)
	}
	select {
	case key := <-c:
		results := n.LPop(key, 1)
		if results != nil {
			n.notify(func() []patch.Op {
				return []patch.Op{{Type: patch.OpTypeLPop, Data: &patch.OpLPop{Key: key}}}
			})
			return key, results[0]
		}
	case <-time.After(timeout):
		break
	}
	return "", nil
}

func (n *Nodis) BRPop(timeout time.Duration, keys ...string) (string, []byte) {
	var c = make(chan string)
	defer n.removeBlockingKeys(c, keys...)
	for _, key := range keys {
		results := n.RPop(key, 1)
		if results != nil {
			n.notify(func() []patch.Op {
				return []patch.Op{{Type: patch.OpTypeRPop, Data: &patch.OpRPop{Key: key}}}
			})
			return key, results[0]
		}
		n.addBlockKey(key, c)
	}
	select {
	case key := <-c:
		results := n.RPop(key, 1)
		if results != nil {
			n.notify(func() []patch.Op {
				return []patch.Op{{Type: patch.OpTypeRPop, Data: &patch.OpRPop{Key: key}}}
			})
			return key, results[0]
		}
	case <-time.After(timeout):
		break
	}
	return "", nil
}
