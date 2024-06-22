package notifier

import (
	"testing"

	"github.com/diiyw/nodis/patch"
)

func TestNotifier_Matched(t *testing.T) {
	pattern := []string{"test"}
	w := New(pattern, nil)
	if w == nil {
		t.Errorf("NewNotifier() = %v, want %v", w, "Notifier{}")
	}
	if !w.Matched("test") {
		t.Errorf("Matched() = %v, want %v", false, true)
	}
}

func TestNotifier_Push(t *testing.T) {
	pattern := []string{"test"}
	w := New(pattern, func(op patch.Op) {
		if op.Data.GetKey() != "test" {
			t.Errorf("Push() = %v, want %v", op.Data.GetKey(), "test")
		}
	})
	if w == nil {
		t.Errorf("NewNotifier() = %v, want %v", w, "Notifier{}")
	}
	w.Push(patch.Op{patch.OpTypeSet, &patch.OpSet{Key: "test", Value: []byte("test")}})
}
