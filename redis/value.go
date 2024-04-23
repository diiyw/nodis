package redis

const (
	StringType  = '+'
	ErrType     = '-'
	IntegerType = ':'
	BulkType    = '$'
	ArrayType   = '*'
	MapType     = '%'
	DoubleType  = ','
	NullType    = '_'
)

var (
	options = map[string]bool{
		"NX":         true,
		"XX":         true,
		"KEEPTTL":    true,
		"GET":        true,
		"LT":         true,
		"GT":         true,
		"CH":         true,
		"INCR":       true,
		"WITHSCORES": true,
	}

	args = map[string]bool{
		"EX":    true,
		"PX":    true,
		"EXAT":  true,
		"PXAT":  true,
		"MATCH": true,
		"COUNT": true,
	}
)

type Value struct {
	typ     uint8
	Str     string
	Integer int64
	Bulk    string
	Double  float64
	Array   []Value
	Map     map[string]Value
	Options map[string]bool
	Args    map[string]Value
}

func StringValue(v string) Value {
	return Value{typ: StringType, Str: v}
}

func ErrorValue(v string) Value {
	return Value{typ: ErrType, Str: v}
}

func BulkValue(v string) Value {
	return Value{typ: BulkType, Bulk: v}
}

func IntegerValue(v int64) Value {
	return Value{typ: IntegerType, Integer: v}
}

func DoubleValue(v float64) Value {
	return Value{typ: DoubleType, Double: v}
}

func ArrayValue(v ...Value) Value {
	return Value{typ: ArrayType, Array: v}
}

func MapValue(v map[string]Value) Value {
	return Value{typ: MapType, Map: v}
}

func NullValue() Value {
	return Value{typ: NullType}
}
