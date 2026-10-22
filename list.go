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
		n := n.inTx(tx)
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
		n := n.inTx(tx)
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
		n := n.inTx(tx)
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
		n := n.inTx(tx)
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
		n := n.inTx(tx)
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
		n := n.inTx(tx)
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
		n := n.inTx(tx)
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
		n := n.inTx(tx)
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
		n := n.inTx(tx)
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
		n := n.inTx(tx)
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
		n := n.inTx(tx)
		meta := tx.writeKey(source, nil)
		if !meta.isOk() {

			return nil
		}
		sourceList := meta.value.(*list.LinkedList)
		if sourceList.LLen() == 0 {
			return nil
		}
		dst := tx.writeKey(destination, n.newList)
		_ = dst.value.(*list.LinkedList)
		v = meta.value.(*list.LinkedList).LPop(1)
		if v == nil {
			return nil
		}
		if meta.value.(*list.LinkedList).LLen() == 0 {
			tx.delKey(source)
		}
		n.signalModifiedKey(source, meta)
		dst.value.(*list.LinkedList).RPush(v...)
		n.notifyBlockingKey(destination)
		tx.storeMeta(dst)
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
		n := n.inTx(tx)
		meta := tx.writeKey(source, nil)
		if !meta.isOk() {
			return nil
		}
		sourceList := meta.value.(*list.LinkedList)
		if sourceList.LLen() == 0 {
			return nil
		}
		dst := tx.writeKey(destination, n.newList)
		_ = dst.value.(*list.LinkedList)
		v = meta.value.(*list.LinkedList).RPop(1)
		if v == nil {
			return nil
		}
		if meta.value.(*list.LinkedList).LLen() == 0 {
			tx.delKey(source)
		}
		n.signalModifiedKey(source, meta)
		dst.value.(*list.LinkedList).LPush(v...)
		n.notifyBlockingKey(destination)
		tx.storeMeta(dst)
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

// Wakes are hints, never ownership transfers. Consumers recheck under store.mu.
func (n *Nodis) notifyBlockingKey(key string) {
	n.blockingKeysMutex.Lock()
	defer n.blockingKeysMutex.Unlock()
	for c := range n.blockingKeys[key] {
		select {
		case c <- struct{}{}:
		default:
		}
	}
}
func (n *Nodis) BLPop(timeout time.Duration, keys ...string) (string, []byte) {
	return n.blockPop(timeout, false, keys...)
}
func (n *Nodis) BRPop(timeout time.Duration, keys ...string) (string, []byte) {
	return n.blockPop(timeout, true, keys...)
}
func (n *Nodis) blockPop(timeout time.Duration, right bool, keys ...string) (string, []byte) {
	wake := make(chan struct{}, 1)
	defer func() {
		n.blockingKeysMutex.Lock()
		defer n.blockingKeysMutex.Unlock()
		for _, key := range keys {
			delete(n.blockingKeys[key], wake)
			if len(n.blockingKeys[key]) == 0 {
				delete(n.blockingKeys, key)
			}
		}
	}()
	var timer *time.Timer
	var deadline <-chan time.Time
	if timeout > 0 {
		timer = time.NewTimer(timeout)
		deadline = timer.C
		defer timer.Stop()
	}
	for {
		var found string
		var value []byte
		var ok bool
		err := n.exec(func(tx *Tx) error {
			scoped := n.inTx(tx)
			for _, key := range keys {
				var values [][]byte
				if right {
					values = scoped.RPop(key, 1)
				} else {
					values = scoped.LPop(key, 1)
				}
				if len(values) > 0 {
					found = key
					value = values[0]
					ok = true
					return nil
				}
			}
			if n.tx == nil {
				n.blockingKeysMutex.Lock()
				for _, key := range keys {
					waiters := n.blockingKeys[key]
					if waiters == nil {
						waiters = make(map[chan struct{}]struct{})
						n.blockingKeys[key] = waiters
					}
					waiters[wake] = struct{}{}
				}
				n.blockingKeysMutex.Unlock()
			}
			return nil
		})
		if ok {
			return found, value
		}
		if err != nil || n.tx != nil || timeout < 0 {
			return "", nil
		}
		select {
		case <-wake:
		case <-deadline:
			return "", nil
		case <-n.done:
			return "", nil
		}
	}
}
