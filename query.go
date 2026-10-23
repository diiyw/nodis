package nodis

import (
	"bytes"
	"cmp"
	"context"
	"encoding/binary"
	"math"
	"slices"
	"strings"
	"time"
)

const (
	DefaultQueryLimit  = 100
	MaxQueryLimit      = 1024
	MaxQueryConditions = 64
	MaxQuerySelect     = 128
	MaxQueryPageBytes  = 8 << 20
)

type CompareOp uint8

const (
	Equal CompareOp = iota + 1
	NotEqual
	Less
	LessEqual
	Greater
	GreaterEqual
	Exists
	Missing
	Prefix
)

// Scalar includes the values supported by query comparisons. Fields themselves
// remain schema-free; comparing different stored kinds never matches.
type Scalar interface {
	bool | string | []byte | int | int8 | int16 | int32 | int64 |
		uint | uint8 | uint16 | uint32 | uint64 | float32 | float64 | time.Time
}

type Condition struct {
	Field string
	Op    CompareOp
	Value Value
}

func Filter[T Scalar](field string, op CompareOp, value T) (Condition, error) {
	inferred, err := ValueOf(value)
	if err != nil {
		return Condition{}, err
	}
	c := Condition{Field: field, Op: op, Value: inferred.Clone()}
	if c.Value.kind == KindString {
		c.Value.text = strings.Clone(c.Value.text)
	}
	if err := c.validate(); err != nil {
		return Condition{}, err
	}
	return c, nil
}

func (c Condition) validate() error {
	if !validName(c.Field) || c.Op < Equal || c.Op > Prefix {
		return ErrInvalidQuery
	}
	kind := c.Value.kind
	if kind > KindTime || kind == KindFloat && math.IsNaN(math.Float64frombits(c.Value.num)) {
		return ErrInvalidQuery
	}
	switch c.Op {
	case Exists, Missing:
		if kind != KindNull {
			return ErrInvalidQuery
		}
	case Prefix:
		if kind != KindString && kind != KindBytes {
			return ErrInvalidQuery
		}
	case Less, LessEqual, Greater, GreaterEqual:
		if kind == KindNull || kind == KindBool {
			return ErrInvalidQuery
		}
	}
	return nil
}

type Query struct {
	Document DocumentID
	Where    []Condition // All conditions must match (AND).
	Select   []string    // Empty selects all readable fields.
	After    ObjectID
	Limit    int // Zero uses DefaultQueryLimit.
}

func (q Query) Validate() error {
	if q.Document == 0 || q.Limit < 0 || q.Limit > MaxQueryLimit || len(q.Where) > MaxQueryConditions || len(q.Select) > MaxQuerySelect {
		return ErrInvalidQuery
	}
	for _, condition := range q.Where {
		if err := condition.validate(); err != nil {
			return err
		}
	}
	for i, name := range q.Select {
		if !validName(name) {
			return ErrInvalidQuery
		}
		for _, previous := range q.Select[:i] {
			if previous == name {
				return ErrInvalidQuery
			}
		}
	}
	return nil
}

// Clone owns the slices and variable-length values used by an async query.
func (q Query) Clone() Query {
	clone := q
	clone.Where = append([]Condition(nil), q.Where...)
	for i := range clone.Where {
		clone.Where[i].Field = strings.Clone(clone.Where[i].Field)
		clone.Where[i].Value = clone.Where[i].Value.Clone()
		if clone.Where[i].Value.kind == KindString {
			clone.Where[i].Value.text = strings.Clone(clone.Where[i].Value.text)
		}
	}
	clone.Select = append([]string(nil), q.Select...)
	for i := range clone.Select {
		clone.Select[i] = strings.Clone(clone.Select[i])
	}
	return clone
}

type QueryPage struct {
	Objects            []PortableObject
	Next               ObjectID // Last returned ID when HasMore; zero otherwise.
	HasMore            bool
	Sequence           uint64
	PermissionRevision uint64
}

type preparedCondition struct {
	field FieldID
	op    CompareOp
	value Value
}

// QueryBuffer owns reusable query scratch and results. FindInto results are
// borrowed until the next use of this buffer. Use one buffer per concurrent call.
type QueryBuffer struct {
	objects    []PortableObject
	data       []byte
	conditions []preparedCondition
	selected   []FieldID
	listeners  []uint64
}

