package storage

import (
	"errors"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"sync"
)

const MaxRecordSize = 64 << 20

var (
	ErrCorrupt        = errors.New("nodis storage: corrupt WAL")
	ErrRecordTooLarge = errors.New("nodis storage: record too large")
	ErrClosed         = errors.New("nodis storage: log closed")
	ErrMemoryOnly     = errors.New("nodis storage: checkpoint requires a directory")
	magic             = [4]byte{'N', 'D', 'W', '1'}
	crcTable          = crc32.MakeTable(crc32.Castagnoli)
)

type Log struct {
	mu      sync.Mutex
	file    *os.File
	dir     string
	closed  bool
	failure error
	writer  FrameWriter[*os.File]
}

func Open(dir string) (*Log, error) {
	if dir == "" {
		return &Log{}, nil
	}
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(filepath.Join(dir, "nodis.wal"), os.O_CREATE|os.O_RDWR, 0o640)
	if err != nil {
		return nil, err
	}
	return &Log{file: file, dir: dir, writer: FrameWriter[*os.File]{Writer: file}}, nil
}

func (l *Log) Replay(apply func(sequence uint64, payload []byte) error) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return ErrClosed
	}
	if l.failure != nil {
		return l.failure
	}
	if l.file == nil {
		return nil
	}
	if _, err := l.file.Seek(0, io.SeekStart); err != nil {
		return err
	}
	reader := FrameReader[*os.File]{Reader: l.file}
	var previous uint64
	for {
		sequence, payload, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err == io.ErrUnexpectedEOF {
			if err := l.file.Truncate(reader.Offset); err != nil {
				return err
			}
			break
		}
		if err != nil {
			return err
		}
		if sequence == 0 || sequence <= previous {
			return ErrCorrupt
		}
		previous = sequence
		if err := apply(sequence, payload); err != nil {
			return err
		}
	}
	_, err := l.file.Seek(0, io.SeekEnd)
	return err
}

func (l *Log) Append(sequence uint64, payload []byte, durable bool) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return ErrClosed
	}
	if l.failure != nil {
		return l.failure
	}
	if l.file == nil {
		return nil
	}
	start, err := l.file.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	rollback := func(writeErr error) error {
		if err := l.file.Truncate(start); err != nil {
			l.failure = errors.Join(writeErr, err)
			return l.failure
		}
		if _, err := l.file.Seek(start, io.SeekStart); err != nil {
			l.failure = errors.Join(writeErr, err)
			return l.failure
		}
		return writeErr
	}
	if err := l.writer.Append(sequence, payload); err != nil {
		return rollback(err)
	}
	if durable {
		if err := l.file.Sync(); err != nil {
			return rollback(err)
		}
	}
	return nil
}

func writeAll[W io.Writer](writer W, data []byte) error {
	for len(data) > 0 {
		n, err := writer.Write(data)
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

func (l *Log) Sync() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return ErrClosed
	}
	if l.failure != nil {
		return l.failure
	}
	if l.file == nil {
		return nil
	}
	return l.file.Sync()
}

func (l *Log) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return nil
	}
	l.closed = true
	if l.file == nil {
		return nil
	}
	err := l.file.Close()
	if err == nil {
		l.file = nil
	}
	return err
}
