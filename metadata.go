package nodis

import (
	"sync"

	"github.com/diiyw/nodis/ds"
)

const (
	KeyStateNormal   uint8 = 1
	KeyStateModified uint8 = 2
)

type metadata struct {
	*sync.RWMutex
	key       *ds.Key
	value     ds.Value
	count     int64
	valueType ds.ValueType
	state     uint8
	writeable bool
}

func newMetadata() *metadata {
	return &metadata{
		RWMutex:   new(sync.RWMutex),
		count:     0,
		value:     nil,
		writeable: false,
	}
}

func (m *metadata) expired(now int64) bool {
	if m == nil {
		return true
	}
	return m.key.Expiration != 0 && m.key.Expiration <= now
}

// modified return if the key is modified
func (m *metadata) modified() bool {
	if m.value == nil {
		return false
	}
	return m.state&KeyStateModified == KeyStateModified
}

// reset the key state
func (m *metadata) reset() {
	m.state = KeyStateNormal
	m.count--
}

func (m *metadata) setValue(value ds.Value) {
	m.value = value
	m.state |= KeyStateNormal
	m.valueType = value.Type()
}

// empty copy the metadata to empty
func (m *metadata) empty() *metadata {
	newM := &metadata{}
	if m != nil {
		newM.RWMutex = m.RWMutex
	}
	return newM
}

func (m *metadata) isOk() bool {
	return m.state&KeyStateNormal == KeyStateNormal
}

func (m *metadata) removeFromMemory() {
	m.value = nil
}

func (m *metadata) commit() {
	if m.RWMutex == nil {
		// empty metadata
		return
	}
	if m.writeable {
		m.writeable = false
		m.Unlock()
		return
	}
	m.RUnlock()
}
