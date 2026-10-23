package nodis

import (
	"cmp"
	"encoding/binary"
	"slices"
	"time"

	"github.com/diiyw/nodis/storage"
)

const (
	snapshotMetadata byte = 240 + iota
	snapshotDocument
	snapshotObject
)

// Checkpoint captures committed state and compacts all historical mutations into
// a snapshot, then recycles the WAL. It blocks commits until publication finishes.
// Administrative callers should keep this DB-level operation on the server.
func (db *DB) Checkpoint() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.closed {
		return ErrClosed
	}
	db.checkpointErr = storage.Checkpoint(db.log, db.sequence, snapshotSource{db: db})
	return db.checkpointErr
}

// CheckpointError reports the most recent manual or periodic checkpoint result.
func (db *DB) CheckpointError() error {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.checkpointErr
}

func (db *DB) checkpointLoop(interval time.Duration) {
	defer close(db.checkpointDone)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-db.checkpointStop:
			return
		case <-ticker.C:
			_ = db.Checkpoint()
		}
	}
}

type snapshotSource struct{ db *DB }

func appendSortedKeys[K cmp.Ordered, V any](dst []K, source map[K]V) []K {
	dst = dst[:0]
	for key := range source {
		dst = append(dst, key)
	}
	slices.Sort(dst)
	return dst
}

func (source snapshotSource) WriteSnapshot(writer *storage.SnapshotWriter) error {
	db := source.db
	buffer := db.snapshotBuffer[:0]
	defer func() { db.snapshotBuffer = buffer[:0] }()
	buffer = append(buffer, snapshotMetadata)
	buffer = binary.LittleEndian.AppendUint64(buffer, uint64(db.nextDocument))
	buffer = binary.LittleEndian.AppendUint64(buffer, uint64(db.nextObject))
	buffer = binary.LittleEndian.AppendUint64(buffer, db.permissionRevision)
	if err := writer.Append(buffer); err != nil {
		return err
	}
	db.snapshotDocuments = appendSortedKeys(db.snapshotDocuments, db.documents)
	for _, id := range db.snapshotDocuments {
		doc := db.documents[id]
		buffer = append(buffer[:0], snapshotDocument)
		buffer = binary.LittleEndian.AppendUint64(buffer, uint64(id))
		buffer = binary.LittleEndian.AppendUint64(buffer, uint64(doc.defaultPipeline))
		buffer = appendString(buffer, doc.name)
		buffer = binary.LittleEndian.AppendUint32(buffer, uint32(doc.nextField))
		for field := FieldID(1); field <= doc.nextField; field++ {
			buffer = appendString(buffer, doc.fieldNames[field])
		}
		if err := writer.Append(buffer); err != nil {
			return err
		}
	}
	for i := range db.rules {
		buffer = appendPolicyRecord(buffer[:0], db.rules[i])
		if err := writer.Append(buffer); err != nil {
			return err
		}
	}
	db.snapshotObjects = db.snapshotObjects[:0]
	for key := range db.objects {
		db.snapshotObjects = append(db.snapshotObjects, key)
	}
	slices.SortFunc(db.snapshotObjects, func(a, b objectKey) int {
		if order := cmp.Compare(a.document, b.document); order != 0 {
			return order
		}
		return cmp.Compare(a.object, b.object)
	})
	for _, key := range db.snapshotObjects {
		buffer = append(buffer[:0], snapshotObject)
		buffer = binary.LittleEndian.AppendUint64(buffer, uint64(key.document))
		buffer = append(buffer, db.objects[key]...)
		if err := writer.Append(buffer); err != nil {
			return err
		}
	}
	return nil
}

// Snapshot recovery preserves dictionary IDs exactly; reconstructing dictionaries
// from live objects would retarget ACLs after a field or object was deleted.
type snapshotRestore struct {
	db                 *DB
	metadata           bool
	permissionRevision uint64
}

