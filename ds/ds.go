package ds

type DataStruct interface {
	Type() DataType
}

type DataType uint8

const (
	// 0 => none, (key didn't exist)
	// 1 => string,
	// 2 => set,
	// 3 => list,
	// 4 => zset,
	// 5 => hash
	// 6 => stream
	None DataType = iota
	String
	Set
	List
	ZSet
	Hash
)

var DataTypeMap = map[DataType]string{
	None:   "none",
	String: "string",
	List:   "list",
	Hash:   "hash",
	Set:    "set",
	ZSet:   "zset",
}
