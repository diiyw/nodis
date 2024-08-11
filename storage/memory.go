package storage

import (
	"sync"

	"github.com/diiyw/nodis/ds"
	"github.com/tidwall/btree"
)

type Memory struct {
	sync.RWMutex
	index []byte
	data  btree.Map[string, ds.Value]
}

// Open initializes the storage.
func (m *Memory) Open() error {
	return nil
}

// Get returns a value from the storage.
func (m *Memory) Get(key string) (ds.Value, error) {
	m.RLock()
	defer m.RUnlock()
	v, ok := m.data.Get(key)
	if !ok {
		return nil, ErrKeyNotFound
	}
	return v, nil
}

// Put sets a value in the storage.
func (m *Memory) Put(key string, value ds.Value, expiration int64) error {
	m.Lock()
	defer m.Unlock()
	m.data.Set(key, value)
	return nil
}

// Delete removes a value from the storage.
func (m *Memory) Delete(key string) error {
	m.Lock()
	defer m.Unlock()
	m.data.Delete(key)
	return nil
}

// GetIndex returns the index.
func (m *Memory) GetIndex() []byte {
	m.RLock()
	defer m.RUnlock()
	return m.index
}

// PutIndex sets the index.
func (m *Memory) PutIndex(index []byte) error {
	m.Lock()
	defer m.Unlock()
	m.index = index
	return nil
}

// Close closes the storage.
func (m *Memory) Close() error {
	return nil
}

// Reset clears the storage.
func (m *Memory) Reset() error {
	m.Lock()
	defer m.Unlock()
	m.data.Clear()
	m.index = nil
	return nil
}
