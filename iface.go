package nodis

import (
	"github.com/diiyw/nodis/pb"
)

type Synchronizer interface {
	Publish(addr string, fn func(c SyncConn)) error
	Subscribe(addr string, fn func(*pb.Op)) error
}

type SyncConn interface {
	Send(*pb.Op) error
	Wait() error
}
