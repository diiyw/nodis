package nodis

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"encoding/binary"

	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/ds/hash"
	"github.com/diiyw/nodis/ds/list"
	"github.com/diiyw/nodis/ds/set"
	"github.com/diiyw/nodis/ds/str"
	"github.com/diiyw/nodis/ds/zset"
	"github.com/diiyw/nodis/fs"
	"github.com/diiyw/nodis/redis"
	"github.com/tidwall/btree"
)

type store struct {
	mu          sync.RWMutex
	metadata    btree.Map[string, *metadata]
	fileSize    int64
	fileId      uint16
	path        string
	current     string
	indexFile   string
	aof         fs.File
	filesystem  fs.Fs
	closed      bool
	watchMu     sync.RWMutex
	watchedKeys btree.Map[string, *list.LinkedListG[*redis.Conn]]
}

func newStore(path string, fileSize int64, filesystem fs.Fs) *store {
	s := &store{
		path:      path,
		fileSize:  fileSize,
		indexFile: filepath.Join(path, "nodis.index"),
	}
	_ = filesystem.MkdirAll(path)
	indexFile, err := filesystem.OpenFile(s.indexFile, os.O_RDWR|os.O_CREATE|os.O_APPEND)
	if err != nil {
		panic(err)
	}
	data, err := indexFile.ReadAll()
	if err != nil {
		panic(err)
	}
	err = indexFile.Close()
	if err != nil {
		panic(err)
	}
	if len(data) > 2 {
		s.fileId = binary.LittleEndian.Uint16(data[:2])
		if len(data[2:]) > 0 {
			data = data[2:]
			for {
				if len(data) == 0 {
					break
				}
				keyLen := int(data[0])
				key := string(data[1 : 1+keyLen])
				data = data[1+keyLen:]
				if len(data) == 0 {
					break
				}
				var m = newMetadata(&Key{}, nil, false)
				m.unmarshal(data)
				data = data[metadataSize:]
				m.state |= KeyStateNormal
				s.metadata.Set(key, m)
			}
		}
	}
	s.filesystem = filesystem
	s.current = filepath.Join(path, "nodis."+strconv.Itoa(int(s.fileId))+".aof")
	s.aof, err = s.filesystem.OpenFile(s.current, os.O_RDWR|os.O_CREATE|os.O_APPEND)
	if err != nil {
		panic(err)
	}
	return s
}

func (s *store) fromStorage(m *metadata) *metadata {
	// try get from storage
	v, err := s.getValueEntryRaw(m.key)
	if err == nil && len(v) > 0 {
		value, err := s.parseValue(v)
		if err != nil {
			log.Println("Parse Value:", err)
			return m
		}
		if value != nil {
			m.setValue(value)
			return m
		}
	}
	return m
}

func (s *store) parseEntryBytes(data []byte) (*ValueEntry, error) {
	var entry = &ValueEntry{}
	if err := entry.decode(data); err != nil {
		return nil, err
	}
	return entry, nil
}

// parseValue the data
func (s *store) parseValue(data []byte) (ds.Value, error) {
	var entry, err = s.parseEntryBytes(data)
	if err != nil {
		return nil, err
	}
	var value ds.Value
	switch ds.ValueType(entry.Type) {
	case ds.String:
		v := str.NewString()
		v.SetValue(entry.Value)
		value = v
	case ds.ZSet:
		z := zset.NewSortedSet()
		z.SetValue(entry.Value)
		value = z
	case ds.List:
		l := list.NewLinkedList()
		l.SetValue(entry.Value)
		value = l
	case ds.Hash:
		h := hash.NewHashMap()
		h.SetValue(entry.Value)
		value = h
	case ds.Set:
		v := set.NewSet()
		v.SetValue(entry.Value)
		value = v
	default:
		panic("unhandled default case")
	}
	return value, nil
}

// saveData flush changed keys to disk
func (s *store) saveData() {
	now := time.Now().UnixMilli()
	s.metadata.Scan(func(key string, m *metadata) bool {
		m.Lock()
		defer m.Unlock()
		if !m.modified() || m.expired(now) || !m.isOk() {
			return true
		}
		if m.value == nil {
			return true
		}
		// saveData to storage
		err := s.saveMetadata(key, m)
		if err != nil {
			log.Println("Flush changes: ", err)
		}
		return true
	})
}

// saveKeyIndex save the key index to disk
func (s *store) saveKeyIndex() error {
	idxFile, err := s.filesystem.OpenFile(s.indexFile+"~", os.O_RDWR|os.O_CREATE|os.O_APPEND)
	if err != nil {
		return err
	}
	var header = make([]byte, 2)
	binary.LittleEndian.PutUint16(header, s.fileId)
	_, err = idxFile.Write(header)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	now := time.Now().UnixMilli()
	s.metadata.Copy().Scan(func(key string, m *metadata) bool {
		m.RLock()
		defer m.RUnlock()
		if m.expired(now) {
			return true
		}
		buf.WriteByte(byte(len(key)))
		buf.WriteString(key)
		buf.Write(m.marshal())
		return true
	})
	_, err = buf.WriteTo(idxFile)
	if err != nil {
		return err
	}
	if err = idxFile.Close(); err != nil {
		log.Println("Close sync error: ", err)
	}
	err = s.filesystem.Rename(s.indexFile+"~", s.indexFile)
	if err != nil {
		return err
	}
	return nil
}

