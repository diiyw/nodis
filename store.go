package nodis

import (
	"log"
	"sync"
	"time"

	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/storage"

	"github.com/diiyw/nodis/ds/list"
	"github.com/diiyw/nodis/redis"
)

type store struct {
	mu          sync.RWMutex
	metadata    map[string]*metadata
	ss          storage.Storage
	closed      bool
	watchMu     sync.RWMutex
	watchedKeys map[string]*list.LinkedListG[*redis.Conn]
}

func newStore(ss storage.Storage) *store {
	s := &store{
		ss:          ss,
		metadata:    make(map[string]*metadata),
		watchedKeys: make(map[string]*list.LinkedListG[*redis.Conn]),
	}
	err := s.ss.Init()
	if err != nil {
		log.Fatal(err)
	}
	s.ss.ScanKeys(func(key *ds.Key) bool {
		var m = newMetadata(key, false)
		m.state |= KeyStateNormal
		s.metadata[key.Name] = m
		return true
	})
	return s
}

// flush changed keys to storage
func (s *store) flush() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UnixMilli()
	for _, m := range s.metadata {
		m.Lock()
		defer m.Unlock()
		if !m.modified() || m.expired(now) || !m.isOk() {
			continue
		}
		if m.value == nil {
			continue
		}
		// save to storage
		err := s.ss.Set(m.key, m.value)
		if err != nil {
			log.Println("Flush changes: ", err)
		}
	}
}

// gc removes expired and unused keys
func (s *store) gc() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	now := time.Now().UnixMilli()
	for key, m := range s.metadata {
		m.Lock()
		defer m.Unlock()
		if m.expired(now) || !m.isOk() {
			delete(s.metadata, key)
			continue
		}
		if m.modified() {
			err := s.ss.Set(m.key, m.value)
			if err != nil {
				log.Println("GC: ", err)
			}
		}
		m.reset()
		if m.count.Load() < 0 {
			m.removeFromMemory()
		}
	}
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
	clear(s.metadata)
	return s.ss.Clear()
}
