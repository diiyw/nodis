package storage

import (
	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/ds/set"
	"github.com/diiyw/nodis/ds/str"
	"strings"
	"testing"
)

func TestSetSerializationBoundaries(t *testing.T) {
	for _, size := range []int{0, 1, 63, 64, 127, 128, 8191, 8192} {
		s := set.NewSet()
		member := strings.Repeat("x", size)
		s.SAdd(member)
		restored, err := DecodeValue(EncodeValue(s))
		if err != nil {
			t.Fatal(size, err)
		}
		if !restored.(*set.Set).SIsMember(member) {
			t.Fatal("corrupt member", size)
		}
	}
}
func TestCorruptRecordsReturnErrors(t *testing.T) {
	for _, data := range [][]byte{nil, {255}, {byte(ds.Set), 128}, {byte(ds.Set), 127}, {byte(ds.Set), 2}, {byte(ds.List), 127}, {byte(ds.Hash), 127}, {byte(ds.ZSet), 127}} {
		if _, err := DecodeValue(data); err == nil {
			t.Fatalf("accepted %v", data)
		}
	}
}
func TestMemoryOwnsFlushedValues(t *testing.T) {
	m := NewMemory()
	key := ds.NewKey("k", 123)
	v := str.NewString()
	input := []byte("old")
	v.Set(input)
	if err := m.Set(key, v); err != nil {
		t.Fatal(err)
	}
	copy(input, "new")
	key.Expiration = 456
	stored, err := m.Get(key)
	if err != nil || string(stored.(*str.String).Get()) != "old" {
		t.Fatal("flushed value aliases live value", err)
	}
	m.ScanKeys(func(k *ds.Key) bool {
		if k.Expiration != 123 {
			t.Error("metadata alias")
		}
		return true
	})
}