// tidy removes expired and unused keys
func (s *store) tidy(keyMaxUseTimes uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	now := time.Now().UnixMilli()
	err := s.saveKeyIndex()
	if err != nil {
		log.Println("Tidy: ", err)
	}
	s.metadata.Scan(func(key string, m *metadata) bool {
		m.Lock()
		defer m.Unlock()
		if m.expired(now) || !m.isOk() {
			s.metadata.Delete(key)
			return true
		}
		if m.useTimes < keyMaxUseTimes {
			if m.modified() {
				// saveData to disk
				err := s.saveMetadata(key, m)
				if err != nil {
					log.Println("Tidy: ", err)
				}
			}
			m.reset()
		}
		return true
	})
}

func (s *store) check() (int64, error) {
	var offset, err = s.aof.FileSize()
	if err != nil {
		return 0, err
	}
	if offset >= s.fileSize {
		err = s.aof.Close()
		if err != nil {
			return 0, err
		}
		// open file with new file id
		s.fileId++
		s.current = filepath.Join(s.path, "nodis."+strconv.Itoa(int(s.fileId))+".aof")
		s.aof, err = s.filesystem.OpenFile(s.current, os.O_RDWR|os.O_CREATE|os.O_APPEND)
		if err != nil {
			return 0, err
		}
		// update index file
		idxFi, err := s.filesystem.OpenFile(s.indexFile, os.O_CREATE|os.O_RDWR)
		if err != nil {
			return 0, err
		}
		defer func() {
			err := idxFi.Close()
			if err != nil {
				log.Println("Close index file error: ", err)
			}
		}()
		var header = make([]byte, 4)
		binary.LittleEndian.PutUint16(header, s.fileId)
		_, err = idxFi.WriteAt(header, 0)
		if err != nil {
			return 0, err
		}
		offset = 0
	}
	return offset, nil
}

func (s *store) saveMetadata(name string, m *metadata) error {
	offset, err := s.check()
	if err != nil {
		return err
	}
	m.key.fileId = s.fileId
	m.key.offset = offset
	v := newValueEntry(name, m.value, m.expiration)
	data := v.encode()
	m.key.size = uint32(len(data))
	_, err = s.aof.Write(data)
	if err != nil {
		return err
	}
	return nil
}

// saveValueEntry a key-value pair into store
func (s *store) saveValueEntry(entry *ValueEntry) error {
	var m = newMetadata(&Key{}, nil, false)
	offset, err := s.check()
	if err != nil {
		return err
	}
	data := entry.encode()
	m.key.fileId = s.fileId
	m.key.offset = offset
	m.key.size = uint32(len(data))
	m.expiration = entry.Expiration
	m.state |= KeyStateNormal
	s.metadata.Set(entry.Key, m)
	_, err = s.aof.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (s *store) saveValueRaw(name string, m *metadata, data []byte) error {
	offset, err := s.check()
	if err != nil {
		return err
	}
	m.key.fileId = s.fileId
	m.key.offset = offset
	m.key.size = uint32(len(data))
	m.state |= KeyStateNormal
	s.metadata.Set(name, m)
	_, err = s.aof.Write(data)
	if err != nil {
		return err
	}
	return nil
}

// getValueEntryRaw get entry raw data
func (s *store) getValueEntryRaw(key *Key) ([]byte, error) {
	if key.fileId == s.fileId {
		data := make([]byte, key.size)
		_, err := s.aof.ReadAt(data, key.offset)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
	// read from other file
	file := filepath.Join(s.path, "nodis."+strconv.Itoa(int(key.fileId))+".aof")
	f, err := s.filesystem.OpenFile(file, os.O_RDONLY)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = f.Close()
		if err != nil {
			log.Println("Close file error: ", err)
		}
	}()
	data := make([]byte, key.size)
	_, err = f.ReadAt(data, key.offset)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// snapshot the store
func (s *store) snapshot(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	snapshotDir := filepath.Join(path, "snapshots", time.Now().Format("20060102_150405"))
	err := s.filesystem.MkdirAll(snapshotDir)
	if err != nil {
		log.Println("Snapshot mkdir error: ", err)
		return
	}
	s.saveData()
	ns := newStore(snapshotDir, s.fileSize, s.filesystem)
	s.metadata.Copy().Scan(func(key string, meta *metadata) bool {
		data, err := s.getValueEntryRaw(meta.key)
		if err != nil {
			log.Println("Snapshot get error: ", err)
			return true
		}
		err = ns.saveValueRaw(key, meta, data)
		if err != nil {
			log.Println("Snapshot put error: ", err)
			return true
		}
		return true
	})
	err = ns.close()
	if err != nil {
		log.Println("Snapshot saveData error: ", err)
	}
}

// close the store
func (s *store) close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	s.saveData()
	err := s.aof.Close()
	if err != nil {
		return err
	}
	return s.saveKeyIndex()
}

// clear the store
func (s *store) clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.metadata.Clear()
	err := s.aof.Truncate(0)
	if err != nil {
		return err
	}
	err = s.filesystem.Remove(s.indexFile)
	if err != nil {
		return err
	}
	return nil
}
