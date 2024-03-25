package watch

import (
	"testing"

	"github.com/diiyw/nodis/pb"
)

func TestWatcher_Matched(t *testing.T) {
	pattern := []string{"test"}
	w := NewWatcher(pattern, nil)
	if w == nil {
		t.Errorf("NewWatcher() = %v, want %v", w, "Watcher{}")
	}
	if !w.Matched("test") {
		t.Errorf("Matched() = %v, want %v", false, true)
	}
}

func TestWatcher_Push(t *testing.T) {
	pattern := []string{"test"}
	w := NewWatcher(pattern, func(op *pb.Operation) {
		if op.Key != "test" {
			t.Errorf("Push() = %v, want %v", op.Key, "test")
		}
	})
	if w == nil {
		t.Errorf("NewWatcher() = %v, want %v", w, "Watcher{}")
	}
	w.Push(&pb.Operation{Key: "test"})
}
