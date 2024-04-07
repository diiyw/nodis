package str

import (
	"encoding/binary"

	"github.com/diiyw/nodis/ds"
)

type String struct {
	V []byte
}

func NewString() *String {
	return &String{}
}

// Type returns the type of the data structure
func (s *String) Type() ds.DataType {
	return ds.String
}

// Set the value
func (s *String) Set(v []byte) {
	s.V = v
}

// Get the value
func (s *String) Get() []byte {
	return s.V
}

// Incr increments the value by 1
func (s *String) Incr() int64 {
	if len(s.V) != 8 {
		s.V = make([]byte, 8)
	}
	v := binary.LittleEndian.Uint64(s.V)
	v++
	binary.LittleEndian.PutUint64(s.V, v)
	return int64(v)
}

// Decr decrements the value by 1
func (s *String) Decr() int64 {
	if len(s.V) != 8 {
		s.V = make([]byte, 8)
	}
	v := binary.LittleEndian.Uint64(s.V)
	v--
	binary.LittleEndian.PutUint64(s.V, v)
	return int64(v)
}

// SetBit set a bit in a key
func (s *String) SetBit(offset int64, value bool) int {
	if offset < 0 {
		return 0
	}
	i := offset / 8
	if i > int64(len(s.V))-1 {
		s.V = append(s.V, make([]byte, i-int64(len(s.V))+1)...)
	}
	by := s.V[i]
	bit := byte(1 << uint(offset%8))
	if value {
		s.V[i] = by | bit
	} else {
		s.V[i] = by &^ bit
	}
	return 1
}

// GetBit get a bit in a key
func (s *String) GetBit(offset int64) int64 {
	return s.getBit(offset)
}

func (s *String) getBit(offset int64) int64 {
	i := offset / 8
	if offset < 0 || i > int64(len(s.V)) {
		return 0
	}
	by := s.V[i]
	bit := byte(1 << uint(offset%8))
	if by&bit != 0 {
		return 1
	}
	return 0
}

// BitCount counts the number of bits set to 1
func (s *String) BitCount(start, end int64) int64 {
	var count int64 = 0
	if start < 0 {
		start = 0
	}
	bl := int64(len(s.V)) * 8
	if end <= 0 {
		end = bl
	}
	if end > int64(len(s.V))*8 {
		end = bl
	}
	for i := start; i < end; i++ {
		if s.getBit(i) == 1 {
			count++
		}
	}
	return count
}

// GetValue the string to bytes
func (s *String) GetValue() []byte {
	return s.V
}

// SetValue the bytes to string
func (s *String) SetValue(data []byte) {
	s.V = data
}
