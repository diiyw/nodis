package nodis

import (
	"errors"
	"hash/crc32"

	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/ds/hash"
	"github.com/diiyw/nodis/ds/list"
	"github.com/diiyw/nodis/ds/set"
	"github.com/diiyw/nodis/ds/str"
	"github.com/diiyw/nodis/ds/zset"
	"github.com/kelindar/binary"
)

var (
	ErrCorruptedData = errors.New("corrupted data")
)

type Entity struct {
	Key       string
	Value     ds.DataStruct
	ExpiredAt int64
}

type dataEntity struct {
	Crc32 uint32
	Type  ds.DataType
	Body  []byte
}

// newEntry creates a new entry
func newEntry(key string, value ds.DataStruct, expiredAt int64) *Entity {
	return &Entity{
		Key:       key,
		Value:     value,
		ExpiredAt: expiredAt,
	}
}

// Marshal marshals the entry
func (e *Entity) Marshal() ([]byte, error) {
	var err error
	data, err := binary.Marshal(e)
	if err != nil {
		return nil, err
	}
	var buf = make([]byte, len(data)+1)
	buf[0] = byte(e.Value.GetType())
	copy(buf[1:], data)
	var block = dataEntity{
		Crc32: crc32.ChecksumIEEE(buf),
		Type:  e.Value.GetType(),
		Body:  data,
	}
	return binary.Marshal(block)
}

// Unmarshal the entry
func (e *Entity) Unmarshal(data []byte) error {
	var block dataEntity
	if err := binary.Unmarshal(data, &block); err != nil {
		return err
	}
	var buf = make([]byte, len(block.Body)+1)
	buf[0] = byte(block.Type)
	copy(buf[1:], block.Body)
	if block.Crc32 != crc32.ChecksumIEEE(buf) {
		return ErrCorruptedData
	}
	switch block.Type {
	case ds.String:
		e.Value = str.NewString()
	case ds.ZSet:
		e.Value = zset.NewSortedSet()
	case ds.List:
		e.Value = list.NewDoublyLinkedList()
	case ds.Hash:
		e.Value = hash.NewHashMap()
	case ds.Set:
		e.Value = set.NewSet()
	}
	return binary.Unmarshal(block.Body, e)
}
