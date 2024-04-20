package redis

import (
	"bufio"
	"io"
	"log"
	"strconv"
	"strings"
)

const (
	StringType  = '+'
	ErrType     = '-'
	IntegerType = ':'
	BulkType    = '$'
	ArrayType   = '*'
	MapType     = '%'
	DoubleType  = ','
	NullType    = '_'
)

var (
	options = map[string]bool{
		"NX":         true,
		"XX":         true,
		"KEEPTTL":    true,
		"GET":        true,
		"LT":         true,
		"GT":         true,
		"CH":         true,
		"INCR":       true,
		"WITHSCORES": true,
	}

	args = map[string]bool{
		"EX":    true,
		"PX":    true,
		"EXAT":  true,
		"PXAT":  true,
		"MATCH": true,
		"COUNT": true,
	}
)

type Value struct {
	typ     uint8
	Str     string
	Integer int64
	Bulk    string
	Double  float64
	Array   []Value
	Map     map[string]Value
	Options map[string]bool
	Args    map[string]Value
}

func StringValue(v string) Value {
	return Value{typ: StringType, Str: v}
}

func ErrorValue(v string) Value {
	return Value{typ: ErrType, Str: v}
}

func BulkValue(v string) Value {
	return Value{typ: BulkType, Bulk: v}
}

func IntegerValue(v int64) Value {
	return Value{typ: IntegerType, Integer: v}
}

func DoubleValue(v float64) Value {
	return Value{typ: DoubleType, Double: v}
}

func ArrayValue(v ...Value) Value {
	return Value{typ: ArrayType, Array: v}
}

func MapValue(v map[string]Value) Value {
	return Value{typ: MapType, Map: v}
}

func NullValue() Value {
	return Value{typ: NullType}
}

type Resp struct {
	reader *bufio.Reader
}

func NewResp(rd io.Reader) *Resp {
	return &Resp{reader: bufio.NewReader(rd)}
}

func (r *Resp) readLine() (line []byte, n int, err error) {
	for {
		b, err := r.reader.ReadByte()
		if err != nil {
			return nil, 0, err
		}
		n += 1
		line = append(line, b)
		if len(line) >= 2 && line[len(line)-2] == '\r' {
			break
		}
	}
	return line[:len(line)-2], n, nil
}

func (r *Resp) readInteger() (x int, n int, err error) {
	line, n, err := r.readLine()
	if err != nil {
		return 0, 0, err
	}
	i64, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return 0, n, err
	}
	return int(i64), n, nil
}

func (r *Resp) Read() (Value, error) {
	_type, err := r.reader.ReadByte()

	if err != nil {
		return Value{}, err
	}

	switch _type {
	case ArrayType:
		return r.readArray()
	case BulkType:
		return r.readBulk()
	default:
		log.Printf("Unknown type: %v \n", string(_type))
		return Value{}, nil
	}
}

func (r *Resp) readArray() (Value, error) {
	v := Value{
		Options: make(map[string]bool),
		Args:    make(map[string]Value),
	}
	v.typ = ArrayType

	// read length of array
	l, _, err := r.readInteger()
	if err != nil {
		return v, err
	}

	// foreach line, parse and read the value
	v.Array = make([]Value, 0)
	var arg string
	for i := 0; i < l; i++ {
		val, err := r.Read()
		if err != nil {
			return v, err
		}
		if arg != "" {
			v.Args[arg] = val
			arg = ""
			continue
		}
		b := strings.ToUpper(val.Bulk)
		// options are special case
		if options[b] && i != 0 {
			v.Options[b] = true
			continue
		}
		// args are special case
		if args[b] {
			arg = b
			continue
		}
		// append parsed value to array
		v.Array = append(v.Array, val)
	}

	return v, nil
}

func (r *Resp) readBulk() (Value, error) {
	v := Value{}

	v.typ = BulkType

	l, _, err := r.readInteger()
	if err != nil {
		return v, err
	}

	bulk := make([]byte, l)

	r.reader.Read(bulk)

	v.Bulk = string(bulk)

	// Read the trailing CRLF
	r.readLine()

	return v, nil
}

// Marshal Value to bytes
func (v Value) Marshal() []byte {
	switch v.typ {
	case ArrayType:
		return v.marshalArray()
	case BulkType:
		return v.marshalBulk()
	case StringType:
		return v.marshalString()
	case NullType:
		return v.marshallNull()
	case ErrType:
		return v.marshallError()
	case IntegerType:
		return v.marshallInteger()
	case MapType:
		return v.marshallMap()
	case DoubleType:
		return v.marshallDouble()
	default:
		return []byte{}
	}
}

func (v Value) marshalString() []byte {
	var bytes []byte
	bytes = append(bytes, StringType)
	bytes = append(bytes, v.Str...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) marshalBulk() []byte {
	var bytes []byte
	bytes = append(bytes, BulkType)
	bytes = append(bytes, strconv.Itoa(len(v.Bulk))...)
	bytes = append(bytes, '\r', '\n')
	bytes = append(bytes, v.Bulk...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) marshalArray() []byte {
	l := len(v.Array)
	var bytes []byte
	bytes = append(bytes, ArrayType)
	bytes = append(bytes, strconv.Itoa(l)...)
	bytes = append(bytes, '\r', '\n')

	for i := 0; i < l; i++ {
		bytes = append(bytes, v.Array[i].Marshal()...)
	}

	return bytes
}

func (v Value) marshallError() []byte {
	var bytes []byte
	bytes = append(bytes, ErrType)
	bytes = append(bytes, v.Str...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) marshallNull() []byte {
	return []byte("$-1\r\n")
}

func (v Value) marshallInteger() []byte {
	var bytes []byte
	bytes = append(bytes, IntegerType)
	bytes = append(bytes, strconv.FormatInt(v.Integer, 10)...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func (v Value) marshallDouble() []byte {
	var bytes []byte
	bytes = append(bytes, DoubleType)
	bytes = append(bytes, strconv.FormatFloat(v.Double, 'f', -1, 64)...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func (v Value) marshallMap() []byte {
	var bytes []byte
	bytes = append(bytes, MapType)
	bytes = append(bytes, strconv.Itoa(len(v.Map))...)
	bytes = append(bytes, '\r', '\n')
	for k, v := range v.Map {
		bytes = append(bytes, StringValue(k).Marshal()...)
		bytes = append(bytes, v.Marshal()...)
	}
	return bytes
}

// Writer

type Writer struct {
	writer io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w}
}

func (w *Writer) Write(v Value) error {
	var bytes = v.Marshal()
	_, err := w.writer.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}
