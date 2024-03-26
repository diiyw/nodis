package nodis

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"

	"encoding/binary"

	"github.com/diiyw/nodis/fs"
	"github.com/diiyw/nodis/pb"
	"github.com/tidwall/btree"
)

type store struct {
	sync.RWMutex
	fileSize   int64
	fileId     uint16
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
	err = idxFile.Close()
	if err != nil {
		panic(err)
	}
	if len(data) > 2 {
		if len(data[2:]) > 0 {
			var idx = &pb.Index{}
			err = proto.Unmarshal(data[2:], idx)
			if err != nil {
				panic(err)
			}
			for _, v := range idx.Items {
				var i = &index{}
				i.unmarshal(v.Data)
				s.index.Set(v.Key, i)
			}
		}
		s.fileId = binary.LittleEndian.Uint16(data[:2])
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
	offset     int64
	expiration int64
	size       uint32
	fileID     uint16
}

// marshal index to bytes
func (i *index) marshal() []byte {
	var b [22]byte
	binary.LittleEndian.PutUint64(b[0:8], uint64(i.offset))
	binary.LittleEndian.PutUint64(b[8:16], uint64(i.expiration))
	binary.LittleEndian.PutUint32(b[16:20], i.size)
	binary.LittleEndian.PutUint16(b[20:22], i.fileID)
	return b[:]
}

// unmarshal bytes to index
func (i *index) unmarshal(b []byte) {
	i.offset = int64(binary.LittleEndian.Uint64(b[0:8]))
	i.expiration = int64(binary.LittleEndian.Uint64(b[8:16]))
	i.size = binary.LittleEndian.Uint32(b[16:20])
	i.fileID = binary.LittleEndian.Uint16(b[20:22])
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

// put a key-value pair into store
func (s *store) put(entry *pb.Entry) error {
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
	idx.fileID = s.fileId
	idx.offset = offset
	idx.size = uint32(len(data))
	idx.expiration = entry.Expiration
	s.index.Set(entry.Key, idx)
	_, err = s.aof.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (s *store) putRaw(key string, data []byte, expiration int64) error {
	s.Lock()
	defer s.Unlock()
	var idx = &index{}
	offset, err := s.check()
	if err != nil {
		return err
	}
	idx.fileID = s.fileId
	idx.offset = offset
	idx.size = uint32(len(data))
	idx.expiration = expiration
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
	if idx.fileID == s.fileId {
		data := make([]byte, idx.size)
		_, err := s.aof.ReadAt(data, idx.offset)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
	// read from other file
	file := filepath.Join(s.path, "nodis."+strconv.Itoa(int(idx.fileID))+".aof")
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
	data := make([]byte, idx.size)
	_, err = f.ReadAt(data, idx.offset)
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
func (s *store) snapshot(path string, entries []*pb.Entry) {
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
		err = ns.put(entry)
		if err != nil {
			log.Println("Snapshot put error: ", err)
			return
		}
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
		err = ns.putRaw(key, data, index.expiration)
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
	var header = make([]byte, 2)
	binary.LittleEndian.PutUint16(header, s.fileId)
	_, err = idxFile.Write(header)
	if err != nil {
		return err
	}
	indexData := &pb.Index{
		Items: make([]*pb.Index_Item, 0, s.index.Len()),
	}
	s.index.Scan(func(key string, i *index) bool {
		indexData.Items = append(indexData.Items, &pb.Index_Item{
			Key:  key,
			Data: i.marshal(),
		})
		return true
	})
	data, err := proto.Marshal(indexData)
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
