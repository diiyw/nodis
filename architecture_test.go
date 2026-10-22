//go:build !wasm

package nodis

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/patch"
	"github.com/diiyw/nodis/redis"
	"github.com/diiyw/nodis/storage"
)

func testDB(t testing.TB) *Nodis {
	t.Helper()
	n := Open(&Options{})
	t.Cleanup(func() {
		if err := n.Close(); err != nil {
			t.Error(err)
		}
	})
	return n
}
func testConn() *redis.Conn {
	return &redis.Conn{Writer: redis.NewWriter(io.Discard), WatchKeys: make(map[string]bool)}
}
func command(n *Nodis, c *redis.Conn, name string, args ...string) string {
	c.Reset()
	GetCommand(name)(n, c, redis.Command{Name: name, Args: args})
	return string(c.Bytes())
}
func TestPersistenceRestart(t *testing.T) {
	path := t.TempDir()
	open := func() *Nodis {
		t.Helper()
		n, err := OpenWithError(&Options{Storage: storage.NewPebble(path, nil)})
		if err != nil {
			t.Fatal(err)
		}
		return n
	}
	n := open()
	n.Set("deleted", []byte("old"), false)
	n.Set("renamed", []byte("move"), false)
	n.Set("ttl", []byte("v"), false)
	if err := n.Flush(); err != nil {
		t.Fatal(err)
	}
	n.Del("deleted")
	if err := n.Rename("renamed", "dst"); err != nil {
		t.Fatal(err)
	}
	n.Expire("ttl", 60)
	if err := n.Close(); err != nil {
		t.Fatal(err)
	}
	n = open()
	defer n.Close()
	if n.Get("deleted") != nil || n.Get("renamed") != nil {
		t.Fatal("deleted source resurrected")
	}
	if string(n.Get("dst")) != "move" || n.PTTL("ttl") <= 0 {
		t.Fatal("value or TTL not durable")
	}
	count := 0
	if err := n.store.ss.ScanKeys(func(*ds.Key) bool { count++; return true }); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("duplicate physical records: %d", count)
	}
	n.Persist("ttl")
	if err := n.Flush(); err != nil {
		t.Fatal(err)
	}
	n.Clear()
	n.Set("after-clear", []byte("new"), false)
	if err := n.Close(); err != nil {
		t.Fatal(err)
	}
	n = open()
	defer n.Close()
	if len(n.Keys("*")) != 1 || string(n.Get("after-clear")) != "new" {
		t.Fatal("clear was not committed with following writes")
	}
}

type failStorage struct {
	*storage.Memory
	fail bool
}

func (s *failStorage) Apply(m []storage.Mutation, clear bool) error {
	if s.fail {
		return errors.New("injected disk failure")
	}
	return s.Memory.Apply(m, clear)
}
func TestFailedFlushRetainsState(t *testing.T) {
	ss := &failStorage{Memory: storage.NewMemory()}
	n := Open(&Options{Storage: ss})
	defer n.Close()
	n.Set("deleted", []byte("old"), false)
	if err := n.Flush(); err != nil {
		t.Fatal(err)
	}
	ss.fail = true
	n.Del("deleted")
	n.Set("hot", []byte("v"), false)
	for i := 0; i < 3; i++ {
		if err := n.store.gc(); err == nil {
			t.Fatal("expected error")
		}
	}
	if string(n.Get("hot")) != "v" {
		t.Fatal("only copy evicted")
	}
	if err := n.Close(); err == nil {
		t.Fatal("Close falsely reported success")
	}
	ss.fail = false
	if err := n.Flush(); err != nil {
		t.Fatal(err)
	}
	if _, err := ss.Get(ds.NewKey("deleted", 0)); !errors.Is(err, storage.ErrKeyNotFound) {
		t.Fatal("lost tombstone")
	}
	if err := n.Close(); err != nil {
		t.Fatal(err)
	}
}
func TestConcurrentOperationsAndGC(t *testing.T) {
	n := testDB(t)
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 32; i++ {
		wg.Go(func() {
			<-start
			for j := 0; j < 100; j++ {
				if _, err := n.Incr("new-counter"); err != nil {
					t.Error(err)
				}
			}
		})
	}
	wg.Go(func() {
		<-start
		for i := 0; i < 50; i++ {
			if err := n.store.gc(); err != nil {
				t.Error(err)
			}
			n.Scan(0, "*", 10, 0)
		}
	})
	close(start)
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("deadlock")
	}
	if got := string(n.Get("new-counter")); got != "3200" {
		t.Fatalf("lost first writes: %s", got)
	}
	if err := n.Rename("new-counter", "new-counter"); err != nil {
		t.Fatal(err)
	}
	n.LPush("q", []byte("a"), []byte("b"))
	if n.RPopLPush("q", "q") == nil {
		t.Fatal("same-key move")
	}
	n.ZAdd("z", "a", 2)
	if got := n.ZUnionStore("z", []string{"z"}, nil, "SUM"); got != 1 {
		t.Fatal(got)
	}
}
func TestColdReads(t *testing.T) {
	n := testDB(t)
	n.Set("hot", []byte("v"), false)
	n.store.gc()
	n.store.gc()
	n.store.gc()
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Go(func() {
			for j := 0; j < 30; j++ {
				if string(n.Get("hot")) != "v" {
					t.Error("cold read")
				}
			}
		})
	}
	wg.Wait()
}

