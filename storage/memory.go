package storage

import (
	"sync"

	"github.com/diiyw/nodis/ds"
	"github.com/tidwall/btree"
)

type Memory struct {
	sync.RWMutex
	keys   btree.Map[string, *Key]
	values btree.Map[string, ds.Value]
}

// Open initializes the storage.
func (m *Memory) Open() error {
	return nil
}

// Get returns a value from the storage.
func (m *Memory) Get(key string) (ds.Value, error) {
	m.RLock()
	defer m.RUnlock()
	v, ok := m.values.Get(key)
	if !ok {
		return nil, ErrKeyNotFound
	}
	return v, nil
}

// Put sets a value in the storage.
func (m *Memory) Put(key *Key, value ds.Value) error {
	m.Lock()
	defer m.Unlock()
	m.values.Set(key.Name, value)
	m.keys.Set(key.Name, key)
	return nil
}

// Delete removes a value from the storage.
func (m *Memory) Delete(key string) error {
	m.Lock()
	defer m.Unlock()
	m.values.Delete(key)
	m.keys.Delete(key)
	return nil
}

// Close closes the storage.
func (m *Memory) Close() error {
	return nil
}

// Snapshot creates a snapshot of the storage.
func (m *Memory) Snapshot() error {
	return nil
}

// ScanKeys returns the keys in the storage.
func (m *Memory) ScanKeys(f func(*Key) bool) {
	m.RLock()
	defer m.RUnlock()
	m.keys.Scan(func(key string, value *Key) bool {
		return f(value)
	})
}

// Reset clears the storage.
func (m *Memory) Reset() error {
	m.Lock()
	defer m.Unlock()
	m.values.Clear()
	m.keys.Clear()
	return nil
}
