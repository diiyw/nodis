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
func newEntry(key string, dataStruct ds.DataStruct, expiration int64) *pb.Entry {
	e := &pb.Entry{
		Key:        key,
		Expiration: expiration,
		Type:       uint32(dataStruct.Type()),
	}
	switch dataStruct.Type() {
	case ds.String:
		e.Value = &pb.Entry_StringValue{
			StringValue: &pb.StringValue{
				Value: dataStruct.(*str.String).GetValue(),
			},
		}
	case ds.ZSet:
		e.Value = &pb.Entry_ZSetValue{
			ZSetValue: &pb.ZSetValue{
				Values: dataStruct.(*zset.SortedSet).GetValue(),
			},
		}
	case ds.List:
		e.Value = &pb.Entry_ListValue{
			ListValue: &pb.ListValue{
				Values: dataStruct.(*list.DoublyLinkedList).GetValue(),
			},
		}
	case ds.Hash:
		e.Value = &pb.Entry_HashValue{
			HashValue: &pb.HashValue{
				Values: dataStruct.(*hash.HashMap).GetValue(),
			},
		}
	case ds.Set:
		e.Value = &pb.Entry_SetValue{
			SetValue: &pb.SetValue{
				Values: dataStruct.(*set.Set).GetValue(),
			},
		}
	}
	return e
}