func TestHotReadsRefreshGC(t *testing.T) {
	n := testDB(t)
	n.Set("hot", []byte("v"), false)
	if err := n.store.gc(); err != nil {
		t.Fatal(err)
	}
	if string(n.Get("hot")) != "v" {
		t.Fatal("read failed")
	}
	if err := n.store.gc(); err != nil {
		t.Fatal(err)
	}
	if n.store.metadata["hot"].value == nil {
		t.Fatal("hot read did not refresh the cache")
	}
	if err := n.store.gc(); err != nil {
		t.Fatal(err)
	}
	if n.store.metadata["hot"].value != nil {
		t.Fatal("idle key retained accumulated access credit")
	}
}

func TestSharedReadsWithWritesAndGC(t *testing.T) {
	n := testDB(t)
	values := [][]byte{[]byte("first"), []byte("second")}
	n.Set("key", values[0], false)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Go(func() {
			for j := 0; j < 100; j++ {
				v := n.Get("key")
				if !bytes.Equal(v, values[0]) && !bytes.Equal(v, values[1]) {
					t.Errorf("unexpected value: %q", v)
				}
				if n.Exists("key") != 1 {
					t.Error("key disappeared")
				}
			}
		})
	}
	wg.Go(func() {
		for i := 0; i < 100; i++ {
			n.Set("key", values[i%2], false)
			if err := n.store.gc(); err != nil {
				t.Error(err)
			}
		}
	})
	wg.Wait()
	if err := n.Close(); err != nil {
		t.Fatal(err)
	}
	if n.Get("key") != nil || n.Exists("key") != 0 {
		t.Fatal("shared reads returned data after close")
	}
}

func TestGCProcessesBoundedBatch(t *testing.T) {
	n := testDB(t)
	for i := 0; i < gcBatchSize+44; i++ {
		n.Set(fmt.Sprintf("gc-%d", i), []byte("v"), false)
	}
	if err := n.store.gc(); err != nil {
		t.Fatal(err)
	}
	visited, remaining := 0, 0
	for _, m := range n.store.metadata {
		if !m.accessed.Load() {
			visited++
		} else {
			remaining++
		}
	}
	if visited != gcBatchSize || remaining != 44 {
		t.Fatalf("GC visited %d entries; %d remain untouched", visited, remaining)
	}
}

