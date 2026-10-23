package protocol

import (
	"testing"

	"github.com/diiyw/nodis"
)

func TestEventRoundTripAndAllocation(t *testing.T) {
	event := nodis.Event{Type: nodis.EventObjectUpdated, Sequence: 9, PermissionRevision: 3, Document: 2, Object: 7, Revision: 4, Data: []byte("object")}
	buffer := make([]byte, 0, 128)
	frame := AppendEvent(buffer, event)
	decoded, err := DecodeEvent(frame)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Sequence != event.Sequence || decoded.Object != event.Object || string(decoded.Data) != "object" {
		t.Fatalf("decoded = %#v", decoded)
	}
	if allocs := testing.AllocsPerRun(1000, func() { _ = AppendEvent(buffer[:0], event) }); allocs != 0 {
		t.Fatalf("encoding allocated: %v", allocs)
	}
	frame[len(frame)-1] ^= 1
	if _, err := DecodeEvent(frame); err == nil {
		t.Fatal("accepted bad checksum")
	}
}
