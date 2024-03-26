// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.9.1
// source: nodis.proto

package pb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type OpType int32

const (
	OpType_None             OpType = 0
	OpType_Clear            OpType = 1
	OpType_Del              OpType = 2
	OpType_Expire           OpType = 3
	OpType_ExpireAt         OpType = 4
	OpType_HClear           OpType = 5
	OpType_HDel             OpType = 6
	OpType_HIncrBy          OpType = 7
	OpType_HIncrByFloat     OpType = 8
	OpType_HMSet            OpType = 9
	OpType_HSet             OpType = 10
	OpType_HSetNX           OpType = 11
	OpType_LInsert          OpType = 12
	OpType_LPop             OpType = 13
	OpType_LPopRPush        OpType = 14
	OpType_LPush            OpType = 15
	OpType_LPushX           OpType = 16
	OpType_LRem             OpType = 17
	OpType_LSet             OpType = 18
	OpType_LTrim            OpType = 19
	OpType_RPop             OpType = 20
	OpType_RPopLPush        OpType = 21
	OpType_RPush            OpType = 22
	OpType_RPushX           OpType = 23
	OpType_SAdd             OpType = 24
	OpType_SRem             OpType = 25
	OpType_Set              OpType = 26
	OpType_ZAdd             OpType = 27
	OpType_ZClear           OpType = 28
	OpType_ZIncrBy          OpType = 29
	OpType_ZRem             OpType = 30
	OpType_ZRemRangeByRank  OpType = 31
	OpType_ZRemRangeByScore OpType = 32
	OpType_Rename           OpType = 33
)

// Enum value maps for OpType.
var (
	OpType_name = map[int32]string{
		0:  "None",
		1:  "Clear",
		2:  "Del",
		3:  "Expire",
		4:  "ExpireAt",
		5:  "HClear",
		6:  "HDel",
		7:  "HIncrBy",
		8:  "HIncrByFloat",
		9:  "HMSet",
		10: "HSet",
		11: "HSetNX",
		12: "LInsert",
		13: "LPop",
		14: "LPopRPush",
		15: "LPush",
		16: "LPushX",
		17: "LRem",
		18: "LSet",
		19: "LTrim",
		20: "RPop",
		21: "RPopLPush",
		22: "RPush",
		23: "RPushX",
		24: "SAdd",
		25: "SRem",
		26: "Set",
		27: "ZAdd",
		28: "ZClear",
		29: "ZIncrBy",
		30: "ZRem",
		31: "ZRemRangeByRank",
		32: "ZRemRangeByScore",
		33: "Rename",
	}
	OpType_value = map[string]int32{
		"None":             0,
		"Clear":            1,
		"Del":              2,
		"Expire":           3,
		"ExpireAt":         4,
		"HClear":           5,
		"HDel":             6,
		"HIncrBy":          7,
		"HIncrByFloat":     8,
		"HMSet":            9,
		"HSet":             10,
		"HSetNX":           11,
		"LInsert":          12,
		"LPop":             13,
		"LPopRPush":        14,
		"LPush":            15,
		"LPushX":           16,
		"LRem":             17,
		"LSet":             18,
		"LTrim":            19,
		"RPop":             20,
		"RPopLPush":        21,
		"RPush":            22,
		"RPushX":           23,
		"SAdd":             24,
		"SRem":             25,
		"Set":              26,
		"ZAdd":             27,
		"ZClear":           28,
		"ZIncrBy":          29,
		"ZRem":             30,
		"ZRemRangeByRank":  31,
		"ZRemRangeByScore": 32,
		"Rename":           33,
	}
)

func (x OpType) Enum() *OpType {
	p := new(OpType)
	*p = x
	return p
}

