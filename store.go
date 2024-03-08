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
	chunkSize int64
	aof       *os.File
	path      string
	current   string
	index     *swiss.Map[string, *index]
	chunkId   uint32
	indexFile string
}

func newStore(path string, chunkSize int64) *store {
	s := &store{
		path:      path,
		chunkSize: chunkSize,
		index:     swiss.NewMap[string, *index](32),
		indexFile: filepath.Join(path, "nodis.index"),
	}
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
		s.chunkId = binary.LittleEndian.Uint32(data[:4])
	}
	s.current = filepath.Join(path, "nodis."+strconv.Itoa(int(s.chunkId))+".aof")
	var err error
	s.aof, err = os.OpenFile(s.current, os.O_CREATE|os.O_RDWR|os.O_APPEND|os.O_SYNC, 0644)
	if err != nil {
		panic(err)
	}
	return s
}

type index struct {
	ChunkID   uint32
	Offset    int64
	Size      uint32
	ExpiredAt int64
}

// put a key-value pair into store
func (s *store) put(key string, value []byte, expiredAt int64) error {
	s.Lock()
	defer s.Unlock()
	var index = &index{}
	stat, err := s.aof.Stat()
	if err != nil {
		return err
	}
	var offset = stat.Size()
	if offset >= s.chunkSize {
		s.aof.Close()
		// open file with new chunk id
		s.chunkId++
		s.current = filepath.Join(s.path, "nodis."+strconv.Itoa(int(s.chunkId))+".aof")
		s.aof, err = os.OpenFile(s.current, os.O_CREATE|os.O_RDWR|os.O_APPEND|os.O_SYNC, 0644)
		if err != nil {
			return err
		}
		// update index file
		idxFi, err := os.OpenFile(s.indexFile, os.O_CREATE|os.O_RDWR|os.O_SYNC, 0644)
		if err != nil {
			return err
		}
		defer idxFi.Close()
		_, err = idxFi.Seek(0, io.SeekStart)
		if err != nil {
			return err
		}
		var header = make([]byte, 4)
		binary.LittleEndian.PutUint32(header, s.chunkId)
		_, err = idxFi.Write(header)
		if err != nil {
			return err
		}
		offset = 0
	}
	index.ChunkID = s.chunkId
	index.Offset = offset
	index.Size = uint32(len(value))
	index.ExpiredAt = expiredAt
	s.index.Put(key, index)
	_, err = s.aof.Write(value)
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
	if index.ChunkID == s.chunkId {
		data := make([]byte, index.Size)
		_, err := s.aof.ReadAt(data, index.Offset)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
	// read from other file
	file := filepath.Join(s.path, "nodis."+strconv.Itoa(int(index.ChunkID))+".aof")
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
func (s *store) snapshot() {
	s.RLock()
	defer s.RUnlock()
	var keys = make(map[string]int64, 0)
	s.index.Iter(func(key string, value *index) bool {
		if value.ExpiredAt > time.Now().Unix() {
			keys[key] = value.ExpiredAt
		}
		return false
	})
	go func(snapshotKeys map[string]int64) {
		snapshotDir := filepath.Join(s.path, "snapshots", time.Now().Format("20060102150405"))
		err := os.MkdirAll(snapshotDir, 0755)
		if err != nil {
			log.Println("Snapshot mkdir error: ", err)
			return
		}
		ns := newStore(snapshotDir, s.chunkSize)
		for key, expiredAt := range snapshotKeys {
			value, _ := s.get(key)
			ns.put(key, value, expiredAt)
		}
		err = ns.close()
		if err != nil {
			log.Println("Snapshot save error: ", err)
		}
	}(keys)
}

// close the store
func (s *store) close() error {
	err := s.aof.Close()
	if err != nil {
		return err
	}
	idxFi, err := os.OpenFile(s.indexFile+"~", os.O_CREATE|os.O_RDWR|os.O_SYNC|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer idxFi.Close()
	_, err = idxFi.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	var header = make([]byte, 4)
	binary.LittleEndian.PutUint32(header, s.chunkId)
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
		panic(err)
	}
	err = os.Rename(s.indexFile+"~", s.indexFile)
	if err != nil {
		return err
	}
	return nil
}
