package nodis

import (
	"sync"

	"github.com/diiyw/nodis/ds"
)

type metadata struct {
	locker   *sync.RWMutex
	key      *Key
	ds       ds.DataStruct
	ok       bool
	writable bool
}

func newEmptyMetadata(locker *sync.RWMutex, writable bool) *metadata {
	return &metadata{
		locker:   locker,
		ok:       false,
		writable: writable,
	}
}

func newMetadata(key *Key, d ds.DataStruct, writable bool, locker *sync.RWMutex) *metadata {
	meta := &metadata{
		locker:   locker,
		key:      key,
		ds:       d,
		writable: writable,
		ok:       true,
	}
	return meta
}

func (t *metadata) isOk() bool {
	return t.ok
}

func (t *metadata) markChanged() {
	if t.ok {
		t.key.changed = true
	}
}

func (t *metadata) commit() {
	if t.writable {
		t.locker.Unlock()
	} else {
		t.locker.RUnlock()
	}
}
