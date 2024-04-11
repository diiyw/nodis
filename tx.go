package nodis

import (
	"sync"

	"github.com/diiyw/nodis/ds"
)

type Tx struct {
	locker   *sync.RWMutex
	key      *Key
	ds       ds.DataStruct
	ok       bool
	writable bool
}

func newEmptyTx(locker *sync.RWMutex, writable bool) *Tx {
	return &Tx{
		locker:   locker,
		ok:       false,
		writable: writable,
	}
}

func newTx(key *Key, d ds.DataStruct, writable bool, locker *sync.RWMutex) *Tx {
	tx := &Tx{
		locker:   locker,
		key:      key,
		ds:       d,
		writable: writable,
		ok:       true,
	}
	return tx
}

func (t *Tx) isOk() bool {
	return t.ok
}

func (t *Tx) markChanged() {
	if t.ok {
		t.key.changed = true
	}
}

func (t *Tx) commit() {
	if t.writable {
		t.locker.Unlock()
	} else {
		t.locker.RUnlock()
	}
}
