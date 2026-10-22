package listener

import (
	"github.com/diiyw/nodis/internal/glob"
	"github.com/diiyw/nodis/patch"
	"log"
	"sync"
)

// Each listener owns one FIFO worker. Never invoke user code under database locks.
type Listener struct {
	OnStop      func() error
	queuedBytes int
	Batch       bool
	pattern     []string
	fn          func(patch.Op)
	mu          sync.Mutex
	wake        *sync.Cond
	queue       []patch.Op
	stopped     bool
}

func New(pattern []string, fn func(patch.Op)) *Listener {
	w := &Listener{pattern: append([]string(nil), pattern...), fn: fn}
	w.wake = sync.NewCond(&w.mu)
	go w.run()
	return w
}
func (w *Listener) Matched(key string) bool {
	for _, p := range w.pattern {
		if glob.Match(p, key) {
			return true
		}
	}
	return false
}
func eventSize(op patch.Op) int {
	if batch, ok := op.Data.(*patch.OpBatch); ok {
		size := 21
		for _, frame := range batch.Frames {
			size += 10 + len(frame)
		}
		return size
	}
	return 0
}
func (w *Listener) Push(op patch.Op) {
	w.mu.Lock()
	if w.stopped {
		w.mu.Unlock()
		return
	}
	size := eventSize(op)
	if w.Batch && (len(w.queue) >= 1024 || size > patch.MaxFrameSize-w.queuedBytes) {
		w.mu.Unlock()
		w.Stop()
		return
	}
	w.queue = append(w.queue, op)
	w.queuedBytes += size
	w.wake.Signal()
	w.mu.Unlock()
}
func (w *Listener) Stop() {
	w.mu.Lock()
	if w.stopped {
		w.mu.Unlock()
		return
	}
	w.stopped = true
	clear(w.queue)
	w.queue = nil
	w.queuedBytes = 0
	w.wake.Broadcast()
	closeFn := w.OnStop
	w.mu.Unlock()
	if closeFn != nil {
		_ = closeFn()
	}
}
func (w *Listener) run() {
	for {
		w.mu.Lock()
		for len(w.queue) == 0 && !w.stopped {
			w.wake.Wait()
		}
		if w.stopped {
			w.mu.Unlock()
			return
		}
		op := w.queue[0]
		w.queuedBytes -= eventSize(op)
		w.queue[0] = patch.Op{}
		w.queue = w.queue[1:]
		w.mu.Unlock()
		func() {
			defer func() {
				if p := recover(); p != nil {
					log.Println("nodis watcher:", p)
				}
			}()
			w.fn(op)
		}()
	}
}

func (w *Listener) Patterns() []string { return w.pattern }
