package storage

import (
	"sync"

	"github.com/diiyw/nodis/ds"
	"github.com/tidwall/btree"
)

type KeyValue struct {
	key   *ds.Key
	value ds.Value
}

type Memory struct {
	sync.RWMutex
	data btree.Map[string, KeyValue]
}

func NewMemory() *Memory {
	return &Memory{}
}

// Init initializes the storage.
func (m *Memory) Init() error {
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
	return v.value, nil
}

// Set sets a value in the storage.
func (m *Memory) Set(key *ds.Key, value ds.Value) error {
	m.Lock()
	defer m.Unlock()
	m.data.Set(key.Name, KeyValue{
		key:   key,
		value: value,
	})
	return nil
}

// Delete removes a value from the storage.
func (m *Memory) Delete(key string) error {
	m.Lock()
	defer m.Unlock()
	m.data.Delete(key)
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
func (m *Memory) ScanKeys(f func(*ds.Key) bool) {
	m.RLock()
	defer m.RUnlock()
	m.data.Scan(func(_ string, kv KeyValue) bool {
		return f(kv.key)
	})
}

// Clear the storage.
func (m *Memory) Clear() error {
	m.Lock()
	defer m.Unlock()
	m.data.Clear()
	return nil
}
