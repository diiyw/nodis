package nodis

import (
	"hash/crc32"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"

	"encoding/binary"

	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/ds/hash"
	"github.com/diiyw/nodis/ds/list"
	"github.com/diiyw/nodis/ds/set"
	"github.com/diiyw/nodis/ds/str"
	"github.com/diiyw/nodis/ds/zset"
	"github.com/diiyw/nodis/fs"
	"github.com/diiyw/nodis/pb"
	"github.com/tidwall/btree"
)

type store struct {
	sync.RWMutex
	fileSize     int64
	fileId       uint16
	path         string
	current      string
	indexFile    string
	aof          fs.File
	keys         btree.Map[string, *Key]
	values       btree.Map[string, ds.DataStruct]
	filesystem   fs.Fs
	locks        []*sync.RWMutex
	lockPoolSize int
	closed       bool
}

func newStore(path string, fileSize int64, lockPoolSize int, filesystem fs.Fs) *store {
	s := &store{
		path:         path,
		fileSize:     fileSize,
		indexFile:    filepath.Join(path, "nodis.index"),
		lockPoolSize: lockPoolSize,
		locks:        make([]*sync.RWMutex, lockPoolSize),
	}
	for i := 0; i < lockPoolSize; i++ {
		s.locks[i] = &sync.RWMutex{}
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
		if len(data[2:]) > 0 {
			var idx = &pb.Index{}
			err = proto.Unmarshal(data[2:], idx)
			if err != nil {
				panic(err)
			}
			for _, v := range idx.Items {
				var i = &Key{}
				i.unmarshal(v.Data)
				s.keys.Set(v.Key, i)
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

func fnv32(key string) uint32 {
	h := uint32(2166136261)
	for i := 0; i < len(key); i++ {
		h *= 16777619
		h ^= uint32(key[i])
	}
	return h
}

func (s *store) spread(hashCode uint32) *sync.RWMutex {
	tableSize := uint32(s.lockPoolSize)
	return s.locks[(tableSize-1)&hashCode]
}

func (s *store) getLocker(key string) *sync.RWMutex {
	return s.spread(fnv32(key))
}

func (s *store) writeKey(key string, newFn func() ds.DataStruct) *metadata {
	locker := s.getLocker(key)
	locker.Lock()
	k, ok := s.keys.Get(key)
	if ok {
		if k.expired(time.Now().UnixMilli()) {
			if newFn == nil {
				return newEmptyMetadata(locker, true)
			}
		}
		d, ok := s.values.Get(key)
		if ok {
			return newMetadata(k, d, true, locker)
		}
		return s.fromStorage(k, true, locker)
	}
	if newFn != nil {
		s.Lock()
		defer s.Unlock()
		d := newFn()
		if d == nil {
			return newEmptyMetadata(locker, true)
		}
		k = newKey()
		s.keys.Set(key, k)
		s.values.Set(key, d)
		meta := newMetadata(k, d, true, locker)
		meta.markChanged()
		return meta
	}
	return newEmptyMetadata(locker, true)
}

func (s *store) readKey(key string) *metadata {
	locker := s.getLocker(key)
	locker.RLock()
	k, ok := s.keys.Get(key)
	if ok {
		if k.expired(time.Now().UnixMilli()) {
			return newEmptyMetadata(locker, false)
		}
		d, ok := s.values.Get(key)
		if !ok {
			// read from storage
			return s.fromStorage(k, false, locker)
		}
		return newMetadata(k, d, false, locker)
	}
	return newEmptyMetadata(locker, false)
}

func (s *store) delKey(key string) {
	s.keys.Delete(key)
	s.values.Delete(key)
}

func (s *store) fromStorage(key *Key, writable bool, locker *sync.RWMutex) *metadata {
	// try get from storage
	v, err := s.getKey(key)
	if err == nil && len(v) > 0 {
		key, d, expiration, err := s.parseDs(v)
		if err != nil {
			log.Println("Parse DataStruct:", err)
			return newEmptyMetadata(locker, writable)
		}
		if d != nil {
			s.values.Set(key, d)
			k := newKey()
			k.expiration = expiration
			s.keys.Set(key, k)
			return newMetadata(k, d, writable, locker)
		}
	}
	return newEmptyMetadata(locker, writable)
}

func (s *store) parseEntry(data []byte) (*pb.Entry, error) {
	c32 := binary.LittleEndian.Uint32(data)
	if c32 != crc32.ChecksumIEEE(data[4:]) {
		return nil, ErrCorruptedData
	}
	var entity pb.Entry
	if err := proto.Unmarshal(data[4:], &entity); err != nil {
		return nil, err
	}
	return &entity, nil
}

// parseDs the data
func (s *store) parseDs(data []byte) (string, ds.DataStruct, int64, error) {
	var entity, err = s.parseEntry(data)
	if err != nil {
		return "", nil, 0, err
	}
	var dataStruct ds.DataStruct
	switch ds.DataType(entity.Type) {
	case ds.String:
		v := str.NewString()
		v.SetValue(entity.GetStringValue().Value)
		dataStruct = v
	case ds.ZSet:
		z := zset.NewSortedSet()
		z.SetValue(entity.GetZSetValue().Values)
		dataStruct = z
	case ds.List:
		l := list.NewDoublyLinkedList()
		l.SetValue(entity.GetListValue().Values)
		dataStruct = l
	case ds.Hash:
		h := hash.NewHashMap()
		h.SetValue(entity.GetHashValue().Values)
		dataStruct = h
	case ds.Set:
		v := set.NewSet()
		v.SetValue(entity.GetSetValue().Values)
		dataStruct = v
	}
	return entity.Key, dataStruct, entity.Expiration, nil
}

// flushChanges flush changed keys to disk
func (s *store) flushChanges() {
	now := time.Now().UnixMilli()
	s.keys.Scan(func(key string, k *Key) bool {
		locker := s.getLocker(key)
		locker.RLock()
		defer locker.RUnlock()
		if !k.changed || k.expired(now) {
			return true
		}
		d, ok := s.values.Get(key)
		if !ok {
			return true
		}
		// save to storage
		err := s.putKv(key, k, d)
		if err != nil {
			log.Println("Flush changes: ", err)
		}
		return true
	})
}

// tidy removes expired and unused keys
func (s *store) tidy(ms int64) {
	s.Lock()
	defer s.Unlock()
	if s.closed {
		return
	}
	now := time.Now().UnixMilli()
	recycleTime := now - ms
	s.keys.Scan(func(key string, k *Key) bool {
		locker := s.getLocker(key)
		locker.Lock()
		defer locker.Unlock()
		if k.expired(now) {
			s.delKey(key)
			return true
		}
		if k.lastUse != 0 && k.lastUse <= recycleTime {
			d, ok := s.values.Get(key)
			if ok {
				k.changed = false
				// save to disk
				err := s.putKv(key, k, d)
				if err != nil {
					log.Println("Recycle: ", err)
				}
			}
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

func (s *store) putKv(k string, key *Key, value ds.DataStruct) error {
	offset, err := s.check()
	if err != nil {
		return err
	}
	key.fileId = s.fileId
	key.offset = offset
	entry := newEntry(k, value, key.expiration)
	data, err := entry.Marshal()
	if err != nil {
		return err
	}
	key.size = uint32(len(data))
	_, err = s.aof.Write(data)
	if err != nil {
		return err
	}
	return nil
}

// putEntry a key-value pair into store
func (s *store) putEntry(entry *pb.Entry) error {
	var key = &Key{}
	offset, err := s.check()
	if err != nil {
		return err
	}
	data, err := entry.Marshal()
	if err != nil {
		return err
	}
	key.fileId = s.fileId
	key.offset = offset
	key.size = uint32(len(data))
	key.expiration = entry.Expiration
	s.keys.Set(entry.Key, key)
	_, err = s.aof.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (s *store) putRaw(key string, data []byte, k *Key) error {
	offset, err := s.check()
	if err != nil {
		return err
	}
	k.fileId = s.fileId
	k.offset = offset
	k.size = uint32(len(data))
	s.keys.Set(key, k)
	_, err = s.aof.Write(data)
	if err != nil {
		return err
	}
	return nil
}

// get a value by key
func (s *store) get(key string) ([]byte, error) {
	k, ok := s.keys.Get(key)
	if !ok {
		return nil, nil
	}
	return s.getKey(k)
}

func (s *store) getKey(key *Key) ([]byte, error) {
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
	s.Lock()
	defer s.Unlock()
	snapshotDir := filepath.Join(path, "snapshots", time.Now().Format("20060102_150405"))
	err := s.filesystem.MkdirAll(snapshotDir)
	if err != nil {
		log.Println("Snapshot mkdir error: ", err)
		return
	}
	s.flushChanges()
	ns := newStore(snapshotDir, s.fileSize, 0, s.filesystem)
	s.keys.Scan(func(key string, k *Key) bool {
		if _, ok := ns.keys.Get(key); !ok {
			return true
		}
		data, err := s.get(key)
		if err != nil {
			log.Println("Snapshot get error: ", err)
			return true
		}
		err = ns.putRaw(key, data, k)
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
	s.closed = true
	s.flushChanges()
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
		Items: make([]*pb.Index_Item, 0, s.keys.Len()),
	}
	s.keys.Scan(func(key string, k *Key) bool {
		locker := s.getLocker(key)
		locker.Lock()
		indexData.Items = append(indexData.Items, &pb.Index_Item{
			Key:  key,
			Data: k.marshal(),
		})
		locker.Unlock()
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
	s.keys.Clear()
	s.values.Clear()
	return nil
}
