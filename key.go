package nodis

import "time"

type Key struct {
	Type string
	TTL  int64
}

func (k Key) Valid() bool {
	return k.TTL == 0 || k.TTL > time.Now().Unix()
}
