package nodis

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/diiyw/nodis/storage"
)

type Options struct {
	Dir   string
	Admin string
	// Deprecated: persistent databases now sync every commit by default.
	Durable bool
	// UnsafeNoSync skips commit-time WAL syncs. A machine crash can lose commits.
	// Flush and Close still sync. Cannot be combined with Durable.
	UnsafeNoSync bool
	EventQueue   int
	// CheckpointInterval compacts storage periodically; zero disables it.
	CheckpointInterval time.Duration
}

type DB struct {
	mu                 sync.RWMutex
	log                *storage.Log
	durable            bool
	admin              string
	closed             bool
	sequence           uint64
	permissionRevision uint64
	nextDocument       DocumentID
	nextObject         ObjectID
	documents          map[DocumentID]*Document
	documentsByName    map[string]DocumentID
	objects            map[objectKey][]byte
	rules              []Rule
	pipelines          map[PipelineID]Pipeline
	listeners          map[uint64]*listener
	nextListener       uint64
	eventQueue         int
	snapshotBuffer     []byte
	snapshotDocuments  []DocumentID
	snapshotObjects    []objectKey
	checkpointStop     chan struct{}
	checkpointDone     chan struct{}
	checkpointErr      error
}

func Open(options Options) (*DB, error) {
	if options.Durable && options.UnsafeNoSync {
		return nil, ErrInvalidValue
	}
	if options.CheckpointInterval < 0 {
		return nil, ErrInvalidValue
	}
	if options.CheckpointInterval > 0 && options.Dir == "" {
		return nil, storage.ErrMemoryOnly
	}
	log, err := storage.Open(options.Dir)
	if err != nil {
		return nil, err
	}
	if options.Admin == "" {
		options.Admin = "admin"
	}
	if options.EventQueue <= 0 {
		options.EventQueue = 256
	}
	db := &DB{
		log: log, durable: !options.UnsafeNoSync, admin: options.Admin,
		documents: make(map[DocumentID]*Document), documentsByName: make(map[string]DocumentID),
		objects: make(map[objectKey][]byte), pipelines: make(map[PipelineID]Pipeline),
		listeners: make(map[uint64]*listener), eventQueue: options.EventQueue,
	}
	restore := snapshotRestore{db: db}
	sequence, found, err := log.RestoreSnapshot(restore.apply)
	if err != nil || found && !restore.metadata {
		_ = log.Close()
		if err == nil {
			err = ErrCorruptData
		}
		return nil, err
	}
	if found {
		db.sequence, db.permissionRevision = sequence, restore.permissionRevision
	}
	if err := log.Replay(func(sequence uint64, payload []byte) error {
		if found && sequence <= db.sequence {
			return nil
		}
		return db.replay(sequence, payload)
	}); err != nil {
		_ = log.Close()
		return nil, err
	}
	if err := db.rebuildIndexes(); err != nil {
		_ = log.Close()
		return nil, err
	}
	if options.CheckpointInterval > 0 {
		db.checkpointStop, db.checkpointDone = make(chan struct{}), make(chan struct{})
		go db.checkpointLoop(options.CheckpointInterval)
	}
	return db, nil
}

func (db *DB) Close() error {
	db.mu.Lock()
	if db.closed {
		db.mu.Unlock()
		return nil
	}
	db.closed = true
	if db.checkpointStop != nil {
		close(db.checkpointStop)
	}
	for id, listener := range db.listeners {
		close(listener.ch)
		delete(db.listeners, id)
	}
	db.mu.Unlock()
	if db.checkpointDone != nil {
		<-db.checkpointDone
	}
	return db.log.Close()
}

func (db *DB) Flush() error {
	db.mu.RLock()
	defer db.mu.RUnlock()
	if db.closed {
		return ErrClosed
	}
	return db.log.Sync()
}

func (db *DB) CreateDocument(actor Principal, name string) (*Document, error) {
	if !validName(name) {
		return nil, ErrInvalidName
	}
	db.mu.Lock()
	if db.closed {
		db.mu.Unlock()
		return nil, ErrClosed
	}
	if !db.allowedLocked(actor, Resource{Scope: ScopeDatabase}, ActionCreate) {
		db.mu.Unlock()
		return nil, ErrPermission
	}
	if _, ok := db.documentsByName[name]; ok {
		db.mu.Unlock()
		return nil, ErrAlreadyExists
	}
	id := db.nextDocument + 1
	next := db.sequence + 1
	payload := appendDocumentRecord(nil, id, name)
	if err := db.log.Append(next, payload, db.durable); err != nil {
		db.mu.Unlock()
		return nil, err
	}
	doc := newDocument(db, id, name)
	db.nextDocument, db.sequence = id, next
	db.documents[id], db.documentsByName[name] = doc, id
	event := Event{Type: EventDocumentCreated, Sequence: next, PermissionRevision: db.permissionRevision, Document: id}
	db.publishLocked(event)
	db.mu.Unlock()
	return doc, nil
}

func newDocument(db *DB, id DocumentID, name string) *Document {
	return &Document{db: db, id: id, name: name, fieldsByName: make(map[string]FieldID), fieldNames: make(map[FieldID]string), indexes: make(map[FieldID]*orderedIndex[scalarKey])}
}

