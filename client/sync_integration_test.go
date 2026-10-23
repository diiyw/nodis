package client

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/diiyw/nodis"
	"github.com/diiyw/nodis/server"
)

type synchronizedClient interface {
	Cached(nodis.DocumentID, nodis.ObjectID) (*nodis.PortableObject, bool)
	SyncStatus() SyncStatus
	Close()
}

func TestPermissionLossAndRestorationRebuildsCache(t *testing.T) {
	for _, mode := range []string{"embedded", "tcp"} {
		t.Run(mode, func(t *testing.T) {
			db, err := nodis.Open(nodis.Options{})
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			admin, user := nodis.Principal{ID: "admin"}, nodis.Principal{ID: "reader"}
			doc, err := db.CreateDocument(admin, "visible")
			if err != nil {
				t.Fatal(err)
			}
			tx := db.Begin(admin)
			draft, err := tx.Create(doc.ID())
			if err != nil {
				t.Fatal(err)
			}
			_ = draft.Set("name", "Ada")
			_ = draft.Set("secret", "private")
			id, err := draft.Save()
			if err != nil {
				t.Fatal(err)
			}
			if err := tx.Commit(); err != nil {
				t.Fatal(err)
			}
			allow := nodis.Rule{SubjectKind: nodis.SubjectUser, Subject: user.ID, Resource: nodis.Resource{Scope: nodis.ScopeDocument, Document: doc.ID()}, Actions: nodis.ActionRead | nodis.ActionDiscover | nodis.ActionSync, Allow: true}
			if err := db.Grant(admin, allow); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			scope := nodis.SyncScope{Documents: []nodis.DocumentID{doc.ID()}}
			var connected synchronizedClient
			if mode == "embedded" {
				connected, err = Connect(ctx, db, user, scope)
			} else {
				listener, listenErr := net.Listen("tcp", "127.0.0.1:0")
				if listenErr != nil {
					t.Fatal(listenErr)
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
				connected, err = Dial(ctx, listener.Addr().String(), []byte("token"), scope)
			}
			if err != nil {
				t.Fatal(err)
			}
			defer connected.Close()
			// A caller modifying its scope must not retarget future subscriptions.
			scope.Documents[0] = 999
			if value, ok := connected.Cached(doc.ID(), id); !ok {
				t.Fatal("initial cache missing")
			} else if _, visible := value.String("secret"); !visible {
				t.Fatal("initial secret missing")
			}
			if err := db.Revoke(admin, allow); err != nil {
				t.Fatal(err)
			}
			awaitStatus(t, connected.SyncStatus, func(status SyncStatus) bool { return !status.Ready && errors.Is(status.Err, nodis.ErrPermission) })
			if _, ok := connected.Cached(doc.ID(), id); ok {
				t.Fatal("cache survived complete permission loss")
			}
			secret, _ := doc.FieldID("secret")
			if err := db.Deny(admin, nodis.Rule{SubjectKind: nodis.SubjectUser, Subject: user.ID, Resource: nodis.Resource{Scope: nodis.ScopeField, Document: doc.ID(), Field: secret}, Actions: nodis.ActionRead | nodis.ActionSync}); err != nil {
				t.Fatal(err)
			}
			if err := db.Grant(admin, allow); err != nil {
				t.Fatal(err)
			}
			permission := db.PermissionRevision()
			awaitStatus(t, connected.SyncStatus, func(status SyncStatus) bool { return status.Ready && status.PermissionRevision == permission })
			value, ok := connected.Cached(doc.ID(), id)
			if !ok {
				t.Fatal("cache did not recover after permission grant")
			}
			if _, visible := value.String("secret"); visible {
				t.Fatal("field permission lost on recovery")
			}
			if name, _ := value.String("name"); name != "Ada" {
				t.Fatal("public field missing after recovery")
			}
			var wg sync.WaitGroup
			for range 8 {
				wg.Go(connected.Close)
			}
			wg.Wait()
			if connected.SyncStatus().Ready || !errors.Is(connected.SyncStatus().Err, nodis.ErrClosed) {
				t.Fatal("closed cache still ready")
			}
		})
	}
}

func TestEmbeddedOperationsAfterClose(t *testing.T) {
	db, err := nodis.Open(nodis.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	admin := nodis.Principal{ID: "admin"}
	doc, _ := db.CreateDocument(admin, "objects")
	client, err := Connect(context.Background(), db, admin, nodis.SyncScope{})
	if err != nil {
		t.Fatal(err)
	}
	client.Close()
	if _, err := client.Create(context.Background(), doc.ID(), map[string]any{"name": "late"}); !errors.Is(err, nodis.ErrClosed) {
		t.Fatal("closed client wrote to database", err)
	}
	if _, err := client.Get(context.Background(), doc.ID(), 1); !errors.Is(err, nodis.ErrClosed) {
		t.Fatal(err)
	}
}
