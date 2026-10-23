package nodis

import "encoding/binary"

const (
	recordDocument       byte = 1
	recordTransaction    byte = 2
	recordPolicy         byte = 3
	recordPolicyRevoke   byte = 4
	recordDocumentDelete byte = 5
	recordDocumentUpdate byte = 6

	mutationPut    byte = 1
	mutationDelete byte = 2
)

func appendString(dst []byte, value string) []byte {
	dst = binary.AppendUvarint(dst, uint64(len(value)))
	return append(dst, value...)
}

func readString(data []byte) (string, int, bool) {
	n, width := binary.Uvarint(data)
	if width <= 0 || n > uint64(len(data)-width) {
		return "", 0, false
	}
	return string(data[width : width+int(n)]), width + int(n), true
}

func appendDocumentRecord(dst []byte, id DocumentID, name string) []byte {
	dst = append(dst, recordDocument)
	dst = binary.LittleEndian.AppendUint64(dst, uint64(id))
	return appendString(dst, name)
}

func appendDocumentDeleteRecord(dst []byte, id DocumentID) []byte {
	dst = append(dst, recordDocumentDelete)
	return binary.LittleEndian.AppendUint64(dst, uint64(id))
}

func appendDocumentUpdateRecord(dst []byte, id DocumentID, name string, pipeline PipelineID) []byte {
	dst = append(dst, recordDocumentUpdate)
	dst = binary.LittleEndian.AppendUint64(dst, uint64(id))
	dst = binary.LittleEndian.AppendUint64(dst, uint64(pipeline))
	return appendString(dst, name)
}

func appendPolicyRecord(dst []byte, rule Rule) []byte {
	return appendRuleRecord(dst, recordPolicy, rule)
}

func appendPolicyRevokeRecord(dst []byte, rule Rule) []byte {
	return appendRuleRecord(dst, recordPolicyRevoke, rule)
}

func appendRuleRecord(dst []byte, record byte, rule Rule) []byte {
	dst = append(dst, record, byte(rule.SubjectKind))
	dst = appendString(dst, rule.Subject)
	dst = append(dst, byte(rule.Resource.Scope))
	dst = binary.LittleEndian.AppendUint64(dst, uint64(rule.Resource.Document))
	dst = binary.LittleEndian.AppendUint64(dst, uint64(rule.Resource.Object))
	dst = binary.LittleEndian.AppendUint32(dst, uint32(rule.Resource.Field))
	dst = binary.LittleEndian.AppendUint16(dst, uint16(rule.Actions))
	if rule.Allow {
		return append(dst, 1)
	}
	return append(dst, 0)
}

func (db *DB) applyRecord(data []byte) error {
	if len(data) == 0 {
		return ErrCorruptData
	}
	switch data[0] {
	case recordDocument:
		if len(data) < 9 {
			return ErrCorruptData
		}
		id := DocumentID(binary.LittleEndian.Uint64(data[1:9]))
		name, n, ok := readString(data[9:])
		if !ok || 9+n != len(data) {
			return ErrCorruptData
		}
		if !validName(name) {
			return ErrCorruptData
		}
		doc := newDocument(db, id, name)
		db.documents[id], db.documentsByName[name] = doc, id
		if id > db.nextDocument {
			db.nextDocument = id
		}
		return nil
	case recordPolicy:
		return db.applyPolicyRecord(data[1:])
	case recordPolicyRevoke:
		rule, err := decodePolicyRecord(data[1:])
		if err != nil {
			return err
		}
		db.removeRuleLocked(rule)
		db.permissionRevision++
		return nil
	case recordDocumentDelete:
		if len(data) != 9 {
			return ErrCorruptData
		}
		return db.applyDocumentDelete(DocumentID(binary.LittleEndian.Uint64(data[1:])))
	case recordDocumentUpdate:
		if len(data) < 17 {
			return ErrCorruptData
		}
		id := DocumentID(binary.LittleEndian.Uint64(data[1:9]))
		pipeline := PipelineID(binary.LittleEndian.Uint64(data[9:17]))
		name, n, ok := readString(data[17:])
		if !ok || 17+n != len(data) || !validName(name) {
			return ErrCorruptData
		}
		doc, ok := db.documents[id]
		if !ok {
			return ErrCorruptData
		}
		delete(db.documentsByName, doc.name)
		doc.name, doc.defaultPipeline = name, pipeline
		db.documentsByName[name] = id
		return nil
	case recordTransaction:
		return db.applyTransactionRecord(data[1:])
	default:
		return ErrCorruptData
	}
}

func (db *DB) applyDocumentDelete(id DocumentID) error {
	doc, ok := db.documents[id]
	if !ok {
		return ErrCorruptData
	}
	delete(db.documents, id)
	delete(db.documentsByName, doc.name)
	doc.indexes, doc.objectIDs = nil, nil
	for key := range db.objects {
		if key.document == id {
			delete(db.objects, key)
		}
	}
	return nil
}

func (db *DB) applyPolicyRecord(data []byte) error {
	rule, err := decodePolicyRecord(data)
	if err != nil {
		return err
	}
	db.rules = append(db.rules, rule)
	db.permissionRevision++
	return nil
}