func TestExpiredKeyRecreated(t *testing.T) {
	n := testDB(t)
	n.SetPX("k", []byte("old"), -1)
	if v, err := n.Incr("k"); err != nil || v != 1 {
		t.Fatal(v, err)
	}
	if string(n.Get("k")) != "1" || n.PTTL("k") != -1 {
		t.Fatal("old expiration retained")
	}
}
func TestOptionsNotMutated(t *testing.T) {
	opts := Options{}
	a := Open(&opts)
	b := Open(&opts)
	defer a.Close()
	defer b.Close()
	a.Set("k", []byte("v"), false)
	a.Flush()
	if opts.Storage != nil || b.Get("k") != nil {
		t.Fatal("shared default storage")
	}
}
func TestTransactionsAndWatch(t *testing.T) {
	n := testDB(t)
	c := testConn()
	n.Set("k", []byte("old"), false)
	command(n, c, "WATCH", "k")
	command(n, c, "MULTI")
	command(n, c, "SET", "k", "new")
	n.Del("k")
	if got := command(n, c, "EXEC"); got != "*-1\r\n" {
		t.Fatalf("WATCH not invalidated: %q", got)
	}
	command(n, c, "WATCH", "k")
	command(n, c, "UNWATCH")
	if len(n.store.watchedKeys) != 0 {
		t.Fatal("watch subscription leaked")
	}
	command(n, c, "MULTI")
	if got := command(n, c, "SET", "k", "first"); got != "+QUEUED\r\n" {
		t.Fatal(got)
	}
	command(n, c, "SET", "k", "final")
	if got := command(n, c, "EXEC"); got != "*2\r\n+OK\r\n+OK\r\n" {
		t.Fatal(got)
	}
	// An embedded observer started during an EXEC execution context must wait.
	entered := make(chan struct{})
	done := make(chan []byte, 1)
	err := n.exec(func(tx *Tx) error {
		scoped := n.inTx(tx)
		scoped.Set("k", []byte("intermediate"), false)
		go func() { close(entered); done <- n.Get("k") }()
		<-entered
		select {
		case <-done:
			t.Fatal("intermediate state visible")
		default:
		}
		scoped.Set("k", []byte("final"), false)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(<-done) != "final" {
		t.Fatal("transaction isolation")
	}
}
func TestBlockingConsumers(t *testing.T) {
	n := testDB(t)
	results := make(chan string, 2)
	for i := 0; i < 2; i++ {
		go func() { _, v := n.BLPop(time.Second, "q"); results <- string(v) }()
	}
	// Registration is inspected under its owner lock, avoiding timing assumptions.
	deadline := time.Now().Add(time.Second)
	for {
		n.blockingKeysMutex.Lock()
		ready := len(n.blockingKeys["q"]) == 2
		n.blockingKeysMutex.Unlock()
		if ready {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("not registered")
		}
		time.Sleep(time.Millisecond)
	}
	n.LPush("q", []byte("one"), []byte("two"))
	a, b := <-results, <-results
	if a == b || a == "" || b == "" {
		t.Fatal(a, b)
	}
}
func TestListenerOrderAndStableIDs(t *testing.T) {
	n := testDB(t)
	first := n.WatchKey([]string{"*"}, func(patch.Op) {})
	values := make(chan string, 20)
	last := n.WatchKey([]string{"*"}, func(op patch.Op) {
		if op.Type == patch.OpTypeSet {
			values <- string(op.Data.(*patch.OpSet).Value)
		}
	})
	n.UnWatchKey(first)
	n.UnWatchKey(first)
	for i := 0; i < 20; i++ {
		n.Set("k", []byte(fmt.Sprint(i)), false)
	}
	for i := 0; i < 20; i++ {
		select {
		case v := <-values:
			if v != fmt.Sprint(i) {
				t.Fatal("reordered", v, i)
			}
		case <-time.After(time.Second):
			t.Fatal("lost event")
		}
	}
	n.UnWatchKey(last)
	n.UnWatchKey(last)
}
func TestPatchTTLAndSnapshot(t *testing.T) {
	n := testDB(t)
	expiry := time.Now().Add(time.Minute).UnixMilli()
	if err := n.ApplyPatch(patch.Op{Type: patch.OpTypeSet, Data: &patch.OpSet{Key: "k", Value: []byte("v"), Expiration: expiry}}); err != nil {
		t.Fatal(err)
	}
	if got := n.PTTL("k"); got <= 0 || got > 60000 {
		t.Fatal(got)
	}
	if err := n.ApplyPatch(patch.Op{Type: patch.OpTypeExpire, Data: &patch.OpExpire{Key: "k", Expiration: expiry}}); err != nil {
		t.Fatal(err)
	}
	if got := n.PTTL("k"); got <= 0 || got > 60000 {
		t.Fatal(got)
	}
	if err := n.Snapshot(); err != nil {
		t.Fatal(err)
	}
	if _, err := n.store.ss.Get(ds.NewKey("k", 0)); err != nil {
		t.Fatal("snapshot skipped pending writes", err)
	}
}
func TestStreamFullSyncAndChanges(t *testing.T) {
	primary := testDB(t)
	replica := testDB(t)
	primary.SetEX("ttl", []byte("initial"), 60)
	primary.SAdd("set", strings.Repeat("x", 64))
	replica.Set("stale", []byte("remove"), false)
	publish := make(chan error, 1)
	go func() { publish <- primary.Broadcast("127.0.0.1:0", []string{"*"}) }()
	stream := primary.options.Channel.(*Stream)
	var addr string
	deadline := time.Now().Add(time.Second)
	for addr == "" {
		stream.mu.Lock()
		if stream.listener != nil {
			addr = stream.listener.Addr().String()
		}
		stream.mu.Unlock()
		if time.Now().After(deadline) {
			t.Fatal("server did not start")
		}
		time.Sleep(time.Millisecond)
	}
	subscribed := make(chan error, 1)
	go func() { subscribed <- replica.Subscribe(addr) }()
	await := func(fn func() bool) {
		t.Helper()
		until := time.Now().Add(3 * time.Second)
		for !fn() {
			select {
			case err := <-subscribed:
				t.Fatalf("subscribe: %v", err)
			default:
			}
			if time.Now().After(until) {
				t.Fatal("replication timeout")
			}
			time.Sleep(time.Millisecond)
		}
	}
	await(func() bool {
		return string(replica.Get("ttl")) == "initial" && replica.Get("stale") == nil && replica.SCard("set") == 1
	})
	c := testConn()
	command(primary, c, "MULTI")
	command(primary, c, "SET", "a", "one")
	command(primary, c, "SET", "b", "two")
	command(primary, c, "EXEC")
	await(func() bool { return string(replica.Get("b")) == "two" })
	if string(replica.Get("a")) != "one" || replica.PTTL("ttl") <= 0 {
		t.Fatal("transaction/ttl mismatch")
	}
	primary.Del("a")
	await(func() bool { return replica.Get("a") == nil })
	primary.Clear()
	await(func() bool { return len(replica.Keys("*")) == 0 })
	primary.Close()
	replica.Close()
	select {
	case <-subscribed:
	case <-time.After(time.Second):
		t.Fatal("subscriber leaked")
	}
	select {
	case <-publish:
	case <-time.After(time.Second):
		t.Fatal("publisher leaked")
	}
}
func TestCommandMalformedOptions(t *testing.T) {
	n := testDB(t)
	c := testConn()
	if got := command(n, c, "SCAN", "0", "MATCH"); !strings.HasPrefix(got, "-ERR") {
		t.Fatal(got)
	} // no parsed options: handler should also validate raw syntax
	if got := command(n, c, "SET", "k", "EX"); got != "+OK\r\n" {
		t.Fatal(got)
	}
	if !bytes.Equal(n.Get("k"), []byte("EX")) {
		t.Fatal("option word treated as data")
	}
}
func BenchmarkHotGet(b *testing.B) {
	n := testDB(b)
	n.Set("hot", []byte("v"), false)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		n.Get("hot")
	}
}
func BenchmarkParallelHotGet(b *testing.B) {
	n := testDB(b)
	n.Set("hot", []byte("v"), false)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n.Get("hot")
		}
	})
}
func BenchmarkHotSet(b *testing.B) {
	n := testDB(b)
	value := []byte("v")
	n.Set("hot", value, false)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		n.Set("hot", value, false)
	}
}
func BenchmarkHotExists(b *testing.B) {
	n := testDB(b)
	n.Set("hot", []byte("v"), false)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		n.Exists("hot")
	}
}
func BenchmarkParallelHotExists(b *testing.B) {
	n := testDB(b)
	n.Set("hot", []byte("v"), false)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n.Exists("hot")
		}
	})
}

