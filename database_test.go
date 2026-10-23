package nodis_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/diiyw/nodis"
	"github.com/diiyw/nodis/client"
)

var errRejected = errors.New("rejected by pipeline")

func TestCRUDPipelinePermissionsSyncAndRecovery(t *testing.T) {
	dir := t.TempDir()
	admin := nodis.Principal{ID: "root", Roles: []string{"admin"}}
	db, err := nodis.Open(nodis.Options{Dir: dir, Admin: "root", Durable: true})
	if err != nil {
		t.Fatal(err)
	}

	doc, err := db.CreateDocument(admin, "users")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.RegisterPipeline(1, nodis.PipelineFunc(func(_ context.Context, p *nodis.PipelineContext) error {
		if value, ok := p.Draft.Value("age"); ok {
			age, integer := value.Int()
			if !integer || age < 0 {
				return errRejected
			}
		}
		return p.Draft.Set("checked", true)
	})); err != nil {
		t.Fatal(err)
	}
	if err := doc.SetDefaultPipeline(admin, 1); err != nil {
		t.Fatal(err)
	}

	tx := db.Begin(admin)
	draft, err := tx.Create(doc.ID())
	if err != nil {
		t.Fatal(err)
	}
	if err := draft.Set("name", "Ada"); err != nil {
		t.Fatal(err)
	}
	if err := draft.Set("age", int64(37)); err != nil {
		t.Fatal(err)
	}
	if err := draft.Set("secret", "compiler"); err != nil {
		t.Fatal(err)
	}
	id, err := draft.Save()
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	object, err := db.Get(admin, doc.ID(), id)
	if err != nil {
		t.Fatal(err)
	}
	if name, ok := object.String("name"); !ok || name != "Ada" {
		t.Fatalf("name = %q, %v", name, ok)
	}
	if checked, ok := object.Bool("checked"); !ok || !checked {
		t.Fatal("pipeline output missing")
	}

	bad := db.Begin(admin)
	first, _ := bad.Create(doc.ID())
	_ = first.Set("age", int64(1))
	firstID, _ := first.Save()
	second, _ := bad.Create(doc.ID())
	_ = second.Set("age", int64(-1))
	secondID, _ := second.Save()
	if err := bad.Commit(); !errors.Is(err, errRejected) {
		t.Fatalf("pipeline error = %v", err)
	}
	if _, err := db.Get(admin, doc.ID(), firstID); !errors.Is(err, nodis.ErrNotFound) {
		t.Fatal("partial transaction was installed")
	}
	if _, err := db.Get(admin, doc.ID(), secondID); !errors.Is(err, nodis.ErrNotFound) {
		t.Fatal("rejected object was installed")
	}

	user := nodis.Principal{ID: "reader", Groups: []string{"staff"}}
	allow := nodis.Rule{SubjectKind: nodis.SubjectGroup, Subject: "staff", Resource: nodis.Resource{Scope: nodis.ScopeDocument, Document: doc.ID()}, Actions: nodis.ActionDiscover | nodis.ActionRead | nodis.ActionSync | nodis.ActionCreate | nodis.ActionUpdate | nodis.ActionExecute}
	if err := db.Grant(admin, allow); err != nil {
		t.Fatal(err)
	}
	secretID, ok := doc.FieldID("secret")
	if !ok {
		t.Fatal("missing field ID")
	}
	denySecret := nodis.Rule{SubjectKind: nodis.SubjectUser, Subject: user.ID, Resource: nodis.Resource{Scope: nodis.ScopeField, Document: doc.ID(), Object: id, Field: secretID}, Actions: nodis.ActionRead | nodis.ActionSync}
	if err := db.Deny(admin, denySecret); err != nil {
		t.Fatal(err)
	}

	c, err := client.Connect(context.Background(), db, user, nodis.SyncScope{Documents: []nodis.DocumentID{doc.ID()}})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	cached, ok := c.Cached(doc.ID(), id)
	if !ok {
		t.Fatal("object was not synchronized")
	}
	if _, ok := cached.String("secret"); ok {
		t.Fatal("secret field leaked into client snapshot")
	}
	if name, ok := cached.String("name"); !ok || name != "Ada" {
		t.Fatal("visible field missing")
	}

	if err := c.Update(context.Background(), doc.ID(), id, object.Revision(), map[string]any{"name": "Ada Lovelace"}); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for {
		cached, ok = c.Cached(doc.ID(), id)
		if ok {
			if name, found := cached.String("name"); found && name == "Ada Lovelace" {
				break
			}
		}
		if time.Now().After(deadline) {
			t.Fatal("incremental synchronization timed out")
		}
		time.Sleep(time.Millisecond)
	}
	adminView, err := db.Get(admin, doc.ID(), id)
	if err != nil {
		t.Fatal(err)
	}
	if secret, ok := adminView.String("secret"); !ok || secret != "compiler" {
		t.Fatal("hidden field was lost during update")
	}
	if err := db.Revoke(admin, denySecret); err != nil {
		t.Fatal(err)
	}
	waitCached(t, c, doc.ID(), id, func(object *nodis.PortableObject) bool {
		value, ok := object.String("secret")
		return ok && value == "compiler"
	})
	if err := db.Deny(admin, denySecret); err != nil {
		t.Fatal(err)
	}
	waitCached(t, c, doc.ID(), id, func(object *nodis.PortableObject) bool { _, ok := object.String("secret"); return !ok })
	future := c.CreateAsync(context.Background(), doc.ID(), map[string]any{"name": "Grace", "age": int64(40)})
	if _, err := future.Wait(context.Background()); err != nil {
		t.Fatal(err)
	}

	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = nodis.Open(nodis.Options{Dir: dir, Admin: "root", Durable: true})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	doc, err = db.Document(admin, "users")
	if err != nil {
		t.Fatal(err)
	}
	restored, err := db.Get(admin, doc.ID(), id)
	if err != nil {
		t.Fatal(err)
	}
	if name, ok := restored.String("name"); !ok || name != "Ada Lovelace" {
		t.Fatal("WAL recovery lost object")
	}
	if db.PermissionRevision() != 4 {
		t.Fatalf("permission revision = %d", db.PermissionRevision())
	}
}

