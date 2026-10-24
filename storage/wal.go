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
	ErrCorrupt         = errors.New("nodis storage: corrupt WAL")
	ErrRecordTooLarge  = errors.New("nodis storage: record too large")
	ErrClosed          = errors.New("nodis storage: log closed")
	ErrMemoryOnly      = errors.New("nodis storage: checkpoint requires a directory")
	ErrLocked          = errors.New("nodis storage: database directory is already locked")
	ErrLockUnsupported = errors.New("nodis storage: database locking is unsupported on this platform")
	magic              = [4]byte{'N', 'D', 'W', '1'}
	crcTable           = crc32.MakeTable(crc32.Castagnoli)
)

type Log struct {
	mu      sync.Mutex
	file    *os.File
	lock    *os.File
	dir     string
	closed  bool
	failure error
	writer  FrameWriter[*os.File]
}

func Open(dir string) (*Log, error) {
	if dir == "" {
		return &Log{}, nil
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, err
	}
	lock, err := lockDirectory(dir)
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(filepath.Join(dir, "nodis.wal"), os.O_CREATE|os.O_RDWR, 0o640)
	if err != nil {
		_ = lock.Close()
		return nil, err
	}
	// Persist the new WAL inode before its directory entry, including any newly
	// created ancestors. Do this on reopen too: a previous process may have died
	// between creation and directory sync without a machine crash.
	err = file.Sync()
	if err == nil {
		for path := dir; ; path = filepath.Dir(path) {
			if err = syncDirectory(path); err != nil || filepath.Dir(path) == path {
				break
			}
		}
	}
	if err == nil {
		_, err = file.Seek(0, io.SeekEnd)
	}
	if err != nil {
		_ = file.Close()
		_ = lock.Close()
		return nil, err
	}
	return &Log{file: file, lock: lock, dir: dir, writer: FrameWriter[*os.File]{Writer: file}}, nil
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
		l.failure = err
		return l.failure
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
				l.failure = err
				return l.failure
			}
			if err := l.file.Sync(); err != nil {
				l.failure = err
				return l.failure
			}
			break
		}
		if err != nil {
			l.failure = err
			return l.failure
		}
		if sequence == 0 || sequence <= previous {
			l.failure = ErrCorrupt
			return l.failure
		}
		previous = sequence
		if err := apply(sequence, payload); err != nil {
			l.failure = err
			return l.failure
		}
	}
	_, l.failure = l.file.Seek(0, io.SeekEnd)
	return l.failure
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
	if len(payload) > MaxRecordSize {
		return ErrRecordTooLarge
	}
	// Never truncate after an I/O failure, especially a failed fsync: whether
	// this record reached stable storage is unknown. Preserve recovery evidence
	// and reject all further operations until the database has been reopened.
	l.failure = appendRecord(&l.writer, sequence, payload, durable)
	return l.failure
}

type walWriter interface {
	io.Writer
	Sync() error
}

func appendRecord[W walWriter](writer *FrameWriter[W], sequence uint64, payload []byte, durable bool) error {
	if err := writer.Append(sequence, payload); err != nil {
		return err
	}
	if durable {
		return writer.Writer.Sync()
	}
	return nil
}

func syncDirectory(path string) error {
	directory, err := os.Open(path)
	if err != nil {
		return err
	}
	return syncAndClose(directory)
}

type syncCloser interface {
	Sync() error
	Close() error
}

func syncAndClose[F syncCloser](file F) error {
	syncErr := file.Sync()
	closeErr := file.Close()
	if syncErr != nil && closeErr != nil {
		return errors.Join(syncErr, closeErr)
	}
	if syncErr != nil {
		return syncErr
	}
	return closeErr
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
	l.failure = l.file.Sync()
	return l.failure
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
	err := syncAndClose(l.file)
	l.file = nil
	if l.lock != nil {
		closeErr := l.lock.Close()
		l.lock = nil
		if closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}
	if l.failure != nil {
		return errors.Join(l.failure, err)
	}
	return err
}
