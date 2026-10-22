package redis

import (
	"bufio"
	"errors"
	internalstrings "github.com/diiyw/nodis/internal/strings"
	"io"
	"net"
	"strconv"
	"strings"
	"unsafe"
)

const defaultSize = 2048
const MaxRequestSize = 64 << 20
const MaxArguments = 1 << 20

var (
	ErrInvalidRequestExceptedArray       = errors.New("ERR invalid request, expected array")
	ErrInvalidRequestExceptedArrayLength = errors.New("ERR invalid array length")
	ErrInvalidRequestExceptedBulk        = errors.New("ERR invalid bulk length")
	ErrSyntax                            = errors.New("ERR syntax error")
)

type Reader struct {
	reader   *bufio.Reader
	buf      []byte
	cmd      Command
	reusable bool
}

func NewReader(rd io.Reader) *Reader {
	return &Reader{reader: bufio.NewReaderSize(rd, defaultSize*2), reusable: true}
}
func (r *Reader) line() ([]byte, error) {
	b, err := r.reader.ReadSlice('\n')
	if err != nil {
		return nil, err
	}
	if len(b) < 2 || b[len(b)-2] != '\r' {
		return nil, ErrSyntax
	}
	return b[:len(b)-2], nil
}
func (r *Reader) number() (int, error) {
	b, err := r.line()
	if err != nil {
		return 0, err
	}
	if len(b) == 0 {
		return 0, ErrSyntax
	}
	n, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil || n < 0 || n > MaxRequestSize {
		return 0, ErrInvalidRequestExceptedBulk
	}
	return int(n), nil
}
func (r *Reader) bulk() (string, error) {
	prefix, err := r.reader.ReadByte()
	if err != nil {
		return "", err
	}
	if prefix != '$' {
		return "", ErrInvalidRequestExceptedBulk
	}
	n, err := r.number()
	if err != nil {
		return "", err
	}
	if len(r.buf) > MaxRequestSize-n {
		return "", ErrInvalidRequestExceptedBulk
	}
	start := len(r.buf)
	r.buf = append(r.buf, make([]byte, n)...)
	if _, err = io.ReadFull(r.reader, r.buf[start:]); err != nil {
		return "", err
	}
	a, err := r.reader.ReadByte()
	if err != nil {
		return "", err
	}
	b, err := r.reader.ReadByte()
	if err != nil {
		return "", err
	}
	if a != '\r' || b != '\n' {
		return "", ErrSyntax
	}
	return unsafe.String(unsafe.SliceData(r.buf[start:]), n), nil
}
func (r *Reader) ReadCommand() error {
	if r.reusable {
		r.buf = r.buf[:0]
	} else {
		r.buf = nil
	}
	r.cmd.Name = ""
	clear(r.cmd.Args)
	r.cmd.Args = r.cmd.Args[:0]
	r.cmd.Options = Options{}
	prefix, err := r.reader.ReadByte()
	if err != nil {
		return err
	}
	if prefix != '*' {
		if err = r.reader.UnreadByte(); err != nil {
			return err
		}
		return r.ReadInlineCommand()
	}
	count, err := r.number()
	if err != nil || count == 0 || count > MaxArguments {
		return ErrInvalidRequestExceptedArrayLength
	}
	for i := 0; i < count; i++ {
		value, err := r.bulk()
		if err != nil {
			return err
		}
		if i == 0 {
			r.cmd.Name = strings.ToUpper(value)
		} else {
			r.cmd.Args = append(r.cmd.Args, value)
		}
	}
	r.options()
	r.reusable = r.cmd.Name == "GET" || r.cmd.Name == "PING" || r.cmd.Name == "EXISTS" || r.cmd.Name == "TTL" || r.cmd.Name == "PTTL"
	return nil
}
func (r *Reader) ReadInlineCommand() error {
	line, err := r.line()
	if err != nil {
		return err
	}
	i := 0
	for i < len(line) {
		for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
			i++
		}
		if i == len(line) {
			break
		}
		start := len(r.buf)
		var quote byte
		if line[i] == '\'' || line[i] == '"' {
			quote = line[i]
			i++
		}
		closed := quote == 0
		for i < len(line) {
			b := line[i]
			if quote != 0 && b == quote {
				i++
				closed = true
				break
			}
			if quote == 0 && (b == ' ' || b == '\t') {
				break
			}
			if b == '\\' && i+1 < len(line) {
				i++
				b = line[i]
			}
			r.buf = append(r.buf, b)
			i++
		}
		if !closed || i < len(line) && line[i] != ' ' && line[i] != '\t' {
			return ErrSyntax
		}
		value := string(r.buf[start:])
		if r.cmd.Name == "" {
			r.cmd.Name = strings.ToUpper(value)
		} else {
			r.cmd.Args = append(r.cmd.Args, value)
		}
	}
	if r.cmd.Name == "" {
		return ErrSyntax
	}
	r.options()
	return nil
}

