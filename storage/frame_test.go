package storage

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestGenericFramesZeroAllocations(t *testing.T) {
	var buffer bytes.Buffer
	buffer.Grow(256)
	writer := FrameWriter[*bytes.Buffer]{Writer: &buffer}
	payload := []byte("immutable object bytes")
	if allocs := testing.AllocsPerRun(1000, func() {
		buffer.Reset()
		if err := writer.Append(7, payload); err != nil {
			panic(err)
		}
	}); allocs != 0 {
		t.Fatalf("FrameWriter: %g allocs", allocs)
	}
	frame := append([]byte(nil), buffer.Bytes()...)
	input := bytes.NewReader(frame)
	reader := FrameReader[*bytes.Reader]{Reader: input}
	_, _, _ = reader.Read()
	if allocs := testing.AllocsPerRun(1000, func() {
		input.Reset(frame)
		sequence, body, err := reader.Read()
		if err != nil || sequence != 7 || !bytes.Equal(body, payload) {
			panic("frame changed")
		}
	}); allocs != 0 {
		t.Fatalf("FrameReader: %g allocs", allocs)
	}
}

func TestWALAppendZeroAllocations(t *testing.T) {
	log, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	payload := []byte("small record")
	var sequence uint64
	if allocs := testing.AllocsPerRun(100, func() {
		sequence++
		if err := log.Append(sequence, payload, false); err != nil {
			panic(err)
		}
	}); allocs != 0 {
		t.Fatalf("Log.Append: %g allocs", allocs)
	}
}

type checkpointRecords struct {
	records [][]byte
	fail    error
}

func (source checkpointRecords) WriteSnapshot(writer *SnapshotWriter) error {
	for _, payload := range source.records {
		if err := writer.Append(payload); err != nil {
			return err
		}
	}
	return source.fail
}

