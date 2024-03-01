package nodis

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/diiyw/nodis/ds/zset"
	"github.com/kelindar/binary"

	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/ds/hash"
	"github.com/diiyw/nodis/ds/list"
	"github.com/diiyw/nodis/ds/set"
	"github.com/dolthub/swiss"
)

type Nodis struct {
	sync.RWMutex
	store    *swiss.Map[string, ds.DataStruct]
	keys     *swiss.Map[string, *Key]
	options  Options
	dbFile   string
	metaFile string
}

func Open(opt Options) *Nodis {
	n := &Nodis{
		store:    swiss.NewMap[string, ds.DataStruct](16),
		keys:     swiss.NewMap[string, *Key](16),
		options:  opt,
		dbFile:   filepath.Join(opt.Path, "nodis.db"),
		metaFile: filepath.Join(opt.Path, "nodis.meta"),
	}
	loadData(n)
	go func() {
		for {
			time.Sleep(opt.SyncInterval)
			err := n.Sync()
			if err != nil {
				log.Println("Sync: ", err)
			}
		}
	}()
	return n
}

func loadData(n *Nodis) {
	metadata, err := os.ReadFile(n.metaFile)
	if err == nil {
		if len(metadata) > 0 {
			var keys = make(map[string]*Key)
			err := binary.Unmarshal(metadata, &keys)
			if err != nil {
				panic(err)
			}
			for name, k := range keys {
				n.keys.Put(name, k)
			}
			dbData, err := os.ReadFile(n.dbFile)
			if err != nil {
				panic(err)
			}
			if len(dbData) == 0 {
				return
			}
			store := make(map[string][]byte, n.keys.Count())
			err = binary.Unmarshal(dbData, &store)
			if err != nil {
				panic(err)
			}
			n.keys.Iter(func(key string, k *Key) bool {
				data, ok := store[key]
				if !ok {
					return false
				}
				switch k.Type {
				case "set":
					s := set.NewSet()
					err := s.Unmarshal(data)
					if err != nil {
						log.Println("SET: Unmarshal ", err)
						return false
					}
					n.store.Put(key, s)
				case "zset":
					z := zset.NewSortedSet()
					err := z.Unmarshal(data)
					if err != nil {
						log.Println("ZSET: Unmarshal ", err)
						return false
					}
					n.store.Put(key, z)
				case "list":
					l := list.NewDoublyLinkedList()
					err := l.Unmarshal(data)
					if err != nil {
						log.Println("LIST: Unmarshal ", err)
						return false
					}
					n.store.Put(key, l)
				case "hash":
					h := hash.NewHashMap()
					err := h.Unmarshal(data)
					if err != nil {
						log.Println("HASH: Unmarshal ", err)
						return false
					}
				}
				return false
			})
		}
	}
}

// Tidy removes expired keys
func (n *Nodis) Tidy() (keys map[string]*Key, store map[string][]byte) {
	n.Lock()
	keys = make(map[string]*Key, n.keys.Count())
	store = make(map[string][]byte, n.store.Count())
	n.keys.Iter(func(key string, k *Key) bool {
		if k.expired() {
			n.store.Delete(key)
			n.keys.Delete(key)
		}
		v, ok := n.store.Get(key)
		if !ok {
			return true
		}
		keys[key] = k
		data, err := v.Marshal()
		if err != nil {
			log.Println("Tidy: ", err)
			return true
		}
		store[key] = data
		return false
	})
	n.Unlock()
	return
}

// Sync saves the data to disk
func (n *Nodis) Sync() error {
	keys, stores := n.Tidy()
	var buf bytes.Buffer
	buf.Grow(64)
	n.RLock()
	defer n.RUnlock()
	err := binary.MarshalTo(stores, &buf)
	if err != nil {
		return err
	}
	err = os.WriteFile(n.dbFile, buf.Bytes(), 0644)
	if err != nil {
		return err
	}
	buf.Reset()
	err = binary.MarshalTo(keys, &buf)
	if err != nil {
		return err
	}
	err = os.WriteFile(n.metaFile, buf.Bytes(), 0644)
	if err != nil {
		return err
	}
	return nil
}

func (n *Nodis) getDs(key string) ds.DataStruct {
	n.Lock()
	defer n.Unlock()
	if !n.exists(key) {
		return nil
	}
	l, ok := n.store.Get(key)
	if !ok {
		return nil
	}
	return l
}

func (n *Nodis) saveDs(key string, ds ds.DataStruct, ttl int64) {
	n.Lock()
	defer n.Unlock()
	n.store.Put(key, ds)
	n.keys.Put(key, newKey(ds.GetType(), ttl))
}
