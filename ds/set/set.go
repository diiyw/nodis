package set

import (
	"path/filepath"
	"sync"

	"github.com/diiyw/nodis/ds"
	"github.com/tidwall/btree"
)

type Set struct {
	sync.RWMutex
	data btree.Set[string]
}

// NewSet creates a new set
func NewSet() *Set {
	return &Set{}
}

// SAdd adds a member to the set
func (s *Set) SAdd(member ...string) int64 {
	s.Lock()
	defer s.Unlock()
	return s.sAdd(member...)
}

func (s *Set) sAdd(member ...string) int64 {
	for _, m := range member {
		s.data.Insert(m)
	}
	return int64(len(member))
}

// SCard gets the set members count.
func (s *Set) SCard() int64 {
	s.RLock()
	defer s.RUnlock()
	return int64(s.data.Len())
}

// SDiff gets the difference between sets.
func (s *Set) SDiff(sets ...*Set) []string {
	s.RLock()
	defer s.RUnlock()
	diff := make([]string, 0, 32)
	s.data.Scan(func(member string) bool {
		found := false
		for _, set := range sets {
			found = set.data.Contains(member)
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
	s.Lock()
	defer s.Unlock()
	for _, member := range diff {
		destination.sAdd(member)
	}
}

// SInter gets the intersection between sets.
func (s *Set) SInter(sets ...*Set) []string {
	s.RLock()
	defer s.RUnlock()
	inter := make([]string, 0, 32)
	s.data.Scan(func(member string) bool {
		found := true
		for _, set := range sets {
			found = set.data.Contains(member)
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
	s.Lock()
	defer s.Unlock()
	for _, member := range inter {
		destination.sAdd(member)
	}
}

// SMembers gets the set members.
func (s *Set) SMembers() []string {
	s.RLock()
	defer s.RUnlock()
	return s.data.Keys()
}

// SIsMember checks if a member is in the set.
func (s *Set) SIsMember(member string) bool {
	s.RLock()
	defer s.RUnlock()
	return s.data.Contains(member)
}

// SRem removes a member from the set.
func (s *Set) SRem(member ...string) int64 {
	s.Lock()
	defer s.Unlock()
	var removed int64 = 0
	for _, m := range member {
		s.data.Delete(m)
		removed++
	}
	return removed
}

// SPop removes and returns a random member from the set.
func (s *Set) SPop(count int64) []string {
	s.Lock()
	defer s.Unlock()
	if count <= 0 {
		return nil
	}
	if count > int64(s.data.Len()) {
		count = int64(s.data.Len())
	}
	members := make([]string, 0, count)
	s.data.Scan(func(member string) bool {
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
	s.RLock()
	defer s.RUnlock()
	union := s.data.Keys()
	for _, set := range sets {
		set.data.Scan(func(member string) bool {
			if !s.data.Contains(member) {
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
	s.Lock()
	defer s.Unlock()
	for _, member := range union {
		destination.sAdd(member)
	}
}

// SScan scans the set members.
func (s *Set) SScan(cursor int64, match string, count int64) (int64, []string) {
	s.RLock()
	defer s.RUnlock()
	keys := make([]string, 0, 32)
	if cursor >= int64(s.data.Len()) {
		return 0, nil
	}
	s.data.Scan(func(member string) bool {
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
		s.data.Insert(member)
	}
}
