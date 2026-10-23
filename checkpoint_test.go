package nodis

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/diiyw/nodis/storage"
)

func putCheckpointObject(t testing.TB, db *DB, actor Principal, document DocumentID, fields map[string]any) ObjectID {
	t.Helper()
	tx := db.Begin(actor)
	draft, err := tx.Create(document)
	if err != nil {
		t.Fatal(err)
	}
	for name, value := range fields {
		if err := draft.Set(name, value); err != nil {
			t.Fatal(err)
		}
	}
	id, err := draft.Save()
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	return id
}

func TestCheckpointPreservesIDsACLsPipelineAndFollowingWrites(t *testing.T) {
	dir := t.TempDir()
	admin := Principal{ID: "admin"}
	reader := Principal{ID: "reader"}
	db, err := Open(Options{Dir: dir, Durable: true})
	if err != nil {
		t.Fatal(err)
	}
	doc, _ := db.CreateDocument(admin, "original")
	deleted := putCheckpointObject(t, db, admin, doc.id, map[string]any{"unused": true, "secret": "old"})
	if err := db.RenameDocument(admin, doc.id, "retained"); err != nil {
		t.Fatal(err)
	}
	if err := db.RegisterPipeline(3, PipelineFunc(func(context.Context, *PipelineContext) error { return nil })); err != nil {
		t.Fatal(err)
	}
	if err := doc.SetDefaultPipeline(admin, 3); err != nil {
		t.Fatal(err)
	}
	id := putCheckpointObject(t, db, admin, doc.id, map[string]any{"name": "Ada", "secret": "private"})
	view, _ := db.Get(admin, doc.id, id)
	before := append([]byte(nil), view.data...)
	secretID, _ := doc.FieldID("secret")
	unusedID, _ := doc.FieldID("unused")
	tx := db.Begin(admin)
	_ = tx.Delete(doc.id, deleted, 1)
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	allow := Rule{SubjectKind: SubjectUser, Subject: reader.ID, Resource: Resource{Scope: ScopeDocument, Document: doc.id}, Actions: ActionRead | ActionDiscover | ActionSync}
	if err := db.Grant(admin, allow); err != nil {
		t.Fatal(err)
	}
	deny := Rule{SubjectKind: SubjectUser, Subject: reader.ID, Resource: Resource{Scope: ScopeField, Document: doc.id, Field: secretID}, Actions: ActionRead}
	if err := db.Deny(admin, deny); err != nil {
		t.Fatal(err)
	}
	// Retain high-water counters even when the highest IDs have been deleted.
	removed, _ := db.CreateDocument(admin, "removed")
	highObject := putCheckpointObject(t, db, admin, removed.id, map[string]any{"x": 1})
	if err := db.DeleteDocument(admin, removed.id); err != nil {
		t.Fatal(err)
	}
	sequence, permission := db.Sequence(), db.PermissionRevision()
	if err := db.Checkpoint(); err != nil {
		t.Fatal(err)
	}
	wal, _ := os.Stat(filepath.Join(dir, "nodis.wal"))
	if wal.Size() != 0 {
		t.Fatal("checkpoint did not compact WAL")
	}
	if db.Sequence() != sequence {
		t.Fatal("checkpoint emitted a user mutation")
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = Open(Options{Dir: dir, Durable: true})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if db.Sequence() != sequence || db.PermissionRevision() != permission {
		t.Fatal("snapshot counters changed")
	}
	restored, err := db.Document(reader, "retained")
	if err != nil {
		t.Fatal(err)
	}
	if field, _ := restored.FieldID("secret"); field != secretID {
		t.Fatal("secret field ID changed")
	}
	if field, _ := restored.FieldID("unused"); field != unusedID {
		t.Fatal("deleted field dictionary entry lost")
	}
	visible, err := db.Get(reader, restored.id, id)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := visible.String("secret"); ok {
		t.Fatal("field ACL lost in checkpoint")
	}
	full, _ := db.Get(admin, restored.id, id)
	if !bytes.Equal(before, full.data) || full.Pipeline() != 3 || restored.defaultPipeline != 3 {
		t.Fatal("object bytes or pipeline changed")
	}
	newDoc, err := db.CreateDocument(admin, "after")
	if err != nil || newDoc.id <= removed.id {
		t.Fatal("document ID reused", err)
	}
	newID := putCheckpointObject(t, db, admin, newDoc.id, map[string]any{"fresh": true})
	if newID <= highObject {
		t.Fatal("object ID reused")
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = Open(Options{Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Get(admin, newDoc.id, newID); err != nil {
		t.Fatal("post-checkpoint WAL did not replay", err)
	}
	if _, err := db.Document(admin, "removed"); !errors.Is(err, ErrNotFound) {
		t.Fatal("deleted document resurrected")
	}
}

func TestCheckpointCrashBeforeWALRecycle(t *testing.T) {
	dir := t.TempDir()
	admin := Principal{ID: "admin"}
	db, _ := Open(Options{Dir: dir})
	doc, _ := db.CreateDocument(admin, "docs")
	id := putCheckpointObject(t, db, admin, doc.id, map[string]any{"name": "checkpoint"})
	oldWAL, _ := os.ReadFile(filepath.Join(dir, "nodis.wal"))
	if err := db.Checkpoint(); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()
	// Simulate a crash after snapshot publication, before WAL truncation.
	if err := os.WriteFile(filepath.Join(dir, "nodis.wal"), oldWAL, 0o640); err != nil {
		t.Fatal(err)
	}
	db, err := Open(Options{Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	view, err := db.Get(admin, doc.id, id)
	if err != nil {
		t.Fatal(err)
	}
	if name, _ := view.String("name"); name != "checkpoint" {
		t.Fatal("duplicate replay changed state")
	}
	putCheckpointObject(t, db, admin, doc.id, map[string]any{"name": "next"})
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = Open(Options{Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if len(db.objects) != 2 {
		t.Fatal("reopen with overlapping WAL lost objects")
	}
}

func TestConcurrentCheckpointAndCommits(t *testing.T) {
	dir := t.TempDir()
	admin := Principal{ID: "admin"}
	db, _ := Open(Options{Dir: dir})
	doc, _ := db.CreateDocument(admin, "concurrent")
	var wg sync.WaitGroup
	wg.Go(func() {
		for range 10 {
			if err := db.Checkpoint(); err != nil {
				t.Error(err)
				return
			}
		}
	})
	for range 3 {
		wg.Go(func() {
			for range 20 {
				tx := db.Begin(admin)
				draft, err := tx.Create(doc.id)
				if err != nil {
					t.Error(err)
					return
				}
				_ = draft.Set("value", int64(42))
				id, _ := draft.Save()
				if err := tx.Commit(); err != nil {
					t.Error(err)
					return
				}
				if _, err := db.Get(admin, doc.id, id); err != nil {
					t.Error(err)
					return
				}
			}
		})
	}
	wg.Wait()
	_ = db.Close()
	db, err := Open(Options{Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if len(db.objects) != 60 {
		t.Fatalf("recovered %d objects", len(db.objects))
	}
}

func TestPeriodicCheckpointAndMemoryOnly(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(Options{Dir: dir, CheckpointInterval: 5 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	_, _ = db.CreateDocument(Principal{ID: "admin"}, "periodic")
	deadline := time.Now().Add(2 * time.Second)
	for {
		if _, err := os.Stat(filepath.Join(dir, "nodis.snapshot")); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("periodic checkpoint did not run")
		}
		time.Sleep(time.Millisecond)
	}
	if err := db.CheckpointError(); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	memory, _ := Open(Options{})
	defer memory.Close()
	if err := memory.Checkpoint(); !errors.Is(err, storage.ErrMemoryOnly) {
		t.Fatal(err)
	}
	if _, err := Open(Options{CheckpointInterval: time.Second}); !errors.Is(err, storage.ErrMemoryOnly) {
		t.Fatal(err)
	}
}
