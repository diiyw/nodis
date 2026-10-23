package client

import (
	"sync"

	"github.com/diiyw/nodis"
)

// SyncStatus describes the current materialized view. Sequence is a watermark of
// received events, not a promise to have seen every database transaction.
type SyncStatus struct {
	Enabled            bool
	Ready              bool
	Sequence           uint64
	PermissionRevision uint64
	Err                error
}

type viewCache struct {
	mu                 sync.RWMutex
	objects            map[key]*nodis.PortableObject
	pending            map[key]*nodis.PortableObject
	status             SyncStatus
	closed             bool
	snapshotStarted    bool
	snapshotSequence   uint64
	snapshotPermission uint64
}

func (c *viewCache) init(enabled bool) {
	c.status.Enabled = enabled
	if enabled {
		c.objects = make(map[key]*nodis.PortableObject)
		c.pending = make(map[key]*nodis.PortableObject)
	}
}

func (c *viewCache) invalidate(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	c.invalidateLocked(err)
}

func (c *viewCache) invalidateLocked(err error) {
	clear(c.objects)
	clear(c.pending)
	c.status.Ready = false
	c.status.Sequence, c.status.PermissionRevision = 0, 0
	c.status.Err = err
	c.snapshotStarted = false
}

func (c *viewCache) close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.invalidateLocked(nodis.ErrClosed)
	c.closed = true
}

// stage keeps a new snapshot invisible until its matching completion marker.
func (c *viewCache) stage(event nodis.Event) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nodis.ErrClosed
	}
	if c.status.Ready || !c.status.Enabled {
		return nodis.ErrInvalidState
	}
	if event.Type != nodis.EventSnapshotObject && event.Type != nodis.EventSnapshotDone {
		return nodis.ErrCorruptData
	}
	if !c.snapshotStarted {
		c.snapshotSequence, c.snapshotPermission = event.Sequence, event.PermissionRevision
		c.snapshotStarted = true
	}
	if event.Sequence != c.snapshotSequence || event.PermissionRevision != c.snapshotPermission {
		return nodis.ErrCorruptData
	}
	if event.Type == nodis.EventSnapshotDone {
		c.objects, c.pending = c.pending, c.objects
		clear(c.pending)
		c.status = SyncStatus{Enabled: true, Ready: true, Sequence: event.Sequence, PermissionRevision: event.PermissionRevision}
		return nil
	}
	k := key{event.Document, event.Object}
	if _, exists := c.pending[k]; exists {
		return nodis.ErrCorruptData
	}
	object, err := ownedObject(event)
	if err != nil {
		return err
	}
	c.pending[k] = object
	return nil
}

func ownedObject(event nodis.Event) (*nodis.PortableObject, error) {
	// Network readers reuse their frame buffer. Cached objects must own bytes.
	object, err := nodis.DecodePortable(append([]byte(nil), event.Data...))
	if err != nil {
		return nil, err
	}
	if event.Document == 0 || event.Object == 0 || object.ID() != event.Object || object.Revision() == 0 || object.Revision() != event.Revision {
		return nil, nodis.ErrCorruptData
	}
	return object, nil
}

func (c *viewCache) apply(event nodis.Event) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nodis.ErrClosed
	}
	if !c.status.Ready {
		return nodis.ErrInvalidState
	}
	switch event.Type {
	case nodis.EventObjectCreated, nodis.EventObjectUpdated, nodis.EventObjectDeleted, nodis.EventDocumentCreated, nodis.EventDocumentUpdated, nodis.EventDocumentDeleted:
		if event.Sequence < c.status.Sequence || event.PermissionRevision != c.status.PermissionRevision {
			return nodis.ErrCorruptData
		}
	}
	switch event.Type {
	case nodis.EventObjectCreated, nodis.EventObjectUpdated:
		object, err := ownedObject(event)
		if err != nil {
			return err
		}
		c.objects[key{event.Document, event.Object}] = object
	case nodis.EventObjectDeleted:
		delete(c.objects, key{event.Document, event.Object})
	case nodis.EventDocumentDeleted:
		for k := range c.objects {
			if k.document == event.Document {
				delete(c.objects, k)
			}
		}
	case nodis.EventDocumentCreated, nodis.EventDocumentUpdated, nodis.EventDocumentRead, nodis.EventObjectRead:
	default:
		return nodis.ErrCorruptData
	}
	if event.Sequence > c.status.Sequence {
		c.status.Sequence = event.Sequence
	}
	return nil
}

func (c *viewCache) cached(document nodis.DocumentID, object nodis.ObjectID) (*nodis.PortableObject, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.status.Ready {
		return nil, false
	}
	value, ok := c.objects[key{document, object}]
	return value, ok
}

func (c *viewCache) syncStatus() SyncStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.status
}
