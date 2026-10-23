package nodis

import (
	"cmp"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"sync"
	"testing"
	"time"
	"unsafe"

	"github.com/diiyw/nodis/storage"
)

// This oracle deliberately enumerates the primary index and evaluates every
// predicate, without using any secondary index or candidate planner.
func scanQueryIDs(db *DB, actor Principal, query Query) []ObjectID {
	db.mu.RLock()
	defer db.mu.RUnlock()
	doc := db.documents[query.Document]
	conditions := make([]preparedCondition, 0, len(query.Where))
	for _, condition := range query.Where {
		field, _ := doc.fieldID(condition.Field)
		conditions = append(conditions, preparedCondition{field, condition.Op, condition.Value})
	}
	var ids []ObjectID
	for _, id := range doc.objectIDs.after(query.After) {
		resource := Resource{Scope: ScopeObject, Document: doc.id, Object: id}
		if !db.allowedLocked(actor, resource, ActionDiscover) || !db.allowedLocked(actor, resource, ActionRead) {
			continue
		}
		view := ObjectView{document: doc, data: db.objects[objectKey{doc.id, id}]}
		if db.queryMatches(actor, &view, conditions) {
			ids = append(ids, id)
		}
	}
	return ids
}

func assertIndexedQuery(t testing.TB, db *DB, actor Principal, query Query) {
	t.Helper()
	want := scanQueryIDs(db, actor, query)
	limit := query.Limit
	if limit == 0 {
		limit = DefaultQueryLimit
	}
	more := len(want) > limit
	if more {
		want = want[:limit]
	}
	page, err := db.Find(context.Background(), actor, query)
	if err != nil {
		t.Fatal(err)
	}
	var next ObjectID
	if more {
		next = want[len(want)-1]
	}
	if !slices.Equal(pageIDs(page), want) || page.HasMore != more || page.Next != next {
		t.Fatalf("indexed query differs from scan: query=%+v IDs=%v want=%v more=%v next=%v", query, pageIDs(page), want, page.HasMore, page.Next)
	}
}

func assertSecondaryIndexState(t testing.TB, db *DB) {
	t.Helper()
	db.mu.RLock()
	defer db.mu.RUnlock()
	for _, doc := range db.documents {
		want := make(map[FieldID][]indexEntry[scalarKey])
		for _, id := range doc.objectIDs {
			view := ObjectView{document: doc, data: db.objects[objectKey{doc.id, id}]}
			for field := range doc.fieldNames {
				value, exists := view.fieldID(field)
				if !exists {
					continue
				}
				if key, ok := scalarIndexKey(value); ok {
					want[field] = append(want[field], indexEntry[scalarKey]{key, id})
				}
			}
		}
		for field, entries := range want {
			if len(entries) > 0 && doc.indexes[field] == nil {
				t.Fatalf("missing index for document=%d field=%d", doc.id, field)
			}
		}
		for field, index := range doc.indexes {
			assertOrderedIndex(t, index, compareScalarKey)
			got := orderedIndexEntries(index)
			entries := want[field]
			slices.SortFunc(entries, func(a, b indexEntry[scalarKey]) int {
				return compareIndexEntry(a, b, compareScalarKey)
			})
			if len(entries) != len(got) {
				t.Fatalf("index cardinality document=%d field=%d got=%d want=%d", doc.id, field, len(got), len(entries))
			}
			for i, entry := range entries {
				if compareIndexEntry(entry, got[i], compareScalarKey) != 0 {
					t.Fatalf("index state document=%d field=%d entry=%d differs from live object", doc.id, field, i)
				}
			}
		}
	}
}

