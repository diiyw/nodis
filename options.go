package nodis

import (
	"time"

	"github.com/diiyw/nodis/fs"
)

const (
	FileSizeKB = 1024
	FileSizeMB = 1024 * FileSizeKB
	FileSizeGB = 1024 * FileSizeMB
)

// Options represents the configuration options for the database.
type Options struct {
	// Path is the path to the database.
	Path string

	// GCDuration is the interval which the database is flushing unused keys to disk.
	// This is useful for reducing the risk of data loss in the event of a crash.
	// It is also used for refreshing hot keys.
	GCDuration time.Duration

	// FileSize is the size of each file. The default value is 1GB.
	FileSize int64

	// SnapshotDuration is the interval at which the database is snapshot.
	// Default 0 for disabling snapshot. and you can call Snapshot manually.
	SnapshotDuration time.Duration

	// Filesystem is the filesystem to use. The default is the memory filesystem.
	Filesystem fs.Fs

	// Synchronizer is the synchronizer to use. The default is nil and no synchronization is performed.
	Synchronizer Synchronizer
}

var DefaultOptions = &Options{
	Path:       "data",
	FileSize:   FileSizeGB,
	GCDuration: 60 * time.Second,
}
