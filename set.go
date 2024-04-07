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
	k, s := n.getDs(key, n.newSet, 0)
	k.changed.Store(true)
	n.notify(pb.NewOp(pb.OpType_SAdd, key).Members(members))
	return s.(*set.Set).SAdd(members...)
}

// SCard gets the set members count.
func (n *Nodis) SCard(key string) int64 {
	_, s := n.getDs(key, nil, 0)
	return s.(*set.Set).SCard()
}

// SDiff gets the difference between sets.
func (n *Nodis) SDiff(keys ...string) []string {
	if len(keys) == 0 {
		return nil
	}
	_, s := n.getDs(keys[0], nil, 0)
	if s == nil {
		return nil
	}
	otherSets := make([]*set.Set, len(keys)-1)
	for i, s := range keys[1:] {
		_, setDs := n.getDs(s, nil, 0)
		if setDs == nil {
			continue
		}
		otherSets[i] = setDs.(*set.Set)
	}
	return s.(*set.Set).SDiff(otherSets...)
}

// SInter gets the intersection between sets.
func (n *Nodis) SInter(keys ...string) []string {
	if len(keys) == 0 {
		return nil
	}
	_, s := n.getDs(keys[0], nil, 0)
	if s == nil {
		return nil
	}
	otherSets := make([]*set.Set, len(keys)-1)
	for i, s := range keys[1:] {
		_, setDs := n.getDs(s, nil, 0)
		if setDs == nil {
			continue
		}
		otherSets[i] = setDs.(*set.Set)
	}
	return s.(*set.Set).SInter(otherSets...)
}

// SIsMember returns if member is a member of the set stored at key.
func (n *Nodis) SIsMember(key, member string) bool {
	_, s := n.getDs(key, nil, 0)
	if s == nil {
		return false
	}
	return s.(*set.Set).SIsMember(member)
}

// SMembers returns all the members of the set value stored at key.
func (n *Nodis) SMembers(key string) []string {
	_, s := n.getDs(key, nil, 0)
	if s == nil {
		return nil
	}
	return s.(*set.Set).SMembers()
}

// SRem removes the specified members from the set stored at key.
func (n *Nodis) SRem(key string, members ...string) int64 {
	k, s := n.getDs(key, nil, 0)
	if s == nil {
		return 0
	}
	k.changed.Store(true)
	n.notify(pb.NewOp(pb.OpType_SRem, key).Members(members))
	return s.(*set.Set).SRem(members...)
}

// SScan scans the set value stored at key.
func (n *Nodis) SScan(key string, cursor int64, match string, count int64) (int64, []string) {
	_, s := n.getDs(key, nil, 0)
	if s == nil {
		return 0, nil
	}
	return s.(*set.Set).SScan(cursor, match, count)
}

// SPop removes and returns a random element from the set value stored at key.
func (n *Nodis) SPop(key string, count int64) []string {
	k, s := n.getDs(key, nil, 0)
	if s == nil {
		return nil
	}
	if count == 0 {
		count = 1
	}
	members := s.(*set.Set).SPop(count)
	k.changed.Store(true)
	n.notify(pb.NewOp(pb.OpType_SRem, key).Members(members))
	return members
}