func TestSecondaryIndexMatchesScanAcrossKinds(t *testing.T) {
	db, err := Open(Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	admin := Principal{ID: "admin"}
	doc, _ := db.CreateDocument(admin, "secondary-kinds")
	values := []any{
		nil, false, true, int64(math.MinInt64), int64(-1), int64(0), int64(1), int64(math.MaxInt64),
		uint64(0), uint64(1), uint64(math.MaxUint64),
		math.Inf(-1), -math.MaxFloat64, -1.0, math.Copysign(0, -1), 0.0, math.SmallestNonzeroFloat64, 1.0, math.MaxFloat64, math.Inf(1), math.NaN(),
		"", "\x00", "a", "ab", "a\xff", "b", "\xff", "\xff\xff", "你好", "你好世界",
		[]byte{}, []byte{0}, []byte{0, 255}, []byte{1}, []byte{255}, []byte{255, 255},
		time.Unix(0, math.MinInt64), time.Unix(0, -1), time.Unix(0, 0), time.Unix(0, math.MaxInt64),
		[]Value{}, []NamedValue{},
	}
	random := rand.New(rand.NewPCG(17, 31))
	for i := range 250 {
		fields := map[string]any{"tag": int64(i % 5)}
		if i%9 != 0 {
			fields["v"] = values[random.IntN(len(values))]
		}
		storeQueryObject(t, db, doc.id, fields)
	}
	// Ensure every boundary, including signed zero, exists even if the random
	// sample missed it. Both zero signs must compare equal and share a range.
	for _, value := range values {
		storeQueryObject(t, db, doc.id, map[string]any{"v": value, "tag": int64(1)})
	}
	user := Principal{ID: "reader"}
	if err := db.Grant(admin, Rule{SubjectKind: SubjectUser, Subject: user.ID, Resource: Resource{Scope: ScopeDocument, Document: doc.id}, Actions: ActionDiscover | ActionRead}); err != nil {
		t.Fatal(err)
	}
	field, _ := doc.FieldID("v")
	for i, id := range doc.objectIDs {
		resource := Resource{Scope: ScopeField, Document: doc.id, Object: id, Field: field}
		if i%7 == 0 {
			if err := db.Deny(admin, Rule{SubjectKind: SubjectUser, Subject: user.ID, Resource: resource, Actions: ActionRead}); err != nil {
				t.Fatal(err)
			}
		}
		if i%11 == 0 {
			resource.Scope = ScopeObject
			if err := db.Deny(admin, Rule{SubjectKind: SubjectUser, Subject: user.ID, Resource: resource, Actions: ActionDiscover}); err != nil {
				t.Fatal(err)
			}
		}
	}
	for i, value := range values {
		for op := Equal; op <= Prefix; op++ {
			predicate := condition(t, "v", op, value)
			if predicate.validate() != nil {
				continue
			}
			t.Run(fmt.Sprintf("%d/%d", i, op), func(t *testing.T) {
				for _, actor := range []Principal{admin, user} {
					for _, after := range []ObjectID{0, 85, 249, 1000} {
						query := Query{Document: doc.id, Where: []Condition{predicate}, After: after, Limit: 7, Select: []string{"tag"}}
						assertIndexedQuery(t, db, actor, query)
						query.Where = append(query.Where, condition(t, "tag", Equal, int64(1)))
						assertIndexedQuery(t, db, actor, query)
					}
				}
			})
		}
	}
	for _, field := range []string{"v", "never-written"} {
		for _, op := range []CompareOp{Exists, Missing} {
			assertIndexedQuery(t, db, user, Query{Document: doc.id, Where: []Condition{{Field: field, Op: op}}, Limit: 7})
		}
	}
	assertSecondaryIndexState(t, db)
}

func TestSecondaryIndexPlannerChoosesSmallestRange(t *testing.T) {
	db, _ := Open(Options{})
	defer db.Close()
	admin := Principal{ID: "admin"}
	doc, _ := db.CreateDocument(admin, "planner")
	var target ObjectID
	for i := range 100 {
		id := storeQueryObject(t, db, doc.id, map[string]any{"unique": int64(i), "group": int64(i % 2)})
		if i == 42 {
			target = id
		}
	}
	group, _ := doc.FieldID("group")
	unique, _ := doc.FieldID("unique")
	groupValue, _ := ValueOf(int64(0))
	uniqueValue, _ := ValueOf(int64(42))
	conditions := []preparedCondition{{group, Equal, groupValue}, {unique, Equal, uniqueValue}}
	var buffer QueryBuffer
	for range 2 {
		buffer.candidates = buffer.candidates[:0]
		ids, err := doc.queryIDs(context.Background(), 0, conditions, &buffer)
		if err != nil || !slices.Equal(ids, []ObjectID{target}) || len(buffer.candidates) != 1 {
			t.Fatalf("did not use narrowest index range: %v, %v", ids, err)
		}
		slices.Reverse(conditions)
	}
	// A late cursor with fewer primary IDs than index entries uses the primary
	// path instead of materializing the broad index range.
	buffer.candidates = buffer.candidates[:0]
	ids, err := doc.queryIDs(context.Background(), 98, conditions[:1], &buffer)
	if err != nil || len(ids) != 2 || len(buffer.candidates) != 0 {
		t.Fatal("planner did not prefer shorter primary scan", ids, err)
	}
	value, _ := ValueOf(int64(1000))
	ids, err = doc.queryIDs(context.Background(), 0, []preparedCondition{{unique, Equal, value}}, &buffer)
	if err != nil || len(ids) != 0 {
		t.Fatal("empty index range did not eliminate candidates", ids, err)
	}
}

func TestSecondaryIndexMutationRollbackAndRecovery(t *testing.T) {
	for _, checkpoint := range []bool{false, true} {
		t.Run(fmt.Sprintf("checkpoint=%v", checkpoint), func(t *testing.T) {
			dir := t.TempDir()
			db, err := Open(Options{Dir: dir, Durable: true})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = db.Close() })
			admin := Principal{ID: "admin"}
			doc, _ := db.CreateDocument(admin, "mutations")
			a := storeQueryObject(t, db, doc.id, map[string]any{"v": "old", "removed": int64(1)})
			b := storeQueryObject(t, db, doc.id, map[string]any{"v": "old"})
			c := storeQueryObject(t, db, doc.id, map[string]any{"v": "deleted"})
			if checkpoint {
				if err := db.Checkpoint(); err != nil {
					t.Fatal(err)
				}
			}
			tx := db.Begin(admin)
			draft, _ := tx.Update(doc.id, a, 1)
			_ = draft.Set("v", int64(42)) // Inferred kind changes without a schema.
			draft.DeleteField("removed")
			_, _ = draft.Save()
			other, _ := tx.Update(doc.id, b, 1)
			_ = other.Set("v", []Value{}) // Scalar becomes a container.
			_, _ = other.Save()
			if err := tx.Delete(doc.id, c, 1); err != nil {
				t.Fatal(err)
			}
			if err := tx.Commit(); err != nil {
				t.Fatal(err)
			}
			injected := errors.New("pipeline rejection")
			if err := db.RegisterPipeline(1, PipelineFunc(func(_ context.Context, p *PipelineContext) error {
				_ = p.Draft.Set("v", "rejected")
				return injected
			})); err != nil {
				t.Fatal(err)
			}
			rejected := db.Begin(admin)
			change, _ := rejected.Update(doc.id, a, 2)
			change.SetPipeline(1)
			_, _ = change.Save()
			if err := rejected.Commit(); !errors.Is(err, injected) {
				t.Fatal(err)
			}
			conflict := db.Begin(admin)
			stale, _ := conflict.Update(doc.id, a, 1)
			_ = stale.Set("v", "conflicted")
			_, _ = stale.Save()
			if err := conflict.Commit(); !errors.Is(err, ErrConflict) {
				t.Fatal(err)
			}
			// Pipeline-produced fields must enter the index at their final values.
			if err := db.RegisterPipeline(2, PipelineFunc(func(_ context.Context, p *PipelineContext) error {
				return p.Draft.Set("v", "pipeline")
			})); err != nil {
				t.Fatal(err)
			}
			piped := db.Begin(admin)
			p, _ := piped.Update(doc.id, b, 2)
			p.SetPipeline(2)
			_, _ = p.Save()
			if err := piped.Commit(); err != nil {
				t.Fatal(err)
			}
			verify := func() {
				t.Helper()
				assertSecondaryIndexState(t, db)
				for _, value := range []any{"old", "deleted", "rejected", "conflicted", "pipeline", int64(42)} {
					assertIndexedQuery(t, db, admin, Query{Document: doc.id, Where: []Condition{condition(t, "v", Equal, value)}})
				}
				assertIndexedQuery(t, db, admin, Query{Document: doc.id, Where: []Condition{condition(t, "removed", Equal, int64(1))}})
				field, _ := doc.FieldID("v")
				entries := db.documents[doc.id].indexes[field]
				if entries.len() != 2 {
					t.Fatal("stale index entries retained", entries)
				}
			}
			verify()
			if err := db.Close(); err != nil {
				t.Fatal(err)
			}
			db, err = Open(Options{Dir: dir})
			if err != nil {
				t.Fatal(err)
			}
			verify()
			current := db.documents[doc.id]
			if err := db.DeleteDocument(admin, doc.id); err != nil {
				t.Fatal(err)
			}
			if current.indexes != nil || current.objectIDs != nil {
				t.Fatal("deleted document retained index storage")
			}
			if err := db.Close(); err != nil {
				t.Fatal(err)
			}
			db, err = Open(Options{Dir: dir})
			if err != nil || len(db.objects) != 0 || len(db.documents) != 0 {
				t.Fatal("deleted document indexes reappeared on recovery", err)
			}
		})
	}
}

