package redis

import (
	"bufio"
	"io"
	"strconv"
	"unsafe"

	"github.com/diiyw/nodis/utils"
)

type Resp struct {
	reader *bufio.Reader
	buf    []byte
	r      int
	l      int
}

const defaultSize = 4096

func NewResp(rd io.Reader) *Resp {
	return &Resp{
		reader: bufio.NewReader(rd),
		buf:    make([]byte, defaultSize),
	}
}

func (r *Resp) grow() {
	r.buf = append(r.buf, make([]byte, defaultSize)...)
}

func (r *Resp) readByte() error {
	b, err := r.reader.ReadByte()
	if err != nil {
		return err
	}
	r.writeBuff(b)
	return nil
}

func (r *Resp) readByteN(n int) error {
	if r.r+r.l+n > defaultSize {
		r.grow()
	}
	_, err := r.reader.Read(r.buf[r.r : r.r+n])
	if err != nil {
		return err
	}
	r.l += n
	return nil
}

func (r *Resp) writeBuff(b byte) {
	if r.r+r.l >= defaultSize {
		r.grow()
	}
	r.buf[r.r+r.l] = b
	r.l++
}

func (r *Resp) bufFirst() byte {
	return r.bufIndex(0)
}

func (r *Resp) bufIndex(i int) byte {
	if i >= 0 {
		return r.buf[r.r+i]
	}
	return r.buf[r.r+r.l+i]
}

func (r *Resp) flushString() string {
	s := unsafe.String(unsafe.SliceData(r.buf[r.r:r.r+r.l]), r.l)
	r.malloc()
	return s
}

func (r *Resp) reset() {
	r.r = 0
	r.l = 0
}

func (r *Resp) malloc() {
	r.r += r.l
	r.l = 0
}

func (r *Resp) readLine() error {
	for {
		err := r.readByte()
		if err != nil {
			return err
		}
		if r.l > 1 && r.bufIndex(-1) == '\n' {
			r.l -= 2
			break
		}
	}
	return nil
}

func (r *Resp) readInteger() (x int, err error) {
	err = r.readLine()
	if err != nil {
		return 0, err
	}
	i64, err := strconv.ParseInt(r.flushString(), 10, 64)
	if err != nil {
		return 0, err
	}
	return int(i64), nil
}

func (r *Resp) Read() (Value, error) {
	r.malloc()
	err := r.readByte()
	if err != nil {
		return Value{}, err
	}

	switch r.bufFirst() {
	case ArrayType:
		r.malloc()
		return r.readArray()
	case BulkType:
		r.malloc()
		return r.readBulk()
	default:
		return r.readInline()
	}
}

func (r *Resp) readArray() (Value, error) {
	v := Value{
		Options: make(map[string]bool),
		Args:    make(map[string]Value),
	}
	v.typ = ArrayType

	// read length of array
	l, err := r.readInteger()
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
		b := utils.ToUpper(val.Bulk)
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

	l, err := r.readInteger()
	if err != nil {
		return v, err
	}

	r.readByteN(l)

	v.Bulk = r.flushString()

	// Read the trailing CRLF
	r.readLine()
	r.malloc()

	return v, nil
}

func (r *Resp) readUtil(end byte) (isEnd bool, err error) {
	for {
		err = r.readByte()
		if err != nil {
			return
		}
		if r.bufIndex(-1) == '\r' {
			r.l--
			continue
		}
		if r.bufIndex(-1) == '\n' {
			isEnd = true
			r.l--
			break
		}
		if r.bufIndex(-1) == end && r.bufIndex(-2) != '\\' {
			r.l--
			break
		}
	}
	return isEnd, nil
}

func (r *Resp) readInline() (Value, error) {
	v := Value{
		typ:   ArrayType,
		Array: make([]Value, 0),
	}
	var c = Value{
		Options: make(map[string]bool),
		Args:    make(map[string]Value),
	}
	isEnd, err := r.readUtil(' ')
	if err != nil {
		return v, err
	}
	c.Bulk = r.flushString()
	v.Array = append(v.Array, c)
	if isEnd {
		return v, nil
	}
	var arg string
	for {
		err := r.readByte()
		if err != nil {
			if err != io.EOF {
				return v, err
			}
			break
		}
		first := r.bufFirst()
		if first == ' ' || first == '\t' {
			r.malloc()
			continue
		}
		if first == '\'' || first == '"' {
			r.malloc()
			isEnd, err = r.readUtil(first)
		} else {
			isEnd, err = r.readUtil(' ')
		}
		if err != nil {
			return v, err
		}
		strV := r.flushString()
		if arg != "" {
			c.Args[arg] = BulkValue(strV)
			arg = ""
			continue
		}
		b := utils.ToUpper(strV)
		// options are special case
		if options[b] {
			c.Options[b] = true
			continue
		}
		// args are special case
		if args[b] {
			arg = b
			continue
		}
		// append parsed value to array
		v.Array = append(v.Array, BulkValue(strV))
		if isEnd {
			break
		}
	}
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
