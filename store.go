package nodis

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/dolthub/swiss"
	"github.com/kelindar/binary"
)

type store struct {
	sync.RWMutex
	fileSize  int64
	aof       *os.File
	path      string
	current   string
	index     *swiss.Map[string, *index]
	fileId    uint32
	indexFile string
}

func newStore(path string, fileSize int64) *store {
	s := &store{
		path:      path,
		fileSize:  fileSize,
		index:     swiss.NewMap[string, *index](32),
		indexFile: filepath.Join(path, "nodis.index"),
	}
	_ = os.MkdirAll(path, 0755)
	data, _ := os.ReadFile(s.indexFile)
	if len(data) >= 4 {
		if len(data[4:]) > 0 {
			var m map[string]*index
			err := binary.Unmarshal(data[4:], &m)
			if err != nil {
				panic(err)
			}
			for k, v := range m {
				s.index.Put(k, v)
			}
		}
		s.fileId = binary.LittleEndian.Uint32(data[:4])
	}
	s.current = filepath.Join(path, "nodis."+strconv.Itoa(int(s.fileId))+".aof")
	var err error
	s.aof, err = os.OpenFile(s.current, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	return s
}

type index struct {
	FileID    uint32
	Offset    int64
	Size      uint32
	ExpiredAt int64
}

func (s *store) check() (int64, error) {
	stat, err := s.aof.Stat()
	if err != nil {
		return 0, err
	}
	var offset = stat.Size()
	if offset >= s.fileSize {
		err = s.aof.Sync()
		if err != nil {
			return 0, err
		}
		s.aof.Close()
		// open file with new file id
		s.fileId++
		s.current = filepath.Join(s.path, "nodis."+strconv.Itoa(int(s.fileId))+".aof")
		s.aof, err = os.OpenFile(s.current, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
		if err != nil {
			return 0, err
		}
		// update index file
		idxFi, err := os.OpenFile(s.indexFile, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			return 0, err
		}
		defer func() {
			err := idxFi.Sync()
			if err != nil {
				log.Println("Index sync error: ", err)
			}
			idxFi.Close()
		}()
		_, err = idxFi.Seek(0, io.SeekStart)
		if err != nil {
			return 0, err
		}
		var header = make([]byte, 4)
		binary.LittleEndian.PutUint32(header, s.fileId)
		_, err = idxFi.Write(header)
		if err != nil {
			return 0, err
		}
		offset = 0
	}
	return offset, nil
}

// put a key-value pair into store
func (s *store) put(entry *Entity) error {
	s.Lock()
	defer s.Unlock()
	var idx = &index{}
	offset, err := s.check()
	if err != nil {
		return err
	}
	data, err := entry.Marshal()
	if err != nil {
		return err
	}
	idx.FileID = s.fileId
	idx.Offset = offset
	idx.Size = uint32(len(data))
	idx.ExpiredAt = entry.ExpiredAt
	s.index.Put(entry.Key, idx)
	_, err = s.aof.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (s *store) putRaw(key string, data []byte, expiredAt int64) error {
	s.Lock()
	defer s.Unlock()
	var index = &index{}
	offset, err := s.check()
	if err != nil {
		return err
	}
	index.FileID = s.fileId
	index.Offset = offset
	index.Size = uint32(len(data))
	index.ExpiredAt = expiredAt
	s.index.Put(key, index)
	_, err = s.aof.Write(data)
	if err != nil {
		return err
	}
	return nil
}

// get a value by key
func (s *store) get(key string) ([]byte, error) {
	s.RLock()
	defer s.RUnlock()
	index, ok := s.index.Get(key)
	if !ok {
		return nil, nil
	}
	if index.FileID == s.fileId {
		data := make([]byte, index.Size)
		_, err := s.aof.ReadAt(data, index.Offset)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
	// read from other file
	file := filepath.Join(s.path, "nodis."+strconv.Itoa(int(index.FileID))+".aof")
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data := make([]byte, index.Size)
	_, err = f.ReadAt(data, index.Offset)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// remove a key-value pair from store
func (s *store) remove(key string) {
	s.Lock()
	defer s.Unlock()
	s.index.Delete(key)
}

// snapshot the store
func (s *store) snapshot(path string, entries []*Entity) {
	s.RLock()
	defer s.RUnlock()
	snapshotDir := filepath.Join(path, "snapshots", time.Now().Format("20060102_150405"))
	err := os.MkdirAll(snapshotDir, 0755)
	if err != nil {
		log.Println("Snapshot mkdir error: ", err)
		return
	}
	ns := newStore(snapshotDir, s.fileSize)
	for _, entry := range entries {
		ns.put(entry)
	}
	s.index.Iter(func(key string, index *index) bool {
		if ns.index.Has(key) {
			return false
		}
		data, err := s.get(key)
		if err != nil {
			log.Println("Snapshot get error: ", err)
			return false
		}
		err = ns.putRaw(key, data, index.ExpiredAt)
		if err != nil {
			log.Println("Snapshot put error: ", err)
			return false
		}
		return false
	})
	err = ns.close()
	if err != nil {
		log.Println("Snapshot save error: ", err)
	}
}

// close the store
func (s *store) close() error {
	s.Lock()
	defer s.Unlock()
	err := s.aof.Sync()
	if err != nil {
		return err
	}
	err = s.aof.Close()
	if err != nil {
		return err
	}
	idxFi, err := os.OpenFile(s.indexFile+"~", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	_, err = idxFi.Seek(0, io.SeekStart)
	if err != nil {
		return idxFi.Close()
	}
	var header = make([]byte, 4)
	binary.LittleEndian.PutUint32(header, s.fileId)
	_, err = idxFi.Write(header)
	if err != nil {
		return err
	}
	var indexData = make(map[string][]byte, 0)
	s.index.Iter(func(key string, value *index) bool {
		data, err := binary.Marshal(value)
		if err != nil {
			return true
		}
		indexData[key] = data
		return false
	})
	data, err := binary.Marshal(indexData)
	if err != nil {
		return err
	}
	_, err = idxFi.Write(data)
	if err != nil {
		return err
	}
	if err = idxFi.Sync(); err != nil {
		return err
	}
	if err = idxFi.Close(); err != nil {
		log.Println("Close sync error: ", err)
	}
	err = os.Rename(s.indexFile+"~", s.indexFile)
	if err != nil {
		return err
	}
	return nil
}

// clear the store
func (s *store) clear() error {
	s.Lock()
	defer s.Unlock()
	err := s.aof.Close()
	if err != nil {
		return err
	}
	err = os.RemoveAll(s.path)
	if err != nil {
		return err
	}
	s.index.Clear()
	return nil
}
