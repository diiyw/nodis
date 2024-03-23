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
