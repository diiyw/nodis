package redis

import (
	"strings"
	"testing"
)

func TestReadInlineSimple(t *testing.T) {
	doc := "ping\r\n"
	r := NewReader(strings.NewReader(doc))
	err := r.ReadCommand()
	if err != nil {
		t.Error(err)
	}
	if r.cmd.Name != "PING" {
		t.Errorf("not equal %s", r.cmd.Name)
	}
}

func TestReadInlineMulti(t *testing.T) {
	doc := "ping\r\nset foo bar\r\n"
	r := NewReader(strings.NewReader(doc))
	err := r.ReadCommand()
	if err != nil {
		t.Error(err)
	}
	if r.cmd.Name != "PING" {
		t.Errorf("not equal %s", r.cmd.Name)
	}
	err = r.ReadCommand()
	if err != nil {
		t.Error(err)
	}
	if r.cmd.Name != "SET" {
		t.Errorf("not equal %s", r.cmd.Name)
	}
	if r.cmd.Args[0] != "foo" {
		t.Errorf("not equal %s", r.cmd.Args[0])
	}
	if r.cmd.Args[1] != "bar" {
		t.Errorf("not equal '%s'", r.cmd.Args[1])
	}
}

func TestReadInlineQuotes(t *testing.T) {
	doc := "set \"foo\" bar\r\nset foo \"bar\"\r\n\r\nset foo \"bar bar2\"\r\n"
	r := NewReader(strings.NewReader(doc))
	err := r.ReadCommand()
	if err != nil {
		t.Error(err)
	}
	if r.cmd.Name != "SET" {
		t.Errorf("not equal %s", r.cmd.Name)
	}
	if r.cmd.Args[0] != "foo" {
		t.Errorf("not equal %s", r.cmd.Args[0])
	}
	if r.cmd.Args[1] != "bar" {
		t.Errorf("not equal '%s'", r.cmd.Args[1])
	}
	err = r.ReadCommand()
	if err != nil {
		t.Error(err)
	}
	if r.cmd.Name != "SET" {
		t.Errorf("not equal %s", r.cmd.Name)
	}
	if r.cmd.Args[0] != "foo" {
		t.Errorf("not equal %s", r.cmd.Args[0])
	}
	if r.cmd.Args[1] != "bar" {
		t.Errorf("not equal '%s'", r.cmd.Args[1])
	}
}

func TestReadInlineSpace(t *testing.T) {
	doc := "set foo \"bar bar2\"\r\nset  foo \"bar\\\"bar2\"\r\n"
	r := NewReader(strings.NewReader(doc))
	err := r.ReadCommand()
	if err != nil {
		t.Error(err)
	}
	if r.cmd.Name != "SET" {
		t.Errorf("not equal %s", r.cmd.Name)
	}
	if r.cmd.Args[0] != "foo" {
		t.Errorf("not equal %s", r.cmd.Args[0])
	}
	if r.cmd.Args[1] != "bar bar2" {
		t.Errorf("not equal '%s'", r.cmd.Args[1])
	}
	err = r.ReadCommand()
	if err != nil {
		t.Error(err)
	}
	if r.cmd.Name != "SET" {
		t.Errorf("not equal %s", r.cmd.Name)
	}
	if r.cmd.Args[0] != "foo" {
		t.Errorf("not equal %s", r.cmd.Args[0])
	}
	if r.cmd.Args[1] != "bar\\\"bar2" {
		t.Errorf("not equal '%s'", r.cmd.Args[1])
	}
}

func TestInlineQuotes(t *testing.T) {
	doc := "set  foo \"bar\\\"\r\nbar3\"\r\n"
	r := NewReader(strings.NewReader(doc))
	err := r.ReadCommand()
	if err != nil {
		t.Error(err)
	}
	if r.cmd.Name != "SET" {
		t.Errorf("not equal %s", r.cmd.Name)
	}
	if r.cmd.Args[0] != "foo" {
		t.Errorf("not equal %s", r.cmd.Args[0])
	}
	if r.cmd.Args[1] != "bar\\\"\r\nbar3" {
		t.Errorf("not equal '%s'", r.cmd.Args[1])
	}
}

func TestReadCommand(t *testing.T) {
	tests := []struct {
		doc      string
		expected Command
	}{
		{
			doc: "ping\r\n",
			expected: Command{
				Name: "PING",
				Args: []string{},
			},
		},
		{
			doc: "set foo bar\r\n",
			expected: Command{
				Name: "SET",
				Args: []string{"foo", "bar"},
			},
		},
		{
			doc: "set \"foo\" \"bar\"\r\n",
			expected: Command{
				Name: "SET",
				Args: []string{"foo", "bar"},
			},
		},
		{
			doc: "set foo \"bar bar2\"\r\n",
			expected: Command{
				Name: "SET",
				Args: []string{"foo", "bar bar2"},
			},
		},
		{
			doc: "set foo \"bar\\\"bar3\"\r\n",
			expected: Command{
				Name: "SET",
				Args: []string{"foo", "bar\\\"bar3"},
			},
		},
	}

	for _, tt := range tests {
		r := NewReader(strings.NewReader(tt.doc))
		err := r.ReadCommand()
		if err != nil {
			t.Error(err)
		}
		if r.cmd.Name != tt.expected.Name {
			t.Errorf("expected %s, got %s", tt.expected.Name, r.cmd.Name)
		}
		for i, arg := range tt.expected.Args {
			if r.cmd.Args[i] != arg {
				t.Errorf("expected %s, got %s", arg, r.cmd.Args[i])
			}
		}
	}
}