func TestHotOperationsZeroAllocs(t *testing.T) {
	n := testDB(t)
	value := []byte("v")
	n.Set("hot", value, false)
	for name, fn := range map[string]func(){"GET": func() { n.Get("hot") }, "SET": func() { n.Set("hot", value, false) }, "EXISTS": func() { n.Exists("hot") }} {
		t.Run(name, func(t *testing.T) {
			if got := testing.AllocsPerRun(1000, fn); got != 0 {
				t.Fatalf("%g allocations/op", got)
			}
		})
	}
}
func TestInvalidSetDoesNotMutate(t *testing.T) {
	n := testDB(t)
	c := testConn()
	n.Set("k", []byte("old"), false)
	for _, args := range [][]string{{"k", "new", "EX", "bad"}, {"k", "new", "EX", "0"}, {"k", "new", "EX", "1", "PX", "1"}, {"k", "new", "KEEPTTL", "EX", "1"}, {"k", "new", "NX", "XX"}} {
		if got := command(n, c, "SET", args...); !strings.HasPrefix(got, "-ERR") {
			t.Fatal(got)
		}
		if string(n.Get("k")) != "old" {
			t.Fatal("invalid command changed data")
		}
	}
	command(n, c, "MULTI")
	command(n, c, "SET", "k", "new", "EX", "bad")
	if got := command(n, c, "EXEC"); !strings.HasPrefix(got, "-EXECABORT") {
		t.Fatal(got)
	}
}
func TestMalformedStateBatchDoesNotReset(t *testing.T) {
	n := testDB(t)
	n.Set("old", []byte("keep"), false)
	batch := patch.Op{Type: patch.OpTypeBatch, Data: &patch.OpBatch{Frames: [][]byte{(patch.Op{Type: patch.OpTypeReset, Data: &patch.OpReset{Patterns: []string{"*"}}}).Encode(), (patch.Op{Type: patch.OpTypeRestore, Data: &patch.OpRestore{Key: "bad", Value: []byte{255}}}).Encode()}}}
	if err := n.ApplyPatch(batch); err == nil {
		t.Fatal("accepted corrupt state")
	}
	if string(n.Get("old")) != "keep" {
		t.Fatal("partially applied malformed batch")
	}
}
func TestCloseStopsWorkers(t *testing.T) {
	n := Open(&Options{GCDuration: time.Millisecond, SnapshotDuration: time.Millisecond})
	n.Set("k", []byte("v"), false)
	if err := n.Close(); err != nil {
		t.Fatal(err)
	}
	if err := n.Close(); err != nil {
		t.Fatal(err)
	}
	if err := n.Flush(); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
	if id := n.WatchKey([]string{"*"}, func(patch.Op) {}); id != -1 {
		t.Fatal("listener created after close")
	}
}

