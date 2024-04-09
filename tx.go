package nodis

import (
	"github.com/diiyw/nodis/ds"
)

type Tx struct {
	key      *Key
	ds       ds.DataStruct
	ok       bool
	writable bool
}

var emptyTx = &Tx{ok: false}

func newTx(key *Key, d ds.DataStruct, writable bool) *Tx {
	return &Tx{
		key:      key,
		ds:       d,
		writable: writable,
		ok:       true,
	}
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
	if t.ok {
		if t.writable {
			t.key.Unlock()
		} else {
			t.key.RUnlock()
		}
	}
}
