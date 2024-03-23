package nodis

import (
	"hash/crc32"
	"log"
	"os"
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
	"github.com/diiyw/nodis/pb"
	"github.com/diiyw/nodis/watch"
	"github.com/tidwall/btree"
	"google.golang.org/protobuf/proto"
)

type Nodis struct {
	sync.RWMutex
	dataStructs btree.Map[string, ds.DataStruct]
	keys        btree.Map[string, *Key]
	options     *Options
	store       *store
	closed      bool
	watchers    btree.Map[string, *watch.Watcher]
}

func Open(opt *Options) *Nodis {
	if opt.FileSize == 0 {
		opt.FileSize = FileSizeGB
	}
	if opt.Filesystem == nil {
		opt.Filesystem = &fs.Disk{}
	}
	n := &Nodis{
		options: opt,
	}
	isDir, err := opt.Filesystem.IsDir(opt.Path)
	if err != nil {
		if os.IsNotExist(err) {
			err = opt.Filesystem.MkdirAll(opt.Path)
			if err != nil {
				panic(err)
			}
		}
	} else if !isDir {
		panic("Path is not a directory")
	}
	n.store = newStore(opt.Path, opt.FileSize, opt.Filesystem)
	go func() {
		if opt.RecycleDuration != 0 {
			for {
				time.Sleep(opt.RecycleDuration)
				n.Recycle()
			}
		}
	}()
	go func() {
		if opt.SnapshotDuration != 0 {
			for {
				time.Sleep(opt.SnapshotDuration)
				n.Snapshot(n.store.path)
				log.Println("Snapshot at", time.Now().Format("2006-01-02 15:04:05"))
			}
		}
	}()
	return n
}

// Snapshot saves the data to disk
func (n *Nodis) Snapshot(path string) {
	n.Recycle()
	n.store.snapshot(path, n.getChangedEntries())
}

// Recycle removes expired and unused keys
func (n *Nodis) Recycle() {
	n.Lock()
	defer n.Unlock()
	if n.closed {
		return
	}
	now := uint32(time.Now().Unix())
	recycleTime := now - uint32(n.options.RecycleDuration.Seconds())
	n.keys.Scan(func(key string, k *Key) bool {
		if k.expired() {
			n.dataStructs.Delete(key)
			n.keys.Delete(key)
			n.store.remove(key)
			return true
		}
		lastUse := k.lastUse.Load()
		if lastUse != 0 && lastUse <= recycleTime {
			d, ok := n.dataStructs.Get(key)
			n.dataStructs.Delete(key)
			n.keys.Delete(key)
			if ok {
				k.changed.Store(false)
				// save to disk
				err := n.store.put(newEntity(key, d, k.Expiration))
				if err != nil {
					log.Println("Recycle: ", err)
				}
			}
		}
		return true
	})
}

// getChangedEntries returns all keys that have been getChangedEntries
func (n *Nodis) getChangedEntries() []*pb.Entity {
	entries := make([]*pb.Entity, 0)
	n.keys.Scan(func(key string, k *Key) bool {
		if !k.changed.Load() || k.expired() {
			return true
		}
		d, ok := n.dataStructs.Get(key)
		if !ok {
			return true
		}
		entries = append(entries, newEntity(key, d, k.Expiration))
		return true
	})
	return entries
}

// Close the store
func (n *Nodis) Close() error {
	n.Lock()
	defer n.Unlock()
	// save values to disk
	entries := n.getChangedEntries()
	for _, entry := range entries {
		err := n.store.put(entry)
		if err != nil {
			return err
		}
	}
	n.closed = true
	return n.store.close()
}

func (n *Nodis) getDs(key string, newFn func() ds.DataStruct, ttl int64) (*Key, ds.DataStruct) {
	n.Lock()
	defer n.Unlock()
	k, ok := n.exists(key)
	if !ok && newFn == nil {
		return nil, nil
	}
	d, ok := n.dataStructs.Get(key)
	if !ok {
		if newFn != nil {
			d = newFn()
			n.dataStructs.Set(key, d)
			k = newKey(d.Type(), ttl)
			k.lastUse.Store(uint32(time.Now().Unix()))
			n.keys.Set(key, k)
			return k, d
		}
		return nil, nil
	}
	k.lastUse.Store(uint32(time.Now().Unix()))
	return k, d
}

// Clear removes all keys from the store
func (n *Nodis) Clear() {
	n.Lock()
	defer n.Unlock()
	n.dataStructs.Clear()
	n.keys.Clear()
	err := n.store.clear()
	if err != nil {
		log.Println("Clear: ", err)
	}
}

// parseDs the data
func parseDs(data []byte) (key string, dataStruct ds.DataStruct, expiration int64, err error) {
	c32 := binary.LittleEndian.Uint32(data)
	if c32 != crc32.ChecksumIEEE(data[4:]) {
		return "", nil, 0, ErrCorruptedData
	}
	var entity pb.Entity
	if err := proto.Unmarshal(data[4:], &entity); err != nil {
		return "", nil, 0, err
	}
	key, expiration = entity.Key, entity.Expiration
	switch ds.DataType(entity.Type) {
	case ds.String:
		s := str.NewString()
		s.SetValue(entity.GetStringValue().Value)
		dataStruct = s
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
		s := set.NewSet()
		s.SetValue(entity.GetSetValue().Values)
		dataStruct = s
	}
	return
}

func (n *Nodis) notify(ops ...*pb.Op) {
	if n.watchers.Len() == 0 {
		return
	}
	go func() {
		n.watchers.Scan(func(key string, w *watch.Watcher) bool {
			for _, op := range ops {
				if w.Matched(op.Key) {
					w.Push(op.Operation)
				}
			}
			return true
		})
	}()
}
