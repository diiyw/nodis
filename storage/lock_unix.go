//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd

package storage

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
)

func lockDirectory(dir string) (*os.File, error) {
	// Keep this inode for the lifetime of the directory. Unlinking a lock file
	// permits another process to lock a different inode under the same name.
	file, err := os.OpenFile(filepath.Join(dir, "nodis.lock"), os.O_CREATE|os.O_RDWR, 0o640)
	if err != nil {
		return nil, err
	}
	for {
		err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err != syscall.EINTR {
			break
		}
	}
	if err != nil {
		_ = file.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
			return nil, ErrLocked
		}
		return nil, err
	}
	return file, nil
}
