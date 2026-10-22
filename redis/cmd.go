package redis

import "strings"

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
	BYLEX      int
	BYSCORE    int
	LIMIT      int
	BYTE       int
	BIT        int
	NUMKEYS    int
	WEIGHTS    int
	AGGREGATE  int
	REV        int
	TYPE       int
	M          int
	KM         int
	FT         int
	MI         int
	ASC        int
	DESC       int
	ANY        int
	WITHCOORD  int
	WITHDIST   int
	WITHHASH   int
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

// Clone owns queued arguments beyond the next reusable parser buffer read.
func (c Command) Clone() Command {
	out := c
	out.Name = strings.Clone(c.Name)
	out.Args = make([]string, len(c.Args))
	for i, arg := range c.Args {
		out.Args[i] = strings.Clone(arg)
	}
	return out
}

// ParseOptions recognizes options using the command's argument grammar.
func (c *Command) ParseOptions() {
	r := Reader{cmd: *c}
	r.cmd.Options = Options{}
	r.options()
	c.Options = r.cmd.Options
}
