package storage

import (
	"github.com/diiyw/nodis/ds"
	"sync"
)

type memoryRecord struct {
	key  ds.Key
	data []byte
}

// Memory stores serialized, owned records: flushed state never aliases hot values.
type Memory struct {
	sync.RWMutex
	data map[string]memoryRecord
}

func NewMemory() *Memory      { return &Memory{data: make(map[string]memoryRecord)} }
func (m *Memory) Init() error { return nil }
func (m *Memory) Get(key *ds.Key) (ds.Value, error) {
	m.RLock()
	defer m.RUnlock()
	r, ok := m.data[key.Name]
	if !ok {
		return nil, ErrKeyNotFound
	}
	return DecodeValue(append([]byte(nil), r.data...))
}
func (m *Memory) Set(k *ds.Key, v ds.Value) error {
	return m.Apply([]Mutation{{Key: *k, Value: v}}, false)
}
func (m *Memory) Delete(k *ds.Key) error { return m.Apply([]Mutation{{Key: *k, Delete: true}}, false) }
func (m *Memory) Clear() error           { return m.Apply(nil, true) }
func (m *Memory) Close() error           { return nil }
func (m *Memory) Snapshot() error        { return nil }
func (m *Memory) ScanKeys(fn func(*ds.Key) bool) error {
	m.RLock()
	defer m.RUnlock()
	for _, r := range m.data {
		k := r.key
		if !fn(&k) {
			break
		}
	}
	return nil
}
func (m *Memory) Apply(mutations []Mutation, reset bool) error {
	records := make([]memoryRecord, len(mutations))
	for i, change := range mutations {
		if !change.Delete {
			records[i] = memoryRecord{key: change.Key, data: EncodeValue(change.Value)}
		}
	}
	m.Lock()
	defer m.Unlock()
	if reset {
		clear(m.data)
	}
	for i, change := range mutations {
		if change.Delete {
			delete(m.data, change.Key.Name)
		} else {
			m.data[change.Key.Name] = records[i]
		}
	}
	return nil
}