func TestWebsocketStateSync(t *testing.T) {
	primary := Open(&Options{Channel: NewWebsocket()})
	replica := Open(&Options{Channel: NewWebsocket()})
	defer primary.Close()
	defer replica.Close()
	primary.Set("scope/a", []byte("one"), false)
	replica.Set("unrelated", []byte("keep"), false)
	published := make(chan error, 1)
	go func() { published <- primary.Broadcast("127.0.0.1:0", []string{"scope/*"}) }()
	ws := primary.options.Channel.(*Websocket)
	var addr string
	deadline := time.Now().Add(time.Second)
	for addr == "" {
		ws.mu.Lock()
		if ws.server != nil {
			addr = ws.server.Addr
		}
		ws.mu.Unlock()
		if time.Now().After(deadline) {
			t.Fatal("WebSocket not started")
		}
		time.Sleep(time.Millisecond)
	}
	subscribed := make(chan error, 1)
	go func() { subscribed <- replica.Subscribe("ws://" + addr) }()
	await := func(fn func() bool) {
		t.Helper()
		deadline := time.Now().Add(2 * time.Second)
		for !fn() {
			select {
			case err := <-subscribed:
				t.Fatal(err)
			default:
			}
			if time.Now().After(deadline) {
				t.Fatal("sync timeout")
			}
			time.Sleep(time.Millisecond)
		}
	}
	await(func() bool { return string(replica.Get("scope/a")) == "one" })
	if err := primary.Rename("scope/a", "outside"); err != nil {
		t.Fatal(err)
	}
	await(func() bool { return replica.Get("scope/a") == nil })
	if replica.Get("outside") != nil {
		t.Fatal("rename escaped subscription filter")
	}
	if err := primary.Rename("outside", "scope/b"); err != nil {
		t.Fatal(err)
	}
	await(func() bool { return string(replica.Get("scope/b")) == "one" })
	primary.Clear()
	await(func() bool { return replica.Get("scope/b") == nil })
	if string(replica.Get("unrelated")) != "keep" {
		t.Fatal("filtered clear removed unrelated state")
	}
	primary.Close()
	replica.Close()
	select {
	case <-subscribed:
	case <-time.After(time.Second):
		t.Fatal("WebSocket subscriber leaked")
	}
	select {
	case <-published:
	case <-time.After(time.Second):
		t.Fatal("WebSocket publisher leaked")
	}
}
func TestRangeBoundsAndMoveTypeValidation(t *testing.T) {
	n := testDB(t)
	n.ZAdd("z", "a", 1)
	n.ZAdd("z", "b", 2)
	if got := n.ZRange("z", 0, int64(^uint64(0)>>1)); len(got) != 2 {
		t.Fatal(got)
	}
	if got := n.ZRange("z", 1, 1); len(got) != 1 || got[0] != "b" {
		t.Fatal(got)
	}
	if got := n.ZRevRange("z", 1, 1); len(got) != 1 || got[0] != "a" {
		t.Fatal(got)
	}
	n.SAdd("source", "member")
	n.Set("wrongtype", []byte("v"), false)
	c := testConn()
	command(n, c, "SMOVE", "source", "wrongtype", "member")
	if !n.SIsMember("source", "member") {
		t.Fatal("failed move removed source")
	}
	command(n, c, "RPOPLPUSH", "wrongtype", "new-list")
	if n.Exists("new-list") != 0 {
		t.Fatal("failed move created destination")
	}
}

