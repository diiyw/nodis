package storage

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// The generic commit path is shared by real files and this fault-injected disk.
type faultDisk struct {
	bytes.Buffer
	writeCall int
	failWrite int
	writeErr  error
	syncErr   error
	closeErr  error
	synced    []byte
	closed    bool
}

func (disk *faultDisk) Write(data []byte) (int, error) {
	disk.writeCall++
	if disk.writeCall == disk.failWrite {
		n, _ := disk.Buffer.Write(data[:len(data)/2])
		return n, disk.writeErr
	}
	return disk.Buffer.Write(data)
}

func (disk *faultDisk) Sync() error {
	if disk.syncErr != nil {
		return disk.syncErr
	}
	disk.synced = append(disk.synced[:0], disk.Bytes()...)
	return nil
}

func (disk *faultDisk) Close() error {
	disk.closed = true
	return disk.closeErr
}

func TestGenericCommitSyncAndFailures(t *testing.T) {
	injected := errors.New("disk failure")
	for _, stage := range []string{"header", "payload", "checksum", "sync", "success", "unsafe"} {
		t.Run(stage, func(t *testing.T) {
			disk := &faultDisk{writeErr: injected}
			writer := FrameWriter[*faultDisk]{Writer: disk}
			if err := appendRecord(&writer, 1, []byte("committed"), true); err != nil {
				t.Fatal(err)
			}
			committed := bytes.Clone(disk.synced)
			switch stage {
			case "header":
				disk.failWrite = disk.writeCall + 1
			case "payload":
				disk.failWrite = disk.writeCall + 2
			case "checksum":
				disk.failWrite = disk.writeCall + 3
			case "sync":
				disk.syncErr = injected
			}
			err := appendRecord(&writer, 2, []byte("following"), stage != "unsafe")
			if stage == "success" {
				if err != nil || !bytes.Equal(disk.Bytes(), disk.synced) {
					t.Fatal("successful commit was not fully synced", err)
				}
			} else if stage == "unsafe" {
				if err != nil || !bytes.Equal(disk.synced, committed) {
					t.Fatal("unsafe commit unexpectedly synced", err)
				}
			} else if !errors.Is(err, injected) || !bytes.Equal(disk.synced, committed) {
				t.Fatal("failure lost acknowledged bytes or was hidden", err)
			}
			if !bytes.HasPrefix(disk.Bytes(), committed) {
				t.Fatal("commit overwrote previously committed data")
			}
			if stage == "sync" {
				// A failed sync does not prove rollback. Keep the complete record
				// for recovery instead of truncating an uncertain transaction.
				reader := FrameReader[*bytes.Reader]{Reader: bytes.NewReader(disk.Bytes())}
				for sequence := uint64(1); sequence <= 2; sequence++ {
					got, _, err := reader.Read()
					if err != nil || got != sequence {
						t.Fatal("sync failure destroyed recovery evidence", got, err)
					}
				}
			}
		})
	}
}

func TestGenericSyncAndCloseReportsFailures(t *testing.T) {
	syncErr := errors.New("sync failure")
	closeErr := errors.New("close failure")
	for _, failures := range [][2]error{{nil, nil}, {syncErr, nil}, {nil, closeErr}, {syncErr, closeErr}} {
		disk := &faultDisk{syncErr: failures[0], closeErr: failures[1]}
		err := syncAndClose(disk)
		if !disk.closed || (err == nil) != (failures[0] == nil && failures[1] == nil) {
			t.Fatal("close skipped or failure hidden", err)
		}
		for _, failure := range failures {
			if failure != nil && !errors.Is(err, failure) {
				t.Fatal("lost failure", failure, err)
			}
		}
	}
}