func (restore *snapshotRestore) apply(data []byte) error {
	if len(data) == 0 {
		return ErrCorruptData
	}
	db := restore.db
	if !restore.metadata && data[0] != snapshotMetadata {
		return ErrCorruptData
	}
	switch data[0] {
	case snapshotMetadata:
		if restore.metadata || len(data) != 25 {
			return ErrCorruptData
		}
		db.nextDocument = DocumentID(binary.LittleEndian.Uint64(data[1:9]))
		db.nextObject = ObjectID(binary.LittleEndian.Uint64(data[9:17]))
		restore.permissionRevision = binary.LittleEndian.Uint64(data[17:25])
		restore.metadata = true
	case snapshotDocument:
		if len(data) < 17 {
			return ErrCorruptData
		}
		id := DocumentID(binary.LittleEndian.Uint64(data[1:9]))
		if id == 0 || id > db.nextDocument {
			return ErrCorruptData
		}
		if _, exists := db.documents[id]; exists {
			return ErrCorruptData
		}
		pipeline := PipelineID(binary.LittleEndian.Uint64(data[9:17]))
		name, n, ok := readString(data[17:])
		if !ok || !validName(name) {
			return ErrCorruptData
		}
		if _, exists := db.documentsByName[name]; exists {
			return ErrCorruptData
		}
		pos := 17 + n
		if len(data)-pos < 4 {
			return ErrCorruptData
		}
		count := binary.LittleEndian.Uint32(data[pos:])
		pos += 4
		if uint64(count) > uint64(len(data)-pos)/2 {
			return ErrCorruptData
		}
		doc := newDocument(db, id, name)
		doc.defaultPipeline = pipeline
		for i := uint32(0); i < count; i++ {
			field, n, ok := readString(data[pos:])
			if !ok || !validName(field) {
				return ErrCorruptData
			}
			pos += n
			if _, exists := doc.fieldsByName[field]; exists {
				return ErrCorruptData
			}
			fieldID, err := doc.internField(field)
			if err != nil || fieldID != FieldID(i+1) {
				return ErrCorruptData
			}
		}
		if pos != len(data) {
			return ErrCorruptData
		}
		db.documents[id], db.documentsByName[name] = doc, id
	case recordPolicy:
		rule, err := decodePolicyRecord(data[1:])
		if err != nil {
			return err
		}
		db.rules = append(db.rules, rule)
	case snapshotObject:
		if len(data) < 9+objectHeaderSize {
			return ErrCorruptData
		}
		document := DocumentID(binary.LittleEndian.Uint64(data[1:9]))
		doc, exists := db.documents[document]
		if !exists {
			return ErrCorruptData
		}
		object := data[9:]
		id := ObjectID(binary.LittleEndian.Uint64(object[6:14]))
		if id == 0 || id > db.nextObject || objectRevision(object) == 0 {
			return ErrCorruptData
		}
		key := objectKey{document: document, object: id}
		if _, exists := db.objects[key]; exists {
			return ErrCorruptData
		}
		if err := validateSnapshotObject(doc, object); err != nil {
			return err
		}
		db.objects[key] = append([]byte(nil), object...)
		doc.objectIDs.add(id)
	default:
		return ErrCorruptData
	}
	return nil
}

func validateSnapshotObject(doc *Document, data []byte) error {
	if len(data) < objectHeaderSize || string(data[:4]) != "NDO1" {
		return ErrCorruptData
	}
	pos := objectHeaderSize
	var previous FieldID
	for range int(binary.LittleEndian.Uint16(data[4:6])) {
		if len(data)-pos < 4 {
			return ErrCorruptData
		}
		id := FieldID(binary.LittleEndian.Uint32(data[pos:]))
		pos += 4
		if id <= previous || id > doc.nextField {
			return ErrCorruptData
		}
		previous = id
		_, used, ok := readValue(data[pos:])
		if !ok {
			return ErrCorruptData
		}
		pos += used
	}
	if pos != len(data) {
		return ErrCorruptData
	}
	return nil
}
