package storage

import (
	"sync"

	"github.com/diiyw/nodis/ds"
)

type KeyValue struct {
	key   *ds.Key
	value ds.Value
}

type Memory struct {
	sync.RWMutex
	data map[string]KeyValue
}

func NewMemory() *Memory {
	return &Memory{
		data: make(map[string]KeyValue),
	}
}

// Init initializes the storage.
func (m *Memory) Init() error {
	return nil
}

// Get returns a value from the storage.
func (m *Memory) Get(key *ds.Key) (ds.Value, error) {
	m.RLock()
	defer m.RUnlock()
	v, ok := m.data[string(key.Encode())]
	if !ok {
		return nil, ErrKeyNotFound
	}
	return v.value, nil
}

// Set sets a value in the storage.
func (m *Memory) Set(key *ds.Key, value ds.Value) error {
	m.Lock()
	defer m.Unlock()
	m.data[string(key.Encode())] = KeyValue{
		key:   key,
		value: value,
	}
	return nil
}

// Delete removes a value from the storage.
func (m *Memory) Delete(key *ds.Key) error {
	m.Lock()
	defer m.Unlock()
	delete(m.data, string(key.Encode()))
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
	for _, kv := range m.data {
		if !f(kv.key) {
			break
		}
	}
}

// Clear the storage.
func (m *Memory) Clear() error {
	m.Lock()
	defer m.Unlock()
	m.data = make(map[string]KeyValue)
	return nil
}
