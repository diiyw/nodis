package nodis

import "time"

type Options struct {
	// Path is the path to the database
	Path string
	// SyncInterval is the interval at which the database is synced to disk
	SyncInterval time.Duration
}

var DefaultOptions = Options{
	SyncInterval: 60 * time.Second,
}
