package nodis

import (
	"encoding/binary"
	"sync"
)

type EventType uint8

const (
	EventDocumentCreated EventType = iota + 1
	EventDocumentRead
	EventDocumentUpdated
	EventDocumentDeleted
	EventObjectCreated
	EventObjectRead
	EventObjectUpdated
	EventObjectDeleted
	EventPolicyChanged
	EventSnapshotObject
	EventSnapshotDone
	EventReset
)

type Event struct {
	Type               EventType
	Sequence           uint64
	PermissionRevision uint64
	Document           DocumentID
	Object             ObjectID
	Revision           uint64
	Data               []byte
}

type listener struct {
	id        uint64
	principal Principal
	scope     SyncScope
	ch        chan Event
}

type SyncScope struct {
	Documents []DocumentID
}

func (s SyncScope) includes(document DocumentID) bool {
	if len(s.Documents) == 0 {
		return true
	}
	for i := range s.Documents {
		if s.Documents[i] == document {
			return true
		}
	}
	return false
}

type Subscription struct {
	db   *DB
	id   uint64
	C    <-chan Event
	once sync.Once
}

func (s *Subscription) Close() {
	if s == nil {
		return
	}
	s.once.Do(func() {
		if s.db != nil {
			s.db.removeListener(s.id)
		}
	})
}

func (db *DB) Subscribe(principal Principal, scope SyncScope) (*Subscription, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.closed {
		return nil, ErrClosed
	}
	for _, document := range scope.Documents {
		resource := Resource{Scope: ScopeDocument, Document: document}
		if !db.allowedLocked(principal, resource, ActionDiscover) || !db.allowedLocked(principal, resource, ActionRead) || !db.allowedLocked(principal, resource, ActionSync) {
			return nil, ErrPermission
		}
	}
	visible := make([]Event, 0)
	for key, data := range db.objects {
		if !scope.includes(key.document) {
			continue
		}
		resource := Resource{Scope: ScopeObject, Document: key.document, Object: key.object}
		if !db.allowedLocked(principal, resource, ActionSync) || !db.allowedLocked(principal, resource, ActionRead) {
			continue
		}
		doc := db.documents[key.document]
		filtered := filterObject(doc, data, func(field FieldID) bool {
			fieldResource := Resource{Scope: ScopeField, Document: key.document, Object: key.object, Field: field}
			return db.allowedLocked(principal, fieldResource, ActionSync) && db.allowedLocked(principal, fieldResource, ActionRead)
		})
		if filtered == nil {
			return nil, ErrCorruptData
		}
		filtered = exportObject(doc, filtered)
		if filtered == nil {
			return nil, ErrCorruptData
		}
		visible = append(visible, Event{Type: EventSnapshotObject, Sequence: db.sequence, PermissionRevision: db.permissionRevision, Document: key.document, Object: key.object, Revision: objectRevision(data), Data: filtered})
	}
	capacity := db.eventQueue
	if capacity < len(visible)+1 {
		capacity = len(visible) + 1
	}
	ch := make(chan Event, capacity)
	for i := range visible {
		ch <- visible[i]
	}
	ch <- Event{Type: EventSnapshotDone, Sequence: db.sequence, PermissionRevision: db.permissionRevision}
	db.nextListener++
	id := db.nextListener
	db.listeners[id] = &listener{id: id, principal: principal, scope: scope, ch: ch}
	return &Subscription{db: db, id: id, C: ch}, nil
}

func objectRevision(data []byte) uint64 {
	if len(data) < objectHeaderSize {
		return 0
	}
	return binary.LittleEndian.Uint64(data[14:22])
}

func (db *DB) listenersLocked() []uint64 {
	if len(db.listeners) == 0 {
		return nil
	}
	ids := make([]uint64, 0, len(db.listeners))
	for id := range db.listeners {
		ids = append(ids, id)
	}
	return ids
}

func (db *DB) removeListener(id uint64) {
	db.mu.Lock()
	if listener, ok := db.listeners[id]; ok {
		delete(db.listeners, id)
		close(listener.ch)
	}
	db.mu.Unlock()
}

func (db *DB) publish(ids []uint64, event Event) {
	for _, id := range ids {
		db.mu.Lock()
		listener, ok := db.listeners[id]
		if !ok {
			db.mu.Unlock()
			continue
		}
		filtered, deliver := db.eventForListenerLocked(listener, event)
		if deliver {
			select {
			case listener.ch <- filtered:
			default:
				delete(db.listeners, id)
				close(listener.ch)
			}
		}
		db.mu.Unlock()
	}
}

// publishLocked queues a committed event while the commit lock is still held,
// preserving WAL sequence order across concurrent writers.
func (db *DB) publishLocked(event Event) {
	for id, listener := range db.listeners {
		filtered, deliver := db.eventForListenerLocked(listener, event)
		if !deliver {
			continue
		}
		select {
		case listener.ch <- filtered:
		default:
			delete(db.listeners, id)
			close(listener.ch)
		}
	}
}

func (db *DB) eventForListenerLocked(listener *listener, event Event) (Event, bool) {
	if event.Type == EventPolicyChanged {
		return Event{Type: EventReset, Sequence: event.Sequence, PermissionRevision: event.PermissionRevision}, true
	}
	if event.Document == 0 || !listener.scope.includes(event.Document) {
		return Event{}, false
	}
	if event.Type == EventDocumentCreated || event.Type == EventDocumentRead || event.Type == EventDocumentUpdated || event.Type == EventDocumentDeleted {
		resource := Resource{Scope: ScopeDocument, Document: event.Document}
		return event, db.allowedLocked(listener.principal, resource, ActionSync)
	}
	resource := Resource{Scope: ScopeObject, Document: event.Document, Object: event.Object}
	if !db.allowedLocked(listener.principal, resource, ActionSync) {
		return Event{}, false
	}
	if event.Type == EventObjectDeleted {
		return event, true
	}
	if event.Type == EventObjectRead {
		return event, db.allowedLocked(listener.principal, resource, ActionRead)
	}
	if !db.allowedLocked(listener.principal, resource, ActionRead) {
		return Event{}, false
	}
	doc := db.documents[event.Document]
	event.Data = filterObject(doc, event.Data, func(field FieldID) bool {
		fieldResource := Resource{Scope: ScopeField, Document: event.Document, Object: event.Object, Field: field}
		return db.allowedLocked(listener.principal, fieldResource, ActionSync) && db.allowedLocked(listener.principal, fieldResource, ActionRead)
	})
	if event.Data == nil {
		return Event{}, false
	}
	event.Data = exportObject(doc, event.Data)
	return event, event.Data != nil
}
