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

// Entry is the entry of the value
type Entry struct {
	Type       uint8
	Expiration int64
	Value      []byte
}

func (e *Entry) encode() []byte {
	var b = make([]byte, 1+8+len(e.Value))
	b[0] = e.Type
	n := binary.PutVarint(b[1:], e.Expiration)
	copy(b[n+1:], e.Value)
	b = b[:n+1+len(e.Value)]
	c32 := crc32.ChecksumIEEE(b)
	var buf = make([]byte, len(b)+4)
	binary.LittleEndian.PutUint32(buf, c32)
	copy(buf[4:], b)
	return buf
}

func (e *Entry) decode(b []byte) error {
	if len(b) < 4 {
		return ErrCorruptedData
	}
	c32 := binary.LittleEndian.Uint32(b)
	b = b[4:]
	if c32 != crc32.ChecksumIEEE(b) {
		return ErrCorruptedData
	}
	e.Type = b[0]
	i, n := binary.Varint(b[1:])
	e.Expiration = i
	e.Value = b[n+1:]
	return nil
}

// NewValueEntry creates a new entity
func NewValueEntry(v ds.Value, expiration int64) *Entry {
	e := &Entry{
		Expiration: expiration,
		Type:       uint8(v.Type()),
	}
	e.Value = v.GetValue()
	return e
}

func parseEntry(data []byte) (*Entry, error) {
	var entry = &Entry{}
	if err := entry.decode(data); err != nil {
		return nil, err
	}
	return entry, nil
}

func parseValue(typ ds.ValueType, data []byte) (ds.Value, error) {
	var value ds.Value
	switch typ {
	case ds.String:
		v := str.NewString()
		v.SetValue(data)
		value = v
	case ds.ZSet:
		z := zset.NewSortedSet()
		z.SetValue(data)
		value = z
	case ds.List:
		l := list.NewLinkedList()
		l.SetValue(data)
		value = l
	case ds.Hash:
		h := hash.NewHashMap()
		h.SetValue(data)
		value = h
	case ds.Set:
		v := set.NewSet()
		v.SetValue(data)
		value = v
	default:
		panic("unhandled default case")
	}
	return value, nil
}
