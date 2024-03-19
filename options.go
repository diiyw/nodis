package nodis

import "time"

const (
	FileSizeKB = 1024
	FileSizeMB = 1024 * FileSizeKB
	FileSizeGB = 1024 * FileSizeMB

	HotDataMode = 0
	MemoryMode  = 1
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

	// In memory mode, the database will not write to disk and will not be able to recover after a crash.
	Mode uint8
}

var DefaultOptions = &Options{
	Path:            "data",
	FileSize:        FileSizeGB,
	RecycleDuration: 60 * time.Second,
}
