package protocol

import (
	"encoding/binary"
	"math"
	"time"
	"unsafe"

	"github.com/diiyw/nodis"
)

const queryHeaderSize = 28
const queryPageHeaderSize = 31

func AppendQuery(dst []byte, query nodis.Query) ([]byte, error) {
	if err := query.Validate(); err != nil {
		return dst, err
	}
	start := len(dst)
	dst = append(dst, 'N', 'D', 'Q', '1')
	dst = binary.LittleEndian.AppendUint64(dst, uint64(query.Document))
	dst = binary.LittleEndian.AppendUint64(dst, uint64(query.After))
	dst = binary.LittleEndian.AppendUint32(dst, uint32(query.Limit))
	dst = binary.LittleEndian.AppendUint16(dst, uint16(len(query.Where)))
	dst = binary.LittleEndian.AppendUint16(dst, uint16(len(query.Select)))
	for _, condition := range query.Where {
		dst = appendQueryName(dst, condition.Field)
		dst = append(dst, byte(condition.Op))
		dst = nodis.AppendValue(dst, condition.Value)
	}
	for _, name := range query.Select {
		dst = appendQueryName(dst, name)
	}
	if len(dst)-start > MaxFrameSize-CommandHeaderSize-4 {
		return dst[:start], ErrInvalidCommand
	}
	return dst, nil
}

type QueryDecoder struct {
	conditions []nodis.Condition
	selected   []string
}

// DecodeQueryInto borrows scalar payloads and names from data and reuses decoder
// slices. The returned query is valid until data or decoder is reused.
func DecodeQueryInto(data []byte, decoder *QueryDecoder) (nodis.Query, error) {
	if decoder == nil || len(data) < queryHeaderSize || len(data) > MaxFrameSize-CommandHeaderSize-4 || string(data[:4]) != "NDQ1" {
		return nodis.Query{}, ErrInvalidCommand
	}
	clear(decoder.conditions)
	clear(decoder.selected)
	decoder.conditions, decoder.selected = decoder.conditions[:0], decoder.selected[:0]
	query := nodis.Query{Document: nodis.DocumentID(binary.LittleEndian.Uint64(data[4:12])), After: nodis.ObjectID(binary.LittleEndian.Uint64(data[12:20]))}
	limit := binary.LittleEndian.Uint32(data[20:24])
	conditions, selected := int(binary.LittleEndian.Uint16(data[24:26])), int(binary.LittleEndian.Uint16(data[26:28]))
	if limit > nodis.MaxQueryLimit || conditions > nodis.MaxQueryConditions || selected > nodis.MaxQuerySelect {
		return nodis.Query{}, ErrInvalidCommand
	}
	query.Limit = int(limit)
	pos := queryHeaderSize
	for range conditions {
		name, used, ok := readQueryName(data[pos:])
		if !ok {
			return nodis.Query{}, ErrInvalidCommand
		}
		pos += used
		if pos == len(data) {
			return nodis.Query{}, ErrInvalidCommand
		}
		op := nodis.CompareOp(data[pos])
		pos++
		value, used, ok := readQueryScalar(data[pos:])
		if !ok {
			return nodis.Query{}, ErrInvalidCommand
		}
		pos += used
		decoder.conditions = append(decoder.conditions, nodis.Condition{Field: name, Op: op, Value: value})
	}
	for range selected {
		name, used, ok := readQueryName(data[pos:])
		if !ok {
			return nodis.Query{}, ErrInvalidCommand
		}
		pos += used
		decoder.selected = append(decoder.selected, name)
	}
	if pos != len(data) {
		return nodis.Query{}, ErrInvalidCommand
	}
	query.Where, query.Select = decoder.conditions, decoder.selected
	if err := query.Validate(); err != nil {
		return nodis.Query{}, ErrInvalidCommand
	}
	return query, nil
}

func appendQueryName(dst []byte, name string) []byte {
	dst = binary.AppendUvarint(dst, uint64(len(name)))
	return append(dst, name...)
}

func readQueryName(data []byte) (string, int, bool) {
	size, width := binary.Uvarint(data)
	if width <= 0 || size == 0 || size > 255 || size > uint64(len(data)-width) {
		return "", 0, false
	}
	payload := data[width : width+int(size)]
	return unsafe.String(unsafe.SliceData(payload), len(payload)), width + int(size), true
}

