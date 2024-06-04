package fs

import (
	"sync"

	"github.com/tidwall/btree"
)

type Memory struct {
	sync.RWMutex
	files btree.Map[string, *MemoryFile]
}

func (m *Memory) OpenFile(filename string, flag int) (File, error) {
	m.Lock()
	defer m.Unlock()
	if f, ok := m.files.Get(filename); ok {
		return f, nil
	}
	mf := &MemoryFile{data: make([]byte, 0, 2048)}
	m.files.Set(filename, mf)
	return mf, nil
}

func (m *Memory) MkdirAll(path string) error {
	return nil
}

func (m *Memory) Rename(oldpath, newpath string) error {
	m.Lock()
	defer m.Unlock()
	if f, ok := m.files.Get(oldpath); ok {
		m.files.Set(newpath, f)
		m.files.Delete(oldpath)
	}
	return nil
}

func (m *Memory) IsDir(path string) (bool, error) {
	return true, nil
}

func (m *Memory) RemoveAll(path string) error {
	m.Lock()
	defer m.Unlock()
	m.files.Clear()
	return nil
}

func (m *Memory) Remove(filename string) error {
	m.Lock()
	defer m.Unlock()
	m.files.Delete(filename)
	return nil
}

type MemoryFile struct {
	data []byte
}

func (m *MemoryFile) ReadAt(b []byte, off int64) (n int, err error) {
	if off >= int64(len(m.data)) {
		return 0, nil
	}
	return copy(b, m.data[off:]), nil
}

func (m *MemoryFile) Write(b []byte) (n int, err error) {
	m.data = append(m.data, b...)
	return len(b), nil
}

func (m *MemoryFile) WriteAt(b []byte, off int64) (n int, err error) {
	if off < 0 {
		return 0, nil
	}
	dataLen := int64(len(m.data))
	writeLen := int64(len(b))
	if off >= dataLen {
		m.data = append(m.data, b...)
		return len(b), nil
	}
	if off+int64(len(b)) > int64(len(m.data)) {
		m.data = append(m.data, make([]byte, off+writeLen-dataLen)...)
	}
	return copy(m.data[off:], b), nil
}

func (m *MemoryFile) Close() error {
	return nil
}

func (m *MemoryFile) FileSize() (int64, error) {
	return int64(len(m.data)), nil
}

func (m *MemoryFile) Truncate(size int64) error {
	if size < 0 {
		return nil
	}
	if size == 0 {
		m.data = make([]byte, 0, 2048)
		return nil
	}
	dataLen := int64(len(m.data))
	if size < dataLen {
		m.data = m.data[:size]
		return nil
	}
	m.data = append(m.data, make([]byte, size-dataLen)...)
	return nil
}

func (m *MemoryFile) ReadAll() ([]byte, error) {
	return m.data, nil
}
