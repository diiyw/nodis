package nodis

import "testing"

func TestNestedValueRoundTrip(t *testing.T) {
	value, err := ValueOf(map[string]any{
		"items": []Value{{kind: KindInt, num: 7}, {kind: KindString, text: "eight"}},
		"raw":   []byte{1, 2, 3},
	})
	if err != nil {
		t.Fatal(err)
	}
	encoded := AppendValue(nil, value)
	decoded, used, err := DecodeValue(encoded)
	if err != nil || used != len(encoded) || decoded.Kind() != KindObject {
		t.Fatalf("decode = %v, %d, %v", decoded.Kind(), used, err)
	}
	if len(decoded.obj) != 2 || decoded.obj[0].Name != "items" || len(decoded.obj[0].Value.list) != 2 {
		t.Fatal("nested object changed")
	}
}