func waitCached(t *testing.T, c *client.Client, document nodis.DocumentID, object nodis.ObjectID, ready func(*nodis.PortableObject) bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for {
		if value, ok := c.Cached(document, object); ok && ready(value) {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("client materialized view did not converge")
		}
		time.Sleep(time.Millisecond)
	}
}

func TestDocumentAndObjectDeleteRecovery(t *testing.T) {
	dir := t.TempDir()
	admin := nodis.Principal{ID: "admin"}
	db, err := nodis.Open(nodis.Options{Dir: dir, Durable: true})
	if err != nil {
		t.Fatal(err)
	}
	doc, err := db.CreateDocument(admin, "temporary")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.RenameDocument(admin, doc.ID(), "renamed"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Document(admin, "renamed"); err != nil {
		t.Fatal(err)
	}
	tx := db.Begin(admin)
	draft, _ := tx.Create(doc.ID())
	_ = draft.Set("value", int64(1))
	id, _ := draft.Save()
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	view, err := db.Get(admin, doc.ID(), id)
	if err != nil {
		t.Fatal(err)
	}
	deleteTx := db.Begin(admin)
	if err := deleteTx.Delete(doc.ID(), id, view.Revision()); err != nil {
		t.Fatal(err)
	}
	if err := deleteTx.Commit(); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Get(admin, doc.ID(), id); !errors.Is(err, nodis.ErrNotFound) {
		t.Fatal("object delete failed")
	}
	if err := db.DeleteDocument(admin, doc.ID()); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = nodis.Open(nodis.Options{Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Document(admin, "renamed"); !errors.Is(err, nodis.ErrNotFound) {
		t.Fatal("document delete was not recovered")
	}
}

func TestIncompleteWALTailIsDiscarded(t *testing.T) {
	dir := t.TempDir()
	admin := nodis.Principal{ID: "admin"}
	db, err := nodis.Open(nodis.Options{Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateDocument(admin, "stable"); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	file, err := os.OpenFile(filepath.Join(dir, "nodis.wal"), os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write([]byte("NDW1\xff")); err != nil {
		t.Fatal(err)
	}
	_ = file.Close()
	db, err = nodis.Open(nodis.Options{Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Document(admin, "stable"); err != nil {
		t.Fatal(err)
	}
}
