package nodis

import (
	"context"
	"encoding/binary"
	"sort"
)

type txState uint8

const (
	txOpen txState = iota
	txCommitted
	txRolledBack
)

type Draft struct {
	tx               *Tx
	document         *Document
	id               ObjectID
	expectedRevision uint64
	pipeline         PipelineID
	fields           map[string]Value
	changed          map[string]struct{}
	readable         map[string]struct{}
	deleted          bool
	saved            bool
}

type Tx struct {
	db        *DB
	principal Principal
	ctx       context.Context
	state     txState
	drafts    []*Draft
}

func (tx *Tx) WithContext(ctx context.Context) *Tx {
	if ctx != nil {
		tx.ctx = ctx
	}
	return tx
}

func (tx *Tx) Create(document DocumentID) (*Draft, error) {
	if tx == nil || tx.state != txOpen {
		return nil, ErrInvalidState
	}
	tx.db.mu.Lock()
	if tx.db.closed {
		tx.db.mu.Unlock()
		return nil, ErrClosed
	}
	doc, ok := tx.db.documents[document]
	if !ok {
		tx.db.mu.Unlock()
		return nil, ErrNotFound
	}
	tx.db.nextObject++
	id := tx.db.nextObject
	pipeline := doc.defaultPipeline
	tx.db.mu.Unlock()
	draft := &Draft{tx: tx, document: doc, id: id, pipeline: pipeline, fields: make(map[string]Value), changed: make(map[string]struct{}), readable: make(map[string]struct{})}
	tx.drafts = append(tx.drafts, draft)
	return draft, nil
}

func (tx *Tx) Update(document DocumentID, object ObjectID, expectedRevision uint64) (*Draft, error) {
	if tx == nil || tx.state != txOpen {
		return nil, ErrInvalidState
	}
	tx.db.mu.RLock()
	if tx.db.closed {
		tx.db.mu.RUnlock()
		return nil, ErrClosed
	}
	doc, ok := tx.db.documents[document]
	if !ok {
		tx.db.mu.RUnlock()
		return nil, ErrNotFound
	}
	data, ok := tx.db.objects[objectKey{document: document, object: object}]
	if !ok {
		tx.db.mu.RUnlock()
		return nil, ErrNotFound
	}
	if !tx.db.allowedLocked(tx.principal, Resource{Scope: ScopeObject, Document: document, Object: object}, ActionRead) {
		tx.db.mu.RUnlock()
		return nil, ErrPermission
	}
	fields, err := decodeObjectFields(doc, data)
	if err != nil {
		tx.db.mu.RUnlock()
		return nil, err
	}
	revision := binary.LittleEndian.Uint64(data[14:22])
	pipeline := PipelineID(binary.LittleEndian.Uint64(data[22:30]))
	readable := make(map[string]struct{}, len(fields))
	for name := range fields {
		field, exists := doc.fieldID(name)
		if exists && tx.db.allowedLocked(tx.principal, Resource{Scope: ScopeField, Document: document, Object: object, Field: field}, ActionRead) {
			readable[name] = struct{}{}
		}
	}
	tx.db.mu.RUnlock()
	if expectedRevision == 0 {
		expectedRevision = revision
	}
	draft := &Draft{tx: tx, document: doc, id: object, expectedRevision: expectedRevision, pipeline: pipeline, fields: fields, changed: make(map[string]struct{}), readable: readable}
	tx.drafts = append(tx.drafts, draft)
	return draft, nil
}

func (tx *Tx) Delete(document DocumentID, object ObjectID, expectedRevision uint64) error {
	draft, err := tx.Update(document, object, expectedRevision)
	if err != nil {
		return err
	}
	draft.deleted, draft.saved = true, true
	return nil
}

func (d *Draft) ID() ObjectID {
	if d == nil {
		return 0
	}
	return d.id
}
func (d *Draft) SetPipeline(id PipelineID) {
	if d != nil {
		d.pipeline = id
	}
}

