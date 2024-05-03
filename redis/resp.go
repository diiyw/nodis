package redis

import (
	"errors"
	"io"
	"net"
	"strconv"
	"unsafe"

	"github.com/diiyw/nodis/utils"
)

type Reader struct {
	reader io.Reader
	buf    []byte
	r      int
	l      int
	cmd    *Command
}

const defaultSize = 4096

func NewReader(rd io.Reader) *Reader {
	return &Reader{
		reader: rd,
		buf:    make([]byte, defaultSize),
		cmd: &Command{
			Args: make([]string, 0),
		},
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
	n, err := r.reader.Read(r.buf[size : r.r+r.l+1])
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

func (r *Reader) firstByte() byte {
	return r.indexByte(0)
}

func (r *Reader) indexByte(i int) byte {
	if i >= 0 {
		return r.buf[r.r+i]
	}
	return r.buf[r.r+r.l+i]
}

func (r *Reader) flushNoCopyString() string {
	s := unsafe.String(unsafe.SliceData(r.buf[r.r:r.r+r.l]), r.l)
	r.malloc()
	return s
}

func (r *Reader) flushCopyString() string {
	s := string(r.buf[r.r : r.r+r.l])
	r.malloc()
	return s
}

func (r *Reader) reset() {
	r.r = 0
	r.l = 0
	r.cmd.Args = make([]string, 0)
	r.cmd.Options = Options{}
}

func (r *Reader) malloc() {
	r.r += r.l
	r.l = 0
}

func (r *Reader) readLine() error {
	for {
		err := r.readByte()
		if err != nil {
			return err
		}
		if r.l > 1 && r.indexByte(-1) == '\n' {
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
	i64, err := strconv.ParseInt(r.flushNoCopyString(), 10, 64)
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
	// Read resp type
	err := r.readByte()
	if err != nil {
		return err
	}
	if r.firstByte() != ArrayType {
		return r.ReadInlineCommand()
	}
	r.malloc()
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
			r.cmd.Name = utils.ToUpper(v)
			continue
		}
		r.cmd.Args = append(r.cmd.Args, v)
		r.readOptions(v, i-1)
	}
	return nil
}

func (r *Reader) readOptions(v string, i int) {
	switch r.cmd.Name {
	case "EXPIRE", "EXPIREAT":
		if i > 1 {
			opt := utils.ToUpper(v)
			switch opt {
			case "NX":
				r.cmd.Options.NX = i
			case "XX":
				r.cmd.Options.XX = i
			case "LT":
				r.cmd.Options.LT = i
			case "GT":
				r.cmd.Options.GT = i
			}
		}
	case "SCAN", "SSCAN", "HSCAN", "ZSCAN", "SPOP":
		if i > 1 {
			opt := utils.ToUpper(v)
			switch opt {
			case "MATCH":
				r.cmd.Options.MATCH = i + 1
			case "COUNT":
				r.cmd.Options.COUNT = i + 1
			case "TYPE":
			}
		}
	case "SET":
		if i > 1 {
			opt := utils.ToUpper(v)
			switch opt {
			case "NX":
				r.cmd.Options.NX = i
			case "XX":
				r.cmd.Options.XX = i
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
			}
		}
	case "ZADD":
		if i > 0 {
			opt := utils.ToUpper(v)
			switch opt {
			case "NX":
				r.cmd.Options.NX = i
			case "XX":
				r.cmd.Options.XX = i
			case "CH":
				r.cmd.Options.CH = i
			case "INCR":
				r.cmd.Options.INCR = i
			case "GT":
				r.cmd.Options.GT = i
			case "LT":
				r.cmd.Options.LT = i
			}
		}
	case "ZRANK", "ZREVRANK", "ZRANGE", "ZREVRANGE", "ZRANGEBYSCORE", "ZREVRANGEBYSCORE":
		if r.cmd.Name == "ZRANK" && i < 1 {
			return
		} else if i < 2 {
			return
		}
		opt := utils.ToUpper(v)
		switch opt {
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
		}
	case "ZUNIONSTORE", "ZINTERSTORE":
		if i > 2 {
			opt := utils.ToUpper(v)
			switch opt {
			case "WEIGHTS":
				r.cmd.Options.WEIGHTS = i + 1
			case "AGGREGATE":
				r.cmd.Options.AGGREGATE = i + 1
			}
		}
	case "CONFIG":
		opt := utils.ToUpper(v)
		switch opt {
		case "GET":
			r.cmd.Options.GET = i + 1
		}
	case "BITCOUNT":
		if i > 2 {
			opt := utils.ToUpper(v)
			switch opt {
			case "BYTE":
				r.cmd.Options.BYTE = i
			case "BIT":
				r.cmd.Options.BIT = i
			}
		}
	}
}

func (r *Reader) readBulk() (string, error) {
	err := r.readByte()
	if err != nil {
		return "", err
	}
	if r.firstByte() != BulkType {
		return "", ErrInvalidRequestExceptedBulk
	}
	r.malloc()
	l, err := r.readInteger()
	if err != nil {
		return "", err
	}
	r.readByteN(l)
	v := r.flushCopyString()
	// Read the trailing CRLF
	r.readLine()
	r.malloc()
	return v, nil
}

func (r *Reader) readUtil(end byte) (bool, error) {
	var lineEnd bool
	for {
		err := r.readByte()
		if err != nil {
			return lineEnd, err
		}
		if r.indexByte(-1) == '\r' {
			r.l--
			continue
		}
		if r.indexByte(-1) == '\n' {
			lineEnd = true
			r.l--
			break
		}
		if r.indexByte(-1) == end && r.indexByte(-2) != '\\' {
			r.l--
			break
		}
	}
	return lineEnd, nil
}

func (r *Reader) ReadInlineCommand() error {
	lineEnd, err := r.readUtil(' ')
	if err != nil {
		return err
	}
	r.cmd.Name = utils.ToUpper(r.flushNoCopyString())
	if lineEnd {
		return nil
	}
	var i = 0
	for {
		err := r.readByte()
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}
		first := r.firstByte()
		if first == ' ' || first == '\t' {
			r.malloc()
			continue
		}
		if first == '\'' || first == '"' {
			r.malloc()
			lineEnd, err = r.readUtil(first)
		} else {
			lineEnd, err = r.readUtil(' ')
		}
		if err != nil {
			return err
		}
		v := r.flushCopyString()
		r.cmd.Args = append(r.cmd.Args, v)
		i++
		r.readOptions(v, i)
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
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: w,
		buf:    make([]byte, defaultSize),
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
	return nil
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
	w.writeBytes(utils.String2Bytes(str)...)
	w.writeBytes('\r', '\n')
}

func (w *Writer) WriteBulk(bulk string) {
	w.writeByte(BulkType)
	w.writeBytes(utils.String2Bytes(strconv.Itoa(len(bulk)))...)
	w.writeBytes('\r', '\n')
	w.writeBytes(utils.String2Bytes(bulk)...)
	w.writeBytes('\r', '\n')
}

func (w *Writer) WriteArray(l int) {
	w.writeByte(ArrayType)
	w.writeBytes(utils.String2Bytes(strconv.Itoa(l))...)
	w.writeBytes('\r', '\n')
}

func (w *Writer) WriteError(err string) {
	w.writeByte(ErrType)
	w.writeBytes(utils.String2Bytes(err)...)
	w.writeBytes('\r', '\n')
}

func (w *Writer) WriteNull() {
	w.writeBytes([]byte("$-1\r\n")...)
}

func (w *Writer) WriteInteger(v int64) {
	w.writeByte(IntegerType)
	w.writeBytes(utils.String2Bytes(strconv.FormatInt(v, 10))...)
	w.writeBytes('\r', '\n')
}

func (w *Writer) WriteDouble(v float64) {
	w.writeByte(DoubleType)
	w.writeBytes(utils.String2Bytes(strconv.FormatFloat(v, 'f', -1, 64))...)
	w.writeBytes('\r', '\n')
}

func (w *Writer) WriteMap(n int) {
	w.writeByte(MapType)
	w.writeBytes(utils.String2Bytes(strconv.Itoa(n))...)
	w.writeBytes('\r', '\n')
}

func (w *Writer) WriteNullMap() {
	w.writeBytes([]byte("%-1\r\n")...)
}

func (w *Writer) WriteOK() {
	w.writeBytes([]byte("+OK\r\n")...)
}

// Close closes the writer
func (w *Writer) Close() error {
	return w.writer.(net.Conn).Close()
}

func (w *Writer) RemoteAddr() string {
	return w.writer.(net.Conn).RemoteAddr().String()
}
