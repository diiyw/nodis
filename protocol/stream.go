package protocol

import (
	"encoding/binary"
	"hash/crc32"
	"io"
)

// CommandReader reads NDC1 frames without allocating once its buffer has enough
// capacity. Payload is borrowed until the next Read. Use one reader per stream.
type CommandReader[R io.Reader] struct {
	Reader R
	prefix [8]byte
	buffer []byte
}

func (r *CommandReader[R]) Read() (Command, error) {
	if _, err := io.ReadFull(r.Reader, r.prefix[:]); err != nil {
		return Command{}, err
	}
	size := binary.LittleEndian.Uint32(r.prefix[4:])
	if string(r.prefix[:4]) != string(commandMagic[:]) || size < CommandHeaderSize+4 || size > MaxFrameSize {
		return Command{}, ErrInvalidCommand
	}
	if cap(r.buffer) < int(size) {
		r.buffer = make([]byte, int(size))
	}
	frame := r.buffer[:int(size)]
	copy(frame, r.prefix[:])
	if _, err := io.ReadFull(r.Reader, frame[8:]); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return Command{}, err
	}
	return DecodeCommand(frame)
}

// CommandWriter writes a header, borrowed payload, and checksum directly to the
// stream. Its scratch arrays are reused; Write must not be called concurrently.
type CommandWriter[W io.Writer] struct {
	Writer W
	header [CommandHeaderSize]byte
	tail   [4]byte
}

func (w *CommandWriter[W]) Write(command Command) error {
	if len(command.Payload) > MaxFrameSize-CommandHeaderSize-4 {
		return ErrInvalidCommand
	}
	copy(w.header[:4], commandMagic[:])
	binary.LittleEndian.PutUint32(w.header[4:8], uint32(CommandHeaderSize+len(command.Payload)+4))
	binary.LittleEndian.PutUint64(w.header[8:16], command.RequestID)
	w.header[16], w.header[17] = byte(command.Op), byte(command.Status)
	binary.LittleEndian.PutUint32(w.header[18:22], uint32(len(command.Payload)))
	checksum := crc32.Update(0, table, w.header[:])
	checksum = crc32.Update(checksum, table, command.Payload)
	binary.LittleEndian.PutUint32(w.tail[:], checksum)
	if err := writeFull(w.Writer, w.header[:]); err != nil {
		return err
	}
	if err := writeFull(w.Writer, command.Payload); err != nil {
		return err
	}
	return writeFull(w.Writer, w.tail[:])
}

func writeFull[W io.Writer](writer W, data []byte) error {
	for len(data) > 0 {
		n, err := writer.Write(data)
		if n < 0 || n > len(data) {
			return io.ErrShortWrite
		}
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
		data = data[n:]
	}
	return nil
}
