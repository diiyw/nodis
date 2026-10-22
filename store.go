package nodis

import (
	"errors"
	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/patch"
	"github.com/diiyw/nodis/redis"
	"github.com/diiyw/nodis/storage"
	"sync"
	"time"
)

var ErrClosed = errors.New("nodis: database is closed")

type store struct {
	tx           Tx
	mu           sync.RWMutex
	metadata     map[string]*metadata
	dirty        map[string]struct{}
	ss           storage.Storage
	closed       bool
	deleted      map[string]struct{}
	changed      map[string]struct{}
	clearPending bool
	events       []patch.Op
	mutations    []storage.Mutation
	watchedKeys  map[string]map[*redis.Conn]struct{}
	gcKeys       []string
	gcPositions  map[string]int
	gcCursor     int
}

const gcBatchSize = 256

func newStore(ss storage.Storage) (*store, error) {
	s := &store{ss: ss, metadata: make(map[string]*metadata), dirty: make(map[string]struct{}), deleted: make(map[string]struct{}), changed: make(map[string]struct{}), watchedKeys: make(map[string]map[*redis.Conn]struct{}), gcPositions: make(map[string]int)}
	s.tx.store = s
	if err := ss.Init(); err != nil {
		return nil, err
	}
	err := ss.ScanKeys(func(k *ds.Key) bool {
		key := *k
		m := newMetadata(&key, false)
		m.state = KeyStateNormal
		s.metadata[k.Name] = m
		s.addGCKey(k.Name)
		return true
	})
	if err != nil {
		_ = ss.Close()
		return nil, err
	}
	return s, nil
}

func (s *store) addGCKey(key string) {
	if _, exists := s.gcPositions[key]; exists {
		return
	}
	s.gcPositions[key] = len(s.gcKeys)
	s.gcKeys = append(s.gcKeys, key)
}

func (s *store) removeGCKey(key string) {
	if index, exists := s.gcPositions[key]; exists {
		s.gcKeys[index] = ""
		delete(s.gcPositions, key)
	}
}

func (s *store) compactGCKeys() {
	keys := s.gcKeys[:0]
	clear(s.gcPositions)
	for _, key := range s.gcKeys {
		if key == "" || s.metadata[key] == nil {
			continue
		}
		s.gcPositions[key] = len(keys)
		keys = append(keys, key)
	}
	clear(s.gcKeys[len(keys):])
	s.gcKeys = keys
	s.gcCursor = 0
}
func (s *store) invalidate(key string) {
	for c := range s.watchedKeys[key] {
		c.WatchKeys[key] = true
	}
}

// flushLocked is the durability boundary. Dirty state and tombstones are retained
// until every mutation succeeds. Built-in backends commit the batch atomically.
func (s *store) flushLocked() error {
	if s.closed {
		return ErrClosed
	}
	s.mutations = s.mutations[:0]
	defer func() { clear(s.mutations); s.mutations = s.mutations[:0] }()
	now := time.Now().UnixMilli()
	for key := range s.deleted {
		s.mutations = append(s.mutations, storage.Mutation{Key: ds.Key{Name: key}, Delete: true})
	}
	for key := range s.dirty {
		m := s.metadata[key]
		if m == nil {
			continue
		}
		if m.expired(now) {
			if _, alreadyDeleted := s.deleted[key]; !alreadyDeleted {
				s.mutations = append(s.mutations, storage.Mutation{Key: *m.key, Delete: true})
			}
			continue
		}
		if m.modified() {
			s.mutations = append(s.mutations, storage.Mutation{Key: *m.key, Value: m.value})
		}
	}
	if len(s.mutations) == 0 && !s.clearPending {
		return nil
	}
	if batch, ok := s.ss.(storage.BatchStorage); ok {
		if err := batch.Apply(s.mutations, s.clearPending); err != nil {
			return err
		}
	} else {
		if s.clearPending {
			if err := s.ss.Clear(); err != nil {
				return err
			}
		}
		for _, mutation := range s.mutations {
			var err error
			if mutation.Delete {
				err = s.ss.Delete(&mutation.Key)
			} else {
				err = s.ss.Set(&mutation.Key, mutation.Value)
			}
			if err != nil {
				return err
			}
		}
	}
	clear(s.deleted)
	for key := range s.dirty {
		if m := s.metadata[key]; m != nil {
			m.state = KeyStateNormal
		}
	}
	clear(s.dirty)
	s.clearPending = false
	return nil
}
func (s *store) flush() error { s.mu.Lock(); defer s.mu.Unlock(); return s.flushLocked() }
func (s *store) gc() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.flushLocked(); err != nil {
		return err
	}
	if s.gcCursor >= len(s.gcKeys) {
		s.compactGCKeys()
	}
	limit := min(gcBatchSize, len(s.gcKeys)-s.gcCursor)
	for processed := 0; processed < limit; processed++ {
		key := s.gcKeys[s.gcCursor]
		s.gcCursor++
		if key == "" {
			continue
		}
		m := s.metadata[key]
		if m == nil {
			continue
		}
		if m.expired(time.Now().UnixMilli()) {
			delete(s.metadata, key)
			s.removeGCKey(key)
			s.deleted[key] = struct{}{}
			s.invalidate(key)
			continue
		}
		if !m.accessed.Swap(false) {
			m.removeFromMemory()
		}
	}
	if s.gcCursor >= len(s.gcKeys) {
		s.compactGCKeys()
	}
	return nil
}
