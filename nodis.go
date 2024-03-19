package nodis

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/diiyw/nodis/ds"
	"github.com/tidwall/btree"
)

type Nodis struct {
	sync.RWMutex
	dataStructs btree.Map[string, ds.DataStruct]
	keys        btree.Map[string, *Key]
	options     *Options
	store       *store
	closed      bool
}

func Open(opt *Options) *Nodis {
	if opt.FileSize == 0 {
		opt.FileSize = FileSizeGB
	}
	n := &Nodis{
		options: opt,
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
				err := n.store.put(newEntry(key, d, k.ExpiredAt))
				if err != nil {
					log.Println("Recycle: ", err)
				}
			}
		}
		return true
	})
}

// getChangedEntries returns all keys that have been getChangedEntries
func (n *Nodis) getChangedEntries() []*Entity {
	entries := make([]*Entity, 0)
	n.keys.Scan(func(key string, k *Key) bool {
		if !k.changed.Load() || k.expired() {
			return true
		}
		d, ok := n.dataStructs.Get(key)
		if !ok {
			return true
		}
		entries = append(entries, newEntry(key, d, k.ExpiredAt))
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
			k = newKey(d.GetType(), ttl)
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

func parseDs(data []byte) (*Entity, error) {
	entry := &Entity{}
	err := entry.Unmarshal(data)
	if err != nil {
		return nil, err
	}
	return entry, nil
}