func TestWriterWriteByte(t *testing.T) {
	w := NewWriter(&strings.Builder{})
	w.writeByte('a')
	if string(w.Bytes()) != "a" {
		t.Errorf("expected 'a', got %s", string(w.Bytes()))
	}
}

func TestWriterWriteBytes(t *testing.T) {
	w := NewWriter(&strings.Builder{})
	w.writeBytes('a', 'b', 'c')
	if string(w.Bytes()) != "abc" {
		t.Errorf("expected 'abc', got %s", string(w.Bytes()))
	}
}

func TestWriterWriteString(t *testing.T) {
	w := NewWriter(&strings.Builder{})
	w.WriteString("hello")
	expected := string(StringType) + "hello\r\n"
	if string(w.Bytes()) != expected {
		t.Errorf("expected %s, got %s", expected, string(w.Bytes()))
	}
}

func TestWriterWriteBulk(t *testing.T) {
	w := NewWriter(&strings.Builder{})
	w.WriteBulk("bulk")
	expected := string(BulkType) + "4\r\nbulk\r\n"
	if string(w.Bytes()) != expected {
		t.Errorf("expected %s, got %s", expected, string(w.Bytes()))
	}
}

func TestWriterWriteArray(t *testing.T) {
	w := NewWriter(&strings.Builder{})
	w.WriteArray(3)
	expected := string(ArrayType) + "3\r\n"
	if string(w.Bytes()) != expected {
		t.Errorf("expected %s, got %s", expected, string(w.Bytes()))
	}
}

func TestWriterWriteError(t *testing.T) {
	w := NewWriter(&strings.Builder{})
	w.WriteError("error")
	expected := string(ErrType) + "error\r\n"
	if string(w.Bytes()) != expected {
		t.Errorf("expected %s, got %s", expected, string(w.Bytes()))
	}
	if !w.HasError() {
		t.Errorf("expected HasError to be true")
	}
}

func TestWriterWriteBulkNull(t *testing.T) {
	w := NewWriter(&strings.Builder{})
	w.WriteBulkNull()
	expected := "$-1\r\n"
	if string(w.Bytes()) != expected {
		t.Errorf("expected %s, got %s", expected, string(w.Bytes()))
	}
}

func TestWriterWriteArrayNull(t *testing.T) {
	w := NewWriter(&strings.Builder{})
	w.WriteArrayNull()
	expected := "*-1\r\n"
	if string(w.Bytes()) != expected {
		t.Errorf("expected %s, got %s", expected, string(w.Bytes()))
	}
}

func TestWriterWriteInt64(t *testing.T) {
	w := NewWriter(&strings.Builder{})
	w.WriteInt64(123)
	expected := string(IntegerType) + "123\r\n"
	if string(w.Bytes()) != expected {
		t.Errorf("expected %s, got %s", expected, string(w.Bytes()))
	}
}

func TestWriterWriteUInt64(t *testing.T) {
	w := NewWriter(&strings.Builder{})
	w.WriteUInt64(123)
	expected := string(IntegerType) + "123\r\n"
	if string(w.Bytes()) != expected {
		t.Errorf("expected %s, got %s", expected, string(w.Bytes()))
	}
}

func TestWriterWriteDouble(t *testing.T) {
	w := NewWriter(&strings.Builder{})
	w.WriteDouble(123.456)
	expected := string(DoubleType) + "123.456\r\n"
	if string(w.Bytes()) != expected {
		t.Errorf("expected %s, got %s", expected, string(w.Bytes()))
	}
}

func TestWriterWriteMap(t *testing.T) {
	w := NewWriter(&strings.Builder{})
	w.WriteMap(2)
	expected := string(MapType) + "2\r\n"
	if string(w.Bytes()) != expected {
		t.Errorf("expected %s, got %s", expected, string(w.Bytes()))
	}
}

func TestWriterWriteNullMap(t *testing.T) {
	w := NewWriter(&strings.Builder{})
	w.WriteNullMap()
	expected := "%-1\r\n"
	if string(w.Bytes()) != expected {
		t.Errorf("expected %s, got %s", expected, string(w.Bytes()))
	}
}

func TestWriterWriteOK(t *testing.T) {
	w := NewWriter(&strings.Builder{})
	w.WriteOK()
	expected := "+OK\r\n"
	if string(w.Bytes()) != expected {
		t.Errorf("expected %s, got %s", expected, string(w.Bytes()))
	}
}

func TestWriterFlush(t *testing.T) {
	builder := &strings.Builder{}
	w := NewWriter(builder)
	w.WriteString("hello")
	err := w.Flush()
	if err != nil {
		t.Error(err)
	}
	expected := string(StringType) + "hello\r\n"
	if builder.String() != expected {
		t.Errorf("expected %s, got %s", expected, builder.String())
	}
	if w.w != 0 {
		t.Errorf("expected w to be 0, got %d", w.w)
	}
	if w.HasError() {
		t.Errorf("expected HasError to be false")
	}
}
