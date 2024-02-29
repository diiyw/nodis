package set

import "github.com/kelindar/binary"

type Set struct {
	v any
}

func NewSet(v any) *Set {
	return &Set{v: v}
}

// Marshal the string to bytes
func (s *Set) Marshal() ([]byte, error) {
	return binary.Marshal(s)
}

// Unmarshal the bytes to string
func (s *Set) Unmarshal(data []byte) error {
	return binary.Unmarshal(data, s)
}
