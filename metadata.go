package nodis

import (
	"sync"

	"github.com/diiyw/nodis/ds"
)

type metadata struct {
	*sync.Mutex
	key *Key
	ds  ds.DataStruct
	ok  bool
}

func (m *metadata) set(key *Key, d ds.DataStruct) *metadata {
	m.Lock()
	m.key = key
	m.ds = d
	m.ok = true
	return m
}

func (m *metadata) isOk() bool {
	return m.ok
}

func (m *metadata) markChanged() {
	if m.ok {
		m.key.changed = true
	}
}

func (m *metadata) commit() {
	if !m.ok {
		return
	}
	m.ds = nil
	m.key = nil
	m.ok = false
	m.Unlock()
}
