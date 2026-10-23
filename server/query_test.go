package server

import (
	"testing"

	"github.com/diiyw/nodis"
	"github.com/diiyw/nodis/protocol"
)

func TestServerQueryExecutionZeroAllocations(t *testing.T) {
	db, err := nodis.Open(nodis.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	admin := nodis.Principal{ID: "admin"}
	doc, _ := db.CreateDocument(admin, "queries")
	for i := range 3 {
		tx := db.Begin(admin)
		draft, err := tx.Create(doc.ID())
		if err != nil {
			t.Fatal(err)
		}
		_ = draft.Set("age", int64(i))
		_ = draft.Set("name", "Ada")
		_, _ = draft.Save()
		if err := tx.Commit(); err != nil {
			t.Fatal(err)
		}
	}
	filter, err := nodis.Filter("age", nodis.GreaterEqual, int64(1))
	if err != nil {
		t.Fatal(err)
	}
	request, err := protocol.AppendQuery(nil, nodis.Query{Document: doc.ID(), Where: []nodis.Condition{filter}, Select: []string{"name"}})
	if err != nil {
		t.Fatal(err)
	}
	var wire serverWire
	var decoder protocol.QueryPageDecoder
	if allocs := testing.AllocsPerRun(1000, func() {
		status, response := wire.find(db, admin, request)
		if status != protocol.StatusOK {
			panic("query rejected")
		}
		page, err := protocol.DecodeQueryPageInto(response, &decoder)
		if err != nil || len(page.Objects) != 2 {
			panic("query result changed")
		}
	}); allocs != 0 {
		t.Fatalf("server query execution/codecs: %g allocs", allocs)
	}
}
