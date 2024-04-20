package set

import (
	"path/filepath"

	"github.com/diiyw/nodis/ds"
	"github.com/tidwall/btree"
)

type Set struct {
	data btree.Map[string, struct{}]
}

// NewSet creates a new set
func NewSet() *Set {
	return &Set{}
}

// SAdd adds a member to the set
func (s *Set) SAdd(member ...string) int64 {
	return s.sAdd(member...)
}

func (s *Set) sAdd(member ...string) int64 {
	for _, m := range member {
		s.data.Set(m, struct{}{})
	}
	return int64(len(member))
}

// SCard gets the set members count.
func (s *Set) SCard() int64 {
	return int64(s.data.Len())
}

// SDiff gets the difference between sets.
func (s *Set) SDiff(sets ...*Set) []string {
	diff := make([]string, 0, 32)
	s.data.Scan(func(member string, _ struct{}) bool {
		found := false
		for _, set := range sets {
			_, found = set.data.Get(member)
			if found {
				break
			}
		}
		if !found {
			diff = append(diff, member)
		}
		return true
	})
	return diff
}

// SDiffStore gets the difference between sets and stores it in a new set.
func (s *Set) SDiffStore(destination *Set, sets ...*Set) {
	diff := s.SDiff(sets...)
	for _, member := range diff {
		destination.sAdd(member)
	}
}

// SInter gets the intersection between sets.
func (s *Set) SInter(sets ...*Set) []string {
	inter := make([]string, 0, 32)
	s.data.Scan(func(member string, _ struct{}) bool {
		found := true
		for _, set := range sets {
			_, found = set.data.Get(member)
			if !found {
				break
			}
		}
		if found {
			inter = append(inter, member)
		}
		return true
	})
	return inter
}

// SInterStore gets the intersection between sets and stores it in a new set.
func (s *Set) SInterStore(destination *Set, sets ...*Set) {
	inter := s.SInter(sets...)
	for _, member := range inter {
		destination.sAdd(member)
	}
}

// SMembers gets the set members.
func (s *Set) SMembers() []string {
	return s.data.Keys()
}

// SIsMember checks if a member is in the set.
func (s *Set) SIsMember(member string) bool {
	_, ok := s.data.Get(member)
	return ok
}

// SRem removes a member from the set.
func (s *Set) SRem(member ...string) int64 {
	var removed int64 = 0
	for _, m := range member {
		s.data.Delete(m)
		removed++
	}
	return removed
}

// SPop removes and returns a random member from the set.
func (s *Set) SPop(count int64) []string {
	if count <= 0 {
		return nil
	}
	if count > int64(s.data.Len()) {
		count = int64(s.data.Len())
	}
	members := make([]string, 0, count)
	s.data.Scan(func(member string, _ struct{}) bool {
		if count > 0 {
			s.data.Delete(member)
			members = append(members, member)
			count--
		}
		return true
	})
	return members
}

// SUnion gets the union between sets.
func (s *Set) SUnion(sets ...*Set) []string {
	union := s.data.Keys()
	for _, set := range sets {
		set.data.Scan(func(member string, _ struct{}) bool {
			if _, ok := s.data.Get(member); !ok {
				union = append(union, member)
			}
			return true
		})
	}
	return union
}

// SUnionStore gets the union between sets and stores it in a new set.
func (s *Set) SUnionStore(destination *Set, sets ...*Set) {
	union := s.SUnion(sets...)
	for _, member := range union {
		destination.sAdd(member)
	}
}

// SScan scans the set members.
func (s *Set) SScan(cursor int64, match string, count int64) (int64, []string) {
	keys := make([]string, 0, 32)
	if cursor >= int64(s.data.Len()) {
		return 0, nil
	}
	s.data.Scan(func(member string, _ struct{}) bool {
		if count > 0 && int64(len(keys)) >= count {
			return false
		}
		if matched, err := filepath.Match(match, member); matched && err == nil {
			keys = append(keys, member)
		}
		return true
	})
	return cursor, keys
}

// Type returns the type of the data structure
func (s *Set) Type() ds.DataType {
	return ds.Set
}

// GetValue the string to bytes
func (s *Set) GetValue() []string {
	return s.SMembers()
}

// SetValue the bytes to string
func (s *Set) SetValue(members []string) {
	for _, member := range members {
		s.data.Set(member, struct{}{})
	}
}