func decodePolicyRecord(data []byte) (Rule, error) {
	if len(data) < 2 {
		return Rule{}, ErrCorruptData
	}
	rule := Rule{SubjectKind: SubjectKind(data[0])}
	name, n, ok := readString(data[1:])
	if !ok {
		return Rule{}, ErrCorruptData
	}
	rule.Subject = name
	data = data[1+n:]
	if len(data) != 1+8+8+4+2+1 {
		return Rule{}, ErrCorruptData
	}
	rule.Resource.Scope = Scope(data[0])
	data = data[1:]
	rule.Resource.Document = DocumentID(binary.LittleEndian.Uint64(data))
	data = data[8:]
	rule.Resource.Object = ObjectID(binary.LittleEndian.Uint64(data))
	data = data[8:]
	rule.Resource.Field = FieldID(binary.LittleEndian.Uint32(data))
	data = data[4:]
	rule.Actions = Action(binary.LittleEndian.Uint16(data))
	data = data[2:]
	rule.Allow = data[0] == 1
	return rule, nil
}

func decodeOwnedValue(data []byte) (Value, int, bool) {
	view, n, ok := readValue(data)
	if !ok {
		return Value{}, 0, false
	}
	v := Value{kind: view.kind}
	switch view.kind {
	case KindNull:
	case KindBool:
		if len(view.payload) != 1 {
			return Value{}, 0, false
		}
		v.num = uint64(view.payload[0])
	case KindInt, KindUint, KindFloat, KindTime:
		if len(view.payload) != 8 {
			return Value{}, 0, false
		}
		v.num = binary.LittleEndian.Uint64(view.payload)
	case KindString:
		v.text = string(view.payload)
	case KindBytes:
		v.data = append([]byte(nil), view.payload...)
	case KindArray:
		count, width := binary.Uvarint(view.payload)
		if width <= 0 {
			return Value{}, 0, false
		}
		pos := width
		v.list = make([]Value, 0, int(count))
		for range count {
			item, used, ok := decodeOwnedValue(view.payload[pos:])
			if !ok {
				return Value{}, 0, false
			}
			pos += used
			v.list = append(v.list, item)
		}
		if pos != len(view.payload) {
			return Value{}, 0, false
		}
	case KindObject:
		count, width := binary.Uvarint(view.payload)
		if width <= 0 {
			return Value{}, 0, false
		}
		pos := width
		v.obj = make([]NamedValue, 0, int(count))
		for range count {
			name, used, ok := readString(view.payload[pos:])
			if !ok {
				return Value{}, 0, false
			}
			pos += used
			item, consumed, ok := decodeOwnedValue(view.payload[pos:])
			if !ok {
				return Value{}, 0, false
			}
			pos += consumed
			v.obj = append(v.obj, NamedValue{Name: name, Value: item})
		}
		if pos != len(view.payload) {
			return Value{}, 0, false
		}
	default:
		return Value{}, 0, false
	}
	return v, n, true
}

func decodeObjectFields(doc *Document, data []byte) (map[string]Value, error) {
	if len(data) < objectHeaderSize || string(data[:4]) != "NDO1" {
		return nil, ErrCorruptData
	}
	fields := make(map[string]Value, int(binary.LittleEndian.Uint16(data[4:6])))
	pos := objectHeaderSize
	for range int(binary.LittleEndian.Uint16(data[4:6])) {
		if len(data)-pos < 4 {
			return nil, ErrCorruptData
		}
		id := FieldID(binary.LittleEndian.Uint32(data[pos:]))
		pos += 4
		doc.mu.RLock()
		name, ok := doc.fieldNames[id]
		doc.mu.RUnlock()
		if !ok {
			return nil, ErrCorruptData
		}
		value, n, ok := decodeOwnedValue(data[pos:])
		if !ok {
			return nil, ErrCorruptData
		}
		pos += n
		fields[name] = value
	}
	if pos != len(data) {
		return nil, ErrCorruptData
	}
	return fields, nil
}

func (db *DB) applyTransactionRecord(data []byte) error {
	count, width := binary.Uvarint(data)
	if width <= 0 {
		return ErrCorruptData
	}
	pos := width
	for range count {
		if len(data)-pos < 1+8+8+8+8 {
			return ErrCorruptData
		}
		op := data[pos]
		pos++
		document := DocumentID(binary.LittleEndian.Uint64(data[pos:]))
		pos += 8
		object := ObjectID(binary.LittleEndian.Uint64(data[pos:]))
		pos += 8
		revision := binary.LittleEndian.Uint64(data[pos:])
		pos += 8
		pipeline := PipelineID(binary.LittleEndian.Uint64(data[pos:]))
		pos += 8
		doc, ok := db.documents[document]
		if !ok {
			return ErrCorruptData
		}
		key := objectKey{document: document, object: object}
		if object > db.nextObject {
			db.nextObject = object
		}
		if op == mutationDelete {
			delete(db.objects, key)
			doc.objectIDs.remove(object)
			continue
		}
		if op != mutationPut {
			return ErrCorruptData
		}
		fieldCount, n := binary.Uvarint(data[pos:])
		if n <= 0 {
			return ErrCorruptData
		}
		pos += n
		fields := make([]fieldValue, 0, int(fieldCount))
		for range fieldCount {
			name, used, ok := readString(data[pos:])
			if !ok || !validName(name) {
				return ErrCorruptData
			}
			pos += used
			value, consumed, ok := decodeOwnedValue(data[pos:])
			if !ok {
				return ErrCorruptData
			}
			pos += consumed
			id, err := doc.internField(name)
			if err != nil {
				return err
			}
			fields = append(fields, fieldValue{id: id, value: value})
		}
		db.objects[key] = encodeObject(nil, object, revision, pipeline, fields)
		doc.objectIDs.add(object)
	}
	if pos != len(data) {
		return ErrCorruptData
	}
	return nil
}
