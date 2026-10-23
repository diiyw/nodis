package storage

import (
	"encoding/binary"
	"hash/crc32"
	"io"
)

// FrameWriter reuses its header and checksum storage. A writer with sufficient
// capacity (or an open file) can encode records without per-record allocations.
type FrameWriter[W io.Writer] struct {
	Writer W
	header [16]byte
	tail   [4]byte
}

func (w *FrameWriter[W]) Append(sequence uint64, payload []byte) error {
	if len(payload) > MaxRecordSize {
		return ErrRecordTooLarge
	}
	copy(w.header[:4], magic[:])
	binary.LittleEndian.PutUint32(w.header[4:8], uint32(len(payload)))
	binary.LittleEndian.PutUint64(w.header[8:], sequence)
	checksum := crc32.Update(0, crcTable, w.header[:])
	checksum = crc32.Update(checksum, crcTable, payload)
	binary.LittleEndian.PutUint32(w.tail[:], checksum)
	if err := writeAll(w.Writer, w.header[:]); err != nil {
		return err
	}
	if err := writeAll(w.Writer, payload); err != nil {
		return err
	}
	return writeAll(w.Writer, w.tail[:])
}

// FrameReader returns borrowed payloads valid until the next Read. Incomplete
// records return io.ErrUnexpectedEOF; a clean record boundary returns io.EOF.
type FrameReader[R io.Reader] struct {
	Reader R
	header [16]byte
	buffer []byte
	Offset int64
}

func (r *FrameReader[R]) Read() (uint64, []byte, error) {
	n, err := io.ReadFull(r.Reader, r.header[:])
	if err != nil {
		if n != 0 && err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return 0, nil, err
	}
	if string(r.header[:4]) != string(magic[:]) {
		return 0, nil, ErrCorrupt
	}
	size := int(binary.LittleEndian.Uint32(r.header[4:8]))
	if size > MaxRecordSize {
		return 0, nil, ErrCorrupt
	}
	if cap(r.buffer) < size+4 {
		r.buffer = make([]byte, size+4)
	}
	r.buffer = r.buffer[:size+4]
	if _, err := io.ReadFull(r.Reader, r.buffer); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return 0, nil, err
	}
	checksum := crc32.Update(0, crcTable, r.header[:])
	checksum = crc32.Update(checksum, crcTable, r.buffer[:size])
	if checksum != binary.LittleEndian.Uint32(r.buffer[size:]) {
		return 0, nil, ErrCorrupt
	}
	r.Offset += int64(20 + size)
	return binary.LittleEndian.Uint64(r.header[8:]), r.buffer[:size], nil
}
