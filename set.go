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
	s := n.getDs(key, n.newSet, 0)
	s.(*set.Set).SAdd(members...)
	return s.(*set.Set).SCard()
}

// SCard gets the set members count.
func (n *Nodis) SCard(key string) int {
	s := n.getDs(key, n.newSet, 0)
	return s.(*set.Set).SCard()
}

// SDiff gets the difference between sets.
func (n *Nodis) SDiff(key string, sets ...string) []string {
	s := n.getDs(key, n.newSet, 0)
	otherSets := make([]*set.Set, len(sets))
	for i, s := range sets {
		setDs := n.getDs(s, nil, 0)
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
	s := n.getDs(key, n.newSet, 0)
	otherSets := make([]*set.Set, len(sets))
	for i, s := range sets {
		setDs := n.getDs(s, nil, 0)
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
	s := n.getDs(key, n.newSet, 0)
	return s.(*set.Set).SIsMember(member)
}

// SMembers returns all the members of the set value stored at key.
func (n *Nodis) SMembers(key string) []string {
	s := n.getDs(key, n.newSet, 0)
	return s.(*set.Set).SMembers()
}
