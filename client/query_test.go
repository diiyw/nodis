package client

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/diiyw/nodis"
	"github.com/diiyw/nodis/server"
)

type queryClient interface {
	Find(context.Context, nodis.Query) (nodis.QueryPage, error)
	FindAsync(context.Context, nodis.Query) Future[nodis.QueryPage]
	Close()
}

func TestAuthoritativeQueryClientsAndAsyncInputOwnership(t *testing.T) {
	for _, mode := range []string{"embedded", "tcp"} {
		t.Run(mode, func(t *testing.T) {
			db, err := nodis.Open(nodis.Options{})
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			admin, user := nodis.Principal{ID: "admin"}, nodis.Principal{ID: "reader"}
			doc, _ := db.CreateDocument(admin, "find")
			var ids []nodis.ObjectID
			for i := range 5 {
				tx := db.Begin(admin)
				draft, err := tx.Create(doc.ID())
				if err != nil {
					t.Fatal(err)
				}
				_ = draft.Set("age", int64(i))
				_ = draft.Set("name", "Ada")
				_ = draft.Set("tag", []byte("x"))
				_ = draft.Set("secret", "private")
				id, err := draft.Save()
				if err != nil {
					t.Fatal(err)
				}
				if err := tx.Commit(); err != nil {
					t.Fatal(err)
				}
				ids = append(ids, id)
			}
			if err := db.Grant(admin, nodis.Rule{SubjectKind: nodis.SubjectUser, Subject: user.ID, Resource: nodis.Resource{Scope: nodis.ScopeDocument, Document: doc.ID()}, Actions: nodis.ActionRead | nodis.ActionDiscover | nodis.ActionSync}); err != nil {
				t.Fatal(err)
			}
			secret, _ := doc.FieldID("secret")
			if err := db.Deny(admin, nodis.Rule{SubjectKind: nodis.SubjectUser, Subject: user.ID, Resource: nodis.Resource{Scope: nodis.ScopeField, Document: doc.ID(), Field: secret}, Actions: nodis.ActionRead}); err != nil {
				t.Fatal(err)
			}
			age, _ := doc.FieldID("age")
			if err := db.Deny(admin, nodis.Rule{SubjectKind: nodis.SubjectUser, Subject: user.ID, Resource: nodis.Resource{Scope: nodis.ScopeField, Document: doc.ID(), Object: ids[2], Field: age}, Actions: nodis.ActionRead}); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			var connected queryClient
			if mode == "embedded" {
				connected, err = Connect(ctx, db, user, nodis.SyncScope{Documents: []nodis.DocumentID{doc.ID()}})
			} else {
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
				connected, err = DialWithOptions(ctx, listener.Addr().String(), []byte("token"), DialOptions{DisableSync: true})
			}
			if err != nil {
				t.Fatal(err)
			}
			defer connected.Close()
			filter, _ := nodis.Filter("age", nodis.GreaterEqual, int64(1))
			query := nodis.Query{Document: doc.ID(), Where: []nodis.Condition{filter}, Select: []string{"name", "secret"}, Limit: 1}
			first, err := connected.Find(ctx, query)
			if err != nil || len(first.Objects) != 1 || first.Objects[0].ID() != ids[1] || !first.HasMore || first.Next != ids[1] {
				t.Fatal("first query page", err)
			}
			if _, ok := first.Objects[0].Field("secret"); ok {
				t.Fatal("query projected hidden field")
			}
			query.After = first.Next
			second, err := connected.FindAsync(ctx, query).Wait(ctx)
			if err != nil || len(second.Objects) != 1 || second.Objects[0].ID() != ids[3] || !second.HasMore {
				t.Fatal("predicate bypassed per-object field ACL", err)
			}
			if name, _ := first.Objects[0].String("name"); name != "Ada" {
				t.Fatal("later response overwrote prior page")
			}
			prefix, _ := nodis.Filter("tag", nodis.Prefix, []byte("x"))
			asyncQuery := nodis.Query{Document: doc.ID(), Where: []nodis.Condition{prefix}, Select: []string{"name"}, Limit: 1}
			future := connected.FindAsync(ctx, asyncQuery)
			bytes, _ := asyncQuery.Where[0].Value.Bytes()
			bytes[0] = 'z'
			asyncQuery.Select[0] = "secret"
			asyncQuery.Where[0].Field = "other"
			result, err := future.Wait(ctx)
			if err != nil || len(result.Objects) != 1 {
				t.Fatal("async query borrowed mutable input", err)
			}
			if name, _ := result.Objects[0].String("name"); name != "Ada" {
				t.Fatal("async projection borrowed mutable input")
			}
			// Invalid container operands are rejected before cloning async inputs.
			cycle := make([]nodis.Value, 1)
			container, _ := nodis.ValueOf(cycle)
			cycle[0] = container
			invalid := nodis.Query{Document: doc.ID(), Where: []nodis.Condition{{Field: "tag", Op: nodis.Equal, Value: container}}}
			if _, err := connected.FindAsync(ctx, invalid).Wait(ctx); !errors.Is(err, nodis.ErrInvalidQuery) {
				t.Fatal("invalid async query was not rejected", err)
			}
			cancelled, stop := context.WithCancel(ctx)
			stop()
			if _, err := connected.Find(cancelled, query); !errors.Is(err, context.Canceled) {
				t.Fatal(err)
			}
			if err := db.Deny(admin, nodis.Rule{SubjectKind: nodis.SubjectUser, Subject: user.ID, Resource: nodis.Resource{Scope: nodis.ScopeDocument, Document: doc.ID()}, Actions: nodis.ActionDiscover}); err != nil {
				t.Fatal(err)
			}
			if _, err := connected.Find(ctx, query); !errors.Is(err, nodis.ErrPermission) {
				t.Fatal("server query bypassed document permissions", err)
			}
			connected.Close()
			if _, err := connected.Find(ctx, query); !errors.Is(err, nodis.ErrClosed) {
				t.Fatal("closed query client remained usable", err)
			}
		})
	}
}