func (db *DB) Find(ctx context.Context, actor Principal, query Query) (QueryPage, error) {
	var buffer QueryBuffer
	return db.FindInto(ctx, actor, query, &buffer)
}

func (db *DB) FindInto(ctx context.Context, actor Principal, query Query, buffer *QueryBuffer) (QueryPage, error) {
	if buffer == nil {
		return QueryPage{}, ErrInvalidQuery
	}
	clear(buffer.objects)
	clear(buffer.conditions)
	buffer.objects, buffer.data, buffer.conditions, buffer.selected, buffer.listeners = buffer.objects[:0], buffer.data[:0], buffer.conditions[:0], buffer.selected[:0], buffer.listeners[:0]
	if err := query.Validate(); err != nil {
		return QueryPage{}, err
	}
	if err := ctx.Err(); err != nil {
		return QueryPage{}, err
	}
	db.mu.RLock()
	page, err := db.findLocked(ctx, actor, query, buffer)
	if err == nil {
		for id := range db.listeners {
			buffer.listeners = append(buffer.listeners, id)
		}
	}
	db.mu.RUnlock()
	if err != nil {
		return QueryPage{}, err
	}
	db.publish(buffer.listeners, Event{Type: EventDocumentRead, Sequence: page.Sequence, PermissionRevision: page.PermissionRevision, Document: query.Document})
	for i := range page.Objects {
		object := &page.Objects[i]
		db.publish(buffer.listeners, Event{Type: EventObjectRead, Sequence: page.Sequence, PermissionRevision: page.PermissionRevision, Document: query.Document, Object: object.ID(), Revision: object.Revision()})
	}
	return page, nil
}

func (db *DB) findLocked(ctx context.Context, actor Principal, query Query, buffer *QueryBuffer) (QueryPage, error) {
	if db.closed {
		return QueryPage{}, ErrClosed
	}
	doc, exists := db.documents[query.Document]
	if !exists {
		return QueryPage{}, ErrNotFound
	}
	resource := Resource{Scope: ScopeDocument, Document: query.Document}
	if !db.allowedLocked(actor, resource, ActionDiscover) || !db.allowedLocked(actor, resource, ActionRead) {
		return QueryPage{}, ErrPermission
	}
	for _, condition := range query.Where {
		field, _ := doc.fieldID(condition.Field)
		buffer.conditions = append(buffer.conditions, preparedCondition{field, condition.Op, condition.Value})
	}
	for _, name := range query.Select {
		if field, ok := doc.fieldID(name); ok {
			buffer.selected = append(buffer.selected, field)
		}
	}
	slices.Sort(buffer.selected)
	limit := query.Limit
	if limit == 0 {
		limit = DefaultQueryLimit
	}
	page := QueryPage{Sequence: db.sequence, PermissionRevision: db.permissionRevision}
	for _, id := range doc.objectIDs.after(query.After) {
		if err := ctx.Err(); err != nil {
			return QueryPage{}, err
		}
		resource := Resource{Scope: ScopeObject, Document: doc.id, Object: id}
		if !db.allowedLocked(actor, resource, ActionDiscover) || !db.allowedLocked(actor, resource, ActionRead) {
			continue
		}
		data, ok := db.objects[objectKey{doc.id, id}]
		if !ok {
			return QueryPage{}, ErrCorruptData
		}
		view := ObjectView{document: doc, data: data}
		if !db.queryMatches(actor, &view, buffer.conditions) {
			continue
		}
		if len(buffer.objects) == limit {
			page.HasMore = true
			break
		}
		start := len(buffer.data)
		var err error
		buffer.data, err = db.appendQueryObject(buffer.data, actor, doc, data, buffer.selected, len(query.Select) > 0)
		if err != nil {
			buffer.data = buffer.data[:start]
			if err == ErrQueryTooLarge && len(buffer.objects) > 0 {
				page.HasMore = true
				break
			}
			return QueryPage{}, err
		}
		buffer.objects = append(buffer.objects, PortableObject{data: buffer.data[start:]})
	}
	if err := ctx.Err(); err != nil {
		return QueryPage{}, err
	}
	page.Objects = buffer.objects
	if page.HasMore {
		page.Next = page.Objects[len(page.Objects)-1].ID()
	}
	return page, nil
}

