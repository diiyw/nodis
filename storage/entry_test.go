package storage

import (
	"reflect"
	"testing"

	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/ds/str"
)

func TestValueEntry_EncodeDecode(t *testing.T) {
	value := str.NewString()
	value.Set([]byte("test value"))
	expiration := int64(1234567890)
	entry := NewValueEntry(value, expiration)

	encoded := entry.encode()

	decoded := &Entry{}
	err := decoded.decode(encoded)
	if err != nil {
		t.Errorf("decode failed: %v", err)
	}

	if decoded.Type != entry.Type {
		t.Errorf("decoded Type = %v, want %v", decoded.Type, entry.Type)
	}

	if decoded.Expiration != entry.Expiration {
		t.Errorf("decoded Expiration = %v, want %v", decoded.Expiration, entry.Expiration)
	}

	if !reflect.DeepEqual(decoded.Value, entry.Value) {
		t.Errorf("decoded Value = %v, want %v", decoded.Value, entry.Value)
	}
}

func TestParseValue(t *testing.T) {
	value := str.NewString()
	value.Set([]byte("test value"))
	expiration := int64(1234567890)
	entry := NewValueEntry(value, expiration)
	encoded := entry.encode()

	parsedEntry, err := parseEntry(encoded)
	if err != nil {
		t.Errorf("parseEntry failed: %v", err)
		return
	}
	if !reflect.DeepEqual(parsedEntry, entry) {
		t.Errorf("parsedEntry = %v, want %v", parsedEntry, entry)
	}
	parsedValue, err := parseValue(ds.String, entry.Value)
	if err != nil {
		t.Errorf("parseValue failed: %v", err)
	}
	if parsedValue.Type() != ds.ValueType(entry.Type) {
		t.Errorf("parsedValue.Type() = %v, want %v", parsedValue.Type(), entry.Type)
	}

	if !reflect.DeepEqual(parsedValue.GetValue(), entry.Value) {
		t.Errorf("parsedValue.GetValue() = %v, want %v", parsedValue.GetValue(), entry.Value)
	}
}
