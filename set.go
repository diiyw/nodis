package nodis

import (
	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/ds/set"
)

func (n *Nodis) newSet() ds.DataStruct {
	return set.NewSet()
}

// SAdd adds the specified members to the set stored at key.
func (n *Nodis) SAdd(key string, members ...string) int {
	k, s := n.getDs(key, n.newSet, 0)
	k.changed.Store(true)
	return s.(*set.Set).SAdd(members...)
}

// SCard gets the set members count.
func (n *Nodis) SCard(key string) int {
	_, s := n.getDs(key, nil, 0)
	return s.(*set.Set).SCard()
}

// SDiff gets the difference between sets.
func (n *Nodis) SDiff(key string, sets ...string) []string {
	_, s := n.getDs(key, n.newSet, 0)
	if s == nil {
		return nil
	}
	otherSets := make([]*set.Set, len(sets))
	for i, s := range sets {
		_, setDs := n.getDs(s, nil, 0)
		if setDs == nil {
			continue
		}
		setDs.RLock()
		otherSets[i] = setDs.(*set.Set)
	}
	defer func() {
		for _, otherSet := range otherSets {
			if otherSet != nil {
				otherSet.RUnlock()
			}
		}
	}()
	return s.(*set.Set).SDiff(otherSets...)
}

// SInter gets the intersection between sets.
func (n *Nodis) SInter(key string, sets ...string) []string {
	_, s := n.getDs(key, nil, 0)
	if s == nil {
		return nil
	}
	otherSets := make([]*set.Set, len(sets))
	for i, s := range sets {
		_, setDs := n.getDs(s, nil, 0)
		if setDs == nil {
			continue
		}
		setDs.RLock()
		otherSets[i] = setDs.(*set.Set)
	}
	defer func() {
		for _, otherSet := range otherSets {
			if otherSet != nil {
				otherSet.RUnlock()
			}
		}
	}()
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
func (n *Nodis) SRem(key string, members ...string) int {
	k, s := n.getDs(key, nil, 0)
	if s == nil {
		return 0
	}
	k.changed.Store(true)
	return s.(*set.Set).SRem(members...)
}
