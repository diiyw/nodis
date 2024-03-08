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

	// TidyDuration is the interval at which the database is tidied.
	// This is useful for reducing the risk of data loss in the event of a crash.
	// It is also used for refreshing hot keys.
	TidyDuration time.Duration

	// ChunkSize is the size of each chunk. The default value is 512MB.
	ChunkSize int64

	// SnapshotDuration is the interval at which the database is snapshotted.
	SnapshotDuration time.Duration
}

var DefaultOptions = &Options{
	Path:             "data",
	ChunkSize:        512 * MB,
	TidyDuration:     60 * time.Second,
	SnapshotDuration: time.Hour,
}
