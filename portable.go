package nodis

import "encoding/binary"

const portableHeaderSize = 30

type PortableObject struct{ data []byte }

func DecodePortable(data []byte) (*PortableObject, error) {
	object, err := ParsePortable(data)
	if err != nil {
		return nil, err
	}
	return &object, nil
}

// ParsePortable validates a borrowed object without allocating a wrapper.
func ParsePortable(data []byte) (PortableObject, error) {
	if len(data) < portableHeaderSize || string(data[:4]) != "NDX1" {
		return PortableObject{}, ErrCorruptData
	}
	object := PortableObject{data: data}
	if !object.valid() {
		return PortableObject{}, ErrCorruptData
	}
	return object, nil
}

func (o *PortableObject) ID() ObjectID {
	if o == nil || len(o.data) < portableHeaderSize {
		return 0
	}
	return ObjectID(binary.LittleEndian.Uint64(o.data[6:14]))
}
func (o *PortableObject) Revision() uint64 {
	if o == nil || len(o.data) < portableHeaderSize {
		return 0
	}
	return binary.LittleEndian.Uint64(o.data[14:22])
}
func (o *PortableObject) Pipeline() PipelineID {
	if o == nil || len(o.data) < portableHeaderSize {
		return 0
	}
	return PipelineID(binary.LittleEndian.Uint64(o.data[22:30]))
}
func (o *PortableObject) Bytes() []byte {
	if o == nil {
		return nil
	}
	return o.data
}

func (o *PortableObject) Field(name string) (ValueView, bool) {
	if o == nil || len(o.data) < portableHeaderSize {
		return ValueView{}, false
	}
	pos := portableHeaderSize
	for range int(binary.LittleEndian.Uint16(o.data[4:6])) {
		field, n, ok := readStringView(o.data[pos:])
		if !ok {
			return ValueView{}, false
		}
		pos += n
		value, used, ok := readValue(o.data[pos:])
		if !ok {
			return ValueView{}, false
		}
		if field == name {
			return value, true
		}
		pos += used
	}
	return ValueView{}, false
}

func (o *PortableObject) String(name string) (string, bool) {
	v, ok := o.Field(name)
	if !ok {
		return "", false
	}
	return v.String()
}
func (o *PortableObject) Int(name string) (int64, bool) {
	v, ok := o.Field(name)
	if !ok {
		return 0, false
	}
	return v.Int()
}
func (o *PortableObject) Bool(name string) (bool, bool) {
	v, ok := o.Field(name)
	if !ok {
		return false, false
	}
	return v.Bool()
}

func (o *PortableObject) valid() bool {
	pos := portableHeaderSize
	for range int(binary.LittleEndian.Uint16(o.data[4:6])) {
		_, n, ok := readStringView(o.data[pos:])
		if !ok {
			return false
		}
		pos += n
		_, used, ok := readValue(o.data[pos:])
		if !ok {
			return false
		}
		pos += used
	}
	return pos == len(o.data)
}

func readStringView(data []byte) (string, int, bool) {
	n, width := binary.Uvarint(data)
	if width <= 0 || n > uint64(len(data)-width) {
		return "", 0, false
	}
	return bytesToString(data[width : width+int(n)]), width + int(n), true
}

func exportObject(doc *Document, data []byte) []byte {
	if len(data) < objectHeaderSize || string(data[:4]) != "NDO1" {
		return nil
	}
	out := make([]byte, 0, len(data)+32)
	out = append(out, 'N', 'D', 'X', '1')
	out = append(out, data[4:objectHeaderSize]...)
	pos := objectHeaderSize
	for range int(binary.LittleEndian.Uint16(data[4:6])) {
		if len(data)-pos < 4 {
			return nil
		}
		id := FieldID(binary.LittleEndian.Uint32(data[pos:]))
		pos += 4
		doc.mu.RLock()
		name, ok := doc.fieldNames[id]
		doc.mu.RUnlock()
		if !ok {
			return nil
		}
		out = appendString(out, name)
		_, used, ok := readValue(data[pos:])
		if !ok {
			return nil
		}
		out = append(out, data[pos:pos+used]...)
		pos += used
	}
	if pos != len(data) {
		return nil
	}
	return out
}

func (o *ObjectView) Portable() ([]byte, error) {
	if o == nil || o.document == nil {
		return nil, ErrNotFound
	}
	data := exportObject(o.document, o.data)
	if data == nil {
		return nil, ErrCorruptData
	}
	return data, nil
}
