// Package patch implements the versioned NDS1 change stream. No code generation
// or external serialization runtime is required.
package patch

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
	"math"
)

const (
	OpTypeNone uint8 = iota
	OpTypeClear
	OpTypeDel
	OpTypeExpire
	OpTypeExpireAt
	OpTypeHClear
	OpTypeHDel
	OpTypeHIncrBy
	OpTypeHIncrByFloat
	OpTypeHMSet
	OpTypeHSet
	OpTypeLInsert
	OpTypeLPop
	OpTypeLPopRPush
	OpTypeLPush
	OpTypeLPushX
	OpTypeLRem
	OpTypeLSet
	OpTypeLTrim
	OpTypeRPop
	OpTypeRPopLPush
	OpTypeRPush
	OpTypeRPushX
	OpTypeSAdd
	OpTypeSRem
	OpTypeSet
	OpTypeZAdd
	OpTypeZClear
	OpTypeZIncrBy
	OpTypeZRem
	OpTypeZRemRangeByRank
	OpTypeZRemRangeByScore
	OpTypeRename
	OpTypePersist
	OpTypeZUnionStore
	OpTypeZInterStore
	OpTypeRenameNX
	OpTypeBatch
	OpTypeRestore
	OpTypeReset
)
const MaxFrameSize = 64 << 20
const headerSize = 17

var ErrInvalidFrame = errors.New("nodis: invalid stream frame")
var ErrSequence = errors.New("nodis: stream sequence gap or replay")
var checksumTable = crc32.MakeTable(crc32.Castagnoli)

type OpData interface{ GetKey() string }
type Op struct {
	Type     uint8
	Sequence uint64
	Data     OpData
}
type OpClear struct {
	Key string
}

func (o *OpClear) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpDel struct {
	Key string
}

func (o *OpDel) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpExpire struct {
	Key        string
	Expiration int64
}

func (o *OpExpire) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpExpireAt struct {
	Key        string
	Expiration int64
}

func (o *OpExpireAt) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpHClear struct {
	Key string
}

func (o *OpHClear) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpHDel struct {
	Key    string
	Fields []string
}

func (o *OpHDel) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpHIncrBy struct {
	Key     string
	Field   string
	IncrInt int64
}

func (o *OpHIncrBy) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpHIncrByFloat struct {
	Key       string
	Field     string
	IncrFloat float64
}

func (o *OpHIncrByFloat) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpHSet struct {
	Key   string
	Field string
	Value []byte
}

func (o *OpHSet) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpHMSet struct {
	Key    string
	Fields []string
	Values [][]byte
}

func (o *OpHMSet) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpLInsert struct {
	Key    string
	Pivot  []byte
	Value  []byte
	Before bool
}

func (o *OpLInsert) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpLPop struct {
	Key   string
	Count int64
}

func (o *OpLPop) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpLPopRPush struct {
	Key    string
	DstKey string
}

func (o *OpLPopRPush) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpLPush struct {
	Key    string
	Values [][]byte
}

func (o *OpLPush) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpLPushX struct {
	Key   string
	Value []byte
}

func (o *OpLPushX) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpLRem struct {
	Key   string
	Value []byte
	Count int64
}

func (o *OpLRem) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpLSet struct {
	Key   string
	Index int64
	Value []byte
}

func (o *OpLSet) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpLTrim struct {
	Key   string
	Start int64
	Stop  int64
}

func (o *OpLTrim) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpRPop struct {
	Key   string
	Count int64
}

func (o *OpRPop) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpRPopLPush struct {
	Key    string
	DstKey string
}

func (o *OpRPopLPush) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpRPush struct {
	Key    string
	Values [][]byte
}

func (o *OpRPush) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpRPushX struct {
	Key   string
	Value []byte
}

func (o *OpRPushX) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpSAdd struct {
	Key     string
	Members []string
}

func (o *OpSAdd) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpSRem struct {
	Key     string
	Members []string
}

