package nodis

import (
	"encoding/binary"
	"sync"

	"github.com/diiyw/nodis/ds"
)

const (
	KeyStateNormal   uint8 = 1
	KeyStateModified uint8 = 2

	metadataSize = 23
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

func (m *metadata) marshal() []byte {
	var b [metadataSize]byte
	binary.LittleEndian.PutUint64(b[14:22], uint64(m.key.Expiration))
	b[22] = uint8(m.valueType)
	return b[:]
}

func (m *metadata) unmarshal(b []byte) *metadata {
	m.key.Expiration = int64(binary.LittleEndian.Uint64(b[14:22]))
	m.valueType = ds.ValueType(b[22])
	return m
}

func (m *metadata) isOk() bool {
	return m.state&KeyStateNormal == KeyStateNormal
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
