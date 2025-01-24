package ds

import (
	"encoding/binary"
	"errors"
)

var (
	ErrCorruptedData = errors.New("corrupted data")
)

type Key struct {
	Name       string
	Expiration int64
}

// NewKey returns a new key.
func NewKey(name string, expiration int64) *Key {
	return &Key{Name: name, Expiration: expiration}
}

// Encode encodes the key.
func (k *Key) Encode() []byte {
	var b = make([]byte, 8+len(k.Name))
	binary.LittleEndian.PutUint64(b, uint64(k.Expiration))
	copy(b[8:], k.Name)
	return b
}

// DecodeKey decodes the key.
func DecodeKey(b []byte) (*Key, error) {
	if len(b) < 8 {
		return nil, ErrCorruptedData
	}
	i := int64(binary.LittleEndian.Uint64(b))
	return &Key{Name: string(b[8:]), Expiration: i}, nil
}

type Value interface {
	Type() ValueType
	GetValue() []byte
}

type ValueType uint8

const (
	// 0 => none, (key didn't exist)
	// 1 => string,
	// 2 => set,
	// 3 => list,
	// 4 => zset,
	// 5 => hash
	// 6 => stream
	None ValueType = iota
	String
	Set
	List
	ZSet
	Hash
)

func (d ValueType) String() string {
	switch d {
	case None:
		return "none"
	case String:
		return "string"
	case List:
		return "list"
	case Hash:
		return "hash"
	case Set:
		return "set"
	case ZSet:
		return "zset"
	default:
		return "none"
	}
}

func StringToDataType(s string) ValueType {
	switch s {
	case "STRING":
		return String
	case "LIST":
		return List
	case "HASH":
		return Hash
	case "SET":
		return Set
	case "ZSET":
		return ZSet
	default:
		return None
	}
}
