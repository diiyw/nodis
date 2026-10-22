package nodis

import (
	"sync/atomic"

	"github.com/diiyw/nodis/ds"
)

const (
	KeyStateNormal   uint8 = 1
	KeyStateModified uint8 = 2
)

// metadata is owned by store.mu. There are no independent key locks.
type metadata struct {
	key       *ds.Key
	value     ds.Value
	accessed  atomic.Bool
	valueType ds.ValueType
	state     uint8
}

func newMetadata(key *ds.Key, _ bool) *metadata { return &metadata{key: key} }
func (m *metadata) expired(now int64) bool {
	return m == nil || m.key.Expiration != 0 && m.key.Expiration <= now
}
func (m *metadata) modified() bool { return m.value != nil && m.state&KeyStateModified != 0 }

// GC clears accessed under the exclusive store lock. Readers only set it, so a
// load before storing avoids repeated writes to a shared cache line.
func (m *metadata) markAccessed() {
	if !m.accessed.Load() {
		m.accessed.Store(true)
	}
}
func (m *metadata) setValue(v ds.Value) {
	m.value = v
	m.state |= KeyStateNormal
	m.valueType = v.Type()
}
func (m *metadata) isOk() bool        { return m != nil && m.state&KeyStateNormal != 0 }
func (m *metadata) removeFromMemory() { m.value = nil }
