//go:build !wasm

package storage

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/cockroachdb/pebble/v2"
	"github.com/diiyw/nodis/ds"
	"path/filepath"
	"sync"
	"time"
)

var ErrUnsupportedFormat = errors.New("nodis: unsupported Pebble data format; use a new data directory")
var formatKey = []byte{0}
var formatVersion = []byte("NDR1")

type Pebble struct {
	sync.RWMutex
	path    string
	options *pebble.Options
	db      *pebble.DB
}

func NewPebble(path string, options *pebble.Options) *Pebble {
	return &Pebble{path: path, options: options}
}
func (p *Pebble) Init() error {
	p.Lock()
	defer p.Unlock()
	if p.db != nil {
		return errors.New("nodis: storage already open")
	}
	db, err := pebble.Open(p.path, p.options)
	if err != nil {
		return err
	}
	fail := func(err error) error { _ = db.Close(); return err }
	version, closer, err := db.Get(formatKey)
	if err == nil {
		valid := bytes.Equal(version, formatVersion)
		_ = closer.Close()
		if !valid {
			return fail(ErrUnsupportedFormat)
		}
	} else if errors.Is(err, pebble.ErrNotFound) {
		iter, e := db.NewIter(nil)
		if e != nil {
			return fail(e)
		}
		nonempty := iter.First()
		e = iter.Error()
		_ = iter.Close()
		if e != nil {
			return fail(e)
		}
		if nonempty {
			return fail(ErrUnsupportedFormat)
		}
		if e = db.Set(formatKey, formatVersion, pebble.Sync); e != nil {
			return fail(e)
		}
	} else {
		return fail(err)
	}
	p.db = db
	return nil
}
func diskKey(name string) []byte {
	k := make([]byte, 1+len(name))
	k[0] = 1
	copy(k[1:], name)
	return k
}
func encodeRecord(k ds.Key, v ds.Value) []byte {
	data := make([]byte, 12)
	copy(data, "NDR1")
	binary.LittleEndian.PutUint64(data[4:], uint64(k.Expiration))
	return append(data, EncodeValue(v)...)
}
func decodeRecord(b []byte) (int64, []byte, error) {
	if len(b) < 13 || string(b[:4]) != "NDR1" {
		return 0, nil, ErrCorruptedData
	}
	return int64(binary.LittleEndian.Uint64(b[4:])), b[12:], nil
}
func (p *Pebble) Get(k *ds.Key) (ds.Value, error) {
	p.RLock()
	defer p.RUnlock()
	if p.db == nil {
		return nil, errors.New("nodis: storage closed")
	}
	data, closer, err := p.db.Get(diskKey(k.Name))
	if errors.Is(err, pebble.ErrNotFound) {
		return nil, ErrKeyNotFound
	}
	if err != nil {
		return nil, err
	}
	defer closer.Close()
	_, value, err := decodeRecord(data)
	if err != nil {
		return nil, err
	}
	return DecodeValue(bytes.Clone(value))
}
func (p *Pebble) Set(k *ds.Key, v ds.Value) error {
	return p.Apply([]Mutation{{Key: *k, Value: v}}, false)
}
func (p *Pebble) Delete(k *ds.Key) error { return p.Apply([]Mutation{{Key: *k, Delete: true}}, false) }
func (p *Pebble) Clear() error           { return p.Apply(nil, true) }
func (p *Pebble) Apply(changes []Mutation, reset bool) error {
	p.RLock()
	defer p.RUnlock()
	if p.db == nil {
		return errors.New("nodis: storage closed")
	}
	batch := p.db.NewBatch()
	defer batch.Close()
	if reset {
		if err := batch.DeleteRange([]byte{1}, []byte{2}, nil); err != nil {
			return err
		}
	}
	for _, change := range changes {
		var err error
		if change.Delete {
			err = batch.Delete(diskKey(change.Key.Name), nil)
		} else {
			err = batch.Set(diskKey(change.Key.Name), encodeRecord(change.Key, change.Value), nil)
		}
		if err != nil {
			return err
		}
	}
	return batch.Commit(pebble.Sync)
}
func (p *Pebble) Close() error {
	p.Lock()
	defer p.Unlock()
	if p.db == nil {
		return nil
	}
	err := p.db.Close()
	if err == nil {
		p.db = nil
	}
	return err
}
func (p *Pebble) Snapshot() error {
	p.RLock()
	defer p.RUnlock()
	if p.db == nil {
		return errors.New("nodis: storage closed")
	}
	return p.db.Checkpoint(filepath.Join(p.path+"-snapshots", fmt.Sprint(time.Now().UnixNano())))
}
func (p *Pebble) ScanKeys(fn func(*ds.Key) bool) error {
	p.RLock()
	defer p.RUnlock()
	if p.db == nil {
		return errors.New("nodis: storage closed")
	}
	iter, err := p.db.NewIter(&pebble.IterOptions{LowerBound: []byte{1}, UpperBound: []byte{2}})
	if err != nil {
		return err
	}
	defer iter.Close()
	for iter.First(); iter.Valid(); iter.Next() {
		expiration, _, err := decodeRecord(iter.Value())
		if err != nil {
			return err
		}
		key := ds.Key{Name: string(iter.Key()[1:]), Expiration: expiration}
		if !fn(&key) {
			break
		}
	}
	return iter.Error()
}
