package notifier

import (
	"path/filepath"

	"github.com/diiyw/nodis/pb"
)

type Notifier struct {
	pattern []string
	fn      func(op *pb.Operation)
}

func New(pattern []string, fn func(op *pb.Operation)) *Notifier {
	return &Notifier{
		pattern: pattern,
		fn:      fn,
	}
}

// Matched checks if the key matches the pattern
func (w *Notifier) Matched(key string) bool {
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
func (w *Notifier) Push(op *pb.Operation) {
	w.fn(op)
}
