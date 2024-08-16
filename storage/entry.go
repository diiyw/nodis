package storage

import (
	"errors"

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
	Type  uint8
	Value []byte
}

func (e *Entry) encode() []byte {
	var b = make([]byte, 1+len(e.Value))
	b[0] = e.Type
	copy(b[1:], e.Value)
	return b
}

func (e *Entry) decode(b []byte) error {
	if len(b) < 1 {
		return ErrCorruptedData
	}
	e.Type = b[0]
	e.Value = b[1:]
	return nil
}

// NewEntry creates a new entity
func NewEntry(v ds.Value) *Entry {
	e := &Entry{
		Type: uint8(v.Type()),
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
