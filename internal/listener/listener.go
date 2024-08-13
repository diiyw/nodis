package listener

import (
	"path/filepath"

	"github.com/diiyw/nodis/patch"
)

type Listener struct {
	pattern []string
	fn      func(op patch.Op)
}

func New(pattern []string, fn func(op patch.Op)) *Listener {
	return &Listener{
		pattern: pattern,
		fn:      fn,
	}
}

// Matched checks if the key matches the pattern
func (w *Listener) Matched(key string) bool {
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
func (w *Listener) Push(op patch.Op) {
	w.fn(op)
}
