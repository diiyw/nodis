package nodis

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/diiyw/nodis/storage"
)

func TestDurabilityDefaults(t *testing.T) {
	for _, config := range []struct {
		name    string
		options Options
		durable bool
	}{
		{name: "default", durable: true},
		{name: "legacy-true", options: Options{Durable: true}, durable: true},
		{name: "legacy-false", options: Options{Durable: false}, durable: true},
		{name: "explicit-unsafe", options: Options{UnsafeNoSync: true}},
	} {
		t.Run(config.name, func(t *testing.T) {
			config.options.Dir = t.TempDir()
			db, err := Open(config.options)
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			if db.durable != config.durable {
				t.Fatal("unexpected commit durability", db.durable)
			}
		})
	}
	if _, err := Open(Options{Durable: true, UnsafeNoSync: true}); !errors.Is(err, ErrInvalidValue) {
		t.Fatal("conflicting durability options accepted", err)
	}
}

func TestRecoveryFailureReleasesDirectoryLock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nodis.wal")
	if err := os.WriteFile(path, bytes.Repeat([]byte{0xff}, 20), 0o640); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if db, err := Open(Options{Dir: dir}); !errors.Is(err, storage.ErrCorrupt) {
			if db != nil {
				_ = db.Close()
			}
			t.Fatal("corruption error was hidden or recovery leaked the lock", err)
		}
	}
}

func TestCrashRecoveryProcess(t *testing.T) {
	if dir := os.Getenv("NODIS_CRASH_TEST_DIR"); dir != "" {
		actor := Principal{ID: "admin"}
		db, err := Open(Options{Dir: dir})
		if err != nil {
			t.Fatal(err)
		}
		doc, err := db.CreateDocument(actor, "survives")
		if err != nil {
			t.Fatal(err)
		}
		putCheckpointObject(t, db, actor, doc.id, map[string]any{"value": "before-checkpoint"})
		mode := os.Getenv("NODIS_CRASH_TEST_MODE")
		if mode != "wal" {
			if err := db.Checkpoint(); err != nil {
				t.Fatal(err)
			}
		}
		if mode != "snapshot" {
			putCheckpointObject(t, db, actor, doc.id, map[string]any{"value": "after-checkpoint"})
		}
		// Tell the parent every preceding commit was acknowledged, then wait to
		// be killed. Deliberately never call Close or Flush in this process.
		fmt.Println("committed")
		var signal [1]byte
		_, _ = os.Stdin.Read(signal[:])
		os.Exit(2)
	}
	for _, mode := range []string{"wal", "snapshot", "snapshot-and-wal"} {
		t.Run(mode, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), "new", "database")
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestCrashRecoveryProcess$")
			command.Env = append(os.Environ(), "NODIS_CRASH_TEST_DIR="+dir, "NODIS_CRASH_TEST_MODE="+mode)
			var stderr bytes.Buffer
			command.Stderr = &stderr
			stdout, err := command.StdoutPipe()
			if err != nil {
				t.Fatal(err)
			}
			stdin, err := command.StdinPipe()
			if err != nil {
				t.Fatal(err)
			}
			defer stdin.Close()
			if err := command.Start(); err != nil {
				t.Fatal(err)
			}
			defer func() {
				_ = command.Process.Kill()
				_ = command.Wait()
			}()
			line, err := bufio.NewReader(stdout).ReadString('\n')
			if err != nil || line != "committed\n" {
				t.Fatal("child failed before acknowledgement", line, err)
			}
			if other, err := Open(Options{Dir: dir}); !errors.Is(err, storage.ErrLocked) {
				if other != nil {
					_ = other.Close()
				}
				t.Fatal("cross-process writer lock failed", err)
			}
			if err := command.Process.Kill(); err != nil {
				t.Fatal(err)
			}
			if err := command.Wait(); err == nil {
				t.Fatal("child unexpectedly exited cleanly", stderr.String())
			}
			db, err := Open(Options{Dir: dir})
			if err != nil {
				t.Fatal("cannot recover after process kill", err)
			}
			defer db.Close()
			actor := Principal{ID: "admin"}
			doc, err := db.Document(actor, "survives")
			if err != nil {
				t.Fatal(err)
			}
			values := []string{"before-checkpoint"}
			if mode != "snapshot" {
				values = append(values, "after-checkpoint")
			}
			for i, expected := range values {
				object, err := db.Get(actor, doc.id, ObjectID(i+1))
				if err != nil {
					t.Fatal("acknowledged object lost", err)
				}
				if actual, ok := object.String("value"); !ok || actual != expected {
					t.Fatal("recovered value changed", actual, expected)
				}
			}
			putCheckpointObject(t, db, actor, doc.id, map[string]any{"value": "after-recovery"})
			if err := db.Checkpoint(); err != nil {
				t.Fatal("recovered database cannot checkpoint", err)
			}
		})
	}
}
