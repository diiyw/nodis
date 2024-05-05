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

func (tx *Tx) spread(key string) *metadata {
	hashCode := utils.Fnv32(key)
	tableSize := uint32(tx.store.metaPoolSize)
	return tx.store.metaPool[(tableSize-1)&hashCode]
}

func (tx *Tx) lockKey(key string) *metadata {
	meta := tx.spread(key)
	meta.Lock()
	tx.lockedMetas = append(tx.lockedMetas, meta)
	return meta
}

func (tx *Tx) rLockKey(key string) *metadata {
	meta := tx.spread(key)
	meta.RLock()
	tx.lockedMetas = append(tx.lockedMetas, meta)
	return meta
}

func (tx *Tx) writeKey(key string, newFn func() ds.DataStruct) *metadata {
	meta := tx.lockKey(key)
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
			meta.set(k, d)
			meta.key.changed = true
			return meta
		}
		meta = tx.store.fromStorage(k, meta)
		meta.key.changed = true
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
		meta.set(k, d)
		meta.key.changed = true
		return meta
	}
	return meta
}

func (tx *Tx) readKey(key string) *metadata {
	meta := tx.rLockKey(key)
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
			return tx.store.fromStorage(k, meta)
		}
		meta.set(k, d)
	}
	return meta
}

func (tx *Tx) commit() {
	for _, meta := range tx.lockedMetas {
		meta.commit()
	}
	tx.lockedMetas = tx.lockedMetas[:0]
}
