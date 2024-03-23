package watch

import (
	"github.com/diiyw/nodis/pb"
	"testing"
)

func TestWatcher_Matched(t *testing.T) {
	pattern := []string{"test"}
	w := NewWatcher(pattern, 0)
	if w == nil {
		t.Errorf("NewWatcher() = %v, want %v", w, "Watcher{}")
	}
	if !w.Matched("test") {
		t.Errorf("Matched() = %v, want %v", false, true)
	}
}

func TestWatcher_Push(t *testing.T) {
	pattern := []string{"test"}
	w := NewWatcher(pattern, 1)
	if w == nil {
		t.Errorf("NewWatcher() = %v, want %v", w, "Watcher{}")
	}
	w.Push(nil)
	op := w.Pop()
	if op != nil {
		t.Errorf("Pop() = %v, want %v", op, nil)
	}
	w.Push(&pb.Operation{
		Key: "test",
	})
	op = w.Pop()
	if op == nil {
		t.Errorf("Pop() = %v, want %v", op, "Operation{}")
	}
	if op.Key != "test" {
		t.Errorf("Pop() = %v, want %v", op.Key, "test")
	}
}
