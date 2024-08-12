package storage

import (
	"github.com/cockroachdb/pebble"
	"github.com/diiyw/nodis/ds"
)

type Pebble struct {
	db *pebble.DB
}

func (p *Pebble) Open(path string) error {
	db, err := pebble.Open(path, &pebble.Options{})
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
	entry, err := parseValueEntry(v)
	if err != nil {
		return nil, err
	}
	return parseValue(entry.Value)
}

// Put the value to the storage
func (p *Pebble) Put(key *ds.Key, value ds.Value) error {
	entry := NewValueEntry(key.Name, value, key.Expiration)
	data := entry.encode()
	return p.db.Set([]byte(key.Name), data, pebble.Sync)
}

// Delete the value from the storage
func (p *Pebble) Delete(key string) error {
	return p.db.Delete([]byte(key), pebble.Sync)
}

// Reset the storage
func (p *Pebble) Reset() error {
	return p.db.Flush()
}

// Close the storage
func (p *Pebble) Close() error {
	return p.db.Close()
}

// Snapshot the storage
func (p *Pebble) Snapshot() error {
	return nil
}

// ScanKeys returns the keys in the storage
func (p *Pebble) ScanKeys(fn func(*ds.Key) bool) {
	iter, err := p.db.NewIter(nil)
	if err != nil {
		return
	}
	defer iter.Close()
	for iter.First(); iter.Valid(); iter.Next() {
		v, err := iter.ValueAndErr()
		if err != nil {
			continue
		}
		entry, err := parseValueEntry(v)
		if err != nil {
			continue
		}
		key := &ds.Key{
			Name:       entry.Key,
			Expiration: entry.Expiration,
		}
		if !fn(key) {
			break
		}
	}
}
