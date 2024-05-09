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

func (d DataType) String() string {
	switch d {
	case None:
		return "NONE"
	case String:
		return "STRING"
	case List:
		return "LIST"
	case Hash:
		return "HASH"
	case Set:
		return "SET"
	case ZSet:
		return "ZSET"
	default:
		return "NONE"
	}
}

func StringToDataType(s string) DataType {
	switch s {
	case "STRING":
		return String
	case "LIST":
		return List
	case "HASH":
		return Hash
	case "SET":
		return Set
	case "ZSET":
		return ZSet
	default:
		return None
	}
}
