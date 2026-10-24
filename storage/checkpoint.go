package storage

import (
	"encoding/binary"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
)

const snapshotName = "nodis.snapshot"

// SnapshotWriter streams bounded, independently checksummed records. The final
// empty frame is mandatory; incomplete snapshots are never accepted on recovery.
type SnapshotWriter struct {
	frames   FrameWriter[*os.File]
	sequence uint64
	count    uint64
	checksum uint32
}

func (w *SnapshotWriter) Append(payload []byte) error {
	if len(payload) == 0 {
		return ErrCorrupt
	}
	if err := w.frames.Append(w.sequence, payload); err != nil {
		return err
	}
	w.count++
	w.checksum = crc32.Update(w.checksum, crcTable, payload)
	return nil
}

type SnapshotSource interface{ WriteSnapshot(*SnapshotWriter) error }

// Checkpoint publishes a durable snapshot before recycling the WAL. The caller
// must hold its commit lock throughout to prevent mutations between the two.
func Checkpoint[S SnapshotSource](l *Log, sequence uint64, source S) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return ErrClosed
	}
	if l.failure != nil {
		return l.failure
	}
	if l.file == nil {
		return ErrMemoryOnly
	}
	temporary, err := os.CreateTemp(l.dir, ".snapshot-*")
	if err != nil {
		return err
	}
	defer func() { _ = temporary.Close(); _ = os.Remove(temporary.Name()) }()
	var header [16]byte
	copy(header[:8], "NDS2\x00\x00\x00\x00")
	binary.LittleEndian.PutUint64(header[8:], sequence)
	if err := writeAll(temporary, header[:]); err != nil {
		return err
	}
	writer := SnapshotWriter{frames: FrameWriter[*os.File]{Writer: temporary}, sequence: sequence}
	if err := source.WriteSnapshot(&writer); err != nil {
		return err
	}
	if err := writer.frames.Append(sequence, nil); err != nil {
		return err
	}
	var footer [16]byte
	copy(footer[:4], "END2")
	binary.LittleEndian.PutUint64(footer[4:12], writer.count)
	binary.LittleEndian.PutUint32(footer[12:16], writer.checksum)
	if err := writeAll(temporary, footer[:]); err != nil {
		return err
	}
	if err := temporary.Sync(); err != nil {
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporary.Name(), filepath.Join(l.dir, snapshotName)); err != nil {
		return err
	}
	// A crash before this directory sync may leave the old snapshot, so retain
	// the WAL until both the file and its directory entry are durable.
	if err := syncDirectory(l.dir); err != nil {
		l.failure = err
		return l.failure
	}
	if err := l.file.Truncate(0); err != nil {
		l.failure = err
		return err
	}
	if _, err := l.file.Seek(0, io.SeekStart); err != nil {
		l.failure = err
		return err
	}
	l.failure = l.file.Sync()
	return l.failure
}

// RestoreSnapshot strictly validates the published file. Missing snapshots are
// allowed for databases created before checkpointing was enabled.
func (l *Log) RestoreSnapshot(apply func([]byte) error) (uint64, bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return 0, false, ErrClosed
	}
	if l.file == nil {
		return 0, false, nil
	}
	file, err := os.Open(filepath.Join(l.dir, snapshotName))
	if os.IsNotExist(err) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	defer file.Close()
	var header [16]byte
	if _, err := io.ReadFull(file, header[:]); err != nil {
		return 0, false, ErrCorrupt
	}
	if string(header[:8]) != "NDS2\x00\x00\x00\x00" {
		return 0, false, ErrCorrupt
	}
	sequence := binary.LittleEndian.Uint64(header[8:])
	reader := FrameReader[*os.File]{Reader: file}
	var count uint64
	var checksum uint32
	for {
		recordSequence, payload, err := reader.Read()
		if err != nil {
			return 0, false, ErrCorrupt
		}
		if recordSequence != sequence {
			return 0, false, ErrCorrupt
		}
		if len(payload) == 0 {
			var footer [16]byte
			if _, err := io.ReadFull(file, footer[:]); err != nil {
				return 0, false, ErrCorrupt
			}
			if string(footer[:4]) != "END2" || binary.LittleEndian.Uint64(footer[4:12]) != count || binary.LittleEndian.Uint32(footer[12:16]) != checksum {
				return 0, false, ErrCorrupt
			}
			var tail [1]byte
			if n, err := file.Read(tail[:]); n != 0 || err != io.EOF {
				return 0, false, ErrCorrupt
			}
			return sequence, true, nil
		}
		count++
		checksum = crc32.Update(checksum, crcTable, payload)
		if err := apply(payload); err != nil {
			return 0, false, err
		}
	}
}
