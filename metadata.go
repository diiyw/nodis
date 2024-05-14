package nodis

import (
	"sync"

	"github.com/diiyw/nodis/ds"
)

type metadata struct {
	*sync.RWMutex
	key       *Key
	value     ds.Value
	writeable bool
}

func (m *metadata) set(key *Key, d ds.Value) *metadata {
	m.key = key
	m.value = d
	if m.key.valueType == 0 {
		m.key.valueType = m.value.Type()
	}
	m.key.state |= KeyStateNormal
	return m
}

func (m *metadata) isOk() bool {
	return m.key.state&KeyStateNormal == KeyStateNormal
}

func (m *metadata) empty() *metadata {
	m.key = &Key{
		state: KeyStateEmpty,
	}
	m.value = nil
	return m
}

func (m *metadata) commit() {
	if m.RWMutex == nil {
		// emptyMetadata
		return
	}
	if m.writeable {
		m.empty()
		m.writeable = false
		m.Unlock()
		return
	}
	m.RUnlock()
}