func TestDirectoryLockAndOpenFailureCleanup(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "database")
	log, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = log.Close() })
	for _, path := range []string{dir, filepath.Join(dir, "..", "database")} {
		other, err := Open(path)
		if other != nil {
			_ = other.Close()
		}
		if !errors.Is(err, ErrLocked) {
			t.Fatal("second writer was allowed", err)
		}
	}
	alias := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(dir, alias); err != nil {
		t.Fatal(err)
	}
	if other, err := Open(alias); !errors.Is(err, ErrLocked) {
		if other != nil {
			_ = other.Close()
		}
		t.Fatal("symlink bypassed lock", err)
	}
	if err := log.Close(); err != nil {
		t.Fatal(err)
	}
	// A directory in place of the WAL forces Open to fail after acquiring the
	// lock. Repairing the path must allow the next Open to acquire it again.
	wal := filepath.Join(dir, "nodis.wal")
	if err := os.Remove(wal); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(wal, 0o750); err != nil {
		t.Fatal(err)
	}
	if other, err := Open(dir); err == nil {
		_ = other.Close()
		t.Fatal("invalid WAL path accepted")
	}
	if err := os.Remove(wal); err != nil {
		t.Fatal(err)
	}
	log, err = Open(dir)
	if err != nil {
		t.Fatal("failed open leaked directory lock", err)
	}
}

func TestRecoveryAtEveryTornFrameBoundary(t *testing.T) {
	var buffer bytes.Buffer
	writer := FrameWriter[*bytes.Buffer]{Writer: &buffer}
	if err := writer.Append(1, []byte("acknowledged")); err != nil {
		t.Fatal(err)
	}
	boundary := buffer.Len()
	if err := writer.Append(2, []byte("interrupted commit")); err != nil {
		t.Fatal(err)
	}
	frames := buffer.Bytes()
	for cut := boundary; cut < len(frames); cut++ {
		dir := t.TempDir()
		path := filepath.Join(dir, "nodis.wal")
		if err := os.WriteFile(path, frames[:cut], 0o640); err != nil {
			t.Fatal(err)
		}
		log, err := Open(dir)
		if err != nil {
			t.Fatal(err)
		}
		count := 0
		if err := log.Replay(func(sequence uint64, payload []byte) error {
			count++
			if sequence != 1 || string(payload) != "acknowledged" {
				t.Fatalf("cut %d damaged acknowledged commit", cut)
			}
			return nil
		}); err != nil || count != 1 {
			t.Fatalf("cut %d: count %d, %v", cut, count, err)
		}
		if err := log.Append(2, []byte("replacement"), true); err != nil {
			t.Fatal(err)
		}
		if err := log.Close(); err != nil {
			t.Fatal(err)
		}
		log, err = Open(dir)
		if err != nil {
			t.Fatal(err)
		}
		count = 0
		if err := log.Replay(func(sequence uint64, payload []byte) error {
			count++
			if sequence != uint64(count) || count == 2 && string(payload) != "replacement" {
				t.Fatal("append after repair was not recoverable")
			}
			return nil
		}); err != nil || count != 2 {
			t.Fatal("reopen after repair", count, err)
		}
		if err := log.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestWALCorruptionIsNotDiscardedAsTornTail(t *testing.T) {
	for _, stage := range []string{"header", "payload", "checksum", "sequence"} {
		t.Run(stage, func(t *testing.T) {
			dir := t.TempDir()
			var buffer bytes.Buffer
			writer := FrameWriter[*bytes.Buffer]{Writer: &buffer}
			for record := uint64(1); record <= 2; record++ {
				sequence := record
				if stage == "sequence" && record == 2 {
					sequence = 1
				}
				if err := writer.Append(sequence, []byte("record")); err != nil {
					t.Fatal(err)
				}
			}
			data := buffer.Bytes()
			switch stage {
			case "header":
				data[0] ^= 1
			case "payload":
				data[16] ^= 1
			case "checksum":
				data[len(data)-1] ^= 1
			}
			path := filepath.Join(dir, "nodis.wal")
			if err := os.WriteFile(path, data, 0o640); err != nil {
				t.Fatal(err)
			}
			log, err := Open(dir)
			if err != nil {
				t.Fatal(err)
			}
			defer log.Close()
			if err := log.Replay(func(uint64, []byte) error { return nil }); !errors.Is(err, ErrCorrupt) {
				t.Fatal("corruption was accepted", err)
			}
			if err := log.Append(3, []byte("must not overwrite corruption"), true); !errors.Is(err, ErrCorrupt) {
				t.Fatal("failed recovery allowed subsequent writes", err)
			}
			after, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(after, data) {
				t.Fatal("corrupt WAL evidence was truncated", err)
			}
		})
	}
}
