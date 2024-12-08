package storage

import (
	"errors"

	"github.com/diiyw/nodis/ds"
)

var (
	ErrKeyNotFound = errors.New("key not found")
)

type Storage interface {
	Init() error
	Get(key *ds.Key) (ds.Value, error)
	Set(key *ds.Key, value ds.Value) error
	Delete(key *ds.Key) error
	Clear() error
	Close() error
	Snapshot() error
	ScanKeys(func(*ds.Key) bool)
}
