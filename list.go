package nodis

import (
	"time"

	"github.com/diiyw/nodis/pb"

	"github.com/diiyw/nodis/ds"

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
		n.notify(func() []*pb.Op {
			return []*pb.Op{pb.NewOp(pb.OpType_LPush, key).Values(values)}
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
		n.notify(func() []*pb.Op {
			return []*pb.Op{pb.NewOp(pb.OpType_RPush, key).Values(values)}
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
		n.notify(func() []*pb.Op {
			return []*pb.Op{pb.NewOp(pb.OpType_LPop, key).Count(count)}
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
		n.notify(func() []*pb.Op {
			return []*pb.Op{pb.NewOp(pb.OpType_RPop, key).Count(count)}
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
		n.notify(func() []*pb.Op {
			return []*pb.Op{pb.NewOp(pb.OpType_LInsert, key).Value(data).Pivot(pivot).Before(before)}
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
		n.notify(func() []*pb.Op {
			return []*pb.Op{pb.NewOp(pb.OpType_LPushX, key).Value(data)}
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
		n.notify(func() []*pb.Op {
			return []*pb.Op{pb.NewOp(pb.OpType_RPushX, key).Value(data)}
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
		ds := meta.value.(*list.LinkedList)
		v = ds.LRem(count, data)
		if ds.LLen() == 0 {
			tx.delKey(key)
		}
		n.signalModifiedKey(key, meta)
		n.notify(func() []*pb.Op {
			return []*pb.Op{pb.NewOp(pb.OpType_LRem, key).Value(data).Count(count)}
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
		n.notify(func() []*pb.Op {
			return []*pb.Op{pb.NewOp(pb.OpType_LSet, key).Value(data).Index(index)}
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
		n.notify(func() []*pb.Op {
			return []*pb.Op{pb.NewOp(pb.OpType_LTrim, key).Start(start).Stop(stop)}
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
		n.notify(func() []*pb.Op {
			return []*pb.Op{pb.NewOp(pb.OpType_LPopRPush, source).DstKey(destination)}
		})
		return nil
	})
	return v[0]
}

func (n *Nodis) RPopLPush(source, destination string) []byte {
	var v = make([][]byte, 0)
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
		n.notify(func() []*pb.Op {
			return []*pb.Op{pb.NewOp(pb.OpType_RPopLPush, source).DstKey(destination)}
		})
		return nil
	})
	return v[0]
}

func (n *Nodis) addBlockKey(key string, c chan string) {
	n.blocklingKeysMutex.Lock()
	cList, ok := n.blocklingKeys.Get(key)
	if !ok {
		cList = list.NewLinkedListG[chan string]()
		cList.LPush(c)
		n.blocklingKeys.Set(key, cList)
	} else {
		cList.LPush(c)
	}
	n.blocklingKeysMutex.Unlock()
}

func (n *Nodis) notifyBlockingKey(key string) {
	n.blocklingKeysMutex.RLock()
	cList, ok := n.blocklingKeys.Get(key)
	n.blocklingKeysMutex.RUnlock()
	if !ok {
		return
	}
	cList.ForRange(func(c chan string) bool {
		c <- key
		return true
	})
}

func (n *Nodis) removeBlockingKeys(rc chan string, keys ...string) {
	n.blocklingKeysMutex.Lock()
	for _, key := range keys {
		cList, ok := n.blocklingKeys.Get(key)
		if !ok {
			n.blocklingKeysMutex.Unlock()
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
	n.blocklingKeysMutex.Unlock()
}

func (n *Nodis) BLPop(timeout time.Duration, keys ...string) (string, []byte) {
	var c = make(chan string)
	defer n.removeBlockingKeys(c, keys...)
	for _, key := range keys {
		results := n.LPop(key, 1)
		if results != nil {
			n.notify(func() []*pb.Op {
				return []*pb.Op{pb.NewOp(pb.OpType_LPop, key)}
			})
			return key, results[0]
		}
		n.addBlockKey(key, c)
	}
	select {
	case key := <-c:
		results := n.LPop(key, 1)
		if results != nil {
			n.notify(func() []*pb.Op {
				return []*pb.Op{pb.NewOp(pb.OpType_LPop, key)}
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
			n.notify(func() []*pb.Op {
				return []*pb.Op{pb.NewOp(pb.OpType_RPop, key)}
			})
			return key, results[0]
		}
		n.addBlockKey(key, c)
	}
	select {
	case key := <-c:
		results := n.LPop(key, 1)
		if results != nil {
			n.notify(func() []*pb.Op {
				return []*pb.Op{pb.NewOp(pb.OpType_LPop, key)}
			})
			return key, results[0]
		}
	case <-time.After(timeout):
		break
	}
	return "", nil
}
