package listener

import (
	"testing"

	"github.com/diiyw/nodis/patch"
)

func TestListener_Matched(t *testing.T) {
	pattern := []string{"test"}
	w := New(pattern, nil)
	if w == nil {
		t.Errorf("NewListener() = %v, want %v", w, "Listener{}")
	}
	if !w.Matched("test") {
		t.Errorf("Matched() = %v, want %v", false, true)
	}
}

func TestListener_Push(t *testing.T) {
	pattern := []string{"test"}
	w := New(pattern, func(op patch.Op) {
		if op.Data.GetKey() != "test" {
			t.Errorf("Push() = %v, want %v", op.Data.GetKey(), "test")
		}
	})
	if w == nil {
		t.Errorf("NewListener() = %v, want %v", w, "Listener{}")
	}
	w.Push(patch.Op{patch.OpTypeSet, &patch.OpSet{Key: "test", Value: []byte("test")}})
}
