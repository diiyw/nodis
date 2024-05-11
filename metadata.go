package nodis

import (
	"sync"
	"time"

	"github.com/diiyw/nodis/ds"
)

type metadata struct {
	*sync.RWMutex
	key       *Key
	ds        ds.DataStruct
	ok        bool
	writeable bool
}

func (m *metadata) set(key *Key, d ds.DataStruct) *metadata {
	m.key = key
	m.ds = d
	if m.key.dataType == 0 {
		m.key.dataType = m.ds.Type()
	}
	m.ok = true
	return m
}

func (m *metadata) isOk() bool {
	return m.ok
}

func (m *metadata) empty() *metadata {
	m.ds = nil
	m.key = nil
	m.ok = false
	return m
}

func (m *metadata) signalModifiedKey() {
	if m.key.state&KeyStateWatching == KeyStateWatching {
		if m.key.state&KeyStateModified == KeyStateModified {
			m.key.state |= KeyStateWatchAfterModified
		} else {
			m.key.state |= KeyStateWatchBeforeModified
		}
	}
	m.key.state |= KeyStateModified
	m.key.modifiedTime = time.Now().Unix()
}

func (m *metadata) watchModified() bool {
	return m.key.state&KeyStateWatchBeforeModified == KeyStateWatchBeforeModified || m.key.state&KeyStateWatchAfterModified == KeyStateWatchAfterModified
}

func (m *metadata) commit() {
	if m.RWMutex == nil {
		// emptyMetadata
		return
	}
	if m.writeable {
		m.empty()
		m.writeable = false
		m.Unlock()
		return
	}
	m.RUnlock()
}
