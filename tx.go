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
	m, ok := tx.store.metadata.Get(key)
	if ok {
		m.Lock()
		m.writeable = true
		tx.lockedMetas = append(tx.lockedMetas, m)
		return m
	}
	tx.store.mu.Lock()
	m = newMetadata(newKey(), nil, true)
	m.Lock()
	tx.store.metadata.Set(key, m)
	tx.lockedMetas = append(tx.lockedMetas, m)
	tx.store.mu.Unlock()
	return m
}

func (tx *Tx) rLockKey(key string) *metadata {
	m, ok := tx.store.metadata.Get(key)
	if ok {
		m.RLock()
		tx.lockedMetas = append(tx.lockedMetas, m)
		return m
	}
	tx.store.mu.Lock()
	m = newMetadata(newKey(), nil, false)
	m.RLock()
	tx.store.metadata.Set(key, m)
	tx.lockedMetas = append(tx.lockedMetas, m)
	tx.store.mu.Unlock()
	return m
}

func (tx *Tx) newKey(m *metadata, key string, newFn func() ds.Value) *metadata {
	if newFn != nil {
		tx.store.mu.Lock()
		defer tx.store.mu.Unlock()
		value := newFn()
		if value == nil {
			return m
		}
		m.key = newKey()
		m.setValue(value)
		m.state |= KeyStateModified
		tx.store.metadata.Set(key, m)
		return m
	}
	return m
}

func (tx *Tx) delKey(key string) {
	tx.store.mu.Lock()
	tx.store.metadata.Delete(key)
	tx.store.mu.Unlock()
}

func (tx *Tx) writeKey(key string, newFn func() ds.Value) *metadata {
	m := tx.lockKey(key)
	if m.isOk() && !m.expired(time.Now().UnixMilli()) {
		// not expired
		if m.value != nil {
			return m
		}
		// if not found in memory, read from storage
		m = tx.store.fromStorage(m)
		return m
	}
	return tx.newKey(m, key, newFn)
}

func (tx *Tx) readKey(key string) *metadata {
	m := tx.rLockKey(key)
	if m.isOk() {
		if m.expired(time.Now().UnixMilli()) {
			return m.empty()
		}
		if m.value == nil {
			// read from storage
			return tx.store.fromStorage(m)
		}
		return m
	}
	return m.empty()
}

func (tx *Tx) commit() {
	for _, meta := range tx.lockedMetas {
		meta.commit()
	}
	tx.lockedMetas = tx.lockedMetas[:0]
}
