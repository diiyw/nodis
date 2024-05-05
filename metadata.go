package nodis

import (
	"sync"

	"github.com/diiyw/nodis/ds"
)

type metadata struct {
	*sync.RWMutex
	key       *Key
	ds        ds.DataStruct
	ok        bool
	writeable bool
}

func (m *metadata) set(key *Key, d ds.DataStruct) *metadata {
	m.key = key
	m.ds = d
	m.ok = true
	return m
}

func (m *metadata) isOk() bool {
	return m.ok
}

func (m *metadata) commit() {
	if m.writeable {
		m.ds = nil
		m.key = nil
		m.ok = false
		m.writeable = false
		m.Unlock()
		return
	}
	m.RUnlock()
}
