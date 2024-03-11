package nodis

import "time"

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

// Options represents the configuration options for the database.
type Options struct {
	// Path is the path to the database.
	Path string

	// RecycleDuration is the interval at which the database is recycled .
	// This is useful for reducing the risk of data loss in the event of a crash.
	// It is also used for refreshing hot keys.
	RecycleDuration time.Duration

	// FileSize is the size of each file. The default value is 1GB.
	FileSize int64

	// SnapshotDuration is the interval at which the database is snapshotted.
	// Default 0 for disabling snapshot. and you can call Snapshot manually.
	SnapshotDuration time.Duration
}

var DefaultOptions = &Options{
	Path:            "data",
	FileSize:        GB,
	RecycleDuration: 60 * time.Second,
}
