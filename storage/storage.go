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
	ScanKeys(func(*ds.Key) bool) error
}

// Mutation is a value replacement or tombstone, keyed only by Name.
type Mutation struct {
	Key    ds.Key
	Value  ds.Value
	Delete bool
}

// BatchStorage makes Flush atomic, including a pending Clear. Custom backends
// without this interface are retried on failure but cannot promise crash atomicity.
type BatchStorage interface {
	Apply(mutations []Mutation, clear bool) error
}
