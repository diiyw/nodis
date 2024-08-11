package storage

import (
	"errors"

	"github.com/diiyw/nodis/ds"
)

var (
	ErrKeyNotFound = errors.New("key not found")
)

type Storage interface {
	Open() error
	Get(key string) (ds.Value, error)
	Put(key string, value ds.Value, expiration int64) error
	Delete(key string) error
	GetIndex() []byte
	PutIndex(index []byte) error
	Reset() error
	Close() error
}