// Interpret option words only in the grammar's option region, never in key/value data.
func (r *Reader) options() {
	start := len(r.cmd.Args)
	switch r.cmd.Name {
	case "SET":
		start = 2
	case "SCAN":
		start = 1
	case "SSCAN", "HSCAN", "ZSCAN", "EXPIRE", "EXPIREAT":
		start = 2
	case "ZADD", "GEOADD":
		for i := 1; i < len(r.cmd.Args); i++ {
			v := strings.ToUpper(r.cmd.Args[i])
			switch v {
			case "NX", "XX", "LT", "GT", "CH", "INCR":
				r.readOptions(v, i)
			default:
				return
			}
		}
		return
	case "ZRANGE", "ZREVRANGE", "ZRANGEBYSCORE", "ZREVRANGEBYSCORE", "BITCOUNT":
		start = 3
	case "GEORADIUS":
		start = 4
	case "GEORADIUSBYMEMBER":
		start = 3
	case "ZUNIONSTORE", "ZINTERSTORE":
		if len(r.cmd.Args) > 1 {
			n, err := strconv.Atoi(r.cmd.Args[1])
			if err == nil && n >= 0 && n <= len(r.cmd.Args)-2 {
				start = 2 + n
			}
		}
	}
	for i := start; i < len(r.cmd.Args); i++ {
		r.readOptions(r.cmd.Args[i], i)
		switch strings.ToUpper(r.cmd.Args[i]) {
		case "EX", "PX", "EXAT", "PXAT", "MATCH", "COUNT", "TYPE", "AGGREGATE":
			i++
		case "LIMIT":
			i += 2
		case "WEIGHTS":
			if len(r.cmd.Args) > 1 {
				n, _ := strconv.Atoi(r.cmd.Args[1])
				if n > 0 && n < len(r.cmd.Args) {
					i += n
				}
			}
		}
	}
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
	if w.w+n <= len(w.buf) {
		return
	}
	newBuf := make([]byte, len(w.buf)+n)
	copy(newBuf, w.buf)
	w.buf = newBuf
}

func (w *Writer) Bytes() []byte {
	return w.buf[:w.w]
}

func (w *Writer) Reset() {
	w.w = 0
	w.err = false
}

func (w *Writer) Push() error {
	defer w.Reset()
	b := w.buf[:w.w]
	for len(b) > 0 {
		n, err := w.writer.Write(b)
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
		b = b[n:]
	}
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
	w.grow(n)
	for _, v := range bs {
		w.writeByte(v)
	}
}

func (w *Writer) WriteString(str string) {
	w.writeByte(StringType)
	w.writeBytes(internalstrings.String2Bytes(str)...)
	w.writeBytes('\r', '\n')
}

func (w *Writer) WriteBulk(bulk string) {
	w.writeByte(BulkType)
	w.writeBytes(internalstrings.String2Bytes(strconv.Itoa(len(bulk)))...)
	w.writeBytes('\r', '\n')
	w.writeBytes(internalstrings.String2Bytes(bulk)...)
	w.writeBytes('\r', '\n')
}

func (w *Writer) WriteArray(l int) {
	w.writeByte(ArrayType)
	w.writeBytes(internalstrings.String2Bytes(strconv.Itoa(l))...)
	w.writeBytes('\r', '\n')
}

func (w *Writer) WriteError(err string) {
	w.err = true
	w.writeByte(ErrType)
	w.writeBytes(internalstrings.String2Bytes(err)...)
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
	w.writeBytes(internalstrings.String2Bytes(strconv.FormatInt(v, 10))...)
	w.writeBytes('\r', '\n')
}

func (w *Writer) WriteUInt64(v uint64) {
	w.writeByte(IntegerType)
	w.writeBytes(internalstrings.String2Bytes(strconv.FormatUint(v, 10))...)
	w.writeBytes('\r', '\n')
}

func (w *Writer) WriteDouble(v float64) {
	w.writeByte(DoubleType)
	w.writeBytes(internalstrings.String2Bytes(strconv.FormatFloat(v, 'f', -1, 64))...)
	w.writeBytes('\r', '\n')
}

func (w *Writer) WriteMap(n int) {
	w.writeByte(MapType)
	w.writeBytes(internalstrings.String2Bytes(strconv.Itoa(n))...)
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
