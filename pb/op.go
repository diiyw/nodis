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