func (d *Draft) Set(name string, input any) error {
	if d == nil || d.tx.state != txOpen {
		return ErrInvalidState
	}
	if !validName(name) {
		return ErrInvalidName
	}
	value, err := ValueOf(input)
	if err != nil {
		return err
	}
	d.fields[name] = value
	d.changed[name] = struct{}{}
	d.readable[name] = struct{}{}
	return nil
}

func (d *Draft) DeleteField(name string) {
	if d == nil || d.tx.state != txOpen {
		return
	}
	delete(d.fields, name)
	d.changed[name] = struct{}{}
}

func (d *Draft) Value(name string) (Value, bool) {
	if _, ok := d.readable[name]; !ok {
		return Value{}, false
	}
	value, ok := d.fields[name]
	return value, ok
}

func (d *Draft) Save() (ObjectID, error) {
	if d == nil || d.tx.state != txOpen || d.deleted {
		return 0, ErrInvalidState
	}
	d.saved = true
	return d.id, nil
}

func (tx *Tx) Rollback() {
	if tx != nil && tx.state == txOpen {
		tx.state = txRolledBack
		tx.drafts = nil
	}
}

func (tx *Tx) Commit() error {
	if tx == nil || tx.state != txOpen {
		return ErrInvalidState
	}
	db := tx.db
	select {
	case <-tx.ctx.Done():
		tx.state = txRolledBack
		return tx.ctx.Err()
	default:
	}
	db.mu.Lock()
	if db.closed {
		db.mu.Unlock()
		return ErrClosed
	}
	active := tx.drafts[:0]
	for _, draft := range tx.drafts {
		if draft.saved {
			active = append(active, draft)
		}
	}
	if len(active) == 0 {
		tx.state = txCommitted
		db.mu.Unlock()
		return nil
	}
	for _, draft := range active {
		if doc, exists := db.documents[draft.document.id]; !exists || doc != draft.document {
			db.mu.Unlock()
			tx.state = txRolledBack
			return ErrNotFound
		}
		resource := Resource{Scope: ScopeObject, Document: draft.document.id, Object: draft.id}
		action := ActionCreate
		var previous *ObjectView
		if draft.expectedRevision != 0 || draft.deleted {
			action = ActionUpdate
			data, ok := db.objects[objectKey{document: draft.document.id, object: draft.id}]
			if !ok {
				db.mu.Unlock()
				tx.state = txRolledBack
				return ErrNotFound
			}
			if binary.LittleEndian.Uint64(data[14:22]) != draft.expectedRevision {
				db.mu.Unlock()
				tx.state = txRolledBack
				return ErrConflict
			}
			visible := filterObject(draft.document, data, func(field FieldID) bool {
				return db.allowedLocked(tx.principal, Resource{Scope: ScopeField, Document: draft.document.id, Object: draft.id, Field: field}, ActionRead)
			})
			if visible == nil {
				db.mu.Unlock()
				tx.state = txRolledBack
				return ErrCorruptData
			}
			previous = &ObjectView{document: draft.document, data: visible}
		}
		if draft.deleted {
			action = ActionDelete
		}
		if !db.allowedLocked(tx.principal, resource, action) {
			db.mu.Unlock()
			tx.state = txRolledBack
			return ErrPermission
		}
		if !draft.deleted && draft.pipeline != 0 {
			pipeline, ok := db.pipelines[draft.pipeline]
			if !ok {
				db.mu.Unlock()
				tx.state = txRolledBack
				return ErrPipelineMissing
			}
			if !db.allowedLocked(tx.principal, resource, ActionExecute) {
				db.mu.Unlock()
				tx.state = txRolledBack
				return ErrPermission
			}
			if err := pipeline.Execute(tx.ctx, &PipelineContext{Principal: tx.principal, Document: draft.document.id, Object: draft.id, Previous: previous, Draft: draft}); err != nil {
				db.mu.Unlock()
				tx.state = txRolledBack
				return err
			}
		}
	}
	prospective := make(map[DocumentID]map[string]FieldID)
	nextFields := make(map[DocumentID]FieldID)
	for _, draft := range active {
		document := draft.document.id
		if _, ok := prospective[document]; !ok {
			prospective[document] = make(map[string]FieldID)
			nextFields[document] = draft.document.nextField
		}
		names := make([]string, 0, len(draft.fields))
		for name := range draft.fields {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			if _, exists := prospective[document][name]; exists {
				continue
			}
			if id, exists := draft.document.fieldID(name); exists {
				prospective[document][name] = id
				continue
			}
			nextFields[document]++
			prospective[document][name] = nextFields[document]
		}
	}
	for _, draft := range active {
		action := ActionCreate
		if draft.expectedRevision != 0 {
			action = ActionUpdate
		}
		if draft.deleted {
			action = ActionDelete
		}
		for name := range draft.changed {
			field := prospective[draft.document.id][name]
			if field == 0 {
				if existing, ok := draft.document.fieldID(name); ok {
					field = existing
				}
			}
			if field == 0 || !db.allowedLocked(tx.principal, Resource{Scope: ScopeField, Document: draft.document.id, Object: draft.id, Field: field}, action) {
				db.mu.Unlock()
				tx.state = txRolledBack
				return ErrPermission
			}
		}
	}
	payload := appendTransactionRecord(nil, active)
	next := db.sequence + 1
	if err := db.log.Append(next, payload, db.durable); err != nil {
		db.mu.Unlock()
		tx.state = txRolledBack
		return err
	}
	events := make([]Event, 0, len(active))
	if err := db.installDrafts(active, next, &events); err != nil {
		db.mu.Unlock()
		tx.state = txRolledBack
		return err
	}
	db.sequence = next
	for i := range events {
		db.publishLocked(events[i])
	}
	tx.state = txCommitted
	db.mu.Unlock()
	return nil
}