func (x OpType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (OpType) Descriptor() protoreflect.EnumDescriptor {
	return file_nodis_proto_enumTypes[0].Descriptor()
}

func (OpType) Type() protoreflect.EnumType {
	return &file_nodis_proto_enumTypes[0]
}

func (x OpType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use OpType.Descriptor instead.
func (OpType) EnumDescriptor() ([]byte, []int) {
	return file_nodis_proto_rawDescGZIP(), []int{0}
}

type Operation struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type       OpType   `protobuf:"varint,1,opt,name=Type,proto3,enum=pb.OpType" json:"Type,omitempty"`
	Key        string   `protobuf:"bytes,2,opt,name=Key,proto3" json:"Key,omitempty"`
	Member     string   `protobuf:"bytes,3,opt,name=Member,proto3" json:"Member,omitempty"`
	Value      []byte   `protobuf:"bytes,4,opt,name=Value,proto3" json:"Value,omitempty"`
	Expiration int64    `protobuf:"varint,5,opt,name=Expiration,proto3" json:"Expiration,omitempty"`
	Score      float64  `protobuf:"fixed64,6,opt,name=Score,proto3" json:"Score,omitempty"`
	Values     [][]byte `protobuf:"bytes,7,rep,name=Values,proto3" json:"Values,omitempty"`
	DstKey     string   `protobuf:"bytes,8,opt,name=DstKey,proto3" json:"DstKey,omitempty"`
	Pivot      []byte   `protobuf:"bytes,9,opt,name=Pivot,proto3" json:"Pivot,omitempty"`
	Count      int64    `protobuf:"varint,10,opt,name=Count,proto3" json:"Count,omitempty"`
	Index      int64    `protobuf:"varint,11,opt,name=Index,proto3" json:"Index,omitempty"`
	Members    []string `protobuf:"bytes,12,rep,name=Members,proto3" json:"Members,omitempty"`
	Start      int64    `protobuf:"varint,13,opt,name=Start,proto3" json:"Start,omitempty"`
	Stop       int64    `protobuf:"varint,14,opt,name=Stop,proto3" json:"Stop,omitempty"`
	Min        float64  `protobuf:"fixed64,15,opt,name=Min,proto3" json:"Min,omitempty"`
	Max        float64  `protobuf:"fixed64,16,opt,name=Max,proto3" json:"Max,omitempty"`
	Field      string   `protobuf:"bytes,17,opt,name=Field,proto3" json:"Field,omitempty"`
	IncrFloat  float64  `protobuf:"fixed64,18,opt,name=IncrFloat,proto3" json:"IncrFloat,omitempty"`
	IncrInt    int64    `protobuf:"varint,19,opt,name=IncrInt,proto3" json:"IncrInt,omitempty"`
	Before     bool     `protobuf:"varint,20,opt,name=Before,proto3" json:"Before,omitempty"`
}

func (x *Operation) Reset() {
	*x = Operation{}
	if protoimpl.UnsafeEnabled {
		mi := &file_nodis_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Operation) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Operation) ProtoMessage() {}

func (x *Operation) ProtoReflect() protoreflect.Message {
	mi := &file_nodis_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Operation.ProtoReflect.Descriptor instead.
func (*Operation) Descriptor() ([]byte, []int) {
	return file_nodis_proto_rawDescGZIP(), []int{0}
}

func (x *Operation) GetType() OpType {
	if x != nil {
		return x.Type
	}
	return OpType_None
}

func (x *Operation) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *Operation) GetMember() string {
	if x != nil {
		return x.Member
	}
	return ""
}

func (x *Operation) GetValue() []byte {
	if x != nil {
		return x.Value
	}
	return nil
}

func (x *Operation) GetExpiration() int64 {
	if x != nil {
		return x.Expiration
	}
	return 0
}

func (x *Operation) GetScore() float64 {
	if x != nil {
		return x.Score
	}
	return 0
}

func (x *Operation) GetValues() [][]byte {
	if x != nil {
		return x.Values
	}
	return nil
}

func (x *Operation) GetDstKey() string {
	if x != nil {
		return x.DstKey
	}
	return ""
}

func (x *Operation) GetPivot() []byte {
	if x != nil {
		return x.Pivot
	}
	return nil
}

func (x *Operation) GetCount() int64 {
	if x != nil {
		return x.Count
	}
	return 0
}

func (x *Operation) GetIndex() int64 {
	if x != nil {
		return x.Index
	}
	return 0
}

func (x *Operation) GetMembers() []string {
	if x != nil {
		return x.Members
	}
	return nil
}

func (x *Operation) GetStart() int64 {
	if x != nil {
		return x.Start
	}
	return 0
}

func (x *Operation) GetStop() int64 {
	if x != nil {
		return x.Stop
	}
	return 0
}

func (x *Operation) GetMin() float64 {
	if x != nil {
		return x.Min
	}
	return 0
}

func (x *Operation) GetMax() float64 {
	if x != nil {
		return x.Max
	}
	return 0
}

func (x *Operation) GetField() string {
	if x != nil {
		return x.Field
	}
	return ""
}

func (x *Operation) GetIncrFloat() float64 {
	if x != nil {
		return x.IncrFloat
	}
	return 0
}

func (x *Operation) GetIncrInt() int64 {
	if x != nil {
		return x.IncrInt
	}
	return 0
}

func (x *Operation) GetBefore() bool {
	if x != nil {
		return x.Before
	}
	return false
}

type KeyScore struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Member string  `protobuf:"bytes,1,opt,name=Member,proto3" json:"Member,omitempty"`
	Score  float64 `protobuf:"fixed64,2,opt,name=Score,proto3" json:"Score,omitempty"`
}

