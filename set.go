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
	meta := n.store.writeKey(key, n.newSet)
	v := meta.ds.(*set.Set).SAdd(members...)
	meta.commit()
	n.notify(pb.NewOp(pb.OpType_SAdd, key).Members(members))
	return v
}

// SCard gets the set members count.
func (n *Nodis) SCard(key string) int64 {
	meta := n.store.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	v := meta.ds.(*set.Set).SCard()
	meta.commit()
	return v
}

// SDiff gets the difference between sets.
func (n *Nodis) SDiff(keys ...string) []string {
	if len(keys) == 0 {
		return nil
	}
	meta := n.store.readKey(keys[0])
	if !meta.isOk() {
		meta.commit()
		return nil
	}
	lockedKeys := []*metadata{}
	otherSets := make([]*set.Set, len(keys)-1)
	for i, s := range keys[1:] {
		metaX := n.store.readKey(s)
		if !meta.isOk() {
			continue
		}
		lockedKeys = append(lockedKeys, metaX)
		otherSets[i] = metaX.ds.(*set.Set)
	}
	v := meta.ds.(*set.Set).SDiff(otherSets...)
	for _, s := range lockedKeys {
		s.commit()
	}
	meta.commit()
	return v
}

// SInter gets the intersection between sets.
func (n *Nodis) SInter(keys ...string) []string {
	if len(keys) == 0 {
		return nil
	}
	meta := n.store.readKey(keys[0])
	if !meta.isOk() {
		meta.commit()
		return nil
	}
	lockedSets := []*metadata{}
	otherSets := make([]*set.Set, len(keys)-1)
	for i, s := range keys[1:] {
		setDs := n.store.readKey(s)
		if !setDs.isOk() {
			continue
		}
		lockedSets = append(lockedSets, setDs)
		otherSets[i] = setDs.ds.(*set.Set)
	}
	v := meta.ds.(*set.Set).SInter(otherSets...)
	for _, s := range lockedSets {
		s.commit()
	}
	meta.commit()
	return v
}

// SIsMember returns if member is a member of the set stored at key.
func (n *Nodis) SIsMember(key, member string) bool {
	meta := n.store.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return false
	}
	v := meta.ds.(*set.Set).SIsMember(member)
	meta.commit()
	return v
}

// SMembers returns all the members of the set value stored at key.
func (n *Nodis) SMembers(key string) []string {
	meta := n.store.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return nil
	}
	v := meta.ds.(*set.Set).SMembers()
	meta.commit()
	return v
}

// SRem removes the specified members from the set stored at key.
func (n *Nodis) SRem(key string, members ...string) int64 {
	meta := n.store.writeKey(key, nil)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	v := meta.ds.(*set.Set).SRem(members...)
	meta.commit()
	n.notify(pb.NewOp(pb.OpType_SRem, key).Members(members))
	return v
}

// SScan scans the set value stored at key.
func (n *Nodis) SScan(key string, cursor int64, match string, count int64) (int64, []string) {
	meta := n.store.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return 0, nil
	}
	c, v := meta.ds.(*set.Set).SScan(cursor, match, count)
	meta.commit()
	return c, v
}

// SPop removes and returns a random element from the set value stored at key.
func (n *Nodis) SPop(key string, count int64) []string {
	meta := n.store.writeKey(key, nil)
	if !meta.isOk() {
		meta.commit()
		return nil
	}
	if count == 0 {
		count = 1
	}
	members := meta.ds.(*set.Set).SPop(count)
	meta.commit()
	n.notify(pb.NewOp(pb.OpType_SRem, key).Members(members))
	return members
}
