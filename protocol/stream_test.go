package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"testing"
	"time"
)

func TestCommandStreamsReuseAndZeroAllocations(t *testing.T) {
	command := Command{RequestID: 42, Op: OpUpdate, Payload: []byte("borrowed payload")}
	var output bytes.Buffer
	output.Grow(256)
	writer := CommandWriter[*bytes.Buffer]{Writer: &output}
	if allocs := testing.AllocsPerRun(1000, func() {
		output.Reset()
		if err := writer.Write(command); err != nil {
			panic(err)
		}
	}); allocs != 0 {
		t.Fatalf("command writer: %g allocs", allocs)
	}
	frame := append([]byte(nil), output.Bytes()...)
	if !bytes.Equal(frame, AppendCommand(nil, command)) {
		t.Fatal("stream format changed")
	}
	input := bytes.NewReader(frame)
	reader := CommandReader[*bytes.Reader]{Reader: input}
	if _, err := reader.Read(); err != nil {
		t.Fatal(err)
	}
	if allocs := testing.AllocsPerRun(1000, func() {
		input.Reset(frame)
		decoded, err := reader.Read()
		if err != nil || decoded.RequestID != command.RequestID || !bytes.Equal(decoded.Payload, command.Payload) {
			panic("frame changed")
		}
	}); allocs != 0 {
		t.Fatalf("command reader: %g allocs", allocs)
	}
}

func TestCommandStreamRejectsDamagedFrames(t *testing.T) {
	frame := AppendCommand(nil, Command{RequestID: 7, Op: OpCreate, Payload: []byte("value")})
	for length := 0; length < len(frame); length++ {
		reader := CommandReader[*bytes.Reader]{Reader: bytes.NewReader(frame[:length])}
		_, err := reader.Read()
		want := io.ErrUnexpectedEOF
		if length == 0 {
			want = io.EOF
		}
		if !errors.Is(err, want) {
			t.Fatalf("truncated at %d: %v", length, err)
		}
	}
	for _, damage := range []string{"magic", "length", "checksum", "oversize"} {
		t.Run(damage, func(t *testing.T) {
			data := append([]byte(nil), frame...)
			switch damage {
			case "magic":
				data[0] ^= 1
			case "length":
				binary.LittleEndian.PutUint32(data[4:8], 8)
			case "checksum":
				data[len(data)-1] ^= 1
			case "oversize":
				binary.LittleEndian.PutUint32(data[4:8], MaxFrameSize+1)
			}
			reader := CommandReader[*bytes.Reader]{Reader: bytes.NewReader(data)}
			if _, err := reader.Read(); !errors.Is(err, ErrInvalidCommand) {
				t.Fatal(err)
			}
			if damage != "checksum" && reader.buffer != nil {
				t.Fatal("invalid prefix allocated a frame")
			}
		})
	}
}

func TestCommandStreamPartialIOAndFrameBoundaries(t *testing.T) {
	var buffer bytes.Buffer
	writer := CommandWriter[fragmentWriter]{Writer: fragmentWriter{&buffer}}
	for _, id := range []uint64{7, 9} {
		if err := writer.Write(Command{RequestID: id, Op: OpGet, Payload: []byte("fragmented")}); err != nil {
			t.Fatal(err)
		}
	}
	reader := CommandReader[fragmentReader]{Reader: fragmentReader{&buffer}}
	for _, id := range []uint64{7, 9} {
		command, err := reader.Read()
		if err != nil || command.RequestID != id || string(command.Payload) != "fragmented" {
			t.Fatalf("frame %d: %+v, %v", id, command, err)
		}
	}
	if _, err := reader.Read(); err != io.EOF {
		t.Fatal(err)
	}
	stalled := CommandWriter[stalledWriter]{Writer: stalledWriter{}}
	if err := stalled.Write(Command{Op: OpGet}); err != io.ErrShortWrite {
		t.Fatal(err)
	}
}

type fragmentWriter struct{ *bytes.Buffer }

func (w fragmentWriter) Write(data []byte) (int, error) {
	return w.Buffer.Write(data[:min(3, len(data))])
}

type fragmentReader struct{ *bytes.Buffer }

func (r fragmentReader) Read(data []byte) (int, error) {
	return r.Buffer.Read(data[:min(2, len(data))])
}

type stalledWriter struct{}

func (stalledWriter) Write([]byte) (int, error) { return 0, nil }

func TestTCPCommandStreamsZeroAllocations(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	done := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			done <- err
			return
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
		reader := CommandReader[net.Conn]{Reader: conn}
		writer := CommandWriter[net.Conn]{Writer: conn}
		for {
			command, err := reader.Read()
			if err == io.EOF {
				done <- nil
				return
			}
			if err != nil {
				done <- err
				return
			}
			if err := writer.Write(command); err != nil {
				done <- err
				return
			}
		}
	}()
	conn, err := net.DialTimeout("tcp", listener.Addr().String(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	reader := CommandReader[net.Conn]{Reader: conn}
	writer := CommandWriter[net.Conn]{Writer: conn}
	command := Command{RequestID: 8, Op: OpGet, Payload: []byte("steady state")}
	if allocs := testing.AllocsPerRun(100, func() {
		if err := writer.Write(command); err != nil {
			panic(err)
		}
		response, err := reader.Read()
		if err != nil || !bytes.Equal(response.Payload, command.Payload) {
			panic("round trip failed")
		}
	}); allocs != 0 {
		t.Fatalf("TCP frame round trip: %g allocs", allocs)
	}
	_ = conn.Close()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}
