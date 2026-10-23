package nodis

import (
	"context"
	"errors"
	"math"
	"slices"
	"sync"
	"testing"
	"time"
)

func storeQueryObject(t testing.TB, db *DB, doc DocumentID, fields map[string]any) ObjectID {
	t.Helper()
	tx := db.Begin(Principal{ID: "admin"})
	draft, err := tx.Create(doc)
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

func condition(t testing.TB, name string, op CompareOp, value any) Condition {
	t.Helper()
	v, err := ValueOf(value)
	if err != nil {
		t.Fatal(err)
	}
	return Condition{Field: name, Op: op, Value: v}
}

func pageIDs(page QueryPage) []ObjectID {
	ids := make([]ObjectID, len(page.Objects))
	for i := range page.Objects {
		ids[i] = page.Objects[i].ID()
	}
	return ids
}

func TestQueryDynamicKindsAndOperators(t *testing.T) {
	db, _ := Open(Options{})
	defer db.Close()
	admin := Principal{ID: "admin"}
	doc, _ := db.CreateDocument(admin, "mixed")
	date := time.Unix(1700000000, 123)
	a := storeQueryObject(t, db, doc.id, map[string]any{"age": int64(42), "name": "Ada", "bytes": []byte{0, 255}, "active": true, "when": date, "null": nil})
	b := storeQueryObject(t, db, doc.id, map[string]any{"age": int(18), "name": "Alan", "active": false})
	u := storeQueryObject(t, db, doc.id, map[string]any{"age": uint64(42)})
	f := storeQueryObject(t, db, doc.id, map[string]any{"age": 42.0})
	n := storeQueryObject(t, db, doc.id, map[string]any{"age": nil})
	missing := storeQueryObject(t, db, doc.id, map[string]any{"other": true})
	nan := storeQueryObject(t, db, doc.id, map[string]any{"age": math.NaN()})
	for _, test := range []struct {
		name  string
		where []Condition
		want  []ObjectID
	}{
		{"integer", []Condition{condition(t, "age", Equal, int64(42))}, []ObjectID{a}},
		{"not-equal", []Condition{condition(t, "age", NotEqual, int64(42))}, []ObjectID{b}},
		{"less", []Condition{condition(t, "age", Less, int64(42))}, []ObjectID{b}},
		{"less-equal", []Condition{condition(t, "age", LessEqual, int64(42))}, []ObjectID{a, b}},
		{"greater", []Condition{condition(t, "age", Greater, int64(18))}, []ObjectID{a}},
		{"greater-equal", []Condition{condition(t, "age", GreaterEqual, int64(42))}, []ObjectID{a}},
		{"uint", []Condition{condition(t, "age", Equal, uint64(42))}, []ObjectID{u}},
		{"float", []Condition{condition(t, "age", Equal, 42.0)}, []ObjectID{f}},
		{"nan-does-not-match", []Condition{condition(t, "age", NotEqual, 42.0)}, nil},
		{"null", []Condition{condition(t, "age", Equal, nil)}, []ObjectID{n}},
		{"missing", []Condition{{Field: "age", Op: Missing}}, []ObjectID{missing}},
		{"unknown-field", []Condition{{Field: "never-written", Op: Missing}}, []ObjectID{a, b, u, f, n, missing, nan}},
		{"exists-includes-null", []Condition{{Field: "age", Op: Exists}}, []ObjectID{a, b, u, f, n, nan}},
		{"prefix", []Condition{condition(t, "name", Prefix, "A")}, []ObjectID{a, b}},
		{"and", []Condition{condition(t, "name", Prefix, "A"), condition(t, "age", Less, int64(20))}, []ObjectID{b}},
		{"boolean", []Condition{condition(t, "active", Equal, true)}, []ObjectID{a}},
		{"bytes", []Condition{condition(t, "bytes", Prefix, []byte{0})}, []ObjectID{a}},
		{"time", []Condition{condition(t, "when", Equal, date)}, []ObjectID{a}},
	} {
		t.Run(test.name, func(t *testing.T) {
			page, err := db.Find(context.Background(), admin, Query{Document: doc.id, Where: test.where})
			if err != nil {
				t.Fatal(err)
			}
			if got := pageIDs(page); !slices.Equal(got, test.want) {
				t.Fatalf("IDs = %v, want %v", got, test.want)
			}
		})
	}
	filter, err := Filter("age", GreaterEqual, int64(18))
	if err != nil {
		t.Fatal(err)
	}
	page, err := db.Find(context.Background(), admin, Query{Document: doc.id, Where: []Condition{filter}, Select: []string{"name", "never-written"}})
	if err != nil || len(page.Objects) != 2 {
		t.Fatal("generic filter failed", err)
	}
	if _, found := page.Objects[0].Field("age"); found {
		t.Fatal("projection included age")
	}
}

func TestQueryFieldACLDoesNotLeakThroughPredicatesOrPagination(t *testing.T) {
	db, _ := Open(Options{})
	defer db.Close()
	admin, user := Principal{ID: "admin"}, Principal{ID: "reader"}
	doc, _ := db.CreateDocument(admin, "private")
	first := storeQueryObject(t, db, doc.id, map[string]any{"name": "Ada", "secret": "private"})
	second := storeQueryObject(t, db, doc.id, map[string]any{"name": "Grace"})
	hidden := storeQueryObject(t, db, doc.id, map[string]any{"name": "hidden", "secret": "other"})
	if err := db.Grant(admin, Rule{SubjectKind: SubjectUser, Subject: user.ID, Resource: Resource{Scope: ScopeDocument, Document: doc.id}, Actions: ActionRead | ActionDiscover}); err != nil {
		t.Fatal(err)
	}
	secret, _ := doc.FieldID("secret")
	if err := db.Deny(admin, Rule{SubjectKind: SubjectUser, Subject: user.ID, Resource: Resource{Scope: ScopeField, Document: doc.id, Field: secret}, Actions: ActionRead}); err != nil {
		t.Fatal(err)
	}
	if err := db.Deny(admin, Rule{SubjectKind: SubjectUser, Subject: user.ID, Resource: Resource{Scope: ScopeObject, Document: doc.id, Object: hidden}, Actions: ActionRead}); err != nil {
		t.Fatal(err)
	}
	page, err := db.Find(context.Background(), user, Query{Document: doc.id, Limit: 2, Select: []string{"name", "secret"}})
	if err != nil || !slices.Equal(pageIDs(page), []ObjectID{first, second}) || page.HasMore || page.Next != 0 {
		t.Fatalf("visible page = %+v, %v", page, err)
	}
	for i := range page.Objects {
		if _, ok := page.Objects[i].Field("secret"); ok {
			t.Fatal("hidden field projected")
		}
	}
	for _, predicate := range []Condition{
		condition(t, "secret", Equal, "private"), condition(t, "secret", NotEqual, "private"), condition(t, "secret", Prefix, "p"), {Field: "secret", Op: Exists}, {Field: "secret", Op: Missing},
	} {
		page, err := db.Find(context.Background(), user, Query{Document: doc.id, Where: []Condition{predicate}, Limit: 1})
		if err != nil || len(page.Objects) != 0 || page.HasMore || page.Next != 0 {
			t.Fatal("predicate leaked hidden field", predicate.Op, err)
		}
	}
	if err := db.Deny(admin, Rule{SubjectKind: SubjectUser, Subject: user.ID, Resource: Resource{Scope: ScopeDocument, Document: doc.id}, Actions: ActionDiscover}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Find(context.Background(), user, Query{Document: doc.id}); !errors.Is(err, ErrPermission) {
		t.Fatal("query bypassed discover permission", err)
	}
}

func TestQueryKeysetPaginationAndDeletedCursor(t *testing.T) {
	db, _ := Open(Options{})
	defer db.Close()
	admin := Principal{ID: "admin"}
	doc, _ := db.CreateDocument(admin, "pages")
	var ids []ObjectID
	for range 5 {
		ids = append(ids, storeQueryObject(t, db, doc.id, map[string]any{"name": "row"}))
	}
	page, err := db.Find(context.Background(), admin, Query{Document: doc.id, Limit: 2})
	if err != nil || !slices.Equal(pageIDs(page), ids[:2]) || !page.HasMore || page.Next != ids[1] {
		t.Fatal("first page", page, err)
	}
	tx := db.Begin(admin)
	_ = tx.Delete(doc.id, ids[1], 1)
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	newID := storeQueryObject(t, db, doc.id, map[string]any{"name": "new"})
	second, err := db.Find(context.Background(), admin, Query{Document: doc.id, After: page.Next, Limit: 2})
	if err != nil || !slices.Equal(pageIDs(second), ids[2:4]) || second.Next != ids[3] {
		t.Fatal("second page", second, err)
	}
	last, err := db.Find(context.Background(), admin, Query{Document: doc.id, After: second.Next, Limit: 2})
	if err != nil || !slices.Equal(pageIDs(last), []ObjectID{ids[4], newID}) || last.HasMore || last.Next != 0 {
		t.Fatal("last page", last, err)
	}
	// Convenience Find owns its result buffers; later queries do not overwrite it.
	if name, _ := page.Objects[0].String("name"); name != "row" {
		t.Fatal("prior page changed")
	}
}

func TestQueryIndexCommitRollbackAndRecovery(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(Options{Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	admin := Principal{ID: "admin"}
	doc, _ := db.CreateDocument(admin, "indexed")
	lowTx, highTx := db.Begin(admin), db.Begin(admin)
	low, _ := lowTx.Create(doc.id)
	_ = low.Set("value", 1)
	lowID, _ := low.Save()
	high, _ := highTx.Create(doc.id)
	_ = high.Set("value", 2)
	highID, _ := high.Save()
	if err := highTx.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := lowTx.Commit(); err != nil {
		t.Fatal(err)
	}
	ordered, err := db.Find(context.Background(), admin, Query{Document: doc.id})
	if err != nil || !slices.Equal(pageIDs(ordered), []ObjectID{lowID, highID}) {
		t.Fatal("out-of-order commits broke primary index", err)
	}
	if err := db.Checkpoint(); err != nil {
		t.Fatal(err)
	}
	tx := db.Begin(admin)
	_ = tx.Delete(doc.id, lowID, 1)
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	finalID := storeQueryObject(t, db, doc.id, map[string]any{"value": 3})
	injected := errors.New("pipeline rejected")
	if err := db.RegisterPipeline(1, PipelineFunc(func(context.Context, *PipelineContext) error { return injected })); err != nil {
		t.Fatal(err)
	}
	if err := doc.SetDefaultPipeline(admin, 1); err != nil {
		t.Fatal(err)
	}
	rejected := db.Begin(admin)
	draft, _ := rejected.Create(doc.id)
	_ = draft.Set("value", 4)
	_, _ = draft.Save()
	if err := rejected.Commit(); !errors.Is(err, injected) {
		t.Fatal(err)
	}
	if !slices.Equal([]ObjectID(doc.objectIDs), []ObjectID{highID, finalID}) {
		t.Fatal("index changed on failed transaction", doc.objectIDs)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = Open(Options{Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	page, err := db.Find(context.Background(), admin, Query{Document: doc.id})
	if err != nil || !slices.Equal(pageIDs(page), []ObjectID{highID, finalID}) {
		t.Fatal("snapshot + WAL index recovery failed", err)
	}
	other, _ := db.CreateDocument(admin, "removed")
	stale := db.Begin(admin)
	pending, _ := stale.Create(other.id)
	_ = pending.Set("value", 1)
	_, _ = pending.Save()
	if err := db.DeleteDocument(admin, other.id); err != nil {
		t.Fatal(err)
	}
	before := db.Sequence()
	if err := stale.Commit(); !errors.Is(err, ErrNotFound) {
		t.Fatal("orphan draft committed", err)
	}
	if db.Sequence() != before {
		t.Fatal("orphan draft reached WAL")
	}
}

func TestQueryByteBudgetAndProjection(t *testing.T) {
	db, _ := Open(Options{})
	defer db.Close()
	admin := Principal{ID: "admin"}
	doc, _ := db.CreateDocument(admin, "large")
	large := make([]byte, MaxQueryPageBytes/2+1)
	first := storeQueryObject(t, db, doc.id, map[string]any{"name": "one", "data": large})
	second := storeQueryObject(t, db, doc.id, map[string]any{"name": "two", "data": large})
	page, err := db.Find(context.Background(), admin, Query{Document: doc.id})
	if err != nil || len(page.Objects) != 1 || page.Next != first || !page.HasMore {
		t.Fatal("byte budget did not paginate", err)
	}
	page, err = db.Find(context.Background(), admin, Query{Document: doc.id, After: page.Next})
	if err != nil || len(page.Objects) != 1 || page.Objects[0].ID() != second || page.HasMore {
		t.Fatal("byte page continuation failed", err)
	}
	oversize := storeQueryObject(t, db, doc.id, map[string]any{"name": "oversize", "data": make([]byte, MaxQueryPageBytes)})
	if _, err := db.Find(context.Background(), admin, Query{Document: doc.id, After: second}); !errors.Is(err, ErrQueryTooLarge) {
		t.Fatal("oversized object accepted", err)
	}
	page, err = db.Find(context.Background(), admin, Query{Document: doc.id, After: second, Select: []string{"name"}})
	if err != nil || len(page.Objects) != 1 || page.Objects[0].ID() != oversize {
		t.Fatal("projection could not reduce page", err)
	}
}

func TestQueryValidationAndClosedDatabase(t *testing.T) {
	for _, query := range []Query{
		{}, {Document: 1, Limit: -1}, {Document: 1, Limit: MaxQueryLimit + 1}, {Document: 1, Select: []string{"same", "same"}},
		{Document: 1, Where: []Condition{condition(t, "x", Prefix, true)}}, {Document: 1, Where: []Condition{condition(t, "x", Equal, []Value{})}},
		{Document: 1, Where: []Condition{condition(t, "x", Equal, math.NaN())}}, {Document: 1, Where: []Condition{{Field: "x", Op: 255}}},
	} {
		if !errors.Is(query.Validate(), ErrInvalidQuery) {
			t.Fatal("invalid query accepted", query)
		}
	}
	db, _ := Open(Options{})
	defer db.Close()
	admin := Principal{ID: "admin"}
	doc, _ := db.CreateDocument(admin, "validate")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := db.Find(ctx, admin, Query{Document: doc.id}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := db.FindInto(context.Background(), admin, Query{Document: doc.id}, nil); !errors.Is(err, ErrInvalidQuery) {
		t.Fatal(err)
	}
	_ = db.Close()
	if _, err := db.Find(context.Background(), admin, Query{Document: doc.id}); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
}

func TestQueryIntoAndOrderedIndexZeroAllocations(t *testing.T) {
	db, _ := Open(Options{EventQueue: 8192})
	defer db.Close()
	admin := Principal{ID: "admin"}
	doc, _ := db.CreateDocument(admin, "allocs")
	for i := range 3 {
		storeQueryObject(t, db, doc.id, map[string]any{"age": int64(i), "name": "Ada"})
	}
	sub, err := db.Subscribe(admin, SyncScope{Documents: []DocumentID{doc.id}})
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Close()
	filter, err := Filter("age", GreaterEqual, int64(1))
	if err != nil {
		t.Fatal(err)
	}
	query := Query{Document: doc.id, Where: []Condition{filter}, Select: []string{"name"}}
	var buffer QueryBuffer
	if allocs := testing.AllocsPerRun(1000, func() {
		page, err := db.FindInto(context.Background(), admin, query, &buffer)
		if err != nil || len(page.Objects) != 2 {
			panic("query failed")
		}
		if name, _ := page.Objects[0].String("name"); name != "Ada" {
			panic("projection failed")
		}
	}); allocs != 0 {
		t.Fatalf("FindInto: %g allocs", allocs)
	}
	ids := orderedIDs[ObjectID](make([]ObjectID, 0, 10))
	ids.add(1)
	ids.add(3)
	if allocs := testing.AllocsPerRun(1000, func() {
		ids.add(2)
		if len(ids.after(1)) != 2 {
			panic("index changed")
		}
		ids.remove(2)
	}); allocs != 0 {
		t.Fatalf("ordered index: %g allocs", allocs)
	}
}

func TestQueryReadEventsAndConcurrentIndexMaintenance(t *testing.T) {
	db, err := Open(Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	admin := Principal{ID: "admin"}
	doc, _ := db.CreateDocument(admin, "concurrent-query")
	id := storeQueryObject(t, db, doc.id, map[string]any{"value": int64(0)})
	sub, err := db.Subscribe(admin, SyncScope{Documents: []DocumentID{doc.id}})
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Close()
	for event := range sub.C {
		if event.Type == EventSnapshotDone {
			break
		}
	}
	sequence := db.Sequence()
	page, err := db.Find(context.Background(), admin, Query{Document: doc.id})
	if err != nil || page.Sequence != sequence || db.Sequence() != sequence {
		t.Fatal("query changed commit sequence", err)
	}
	for _, expected := range []EventType{EventDocumentRead, EventObjectRead} {
		select {
		case event := <-sub.C:
			if event.Type != expected || event.Document != doc.id || event.Sequence != sequence {
				t.Fatal("read event changed", event)
			}
			if expected == EventObjectRead && event.Object != id {
				t.Fatal("wrong query read object")
			}
		case <-time.After(time.Second):
			t.Fatal("query read event missing")
		}
	}
	sub.Close()
	var wg sync.WaitGroup
	for range 2 {
		wg.Go(func() {
			for range 20 {
				tx := db.Begin(admin)
				draft, err := tx.Create(doc.id)
				if err != nil {
					t.Error(err)
					return
				}
				_ = draft.Set("value", int64(42))
				_, _ = draft.Save()
				if err := tx.Commit(); err != nil {
					t.Error(err)
					return
				}
			}
		})
	}
	wg.Go(func() {
		var buffer QueryBuffer
		for range 30 {
			page, err := db.FindInto(context.Background(), admin, Query{Document: doc.id}, &buffer)
			if err != nil {
				t.Error(err)
				return
			}
			var previous ObjectID
			for i := range page.Objects {
				id := page.Objects[i].ID()
				if id <= previous {
					t.Error("concurrent primary index not ordered")
					return
				}
				previous = id
			}
		}
	})
	wg.Wait()
	page, err = db.Find(context.Background(), admin, Query{Document: doc.id})
	if err != nil || len(page.Objects) != 41 {
		t.Fatal("concurrent queries lost objects", err)
	}
}
