package nodis

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/diiyw/nodis/ds/set"

	"github.com/diiyw/nodis/ds/zset"

	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/ds/hash"
	"github.com/diiyw/nodis/ds/list"
	"github.com/diiyw/nodis/ds/str"
	"github.com/dolthub/swiss"
)

type Nodis struct {
	sync.RWMutex
	dataStructs *swiss.Map[string, ds.DataStruct]
	keys        *swiss.Map[string, *Key]
	options     *Options
	store       *store
}

func Open(opt *Options) *Nodis {
	n := &Nodis{
		dataStructs: swiss.NewMap[string, ds.DataStruct](16),
		keys:        swiss.NewMap[string, *Key](16),
		options:     opt,
	}
	stat, err := os.Stat(opt.Path)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(opt.Path, 0755)
			if err != nil {
				panic(err)
			}
		}
	} else if !stat.IsDir() {
		panic("Path is not a directory")
	}
	n.store = newStore(opt.Path, opt.FileSize)
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
	n.store.snapshot(path)
}

// Recycle removes expired and unused keys
func (n *Nodis) Recycle() {
	n.Lock()
	defer n.Unlock()
	n.keys.Iter(func(key string, k *Key) bool {
		if k.expired() {
			n.dataStructs.Delete(key)
			n.keys.Delete(key)
			n.store.remove(key)
			return false
		}
		if k.lastUse != 0 && k.lastUse < time.Now().Unix()-int64(n.options.RecycleDuration.Seconds()) {
			d, ok := n.dataStructs.Get(key)
			n.dataStructs.Delete(key)
			n.keys.Delete(key)
			if ok {
				go func() {
					// save to disk
					data, err := d.Marshal()
					if err != nil {
						log.Println("Tidy Marshal: ", err)
						return
					}
					n.store.put(key, append([]byte{byte(d.GetType())}, data...), k.ExpiredAt)
				}()
			}
		}
		if !n.store.index.Has(key) {
			d, ok := n.dataStructs.Get(key)
			if ok {
				go func() {
					// save to disk
					data, err := d.Marshal()
					if err != nil {
						log.Println("Tidy Marshal: ", err)
						return
					}
					n.store.put(key, append([]byte{byte(d.GetType())}, data...), k.ExpiredAt)
				}()
			}
		}
		return false
	})
}

// sync saves the data to disk
func (n *Nodis) sync() error {
	n.keys.Iter(func(key string, k *Key) bool {
		if !k.changed {
			return false
		}
		d, ok := n.dataStructs.Get(key)
		if !ok {
			return false
		}
		data, err := d.Marshal()
		if err != nil {
			log.Println("Sync Marshal: ", err)
		}
		k.changed = false
		n.store.put(key, append([]byte{byte(d.GetType())}, data...), k.ExpiredAt)
		return false
	})
	return nil
}

// Close the store
func (n *Nodis) Close() error {
	n.Lock()
	defer n.Unlock()
	// save values to disk
	n.sync()
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
			n.dataStructs.Put(key, d)
			k = newKey(d.GetType(), ttl)
			n.keys.Put(key, k)
			return k, d
		}
		return nil, nil
	}
	return k, d
}

// Clear removes all keys from the store
func (n *Nodis) Clear(key string) {
	n.Lock()
	defer n.Unlock()
	_, ok := n.exists(key)
	if !ok {
		return
	}
	n.dataStructs.Delete(key)
	n.keys.Delete(key)
	n.store.clear()
}

func parseDs(data []byte) (d ds.DataStruct) {
	dataType := ds.DataType(data[0])
	data = data[1:]
	switch dataType {
	case ds.String:
		s := str.NewString()
		err := s.Unmarshal(data)
		if err != nil {
			log.Println("String: Unmarshal ", err)
		}
		d = s
	case ds.ZSet:
		z := zset.NewSortedSet()
		err := z.Unmarshal(data)
		if err != nil {
			log.Println("ZSET: Unmarshal ", err)
		}
		d = z
	case ds.List:
		l := list.NewDoublyLinkedList()
		err := l.Unmarshal(data)
		if err != nil {
			log.Println("LIST: Unmarshal ", err)
		}
		d = l
	case ds.Hash:
		h := hash.NewHashMap()
		err := h.Unmarshal(data)
		if err != nil {
			log.Println("HASH: Unmarshal ", err)
		}
		d = h
	case ds.Set:
		s := set.NewSet()
		err := s.Unmarshal(data)
		if err != nil {
			log.Println("SET: Unmarshal ", err)
		}
		d = s
	}
	return
}
