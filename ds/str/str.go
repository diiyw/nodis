package str

import (
	"sync"

	"github.com/diiyw/nodis/ds"
	"github.com/kelindar/binary"
)

type String struct {
	sync.RWMutex
	V []byte
}

func NewString() *String {
	return &String{}
}

// GetType returns the type of the data structure
func (s *String) GetType() ds.DataType {
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

// Marshal the string to bytes
func (s *String) Marshal() ([]byte, error) {
	return binary.Marshal(s.V)
}

// Unmarshal the bytes to string
func (s *String) Unmarshal(data []byte) error {
	return binary.Unmarshal(data, &s.V)
}
