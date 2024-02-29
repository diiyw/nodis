package nodis

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/kelindar/binary"

	"github.com/diyyw/nodis/ds"
	"github.com/diyyw/nodis/ds/set"
	"github.com/dolthub/swiss"
)

type Nodis struct {
	sync.RWMutex
	store    *swiss.Map[string, ds.DataStruct]
	keys     *swiss.Map[string, Key]
	options  Options
	dbFile   string
	metaFile string
}

func Open(opt Options) *Nodis {
	n := &Nodis{
		store:    swiss.NewMap[string, ds.DataStruct](16),
		keys:     swiss.NewMap[string, Key](16),
		options:  opt,
		dbFile:   filepath.Join(opt.Path, "nodis.db"),
		metaFile: filepath.Join(opt.Path, "nodis.meta"),
	}
	data, err := os.ReadFile(n.dbFile)
	if err != nil && err.Error() != "open testdata/nodis.db: no such file or directory" {
		panic(err)
	}
	for len(data) > 0 {
		err := binary.Unmarshal(data, n)
		if err != nil {
			panic(err)
		}
	}
	go func() {
		for {
			time.Sleep(opt.SyncInterval)
			n.Sync()
		}
	}()
	return n
}

// Del a key
func (n *Nodis) Del(key string) {
	n.Lock()
	n.store.Delete(key)
	n.keys.Delete(key)
	n.Unlock()
}

// Exists checks if a key exists
func (n *Nodis) Exists(key string) bool {
	n.RLock()
	k, ok := n.keys.Get(key)
	n.RUnlock()
	return ok && k.Valid()
}

// Expire the keys
func (n *Nodis) Expire(key string, seconds time.Duration) {
	n.Lock()
	k, ok := n.keys.Get(key)
	if !ok {
		n.Unlock()
		return
	}
	k.TTL = time.Now().Add(seconds).Unix()
	n.Unlock()
}

// ExpireAt the keys
func (n *Nodis) ExpireAt(key string, timestamp time.Time) {
	n.Lock()
	k, ok := n.keys.Get(key)
	if !ok || !k.Valid() {
		n.Unlock()
		return
	}
	k.TTL = timestamp.Unix()
	n.Unlock()
}

// Keys gets the keys
func (n *Nodis) Keys(pattern string) []string {
	n.RLock()
	keys := make([]string, 0, n.keys.Count())
	n.keys.Iter(func(key string, k Key) bool {
		matched, _ := filepath.Match(pattern, key)
		if matched && k.Valid() {
			keys = append(keys, key)
		}
		return true
	})
	n.RUnlock()
	return keys
}

// TTL gets the TTL
func (n *Nodis) TTL(key string) time.Duration {
	n.RLock()
	k, ok := n.keys.Get(key)
	if !ok || !k.Valid() {
		n.RUnlock()
		return -1
	}
	n.RUnlock()
	return time.Duration(k.TTL - time.Now().Unix())
}

// Rename a key
func (n *Nodis) Rename(key, newkey string) error {
	n.Lock()
	nk, ok := n.keys.Get(newkey)
	if ok && nk.Valid() {
		n.Unlock()
		return errors.New("newkey exists")
	}
	k, ok := n.keys.Get(key)
	if !ok || !k.Valid() {
		n.Unlock()
		return errors.New("key does not exist")
	}
	v, ok := n.store.Get(key)
	if !ok {
		n.Unlock()
		return errors.New("key does not exist")
	}
	n.store.Delete(key)
	n.store.Put(newkey, v)
	n.keys.Delete(key)
	n.keys.Put(newkey, k)
	n.Unlock()
	return nil
}

// Type gets the type of a key
func (n *Nodis) Type(key string) string {
	n.RLock()
	k, ok := n.keys.Get(key)
	if !ok || !k.Valid() {
		n.RUnlock()
		return "none"
	}
	n.RUnlock()
	return k.Type
}

// Tidy removes expired keys
func (n *Nodis) Tidy() (keys map[string]Key, store map[string][]byte) {
	n.Lock()
	keys = make(map[string]Key, n.keys.Count())
	store = make(map[string][]byte, n.store.Count())
	n.keys.Iter(func(key string, k Key) bool {
		if !k.Valid() {
			n.store.Delete(key)
			n.keys.Delete(key)
		}
		ds, ok := n.store.Get(key)
		if !ok {
			return true
		}
		keys[key] = k
		data, err := ds.Marshal()
		if err != nil {
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

// Set a key with a value and a TTL
func (n *Nodis) Set(key string, value any, ttl int64) {
	n.Lock()
	n.store.Put(key, set.NewSet(value))
	n.keys.Put(key, Key{TTL: ttl})
	n.Unlock()
}

// Get a key
func (n *Nodis) Get(key string) (any, bool) {
	n.RLock()
	k, ok := n.keys.Get(key)
	if !ok {
		return nil, false
	}
	if !k.Valid() {
		return nil, false
	}
	v, ok := n.store.Get(key)
	if !ok {
		return nil, false
	}
	n.RUnlock()
	return v, true
}
