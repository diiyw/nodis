package hash

import (
	"path/filepath"
	"strconv"
	"sync"

	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/utils"
	"github.com/dolthub/swiss"
	"github.com/kelindar/binary"
)

type HashMap struct {
	sync.RWMutex
	data *swiss.Map[string, []byte]
}

// NewHashMap creates a new hash
func NewHashMap() *HashMap {
	return &HashMap{
		data: swiss.NewMap[string, []byte](32),
	}
}

// GetType returns the type of the data structure
func (s *HashMap) GetType() ds.DataType {
	return ds.Hash
}

// HSet sets the value of a hash
func (s *HashMap) HSet(key string, value []byte) {
	s.data.Put(key, value)
}

// HGet gets the value of a hash
func (s *HashMap) HGet(key string) []byte {
	v, ok := s.data.Get(key)
	if !ok {
		return nil
	}
	return v
}

// HDel deletes the value of a hash
func (s *HashMap) HDel(key string) {
	s.data.Delete(key)
}

// HLen gets the length of a hash
func (s *HashMap) HLen() int {
	return s.data.Count()
}

// HKeys gets the keys of a hash
func (s *HashMap) HKeys() []string {
	keys := make([]string, 0, s.data.Count())
	s.data.Iter(func(key string, _ []byte) bool {
		keys = append(keys, key)
		return false
	})
	return keys
}

// HExists checks if a key exists in a hash
func (s *HashMap) HExists(key string) bool {
	_, ok := s.data.Get(key)
	return ok
}

// HGetAll gets all the values of a hash
func (s *HashMap) HGetAll() map[string][]byte {
	values := make(map[string][]byte, s.data.Count())
	s.data.Iter(func(key string, value []byte) bool {
		values[key] = value
		return false
	})
	return values
}

// HIncrBy increments the value of a hash
func (s *HashMap) HIncrBy(key string, value int64) int64 {
	v, ok := s.data.Get(key)
	if !ok {
		s.data.Put(key, []byte(strconv.FormatInt(value, 10)))
		return 0
	}
	vi, _ := strconv.ParseInt(utils.Byte2String(v), 10, 64)
	i := vi + value
	s.data.Put(key, []byte(strconv.FormatInt(i, 10)))
	return i
}

// HIncByFloat increments the value of a hash
func (s *HashMap) HIncrByFloat(key string, value float64) float64 {
	v, ok := s.data.Get(key)
	if !ok {
		s.data.Put(key, []byte(strconv.FormatFloat(value, 'f', -1, 64)))
		return 0
	}
	vf, _ := strconv.ParseFloat(utils.Byte2String(v), 64)
	f := vf + value
	s.data.Put(key, []byte(strconv.FormatFloat(f, 'f', -1, 64)))
	return f
}

// HMSet sets the values of a hash
func (s *HashMap) HMSet(values map[string][]byte) {
	for key, value := range values {
		s.data.Put(key, value)
	}
}

// HMGet gets the values of a hash
func (s *HashMap) HMGet(fields ...string) [][]byte {
	values := make([][]byte, 0, len(fields))
	for _, key := range fields {
		value, ok := s.data.Get(key)
		if ok {
			values = append(values, value)
		} else {
			values = append(values, nil)
		}
	}
	return values
}

// HSetNX sets the value of a hash if it does not exist
func (s *HashMap) HSetNX(key string, value []byte) bool {
	_, ok := s.data.Get(key)
	if ok {
		return false
	}
	s.data.Put(key, value)
	return true
}

// HVals gets the values of a hash
func (s *HashMap) HVals() [][]byte {
	values := make([][]byte, 0, s.data.Count())
	s.data.Iter(func(_ string, value []byte) bool {
		values = append(values, value)
		return false
	})
	return values
}

// HScan scans the values of a hash
func (s *HashMap) HScan(cursor int, match string, count int) (int, map[string][]byte) {
	values := make(map[string][]byte, s.data.Count())
	i := 0
	s.data.Iter(func(key string, value []byte) bool {
		matched, _ := filepath.Match(match, key)
		if matched && i >= cursor {
			values[key] = value
		}
		i++
		return i >= cursor+count
	})
	return i, values
}

// Marshal the set to bytes
func (s *HashMap) MarshalBinary() ([]byte, error) {
	var data = s.HGetAll()
	return binary.Marshal(data)
}

// Unmarshal the set from bytes
func (s *HashMap) UnmarshalBinary(data []byte) error {
	var values map[string][]byte
	if err := binary.Unmarshal(data, &values); err != nil {
		return err
	}

	s.data = swiss.NewMap[string, []byte](32)
	for key, value := range values {
		s.data.Put(key, value)
	}
	return nil
}
