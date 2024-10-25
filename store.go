package nodis

import (
	"log"
	"sync"
	"time"

	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/storage"

	"github.com/diiyw/nodis/ds/list"
	"github.com/diiyw/nodis/redis"
	"github.com/tidwall/btree"
)

type store struct {
	mu          sync.RWMutex
	metadata    btree.Map[string, *metadata]
	ss          storage.Storage
	closed      bool
	watchMu     sync.RWMutex
	watchedKeys btree.Map[string, *list.LinkedListG[*redis.Conn]]
}

func newStore(ss storage.Storage) *store {
	s := &store{ss: ss}
	err := s.ss.Init()
	if err != nil {
		log.Fatal(err)
	}
	s.ss.ScanKeys(func(key *ds.Key) bool {
		var m = newMetadata()
		m.key = key
		m.state |= KeyStateNormal
		s.metadata.Set(key.Name, m)
		return true
	})
	return s
}

// flush changed keys to storage
func (s *store) flush() {
	s.mu.Lock()
	defer s.mu.Unlock()
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
		err := s.ss.Put(m.key, m.value)
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
			err := s.ss.Put(m.key, m.value)
			if err != nil {
				log.Println("GC: ", err)
			}
		}
		m.reset()
		if m.count < 0 {
			m.removeFromMemory()
		}
		return true
	})
}

// close the store
func (s *store) close() error {
	s.closed = true
	s.flush()
	return s.ss.Close()
}

// clear the store
func (s *store) clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.metadata.Clear()
	return s.ss.Clear()
}