func (o *OpSRem) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpSet struct {
	Key        string
	Value      []byte
	KeepTTL    bool
	Expiration int64
}

func (o *OpSet) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpZAdd struct {
	Key    string
	Member string
	Score  float64
}

func (o *OpZAdd) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpZClear struct {
	Key string
}

func (o *OpZClear) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpZIncrBy struct {
	Key    string
	Member string
	Score  float64
}

func (o *OpZIncrBy) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpZRem struct {
	Key     string
	Member  string
	Members []string
}

func (o *OpZRem) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpZRemRangeByRank struct {
	Key   string
	Start int64
	Stop  int64
}

func (o *OpZRemRangeByRank) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpZRemRangeByScore struct {
	Key  string
	Mode int64
	Min  float64
	Max  float64
}

func (o *OpZRemRangeByScore) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpRename struct {
	Key    string
	DstKey string
}

func (o *OpRename) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpPersist struct {
	Key string
}

func (o *OpPersist) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpZUnionStore struct {
	Key       string
	Keys      []string
	Weights   []float64
	Aggregate string
}

func (o *OpZUnionStore) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpZInterStore struct {
	Key       string
	Keys      []string
	Weights   []float64
	Aggregate string
}

func (o *OpZInterStore) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpRenameNX struct {
	Key    string
	DstKey string
}

func (o *OpRenameNX) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpBatch struct {
	Key    string
	Frames [][]byte
}

func (o *OpBatch) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpRestore struct {
	Key        string
	Value      []byte
	Expiration int64
}

func (o *OpRestore) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

type OpReset struct {
	Key      string
	Patterns []string
}

func (o *OpReset) GetKey() string {
	if o == nil {
		return ""
	}
	return o.Key
}

