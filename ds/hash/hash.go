package hash

import (
	"path/filepath"

	"github.com/dolthub/swiss"
	"github.com/kelindar/binary"
)

type Hash struct {
	data *swiss.Map[string, any]
}

// NewHash creates a new hash
func NewHash() *Hash {
	return &Hash{
		data: swiss.NewMap[string, any](32),
	}
}

// HSet sets the value of a hash
func (s *Hash) HSet(key string, value any) {
	s.data.Put(key, value)
}

// HGet gets the value of a hash
func (s *Hash) HGet(key string) (any, bool) {
	return s.data.Get(key)
}

// HDel deletes the value of a hash
func (s *Hash) HDel(key string) {
	s.data.Delete(key)
}

// HLen gets the length of a hash
func (s *Hash) HLen() int {
	return s.data.Count()
}

// HKeys gets the keys of a hash
func (s *Hash) HKeys() []string {
	keys := make([]string, 0, s.data.Count())
	s.data.Iter(func(key string, _ any) bool {
		keys = append(keys, key)
		return false
	})
	return keys
}

// HExists checks if a key exists in a hash
func (s *Hash) HExists(key string) bool {
	_, ok := s.data.Get(key)
	return ok
}

// HGetAll gets all the values of a hash
func (s *Hash) HGetAll() map[string]any {
	values := make(map[string]any, s.data.Count())
	s.data.Iter(func(key string, value any) bool {
		values[key] = value
		return false
	})
	return values
}

// HIncrBy increments the value of a hash
func (s *Hash) HIncrBy(key string, value int) {
	v, ok := s.data.Get(key)
	if !ok {
		s.data.Put(key, value)
		return
	}
	s.data.Put(key, v.(int)+value)
}

// HIncByFloat increments the value of a hash
func (s *Hash) HIncrByFloat(key string, value float64) {
	v, ok := s.data.Get(key)
	if !ok {
		s.data.Put(key, value)
		return
	}
	s.data.Put(key, v.(float64)+value)
}

// HMSet sets the values of a hash
func (s *Hash) HMSet(values map[string]any) {
	for key, value := range values {
		s.data.Put(key, value)
	}
}

// HMGet gets the values of a hash
func (s *Hash) HMGet(keys ...string) map[string]any {
	values := make(map[string]any, len(keys))
	for _, key := range keys {
		value, ok := s.data.Get(key)
		if ok {
			values[key] = value
		}
	}
	return values
}

// HSetNX sets the value of a hash if it does not exist
func (s *Hash) HSetNX(key string, value any) bool {
	_, ok := s.data.Get(key)
	if ok {
		return false
	}
	s.data.Put(key, value)
	return true
}

// HVals gets the values of a hash
func (s *Hash) HVals() []any {
	values := make([]any, 0, s.data.Count())
	s.data.Iter(func(_ string, value any) bool {
		values = append(values, value)
		return false
	})
	return values
}

// HScan scans the values of a hash
func (s *Hash) HScan(cursor int, match string, count int) (int, map[string]any) {
	values := make(map[string]any, s.data.Count())
	i := 0
	s.data.Iter(func(key string, value any) bool {
		matched, _ := filepath.Match(match, key)
		if matched && i >= cursor {
			values[key] = value
		}
		i++
		return i >= cursor+count
	})
	return i - 1, values
}

// Marshal the set to bytes
func (s *Hash) Marshal() ([]byte, error) {
	return binary.Marshal(s.data)
}

// Unmarshal the set from bytes
func (s *Hash) Unmarshal(data []byte) error {
	return binary.Unmarshal(data, &s.data)
}
