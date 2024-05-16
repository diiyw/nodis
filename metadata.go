package nodis

import (
	"encoding/binary"
	"sync"

	"github.com/diiyw/nodis/ds"
)

const (
	KeyStateNormal   uint8 = 1
	KeyStateModified uint8 = 2
)

type metadata struct {
	*sync.RWMutex
	key        *Key
	value      ds.Value
	useTimes   uint64
	expiration int64
	valueType  ds.ValueType
	state      uint8
	writeable  bool
}

func newMetadata(key *Key, value ds.Value, writeable bool) *metadata {
	return &metadata{
		RWMutex:   new(sync.RWMutex),
		useTimes:  1,
		key:       key,
		value:     value,
		writeable: writeable,
	}
}

func (m *metadata) expired(now int64) bool {
	if m == nil {
		return true
	}
	return m.expiration != 0 && m.expiration <= now
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
	m.useTimes = 0
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
	var b [23]byte
	binary.LittleEndian.PutUint16(b[0:2], m.key.fileId)
	binary.LittleEndian.PutUint64(b[2:10], uint64(m.key.offset))
	binary.LittleEndian.PutUint32(b[10:14], m.key.size)
	binary.LittleEndian.PutUint64(b[14:22], uint64(m.expiration))
	b[22] = uint8(m.valueType)
	return b[:]
}

func (m *metadata) unmarshal(b []byte) *metadata {
	m.key.fileId = binary.LittleEndian.Uint16(b[0:2])
	m.key.offset = int64(binary.LittleEndian.Uint64(b[2:10]))
	m.key.size = binary.LittleEndian.Uint32(b[10:14])
	m.expiration = int64(binary.LittleEndian.Uint64(b[14:22]))
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
