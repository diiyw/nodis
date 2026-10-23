package protocol

import (
	"encoding/binary"
	"math"
	"testing"
	"time"

	"github.com/diiyw/nodis"
)

func queryCodecFixture(t testing.TB) (nodis.Query, nodis.QueryPage) {
	t.Helper()
	query := nodis.Query{Document: 7, After: 2, Limit: 10, Select: []string{"name"}}
	for _, input := range []any{nil, true, int64(-5), uint64(math.MaxUint64), 1.25, "Ada", []byte{0, 255}, time.Unix(1700000000, 123)} {
		value, err := nodis.ValueOf(input)
		if err != nil {
			t.Fatal(err)
		}
		query.Where = append(query.Where, nodis.Condition{Field: "field", Op: nodis.Equal, Value: value})
	}
	data := make([]byte, 30)
	copy(data, "NDX1")
	binary.LittleEndian.PutUint16(data[4:6], 1)
	binary.LittleEndian.PutUint64(data[6:14], 3)
	binary.LittleEndian.PutUint64(data[14:22], 2)
	data = appendQueryName(data, "name")
	value, _ := nodis.ValueOf("Ada")
	data = nodis.AppendValue(data, value)
	object, err := nodis.ParsePortable(data)
	if err != nil {
		t.Fatal(err)
	}
	return query, nodis.QueryPage{Objects: []nodis.PortableObject{object}, Sequence: 9, PermissionRevision: 4, Next: 3, HasMore: true}
}

func TestQueryCodecsRoundTripAndZeroAllocations(t *testing.T) {
	query, page := queryCodecFixture(t)
	request, err := AppendQuery(nil, query)
	if err != nil {
		t.Fatal(err)
	}
	var decoder QueryDecoder
	decoded, err := DecodeQueryInto(request, &decoder)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Document != query.Document || decoded.After != query.After || decoded.Limit != query.Limit || len(decoded.Where) != len(query.Where) {
		t.Fatal("query metadata changed")
	}
	for i := range query.Where {
		want, got := nodis.AppendValue(nil, query.Where[i].Value), nodis.AppendValue(nil, decoded.Where[i].Value)
		if string(want) != string(got) {
			t.Fatal("query value changed", i)
		}
	}
	owned := decoded.Clone()
	clear(request)
	if value, _ := owned.Where[5].Value.String(); value != "Ada" || owned.Where[5].Field != "field" || owned.Select[0] != "name" {
		t.Fatal("cloned query retained borrowed strings")
	}
	request, err = AppendQuery(request[:0], query)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := AppendQueryPage(nil, page)
	if err != nil {
		t.Fatal(err)
	}
	var pageDecoder QueryPageDecoder
	result, err := DecodeQueryPageInto(encoded, &pageDecoder)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Objects) != 1 || result.Next != 3 || !result.HasMore || result.Sequence != 9 || result.PermissionRevision != 4 {
		t.Fatal("page metadata changed")
	}
	if name, _ := result.Objects[0].String("name"); name != "Ada" {
		t.Fatal("page object changed")
	}
	requestBuffer := make([]byte, 0, len(request))
	responseBuffer := make([]byte, 0, len(encoded))
	if allocs := testing.AllocsPerRun(1000, func() {
		if _, err := AppendQuery(requestBuffer[:0], query); err != nil {
			panic(err)
		}
		if _, err := DecodeQueryInto(request, &decoder); err != nil {
			panic(err)
		}
		if _, err := AppendQueryPage(responseBuffer[:0], page); err != nil {
			panic(err)
		}
		if _, err := DecodeQueryPageInto(encoded, &pageDecoder); err != nil {
			panic(err)
		}
	}); allocs != 0 {
		t.Fatalf("query/page codecs: %g allocs", allocs)
	}
}

func TestQueryDecoderRejectsMalformedInput(t *testing.T) {
	query, page := queryCodecFixture(t)
	request, _ := AppendQuery(nil, query)
	response, _ := AppendQueryPage(nil, page)
	for size := 0; size < len(request); size++ {
		var decoder QueryDecoder
		if _, err := DecodeQueryInto(request[:size], &decoder); err == nil {
			t.Fatal("accepted truncated query", size)
		}
	}
	for size := 0; size < len(response); size++ {
		var decoder QueryPageDecoder
		if _, err := DecodeQueryPageInto(response[:size], &decoder); err == nil {
			t.Fatal("accepted truncated page", size)
		}
	}
	for _, damage := range []string{"magic", "document", "limit", "conditions", "select", "operator", "trailing"} {
		t.Run(damage, func(t *testing.T) {
			data := append([]byte(nil), request...)
			switch damage {
			case "magic":
				data[0] ^= 1
			case "document":
				clear(data[4:12])
			case "limit":
				binary.LittleEndian.PutUint32(data[20:24], math.MaxUint32)
			case "conditions":
				binary.LittleEndian.PutUint16(data[24:26], math.MaxUint16)
			case "select":
				binary.LittleEndian.PutUint16(data[26:28], math.MaxUint16)
			case "operator":
				data[queryHeaderSize+1+len("field")] = 255
			case "trailing":
				data = append(data, 0)
			}
			var decoder QueryDecoder
			if _, err := DecodeQueryInto(data, &decoder); err == nil {
				t.Fatal("accepted malformed query")
			}
		})
	}
	for _, damage := range []string{"flag", "cursor", "count", "length", "object-id", "revision", "trailing"} {
		t.Run("page-"+damage, func(t *testing.T) {
			data := append([]byte(nil), response...)
			switch damage {
			case "flag":
				data[28] = 2
			case "cursor":
				binary.LittleEndian.PutUint64(data[20:28], 99)
			case "count":
				binary.LittleEndian.PutUint16(data[29:31], math.MaxUint16)
			case "length":
				binary.LittleEndian.PutUint32(data[31:35], math.MaxUint32)
			case "object-id":
				clear(data[35+6 : 35+14])
			case "revision":
				clear(data[35+14 : 35+22])
			case "trailing":
				data = append(data, 0)
			}
			var decoder QueryPageDecoder
			if _, err := DecodeQueryPageInto(data, &decoder); err == nil {
				t.Fatal("accepted malformed page")
			}
		})
	}
	// Containers with attacker-controlled element counts are never decoded.
	for _, body := range [][]byte{{byte(nodis.KindArray), 10, 255, 255, 255, 255, 255, 255, 255, 255, 255, 1}, {byte(nodis.KindBool), 1, 2}, {byte(nodis.KindNull), 1, 0}, {byte(nodis.KindInt), 7, 0, 0, 0, 0, 0, 0, 0}} {
		if _, _, ok := readQueryScalar(body); ok {
			t.Fatal("malformed scalar accepted")
		}
	}
}

func FuzzQueryDecoders(f *testing.F) {
	query, page := queryCodecFixture(f)
	request, _ := AppendQuery(nil, query)
	response, _ := AppendQueryPage(nil, page)
	f.Add(request)
	f.Add(response)
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, data []byte) {
		var decoder QueryDecoder
		if query, err := DecodeQueryInto(data, &decoder); err == nil {
			if err := query.Validate(); err != nil {
				t.Fatal("decoded invalid query")
			}
			if _, err := AppendQuery(nil, query); err != nil {
				t.Fatal(err)
			}
		}
		var pageDecoder QueryPageDecoder
		if page, err := DecodeQueryPageInto(data, &pageDecoder); err == nil {
			if _, err := AppendQueryPage(nil, page); err != nil {
				t.Fatal("decoded non-encodable page", err)
			}
		}
	})
}
