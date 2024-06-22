package ds

type Value interface {
	Type() ValueType
	GetValue() []byte
}

type ValueType uint8

const (
	// 0 => none, (key didn't exist)
	// 1 => string,
	// 2 => set,
	// 3 => list,
	// 4 => zset,
	// 5 => hash
	// 6 => stream
	None ValueType = iota
	String
	Set
	List
	ZSet
	Hash
)

func (d ValueType) String() string {
	switch d {
	case None:
		return "none"
	case String:
		return "string"
	case List:
		return "list"
	case Hash:
		return "hash"
	case Set:
		return "set"
	case ZSet:
		return "zset"
	default:
		return "none"
	}
}

func StringToDataType(s string) ValueType {
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
