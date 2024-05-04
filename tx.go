package nodis

import (
	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/utils"
	"time"
)

type Tx struct {
	store       *store
	lockedMetas []*metadata
}

func newTx(store *store) *Tx {
	return &Tx{
		store:       store,
		lockedMetas: make([]*metadata, 0),
	}
}

func (tx *Tx) spread(hashCode uint32) *metadata {
	tableSize := uint32(tx.store.metaPoolSize)
	meta := tx.store.metaPool[hashCode&tableSize]
	tx.lockedMetas = append(tx.lockedMetas, meta)
	return meta
}

func (tx *Tx) writeKey(key string, newFn func() ds.DataStruct) *metadata {
	meta := tx.spread(utils.Fnv32(key))
	meta.Lock()
	meta.writeable = true
	tx.store.mu.Lock()
	defer tx.store.mu.Unlock()
	k, ok := tx.store.keys.Get(key)
	if ok {
		if k.expired(time.Now().UnixMilli()) {
			if newFn == nil {
				return meta
			}
		}
		d, ok := tx.store.values.Get(key)
		if ok {
			meta.set(k, d, true)
			return meta
		}
		meta = tx.store.fromStorage(k, meta, true)
		return meta
	}
	if newFn != nil {
		d := newFn()
		if d == nil {
			return meta
		}
		k = newKey()
		tx.store.keys.Set(key, k)
		tx.store.values.Set(key, d)
		meta.set(k, d, true)
		return meta
	}
	return meta
}

func (tx *Tx) readKey(key string) *metadata {
	meta := tx.spread(utils.Fnv32(key))
	meta.RLock()
	tx.store.mu.RLock()
	defer tx.store.mu.RUnlock()
	k, ok := tx.store.keys.Get(key)
	if ok {
		if k.expired(time.Now().UnixMilli()) {
			return meta
		}
		d, ok := tx.store.values.Get(key)
		if !ok {
			// read from storage
			return tx.store.fromStorage(k, meta, false)
		}
		meta.set(k, d, false)
		return meta
	}
	return meta
}

func (tx *Tx) commit() {
	for _, meta := range tx.lockedMetas {
		meta.commit()
	}
	tx.lockedMetas = tx.lockedMetas[:0]
}
