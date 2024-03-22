package watch

import (
	"path/filepath"

	"github.com/diiyw/nodis/pb"
)

type Watcher struct {
	pattern []string
	queue   chan *pb.Operation
}

func NewWatcher(pattern []string, capacity int) *Watcher {
	return &Watcher{
		pattern: pattern,
		queue:   make(chan *pb.Operation, capacity),
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

// Push pushs a operation into the watch queue
func (w *Watcher) Push(op *pb.Operation) {
	w.queue <- op
}

// Pop pops a operation from the watch queue
func (w *Watcher) Pop() *pb.Operation {
	return <-w.queue
}
