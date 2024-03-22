package ds

type DataStruct interface {
	Lock()
	Unlock()
	RLock()
	RUnlock()
	Type() DataType
}

type DataType uint8

const (
	None DataType = iota
	String
	List
	Hash
	Set
	ZSet
)

var DataTypeMap = map[DataType]string{
	None:   "none",
	String: "string",
	List:   "list",
	Hash:   "hash",
	Set:    "set",
	ZSet:   "zset",
}
