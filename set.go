package nodis

import (
	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/ds/set"
	"github.com/diiyw/nodis/patch"
)

func (n *Nodis) newSet() ds.Value {
	return set.NewSet()
}

// SAdd adds the specified members to the set stored at key.
func (n *Nodis) SAdd(key string, members ...string) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		n := n.inTx(tx)
		meta := tx.writeKey(key, n.newSet)
		v = meta.value.(*set.Set).SAdd(members...)
		n.signalModifiedKey(key, meta)
		n.notify(func() []patch.Op {
			return []patch.Op{{Type: patch.OpTypeSAdd, Data: &patch.OpSAdd{Key: key, Members: members}}}
		})
		return nil
	})
	return v
}

// SCard gets the set members count.
func (n *Nodis) SCard(key string) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		v = meta.value.(*set.Set).SCard()
		return nil
	})
	return v
}

// SDiff gets the difference between sets.
func (n *Nodis) SDiff(keys ...string) []string {
	if len(keys) == 0 {
		return nil
	}
	var v []string
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(keys[0])
		if !meta.isOk() {
			return nil
		}
		otherSets := make([]*set.Set, 0, len(keys)-1)
		for _, s := range keys[1:] {
			metaX := tx.readKey(s)
			if !metaX.isOk() {
				continue
			}
			otherSets = append(otherSets, metaX.value.(*set.Set))
		}
		v = meta.value.(*set.Set).SDiff(otherSets...)
		return nil
	})
	return v
}

// SDiffStore stores the difference between sets.
func (n *Nodis) SDiffStore(destination string, keys ...string) int64 {
	var count int64
	_ = n.exec(func(tx *Tx) error {
		n := n.inTx(tx)
		members := n.SDiff(keys...)
		n.Del(destination)
		if len(members) > 0 {
			count = n.SAdd(destination, members...)
		}
		return nil
	})
	return count
}

func (n *Nodis) SInter(keys ...string) []string {
	if len(keys) == 0 {
		return nil
	}
	if len(keys) == 1 {
		return n.SMembers(keys[0])
	}
	var v []string
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(keys[0])
		if !meta.isOk() {
			return nil
		}
		otherSets := make([]*set.Set, 0, len(keys)-1)
		for _, s := range keys[1:] {
			setDs := tx.readKey(s)
			if !setDs.isOk() {
				return nil
			}
			otherSets = append(otherSets, setDs.value.(*set.Set))
		}
		v = meta.value.(*set.Set).SInter(otherSets...)
		return nil
	})
	return v
}

func (n *Nodis) SInterStore(destination string, keys ...string) int64 {
	var count int64
	_ = n.exec(func(tx *Tx) error {
		n := n.inTx(tx)
		members := n.SInter(keys...)
		n.Del(destination)
		if len(members) > 0 {
			count = n.SAdd(destination, members...)
		}
		return nil
	})
	return count
}

func (n *Nodis) SUnion(keys ...string) []string {
	if len(keys) == 0 {
		return nil
	}
	if len(keys) == 1 {
		return n.SMembers(keys[0])
	}
	var v []string
	_ = n.exec(func(tx *Tx) error {
		otherSets := make([]*set.Set, 0, len(keys))
		for _, s := range keys {
			setDs := tx.readKey(s)
			if !setDs.isOk() {
				continue
			}
			otherSets = append(otherSets, setDs.value.(*set.Set))
		}
		if len(otherSets) == 0 {
			return nil
		}
		v = otherSets[0].SUnion(otherSets[1:]...)
		return nil
	})
	return v
}

func (n *Nodis) SUnionStore(destination string, keys ...string) int64 {
	var count int64
	_ = n.exec(func(tx *Tx) error {
		n := n.inTx(tx)
		members := n.SUnion(keys...)
		n.Del(destination)
		if len(members) > 0 {
			count = n.SAdd(destination, members...)
		}
		return nil
	})
	return count
}

func (n *Nodis) SIsMember(key, member string) bool {
	var v bool
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		v = meta.value.(*set.Set).SIsMember(member)
		return nil
	})
	return v
}

// SMembers returns all the members of the set value stored at key.
func (n *Nodis) SMembers(key string) []string {
	var v []string
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		v = meta.value.(*set.Set).SMembers()
		return nil
	})
	return v
}

// SRem removes the specified members from the set stored at key.
func (n *Nodis) SRem(key string, members ...string) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		n := n.inTx(tx)
		meta := tx.writeKey(key, nil)
		if !meta.isOk() {
			return nil
		}
		v = meta.value.(*set.Set).SRem(members...)
		n.signalModifiedKey(key, meta)
		n.notify(func() []patch.Op {
			return []patch.Op{{Type: patch.OpTypeSRem, Data: &patch.OpSRem{Key: key, Members: members}}}
		})
		return nil
	})
	return v
}

// SScan scans the set value stored at key.
func (n *Nodis) SScan(key string, cursor int64, match string, count int64) (int64, []string) {
	var c int64
	var v []string
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		c, v = meta.value.(*set.Set).SScan(cursor, match, count)
		return nil
	})
	return c, v
}

// SPop removes and returns a random element from the set value stored at key.
func (n *Nodis) SPop(key string, count int64) []string {
	var v []string
	_ = n.exec(func(tx *Tx) error {
		n := n.inTx(tx)
		meta := tx.writeKey(key, nil)
		if !meta.isOk() {
			return nil
		}
		if count == 0 {
			count = 1
		}
		v = meta.value.(*set.Set).SPop(count)
		n.signalModifiedKey(key, meta)
		n.notify(func() []patch.Op {
			return []patch.Op{{Type: patch.OpTypeSRem, Data: &patch.OpSRem{Key: key, Members: v}}}
		})
		return nil
	})
	return v
}

// SMove moves a member from one set to another.
func (n *Nodis) SMove(source, destination, member string) bool {
	var moved bool
	_ = n.exec(func(tx *Tx) error {
		n := n.inTx(tx)
		src := tx.writeKey(source, nil)
		if !src.isOk() {
			return nil
		}
		sourceSet := src.value.(*set.Set)
		if !sourceSet.SIsMember(member) {
			return nil
		}
		if source == destination {
			moved = true
			return nil
		}
		dst := tx.writeKey(destination, n.newSet)
		destinationSet := dst.value.(*set.Set)
		sourceSet.SRem(member)
		destinationSet.SAdd(member)
		if sourceSet.SCard() == 0 {
			tx.delKey(source)
		}
		n.signalModifiedKey(source, src)
		n.signalModifiedKey(destination, dst)
		n.notify(func() []patch.Op {
			return []patch.Op{{Type: patch.OpTypeSRem, Data: &patch.OpSRem{Key: source, Members: []string{member}}}, {Type: patch.OpTypeSAdd, Data: &patch.OpSAdd{Key: destination, Members: []string{member}}}}
		})
		moved = true
		return nil
	})
	return moved
}

// SRandMember returns one or more random elements from the set value stored at key.
func (n *Nodis) SRandMember(key string, count int64) []string {
	var v []string
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		v = meta.value.(*set.Set).SRandMember(count)
		return nil
	})
	return v
}
