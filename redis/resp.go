package redis

import (
	"errors"
	"io"
	"net"
	"strconv"
	"unsafe"

	"github.com/diiyw/nodis/internal/strings"
)

type Reader struct {
	reader io.Reader
	buf    []byte
	// Read index
	r int
	// Has read length
	l   int
	cmd Command
}

const defaultSize = 4096

func NewReader(rd io.Reader) *Reader {
	return &Reader{
		reader: rd,
		buf:    make([]byte, defaultSize*2),
	}
}

func (r *Reader) grow(n int) {
	if r.r+r.l+n <= len(r.buf) {
		return
	}
	c := r.r + r.l + n
	newBuf := make([]byte, c)
	copy(newBuf, r.buf)
	r.buf = newBuf
}

func (r *Reader) readByte() error {
	size := r.r + r.l
	r.grow(defaultSize)
	n, err := r.reader.Read(r.buf[size : size+1])
	if err != nil {
		return err
	}
	if n == 0 {
		return io.EOF
	}
	r.l++
	return nil
}

func (r *Reader) readByteN(n int) error {
	r.grow(n)
READ:
	start := r.r + r.l
	end := start + (n - r.l)
	rn, err := r.reader.Read(r.buf[start:end])
	if err != nil {
		return err
	}
	r.l += rn
	if r.l < n {
		goto READ
	}
	return nil
}

func (r *Reader) peekByte(i int) byte {
	if i >= 0 {
		return r.buf[r.r+i]
	}
	return r.buf[r.r+r.l+i]
}

func (r *Reader) String() string {
	s := unsafe.String(unsafe.SliceData(r.buf[r.r:r.r+r.l]), r.l)
	r.discard()
	return s
}

func (r *Reader) reset() {
	r.discard()
	r.cmd = Command{
		Args: make([]string, 0),
	}
}

func (r *Reader) discard() {
	r.r += r.l
	r.l = 0
}

func (r *Reader) readLine() error {
	for {
		err := r.readByte()
		if err != nil {
			return err
		}
		if r.l > 1 && r.peekByte(-1) == '\n' {
			r.l -= 2
			break
		}
	}
	return nil
}

func (r *Reader) readInteger() (x int, err error) {
	err = r.readLine()
	if err != nil {
		return 0, err
	}
	i64, err := strconv.ParseInt(r.String(), 10, 64)
	if err != nil {
		return 0, err
	}
	return int(i64), nil
}

var (
	ErrInvalidRequestExceptedArray       = errors.New("invalid request, expected array")
	ErrInvalidRequestExceptedArrayLength = errors.New("invalid request, expected array length")
	ErrInvalidRequestExceptedBulk        = errors.New("invalid request, expected bulk")
)

func (r *Reader) ReadCommand() error {
	r.reset()
	if r.peekByte(0) != ArrayType {
		return r.ReadInlineCommand()
	}
	// Read resp type
	err := r.readByte()
	if err != nil {
		return err
	}
	r.discard()
	// Read length of array
	l, err := r.readInteger()
	if err != nil {
		return ErrInvalidRequestExceptedArrayLength
	}
	var v string
	for i := 0; i < l; i++ {
		// Read first args, it should be command name
		v, err = r.readBulk()
		if err != nil {
			return ErrInvalidRequestExceptedArray
		}
		if i == 0 {
			r.cmd.Name = strings.ToUpper(v)
			continue
		}
		r.cmd.Args = append(r.cmd.Args, v)
		r.readOptions(v, i-1)
	}
	return nil
}

func (r *Reader) readOptions(v string, i int) {
	opt := strings.ToUpper(v)
	switch opt {
	case "NX":
		r.cmd.Options.NX = i
	case "XX":
		r.cmd.Options.XX = i
	case "LT":
		r.cmd.Options.LT = i
	case "GT":
		r.cmd.Options.GT = i
	case "MATCH":
		r.cmd.Options.MATCH = i + 1
	case "COUNT":
		r.cmd.Options.COUNT = i + 1
	case "TYPE":
		r.cmd.Options.TYPE = i + 1
	case "EX":
		r.cmd.Options.EX = i + 1
	case "EXAT":
		r.cmd.Options.EXAT = i + 1
	case "PX":
		r.cmd.Options.PX = i + 1
	case "PXAT":
		r.cmd.Options.PXAT = i + 1
	case "GET":
		r.cmd.Options.GET = i
	case "KEEPTTL":
		r.cmd.Options.KEEPTTL = i
	case "CH":
		r.cmd.Options.CH = i
	case "INCR":
		r.cmd.Options.INCR = i
	case "WITHSCORES":
		r.cmd.Options.WITHSCORES = i
	case "LIMIT":
		r.cmd.Options.LIMIT = i + 1
	case "BYSCORE":
		r.cmd.Options.BYSCORE = i
	case "BYLEX":
		r.cmd.Options.BYLEX = i
	case "REV":
		r.cmd.Options.REV = i
	case "WEIGHTS":
		r.cmd.Options.WEIGHTS = i + 1
	case "AGGREGATE":
		r.cmd.Options.AGGREGATE = i + 1
	case "BYTE":
		r.cmd.Options.BYTE = i
	case "BIT":
		r.cmd.Options.BIT = i
	case "KM":
		r.cmd.Options.KM = i
	case "M":
		r.cmd.Options.M = i
	case "FT":
		r.cmd.Options.FT = i
	case "MI":
		r.cmd.Options.MI = i
	case "ASC":
		r.cmd.Options.ASC = i
	case "DESC":
		r.cmd.Options.DESC = i
	case "ANY":
		r.cmd.Options.ANY = i
	case "WITHDIST":
		r.cmd.Options.WITHDIST = i
	case "WITHCOORD":
		r.cmd.Options.WITHCOORD = i
	case "WITHHASH":
		r.cmd.Options.WITHHASH = i
	}
}

