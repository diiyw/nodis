package nodis

import (
	"strconv"
	"time"
	"unsafe"

	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/ds/str"
	"github.com/diiyw/nodis/pb"
)

func (n *Nodis) newStr() ds.DataStruct {
	return str.NewString()
}

// Set a key with a value and a TTL
func (n *Nodis) Set(key string, value []byte) {
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, n.newStr)
		meta.ds.(*str.String).Set(value)
		n.notify(pb.NewOp(pb.OpType_Set, key).Value(value))
		return nil
	})
}

// GetSet
func (n *Nodis) GetSet(key string, value []byte) []byte {
	var v []byte
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, n.newStr)
		v = meta.ds.(*str.String).GetSet(value)
		n.notify(pb.NewOp(pb.OpType_Set, key).Value(value))
		return nil
	})
	return v
}

// SetEX set a key with specified expire time, in seconds (a positive integer).
func (n *Nodis) SetEX(key string, value []byte, seconds int64) {
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, n.newStr)
		meta.key.expiration = time.Now().UnixMilli()
		meta.key.expiration += seconds * 1000
		meta.ds.(*str.String).Set(value)
		n.notify(pb.NewOp(pb.OpType_Set, key).Value(value).Expiration(meta.key.expiration))
		return nil
	})
}

// SetPX set a key with specified expire time, in milliseconds (a positive integer).
func (n *Nodis) SetPX(key string, value []byte, milliseconds int64) {
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, n.newStr)
		meta.key.expiration = time.Now().UnixMilli()
		meta.key.expiration += milliseconds
		n.notify(pb.NewOp(pb.OpType_Set, key).Value(value).Expiration(meta.key.expiration))
		meta.ds.(*str.String).Set(value)
		return nil
	})
}

// SetNX set a key with a value if it does not exist
func (n *Nodis) SetNX(key string, value []byte) bool {
	var ok bool
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, nil)
		if meta.isOk() {
			return nil
		}
		meta = tx.newKey(meta, key, n.newStr)
		n.notify(pb.NewOp(pb.OpType_Set, key).Value(value))
		meta.ds.(*str.String).Set(value)
		ok = true
		return nil
	})
	return ok
}

// SetXX set a key with a value if it exists
func (n *Nodis) SetXX(key string, value []byte) bool {
	var ok bool
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, nil)
		if !meta.isOk() {
			return nil
		}
		n.notify(pb.NewOp(pb.OpType_Set, key).Value(value))
		meta.ds.(*str.String).Set(value)
		ok = true
		return nil
	})
	return ok
}

// Get a key
func (n *Nodis) Get(key string) []byte {
	var v []byte
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		v = meta.ds.(*str.String).Get()
		return nil
	})
	return v
}

// Incr increment the integer value of a key by one
func (n *Nodis) Incr(key string) (int64, error) {
	var v int64
	err := n.exec(func(tx *Tx) error {
		var err error
		meta := tx.writeKey(key, n.newStr)
		v, err = meta.ds.(*str.String).Incr(1)
		if err != nil {

			return err
		}
		vv := strconv.FormatInt(v, 10)
		m := unsafe.Slice(unsafe.StringData(vv), len(vv))
		n.notify(pb.NewOp(pb.OpType_Set, key).Value(m))
		return nil
	})
	return v, err
}

func (n *Nodis) IncrBy(key string, increment int64) (int64, error) {
	var v int64
	err := n.exec(func(tx *Tx) error {
		var err error
		meta := tx.writeKey(key, n.newStr)
		v, err = meta.ds.(*str.String).Incr(increment)
		if err != nil {
			return nil
		}
		vv := strconv.FormatInt(v, 10)
		m := unsafe.Slice(unsafe.StringData(vv), len(vv))
		n.notify(pb.NewOp(pb.OpType_Set, key).Value(m))
		return nil
	})
	return v, err
}

// Decr decrement the integer value of a key by one
func (n *Nodis) Decr(key string) (int64, error) {
	var v int64
	err := n.exec(func(tx *Tx) error {
		var err error
		meta := tx.writeKey(key, n.newStr)
		v, err = meta.ds.(*str.String).Decr(1)
		if err != nil {

			return err
		}
		vv := strconv.FormatInt(v, 10)
		m := unsafe.Slice(unsafe.StringData(vv), len(vv))
		n.notify(pb.NewOp(pb.OpType_Set, key).Value(m))
		return nil
	})
	return v, err
}

func (n *Nodis) DecrBy(key string, decrement int64) (int64, error) {
	var v int64
	err := n.exec(func(tx *Tx) error {
		var err error
		meta := tx.writeKey(key, n.newStr)
		v, err = meta.ds.(*str.String).Decr(decrement)
		if err != nil {

			return err
		}
		vv := strconv.FormatInt(v, 10)
		m := unsafe.Slice(unsafe.StringData(vv), len(vv))
		n.notify(pb.NewOp(pb.OpType_Set, key).Value(m))
		return nil
	})
	return v, err
}

func (n *Nodis) IncrByFloat(key string, increment float64) (float64, error) {
	var v float64
	err := n.exec(func(tx *Tx) error {
		var err error
		meta := tx.writeKey(key, n.newStr)
		v, err = meta.ds.(*str.String).IncrByFloat(increment)
		if err != nil {
			return err
		}
		vv := strconv.FormatFloat(v, 'f', -1, 64)
		m := unsafe.Slice(unsafe.StringData(vv), len(vv))
		n.notify(pb.NewOp(pb.OpType_Set, key).Value(m))
		return nil
	})
	return v, err
}

// SetBit set a bit in a key
func (n *Nodis) SetBit(key string, offset int64, value bool) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, n.newStr)
		k := meta.ds.(*str.String)
		v = k.SetBit(offset, value)
		n.notify(pb.NewOp(pb.OpType_Set, key).Value(k.Get()))
		return nil
	})
	return v
}

// GetBit get a bit in a key
func (n *Nodis) GetBit(key string, offset int64) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		v = meta.ds.(*str.String).GetBit(offset)
		return nil
	})
	return v
}

// BitCount returns the number of bits set to 1
func (n *Nodis) BitCount(key string, start, end int64, bit bool) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		if bit {
			v = meta.ds.(*str.String).BitCountByBit(start, end)
		} else {
			v = meta.ds.(*str.String).BitCount(start, end)
		}
		return nil
	})
	return v
}

// Append a value to a key
func (n *Nodis) Append(key string, value []byte) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, n.newStr)
		k := meta.ds.(*str.String)
		v = k.Append(value)
		n.notify(pb.NewOp(pb.OpType_Set, key).Value(k.Get()))
		return nil
	})
	return v
}

// GetRange returns the substring of the string value stored at key, determined by the offsets start and end (both are inclusive).
func (n *Nodis) GetRange(key string, start, end int64) []byte {
	var v []byte
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {

			return nil
		}
		v = meta.ds.(*str.String).GetRange(start, end)
		return nil
	})
	return v
}

// StrLen returns the length of the string value stored at key
func (n *Nodis) StrLen(key string) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		v = meta.ds.(*str.String).Strlen()
		return nil
	})
	return v
}

// MSet sets the given keys to their respective values
func (n *Nodis) MSet(pairs ...string) {
	if len(pairs)%2 != 0 {
		return
	}
	for i := 0; i < len(pairs); i += 2 {
		n.Set(pairs[i], unsafe.Slice(unsafe.StringData(pairs[i+1]), len(pairs[i+1])))
	}
}