func (x *KeyScore) Reset() {
	*x = KeyScore{}
	if protoimpl.UnsafeEnabled {
		mi := &file_nodis_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KeyScore) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KeyScore) ProtoMessage() {}

func (x *KeyScore) ProtoReflect() protoreflect.Message {
	mi := &file_nodis_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KeyScore.ProtoReflect.Descriptor instead.
func (*KeyScore) Descriptor() ([]byte, []int) {
	return file_nodis_proto_rawDescGZIP(), []int{1}
}

func (x *KeyScore) GetMember() string {
	if x != nil {
		return x.Member
	}
	return ""
}

func (x *KeyScore) GetScore() float64 {
	if x != nil {
		return x.Score
	}
	return 0
}

type ZSetValue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Values []*KeyScore `protobuf:"bytes,2,rep,name=Values,proto3" json:"Values,omitempty"`
}

func (x *ZSetValue) Reset() {
	*x = ZSetValue{}
	if protoimpl.UnsafeEnabled {
		mi := &file_nodis_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ZSetValue) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ZSetValue) ProtoMessage() {}

func (x *ZSetValue) ProtoReflect() protoreflect.Message {
	mi := &file_nodis_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ZSetValue.ProtoReflect.Descriptor instead.
func (*ZSetValue) Descriptor() ([]byte, []int) {
	return file_nodis_proto_rawDescGZIP(), []int{2}
}

func (x *ZSetValue) GetValues() []*KeyScore {
	if x != nil {
		return x.Values
	}
	return nil
}

type ListValue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Values [][]byte `protobuf:"bytes,2,rep,name=Values,proto3" json:"Values,omitempty"`
}

func (x *ListValue) Reset() {
	*x = ListValue{}
	if protoimpl.UnsafeEnabled {
		mi := &file_nodis_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListValue) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListValue) ProtoMessage() {}

func (x *ListValue) ProtoReflect() protoreflect.Message {
	mi := &file_nodis_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListValue.ProtoReflect.Descriptor instead.
func (*ListValue) Descriptor() ([]byte, []int) {
	return file_nodis_proto_rawDescGZIP(), []int{3}
}

func (x *ListValue) GetValues() [][]byte {
	if x != nil {
		return x.Values
	}
	return nil
}

type StringValue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Value []byte `protobuf:"bytes,2,opt,name=Value,proto3" json:"Value,omitempty"`
}

func (x *StringValue) Reset() {
	*x = StringValue{}
	if protoimpl.UnsafeEnabled {
		mi := &file_nodis_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StringValue) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StringValue) ProtoMessage() {}

func (x *StringValue) ProtoReflect() protoreflect.Message {
	mi := &file_nodis_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StringValue.ProtoReflect.Descriptor instead.
func (*StringValue) Descriptor() ([]byte, []int) {
	return file_nodis_proto_rawDescGZIP(), []int{4}
}

func (x *StringValue) GetValue() []byte {
	if x != nil {
		return x.Value
	}
	return nil
}

type MemberBytes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Member string `protobuf:"bytes,1,opt,name=Member,proto3" json:"Member,omitempty"`
	Value  []byte `protobuf:"bytes,2,opt,name=Value,proto3" json:"Value,omitempty"`
}

func (x *MemberBytes) Reset() {
	*x = MemberBytes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_nodis_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MemberBytes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MemberBytes) ProtoMessage() {}

func (x *MemberBytes) ProtoReflect() protoreflect.Message {
	mi := &file_nodis_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MemberBytes.ProtoReflect.Descriptor instead.
func (*MemberBytes) Descriptor() ([]byte, []int) {
	return file_nodis_proto_rawDescGZIP(), []int{5}
}

func (x *MemberBytes) GetMember() string {
	if x != nil {
		return x.Member
	}
	return ""
}

func (x *MemberBytes) GetValue() []byte {
	if x != nil {
		return x.Value
	}
	return nil
}

type SetValue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Values []string `protobuf:"bytes,2,rep,name=Values,proto3" json:"Values,omitempty"`
}

func (x *SetValue) Reset() {
	*x = SetValue{}
	if protoimpl.UnsafeEnabled {
		mi := &file_nodis_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetValue) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetValue) ProtoMessage() {}

func (x *SetValue) ProtoReflect() protoreflect.Message {
	mi := &file_nodis_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetValue.ProtoReflect.Descriptor instead.
func (*SetValue) Descriptor() ([]byte, []int) {
	return file_nodis_proto_rawDescGZIP(), []int{6}
}

func (x *SetValue) GetValues() []string {
	if x != nil {
		return x.Values
	}
	return nil
}

type HashValue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Values []*MemberBytes `protobuf:"bytes,2,rep,name=Values,proto3" json:"Values,omitempty"`
}

