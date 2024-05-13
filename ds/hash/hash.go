package hash

import (
	"errors"
	"path/filepath"
	"strconv"
	"unsafe"

	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/pb"
	"github.com/tidwall/btree"
)

type HashMap struct {
	data btree.Map[string, []byte]
}

// NewHashMap creates a new hash
func NewHashMap() *HashMap {
	return &HashMap{}
}

// Type returns the type of the data structure
func (s *HashMap) Type() ds.ValueType {
	return ds.Hash
}

// HSet sets the value of a hash
func (s *HashMap) HSet(key string, value []byte) int64 {
	_, replaced := s.data.Set(key, value)
	if replaced {
		return 0
	}
	return 1
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
func (s *HashMap) HDel(key ...string) int64 {
	var v int64 = 0
	for _, k := range key {
		_, deleted := s.data.Delete(k)
		if deleted {
			v++
		}
	}
	return v
}

// HLen gets the length of a hash
func (s *HashMap) HLen() int64 {
	return int64(s.data.Len())
}

// HKeys gets the keys of a hash
func (s *HashMap) HKeys() []string {
	return s.data.Keys()
}

// HExists checks if a key exists in a hash
func (s *HashMap) HExists(key string) bool {
	_, ok := s.data.Get(key)
	return ok
}

// HGetAll gets all the values of a hash
func (s *HashMap) HGetAll() map[string][]byte {
	values := make(map[string][]byte, s.data.Len())
	s.data.Scan(func(key string, value []byte) bool {
		values[key] = value
		return true
	})
	return values
}

// HIncrBy increments the value of a hash
func (s *HashMap) HIncrBy(key string, value int64) (int64, error) {
	v, ok := s.data.Get(key)
	if !ok {
		s.data.Set(key, []byte(strconv.FormatInt(value, 10)))
		return value, nil
	}
	vi, err := strconv.ParseInt(*(*string)(unsafe.Pointer(&v)), 10, 64)
	if err != nil {
		return 0, errors.New("ERR hash value is not an integer")
	}
	i := vi + value
	s.data.Set(key, []byte(strconv.FormatInt(i, 10)))
	return i, nil
}

// HIncByFloat increments the value of a hash
func (s *HashMap) HIncrByFloat(key string, value float64) (float64, error) {
	v, ok := s.data.Get(key)
	if !ok {
		s.data.Set(key, []byte(strconv.FormatFloat(value, 'f', -1, 64)))
		return value, nil
	}
	vf, err := strconv.ParseFloat(*(*string)(unsafe.Pointer(&v)), 64)
	if err != nil {
		return 0, errors.New("ERR hash value is not an integer")
	}
	f := vf + value
	s.data.Set(key, []byte(strconv.FormatFloat(f, 'f', -1, 64)))
	return f, nil
}

// HMSet sets the values of a hash
func (s *HashMap) HMSet(values map[string][]byte) {
	for key, value := range values {
		s.data.Set(key, value)
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
	s.data.Set(key, value)
	return true
}

// HVals gets the values of a hash
func (s *HashMap) HVals() [][]byte {
	return s.data.Values()
}

// HScan scans the values of a hash
func (s *HashMap) HScan(cursor int64, match string, count int64) (int64, map[string][]byte) {
	values := make(map[string][]byte, s.data.Len())
	var i int64 = 0
	s.data.Scan(func(key string, value []byte) bool {
		matched, _ := filepath.Match(match, key)
		if matched && i >= cursor {
			values[key] = value
		}
		i++
		return i < cursor+count
	})
	return i, values
}

// HStrLen gets the length of a hash
func (s *HashMap) HStrLen(field string) int64 {
	v, ok := s.data.Get(field)
	if !ok {
		return 0
	}
	return int64(len(*(*string)(unsafe.Pointer(&v))))
}

func (s *HashMap) GetValue() []*pb.MemberBytes {
	values := make([]*pb.MemberBytes, 0, s.data.Len())
	s.data.Scan(func(key string, value []byte) bool {
		values = append(values, &pb.MemberBytes{Member: key, Value: value})
		return true
	})
	return values
}

// SetValue the set from bytes
func (s *HashMap) SetValue(values []*pb.MemberBytes) {
	for _, v := range values {
		s.data.Set(v.Member, v.Value)
	}
}