func TestCheckpointPublicationAndFailure(t *testing.T) {
	dir := t.TempDir()
	log, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	if err := log.Append(1, []byte("original"), true); err != nil {
		t.Fatal(err)
	}
	injected := errors.New("source failed")
	if err := Checkpoint(log, 1, checkpointRecords{records: [][]byte{[]byte("draft")}, fail: injected}); !errors.Is(err, injected) {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, snapshotName)); !os.IsNotExist(err) {
		t.Fatal("failed checkpoint was published")
	}
	var records int
	if err := log.Replay(func(_ uint64, data []byte) error {
		records++
		if string(data) != "original" {
			t.Fatal("WAL changed")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if records != 1 {
		t.Fatal("failed checkpoint lost WAL")
	}
	if err := Checkpoint(log, 1, checkpointRecords{records: [][]byte{[]byte("saved")}}); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filepath.Join(dir, snapshotName))
	if err != nil {
		t.Fatal(err)
	}
	if err := Checkpoint(log, 1, checkpointRecords{fail: injected}); !errors.Is(err, injected) {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(filepath.Join(dir, snapshotName))
	if !bytes.Equal(before, after) {
		t.Fatal("failed replacement destroyed snapshot")
	}
	sequence, found, err := log.RestoreSnapshot(func(data []byte) error {
		if string(data) != "saved" {
			t.Fatal("bad snapshot record")
		}
		return nil
	})
	if err != nil || !found || sequence != 1 {
		t.Fatalf("restore = %d, %v, %v", sequence, found, err)
	}
	stat, _ := os.Stat(filepath.Join(dir, "nodis.wal"))
	if stat.Size() != 0 {
		t.Fatal("WAL not recycled")
	}
}

func TestSnapshotRejectsTruncationCorruptionAndTrailingBytes(t *testing.T) {
	for _, damage := range []string{"truncate", "checksum", "trailing", "missing-record", "duplicate-record"} {
		t.Run(damage, func(t *testing.T) {
			dir := t.TempDir()
			log, err := Open(dir)
			if err != nil {
				t.Fatal(err)
			}
			defer log.Close()
			if err := Checkpoint(log, 7, checkpointRecords{records: [][]byte{[]byte("record")}}); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(dir, snapshotName)
			data, _ := os.ReadFile(path)
			switch damage {
			case "truncate":
				data = data[:len(data)-1]
			case "checksum":
				data[33] ^= 1
			case "trailing":
				data = append(data, 1)
			case "missing-record":
				// Keep the valid terminator/footer, but omit the whole data frame.
				data = append(data[:16], data[16+20+len("record"):]...)
			case "duplicate-record":
				end := 16 + 20 + len("record")
				frame := append([]byte(nil), data[16:end]...)
				data = append(data[:end], append(frame, data[end:]...)...)
			}
			if err := os.WriteFile(path, data, 0o640); err != nil {
				t.Fatal(err)
			}
			if _, _, err := log.RestoreSnapshot(func([]byte) error { return nil }); !errors.Is(err, ErrCorrupt) {
				t.Fatalf("damaged snapshot = %v", err)
			}
		})
	}
}

func TestGenericFrameShortWrites(t *testing.T) {
	writer := FrameWriter[shortWriter]{Writer: shortWriter{}}
	if err := writer.Append(1, nil); !errors.Is(err, io.ErrShortWrite) {
		t.Fatal(err)
	}
}

func TestStorageFailureStopsFurtherWrites(t *testing.T) {
	for _, operation := range []string{"append-rollback", "checkpoint-recycle"} {
		t.Run(operation, func(t *testing.T) {
			dir := t.TempDir()
			log, err := Open(dir)
			if err != nil {
				t.Fatal(err)
			}
			defer log.Close()
			if err := log.Append(1, []byte("committed"), true); err != nil {
				t.Fatal(err)
			}
			if err := log.file.Close(); err != nil {
				t.Fatal(err)
			}
			// Force write/truncate failures while preserving the committed bytes.
			readOnly, err := os.Open(filepath.Join(dir, "nodis.wal"))
			if err != nil {
				t.Fatal(err)
			}
			log.file, log.writer.Writer = readOnly, readOnly
			if _, err := readOnly.Seek(0, io.SeekEnd); err != nil {
				t.Fatal(err)
			}
			var failure error
			if operation == "append-rollback" {
				failure = log.Append(2, []byte("uncommitted"), true)
			} else {
				failure = Checkpoint(log, 1, checkpointRecords{records: [][]byte{[]byte("committed")}})
			}
			if failure == nil || log.failure == nil {
				t.Fatal("storage failure was not latched")
			}
			if err := readOnly.Close(); err != nil {
				t.Fatal(err)
			}
			writable, err := os.OpenFile(filepath.Join(dir, "nodis.wal"), os.O_RDWR, 0)
			if err != nil {
				t.Fatal(err)
			}
			log.file, log.writer.Writer = writable, writable
			if err := log.Append(2, []byte("must not be written"), false); err != failure {
				t.Fatal("append ignored storage failure", err)
			}
			if err := log.Sync(); err != failure {
				t.Fatal("sync ignored storage failure", err)
			}
			if err := Checkpoint(log, 1, checkpointRecords{}); err != failure {
				t.Fatal("checkpoint ignored storage failure", err)
			}
			if err := log.Close(); err != nil {
				t.Fatal(err)
			}
			reopened, err := Open(dir)
			if err != nil {
				t.Fatal(err)
			}
			defer reopened.Close()
			count := 0
			if err := reopened.Replay(func(sequence uint64, payload []byte) error {
				count++
				if sequence != 1 || string(payload) != "committed" {
					t.Fatal("committed WAL changed")
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
			if count != 1 {
				t.Fatal("committed WAL lost")
			}
			if operation == "checkpoint-recycle" {
				if sequence, found, err := reopened.RestoreSnapshot(func(payload []byte) error {
					if string(payload) != "committed" {
						t.Fatal("published snapshot changed")
					}
					return nil
				}); err != nil || !found || sequence != 1 {
					t.Fatal("published snapshot lost", err)
				}
			}
		})
	}
}

type shortWriter struct{}

func (shortWriter) Write([]byte) (int, error) { return 0, nil }
