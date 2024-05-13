package notifier

import (
	"testing"

	"github.com/diiyw/nodis/pb"
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
	w := New(pattern, func(op *pb.Operation) {
		if op.Key != "test" {
			t.Errorf("Push() = %v, want %v", op.Key, "test")
		}
	})
	if w == nil {
		t.Errorf("NewNotifier() = %v, want %v", w, "Notifier{}")
	}
	w.Push(&pb.Operation{Key: "test"})
}
