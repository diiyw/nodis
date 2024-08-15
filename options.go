package nodis

import (
	"github.com/diiyw/nodis/storage"
	"time"
)

const (
	FileSizeKB = 1024
	FileSizeMB = 1024 * FileSizeKB
	FileSizeGB = 1024 * FileSizeMB
)

// Options represents the configuration options for the database.
type Options struct {
	// GCDuration is the interval which the database is flushing unused keys to disk.
	// This is useful for reducing the risk of data loss in the event of a crash.
	// It is also used for refreshing hot keys.
	GCDuration time.Duration

	// SnapshotDuration is the interval at which the database is snapshot.
	// Default 0 for disabling snapshot. and you can call Snapshot manually.
	SnapshotDuration time.Duration

	// Storage is the storage to use. The default is the disk storage.
	Storage storage.Storage

	// Channel is the Pub/Sub channel.
	Channel Channel
}

var DefaultOptions = &Options{
	GCDuration: 60 * time.Second,
}
