package nodis

import "github.com/diiyw/nodis/patch"

type Synchronizer interface {
	Publish(addr string, fn func(c SyncConn)) error
	Subscribe(addr string, fn func(op patch.Op)) error
}

type SyncConn interface {
	Send(op patch.Op) error
	Wait() error
}
