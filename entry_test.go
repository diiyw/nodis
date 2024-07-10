package nodis

import (
	"errors"
	"github.com/diiyw/nodis/ds/str"
	"testing"
	"time"
)

func TestValueEntry_EncodeAndDecode(t *testing.T) {
	entry := &ValueEntry{
		Type:       1,
		Expiration: time.Now().Unix(),
		Key:        "testKey",
		Value:      []byte("testValue"),
	}

	encoded := entry.encode()

	decodedEntry := &ValueEntry{}
	err := decodedEntry.decode(encoded)
	if err != nil {
		t.Errorf("Decode failed with error: %v", err)
	}

	if decodedEntry.Type != entry.Type ||
		decodedEntry.Expiration != entry.Expiration ||
		decodedEntry.Key != entry.Key ||
		string(decodedEntry.Value) != string(entry.Value) {
		t.Errorf("Decoded entry does not match original entry")
	}
}

func TestValueEntry_DecodeCorruptedData(t *testing.T) {
	corruptedData := []byte{0, 1, 2, 3}

	entry := &ValueEntry{}
	err := entry.decode(corruptedData)
	if !errors.Is(err, ErrCorruptedData) {
		t.Errorf("Expected ErrCorruptedData, but got %v", err)
	}
}

func TestNewValueEntry(t *testing.T) {
	key := "testKey"
	value := str.NewString()
	value.SetValue([]byte("testValue"))
	expiration := time.Now().Unix()

	entry := newValueEntry(key, value, expiration)

	if entry.Type != uint8(value.Type()) ||
		entry.Expiration != expiration ||
		entry.Key != key ||
		string(entry.Value) != string(value.GetValue()) {
		t.Errorf("NewValueEntry did not correctly initialize entry")
	}
}
