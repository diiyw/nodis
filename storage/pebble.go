package storage

import (
	"os"
	"path/filepath"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/diiyw/nodis/ds"
)

type Pebble struct {
	path    string
	options *pebble.Options
	db      *pebble.DB
}

// NewPebble creates a new Pebble storage
func NewPebble(path string, options *pebble.Options) *Pebble {
	return &Pebble{
		path:    path,
		options: options,
	}
}

func (p *Pebble) Init() error {
	db, err := pebble.Open(p.path, p.options)
	if err != nil {
		return err
	}
	p.db = db
	return nil
}

// Get the value from the storage
func (p *Pebble) Get(key string) (ds.Value, error) {
	v, closer, err := p.db.Get([]byte(key))
	if err != nil {
		return nil, err
	}
	defer closer.Close()
	entry, err := parseEntry(v)
	if err != nil {
		return nil, err
	}
	return entry.GetValue()
}

// Put the value to the storage
func (p *Pebble) Put(key *ds.Key, value ds.Value) error {
	entry := NewEntry(value)
	data := entry.encode()
	return p.db.Set([]byte(key.Name), data, pebble.Sync)
}

// Delete the value from the storage
func (p *Pebble) Delete(key string) error {
	return p.db.Delete([]byte(key), pebble.Sync)
}

// Clear the storage
func (p *Pebble) Clear() error {
	err := p.db.Close()
	if err != nil {
		return err
	}
	err = os.RemoveAll(p.path)
	if err != nil {
		return err
	}
	db, err := pebble.Open(p.path, p.options)
	if err != nil {
		return err
	}
	p.db = db
	return nil
}

// Close the storage
func (p *Pebble) Close() error {
	return p.db.Close()
}

// Snapshot the storage
func (p *Pebble) Snapshot() error {
	dstDir := time.Now().Format("20060102150405")
	return p.db.Checkpoint(filepath.Join(p.path, dstDir))
}

// ScanKeys returns the keys in the storage
func (p *Pebble) ScanKeys(fn func(*ds.Key) bool) {
	iter, err := p.db.NewIter(nil)
	if err != nil {
		return
	}
	defer iter.Close()
	for iter.First(); iter.Valid(); iter.Next() {
		key, err := ds.DecodeKey(iter.Key())
		if err != nil {
			continue
		}
		if !fn(key) {
			break
		}
	}
}
