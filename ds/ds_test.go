package ds

import (
	"testing"
)

func TestKey_EncodeDecode(t *testing.T) {
	key := NewKey("test key", 100)
	encoded := key.Encode()
	decoded := &Key{}
	decoded, err := DecodeKey(encoded)
	if err != nil {
		t.Errorf("decode failed: %v", err)
	}

	if decoded.Name != key.Name {
		t.Errorf("decoded = %v, want %v", decoded, key)
	}
}
