package set

import (
	"sync"

	"github.com/kelindar/binary"
)

type Set struct {
	sync.RWMutex
	V []byte
}

func NewSet() *Set {
	return &Set{}
}

// GetType returns the type of the data structure
func (s *Set) GetType() string {
	return "set"
}

// Set the value
func (s *Set) Set(v []byte) {
	s.Lock()
	s.V = v
	s.Unlock()
}

// Get the value
func (s *Set) Get() []byte {
	s.RLock()
	defer s.RUnlock()
	return s.V
}

// Marshal the string to bytes
func (s *Set) Marshal() ([]byte, error) {
	return binary.Marshal(s.V)
}

// Unmarshal the bytes to string
func (s *Set) Unmarshal(data []byte) error {
	return binary.Unmarshal(data, &s.V)
}
