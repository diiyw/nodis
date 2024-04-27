package redis

type Command struct {
	Name    string
	Args    []string
	Options Options
}

type Options struct {
	NX         int
	XX         int
	KEEPTTL    int
	GET        int
	LT         int
	GT         int
	CH         int
	INCR       int
	WITHSCORES int
	EX         int
	PX         int
	EXAT       int
	PXAT       int
	MATCH      int
	COUNT      int
	OFFSET     int
	BYLEX      int
	BYSCORE    int
}

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