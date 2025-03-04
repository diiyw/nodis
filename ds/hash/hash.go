package hash

import (
	"encoding/binary"
	"errors"
	"path/filepath"
	"strconv"
	"unsafe"

	"github.com/diiyw/nodis/ds"
)

type HashMap struct {
	data map[string][]byte
}

type keyValuePair struct {
	key   string
	value []byte
}

func (kvPair *keyValuePair) encode() []byte {
	var kLen = len(kvPair.key)
	var b = make([]byte, 8+kLen+len(kvPair.value))
	n := binary.PutVarint(b, int64(kLen))
	copy(b[n:], kvPair.key)
	n += kLen
	n += copy(b[n:], kvPair.value)
	return b[:n]
}

func decodeKeyValuePair(b []byte) *keyValuePair {
	l, n := binary.Varint(b)
	b = b[n:]
	key := string(b[:l])
	return &keyValuePair{
		key:   key,
		value: b[l:],
	}
}

// NewHashMap creates a new hash
func NewHashMap() *HashMap {
	return &HashMap{
		data: make(map[string][]byte),
	}
}

// Type returns the type of the data structure
func (s *HashMap) Type() ds.ValueType {
	return ds.Hash
}

// HSet sets the value of a hash
func (s *HashMap) HSet(key string, value []byte) int64 {
	_, replaced := s.data[key]
	s.data[key] = value
	if replaced {
		return 0
	}
	return 1
}

// HGet gets the value of a hash
func (s *HashMap) HGet(key string) []byte {
	v, ok := s.data[key]
	if !ok {
		return nil
	}
	return v
}

// HDel deletes the value of a hash
func (s *HashMap) HDel(key ...string) int64 {
	var v int64 = 0
	for _, k := range key {
		_, deleted := s.data[k]
		if deleted {
			delete(s.data, k)
			v++
		}
	}
	return v
}

// HLen gets the length of a hash
func (s *HashMap) HLen() int64 {
	return int64(len(s.data))
}

// HKeys gets the keys of a hash
func (s *HashMap) HKeys() []string {
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys
}

// HExists checks if a key exists in a hash
func (s *HashMap) HExists(key string) bool {
	_, ok := s.data[key]
	return ok
}

// HGetAll gets all the values of a hash
func (s *HashMap) HGetAll() map[string][]byte {
	values := make(map[string][]byte, len(s.data))
	for k, v := range s.data {
		values[k] = v
	}
	return values
}

// HIncrBy increments the value of a hash
func (s *HashMap) HIncrBy(key string, value int64) (int64, error) {
	v, ok := s.data[key]
	if !ok {
		s.data[key] = []byte(strconv.FormatInt(value, 10))
		return value, nil
	}
	vi, err := strconv.ParseInt(*(*string)(unsafe.Pointer(&v)), 10, 64)
	if err != nil {
		return 0, errors.New("ERR hash value is not an integer")
	}
	i := vi + value
	s.data[key] = []byte(strconv.FormatInt(i, 10))
	return i, nil
}

// HIncrByFloat increments the value of a hash
func (s *HashMap) HIncrByFloat(key string, value float64) (float64, error) {
	v, ok := s.data[key]
	if !ok {
		s.data[key] = []byte(strconv.FormatFloat(value, 'f', -1, 64))
		return value, nil
	}
	vf, err := strconv.ParseFloat(*(*string)(unsafe.Pointer(&v)), 64)
	if err != nil {
		return 0, errors.New("ERR hash value is not an integer")
	}
	f := vf + value
	s.data[key] = []byte(strconv.FormatFloat(f, 'f', -1, 64))
	return f, nil
}

// HMSet sets the values of a hash
func (s *HashMap) HMSet(values map[string][]byte) {
	for key, value := range values {
		s.data[key] = value
	}
}

// HMGet gets the values of a hash
func (s *HashMap) HMGet(fields ...string) [][]byte {
	values := make([][]byte, len(fields))
	for i, key := range fields {
		value, ok := s.data[key]
		if ok {
			values[i] = value
		} else {
			values[i] = nil
		}
	}
	return values
}

// HSetNX sets the value of a hash if it does not exist
func (s *HashMap) HSetNX(key string, value []byte) bool {
	_, ok := s.data[key]
	if ok {
		return false
	}
	s.data[key] = value
	return true
}

// HVals gets the values of a hash
func (s *HashMap) HVals() [][]byte {
	values := make([][]byte, 0, len(s.data))
	for _, v := range s.data {
		values = append(values, v)
	}
	return values
}

// HScan scans the values of a hash
func (s *HashMap) HScan(cursor int64, match string, count int64) (int64, map[string][]byte) {
	values := make(map[string][]byte, len(s.data))
	var i int64 = 0
	for k, v := range s.data {
		matched, _ := filepath.Match(match, k)
		if matched && i >= cursor {
			values[k] = v
		}
		i++
		if i >= cursor+count {
			break
		}
	}
	return i, values
}

// HStrLen gets the length of a hash
func (s *HashMap) HStrLen(field string) int64 {
	v, ok := s.data[field]
	if !ok {
		return 0
	}
	return int64(len(*(*string)(unsafe.Pointer(&v))))
}

func (s *HashMap) GetValue() []byte {
	values := make([]byte, 0, 1024*100)
	for key, value := range s.data {
		kvPair := &keyValuePair{
			key:   key,
			value: value,
		}
		data := kvPair.encode()
		dataLen := len(data)
		var b = make([]byte, 8+dataLen)
		n := binary.PutVarint(b, int64(dataLen))
		copy(b[n:], data)
		values = append(values, b[:n+dataLen]...)
	}
	return values
}

// SetValue the set from bytes
func (s *HashMap) SetValue(values []byte) {
	for {
		if len(values) == 0 {
			break
		}
		dataLen, n := binary.Varint(values)
		if n <= 0 {
			break
		}
		end := n + int(dataLen)
		item := decodeKeyValuePair(values[n:end])
		s.HSet(item.key, item.value)
		values = values[end:]
	}
}