func (db *DB) Document(actor Principal, name string) (*Document, error) {
	db.mu.RLock()
	if db.closed {
		db.mu.RUnlock()
		return nil, ErrClosed
	}
	id, ok := db.documentsByName[name]
	if !ok {
		db.mu.RUnlock()
		return nil, ErrNotFound
	}
	doc := db.documents[id]
	resource := Resource{Scope: ScopeDocument, Document: id}
	if !db.allowedLocked(actor, resource, ActionDiscover) || !db.allowedLocked(actor, resource, ActionRead) {
		db.mu.RUnlock()
		return nil, ErrPermission
	}
	listeners := db.listenersLocked()
	event := Event{Type: EventDocumentRead, Sequence: db.sequence, PermissionRevision: db.permissionRevision, Document: id}
	db.mu.RUnlock()
	db.publish(listeners, event)
	return doc, nil
}

func (db *DB) DocumentByID(actor Principal, id DocumentID) (*Document, error) {
	db.mu.RLock()
	if db.closed {
		db.mu.RUnlock()
		return nil, ErrClosed
	}
	doc, ok := db.documents[id]
	if !ok {
		db.mu.RUnlock()
		return nil, ErrNotFound
	}
	resource := Resource{Scope: ScopeDocument, Document: id}
	if !db.allowedLocked(actor, resource, ActionDiscover) || !db.allowedLocked(actor, resource, ActionRead) {
		db.mu.RUnlock()
		return nil, ErrPermission
	}
	listeners := db.listenersLocked()
	event := Event{Type: EventDocumentRead, Sequence: db.sequence, PermissionRevision: db.permissionRevision, Document: id}
	db.mu.RUnlock()
	db.publish(listeners, event)
	return doc, nil
}

func (db *DB) Get(actor Principal, document DocumentID, object ObjectID) (*ObjectView, error) {
	db.mu.RLock()
	if db.closed {
		db.mu.RUnlock()
		return nil, ErrClosed
	}
	doc, ok := db.documents[document]
	if !ok {
		db.mu.RUnlock()
		return nil, ErrNotFound
	}
	resource := Resource{Scope: ScopeObject, Document: document, Object: object}
	if !db.allowedLocked(actor, resource, ActionRead) {
		db.mu.RUnlock()
		return nil, ErrPermission
	}
	data, ok := db.objects[objectKey{document: document, object: object}]
	if !ok {
		db.mu.RUnlock()
		return nil, ErrNotFound
	}
	filtered := filterObject(doc, data, func(field FieldID) bool {
		return db.allowedLocked(actor, Resource{Scope: ScopeField, Document: document, Object: object, Field: field}, ActionRead)
	})
	if filtered == nil {
		db.mu.RUnlock()
		return nil, ErrCorruptData
	}
	listeners := db.listenersLocked()
	event := Event{Type: EventObjectRead, Sequence: db.sequence, PermissionRevision: db.permissionRevision, Document: document, Object: object, Revision: objectRevision(data)}
	db.mu.RUnlock()
	db.publish(listeners, event)
	return &ObjectView{document: doc, data: filtered}, nil
}

func (db *DB) RenameDocument(actor Principal, id DocumentID, name string) error {
	if !validName(name) {
		return ErrInvalidName
	}
	db.mu.Lock()
	if db.closed {
		db.mu.Unlock()
		return ErrClosed
	}
	doc, ok := db.documents[id]
	if !ok {
		db.mu.Unlock()
		return ErrNotFound
	}
	if !db.allowedLocked(actor, Resource{Scope: ScopeDocument, Document: id}, ActionUpdate) {
		db.mu.Unlock()
		return ErrPermission
	}
	if existing, exists := db.documentsByName[name]; exists && existing != id {
		db.mu.Unlock()
		return ErrAlreadyExists
	}
	next := db.sequence + 1
	if err := db.log.Append(next, appendDocumentUpdateRecord(nil, id, name, doc.defaultPipeline), db.durable); err != nil {
		db.mu.Unlock()
		return err
	}
	delete(db.documentsByName, doc.name)
	doc.name = name
	db.documentsByName[name] = id
	db.sequence = next
	event := Event{Type: EventDocumentUpdated, Sequence: next, PermissionRevision: db.permissionRevision, Document: id}
	db.publishLocked(event)
	db.mu.Unlock()
	return nil
}

func (db *DB) DeleteDocument(actor Principal, id DocumentID) error {
	db.mu.Lock()
	if db.closed {
		db.mu.Unlock()
		return ErrClosed
	}
	doc, ok := db.documents[id]
	if !ok {
		db.mu.Unlock()
		return ErrNotFound
	}
	if !db.allowedLocked(actor, Resource{Scope: ScopeDocument, Document: id}, ActionDelete) {
		db.mu.Unlock()
		return ErrPermission
	}
	next := db.sequence + 1
	if err := db.log.Append(next, appendDocumentDeleteRecord(nil, id), db.durable); err != nil {
		db.mu.Unlock()
		return err
	}
	delete(db.documents, id)
	delete(db.documentsByName, doc.name)
	doc.indexes, doc.objectIDs = nil, nil
	for key := range db.objects {
		if key.document == id {
			delete(db.objects, key)
		}
	}
	db.sequence = next
	event := Event{Type: EventDocumentDeleted, Sequence: next, PermissionRevision: db.permissionRevision, Document: id}
	db.publishLocked(event)
	db.mu.Unlock()
	return nil
}

func (db *DB) Begin(actor Principal) *Tx {
	return &Tx{db: db, principal: actor, state: txOpen, ctx: context.Background()}
}

func (db *DB) Sequence() uint64 { db.mu.RLock(); defer db.mu.RUnlock(); return db.sequence }
func (db *DB) PermissionRevision() uint64 {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.permissionRevision
}

func (db *DB) replay(sequence uint64, payload []byte) error {
	if sequence <= db.sequence {
		return errors.New("nodis: WAL sequence is not increasing")
	}
	if err := db.applyRecord(payload); err != nil {
		return err
	}
	db.sequence = sequence
	return nil
}
