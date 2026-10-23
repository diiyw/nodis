package server_test

import (
	"context"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/diiyw/nodis"
	"github.com/diiyw/nodis/client"
	"github.com/diiyw/nodis/server"
)

func TestRemoteCRUDAndSynchronization(t *testing.T) {
	admin := nodis.Principal{ID: "root", Roles: []string{"admin"}}
	db, err := nodis.Open(nodis.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	doc, err := db.CreateDocument(admin, "people")
	if err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	s := server.New(db, server.StaticTokens{"secret-token": admin})
	done := make(chan error, 1)
	go func() { done <- s.Serve(listener) }()
	defer func() {
		_ = s.Close()
		if err := <-done; err != nil {
			t.Error(err)
		}
	}()

	remote, err := client.Dial(context.Background(), listener.Addr().String(), []byte("secret-token"), nodis.SyncScope{Documents: []nodis.DocumentID{doc.ID()}})
	if err != nil {
		t.Fatal(err)
	}
	defer remote.Close()
	if remoteID, err := remote.Document(context.Background(), "people"); err != nil || remoteID != doc.ID() {
		t.Fatalf("remote document = %d, %v", remoteID, err)
	}
	id, err := remote.Create(context.Background(), doc.ID(), map[string]any{"name": "Ada", "age": int64(37)})
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for {
		if object, ok := remote.Cached(doc.ID(), id); ok {
			if name, ok := object.String("name"); ok && name == "Ada" {
				break
			}
		}
		if time.Now().After(deadline) {
			t.Fatal("remote synchronization timed out")
		}
		time.Sleep(time.Millisecond)
	}
	object, err := remote.Get(context.Background(), doc.ID(), id)
	if err != nil {
		t.Fatal(err)
	}
	if err := remote.Update(context.Background(), doc.ID(), id, object.Revision(), map[string]any{"name": "Ada Lovelace"}); err != nil {
		t.Fatal(err)
	}
	object, err = remote.Get(context.Background(), doc.ID(), id)
	if err != nil {
		t.Fatal(err)
	}
	if name, ok := object.String("name"); !ok || name != "Ada Lovelace" {
		t.Fatal("remote update missing")
	}
	if err := remote.Delete(context.Background(), doc.ID(), id, object.Revision()); err != nil {
		t.Fatal(err)
	}
	if _, err := remote.Get(context.Background(), doc.ID(), id); err != nodis.ErrNotFound {
		t.Fatalf("get deleted = %v", err)
	}
}

func TestRemoteAuthentication(t *testing.T) {
	db, _ := nodis.Open(nodis.Options{})
	defer db.Close()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	s := server.New(db, server.StaticTokens{"valid": {ID: "user"}})
	done := make(chan error, 1)
	go func() { done <- s.Serve(listener) }()
	if _, err := client.Dial(context.Background(), listener.Addr().String(), []byte("invalid"), nodis.SyncScope{}); err != nodis.ErrPermission {
		t.Fatalf("dial = %v", err)
	}
	_ = s.Close()
	<-done
}

func TestRemoteWithoutAutomaticSync(t *testing.T) {
	db, err := nodis.Open(nodis.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	admin := nodis.Principal{ID: "admin"}
	user := nodis.Principal{ID: "writer"}
	doc, err := db.CreateDocument(admin, "request-only")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Grant(admin, nodis.Rule{
		SubjectKind: nodis.SubjectUser, Subject: user.ID,
		Resource: nodis.Resource{Scope: nodis.ScopeDocument, Document: doc.ID()},
		Actions:  nodis.ActionDiscover | nodis.ActionRead | nodis.ActionCreate,
	}); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	s := server.New(db, server.StaticTokens{"token": user})
	done := make(chan error, 1)
	go func() { done <- s.Serve(listener) }()
	defer func() {
		_ = s.Close()
		if err := <-done; err != nil {
			t.Error(err)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	remote, err := client.DialWithOptions(ctx, listener.Addr().String(), []byte("token"), client.DialOptions{DisableSync: true})
	if err != nil {
		t.Fatal(err)
	}
	defer remote.Close()
	id, err := remote.CreateAsync(ctx, doc.ID(), map[string]any{"name": "Ada"}).Wait(ctx)
	if err != nil {
		t.Fatal(err)
	}
	object, err := remote.Get(ctx, doc.ID(), id)
	if err != nil {
		t.Fatal(err)
	}
	if name, _ := object.String("name"); name != "Ada" {
		t.Fatal("authoritative read changed")
	}
	if _, ok := remote.Cached(doc.ID(), id); ok {
		t.Fatal("disabled synchronization populated cache")
	}
	// Ordinary requests remain available to a principal lacking sync permission.
	if synced, err := client.Dial(ctx, listener.Addr().String(), []byte("token"), nodis.SyncScope{Documents: []nodis.DocumentID{doc.ID()}}); err == nil {
		synced.Close()
		t.Fatal("subscription accepted without sync permission")
	}
}

func TestCanceledWriteCanCommitWithoutBeingReplayed(t *testing.T) {
	db, err := nodis.Open(nodis.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	admin := nodis.Principal{ID: "admin"}
	doc, err := db.CreateDocument(admin, "uncertain")
	if err != nil {
		t.Fatal(err)
	}
	started := make(chan nodis.ObjectID, 1)
	release := make(chan struct{})
	var releaseOnce sync.Once
	var executions atomic.Int32
	if err := db.RegisterPipeline(7, nodis.PipelineFunc(func(ctx context.Context, pipeline *nodis.PipelineContext) error {
		executions.Add(1)
		started <- pipeline.Object
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})); err != nil {
		t.Fatal(err)
	}
	if err := doc.SetDefaultPipeline(admin, 7); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	s := server.New(db, server.StaticTokens{"token": admin})
	done := make(chan error, 1)
	go func() { done <- s.Serve(listener) }()
	defer func() {
		_ = s.Close()
		if err := <-done; err != nil {
			t.Error(err)
		}
	}()
	// Release before the DB close even when an assertion fails mid-request.
	defer releaseOnce.Do(func() { close(release) })
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	remote, err := client.DialWithOptions(ctx, listener.Addr().String(), []byte("token"), client.DialOptions{DisableSync: true})
	if err != nil {
		t.Fatal(err)
	}
	defer remote.Close()
	requestCtx, cancelRequest := context.WithCancel(context.Background())
	defer cancelRequest()
	future := remote.CreateAsync(requestCtx, doc.ID(), map[string]any{"name": "once"})
	var id nodis.ObjectID
	select {
	case id = <-started:
	case <-ctx.Done():
		t.Fatal("pipeline did not start")
	}
	cancelRequest()
	if _, err := future.Wait(ctx); !errors.Is(err, client.ErrOutcomeUnknown) || !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled write = %v", err)
	}
	releaseOnce.Do(func() { close(release) })
	object, err := remote.Get(ctx, doc.ID(), id)
	if err != nil {
		t.Fatal("server commit was not recoverable through a fresh connection", err)
	}
	if name, _ := object.String("name"); name != "once" {
		t.Fatal("committed object changed")
	}
	if executions.Load() != 1 {
		t.Fatal("client replayed uncertain write")
	}
}
