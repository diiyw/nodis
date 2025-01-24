package nodis

import (
	"time"

	"github.com/diiyw/nodis/ds"
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

func (tx *Tx) lockKey(key string) *metadata {
	tx.store.mu.RLock()
	m, ok := tx.store.metadata.Get(key)
	tx.store.mu.RUnlock()
	if ok {
		tx.lockMeta(m)
		return m
	}
	return newMetadata(ds.NewKey(key, 0), true)
}

func (tx *Tx) rLockKey(key string) *metadata {
	tx.store.mu.RLock()
	m, ok := tx.store.metadata.Get(key)
	tx.store.mu.RUnlock()
	if ok {
		tx.rLockMeta(m)
		return m
	}
	return newMetadata(ds.NewKey(key, 0), false)
}

func (tx *Tx) lockMeta(m *metadata) {
	m.Lock()
	m.writeable = true
	tx.lockedMetas = append(tx.lockedMetas, m)
	m.count.Add(1)
}

func (tx *Tx) rLockMeta(m *metadata) {
	m.RLock()
	tx.lockedMetas = append(tx.lockedMetas, m)
	m.count.Add(1)
}

func (tx *Tx) storeMeta(m *metadata) {
	tx.store.mu.Lock()
	tx.store.metadata.Set(m.key.Name, m)
	tx.store.mu.Unlock()
}

func (tx *Tx) resetMeta(m *metadata, newFn func() ds.Value) {
	m.count.Store(0)
	m.value = nil
	if newFn != nil {
		m.setValue(newFn())
	}
	m.state = KeyStateNormal
}

func (tx *Tx) newStoredMetadata(m *metadata, newFn func() ds.Value) *metadata {
	tx.store.mu.Lock()
	value := newFn()
	m.setValue(value)
	m.state |= KeyStateModified
	tx.lockMeta(m)
	tx.store.metadata.Set(m.key.Name, m)
	tx.store.mu.Unlock()
	return m
}

func (tx *Tx) delKey(key string) {
	tx.store.mu.Lock()
	tx.store.metadata.Delete(key)
	tx.store.mu.Unlock()
}

func (tx *Tx) writeKey(key string, newFn func() ds.Value) *metadata {
	m := tx.lockKey(key)
	if m.isOk() {
		if m.expired(time.Now().UnixMilli()) {
			tx.resetMeta(m, newFn)
			return m
		}
		// not expired
		if m.value != nil {
			return m
		}
		// if not found in memory, read from storage
		v, err := tx.store.ss.Get(m.key)
		if err != nil {
			tx.resetMeta(m, newFn)
			return m
		}
		m.setValue(v)
		return m
	}
	if newFn == nil {
		return m
	}
	return tx.newStoredMetadata(m, newFn)
}

func (tx *Tx) readKey(key string) *metadata {
	m := tx.rLockKey(key)
	if m.isOk() {
		if m.expired(time.Now().UnixMilli()) {
			return newMetadata(m.key, false)
		}
		// not expired
		if m.value != nil {
			return m
		}
		// if not found in memory, read from storage
		v, err := tx.store.ss.Get(m.key)
		if err != nil {
			return newMetadata(m.key, false)
		}
		m.setValue(v)
		return m
	}
	return newMetadata(m.key, false)
}

func (tx *Tx) commit() {
	for _, meta := range tx.lockedMetas {
		meta.commit()
	}
	tx.lockedMetas = tx.lockedMetas[:0]
}
