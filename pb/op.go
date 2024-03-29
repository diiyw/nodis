package pb

type Op struct {
	*Operation
}

func NewOp(typ OpType, key string) *Op {
	return &Op{
		Operation: &Operation{
			Type: typ,
			Key:  key,
		},
	}
}

func (o *Op) Value(v []byte) *Op {
	o.Operation.Value = v
	return o
}

// Values set the values
func (o *Op) Values(v [][]byte) *Op {
	o.Operation.Values = v
	return o
}

// Score set the score
func (o *Op) Score(score float64) *Op {
	o.Operation.Score = score
	return o
}

// Member set the member
func (o *Op) Member(member string) *Op {
	o.Operation.Member = member
	return o
}

// Expiration set the expiration
func (o *Op) Expiration(seconds int64) *Op {
	o.Operation.Expiration = seconds
	return o
}

// DstKey set the destination key
func (o *Op) DstKey(key string) *Op {
	o.Operation.DstKey = key
	return o
}

// Pivot set the pivot
func (o *Op) Pivot(pivot []byte) *Op {
	o.Operation.Pivot = pivot
	return o
}

// Count set the count
func (o *Op) Count(count int64) *Op {
	o.Operation.Count = count
	return o
}

// Index set the index
func (o *Op) Index(index int64) *Op {
	o.Operation.Index = index
	return o
}

// Members set the members
func (o *Op) Members(members []string) *Op {
	o.Operation.Members = members
	return o
}

// Start set the start
func (o *Op) Start(start int64) *Op {
	o.Operation.Start = start
	return o
}

// Stop set the stop
func (o *Op) Stop(stop int64) *Op {
	o.Operation.Stop = stop
	return o
}

// Min set the min
func (o *Op) Min(min float64) *Op {
	o.Operation.Min = min
	return o
}

// Max set the max
func (o *Op) Max(max float64) *Op {
	o.Operation.Max = max
	return o
}

// Field set the field
func (o *Op) Field(field string) *Op {
	o.Operation.Field = field
	return o
}

// IncrFloat set the increment
func (o *Op) IncrFloat(i float64) *Op {
	o.Operation.IncrFloat = i
	return o
}

// IncrInt set the increment
func (o *Op) IncrInt(i int64) *Op {
	o.Operation.IncrInt = i
	return o
}

// Before set the before
func (o *Op) Before(before bool) *Op {
	o.Operation.Before = before
	return o
}
