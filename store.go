package nodis

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/diiyw/nodis/fs"
	"github.com/kelindar/binary"
	"github.com/tidwall/btree"
)

type store struct {
	sync.RWMutex
	fileSize   int64
	fileId     uint32
	path       string
	current    string
	indexFile  string
	aof        fs.File
	index      btree.Map[string, *index]
	filesystem fs.Fs
}

func newStore(path string, fileSize int64, filesystem fs.Fs) *store {
	s := &store{
		path:      path,
		fileSize:  fileSize,
		indexFile: filepath.Join(path, "nodis.index"),
	}

	_ = filesystem.MkdirAll(path)
	idxFile, err := filesystem.OpenFile(s.indexFile, os.O_RDWR|os.O_CREATE|os.O_APPEND)
	if err != nil {
		panic(err)
	}
	data, err := idxFile.ReadAll()
	if err != nil {
		panic(err)
	}
	if len(data) > 4 {
		if len(data[4:]) > 0 {
			var m map[string]*index
			err := binary.Unmarshal(data[4:], &m)
			if err != nil {
				panic(err)
			}
			for k, v := range m {
				s.index.Set(k, v)
			}
		}
		s.fileId = binary.LittleEndian.Uint32(data[:4])
	}
	s.filesystem = filesystem
	s.current = filepath.Join(path, "nodis."+strconv.Itoa(int(s.fileId))+".aof")
	s.aof, err = s.filesystem.OpenFile(s.current, os.O_RDWR|os.O_CREATE|os.O_APPEND)
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
		binary.LittleEndian.PutUint32(header, s.fileId)
		_, err = idxFi.WriteAt(header, 0)
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
	s.index.Set(entry.Key, idx)
	_, err = s.aof.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (s *store) putRaw(key string, data []byte, expiredAt int64) error {
	s.Lock()
	defer s.Unlock()
	var idx = &index{}
	offset, err := s.check()
	if err != nil {
		return err
	}
	idx.FileID = s.fileId
	idx.Offset = offset
	idx.Size = uint32(len(data))
	idx.ExpiredAt = expiredAt
	s.index.Set(key, idx)
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
	idx, ok := s.index.Get(key)
	if !ok {
		return nil, nil
	}
	if idx.FileID == s.fileId {
		data := make([]byte, idx.Size)
		_, err := s.aof.ReadAt(data, idx.Offset)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
	// read from other file
	file := filepath.Join(s.path, "nodis."+strconv.Itoa(int(idx.FileID))+".aof")
	f, err := s.filesystem.OpenFile(file, os.O_RDONLY)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data := make([]byte, idx.Size)
	_, err = f.ReadAt(data, idx.Offset)
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
	err := s.filesystem.MkdirAll(snapshotDir)
	if err != nil {
		log.Println("Snapshot mkdir error: ", err)
		return
	}
	ns := newStore(snapshotDir, s.fileSize, s.filesystem)
	for _, entry := range entries {
		ns.put(entry)
	}
	s.index.Scan(func(key string, index *index) bool {
		if _, ok := ns.index.Get(key); !ok {
			return true
		}
		data, err := s.get(key)
		if err != nil {
			log.Println("Snapshot get error: ", err)
			return true
		}
		err = ns.putRaw(key, data, index.ExpiredAt)
		if err != nil {
			log.Println("Snapshot put error: ", err)
			return true
		}
		return true
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
	err := s.aof.Close()
	if err != nil {
		return err
	}
	idxFile, err := s.filesystem.OpenFile(s.indexFile+"~", os.O_RDWR|os.O_CREATE|os.O_APPEND)
	if err != nil {
		return err
	}
	var header = make([]byte, 4)
	binary.LittleEndian.PutUint32(header, s.fileId)
	_, err = idxFile.Write(header)
	if err != nil {
		return err
	}
	var indexData = make(map[string][]byte, 0)
	s.index.Scan(func(key string, value *index) bool {
		data, err := binary.Marshal(value)
		if err != nil {
			return false
		}
		indexData[key] = data
		return true
	})
	data, err := binary.Marshal(indexData)
	if err != nil {
		return err
	}
	_, err = idxFile.Write(data)
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

// clear the store
func (s *store) clear() error {
	s.Lock()
	defer s.Unlock()
	err := s.aof.Truncate(0)
	if err != nil {
		return err
	}
	s.index.Clear()
	return nil
}
