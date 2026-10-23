package nodis

import (
	"cmp"
	"context"
	"encoding/binary"
	"math"
	"slices"
	"strings"
)

type indexEntry[K any] struct {
	key K
	id  ObjectID
}

func compareIndexEntry[K any](a, b indexEntry[K], compare func(K, K) int) int {
	if order := compare(a.key, b.key); order != 0 {
		return order
	}
	return cmp.Compare(a.id, b.id)
}

// Text keys borrow immutable encoded object bytes, including byte-valued fields.
// Numeric keys preserve kind and bits; comparison normalizes floating-point zero.
type scalarKey struct {
	kind Kind
	num  uint64
	text string
}

func compareScalarKey(a, b scalarKey) int {
	if order := cmp.Compare(a.kind, b.kind); order != 0 {
		return order
	}
	switch a.kind {
	case KindInt, KindTime:
		return cmp.Compare(int64(a.num), int64(b.num))
	case KindFloat:
		return cmp.Compare(math.Float64frombits(a.num), math.Float64frombits(b.num))
	case KindString, KindBytes:
		return cmp.Compare(a.text, b.text)
	default:
		return cmp.Compare(a.num, b.num)
	}
}

func scalarIndexKey(value ValueView) (scalarKey, bool) {
	key := scalarKey{kind: value.kind}
	switch value.kind {
	case KindNull:
		return key, len(value.payload) == 0
	case KindBool:
		v, ok := value.Bool()
		if v {
			key.num = 1
		}
		return key, ok
	case KindInt, KindUint, KindFloat, KindTime:
		if len(value.payload) != 8 {
			return key, false
		}
		key.num = binary.LittleEndian.Uint64(value.payload)
		return key, value.kind != KindFloat || !math.IsNaN(math.Float64frombits(key.num))
	case KindString, KindBytes:
		key.text = bytesToString(value.payload)
		return key, true
	default:
		// Containers and NaN still participate in Exists/Missing through a scan.
		return key, false
	}
}

type indexMutation uint8

const (
	indexAppend indexMutation = iota // Unsorted bulk recovery; sort once afterward.
	indexAdd
	indexRemove
)

func (doc *Document) indexObject(data []byte, mutation indexMutation) error {
	if len(data) < objectHeaderSize || string(data[:4]) != "NDO1" {
		return ErrCorruptData
	}
	id := ObjectID(binary.LittleEndian.Uint64(data[6:14]))
	pos := objectHeaderSize
	var previous FieldID
	for range int(binary.LittleEndian.Uint16(data[4:6])) {
		if len(data)-pos < 4 {
			return ErrCorruptData
		}
		field := FieldID(binary.LittleEndian.Uint32(data[pos:]))
		pos += 4
		if field <= previous || field > doc.nextField {
			return ErrCorruptData
		}
		previous = field
		value, used, ok := readValue(data[pos:])
		if !ok {
			return ErrCorruptData
		}
		pos += used
		key, indexed := scalarIndexKey(value)
		if !indexed {
			continue
		}
		index := doc.indexes[field]
		if mutation == indexRemove {
			if index != nil {
				index.remove(key, id, compareScalarKey)
			}
			continue
		}
		if index == nil {
			index = new(orderedIndex[scalarKey])
			doc.indexes[field] = index
		}
		if mutation == indexAppend {
			index.appendUnsorted(indexEntry[scalarKey]{key, id})
		} else {
			index.add(key, id, compareScalarKey)
		}
	}
	if pos != len(data) {
		return ErrCorruptData
	}
	return nil
}

// Rebuild only the final live state, after snapshot restore and complete WAL
// replay. Indexes are derived data and require no persistent format changes.
func (db *DB) rebuildIndexes() error {
	for key, data := range db.objects {
		doc := db.documents[key.document]
		if doc == nil {
			return ErrCorruptData
		}
		if err := doc.indexObject(data, indexAppend); err != nil {
			return err
		}
	}
	for _, doc := range db.documents {
		for _, index := range doc.indexes {
			index.build(compareScalarKey)
		}
	}
	return nil
}

type indexRanges[K any] struct {
	first, second indexSpan[K]
}

func (ranges indexRanges[K]) count() int { return ranges.first.length + ranges.second.length }

func scalarIndexRanges(index *orderedIndex[scalarKey], condition preparedCondition) indexRanges[scalarKey] {
	if index == nil || index.len() == 0 {
		return indexRanges[scalarKey]{}
	}
	value := condition.value
	key := scalarKey{kind: value.kind, num: value.num, text: value.text}
	if value.kind == KindBytes {
		key.text = bytesToString(value.data)
	} else if value.kind == KindBool && key.num != 0 {
		key.num = 1
	}
	lower := func() indexPosition {
		return index.bound(func(value scalarKey) bool { return compareScalarKey(value, key) < 0 })
	}
	upper := func() indexPosition {
		return index.bound(func(value scalarKey) bool { return compareScalarKey(value, key) <= 0 })
	}
	start := func() indexPosition {
		return index.bound(func(value scalarKey) bool { return value.kind < key.kind })
	}
	end := func() indexPosition {
		return index.bound(func(value scalarKey) bool { return value.kind <= key.kind })
	}
	switch condition.op {
	case Equal:
		return indexRanges[scalarKey]{first: index.span(lower(), upper())}
	case NotEqual:
		return indexRanges[scalarKey]{first: index.span(start(), lower()), second: index.span(upper(), end())}
	case Less:
		return indexRanges[scalarKey]{first: index.span(start(), lower())}
	case LessEqual:
		return indexRanges[scalarKey]{first: index.span(start(), upper())}
	case Greater:
		return indexRanges[scalarKey]{first: index.span(upper(), end())}
	case GreaterEqual:
		return indexRanges[scalarKey]{first: index.span(lower(), end())}
	case Prefix:
		// The matching prefix interval needs no allocated successor string,
		// including for empty and all-FF byte prefixes.
		last := index.bound(func(value scalarKey) bool {
			return value.kind < key.kind || value.kind == key.kind && (value.text < key.text || strings.HasPrefix(value.text, key.text))
		})
		return indexRanges[scalarKey]{first: index.span(lower(), last)}
	default:
		return indexRanges[scalarKey]{}
	}
}

// Pick the narrowest scalar range when it is smaller than the primary scan.
// Candidate cardinalities are internal; ACLs and all predicates are rechecked
// before projection, events, or pagination metadata can expose any result.
func (doc *Document) queryIDs(ctx context.Context, after ObjectID, conditions []preparedCondition, buffer *QueryBuffer) ([]ObjectID, error) {
	primary := doc.objectIDs.after(after)
	bestCount := len(primary)
	var best indexRanges[scalarKey]
	indexed := false
	for _, condition := range conditions {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if condition.op == Exists || condition.op == Missing {
			continue
		}
		ranges := scalarIndexRanges(doc.indexes[condition.field], condition)
		if count := ranges.count(); count < bestCount {
			best, bestCount, indexed = ranges, count, true
			if count == 0 {
				return nil, nil
			}
		}
	}
	if !indexed {
		return primary, nil
	}
	for _, span := range [2]indexSpan[scalarKey]{best.first, best.second} {
		cursor := span.cursor()
		for entry, ok := cursor.next(); ok; entry, ok = cursor.next() {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			if entry.id > after {
				buffer.candidates = append(buffer.candidates, entry.id)
			}
		}
	}
	slices.Sort(buffer.candidates)
	return buffer.candidates, ctx.Err()
}