func (x *HashValue) Reset() {
	*x = HashValue{}
	if protoimpl.UnsafeEnabled {
		mi := &file_nodis_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HashValue) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HashValue) ProtoMessage() {}

func (x *HashValue) ProtoReflect() protoreflect.Message {
	mi := &file_nodis_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HashValue.ProtoReflect.Descriptor instead.
func (*HashValue) Descriptor() ([]byte, []int) {
	return file_nodis_proto_rawDescGZIP(), []int{7}
}

func (x *HashValue) GetValues() []*MemberBytes {
	if x != nil {
		return x.Values
	}
	return nil
}

type Entry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type uint32 `protobuf:"varint,1,opt,name=Type,proto3" json:"Type,omitempty"`
	Key  string `protobuf:"bytes,2,opt,name=Key,proto3" json:"Key,omitempty"`
	// Types that are assignable to Value:
	//
	//	*Entry_StringValue
	//	*Entry_ListValue
	//	*Entry_SetValue
	//	*Entry_HashValue
	//	*Entry_ZSetValue
	Value      isEntry_Value `protobuf_oneof:"Value"`
	Expiration int64         `protobuf:"varint,8,opt,name=Expiration,proto3" json:"Expiration,omitempty"`
}

func (x *Entry) Reset() {
	*x = Entry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_nodis_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Entry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Entry) ProtoMessage() {}

func (x *Entry) ProtoReflect() protoreflect.Message {
	mi := &file_nodis_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Entry.ProtoReflect.Descriptor instead.
func (*Entry) Descriptor() ([]byte, []int) {
	return file_nodis_proto_rawDescGZIP(), []int{8}
}

func (x *Entry) GetType() uint32 {
	if x != nil {
		return x.Type
	}
	return 0
}

func (x *Entry) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (m *Entry) GetValue() isEntry_Value {
	if m != nil {
		return m.Value
	}
	return nil
}

func (x *Entry) GetStringValue() *StringValue {
	if x, ok := x.GetValue().(*Entry_StringValue); ok {
		return x.StringValue
	}
	return nil
}

func (x *Entry) GetListValue() *ListValue {
	if x, ok := x.GetValue().(*Entry_ListValue); ok {
		return x.ListValue
	}
	return nil
}

func (x *Entry) GetSetValue() *SetValue {
	if x, ok := x.GetValue().(*Entry_SetValue); ok {
		return x.SetValue
	}
	return nil
}

func (x *Entry) GetHashValue() *HashValue {
	if x, ok := x.GetValue().(*Entry_HashValue); ok {
		return x.HashValue
	}
	return nil
}

func (x *Entry) GetZSetValue() *ZSetValue {
	if x, ok := x.GetValue().(*Entry_ZSetValue); ok {
		return x.ZSetValue
	}
	return nil
}

func (x *Entry) GetExpiration() int64 {
	if x != nil {
		return x.Expiration
	}
	return 0
}

type isEntry_Value interface {
	isEntry_Value()
}

type Entry_StringValue struct {
	StringValue *StringValue `protobuf:"bytes,3,opt,name=StringValue,proto3,oneof"`
}

type Entry_ListValue struct {
	ListValue *ListValue `protobuf:"bytes,4,opt,name=ListValue,proto3,oneof"`
}

type Entry_SetValue struct {
	SetValue *SetValue `protobuf:"bytes,5,opt,name=SetValue,proto3,oneof"`
}

type Entry_HashValue struct {
	HashValue *HashValue `protobuf:"bytes,6,opt,name=HashValue,proto3,oneof"`
}

type Entry_ZSetValue struct {
	ZSetValue *ZSetValue `protobuf:"bytes,7,opt,name=ZSetValue,proto3,oneof"`
}

func (*Entry_StringValue) isEntry_Value() {}

func (*Entry_ListValue) isEntry_Value() {}

func (*Entry_SetValue) isEntry_Value() {}

func (*Entry_HashValue) isEntry_Value() {}

func (*Entry_ZSetValue) isEntry_Value() {}

type Index struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Items []*Index_Item `protobuf:"bytes,1,rep,name=Items,proto3" json:"Items,omitempty"`
}

func (x *Index) Reset() {
	*x = Index{}
	if protoimpl.UnsafeEnabled {
		mi := &file_nodis_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Index) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Index) ProtoMessage() {}

func (x *Index) ProtoReflect() protoreflect.Message {
	mi := &file_nodis_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Index.ProtoReflect.Descriptor instead.
func (*Index) Descriptor() ([]byte, []int) {
	return file_nodis_proto_rawDescGZIP(), []int{9}
}

func (x *Index) GetItems() []*Index_Item {
	if x != nil {
		return x.Items
	}
	return nil
}

type Index_Item struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key  string `protobuf:"bytes,1,opt,name=Key,proto3" json:"Key,omitempty"`
	Data []byte `protobuf:"bytes,2,opt,name=Data,proto3" json:"Data,omitempty"`
}

