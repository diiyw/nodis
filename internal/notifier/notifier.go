package notifier

import (
	"github.com/diiyw/nodis/patch"
	"path/filepath"
)

type Notifier struct {
	pattern []string
	fn      func(op patch.Op)
}

func New(pattern []string, fn func(op patch.Op)) *Notifier {
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
func (w *Notifier) Push(op patch.Op) {
	w.fn(op)
}
