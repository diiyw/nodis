package nodis

import (
	"github.com/diiyw/nodis/storage"
	"sync"
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
		m.Lock()
		m.writeable = true
		tx.lockedMetas = append(tx.lockedMetas, m)
		m.count++
		return m
	}
	return m.empty()
}

func (tx *Tx) rLockKey(key string) *metadata {
	tx.store.mu.RLock()
	m, ok := tx.store.metadata.Get(key)
	tx.store.mu.RUnlock()
	if ok {
		m.RLock()
		tx.lockedMetas = append(tx.lockedMetas, m)
		m.count++
		return m
	}
	return m.empty()
}

func (tx *Tx) newKey(m *metadata, key string, newFn func() ds.Value) *metadata {
	if newFn != nil {
		tx.store.mu.Lock()
		if m.RWMutex == nil {
			m.RWMutex = new(sync.RWMutex)
		}
		value := newFn()
		m.key = storage.NewKey(key, 0)
		m.setValue(value)
		m.state |= KeyStateModified
		tx.store.metadata.Set(key, m)
		tx.store.mu.Unlock()
		return m
	}
	return m.empty()
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
			return tx.newKey(m, key, newFn)
		}
		// not expired
		if m.value != nil {
			return m
		}
		// if not found in memory, read from storage
		v, err := tx.store.sg.Get(key)
		if err != nil {
			return tx.newKey(m, key, newFn)
		}
		m.setValue(v)
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
		// not expired
		if m.value != nil {
			return m
		}
		// if not found in memory, read from storage
		v, err := tx.store.sg.Get(key)
		if err != nil {
			return m.empty()
		}
		m.setValue(v)
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
