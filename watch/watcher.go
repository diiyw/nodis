package watch

import (
	"path/filepath"

	"github.com/diiyw/nodis/pb"
)

type Watcher struct {
	pattern []string
	fn      func(op *pb.Operation)
}

func NewWatcher(pattern []string, fn func(op *pb.Operation)) *Watcher {
	return &Watcher{
		pattern: pattern,
		fn:      fn,
	}
}

// Matched checks if the key matches the pattern
func (w *Watcher) Matched(key string) bool {
	for _, p := range w.pattern {
		matched, err := filepath.Match(p, key)
		if err != nil {
			continue
		}
		if matched {
			return true
		}
	}
	return false
}

// Push sends the operation to the watcher
func (w *Watcher) Push(op *pb.Operation) {
	w.fn(op)
}
