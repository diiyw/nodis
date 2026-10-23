package nodis

import "testing"

func TestScalarLookupHasZeroAllocations(t *testing.T) {
	doc := newDocument(nil, 1, "test")
	id, _ := doc.internField("age")
	data := encodeObject(make([]byte, 0, 64), 1, 1, 0, []fieldValue{{id: id, value: Value{kind: KindInt, num: 42}}})
	view := &ObjectView{document: doc, data: data}
	if allocs := testing.AllocsPerRun(1000, func() {
		value, ok := view.Int("age")
		if !ok || value != 42 {
			panic("bad value")
		}
	}); allocs != 0 {
		t.Fatalf("scalar lookup allocated: %v", allocs)
	}
}
