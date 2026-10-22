package redis

import (
	"bytes"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

func TestRejectMalformedRESP(t *testing.T) {
	for _, request := range []string{"*1\r\n$-1\r\n", "*0\r\n", "*-1\r\n", "*99999999\r\n", "*1\r\n$999999999\r\n", "*1\r\n$4\r\nPINGxx", "*1\r\n$4\r\nPIN"} {
		t.Run(request, func(t *testing.T) {
			r := NewReader(strings.NewReader(request))
			if err := r.ReadCommand(); err == nil {
				t.Fatal("accepted malformed request")
			}
		})
	}
}
func TestRESPOptionsAndQueueOwnership(t *testing.T) {
	r := NewReader(strings.NewReader("*3\r\n$3\r\nSET\r\n$1\r\nk\r\n$2\r\nEX\r\n*3\r\n$3\r\nSET\r\n$1\r\nk\r\n$2\r\nXX\r\n"))
	if err := r.ReadCommand(); err != nil {
		t.Fatal(err)
	}
	if r.cmd.Options.EX != 0 {
		t.Fatal("data parsed as option")
	}
	queued := r.cmd.Clone()
	if err := r.ReadCommand(); err != nil {
		t.Fatal(err)
	}
	if queued.Args[1] != "EX" || r.cmd.Args[1] != "XX" {
		t.Fatal("queued args overwritten")
	}
	inline := NewReader(strings.NewReader("SET k 'hello world' EX 5\r\n"))
	if err := inline.ReadCommand(); err != nil {
		t.Fatal(err)
	}
	if inline.cmd.Args[1] != "hello world" || inline.cmd.Options.EX != 3 {
		t.Fatal(inline.cmd)
	}
}
func TestConnectionRejectsMalformedWithoutPanic(t *testing.T) {
	a, b := net.Pipe()
	done := make(chan struct{})
	go func() { handleConn(a, func(c *Conn, cmd Command) { c.WriteString("PONG") }); close(done) }()
	defer b.Close()
	b.SetDeadline(time.Now().Add(time.Second))
	if _, err := b.Write([]byte("*1\r\n$-1\r\n")); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 256)
	n, err := b.Read(buf)
	if err != nil || !bytes.HasPrefix(buf[:n], []byte("-ERR")) {
		t.Fatal(string(buf[:n]), err)
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("connection not closed")
	}
}

type shortWriter struct{ bytes.Buffer }

func (w *shortWriter) Write(b []byte) (int, error) {
	if len(b) > 2 {
		b = b[:2]
	}
	return w.Buffer.Write(b)
}
func TestWriterHandlesShortWrites(t *testing.T) {
	dst := &shortWriter{}
	w := NewWriter(dst)
	w.WriteString("hello")
	if err := w.Push(); err != nil {
		t.Fatal(err)
	}
	if dst.String() != "+hello\r\n" {
		t.Fatal(dst.String())
	}
}
func FuzzRESP(f *testing.F) {
	for _, seed := range []string{"*1\r\n$4\r\nPING\r\n", "*1\r\n$-1\r\n", "SET k EX\r\n"} {
		f.Add([]byte(seed))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1<<20 {
			return
		}
		_ = NewReader(bytes.NewReader(data)).ReadCommand()
	})
}
func BenchmarkParseRESP(b *testing.B) {
	payload := []byte("*2\r\n$3\r\nGET\r\n$3\r\nhot\r\n")
	src := bytes.NewReader(payload)
	r := NewReader(src)
	r.ReadCommand()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		src.Reset(payload)
		if err := r.ReadCommand(); err != nil {
			b.Fatal(err)
		}
	}
}

var _ io.Writer = (*shortWriter)(nil)

func TestWriteArgumentsRemainOwned(t *testing.T) {
	r := NewReader(strings.NewReader("*3\r\n$3\r\nSET\r\n$3\r\nold\r\n$3\r\none\r\n*3\r\n$3\r\nSET\r\n$3\r\nnew\r\n$3\r\ntwo\r\n"))
	if err := r.ReadCommand(); err != nil {
		t.Fatal(err)
	}
	key, value := r.cmd.Args[0], r.cmd.Args[1]
	if err := r.ReadCommand(); err != nil {
		t.Fatal(err)
	}
	if key != "old" || value != "one" {
		t.Fatal("retained database strings were overwritten")
	}
}
