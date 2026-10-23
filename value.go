package nodis

import (
	"encoding/binary"
	"math"
	"sort"
	"time"
	"unsafe"
)

type Kind uint8

const (
	KindNull Kind = iota
	KindBool
	KindInt
	KindUint
	KindFloat
	KindString
	KindBytes
	KindTime
	KindArray
	KindObject
)

type NamedValue struct {
	Name  string
	Value Value
}

type Value struct {
	kind Kind
	num  uint64
	text string
	data []byte
	list []Value
	obj  []NamedValue
}

func (v Value) Kind() Kind             { return v.kind }
func (v Value) Bool() (bool, bool)     { return v.num != 0, v.kind == KindBool }
func (v Value) Int() (int64, bool)     { return int64(v.num), v.kind == KindInt }
func (v Value) Uint() (uint64, bool)   { return v.num, v.kind == KindUint }
func (v Value) Float() (float64, bool) { return math.Float64frombits(v.num), v.kind == KindFloat }
func (v Value) String() (string, bool) { return v.text, v.kind == KindString }
func (v Value) Bytes() ([]byte, bool)  { return v.data, v.kind == KindBytes }

func (v Value) Clone() Value {
	clone := v
	if v.data != nil {
		clone.data = append([]byte(nil), v.data...)
	}
	if v.list != nil {
		clone.list = make([]Value, len(v.list))
		for i := range v.list {
			clone.list[i] = v.list[i].Clone()
		}
	}
	if v.obj != nil {
		clone.obj = make([]NamedValue, len(v.obj))
		for i := range v.obj {
			clone.obj[i] = NamedValue{Name: v.obj[i].Name, Value: v.obj[i].Value.Clone()}
		}
	}
	return clone
}

func ValueOf(input any) (Value, error) {
	switch v := input.(type) {
	case nil:
		return Value{kind: KindNull}, nil
	case Value:
		return v, nil
	case bool:
		if v {
			return Value{kind: KindBool, num: 1}, nil
		}
		return Value{kind: KindBool}, nil
	case int:
		return Value{kind: KindInt, num: uint64(int64(v))}, nil
	case int8:
		return Value{kind: KindInt, num: uint64(int64(v))}, nil
	case int16:
		return Value{kind: KindInt, num: uint64(int64(v))}, nil
	case int32:
		return Value{kind: KindInt, num: uint64(int64(v))}, nil
	case int64:
		return Value{kind: KindInt, num: uint64(v)}, nil
	case uint:
		return Value{kind: KindUint, num: uint64(v)}, nil
	case uint8:
		return Value{kind: KindUint, num: uint64(v)}, nil
	case uint16:
		return Value{kind: KindUint, num: uint64(v)}, nil
	case uint32:
		return Value{kind: KindUint, num: uint64(v)}, nil
	case uint64:
		return Value{kind: KindUint, num: v}, nil
	case float32:
		return Value{kind: KindFloat, num: math.Float64bits(float64(v))}, nil
	case float64:
		return Value{kind: KindFloat, num: math.Float64bits(v)}, nil
	case string:
		return Value{kind: KindString, text: v}, nil
	case []byte:
		return Value{kind: KindBytes, data: v}, nil
	case time.Time:
		return Value{kind: KindTime, num: uint64(v.UnixNano())}, nil
	case []Value:
		return Value{kind: KindArray, list: v}, nil
	case []NamedValue:
		return Value{kind: KindObject, obj: v}, nil
	case map[string]any:
		values := make([]NamedValue, 0, len(v))
		for name, item := range v {
			value, err := ValueOf(item)
			if err != nil {
				return Value{}, err
			}
			values = append(values, NamedValue{Name: name, Value: value})
		}
		sort.Slice(values, func(i, j int) bool { return values[i].Name < values[j].Name })
		return Value{kind: KindObject, obj: values}, nil
	default:
		return Value{}, ErrInvalidValue
	}
}

