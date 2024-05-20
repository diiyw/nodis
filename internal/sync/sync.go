package sync

import (
	"github.com/diiyw/nodis/pb"
)

type Synchronizer interface {
	Publish(addr string, fn func(c Conn)) error
	Subscribe(addr string, fn func(*pb.Op)) error
}

type Conn interface {
	Send(*pb.Op) error
	Wait() error
}
