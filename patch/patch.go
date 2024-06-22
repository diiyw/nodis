package patch

import (
	"errors"
	"google.golang.org/protobuf/proto"
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
)

type OpData interface {
	proto.Message
	GetKey() string
}

type Op struct {
	Type uint8
	Data OpData
}

func (o *Op) Encode() []byte {
	data, _ := proto.Marshal(o.Data)
	return append([]byte{o.Type}, data...)
}

func DecodeOp(data []byte) (op Op, err error) {
	op.Type = data[0]
	switch op.Type {
	case OpTypeClear:
		op.Data = &OpClear{}
	case OpTypeDel:
		op.Data = &OpDel{}
	case OpTypeExpire:
		op.Data = &OpExpire{}
	case OpTypeExpireAt:
		op.Data = &OpExpireAt{}
	case OpTypeHClear:
		op.Data = &OpHClear{}
	case OpTypeHDel:
		op.Data = &OpHDel{}
	case OpTypeHIncrBy:
		op.Data = &OpHIncrBy{}
	case OpTypeHIncrByFloat:
		op.Data = &OpHIncrByFloat{}
	case OpTypeHMSet:
		op.Data = &OpHMSet{}
	case OpTypeHSet:
		op.Data = &OpHSet{}
	case OpTypeLInsert:
		op.Data = &OpLInsert{}
	case OpTypeLPop:
		op.Data = &OpLPop{}
	case OpTypeLPopRPush:
		op.Data = &OpLPopRPush{}
	case OpTypeLPush:
		op.Data = &OpLPush{}
	case OpTypeLPushX:
		op.Data = &OpLPushX{}
	case OpTypeLRem:
		op.Data = &OpLRem{}
	case OpTypeLSet:
		op.Data = &OpLSet{}
	case OpTypeLTrim:
		op.Data = &OpLTrim{}
	case OpTypeRPop:
		op.Data = &OpRPop{}
	case OpTypeRPopLPush:
		op.Data = &OpRPopLPush{}
	case OpTypeRPush:
		op.Data = &OpRPush{}
	case OpTypeRPushX:
		op.Data = &OpRPushX{}
	case OpTypeSAdd:
		op.Data = &OpSAdd{}
	case OpTypeSRem:
		op.Data = &OpSRem{}
	case OpTypeSet:
		op.Data = &OpSet{}
	case OpTypeZAdd:
		op.Data = &OpZAdd{}
	case OpTypeZClear:
		op.Data = &OpZClear{}
	case OpTypeZIncrBy:
		op.Data = &OpZIncrBy{}
	case OpTypeZRem:
		op.Data = &OpZRem{}
	case OpTypeZRemRangeByRank:
		op.Data = &OpZRemRangeByRank{}
	case OpTypeZRemRangeByScore:
		op.Data = &OpZRemRangeByScore{}
	case OpTypeRename:
		op.Data = &OpRename{}
	case OpTypePersist:
		op.Data = &OpPersist{}
	case OpTypeZUnionStore:
		op.Data = &OpZUnionStore{}
	case OpTypeZInterStore:
		op.Data = &OpZInterStore{}
	case OpTypeRenameNX:
		op.Data = &OpRenameNX{}
	default:
		err = errors.New("unknown operation type")
	}
	err = proto.Unmarshal(data[1:], op.Data)
	return
}
