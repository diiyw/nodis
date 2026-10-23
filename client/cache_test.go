package client

import (
	"encoding/binary"
	"errors"
	"slices"
	"testing"

	"github.com/diiyw/nodis"
)

func objectEvent(id, revision, sequence, permission uint64, fields map[string]string) nodis.Event {
	data := make([]byte, 30)
	copy(data, "NDX1")
	binary.LittleEndian.PutUint16(data[4:6], uint16(len(fields)))
	binary.LittleEndian.PutUint64(data[6:14], id)
	binary.LittleEndian.PutUint64(data[14:22], revision)
	names := make([]string, 0, len(fields))
	for name := range fields {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		data = binary.AppendUvarint(data, uint64(len(name)))
		data = append(data, name...)
		value, _ := nodis.ValueOf(fields[name])
		data = nodis.AppendValue(data, value)
	}
	return nodis.Event{Type: nodis.EventSnapshotObject, Document: 1, Object: nodis.ObjectID(id), Revision: revision, Sequence: sequence, PermissionRevision: permission, Data: data}
}

func snapshotDone(sequence, permission uint64) nodis.Event {
	return nodis.Event{Type: nodis.EventSnapshotDone, Sequence: sequence, PermissionRevision: permission}
}

func TestCacheAtomicSnapshotsOwnTheirBytes(t *testing.T) {
	var cache viewCache
	cache.init(true)
	first := objectEvent(1, 1, 5, 1, map[string]string{"name": "Ada", "secret": "old"})
	if err := cache.stage(first); err != nil {
		t.Fatal(err)
	}
	if _, ok := cache.cached(1, 1); ok {
		t.Fatal("partial snapshot exposed")
	}
	if err := cache.stage(snapshotDone(5, 1)); err != nil {
		t.Fatal(err)
	}
	old, ok := cache.cached(1, 1)
	if !ok {
		t.Fatal("completed snapshot missing")
	}
	clear(first.Data)
	if name, _ := old.String("name"); name != "Ada" {
		t.Fatal("cache retained borrowed frame")
	}
	cache.invalidate(errors.New("connection lost"))
	if _, ok := cache.cached(1, 1); ok || cache.syncStatus().Ready {
		t.Fatal("disconnected cache remained visible")
	}
	partial := objectEvent(2, 1, 7, 2, map[string]string{"name": "partial"})
	if err := cache.stage(partial); err != nil {
		t.Fatal(err)
	}
	if err := cache.stage(snapshotDone(8, 2)); err == nil {
		t.Fatal("accepted mismatched completion marker")
	}
	cache.invalidate(nodis.ErrCorruptData)
	newEvent := objectEvent(1, 2, 9, 3, map[string]string{"name": "Grace"})
	if err := cache.stage(newEvent); err != nil {
		t.Fatal(err)
	}
	if _, ok := cache.cached(1, 1); ok {
		t.Fatal("replacement snapshot exposed before completion")
	}
	if err := cache.stage(snapshotDone(9, 3)); err != nil {
		t.Fatal(err)
	}
	if _, ok := cache.cached(1, 2); ok {
		t.Fatal("failed snapshot survived replacement")
	}
	current, _ := cache.cached(1, 1)
	if _, ok := current.String("secret"); ok {
		t.Fatal("removed field retained")
	}
	if value, _ := old.String("secret"); value != "old" {
		t.Fatal("previously returned immutable object changed")
	}
	clear(newEvent.Data)
	if name, _ := current.String("name"); name != "Grace" {
		t.Fatal("replacement borrowed source bytes")
	}
	status := cache.syncStatus()
	if !status.Enabled || !status.Ready || status.Sequence != 9 || status.PermissionRevision != 3 || status.Err != nil {
		t.Fatalf("status = %+v", status)
	}
}

func TestCacheRejectsInvalidSnapshotAndLiveEvents(t *testing.T) {
	for _, damage := range []string{"duplicate", "id", "revision", "permission", "live-in-snapshot"} {
		t.Run(damage, func(t *testing.T) {
			var cache viewCache
			cache.init(true)
			event := objectEvent(1, 1, 5, 2, map[string]string{"name": "Ada"})
			if err := cache.stage(event); err != nil {
				t.Fatal(err)
			}
			switch damage {
			case "id":
				event.Object++
			case "revision":
				event = objectEvent(2, 1, 5, 2, map[string]string{"name": "Ada"})
				event.Revision++
			case "permission":
				event.PermissionRevision++
			case "live-in-snapshot":
				event.Type = nodis.EventObjectUpdated
			}
			if err := cache.stage(event); err == nil {
				t.Fatal("invalid snapshot accepted")
			}
			if cache.syncStatus().Ready {
				t.Fatal("invalid snapshot became visible")
			}
		})
	}
	var cache viewCache
	cache.init(true)
	if err := cache.stage(snapshotDone(5, 2)); err != nil {
		t.Fatal(err)
	}
	if err := cache.apply(nodis.Event{Type: nodis.EventObjectDeleted, Sequence: 4, PermissionRevision: 2}); err == nil {
		t.Fatal("out of order mutation accepted")
	}
	if err := cache.apply(nodis.Event{Type: nodis.EventDocumentDeleted, Sequence: 6, PermissionRevision: 1}); err == nil {
		t.Fatal("old permissions accepted")
	}
	cache.close()
	cache.invalidate(nil)
	if !errors.Is(cache.syncStatus().Err, nodis.ErrClosed) {
		t.Fatal("close status overwritten")
	}
	if err := cache.stage(snapshotDone(7, 3)); !errors.Is(err, nodis.ErrClosed) {
		t.Fatal("closed cache accepted snapshot")
	}
}

func TestCacheLookupStatusAndInvalidationZeroAllocations(t *testing.T) {
	var cache viewCache
	cache.init(true)
	event := objectEvent(1, 1, 1, 1, map[string]string{"name": "Ada"})
	if err := cache.stage(event); err != nil {
		t.Fatal(err)
	}
	if err := cache.stage(snapshotDone(1, 1)); err != nil {
		t.Fatal(err)
	}
	if allocs := testing.AllocsPerRun(1000, func() {
		value, ok := cache.cached(1, 1)
		if !ok {
			panic("cache lookup failed")
		}
		if name, _ := value.String("name"); name != "Ada" {
			panic("name changed")
		}
		if !cache.syncStatus().Ready {
			panic("cache not ready")
		}
	}); allocs != 0 {
		t.Fatalf("lookup/status: %g allocs", allocs)
	}
	if allocs := testing.AllocsPerRun(1000, func() {
		cache.invalidate(nil)
		if err := cache.stage(snapshotDone(1, 1)); err != nil {
			panic(err)
		}
	}); allocs != 0 {
		t.Fatalf("invalidate/snapshot publication: %g allocs", allocs)
	}
}
