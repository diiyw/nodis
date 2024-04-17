package nodis

import (
	"sync"

	"github.com/diiyw/nodis/ds"
)

var metaPool = sync.Pool{
	New: func() any {
		return new(metadata)
	},
}

type metadata struct {
	key      *Key
	ds       ds.DataStruct
	locker   *sync.RWMutex
	ok       bool
	writable bool
}

func newEmptyMetadata(locker *sync.RWMutex, writable bool) *metadata {
	m := metaPool.Get().(*metadata)
	m.locker = locker
	m.writable = writable
	return m
}

func newMetadata(key *Key, d ds.DataStruct, writable bool, locker *sync.RWMutex) *metadata {
	m := metaPool.Get().(*metadata)
	m.locker = locker
	m.key = key
	m.ds = d
	m.writable = writable
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
	if m.writable {
		m.locker.Unlock()
	} else {
		m.locker.RUnlock()
	}
	m.ds = nil
	m.locker = nil
	m.key = nil
	m.ok = false
	m.writable = false
	metaPool.Put(m)
}
