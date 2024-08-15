package nodis

import "github.com/diiyw/nodis/patch"

type Channel interface {
	Publish(addr string, fn func(c ChannelConn)) error
	Subscribe(addr string, fn func(op patch.Op)) error
}

type ChannelConn interface {
	Send(op patch.Op) error
	Wait() error
}