func (x *Index_Item) Reset() {
	*x = Index_Item{}
	if protoimpl.UnsafeEnabled {
		mi := &file_nodis_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Index_Item) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Index_Item) ProtoMessage() {}

func (x *Index_Item) ProtoReflect() protoreflect.Message {
	mi := &file_nodis_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Index_Item.ProtoReflect.Descriptor instead.
func (*Index_Item) Descriptor() ([]byte, []int) {
	return file_nodis_proto_rawDescGZIP(), []int{9, 0}
}

func (x *Index_Item) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *Index_Item) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

var File_nodis_proto protoreflect.FileDescriptor

var file_nodis_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x6e, 0x6f, 0x64, 0x69, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02, 0x70,
	0x62, 0x22, 0xe1, 0x03, 0x0a, 0x09, 0x4f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12,
	0x1e, 0x0a, 0x04, 0x54, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0a, 0x2e,
	0x70, 0x62, 0x2e, 0x4f, 0x70, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x54, 0x79, 0x70, 0x65, 0x12,
	0x10, 0x0a, 0x03, 0x4b, 0x65, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x4b, 0x65,
	0x79, 0x12, 0x16, 0x0a, 0x06, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x06, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x14, 0x0a, 0x05, 0x56, 0x61, 0x6c,
	0x75, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12,
	0x1e, 0x0a, 0x0a, 0x45, 0x78, 0x70, 0x69, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x0a, 0x45, 0x78, 0x70, 0x69, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12,
	0x14, 0x0a, 0x05, 0x53, 0x63, 0x6f, 0x72, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x01, 0x52, 0x05,
	0x53, 0x63, 0x6f, 0x72, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x18,
	0x07, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x06, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x12, 0x16, 0x0a,
	0x06, 0x44, 0x73, 0x74, 0x4b, 0x65, 0x79, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x44,
	0x73, 0x74, 0x4b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x50, 0x69, 0x76, 0x6f, 0x74, 0x18, 0x09,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x50, 0x69, 0x76, 0x6f, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x43,
	0x6f, 0x75, 0x6e, 0x74, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x43, 0x6f, 0x75, 0x6e,
	0x74, 0x12, 0x14, 0x0a, 0x05, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x05, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x12, 0x18, 0x0a, 0x07, 0x4d, 0x65, 0x6d, 0x62, 0x65,
	0x72, 0x73, 0x18, 0x0c, 0x20, 0x03, 0x28, 0x09, 0x52, 0x07, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72,
	0x73, 0x12, 0x14, 0x0a, 0x05, 0x53, 0x74, 0x61, 0x72, 0x74, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x05, 0x53, 0x74, 0x61, 0x72, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x53, 0x74, 0x6f, 0x70, 0x18,
	0x0e, 0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x53, 0x74, 0x6f, 0x70, 0x12, 0x10, 0x0a, 0x03, 0x4d,
	0x69, 0x6e, 0x18, 0x0f, 0x20, 0x01, 0x28, 0x01, 0x52, 0x03, 0x4d, 0x69, 0x6e, 0x12, 0x10, 0x0a,
	0x03, 0x4d, 0x61, 0x78, 0x18, 0x10, 0x20, 0x01, 0x28, 0x01, 0x52, 0x03, 0x4d, 0x61, 0x78, 0x12,
	0x14, 0x0a, 0x05, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x18, 0x11, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x46, 0x69, 0x65, 0x6c, 0x64, 0x12, 0x1c, 0x0a, 0x09, 0x49, 0x6e, 0x63, 0x72, 0x46, 0x6c, 0x6f,
	0x61, 0x74, 0x18, 0x12, 0x20, 0x01, 0x28, 0x01, 0x52, 0x09, 0x49, 0x6e, 0x63, 0x72, 0x46, 0x6c,
	0x6f, 0x61, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x49, 0x6e, 0x63, 0x72, 0x49, 0x6e, 0x74, 0x18, 0x13,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x49, 0x6e, 0x63, 0x72, 0x49, 0x6e, 0x74, 0x12, 0x16, 0x0a,
	0x06, 0x42, 0x65, 0x66, 0x6f, 0x72, 0x65, 0x18, 0x14, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x42,
	0x65, 0x66, 0x6f, 0x72, 0x65, 0x22, 0x38, 0x0a, 0x08, 0x4b, 0x65, 0x79, 0x53, 0x63, 0x6f, 0x72,
	0x65, 0x12, 0x16, 0x0a, 0x06, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x06, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x14, 0x0a, 0x05, 0x53, 0x63, 0x6f,
	0x72, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x01, 0x52, 0x05, 0x53, 0x63, 0x6f, 0x72, 0x65, 0x22,
	0x31, 0x0a, 0x09, 0x5a, 0x53, 0x65, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x24, 0x0a, 0x06,
	0x56, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x70,
	0x62, 0x2e, 0x4b, 0x65, 0x79, 0x53, 0x63, 0x6f, 0x72, 0x65, 0x52, 0x06, 0x56, 0x61, 0x6c, 0x75,
	0x65, 0x73, 0x22, 0x23, 0x0a, 0x09, 0x4c, 0x69, 0x73, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12,
	0x16, 0x0a, 0x06, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0c, 0x52,
	0x06, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x22, 0x23, 0x0a, 0x0b, 0x53, 0x74, 0x72, 0x69, 0x6e,
	0x67, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x3b, 0x0a, 0x0b,
	0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x42, 0x79, 0x74, 0x65, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x4d,
	0x65, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x4d, 0x65, 0x6d,
	0x62, 0x65, 0x72, 0x12, 0x14, 0x0a, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x22, 0x0a, 0x08, 0x53, 0x65, 0x74,
	0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x18,
	0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x06, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x22, 0x34, 0x0a,
	0x09, 0x48, 0x61, 0x73, 0x68, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x27, 0x0a, 0x06, 0x56, 0x61,
	0x6c, 0x75, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x70, 0x62, 0x2e,
	0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x42, 0x79, 0x74, 0x65, 0x73, 0x52, 0x06, 0x56, 0x61, 0x6c,
	0x75, 0x65, 0x73, 0x22, 0xc4, 0x02, 0x0a, 0x05, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x12, 0x0a,
	0x04, 0x54, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x54, 0x79, 0x70,
	0x65, 0x12, 0x10, 0x0a, 0x03, 0x4b, 0x65, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x4b, 0x65, 0x79, 0x12, 0x33, 0x0a, 0x0b, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x56, 0x61, 0x6c,
	0x75, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x70, 0x62, 0x2e, 0x53, 0x74,
	0x72, 0x69, 0x6e, 0x67, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x48, 0x00, 0x52, 0x0b, 0x53, 0x74, 0x72,
	0x69, 0x6e, 0x67, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x2d, 0x0a, 0x09, 0x4c, 0x69, 0x73, 0x74,
	0x56, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x70, 0x62,
	0x2e, 0x4c, 0x69, 0x73, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x48, 0x00, 0x52, 0x09, 0x4c, 0x69,
	0x73, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x2a, 0x0a, 0x08, 0x53, 0x65, 0x74, 0x56, 0x61,
	0x6c, 0x75, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x70, 0x62, 0x2e, 0x53,
	0x65, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x48, 0x00, 0x52, 0x08, 0x53, 0x65, 0x74, 0x56, 0x61,
	0x6c, 0x75, 0x65, 0x12, 0x2d, 0x0a, 0x09, 0x48, 0x61, 0x73, 0x68, 0x56, 0x61, 0x6c, 0x75, 0x65,
	0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x70, 0x62, 0x2e, 0x48, 0x61, 0x73, 0x68,
	0x56, 0x61, 0x6c, 0x75, 0x65, 0x48, 0x00, 0x52, 0x09, 0x48, 0x61, 0x73, 0x68, 0x56, 0x61, 0x6c,
	0x75, 0x65, 0x12, 0x2d, 0x0a, 0x09, 0x5a, 0x53, 0x65, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x18,
	0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x70, 0x62, 0x2e, 0x5a, 0x53, 0x65, 0x74, 0x56,
	0x61, 0x6c, 0x75, 0x65, 0x48, 0x00, 0x52, 0x09, 0x5a, 0x53, 0x65, 0x74, 0x56, 0x61, 0x6c, 0x75,
	0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x45, 0x78, 0x70, 0x69, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18,
	0x08, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0a, 0x45, 0x78, 0x70, 0x69, 0x72, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x42, 0x07, 0x0a, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x5b, 0x0a, 0x05, 0x49, 0x6e,
	0x64, 0x65, 0x78, 0x12, 0x24, 0x0a, 0x05, 0x49, 0x74, 0x65, 0x6d, 0x73, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x70, 0x62, 0x2e, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x2e, 0x49, 0x74,
	0x65, 0x6d, 0x52, 0x05, 0x49, 0x74, 0x65, 0x6d, 0x73, 0x1a, 0x2c, 0x0a, 0x04, 0x49, 0x74, 0x65,
	0x6d, 0x12, 0x10, 0x0a, 0x03, 0x4b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x4b, 0x65, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x44, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x04, 0x44, 0x61, 0x74, 0x61, 0x2a, 0xa3, 0x03, 0x0a, 0x06, 0x4f, 0x70, 0x54, 0x79,
	0x70, 0x65, 0x12, 0x08, 0x0a, 0x04, 0x4e, 0x6f, 0x6e, 0x65, 0x10, 0x00, 0x12, 0x09, 0x0a, 0x05,
	0x43, 0x6c, 0x65, 0x61, 0x72, 0x10, 0x01, 0x12, 0x07, 0x0a, 0x03, 0x44, 0x65, 0x6c, 0x10, 0x02,
	0x12, 0x0a, 0x0a, 0x06, 0x45, 0x78, 0x70, 0x69, 0x72, 0x65, 0x10, 0x03, 0x12, 0x0c, 0x0a, 0x08,
	0x45, 0x78, 0x70, 0x69, 0x72, 0x65, 0x41, 0x74, 0x10, 0x04, 0x12, 0x0a, 0x0a, 0x06, 0x48, 0x43,
	0x6c, 0x65, 0x61, 0x72, 0x10, 0x05, 0x12, 0x08, 0x0a, 0x04, 0x48, 0x44, 0x65, 0x6c, 0x10, 0x06,
	0x12, 0x0b, 0x0a, 0x07, 0x48, 0x49, 0x6e, 0x63, 0x72, 0x42, 0x79, 0x10, 0x07, 0x12, 0x10, 0x0a,
	0x0c, 0x48, 0x49, 0x6e, 0x63, 0x72, 0x42, 0x79, 0x46, 0x6c, 0x6f, 0x61, 0x74, 0x10, 0x08, 0x12,
	0x09, 0x0a, 0x05, 0x48, 0x4d, 0x53, 0x65, 0x74, 0x10, 0x09, 0x12, 0x08, 0x0a, 0x04, 0x48, 0x53,
	0x65, 0x74, 0x10, 0x0a, 0x12, 0x0a, 0x0a, 0x06, 0x48, 0x53, 0x65, 0x74, 0x4e, 0x58, 0x10, 0x0b,
	0x12, 0x0b, 0x0a, 0x07, 0x4c, 0x49, 0x6e, 0x73, 0x65, 0x72, 0x74, 0x10, 0x0c, 0x12, 0x08, 0x0a,
	0x04, 0x4c, 0x50, 0x6f, 0x70, 0x10, 0x0d, 0x12, 0x0d, 0x0a, 0x09, 0x4c, 0x50, 0x6f, 0x70, 0x52,
	0x50, 0x75, 0x73, 0x68, 0x10, 0x0e, 0x12, 0x09, 0x0a, 0x05, 0x4c, 0x50, 0x75, 0x73, 0x68, 0x10,
	0x0f, 0x12, 0x0a, 0x0a, 0x06, 0x4c, 0x50, 0x75, 0x73, 0x68, 0x58, 0x10, 0x10, 0x12, 0x08, 0x0a,
	0x04, 0x4c, 0x52, 0x65, 0x6d, 0x10, 0x11, 0x12, 0x08, 0x0a, 0x04, 0x4c, 0x53, 0x65, 0x74, 0x10,
	0x12, 0x12, 0x09, 0x0a, 0x05, 0x4c, 0x54, 0x72, 0x69, 0x6d, 0x10, 0x13, 0x12, 0x08, 0x0a, 0x04,
	0x52, 0x50, 0x6f, 0x70, 0x10, 0x14, 0x12, 0x0d, 0x0a, 0x09, 0x52, 0x50, 0x6f, 0x70, 0x4c, 0x50,
	0x75, 0x73, 0x68, 0x10, 0x15, 0x12, 0x09, 0x0a, 0x05, 0x52, 0x50, 0x75, 0x73, 0x68, 0x10, 0x16,
	0x12, 0x0a, 0x0a, 0x06, 0x52, 0x50, 0x75, 0x73, 0x68, 0x58, 0x10, 0x17, 0x12, 0x08, 0x0a, 0x04,
	0x53, 0x41, 0x64, 0x64, 0x10, 0x18, 0x12, 0x08, 0x0a, 0x04, 0x53, 0x52, 0x65, 0x6d, 0x10, 0x19,
	0x12, 0x07, 0x0a, 0x03, 0x53, 0x65, 0x74, 0x10, 0x1a, 0x12, 0x08, 0x0a, 0x04, 0x5a, 0x41, 0x64,
	0x64, 0x10, 0x1b, 0x12, 0x0a, 0x0a, 0x06, 0x5a, 0x43, 0x6c, 0x65, 0x61, 0x72, 0x10, 0x1c, 0x12,
	0x0b, 0x0a, 0x07, 0x5a, 0x49, 0x6e, 0x63, 0x72, 0x42, 0x79, 0x10, 0x1d, 0x12, 0x08, 0x0a, 0x04,
	0x5a, 0x52, 0x65, 0x6d, 0x10, 0x1e, 0x12, 0x13, 0x0a, 0x0f, 0x5a, 0x52, 0x65, 0x6d, 0x52, 0x61,
	0x6e, 0x67, 0x65, 0x42, 0x79, 0x52, 0x61, 0x6e, 0x6b, 0x10, 0x1f, 0x12, 0x14, 0x0a, 0x10, 0x5a,
	0x52, 0x65, 0x6d, 0x52, 0x61, 0x6e, 0x67, 0x65, 0x42, 0x79, 0x53, 0x63, 0x6f, 0x72, 0x65, 0x10,
	0x20, 0x12, 0x0a, 0x0a, 0x06, 0x52, 0x65, 0x6e, 0x61, 0x6d, 0x65, 0x10, 0x21, 0x42, 0x07, 0x5a,
	0x05, 0x2e, 0x2e, 0x2f, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_nodis_proto_rawDescOnce sync.Once
	file_nodis_proto_rawDescData = file_nodis_proto_rawDesc
)