type readFailureStorage struct{ *storage.Memory }

func (s *readFailureStorage) Get(*ds.Key) (ds.Value, error) {
	return nil, errors.New("injected read failure")
}
func TestStorageReadErrorIsOneReply(t *testing.T) {
	ss := &readFailureStorage{storage.NewMemory()}
	n := Open(&Options{Storage: ss})
	defer n.Close()
	n.Set("k", []byte("v"), false)
	n.store.gc()
	n.store.gc()
	n.store.gc()
	got := command(n, testConn(), "GET", "k")
	if strings.Count(got, "\r\n") != 1 || !strings.HasPrefix(got, "-ERR") {
		t.Fatalf("invalid error reply: %q", got)
	}
}
func TestGeoChangesAreDirty(t *testing.T) {
	n := testDB(t)
	member := &GeoMember{Member: "place", Longitude: 10, Latitude: 20}
	n.GeoAdd("geo", member)
	if err := n.Flush(); err != nil {
		t.Fatal(err)
	}
	n.GeoAdd("geo", &GeoMember{Member: "place", Longitude: 11, Latitude: 21})
	if err := n.Flush(); err != nil {
		t.Fatal(err)
	}
	stored, err := n.store.ss.Get(ds.NewKey("geo", 0))
	if err != nil {
		t.Fatal(err)
	}
	expected, _ := n.ZScore("geo", "place")
	decoded, err := storage.DecodeValue(storage.EncodeValue(stored))
	if err != nil {
		t.Fatal(err)
	}
	temp := testDB(t)
	if err := temp.ApplyPatch(patch.Op{Type: patch.OpTypeRestore, Data: &patch.OpRestore{Key: "geo", Value: storage.EncodeValue(decoded)}}); err != nil {
		t.Fatal(err)
	}
	actual, _ := temp.ZScore("geo", "place")
	if actual != expected {
		t.Fatal("GEO update was not flushed")
	}
}