func TestSecondaryIndexFailedWritesLeaveCommittedState(t *testing.T) {
	for _, mode := range []string{"permission", "wal", "rollback"} {
		t.Run(mode, func(t *testing.T) {
			db, err := Open(Options{Dir: t.TempDir()})
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			admin, user := Principal{ID: "admin"}, Principal{ID: "writer"}
			doc, _ := db.CreateDocument(admin, "failed-writes")
			id := storeQueryObject(t, db, doc.id, map[string]any{"v": int64(1)})
			if err := db.Grant(admin, Rule{SubjectKind: SubjectUser, Subject: user.ID, Resource: Resource{Scope: ScopeDocument, Document: doc.id}, Actions: ActionRead | ActionDiscover | ActionUpdate}); err != nil {
				t.Fatal(err)
			}
			actor, expected := admin, error(storage.ErrClosed)
			if mode == "permission" {
				actor, expected = user, ErrPermission
				field, _ := doc.FieldID("v")
				if err := db.Deny(admin, Rule{SubjectKind: SubjectUser, Subject: user.ID, Resource: Resource{Scope: ScopeField, Document: doc.id, Field: field}, Actions: ActionUpdate}); err != nil {
					t.Fatal(err)
				}
			}
			tx := db.Begin(actor)
			draft, err := tx.Update(doc.id, id, 1)
			if err != nil {
				t.Fatal(err)
			}
			_ = draft.Set("v", int64(2))
			_, _ = draft.Save()
			sequence := db.Sequence()
			if mode == "rollback" {
				tx.Rollback()
				expected = ErrInvalidState
			} else if mode == "wal" {
				if err := db.log.Close(); err != nil {
					t.Fatal(err)
				}
			}
			if err := tx.Commit(); !errors.Is(err, expected) {
				t.Fatal("wrong write failure", err)
			}
			if db.Sequence() != sequence {
				t.Fatal("rejected write changed sequence")
			}
			assertSecondaryIndexState(t, db)
			for _, value := range []int64{1, 2} {
				assertIndexedQuery(t, db, admin, Query{Document: doc.id, Where: []Condition{condition(t, "v", Equal, value)}})
			}
		})
	}
}