func appendTransactionRecord(dst []byte, drafts []*Draft) []byte {
	dst = append(dst, recordTransaction)
	dst = binary.AppendUvarint(dst, uint64(len(drafts)))
	for _, draft := range drafts {
		if draft.deleted {
			dst = append(dst, mutationDelete)
		} else {
			dst = append(dst, mutationPut)
		}
		dst = binary.LittleEndian.AppendUint64(dst, uint64(draft.document.id))
		dst = binary.LittleEndian.AppendUint64(dst, uint64(draft.id))
		revision := draft.expectedRevision + 1
		dst = binary.LittleEndian.AppendUint64(dst, revision)
		dst = binary.LittleEndian.AppendUint64(dst, uint64(draft.pipeline))
		if draft.deleted {
			continue
		}
		names := make([]string, 0, len(draft.fields))
		for name := range draft.fields {
			names = append(names, name)
		}
		sort.Strings(names)
		dst = binary.AppendUvarint(dst, uint64(len(names)))
		for _, name := range names {
			dst = appendString(dst, name)
			dst = appendValue(dst, draft.fields[name])
		}
	}
	return dst
}

func (db *DB) installDrafts(drafts []*Draft, sequence uint64, events *[]Event) error {
	for _, draft := range drafts {
		key := objectKey{document: draft.document.id, object: draft.id}
		if draft.deleted {
			delete(db.objects, key)
			draft.document.objectIDs.remove(draft.id)
			*events = append(*events, Event{Type: EventObjectDeleted, Sequence: sequence, PermissionRevision: db.permissionRevision, Document: key.document, Object: key.object, Revision: draft.expectedRevision + 1})
			continue
		}
		names := make([]string, 0, len(draft.fields))
		for name := range draft.fields {
			names = append(names, name)
		}
		sort.Strings(names)
		fields := make([]fieldValue, 0, len(names))
		for _, name := range names {
			id, err := draft.document.internField(name)
			if err != nil {
				return err
			}
			fields = append(fields, fieldValue{id: id, value: draft.fields[name]})
		}
		data := encodeObject(nil, draft.id, draft.expectedRevision+1, draft.pipeline, fields)
		db.objects[key] = data
		draft.document.objectIDs.add(draft.id)
		typeOfEvent := EventObjectUpdated
		if draft.expectedRevision == 0 {
			typeOfEvent = EventObjectCreated
		}
		*events = append(*events, Event{Type: typeOfEvent, Sequence: sequence, PermissionRevision: db.permissionRevision, Document: key.document, Object: key.object, Revision: draft.expectedRevision + 1, Data: data})
	}
	return nil
}
