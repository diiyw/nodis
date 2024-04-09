package nodis

import (
	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/ds/set"
	"github.com/diiyw/nodis/pb"
)

func (n *Nodis) newSet() ds.DataStruct {
	return set.NewSet()
}

// SAdd adds the specified members to the set stored at key.
func (n *Nodis) SAdd(key string, members ...string) int64 {
	tx := n.writeKey(key, n.newSet)
	v := tx.ds.(*set.Set).SAdd(members...)
	tx.commit()
	n.notify(pb.NewOp(pb.OpType_SAdd, key).Members(members))
	return v
}

// SCard gets the set members count.
func (n *Nodis) SCard(key string) int64 {
	tx := n.writeKey(key, nil)
	if !tx.isOk() {
		return 0
	}
	v := tx.ds.(*set.Set).SCard()
	tx.commit()
	return v
}

// SDiff gets the difference between sets.
func (n *Nodis) SDiff(keys ...string) []string {
	if len(keys) == 0 {
		return nil
	}
	tx := n.readKey(keys[0])
	if !tx.isOk() {
		return nil
	}
	lockedKeys := []*Tx{}
	otherSets := make([]*set.Set, len(keys)-1)
	for i, s := range keys[1:] {
		setDsTx := n.readKey(s)
		if !setDsTx.isOk() {
			continue
		}
		lockedKeys = append(lockedKeys, setDsTx)
		otherSets[i] = setDsTx.ds.(*set.Set)
	}
	v := tx.ds.(*set.Set).SDiff(otherSets...)
	for _, s := range lockedKeys {
		s.commit()
	}
	tx.commit()
	return v
}

// SInter gets the intersection between sets.
func (n *Nodis) SInter(keys ...string) []string {
	if len(keys) == 0 {
		return nil
	}
	tx := n.readKey(keys[0])
	if !tx.isOk() {
		return nil
	}
	lockedSets := []*Tx{}
	otherSets := make([]*set.Set, len(keys)-1)
	for i, s := range keys[1:] {
		setDs := n.readKey(s)
		if !setDs.isOk() {
			continue
		}
		lockedSets = append(lockedSets, setDs)
		otherSets[i] = setDs.ds.(*set.Set)
	}
	v := tx.ds.(*set.Set).SInter(otherSets...)
	for _, s := range lockedSets {
		s.commit()
	}
	tx.commit()
	return v
}

// SIsMember returns if member is a member of the set stored at key.
func (n *Nodis) SIsMember(key, member string) bool {
	tx := n.readKey(key)
	if !tx.isOk() {
		return false
	}
	v := tx.ds.(*set.Set).SIsMember(member)
	tx.commit()
	return v
}

// SMembers returns all the members of the set value stored at key.
func (n *Nodis) SMembers(key string) []string {
	tx := n.readKey(key)
	if !tx.isOk() {
		return nil
	}
	v := tx.ds.(*set.Set).SMembers()
	tx.commit()
	return v
}

// SRem removes the specified members from the set stored at key.
func (n *Nodis) SRem(key string, members ...string) int64 {
	tx := n.writeKey(key, nil)
	if !tx.isOk() {
		return 0
	}
	v := tx.ds.(*set.Set).SRem(members...)
	tx.commit()
	n.notify(pb.NewOp(pb.OpType_SRem, key).Members(members))
	return v
}

// SScan scans the set value stored at key.
func (n *Nodis) SScan(key string, cursor int64, match string, count int64) (int64, []string) {
	tx := n.readKey(key)
	if !tx.isOk() {
		return 0, nil
	}
	c, v := tx.ds.(*set.Set).SScan(cursor, match, count)
	tx.commit()
	return c, v
}

// SPop removes and returns a random element from the set value stored at key.
func (n *Nodis) SPop(key string, count int64) []string {
	tx := n.writeKey(key, nil)
	if !tx.isOk() {
		return nil
	}
	if count == 0 {
		count = 1
	}
	members := tx.ds.(*set.Set).SPop(count)
	tx.commit()
	n.notify(pb.NewOp(pb.OpType_SRem, key).Members(members))
	return members
}
