package nodis

import "sync"

type Document struct {
	db              *DB
	id              DocumentID
	name            string
	mu              sync.RWMutex
	fieldsByName    map[string]FieldID
	fieldNames      map[FieldID]string
	nextField       FieldID
	defaultPipeline PipelineID
	objectIDs       orderedIDs[ObjectID]
	indexes         map[FieldID]*orderedIndex[scalarKey]
}

func (d *Document) ID() DocumentID {
	if d == nil {
		return 0
	}
	return d.id
}
func (d *Document) Name() string {
	if d == nil {
		return ""
	}
	if d.db != nil {
		d.db.mu.RLock()
		defer d.db.mu.RUnlock()
	}
	return d.name
}

func (d *Document) SetDefaultPipeline(actor Principal, pipeline PipelineID) error {
	if d == nil || d.db == nil {
		return ErrNotFound
	}
	d.db.mu.Lock()
	if d.db.closed {
		d.db.mu.Unlock()
		return ErrClosed
	}
	if current, ok := d.db.documents[d.id]; !ok || current != d {
		d.db.mu.Unlock()
		return ErrNotFound
	}
	if !d.db.allowedLocked(actor, Resource{Scope: ScopeDocument, Document: d.id}, ActionUpdate) {
		d.db.mu.Unlock()
		return ErrPermission
	}
	if pipeline != 0 {
		if _, ok := d.db.pipelines[pipeline]; !ok {
			d.db.mu.Unlock()
			return ErrPipelineMissing
		}
	}
	next := d.db.sequence + 1
	if err := d.db.log.Append(next, appendDocumentUpdateRecord(nil, d.id, d.name, pipeline), d.db.durable); err != nil {
		d.db.mu.Unlock()
		return err
	}
	d.defaultPipeline = pipeline
	d.db.sequence = next
	event := Event{Type: EventDocumentUpdated, Sequence: next, PermissionRevision: d.db.permissionRevision, Document: d.id}
	d.db.publishLocked(event)
	d.db.mu.Unlock()
	return nil
}

func (d *Document) fieldID(name string) (FieldID, bool) {
	d.mu.RLock()
	id, ok := d.fieldsByName[name]
	d.mu.RUnlock()
	return id, ok
}

func (d *Document) FieldID(name string) (FieldID, bool) { return d.fieldID(name) }

func (d *Document) internField(name string) (FieldID, error) {
	if !validName(name) {
		return 0, ErrInvalidName
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if id, ok := d.fieldsByName[name]; ok {
		return id, nil
	}
	d.nextField++
	id := d.nextField
	d.fieldsByName[name] = id
	d.fieldNames[id] = name
	return id, nil
}
