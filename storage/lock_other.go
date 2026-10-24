//go:build !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd

package storage

import "os"

func lockDirectory(string) (*os.File, error) {
	// Refuse persistent writes rather than silently allowing concurrent writers.
	// In-memory databases do not need a filesystem lock.
	return nil, ErrLockUnsupported
}
