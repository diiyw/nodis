package nodis

import (
	"errors"

	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/ds/hash"
	"github.com/diiyw/nodis/ds/list"
	"github.com/diiyw/nodis/ds/set"
	"github.com/diiyw/nodis/ds/str"
	"github.com/diiyw/nodis/ds/zset"
	"github.com/diiyw/nodis/pb"
)

var (
	ErrCorruptedData = errors.New("corrupted data")
)

// newEntry creates a new entity
func newEntry(key string, value ds.Value, expiration int64) *pb.Entry {
	e := &pb.Entry{
		Key:        key,
		Expiration: expiration,
		Type:       uint32(value.Type()),
	}
	switch value.Type() {
	case ds.String:
		e.Value = &pb.Entry_StringValue{
			StringValue: &pb.StringValue{
				Value: value.(*str.String).GetValue(),
			},
		}
	case ds.ZSet:
		e.Value = &pb.Entry_ZSetValue{
			ZSetValue: &pb.ZSetValue{
				Values: value.(*zset.SortedSet).GetValue(),
			},
		}
	case ds.List:
		e.Value = &pb.Entry_ListValue{
			ListValue: &pb.ListValue{
				Values: value.(*list.DoublyLinkedList).GetValue(),
			},
		}
	case ds.Hash:
		e.Value = &pb.Entry_HashValue{
			HashValue: &pb.HashValue{
				Values: value.(*hash.HashMap).GetValue(),
			},
		}
	case ds.Set:
		e.Value = &pb.Entry_SetValue{
			SetValue: &pb.SetValue{
				Values: value.(*set.Set).GetValue(),
			},
		}
	}
	return e
}
