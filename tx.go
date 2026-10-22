package nodis

import (
	"github.com/diiyw/nodis/ds"
	"time"
)

// Tx borrows the store execution lock. Nested operations use the same context.
// It is an isolation boundary, not a rollback transaction.
type Tx struct {
	store          *store
	err            error
	suppressEvents bool
	cleared        bool
}

func (tx *Tx) lockKey(key string) *metadata {
	if m := tx.store.metadata[key]; m != nil {
		m.markAccessed()
		return m
	}
	return newMetadata(ds.NewKey(key, 0), true)
}
func (tx *Tx) storeMeta(m *metadata) {
	tx.store.metadata[m.key.Name] = m
	tx.store.addGCKey(m.key.Name)
	tx.store.dirty[m.key.Name] = struct{}{}
	m.markAccessed()
	delete(tx.store.deleted, m.key.Name)
}
func (tx *Tx) resetMeta(m *metadata, fn func() ds.Value) {
	m.key.Expiration = 0
	m.value = nil
	m.state = 0
	m.accessed.Store(fn != nil)
	if fn != nil {
		m.setValue(fn())
		m.state |= KeyStateModified
	}
}
func (tx *Tx) newStoredMetadata(m *metadata, fn func() ds.Value) *metadata {
	tx.resetMeta(m, fn)
	tx.storeMeta(m)
	return m
}
func (tx *Tx) delKey(key string) {
	delete(tx.store.metadata, key)
	tx.store.removeGCKey(key)
	delete(tx.store.dirty, key)
	tx.store.deleted[key] = struct{}{}
	tx.store.changed[key] = struct{}{}
	tx.store.invalidate(key)
}
func (tx *Tx) writeKey(key string, fn func() ds.Value) *metadata {
	m := tx.lockKey(key)
	if m.isOk() && m.expired(time.Now().UnixMilli()) {
		tx.delKey(key)
		tx.resetMeta(m, nil)
	}
	if !m.isOk() {
		if fn != nil {
			return tx.newStoredMetadata(m, fn)
		}
		return m
	}
	if m.value == nil {
		v, err := tx.store.ss.Get(m.key)
		if err != nil {
			tx.err = err
			// Storage failures must never be interpreted as permission to overwrite data.
			panic(err)
		}
		m.setValue(v)
	}
	return m
}
func (tx *Tx) readKey(key string) *metadata {
	m := tx.store.metadata[key]
	if !m.isOk() || m.expired(time.Now().UnixMilli()) {
		return nil
	}
	m.markAccessed()
	if m.value == nil {
		v, err := tx.store.ss.Get(m.key)
		if err != nil {
			tx.err = err
			panic(err)
		}
		m.setValue(v)
	}
	return m
}
