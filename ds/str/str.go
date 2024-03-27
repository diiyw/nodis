package str

import (
	"sync"

	"github.com/diiyw/nodis/ds"
)

type String struct {
	sync.RWMutex
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
	s.Lock()
	s.V = v
	s.Unlock()
}

// Get the value
func (s *String) Get() []byte {
	s.RLock()
	defer s.RUnlock()
	return s.V
}

// SetBit set a bit in a key
func (s *String) SetBit(offset int64, value int) int {
	s.Lock()
	defer s.Unlock()
	if offset < 0 {
		return 0
	}
	i := offset / 8
	if i > int64(len(s.V))-1 {
		s.V = append(s.V, make([]byte, i-int64(len(s.V))+1)...)
	}
	by := s.V[i]
	bit := byte(1 << uint(offset%8))
	if value == 1 {
		s.V[i] = by | bit
	} else {
		s.V[i] = by &^ bit
	}
	return 1
}

// GetBit get a bit in a key
func (s *String) GetBit(offset int64) int {
	s.RLock()
	defer s.RUnlock()
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
func (s *String) BitCount() int {
	s.RLock()
	defer s.RUnlock()
	count := 0
	for _, b := range s.V {
		for i := 0; i < 8; i++ {
			if b&(1<<uint(i)) != 0 {
				count++
			}
		}
	}
	return count
}

// GetValue the string to bytes
func (s *String) GetValue() []byte {
	s.RLock()
	defer s.RUnlock()
	return s.V
}

// SetValue the bytes to string
func (s *String) SetValue(data []byte) {
	s.V = data
}