func (r *Reader) readBulk() (string, error) {
	err := r.readByte()
	if err != nil {
		return "", err
	}
	if r.peekByte(0) != BulkType {
		return "", ErrInvalidRequestExceptedBulk
	}
	r.discard()
	l, err := r.readInteger()
	if err != nil {
		return "", err
	}
	err = r.readByteN(l)
	if err != nil {
		return "", err
	}
	v := r.String()
	// Read the trailing CRLF
	err = r.readLine()
	if err != nil {
		return "", err
	}
	r.discard()
	return v, nil
}

func (r *Reader) readUtil(end byte) (bool, error) {
	var lineEnd bool
	for {
		err := r.readByte()
		if err != nil {
			return lineEnd, err
		}
		if r.peekByte(-1) == '\r' {
			r.l--
			continue
		}
		if r.peekByte(-1) == '\n' {
			lineEnd = true
			r.l--
			break
		}
		if r.peekByte(-1) == end && r.peekByte(-2) != '\\' {
			r.l--
			break
		}
	}
	return lineEnd, nil
}

func (r *Reader) ReadInlineCommand() error {
	var index = 0
	var lineEnd = false
	var err error
	for {
		err = r.readByte()
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}
		first := r.peekByte(0)
		if first == ' ' || first == '\t' || first == '\r' {
			r.discard()
			continue
		}
		if first == '\n' {
			r.discard()
			lineEnd = true
			break
		}
		if first == '\'' || first == '"' {
			r.discard()
			lineEnd, err = r.readUtil(first)
		} else {
			lineEnd, err = r.readUtil(' ')
		}
		if err != nil {
			return err
		}
		v := r.String()
		if index == 0 {
			r.cmd.Name = strings.ToUpper(v)
		} else {
			r.cmd.Args = append(r.cmd.Args, v)
			r.readOptions(v, index)
		}
		index++
		if lineEnd {
			break
		}
	}
	return nil
}

// Writer is a RESP writer
type Writer struct {
	writer io.Writer
	buf    []byte
	w      int
	err    bool
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: w,
		buf:    make([]byte, defaultSize*2),
	}
}

func (w *Writer) grow(n int) {
	newBuf := make([]byte, len(w.buf)+n)
	copy(newBuf, w.buf)
	w.buf = newBuf
}

func (w *Writer) Bytes() []byte {
	return w.buf[:w.w]
}

func (w *Writer) Flush() error {
	_, err := w.writer.Write(w.buf[:w.w])
	if err != nil {
		return err
	}
	w.w = 0
	w.err = false
	return nil
}

func (w *Writer) HasError() bool {
	return w.err
}

func (w *Writer) writeByte(b byte) {
	if w.w >= len(w.buf) {
		w.grow(defaultSize)
	}
	w.buf[w.w] = b
	w.w++
}

func (w *Writer) writeBytes(bs ...byte) {
	n := len(bs)
	if w.w+n >= len(w.buf) {
		w.grow(n)
	}
	for _, v := range bs {
		w.writeByte(v)
	}
}

func (w *Writer) WriteString(str string) {
	w.writeByte(StringType)
	w.writeBytes(strings.String2Bytes(str)...)
	w.writeBytes('\r', '\n')
}

func (w *Writer) WriteBulk(bulk string) {
	w.writeByte(BulkType)
	w.writeBytes(strings.String2Bytes(strconv.Itoa(len(bulk)))...)
	w.writeBytes('\r', '\n')
	w.writeBytes(strings.String2Bytes(bulk)...)
	w.writeBytes('\r', '\n')
}

func (w *Writer) WriteArray(l int) {
	w.writeByte(ArrayType)
	w.writeBytes(strings.String2Bytes(strconv.Itoa(l))...)
	w.writeBytes('\r', '\n')
}

func (w *Writer) WriteError(err string) {
	w.err = true
	w.writeByte(ErrType)
	w.writeBytes(strings.String2Bytes(err)...)
	w.writeBytes('\r', '\n')
}

func (w *Writer) WriteBulkNull() {
	w.writeBytes([]byte("$-1\r\n")...)
}

func (w *Writer) WriteArrayNull() {
	w.writeBytes([]byte("*-1\r\n")...)
}

func (w *Writer) WriteInt64(v int64) {
	w.writeByte(IntegerType)
	w.writeBytes(strings.String2Bytes(strconv.FormatInt(v, 10))...)
	w.writeBytes('\r', '\n')
}

func (w *Writer) WriteUInt64(v uint64) {
	w.writeByte(IntegerType)
	w.writeBytes(strings.String2Bytes(strconv.FormatUint(v, 10))...)
	w.writeBytes('\r', '\n')
}

func (w *Writer) WriteDouble(v float64) {
	w.writeByte(DoubleType)
	w.writeBytes(strings.String2Bytes(strconv.FormatFloat(v, 'f', -1, 64))...)
	w.writeBytes('\r', '\n')
}

func (w *Writer) WriteMap(n int) {
	w.writeByte(MapType)
	w.writeBytes(strings.String2Bytes(strconv.Itoa(n))...)
	w.writeBytes('\r', '\n')
}

func (w *Writer) WriteNullMap() {
	w.writeBytes([]byte("%-1\r\n")...)
}

func (w *Writer) WriteOK() {
	w.writeBytes([]byte("+OK\r\n")...)
}

func (w *Writer) RemoteAddr() string {
	if c, ok := w.writer.(net.Conn); ok {
		return c.RemoteAddr().String()
	}
	return ""
}