func file_nodis_proto_rawDescGZIP() []byte {
	file_nodis_proto_rawDescOnce.Do(func() {
		file_nodis_proto_rawDescData = protoimpl.X.CompressGZIP(file_nodis_proto_rawDescData)
	})
	return file_nodis_proto_rawDescData
}

var file_nodis_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_nodis_proto_msgTypes = make([]protoimpl.MessageInfo, 11)
var file_nodis_proto_goTypes = []interface{}{
	(OpType)(0),         // 0: pb.OpType
	(*Operation)(nil),   // 1: pb.Operation
	(*KeyScore)(nil),    // 2: pb.KeyScore
	(*ZSetValue)(nil),   // 3: pb.ZSetValue
	(*ListValue)(nil),   // 4: pb.ListValue
	(*StringValue)(nil), // 5: pb.StringValue
	(*MemberBytes)(nil), // 6: pb.MemberBytes
	(*SetValue)(nil),    // 7: pb.SetValue
	(*HashValue)(nil),   // 8: pb.HashValue
	(*Entry)(nil),       // 9: pb.Entry
	(*Index)(nil),       // 10: pb.Index
	(*Index_Item)(nil),  // 11: pb.Index.Item
}
var file_nodis_proto_depIdxs = []int32{
	0,  // 0: pb.Operation.Type:type_name -> pb.OpType
	2,  // 1: pb.ZSetValue.Values:type_name -> pb.KeyScore
	6,  // 2: pb.HashValue.Values:type_name -> pb.MemberBytes
	5,  // 3: pb.Entry.StringValue:type_name -> pb.StringValue
	4,  // 4: pb.Entry.ListValue:type_name -> pb.ListValue
	7,  // 5: pb.Entry.SetValue:type_name -> pb.SetValue
	8,  // 6: pb.Entry.HashValue:type_name -> pb.HashValue
	3,  // 7: pb.Entry.ZSetValue:type_name -> pb.ZSetValue
	11, // 8: pb.Index.Items:type_name -> pb.Index.Item
	9,  // [9:9] is the sub-list for method output_type
	9,  // [9:9] is the sub-list for method input_type
	9,  // [9:9] is the sub-list for extension type_name
	9,  // [9:9] is the sub-list for extension extendee
	0,  // [0:9] is the sub-list for field type_name
}

func init() { file_nodis_proto_init() }
func file_nodis_proto_init() {
	if File_nodis_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_nodis_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Operation); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_nodis_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*KeyScore); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_nodis_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ZSetValue); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_nodis_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListValue); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_nodis_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StringValue); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_nodis_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MemberBytes); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_nodis_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetValue); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_nodis_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HashValue); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_nodis_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Entry); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_nodis_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Index); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_nodis_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Index_Item); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_nodis_proto_msgTypes[8].OneofWrappers = []interface{}{
		(*Entry_StringValue)(nil),
		(*Entry_ListValue)(nil),
		(*Entry_SetValue)(nil),
		(*Entry_HashValue)(nil),
		(*Entry_ZSetValue)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_nodis_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   11,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_nodis_proto_goTypes,
		DependencyIndexes: file_nodis_proto_depIdxs,
		EnumInfos:         file_nodis_proto_enumTypes,
		MessageInfos:      file_nodis_proto_msgTypes,
	}.Build()
	File_nodis_proto = out.File
	file_nodis_proto_rawDesc = nil
	file_nodis_proto_goTypes = nil
	file_nodis_proto_depIdxs = nil
}