func TestSecondaryIndexRejectsOversizedFieldCountBeforeWAL(t *testing.T) {
	db, _ := Open(Options{})
	defer db.Close()
	admin := Principal{ID: "admin"}
	doc, _ := db.CreateDocument(admin, "field-count")
	id := storeQueryObject(t, db, doc.id, map[string]any{"v": int64(1)})
	if err := db.RegisterPipeline(1, PipelineFunc(func(_ context.Context, p *PipelineContext) error {
		for i := range math.MaxUint16 + 1 {
			if err := p.Draft.Set(fmt.Sprint(i), int64(i)); err != nil {
				return err
			}
		}
		return nil
	})); err != nil {
		t.Fatal(err)
	}
	tx := db.Begin(admin)
	draft, _ := tx.Update(doc.id, id, 1)
	draft.SetPipeline(1)
	_, _ = draft.Save()
	sequence := db.Sequence()
	if err := tx.Commit(); !errors.Is(err, ErrInvalidValue) {
		t.Fatal("oversized compact field count was not rejected", err)
	}
	if db.Sequence() != sequence || len(doc.fieldNames) != 1 {
		t.Fatal("invalid pipeline result reached WAL or field dictionary")
	}
	assertSecondaryIndexState(t, db)
}

func TestSecondaryIndexRepeatedDraftMutations(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(Options{Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	admin := Principal{ID: "admin"}
	doc, _ := db.CreateDocument(admin, "repeated-drafts")
	for _, mutations := range []string{"uu", "dd", "du", "ud"} {
		id := storeQueryObject(t, db, doc.id, map[string]any{"v": "original"})
		tx := db.Begin(admin)
		for i, mutation := range mutations {
			if mutation == 'd' {
				if err := tx.Delete(doc.id, id, 1); err != nil {
					t.Fatal(err)
				}
			} else {
				draft, err := tx.Update(doc.id, id, 1)
				if err != nil {
					t.Fatal(err)
				}
				_ = draft.Set("v", fmt.Sprint(i))
				_, _ = draft.Save()
			}
		}
		if err := tx.Commit(); err != nil {
			t.Fatal("repeated draft index maintenance failed", err)
		}
		assertSecondaryIndexState(t, db)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = Open(Options{Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	assertSecondaryIndexState(t, db)
}

func TestSecondaryIndexRandomMutationsAndRecovery(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(Options{Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	admin := Principal{ID: "admin"}
	doc, _ := db.CreateDocument(admin, "random-mutations")
	other, _ := db.CreateDocument(admin, "other-document")
	storeQueryObject(t, db, other.id, map[string]any{"v": "other"})
	values := []any{nil, true, int64(-1), uint64(math.MaxUint64), math.NaN(), math.Copysign(0, -1), 0.0, "abc", []byte{255}, []Value{}, []NamedValue{}}
	random := rand.New(rand.NewPCG(123, 456))
	var live []ObjectID
	for step := range 150 {
		tx := db.Begin(admin)
		drafts := 1 + random.IntN(3)
		for range drafts {
			var draft *Draft
			if len(live) == 0 || random.IntN(3) == 0 {
				draft, err = tx.Create(doc.id)
				if err == nil {
					live = append(live, draft.id)
				}
			} else {
				pos := random.IntN(len(live))
				id := live[pos]
				// Use each existing object at most once per transaction.
				duplicate := false
				for _, previous := range tx.drafts {
					duplicate = duplicate || previous.id == id
				}
				if duplicate {
					continue
				}
				if random.IntN(4) == 0 {
					if err := tx.Delete(doc.id, id, 0); err != nil {
						t.Fatal(err)
					}
					live = slices.Delete(live, pos, pos+1)
					continue
				}
				draft, err = tx.Update(doc.id, id, 0)
			}
			if err != nil {
				t.Fatal(err)
			}
			if random.IntN(4) == 0 && draft.expectedRevision != 0 {
				draft.DeleteField("v")
			} else {
				_ = draft.Set("v", values[random.IntN(len(values))])
			}
			_ = draft.Set("other", int64(step))
			_, _ = draft.Save()
		}
		if err := tx.Commit(); err != nil {
			t.Fatal(err)
		}
		assertSecondaryIndexState(t, db)
		for _, value := range values {
			predicate := condition(t, "v", Equal, value)
			if predicate.validate() == nil {
				assertIndexedQuery(t, db, admin, Query{Document: doc.id, Where: []Condition{predicate}, Limit: 3})
			}
		}
		if step%25 == 24 {
			if step%50 == 49 {
				if err := db.Checkpoint(); err != nil {
					t.Fatal(err)
				}
			}
			if err := db.Close(); err != nil {
				t.Fatal(err)
			}
			db, err = Open(Options{Dir: dir})
			if err != nil {
				t.Fatal(err)
			}
			doc = db.documents[doc.id]
			assertSecondaryIndexState(t, db)
			assertIndexedQuery(t, db, admin, Query{Document: other.id, Where: []Condition{condition(t, "v", Equal, "other")}})
		}
	}
}

func TestSecondaryIndexReplacesBorrowedKeyAndClearsTail(t *testing.T) {
	db, _ := Open(Options{})
	defer db.Close()
	admin := Principal{ID: "admin"}
	doc, _ := db.CreateDocument(admin, "borrowed-keys")
	id := storeQueryObject(t, db, doc.id, map[string]any{"v": "unchanged"})
	field, _ := doc.FieldID("v")
	previous := doc.indexes[field].nodes[doc.indexes[field].root].entry.key.text
	tx := db.Begin(admin)
	draft, _ := tx.Update(doc.id, id, 1)
	_, _ = draft.Save()
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	entry := doc.indexes[field].nodes[doc.indexes[field].root].entry
	view := ObjectView{document: doc, data: db.objects[objectKey{doc.id, id}]}
	current, _ := view.String("v")
	if unsafe.StringData(entry.key.text) != unsafe.StringData(current) || unsafe.StringData(entry.key.text) == unsafe.StringData(previous) {
		t.Fatal("unchanged value retained previous object bytes")
	}
	tx = db.Begin(admin)
	_ = tx.Delete(doc.id, id, 2)
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	index := doc.indexes[field]
	if index.len() != 0 || index.nodes[index.free].entry != (indexEntry[scalarKey]{}) {
		t.Fatal("removed index entry retained borrowed bytes")
	}
}

func TestSecondaryIndexZeroAllocations(t *testing.T) {
	index := orderedIndex[int64]{nodes: make([]indexNode[int64], 1, 9)}
	index.add(-1, 1, cmp.Compare[int64])
	index.add(1, 3, cmp.Compare[int64])
	if allocs := testing.AllocsPerRun(1000, func() {
		index.add(0, 2, cmp.Compare[int64])
		index.add(0, 4, cmp.Compare[int64])
		index.add(0, 2, cmp.Compare[int64])
		if index.lower(0, cmp.Compare[int64]) != 1 || index.upper(0, cmp.Compare[int64]) != 3 {
			panic("generic index bounds changed")
		}
		index.remove(0, 2, cmp.Compare[int64])
		index.remove(0, 4, cmp.Compare[int64])
	}); allocs != 0 {
		t.Fatalf("prepared generic secondary index: %g allocs", allocs)
	}
	db, _ := Open(Options{})
	defer db.Close()
	admin := Principal{ID: "admin"}
	doc, _ := db.CreateDocument(admin, "allocations")
	for i := range 50 {
		storeQueryObject(t, db, doc.id, map[string]any{"v": int64(i), "name": fmt.Sprintf("item-%02d", i)})
	}
	for _, predicate := range []Condition{
		condition(t, "v", Equal, int64(25)), condition(t, "v", NotEqual, int64(25)),
		condition(t, "v", Less, int64(5)), condition(t, "v", GreaterEqual, int64(45)),
		condition(t, "name", Prefix, "item-1"),
	} {
		query := Query{Document: doc.id, Where: []Condition{predicate}, Select: []string{"name"}, Limit: 3}
		var buffer QueryBuffer
		if allocs := testing.AllocsPerRun(1000, func() {
			page, err := db.FindInto(context.Background(), admin, query, &buffer)
			if err != nil || len(page.Objects) == 0 {
				panic("indexed query failed")
			}
		}); allocs != 0 {
			t.Fatalf("indexed FindInto op=%d: %g allocs", predicate.Op, allocs)
		}
	}
	// Index maintenance on compact bytes itself reuses prepared index capacity;
	// transaction draft construction and WAL record construction still allocate.
	data := db.objects[objectKey{doc.id, 1}]
	if allocs := testing.AllocsPerRun(1000, func() {
		if doc.indexObject(data, indexRemove) != nil || doc.indexObject(data, indexAdd) != nil {
			panic("index maintenance failed")
		}
	}); allocs != 0 {
		t.Fatalf("prepared encoded-object index maintenance: %g allocs", allocs)
	}
}

func TestSecondaryIndexConcurrentUpdatesAndQueries(t *testing.T) {
	db, _ := Open(Options{})
	defer db.Close()
	admin := Principal{ID: "admin"}
	doc, _ := db.CreateDocument(admin, "concurrent-secondary")
	id := storeQueryObject(t, db, doc.id, map[string]any{"v": int64(0)})
	for i := range 20 {
		storeQueryObject(t, db, doc.id, map[string]any{"other": int64(i)})
	}
	var wg sync.WaitGroup
	wg.Go(func() {
		for i := range 100 {
			tx := db.Begin(admin)
			draft, err := tx.Update(doc.id, id, 0)
			if err != nil {
				t.Error(err)
				return
			}
			_ = draft.Set("v", int64(i%2))
			_, _ = draft.Save()
			if err := tx.Commit(); err != nil {
				t.Error(err)
				return
			}
		}
	})
	for range 2 {
		wg.Go(func() {
			var buffer QueryBuffer
			for i := range 100 {
				query := Query{Document: doc.id, Where: []Condition{condition(t, "v", Equal, int64(i%2))}}
				page, err := db.FindInto(context.Background(), admin, query, &buffer)
				if err != nil || len(page.Objects) > 1 {
					t.Error("concurrent indexed query failed", err)
					return
				}
				if len(page.Objects) == 1 {
					value, _ := page.Objects[0].Int("v")
					if value != int64(i%2) {
						t.Error("index and object revision diverged")
						return
					}
				}
			}
		})
	}
	wg.Wait()
	assertSecondaryIndexState(t, db)
}

type cancelDuringPlanning struct {
	context.Context
	cancel context.CancelFunc
	calls  int
}

func (ctx *cancelDuringPlanning) Err() error {
	ctx.calls++
	if ctx.calls == 5 {
		ctx.cancel()
	}
	return ctx.Context.Err()
}

func TestSecondaryIndexPlanningCancellationAndBufferReuse(t *testing.T) {
	db, _ := Open(Options{})
	defer db.Close()
	admin := Principal{ID: "admin"}
	doc, _ := db.CreateDocument(admin, "cancel-planner")
	for i := range 100 {
		storeQueryObject(t, db, doc.id, map[string]any{"v": int64(i % 2)})
	}
	query := Query{Document: doc.id, Where: []Condition{condition(t, "v", Equal, int64(1))}}
	base, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx := &cancelDuringPlanning{Context: base, cancel: cancel}
	var buffer QueryBuffer
	if _, err := db.FindInto(ctx, admin, query, &buffer); !errors.Is(err, context.Canceled) {
		t.Fatal("planner ignored cancellation", err)
	}
	if len(buffer.candidates) == 0 {
		t.Fatal("cancellation test did not reach candidate materialization")
	}
	page, err := db.FindInto(context.Background(), admin, query, &buffer)
	if err != nil || len(page.Objects) != 50 || page.HasMore {
		t.Fatal("canceled query corrupted reused buffer", err)
	}
}

func BenchmarkSecondaryIndexFindInto(b *testing.B) {
	for _, count := range []int{1000, 10000} {
		b.Run(fmt.Sprint(count), func(b *testing.B) {
			db, _ := Open(Options{})
			defer db.Close()
			admin := Principal{ID: "admin"}
			doc, _ := db.CreateDocument(admin, "benchmark-index")
			for i := range count {
				storeQueryObject(b, db, doc.id, map[string]any{"v": int64(i), "name": fmt.Sprintf("item-%05d", i)})
			}
			for _, test := range []struct {
				name  string
				where Condition
				want  int
			}{
				{"equal", condition(b, "v", Equal, int64(count/2)), 1},
				{"range", condition(b, "v", GreaterEqual, int64(count-10)), 10},
				{"prefix", condition(b, "name", Prefix, "item-0000"), 10},
			} {
				b.Run(test.name, func(b *testing.B) {
					query := Query{Document: doc.id, Where: []Condition{test.where}, Select: []string{"name"}}
					var buffer QueryBuffer
					_, _ = db.FindInto(context.Background(), admin, query, &buffer)
					b.ReportAllocs()
					b.ResetTimer()
					for b.Loop() {
						page, err := db.FindInto(context.Background(), admin, query, &buffer)
						if err != nil || len(page.Objects) != test.want {
							b.Fatal(err)
						}
					}
				})
			}
		})
	}
}

func FuzzScalarIndexRanges(f *testing.F) {
	f.Add([]byte("abc\x00\xff\xff\xff\xffabcdefghi"), []byte("abc"), byte(KindString), byte(Prefix-1))
	f.Add([]byte{0, 255, 255, 255, 255, 255, 255, 255, 0}, []byte{0}, byte(KindFloat), byte(Equal-1))
	f.Add([]byte{255, 255, 255, 255, 255, 255, 255, 255}, []byte{255}, byte(KindBytes), byte(Prefix-1))
	f.Fuzz(func(t *testing.T, data, operand []byte, kindByte, opByte byte) {
		if len(data) > 1024 || len(operand) > 128 {
			t.Skip()
		}
		kind := Kind(kindByte % byte(KindTime+1))
		op := CompareOp(opByte%byte(Prefix) + 1)
		if op == Exists || op == Missing {
			return
		}
		valueFor := func(raw []byte) Value {
			value := Value{kind: kind}
			switch kind {
			case KindBool:
				if len(raw) > 0 {
					value.num = uint64(raw[0] & 1)
				}
			case KindInt, KindUint, KindFloat, KindTime:
				var bits [8]byte
				copy(bits[:], raw)
				value.num = binary.LittleEndian.Uint64(bits[:])
			case KindString:
				value.text = string(raw)
			case KindBytes:
				value.data = raw
			}
			return value
		}
		expected := valueFor(operand)
		condition := preparedCondition{1, op, expected}
		if (Condition{Field: "v", Op: op, Value: expected}).validate() != nil {
			return
		}
		var index orderedIndex[scalarKey]
		var want []ObjectID
		for pos := 0; pos < len(data); pos += 8 {
			value := valueFor(data[pos:min(pos+8, len(data))])
			view, _, _ := readValue(appendValue(nil, value))
			id := ObjectID(pos/8 + 1)
			if key, ok := scalarIndexKey(view); ok {
				index.add(key, id, compareScalarKey)
			}
			if matchComparison(view, op, expected) {
				want = append(want, id)
			}
		}
		ranges := scalarIndexRanges(&index, condition)
		var got []ObjectID
		for _, span := range [2]indexSpan[scalarKey]{ranges.first, ranges.second} {
			cursor := span.cursor()
			for entry, ok := cursor.next(); ok; entry, ok = cursor.next() {
				got = append(got, entry.id)
			}
		}
		slices.Sort(got)
		if !slices.Equal(got, want) {
			t.Fatalf("index bounds kind=%v op=%v got=%v want=%v", kind, op, got, want)
		}
	})
}