// Unlike the general value decoder, this accepts no nested containers and never
// allocates storage based on an untrusted element count.
func readQueryScalar(data []byte) (nodis.Value, int, bool) {
	if len(data) < 2 {
		return nodis.Value{}, 0, false
	}
	kind := nodis.Kind(data[0])
	size, width := binary.Uvarint(data[1:])
	if width <= 0 || size > uint64(len(data)-1-width) || kind > nodis.KindTime {
		return nodis.Value{}, 0, false
	}
	payload := data[1+width : 1+width+int(size)]
	var value nodis.Value
	switch kind {
	case nodis.KindNull:
		if len(payload) != 0 {
			return value, 0, false
		}
	case nodis.KindBool:
		if len(payload) != 1 || payload[0] > 1 {
			return value, 0, false
		}
		value, _ = nodis.ValueOf(payload[0] == 1)
	case nodis.KindInt, nodis.KindUint, nodis.KindFloat, nodis.KindTime:
		if len(payload) != 8 {
			return value, 0, false
		}
		number := binary.LittleEndian.Uint64(payload)
		switch kind {
		case nodis.KindInt:
			value, _ = nodis.ValueOf(int64(number))
		case nodis.KindUint:
			value, _ = nodis.ValueOf(number)
		case nodis.KindFloat:
			value, _ = nodis.ValueOf(math.Float64frombits(number))
		case nodis.KindTime:
			value, _ = nodis.ValueOf(time.Unix(0, int64(number)))
		}
	case nodis.KindString:
		value, _ = nodis.ValueOf(unsafe.String(unsafe.SliceData(payload), len(payload)))
	case nodis.KindBytes:
		value, _ = nodis.ValueOf(payload)
	}
	return value, 1 + width + int(size), true
}

func AppendQueryPage(dst []byte, page nodis.QueryPage) ([]byte, error) {
	if len(page.Objects) > nodis.MaxQueryLimit {
		return dst, ErrInvalidCommand
	}
	if !validPageCursor(page) {
		return dst, ErrInvalidCommand
	}
	start := len(dst)
	dst = append(dst, 'N', 'Q', 'P', '1')
	dst = binary.LittleEndian.AppendUint64(dst, page.Sequence)
	dst = binary.LittleEndian.AppendUint64(dst, page.PermissionRevision)
	dst = binary.LittleEndian.AppendUint64(dst, uint64(page.Next))
	flag := byte(0)
	if page.HasMore {
		flag = 1
	}
	dst = append(dst, flag)
	dst = binary.LittleEndian.AppendUint16(dst, uint16(len(page.Objects)))
	var previous nodis.ObjectID
	size := 0
	for i := range page.Objects {
		object := &page.Objects[i]
		data := object.Bytes()
		if object.ID() <= previous || object.Revision() == 0 || len(data) > nodis.MaxQueryPageBytes-size {
			return dst[:start], ErrInvalidCommand
		}
		size += len(data)
		previous = object.ID()
		dst = binary.LittleEndian.AppendUint32(dst, uint32(len(data)))
		dst = append(dst, data...)
	}
	return dst, nil
}

type QueryPageDecoder struct{ objects []nodis.PortableObject }

// DecodeQueryPageInto returns borrowed object bytes and reuses result wrappers.
func DecodeQueryPageInto(data []byte, decoder *QueryPageDecoder) (nodis.QueryPage, error) {
	if decoder == nil || len(data) < queryPageHeaderSize || len(data) > queryPageHeaderSize+4*nodis.MaxQueryLimit+nodis.MaxQueryPageBytes || string(data[:4]) != "NQP1" || data[28] > 1 {
		return nodis.QueryPage{}, ErrInvalidCommand
	}
	count := int(binary.LittleEndian.Uint16(data[29:31]))
	if count > nodis.MaxQueryLimit || count > (len(data)-queryPageHeaderSize)/34 {
		return nodis.QueryPage{}, ErrInvalidCommand
	}
	clear(decoder.objects)
	decoder.objects = decoder.objects[:0]
	page := nodis.QueryPage{Sequence: binary.LittleEndian.Uint64(data[4:12]), PermissionRevision: binary.LittleEndian.Uint64(data[12:20]), Next: nodis.ObjectID(binary.LittleEndian.Uint64(data[20:28])), HasMore: data[28] == 1}
	pos, size := queryPageHeaderSize, 0
	var previous nodis.ObjectID
	for range count {
		if len(data)-pos < 4 {
			return nodis.QueryPage{}, ErrInvalidCommand
		}
		length := uint64(binary.LittleEndian.Uint32(data[pos:]))
		pos += 4
		if length > uint64(len(data)-pos) || length > uint64(nodis.MaxQueryPageBytes-size) {
			return nodis.QueryPage{}, ErrInvalidCommand
		}
		object, err := nodis.ParsePortable(data[pos : pos+int(length)])
		if err != nil || object.ID() <= previous || object.Revision() == 0 {
			return nodis.QueryPage{}, ErrInvalidCommand
		}
		pos += int(length)
		size += int(length)
		previous = object.ID()
		decoder.objects = append(decoder.objects, object)
	}
	if pos != len(data) {
		return nodis.QueryPage{}, ErrInvalidCommand
	}
	page.Objects = decoder.objects
	if !validPageCursor(page) {
		return nodis.QueryPage{}, ErrInvalidCommand
	}
	return page, nil
}

func validPageCursor(page nodis.QueryPage) bool {
	if !page.HasMore {
		return page.Next == 0
	}
	return len(page.Objects) > 0 && page.Next == page.Objects[len(page.Objects)-1].ID()
}
