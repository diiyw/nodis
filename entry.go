package nodis

import (
	"encoding/binary"
	"errors"
	"hash/crc32"

	"github.com/diiyw/nodis/ds"
)

var (
	ErrCorruptedData = errors.New("corrupted data")
)

// ValueEntry is the entry of the value
type ValueEntry struct {
	Type       uint8
	Expiration int64
	Key        string
	Value      []byte
}

func (v *ValueEntry) encode() []byte {
	var keyLen = len(v.Key)
	var b = make([]byte, 1+8+1+keyLen+len(v.Value))
	b[0] = v.Type
	n := binary.PutVarint(b[1:], v.Expiration)
	b[n+1+1] = byte(keyLen)
	copy(b[n+1+1+1:], v.Key)
	copy(b[n+1+1+1+keyLen:], v.Value)
	b = b[:n+1+1+1+keyLen+len(v.Value)]
	c32 := crc32.ChecksumIEEE(b)
	var buf = make([]byte, len(b)+4)
	binary.LittleEndian.PutUint32(buf, c32)
	copy(buf[4:], b)
	return buf
}

func (v *ValueEntry) decode(b []byte) error {
	if len(b) < 4 {
		return ErrCorruptedData
	}
	c32 := binary.LittleEndian.Uint32(b)
	b = b[4:]
	if c32 != crc32.ChecksumIEEE(b) {
		return ErrCorruptedData
	}
	v.Type = b[0]
	i, n := binary.Varint(b[1:])
	v.Expiration = i
	// type+expiration+keyIndex
	keyLen := int(b[1+n+1])
	if len(b) < keyLen {
		return ErrCorruptedData
	}
	v.Key = string(b[1+n+1+1 : 1+n+1+1+keyLen])
	v.Value = b[1+n+1+1+keyLen:]
	return nil
}

// newValueEntry creates a new entity
func newValueEntry(key string, v ds.Value, expiration int64) *ValueEntry {
	e := &ValueEntry{
		Key:        key,
		Expiration: expiration,
		Type:       uint8(v.Type()),
	}
	e.Value = v.GetValue()
	return e
}
