package patch

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"reflect"
	"testing"
)

func allOperations() []OpData {
	return []OpData{&OpClear{}, &OpDel{}, &OpExpire{}, &OpExpireAt{}, &OpHClear{}, &OpHDel{}, &OpHIncrBy{}, &OpHIncrByFloat{}, &OpHMSet{}, &OpHSet{}, &OpLInsert{}, &OpLPop{}, &OpLPopRPush{}, &OpLPush{}, &OpLPushX{}, &OpLRem{}, &OpLSet{}, &OpLTrim{}, &OpRPop{}, &OpRPopLPush{}, &OpRPush{}, &OpRPushX{}, &OpSAdd{}, &OpSRem{}, &OpSet{}, &OpZAdd{}, &OpZClear{}, &OpZIncrBy{}, &OpZRem{}, &OpZRemRangeByRank{}, &OpZRemRangeByScore{}, &OpRename{}, &OpPersist{}, &OpZUnionStore{}, &OpZInterStore{}, &OpRenameNX{}, &OpBatch{}, &OpRestore{}, &OpReset{}}
}
func fill(v reflect.Value) {
	switch v.Kind() {
	case reflect.String:
		v.SetString("中文\x00wire")
	case reflect.Int64:
		v.SetInt(-123456789)
	case reflect.Float64:
		v.SetFloat(-1.25)
	case reflect.Bool:
		v.SetBool(true)
	case reflect.Slice:
		v.Set(reflect.MakeSlice(v.Type(), 2, 2))
		for i := 0; i < v.Len(); i++ {
			if v.Type().Elem().Kind() == reflect.Uint8 {
				v.Index(i).SetUint(255)
			} else {
				fill(v.Index(i))
			}
		}
	}
}
func TestEveryOperationRoundTrip(t *testing.T) {
	for i, data := range allOperations() {
		t.Run(reflect.TypeOf(data).String(), func(t *testing.T) {
			v := reflect.ValueOf(data).Elem()
			for j := 0; j < v.NumField(); j++ {
				fill(v.Field(j))
			}
			op := Op{Type: uint8(i + 1), Sequence: 42, Data: data}
			frame, err := op.AppendEncode(nil)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := DecodeOp(frame)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(op, decoded) {
				t.Fatalf("round trip differs: %#v %#v", op, decoded)
			}
			for j := 0; j < len(frame); j++ {
				if _, err := DecodeOp(frame[:j]); err == nil {
					t.Fatal("truncated frame accepted", j)
				}
			}
		})
	}
}
func TestDecoderBoundsAndChecksum(t *testing.T) {
	valid := Op{Type: OpTypeSet, Data: &OpSet{Key: "k", Value: []byte("v")}}.Encode()
	for _, index := range []int{0, 4, 8, 16, len(valid) - 1} {
		bad := bytes.Clone(valid)
		bad[index] ^= 255
		if _, err := DecodeOp(bad); err == nil {
			t.Fatal("corrupt frame accepted")
		}
	}
	invalid := make([]byte, 8)
	copy(invalid, "NDS1")
	binary.BigEndian.PutUint32(invalid[4:], MaxFrameSize+1)
	if _, err := NewDecoder(bytes.NewReader(invalid)).Read(); !errors.Is(err, ErrInvalidFrame) {
		t.Fatal(err)
	}
}

type fragmentReader struct{ io.Reader }

func (r fragmentReader) Read(b []byte) (int, error) {
	if len(b) > 1 {
		b = b[:1]
	}
	return r.Reader.Read(b)
}
func TestStreamFramingAndSequences(t *testing.T) {
	var stream bytes.Buffer
	e := NewEncoder(&stream)
	for i := 0; i < 3; i++ {
		if err := e.Write(Op{Type: OpTypeDel, Data: &OpDel{Key: "k"}}); err != nil {
			t.Fatal(err)
		}
	}
	d := NewDecoder(fragmentReader{&stream})
	for i := 1; i <= 3; i++ {
		op, err := d.Read()
		if err != nil || op.Sequence != uint64(i) {
			t.Fatal(op, err)
		}
	}
	frame := Op{Type: OpTypeDel, Sequence: 2, Data: &OpDel{Key: "k"}}.Encode()
	if _, err := NewDecoder(bytes.NewReader(frame)).Read(); !errors.Is(err, ErrSequence) {
		t.Fatal(err)
	}
}
func TestAppendEncodeZeroAllocs(t *testing.T) {
	op := Op{Type: OpTypeSet, Data: &OpSet{Key: "key", Value: []byte("value")}}
	buf := make([]byte, 0, 256)
	if got := testing.AllocsPerRun(1000, func() {
		if _, err := op.AppendEncode(buf[:0]); err != nil {
			panic(err)
		}
	}); got != 0 {
		t.Fatal("allocations:", got)
	}
}
func FuzzDecodeOp(f *testing.F) {
	f.Add(Op{Type: OpTypeDel, Data: &OpDel{Key: "k"}}.Encode())
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, b []byte) { _, _ = DecodeOp(b) })
}