func appendValue(dst []byte, v Value) []byte {
	dst = append(dst, byte(v.kind))
	switch v.kind {
	case KindNull:
		return append(dst, 0)
	case KindBool:
		dst = append(dst, 1, byte(v.num))
		return dst
	case KindInt, KindUint, KindFloat, KindTime:
		dst = append(dst, 8)
		return binary.LittleEndian.AppendUint64(dst, v.num)
	case KindString:
		dst = binary.AppendUvarint(dst, uint64(len(v.text)))
		return append(dst, v.text...)
	case KindBytes:
		dst = binary.AppendUvarint(dst, uint64(len(v.data)))
		return append(dst, v.data...)
	case KindArray:
		start := len(dst)
		dst = append(dst, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
		body := len(dst)
		dst = binary.AppendUvarint(dst, uint64(len(v.list)))
		for i := range v.list {
			dst = appendValue(dst, v.list[i])
		}
		end := len(dst)
		n := binary.PutUvarint(dst[start:body], uint64(end-body))
		copy(dst[start+n:], dst[body:end])
		return dst[:end-(10-n)]
	case KindObject:
		start := len(dst)
		dst = append(dst, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
		body := len(dst)
		dst = binary.AppendUvarint(dst, uint64(len(v.obj)))
		for i := range v.obj {
			dst = binary.AppendUvarint(dst, uint64(len(v.obj[i].Name)))
			dst = append(dst, v.obj[i].Name...)
			dst = appendValue(dst, v.obj[i].Value)
		}
		end := len(dst)
		n := binary.PutUvarint(dst[start:body], uint64(end-body))
		copy(dst[start+n:], dst[body:end])
		return dst[:end-(10-n)]
	default:
		return dst
	}
}

// AppendValue appends the canonical representation of a previously inferred value.
func AppendValue(dst []byte, value Value) []byte { return appendValue(dst, value) }

// DecodeValue decodes one canonical value and returns the number of consumed bytes.
func DecodeValue(data []byte) (Value, int, error) {
	value, consumed, ok := decodeOwnedValue(data)
	if !ok {
		return Value{}, 0, ErrCorruptData
	}
	return value, consumed, nil
}

type ValueView struct {
	kind    Kind
	payload []byte
}

func (v ValueView) Kind() Kind   { return v.kind }
func (v ValueView) IsNull() bool { return v.kind == KindNull }

func (v ValueView) Bool() (bool, bool) {
	return len(v.payload) == 1 && v.kind == KindBool && v.payload[0] != 0, len(v.payload) == 1 && v.kind == KindBool
}

func (v ValueView) Int() (int64, bool) {
	if v.kind != KindInt || len(v.payload) != 8 {
		return 0, false
	}
	return int64(binary.LittleEndian.Uint64(v.payload)), true
}

func (v ValueView) Uint() (uint64, bool) {
	if v.kind != KindUint || len(v.payload) != 8 {
		return 0, false
	}
	return binary.LittleEndian.Uint64(v.payload), true
}

func (v ValueView) Float() (float64, bool) {
	if v.kind != KindFloat || len(v.payload) != 8 {
		return 0, false
	}
	return math.Float64frombits(binary.LittleEndian.Uint64(v.payload)), true
}

func (v ValueView) Time() (time.Time, bool) {
	if v.kind != KindTime || len(v.payload) != 8 {
		return time.Time{}, false
	}
	return time.Unix(0, int64(binary.LittleEndian.Uint64(v.payload))), true
}

// String returns an immutable borrowed view backed by the object's encoded bytes.
func (v ValueView) String() (string, bool) {
	if v.kind != KindString {
		return "", false
	}
	if len(v.payload) == 0 {
		return "", true
	}
	return unsafe.String(unsafe.SliceData(v.payload), len(v.payload)), true
}

// Bytes returns immutable borrowed bytes backed by the object's encoded bytes.
func (v ValueView) Bytes() ([]byte, bool) {
	if v.kind != KindBytes {
		return nil, false
	}
	return v.payload, true
}

func readValue(data []byte) (ValueView, int, bool) {
	if len(data) < 2 {
		return ValueView{}, 0, false
	}
	kind := Kind(data[0])
	n, width := binary.Uvarint(data[1:])
	if width <= 0 || n > uint64(len(data)-1-width) {
		return ValueView{}, 0, false
	}
	if kind > KindObject {
		return ValueView{}, 0, false
	}
	start := 1 + width
	return ValueView{kind: kind, payload: data[start : start+int(n)]}, start + int(n), true
}
