package nodis

import (
	"encoding/binary"
	"sort"
	"unsafe"
)

const objectHeaderSize = 30

type fieldValue struct {
	id    FieldID
	value Value
}

type ObjectView struct {
	document *Document
	data     []byte
}

func (o *ObjectView) ID() ObjectID {
	if o == nil || len(o.data) < objectHeaderSize {
		return 0
	}
	return ObjectID(binary.LittleEndian.Uint64(o.data[6:14]))
}

func (o *ObjectView) Revision() uint64 {
	if o == nil || len(o.data) < objectHeaderSize {
		return 0
	}
	return binary.LittleEndian.Uint64(o.data[14:22])
}

func (o *ObjectView) Pipeline() PipelineID {
	if o == nil || len(o.data) < objectHeaderSize {
		return 0
	}
	return PipelineID(binary.LittleEndian.Uint64(o.data[22:30]))
}

func (o *ObjectView) Field(name string) (ValueView, bool) {
	if o == nil || o.document == nil {
		return ValueView{}, false
	}
	id, ok := o.document.fieldID(name)
	if !ok {
		return ValueView{}, false
	}
	return o.fieldID(id)
}

func (o *ObjectView) fieldID(id FieldID) (ValueView, bool) {
	if len(o.data) < objectHeaderSize || string(o.data[:4]) != "NDO1" {
		return ValueView{}, false
	}
	count := int(binary.LittleEndian.Uint16(o.data[4:6]))
	pos := objectHeaderSize
	for range count {
		if len(o.data)-pos < 4 {
			return ValueView{}, false
		}
		fieldID := FieldID(binary.LittleEndian.Uint32(o.data[pos:]))
		pos += 4
		value, n, ok := readValue(o.data[pos:])
		if !ok {
			return ValueView{}, false
		}
		if fieldID == id {
			return value, true
		}
		if fieldID > id {
			return ValueView{}, false
		}
		pos += n
	}
	return ValueView{}, false
}

func (o *ObjectView) String(name string) (string, bool) {
	v, ok := o.Field(name)
	if !ok {
		return "", false
	}
	return v.String()
}
func (o *ObjectView) Int(name string) (int64, bool) {
	v, ok := o.Field(name)
	if !ok {
		return 0, false
	}
	return v.Int()
}
func (o *ObjectView) Uint(name string) (uint64, bool) {
	v, ok := o.Field(name)
	if !ok {
		return 0, false
	}
	return v.Uint()
}
func (o *ObjectView) Float(name string) (float64, bool) {
	v, ok := o.Field(name)
	if !ok {
		return 0, false
	}
	return v.Float()
}
func (o *ObjectView) Bool(name string) (bool, bool) {
	v, ok := o.Field(name)
	if !ok {
		return false, false
	}
	return v.Bool()
}
func (o *ObjectView) Bytes(name string) ([]byte, bool) {
	v, ok := o.Field(name)
	if !ok {
		return nil, false
	}
	return v.Bytes()
}

func encodeObject(dst []byte, id ObjectID, revision uint64, pipeline PipelineID, fields []fieldValue) []byte {
	sort.Slice(fields, func(i, j int) bool { return fields[i].id < fields[j].id })
	dst = append(dst, 'N', 'D', 'O', '1')
	dst = binary.LittleEndian.AppendUint16(dst, uint16(len(fields)))
	dst = binary.LittleEndian.AppendUint64(dst, uint64(id))
	dst = binary.LittleEndian.AppendUint64(dst, revision)
	dst = binary.LittleEndian.AppendUint64(dst, uint64(pipeline))
	for i := range fields {
		dst = binary.LittleEndian.AppendUint32(dst, uint32(fields[i].id))
		dst = appendValue(dst, fields[i].value)
	}
	return dst
}

func filterObject(doc *Document, data []byte, allowed func(FieldID) bool) []byte {
	if len(data) < objectHeaderSize {
		return nil
	}
	out := make([]byte, 0, len(data))
	out = append(out, data[:objectHeaderSize]...)
	count := int(binary.LittleEndian.Uint16(data[4:6]))
	visible := uint16(0)
	pos := objectHeaderSize
	for range count {
		if len(data)-pos < 4 {
			return nil
		}
		start := pos
		id := FieldID(binary.LittleEndian.Uint32(data[pos:]))
		pos += 4
		_, n, ok := readValue(data[pos:])
		if !ok {
			return nil
		}
		pos += n
		if allowed(id) {
			out = append(out, data[start:pos]...)
			visible++
		}
	}
	binary.LittleEndian.PutUint16(out[4:6], visible)
	return out
}

func bytesToString(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	return unsafe.String(unsafe.SliceData(data), len(data))
}