func (db *DB) queryMatches(actor Principal, view *ObjectView, conditions []preparedCondition) bool {
	for _, condition := range conditions {
		resource := Resource{Scope: ScopeField, Document: view.document.id, Object: view.ID(), Field: condition.field}
		// Even Missing/NotEqual must not reveal a hidden field's presence/value.
		if !db.allowedLocked(actor, resource, ActionRead) {
			return false
		}
		value, exists := view.fieldID(condition.field)
		if condition.op == Exists {
			if !exists {
				return false
			}
			continue
		}
		if condition.op == Missing {
			if exists {
				return false
			}
			continue
		}
		if !exists || !matchComparison(value, condition.op, condition.value) {
			return false
		}
	}
	return true
}

func matchComparison(value ValueView, op CompareOp, expected Value) bool {
	if value.kind != expected.kind {
		return false
	}
	var order int
	switch value.kind {
	case KindNull:
		if len(value.payload) != 0 {
			return false
		}
	case KindBool:
		v, ok := value.Bool()
		if !ok {
			return false
		}
		if v != (expected.num != 0) {
			order = 1
		}
	case KindInt, KindTime:
		if len(value.payload) != 8 {
			return false
		}
		order = cmp.Compare(int64(binary.LittleEndian.Uint64(value.payload)), int64(expected.num))
	case KindUint:
		v, ok := value.Uint()
		if !ok {
			return false
		}
		order = cmp.Compare(v, expected.num)
	case KindFloat:
		v, ok := value.Float()
		if !ok || math.IsNaN(v) {
			return false
		}
		order = cmp.Compare(v, math.Float64frombits(expected.num))
	case KindString:
		v, _ := value.String()
		if op == Prefix {
			return strings.HasPrefix(v, expected.text)
		}
		order = cmp.Compare(v, expected.text)
	case KindBytes:
		if op == Prefix {
			return bytes.HasPrefix(value.payload, expected.data)
		}
		order = bytes.Compare(value.payload, expected.data)
	default:
		return false
	}
	switch op {
	case Equal:
		return order == 0
	case NotEqual:
		return order != 0
	case Less:
		return order < 0
	case LessEqual:
		return order <= 0
	case Greater:
		return order > 0
	case GreaterEqual:
		return order >= 0
	default:
		return false
	}
}

func (db *DB) appendQueryObject(dst []byte, actor Principal, doc *Document, data []byte, selected []FieldID, projection bool) ([]byte, error) {
	if len(data) < objectHeaderSize || string(data[:4]) != "NDO1" {
		return dst, ErrCorruptData
	}
	if len(dst)+portableHeaderSize > MaxQueryPageBytes {
		return dst, ErrQueryTooLarge
	}
	start := len(dst)
	dst = append(dst, 'N', 'D', 'X', '1')
	dst = append(dst, data[4:objectHeaderSize]...)
	var visible uint16
	pos := objectHeaderSize
	object := ObjectID(binary.LittleEndian.Uint64(data[6:14]))
	for range int(binary.LittleEndian.Uint16(data[4:6])) {
		if len(data)-pos < 4 {
			return dst, ErrCorruptData
		}
		field := FieldID(binary.LittleEndian.Uint32(data[pos:]))
		pos += 4
		_, used, ok := readValue(data[pos:])
		if !ok {
			return dst, ErrCorruptData
		}
		_, included := slices.BinarySearch(selected, field)
		if (!projection || included) && db.allowedLocked(actor, Resource{Scope: ScopeField, Document: doc.id, Object: object, Field: field}, ActionRead) {
			doc.mu.RLock()
			name, exists := doc.fieldNames[field]
			doc.mu.RUnlock()
			if !exists {
				return dst, ErrCorruptData
			}
			// Names are at most 255 bytes, hence at most two varint bytes.
			width := 1
			if len(name) >= 128 {
				width++
			}
			if len(dst)+width+len(name)+used > MaxQueryPageBytes {
				return dst, ErrQueryTooLarge
			}
			dst = appendString(dst, name)
			dst = append(dst, data[pos:pos+used]...)
			visible++
		}
		pos += used
	}
	if pos != len(data) {
		return dst, ErrCorruptData
	}
	binary.LittleEndian.PutUint16(dst[start+4:start+6], visible)
	return dst, nil
}
