package str

import (
	"strconv"
	"unsafe"

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

// GetSet set the value and return the old value
func (s *String) GetSet(v []byte) []byte {
	old := s.V
	s.V = v
	return old
}

// Get the value
func (s *String) Get() []byte {
	return s.V
}

// Incr increments the value by 1
func (s *String) Incr(step int64) (int64, error) {
	var v string
	if len(s.V) == 0 {
		v = "0"
	} else {
		v = unsafe.String(unsafe.SliceData(s.V), len(s.V))
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, err
	}
	n += step
	nn := strconv.FormatInt(n, 10)
	s.V = unsafe.Slice(unsafe.StringData(nn), len(nn))
	return n, nil
}

// Decr decrements the value by 1
func (s *String) Decr(step int64) (int64, error) {
	var v string
	if len(s.V) == 0 {
		v = "0"
	} else {
		v = unsafe.String(unsafe.SliceData(s.V), len(s.V))
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, err
	}
	n -= step
	nn := strconv.FormatInt(n, 10)
	s.V = unsafe.Slice(unsafe.StringData(nn), len(nn))
	return n, nil
}

// IncrByFloat increments the value by a float
func (s *String) IncrByFloat(step float64) (float64, error) {
	var v string
	if len(s.V) == 0 {
		v = "0"
	} else {
		v = unsafe.String(unsafe.SliceData(s.V), len(s.V))
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, err
	}
	n += step
	nn := strconv.FormatFloat(n, 'f', -1, 64)
	s.V = unsafe.Slice(unsafe.StringData(nn), len(nn))
	return n, nil
}

// SetBit set a bit in a key
func (s *String) SetBit(offset int64, value bool) int64 {
	if offset < 0 {
		return 0
	}
	i := offset / 8
	if i > int64(len(s.V))-1 {
		s.V = append(s.V, make([]byte, i-int64(len(s.V))+1)...)
	}
	by := s.V[i]
	bit := byte(1 << (7 - uint(offset%8)))
	old := by & bit
	if value {
		s.V[i] = by | bit
	} else {
		s.V[i] = by &^ bit
	}
	if old != 0 {
		return 1
	}
	return 0
}

// GetBit get a bit in a key
func (s *String) GetBit(offset int64) int64 {
	return s.getBit(offset)
}

func (s *String) getBit(offset int64) int64 {
	i := offset / 8
	if offset < 0 || len(s.V) == 0 || i > int64(len(s.V))-1 {
		return 0
	}
	by := s.V[i]
	bit := byte(1 << (7 - uint(offset%8)))
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
	if start >= int64(len(s.V)) {
		return 0
	}
	bl := int64(len(s.V))
	if end <= 0 {
		end += bl + 1
	}
	if end > bl {
		end = bl
	}
	if start > end {
		return 0
	}
	if start == end {
		end++
	}
	for _, v := range s.V[start:end] {
		for i := 0; i < 8; i++ {
			if v&(1<<uint(i)) != 0 {
				count++
			}
		}
	}
	return count
}

// BitCountByBit counts the number of bits set to 1 by bit
func (s *String) BitCountByBit(start, end int64) int64 {
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

func (s *String) Append(data []byte) int64 {
	s.V = append(s.V, data...)
	return int64(len(s.V))
}

// GetRange returns the substring of the value
func (s *String) GetRange(start, end int64) []byte {
	bl := int64(len(s.V))
	if start < 0 {
		start = bl + start
	}
	if start >= int64(len(s.V)) {
		return nil
	}
	end += 1
	if end <= 0 {
		end += bl
	}
	if end > bl {
		end = bl
	}
	if start > end {
		return nil
	}
	return s.V[start:end]
}

// Strlen returns the length of the value
func (s *String) Strlen() int64 {
	return int64(len(s.V))
}

// SetRange sets the value at the offset
func (s *String) SetRange(offset int64, data []byte) int64 {
	if offset < 0 {
		return 0
	}
	vLen := int64(len(s.V))
	if offset > vLen {
		s.V = append(s.V, make([]byte, offset-vLen...)...)
	}
	if offset+vLen > vLen {
		s.V = append(s.V, make([]byte, offset+vLen-vLen...)...)
	}
	copy(s.V[offset:], data)
	return int64(len(s.V))
}

// GetValue the string to bytes
func (s *String) GetValue() []byte {
	return s.V
}

// SetValue the bytes to string
func (s *String) SetValue(data []byte) {
	s.V = data
}