func appendBytes(dst, b []byte) []byte {
	dst = binary.AppendUvarint(dst, uint64(len(b)))
	return append(dst, b...)
}
func appendString(dst []byte, s string) []byte {
	dst = binary.AppendUvarint(dst, uint64(len(s)))
	return append(dst, s...)
}
func (op Op) AppendEncode(dst []byte) ([]byte, error) {
	start := len(dst)
	dst = append(dst, "NDS1"...)
	dst = append(dst, make([]byte, 13)...)
	binary.BigEndian.PutUint64(dst[start+8:], op.Sequence)
	dst[start+16] = op.Type
	switch d := op.Data.(type) {
	case *OpClear:
		if d == nil || op.Type != OpTypeClear {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
	case *OpDel:
		if d == nil || op.Type != OpTypeDel {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
	case *OpExpire:
		if d == nil || op.Type != OpTypeExpire {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
		dst = binary.AppendVarint(dst, d.Expiration)
	case *OpExpireAt:
		if d == nil || op.Type != OpTypeExpireAt {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
		dst = binary.AppendVarint(dst, d.Expiration)
	case *OpHClear:
		if d == nil || op.Type != OpTypeHClear {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
	case *OpHDel:
		if d == nil || op.Type != OpTypeHDel {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
		dst = binary.AppendUvarint(dst, uint64(len(d.Fields)))
		for _, v := range d.Fields {
			dst = appendString(dst, v)
		}
	case *OpHIncrBy:
		if d == nil || op.Type != OpTypeHIncrBy {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
		dst = appendString(dst, d.Field)
		dst = binary.AppendVarint(dst, d.IncrInt)
	case *OpHIncrByFloat:
		if d == nil || op.Type != OpTypeHIncrByFloat {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
		dst = appendString(dst, d.Field)
		dst = binary.LittleEndian.AppendUint64(dst, math.Float64bits(d.IncrFloat))
	case *OpHSet:
		if d == nil || op.Type != OpTypeHSet {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
		dst = appendString(dst, d.Field)
		dst = appendBytes(dst, d.Value)
	case *OpHMSet:
		if d == nil || op.Type != OpTypeHMSet {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
		dst = binary.AppendUvarint(dst, uint64(len(d.Fields)))
		for _, v := range d.Fields {
			dst = appendString(dst, v)
		}
		dst = binary.AppendUvarint(dst, uint64(len(d.Values)))
		for _, v := range d.Values {
			dst = appendBytes(dst, v)
		}
	case *OpLInsert:
		if d == nil || op.Type != OpTypeLInsert {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
		dst = appendBytes(dst, d.Pivot)
		dst = appendBytes(dst, d.Value)
		if d.Before {
			dst = append(dst, 1)
		} else {
			dst = append(dst, 0)
		}
	case *OpLPop:
		if d == nil || op.Type != OpTypeLPop {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
		dst = binary.AppendVarint(dst, d.Count)
	case *OpLPopRPush:
		if d == nil || op.Type != OpTypeLPopRPush {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
		dst = appendString(dst, d.DstKey)
	case *OpLPush:
		if d == nil || op.Type != OpTypeLPush {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
		dst = binary.AppendUvarint(dst, uint64(len(d.Values)))
		for _, v := range d.Values {
			dst = appendBytes(dst, v)
		}
	case *OpLPushX:
		if d == nil || op.Type != OpTypeLPushX {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
		dst = appendBytes(dst, d.Value)
	case *OpLRem:
		if d == nil || op.Type != OpTypeLRem {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
		dst = appendBytes(dst, d.Value)
		dst = binary.AppendVarint(dst, d.Count)
	case *OpLSet:
		if d == nil || op.Type != OpTypeLSet {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
		dst = binary.AppendVarint(dst, d.Index)
		dst = appendBytes(dst, d.Value)
	case *OpLTrim:
		if d == nil || op.Type != OpTypeLTrim {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
		dst = binary.AppendVarint(dst, d.Start)
		dst = binary.AppendVarint(dst, d.Stop)
	case *OpRPop:
		if d == nil || op.Type != OpTypeRPop {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
		dst = binary.AppendVarint(dst, d.Count)
	case *OpRPopLPush:
		if d == nil || op.Type != OpTypeRPopLPush {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
		dst = appendString(dst, d.DstKey)
	case *OpRPush:
		if d == nil || op.Type != OpTypeRPush {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
		dst = binary.AppendUvarint(dst, uint64(len(d.Values)))
		for _, v := range d.Values {
			dst = appendBytes(dst, v)
		}
	case *OpRPushX:
		if d == nil || op.Type != OpTypeRPushX {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
		dst = appendBytes(dst, d.Value)
	case *OpSAdd:
		if d == nil || op.Type != OpTypeSAdd {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
		dst = binary.AppendUvarint(dst, uint64(len(d.Members)))
		for _, v := range d.Members {
			dst = appendString(dst, v)
		}
	case *OpSRem:
		if d == nil || op.Type != OpTypeSRem {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
		dst = binary.AppendUvarint(dst, uint64(len(d.Members)))
		for _, v := range d.Members {
			dst = appendString(dst, v)
		}
	case *OpSet:
		if d == nil || op.Type != OpTypeSet {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
		dst = appendBytes(dst, d.Value)
		if d.KeepTTL {
			dst = append(dst, 1)
		} else {
			dst = append(dst, 0)
		}
		dst = binary.AppendVarint(dst, d.Expiration)
	case *OpZAdd:
		if d == nil || op.Type != OpTypeZAdd {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
		dst = appendString(dst, d.Member)
		dst = binary.LittleEndian.AppendUint64(dst, math.Float64bits(d.Score))
	case *OpZClear:
		if d == nil || op.Type != OpTypeZClear {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
	case *OpZIncrBy:
		if d == nil || op.Type != OpTypeZIncrBy {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
		dst = appendString(dst, d.Member)
		dst = binary.LittleEndian.AppendUint64(dst, math.Float64bits(d.Score))
	case *OpZRem:
		if d == nil || op.Type != OpTypeZRem {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
		dst = appendString(dst, d.Member)
		dst = binary.AppendUvarint(dst, uint64(len(d.Members)))
		for _, v := range d.Members {
			dst = appendString(dst, v)
		}
	case *OpZRemRangeByRank:
		if d == nil || op.Type != OpTypeZRemRangeByRank {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
		dst = binary.AppendVarint(dst, d.Start)
		dst = binary.AppendVarint(dst, d.Stop)
	case *OpZRemRangeByScore:
		if d == nil || op.Type != OpTypeZRemRangeByScore {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
		dst = binary.AppendVarint(dst, d.Mode)
		dst = binary.LittleEndian.AppendUint64(dst, math.Float64bits(d.Min))
		dst = binary.LittleEndian.AppendUint64(dst, math.Float64bits(d.Max))
	case *OpRename:
		if d == nil || op.Type != OpTypeRename {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
		dst = appendString(dst, d.DstKey)
	case *OpPersist:
		if d == nil || op.Type != OpTypePersist {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
	case *OpZUnionStore:
		if d == nil || op.Type != OpTypeZUnionStore {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
		dst = binary.AppendUvarint(dst, uint64(len(d.Keys)))
		for _, v := range d.Keys {
			dst = appendString(dst, v)
		}
		dst = binary.AppendUvarint(dst, uint64(len(d.Weights)))
		for _, v := range d.Weights {
			dst = binary.LittleEndian.AppendUint64(dst, math.Float64bits(v))
		}
		dst = appendString(dst, d.Aggregate)
	case *OpZInterStore:
		if d == nil || op.Type != OpTypeZInterStore {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
		dst = binary.AppendUvarint(dst, uint64(len(d.Keys)))
		for _, v := range d.Keys {
			dst = appendString(dst, v)
		}
		dst = binary.AppendUvarint(dst, uint64(len(d.Weights)))
		for _, v := range d.Weights {
			dst = binary.LittleEndian.AppendUint64(dst, math.Float64bits(v))
		}
		dst = appendString(dst, d.Aggregate)
	case *OpRenameNX:
		if d == nil || op.Type != OpTypeRenameNX {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
		dst = appendString(dst, d.DstKey)
	case *OpBatch:
		if d == nil || op.Type != OpTypeBatch {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
		dst = binary.AppendUvarint(dst, uint64(len(d.Frames)))
		for _, v := range d.Frames {
			dst = appendBytes(dst, v)
		}
	case *OpRestore:
		if d == nil || op.Type != OpTypeRestore {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
		dst = appendBytes(dst, d.Value)
		dst = binary.AppendVarint(dst, d.Expiration)
	case *OpReset:
		if d == nil || op.Type != OpTypeReset {
			return dst[:start], ErrInvalidFrame
		}
		dst = appendString(dst, d.Key)
		dst = binary.AppendUvarint(dst, uint64(len(d.Patterns)))
		for _, v := range d.Patterns {
			dst = appendString(dst, v)
		}
	default:
		return dst[:start], ErrInvalidFrame
	}
	size := len(dst) - start + 4
	if size > MaxFrameSize {
		return dst[:start], ErrInvalidFrame
	}
	binary.BigEndian.PutUint32(dst[start+4:], uint32(size))
	dst = binary.BigEndian.AppendUint32(dst, crc32.Checksum(dst[start:], checksumTable))
	return dst, nil
}
func (op Op) Encode() []byte {
	b, err := op.AppendEncode(nil)
	if err != nil {
		panic(err)
	}
	return b
}

type payloadReader struct {
	p   []byte
	err error
}

func (r *payloadReader) unsigned() uint64 {
	if r.err != nil {
		return 0
	}
	v, n := binary.Uvarint(r.p)
	if n <= 0 {
		r.err = ErrInvalidFrame
		return 0
	}
	r.p = r.p[n:]
	return v
}
func (r *payloadReader) integer() int64 {
	if r.err != nil {
		return 0
	}
	v, n := binary.Varint(r.p)
	if n <= 0 {
		r.err = ErrInvalidFrame
		return 0
	}
	r.p = r.p[n:]
	return v
}
func (r *payloadReader) bytes() []byte {
	n := r.unsigned()
	if n > uint64(len(r.p)) {
		r.err = ErrInvalidFrame
		return nil
	}
	v := append([]byte{}, r.p[:int(n)]...)
	r.p = r.p[int(n):]
	return v
}
func (r *payloadReader) str() string {
	n := r.unsigned()
	if n > uint64(len(r.p)) {
		r.err = ErrInvalidFrame
		return ""
	}
	v := string(r.p[:int(n)])
	r.p = r.p[int(n):]
	return v
}
func (r *payloadReader) float() float64 {
	if len(r.p) < 8 {
		r.err = ErrInvalidFrame
		return 0
	}
	v := math.Float64frombits(binary.LittleEndian.Uint64(r.p))
	r.p = r.p[8:]
	return v
}
func (r *payloadReader) boolean() bool {
	if len(r.p) < 1 || r.p[0] > 1 {
		r.err = ErrInvalidFrame
		return false
	}
	v := r.p[0] != 0
	r.p = r.p[1:]
	return v
}
func (r *payloadReader) count(minSize uint64) int {
	n := r.unsigned()
	if n > uint64(len(r.p))/minSize {
		r.err = ErrInvalidFrame
		return 0
	}
	return int(n)
}
func DecodeOp(b []byte) (Op, error) {
	if len(b) < headerSize+4 || len(b) > MaxFrameSize || string(b[:4]) != "NDS1" || int(binary.BigEndian.Uint32(b[4:])) != len(b) {
		return Op{}, ErrInvalidFrame
	}
	if crc32.Checksum(b[:len(b)-4], checksumTable) != binary.BigEndian.Uint32(b[len(b)-4:]) {
		return Op{}, ErrInvalidFrame
	}
	op := Op{Type: b[16], Sequence: binary.BigEndian.Uint64(b[8:])}
	r := payloadReader{p: b[17 : len(b)-4]}
	switch op.Type {
	case OpTypeClear:
		d := &OpClear{}
		d.Key = r.str()
		op.Data = d
	case OpTypeDel:
		d := &OpDel{}
		d.Key = r.str()
		op.Data = d
	case OpTypeExpire:
		d := &OpExpire{}
		d.Key = r.str()
		d.Expiration = r.integer()
		op.Data = d
	case OpTypeExpireAt:
		d := &OpExpireAt{}
		d.Key = r.str()
		d.Expiration = r.integer()
		op.Data = d
	case OpTypeHClear:
		d := &OpHClear{}
		d.Key = r.str()
		op.Data = d
	case OpTypeHDel:
		d := &OpHDel{}
		d.Key = r.str()
		d.Fields = make([]string, r.count(1))
		for i := range d.Fields {
			d.Fields[i] = r.str()
		}
		op.Data = d
	case OpTypeHIncrBy:
		d := &OpHIncrBy{}
		d.Key = r.str()
		d.Field = r.str()
		d.IncrInt = r.integer()
		op.Data = d
	case OpTypeHIncrByFloat:
		d := &OpHIncrByFloat{}
		d.Key = r.str()
		d.Field = r.str()
		d.IncrFloat = r.float()
		op.Data = d
	case OpTypeHSet:
		d := &OpHSet{}
		d.Key = r.str()
		d.Field = r.str()
		d.Value = r.bytes()
		op.Data = d
	case OpTypeHMSet:
		d := &OpHMSet{}
		d.Key = r.str()
		d.Fields = make([]string, r.count(1))
		for i := range d.Fields {
			d.Fields[i] = r.str()
		}
		d.Values = make([][]byte, r.count(1))
		for i := range d.Values {
			d.Values[i] = r.bytes()
		}
		op.Data = d
	case OpTypeLInsert:
		d := &OpLInsert{}
		d.Key = r.str()
		d.Pivot = r.bytes()
		d.Value = r.bytes()
		d.Before = r.boolean()
		op.Data = d
	case OpTypeLPop:
		d := &OpLPop{}
		d.Key = r.str()
		d.Count = r.integer()
		op.Data = d
	case OpTypeLPopRPush:
		d := &OpLPopRPush{}
		d.Key = r.str()
		d.DstKey = r.str()
		op.Data = d
	case OpTypeLPush:
		d := &OpLPush{}
		d.Key = r.str()
		d.Values = make([][]byte, r.count(1))
		for i := range d.Values {
			d.Values[i] = r.bytes()
		}
		op.Data = d
	case OpTypeLPushX:
		d := &OpLPushX{}
		d.Key = r.str()
		d.Value = r.bytes()
		op.Data = d
	case OpTypeLRem:
		d := &OpLRem{}
		d.Key = r.str()
		d.Value = r.bytes()
		d.Count = r.integer()
		op.Data = d
	case OpTypeLSet:
		d := &OpLSet{}
		d.Key = r.str()
		d.Index = r.integer()
		d.Value = r.bytes()
		op.Data = d
	case OpTypeLTrim:
		d := &OpLTrim{}
		d.Key = r.str()
		d.Start = r.integer()
		d.Stop = r.integer()
		op.Data = d
	case OpTypeRPop:
		d := &OpRPop{}
		d.Key = r.str()
		d.Count = r.integer()
		op.Data = d
	case OpTypeRPopLPush:
		d := &OpRPopLPush{}
		d.Key = r.str()
		d.DstKey = r.str()
		op.Data = d
	case OpTypeRPush:
		d := &OpRPush{}
		d.Key = r.str()
		d.Values = make([][]byte, r.count(1))
		for i := range d.Values {
			d.Values[i] = r.bytes()
		}
		op.Data = d
	case OpTypeRPushX:
		d := &OpRPushX{}
		d.Key = r.str()
		d.Value = r.bytes()
		op.Data = d
	case OpTypeSAdd:
		d := &OpSAdd{}
		d.Key = r.str()
		d.Members = make([]string, r.count(1))
		for i := range d.Members {
			d.Members[i] = r.str()
		}
		op.Data = d
	case OpTypeSRem:
		d := &OpSRem{}
		d.Key = r.str()
		d.Members = make([]string, r.count(1))
		for i := range d.Members {
			d.Members[i] = r.str()
		}
		op.Data = d
	case OpTypeSet:
		d := &OpSet{}
		d.Key = r.str()
		d.Value = r.bytes()
		d.KeepTTL = r.boolean()
		d.Expiration = r.integer()
		op.Data = d
	case OpTypeZAdd:
		d := &OpZAdd{}
		d.Key = r.str()
		d.Member = r.str()
		d.Score = r.float()
		op.Data = d
	case OpTypeZClear:
		d := &OpZClear{}
		d.Key = r.str()
		op.Data = d
	case OpTypeZIncrBy:
		d := &OpZIncrBy{}
		d.Key = r.str()
		d.Member = r.str()
		d.Score = r.float()
		op.Data = d
	case OpTypeZRem:
		d := &OpZRem{}
		d.Key = r.str()
		d.Member = r.str()
		d.Members = make([]string, r.count(1))
		for i := range d.Members {
			d.Members[i] = r.str()
		}
		op.Data = d
	case OpTypeZRemRangeByRank:
		d := &OpZRemRangeByRank{}
		d.Key = r.str()
		d.Start = r.integer()
		d.Stop = r.integer()
		op.Data = d
	case OpTypeZRemRangeByScore:
		d := &OpZRemRangeByScore{}
		d.Key = r.str()
		d.Mode = r.integer()
		d.Min = r.float()
		d.Max = r.float()
		op.Data = d
	case OpTypeRename:
		d := &OpRename{}
		d.Key = r.str()
		d.DstKey = r.str()
		op.Data = d
	case OpTypePersist:
		d := &OpPersist{}
		d.Key = r.str()
		op.Data = d
	case OpTypeZUnionStore:
		d := &OpZUnionStore{}
		d.Key = r.str()
		d.Keys = make([]string, r.count(1))
		for i := range d.Keys {
			d.Keys[i] = r.str()
		}
		d.Weights = make([]float64, r.count(8))
		for i := range d.Weights {
			d.Weights[i] = r.float()
		}
		d.Aggregate = r.str()
		op.Data = d
	case OpTypeZInterStore:
		d := &OpZInterStore{}
		d.Key = r.str()
		d.Keys = make([]string, r.count(1))
		for i := range d.Keys {
			d.Keys[i] = r.str()
		}
		d.Weights = make([]float64, r.count(8))
		for i := range d.Weights {
			d.Weights[i] = r.float()
		}
		d.Aggregate = r.str()
		op.Data = d
	case OpTypeRenameNX:
		d := &OpRenameNX{}
		d.Key = r.str()
		d.DstKey = r.str()
		op.Data = d
	case OpTypeBatch:
		d := &OpBatch{}
		d.Key = r.str()
		d.Frames = make([][]byte, r.count(1))
		for i := range d.Frames {
			d.Frames[i] = r.bytes()
		}
		op.Data = d
	case OpTypeRestore:
		d := &OpRestore{}
		d.Key = r.str()
		d.Value = r.bytes()
		d.Expiration = r.integer()
		op.Data = d
	case OpTypeReset:
		d := &OpReset{}
		d.Key = r.str()
		d.Patterns = make([]string, r.count(1))
		for i := range d.Patterns {
			d.Patterns[i] = r.str()
		}
		op.Data = d
	default:
		return Op{}, ErrInvalidFrame
	}
	if r.err != nil || len(r.p) != 0 {
		return Op{}, ErrInvalidFrame
	}
	return op, nil
}

// Encoder reuses its buffer; Write performs no allocations after capacity warms.
type Encoder struct {
	w        io.Writer
	buf      []byte
	sequence uint64
}

func NewEncoder(w io.Writer) *Encoder { return &Encoder{w: w} }
func (e *Encoder) Write(op Op) error {
	e.sequence++
	op.Sequence = e.sequence
	b, err := op.AppendEncode(e.buf[:0])
	if err != nil {
		return err
	}
	e.buf = b
	for len(b) > 0 {
		n, err := e.w.Write(b)
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
		b = b[n:]
	}
	return nil
}

type Decoder struct {
	r        io.Reader
	buf      []byte
	sequence uint64
}

func NewDecoder(r io.Reader) *Decoder { return &Decoder{r: r, buf: make([]byte, 8)} }
func (d *Decoder) Read() (Op, error) {
	if _, err := io.ReadFull(d.r, d.buf[:8]); err != nil {
		return Op{}, err
	}
	size := int(binary.BigEndian.Uint32(d.buf[4:8]))
	if string(d.buf[:4]) != "NDS1" || size < headerSize+4 || size > MaxFrameSize {
		return Op{}, ErrInvalidFrame
	}
	if cap(d.buf) < size {
		b := make([]byte, size)
		copy(b, d.buf[:8])
		d.buf = b
	} else {
		d.buf = d.buf[:size]
	}
	if _, err := io.ReadFull(d.r, d.buf[8:]); err != nil {
		return Op{}, err
	}
	op, err := DecodeOp(d.buf)
	if err != nil {
		return Op{}, err
	}
	if op.Sequence != d.sequence+1 {
		return Op{}, ErrSequence
	}
	d.sequence = op.Sequence
	return op, nil
}
