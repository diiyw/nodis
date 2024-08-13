package storage

import (
	"encoding/binary"
	"errors"
	"hash/crc32"

	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/ds/hash"
	"github.com/diiyw/nodis/ds/list"
	"github.com/diiyw/nodis/ds/set"
	"github.com/diiyw/nodis/ds/str"
	"github.com/diiyw/nodis/ds/zset"
)

var (
	ErrCorruptedData = errors.New("corrupted values")
)

// ValueEntry is the entry of the value
type ValueEntry struct {
	Type       uint8
	Expiration int64
	Value      []byte
}

func (v *ValueEntry) encode() []byte {
	var b = make([]byte, 1+8+len(v.Value))
	b[0] = v.Type
	n := binary.PutVarint(b[1:], v.Expiration)
	copy(b[n+1:], v.Value)
	b = b[:n+1+len(v.Value)]
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
	v.Value = b[n+1:]
	return nil
}

// NewValueEntry creates a new entity
func NewValueEntry(v ds.Value, expiration int64) *ValueEntry {
	e := &ValueEntry{
		Expiration: expiration,
		Type:       uint8(v.Type()),
	}
	e.Value = v.GetValue()
	return e
}

func parseValueEntry(data []byte) (*ValueEntry, error) {
	var entry = &ValueEntry{}
	if err := entry.decode(data); err != nil {
		return nil, err
	}
	return entry, nil
}

func parseValue(data []byte) (*ValueEntry, ds.Value, error) {
	var entry, err = parseValueEntry(data)
	if err != nil {
		return nil, nil, err
	}
	var value ds.Value
	switch ds.ValueType(entry.Type) {
	case ds.String:
		v := str.NewString()
		v.SetValue(entry.Value)
		value = v
	case ds.ZSet:
		z := zset.NewSortedSet()
		z.SetValue(entry.Value)
		value = z
	case ds.List:
		l := list.NewLinkedList()
		l.SetValue(entry.Value)
		value = l
	case ds.Hash:
		h := hash.NewHashMap()
		h.SetValue(entry.Value)
		value = h
	case ds.Set:
		v := set.NewSet()
		v.SetValue(entry.Value)
		value = v
	default:
		panic("unhandled default case")
	}
	return entry, value, nil
}
