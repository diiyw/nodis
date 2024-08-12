package nodis

import (
	"github.com/diiyw/nodis/storage"
	"log"
	"sync"
	"time"

	"github.com/diiyw/nodis/ds/list"
	"github.com/diiyw/nodis/redis"
	"github.com/tidwall/btree"
)

type store struct {
	mu          sync.RWMutex
	metadata    btree.Map[string, *metadata]
	sg          storage.Storage
	closed      bool
	watchMu     sync.RWMutex
	watchedKeys btree.Map[string, *list.LinkedListG[*redis.Conn]]
}

func newStore(sg storage.Storage) *store {
	s := &store{sg: sg}
	s.sg.ScanKeys(func(key *ds.Key) bool {
		var m = newMetadata()
		m.key = key
		m.state |= KeyStateNormal
		s.metadata.Set(key.Name, m)
		return true
	})
	return s
}

// sync flush changed keys to disk
func (s *store) sync() {
	now := time.Now().UnixMilli()
	s.metadata.Scan(func(key string, m *metadata) bool {
		m.Lock()
		defer m.Unlock()
		if !m.modified() || m.expired(now) || !m.isOk() {
			return true
		}
		if m.value == nil {
			return true
		}
		// save to storage
		err := s.sg.Put(m.key, m.value)
		if err != nil {
			log.Println("Flush changes: ", err)
		}
		return true
	})
}

// gc removes expired and unused keys
func (s *store) gc() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	now := time.Now().UnixMilli()
	s.metadata.Scan(func(key string, m *metadata) bool {
		m.Lock()
		defer m.Unlock()
		if m.expired(now) || !m.isOk() {
			s.metadata.Delete(key)
			return true
		}
		if m.modified() {
			// sync to disk
			err := s.sg.Put(m.key, m.value)
			if err != nil {
				log.Println("GC: ", err)
			}
		}
		m.reset()
		if m.count < 0 {
			s.metadata.Delete(key)
		}
		return true
	})
}

// close the store
func (s *store) close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sync()
	s.closed = true
	return s.sg.Close()
}

// clear the store
func (s *store) clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.metadata.Clear()
	return s.sg.Reset()
}
