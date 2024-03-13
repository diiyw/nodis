package set

import (
	"sync"

	"github.com/diiyw/nodis/ds"
	"github.com/dolthub/swiss"
	"github.com/kelindar/binary"
)

type Set struct {
	sync.RWMutex
	data *swiss.Map[string, struct{}]
}

// NewSet creates a new set
func NewSet() *Set {
	return &Set{
		data: swiss.NewMap[string, struct{}](32),
	}
}

// SAdd adds a member to the set
func (s *Set) SAdd(member ...string) {
	s.Lock()
	defer s.Unlock()
	s.sAdd(member...)
}

func (s *Set) sAdd(member ...string) {
	for _, m := range member {
		s.data.Put(m, struct{}{})
	}
}

// SCard gets the set members count.
func (s *Set) SCard() int {
	s.RLock()
	defer s.RUnlock()
	return s.data.Count()
}

// SDiff gets the difference between sets.
func (s *Set) SDiff(sets ...*Set) []string {
	s.RLock()
	defer s.RUnlock()
	diff := make([]string, 0, 32)
	s.data.Iter(func(member string, _ struct{}) bool {
		found := false
		for _, set := range sets {
			found = set.data.Has(member)
			if found {
				break
			}
		}
		if !found {
			diff = append(diff, member)
		}
		return false
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
	s.data.Iter(func(member string, _ struct{}) bool {
		found := true
		for _, set := range sets {
			found = set.data.Has(member)
			if !found {
				break
			}
		}
		if found {
			inter = append(inter, member)
		}
		return false
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
	var keys = make([]string, 0, s.data.Count())
	s.data.Iter(func(key string, _ struct{}) bool {
		keys = append(keys, key)
		return false
	})
	return keys
}

// SIsMember checks if a member is in the set.
func (s *Set) SIsMember(member string) bool {
	s.RLock()
	defer s.RUnlock()
	_, ok := s.data.Get(member)
	return ok
}

// SRem removes a member from the set.
func (s *Set) SRem(member string) {
	s.Lock()
	defer s.Unlock()
	s.data.Delete(member)
}

// SUnion gets the union between sets.
func (s *Set) SUnion(sets ...*Set) []string {
	s.RLock()
	defer s.RUnlock()
	union := make([]string, 0, 32)
	s.data.Iter(func(member string, _ struct{}) bool {
		union = append(union, member)
		return false
	})
	for _, set := range sets {
		set.data.Iter(func(member string, _ struct{}) bool {
			if !s.data.Has(member) {
				union = append(union, member)
			}
			return false
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

// GetType returns the type of the data structure
func (s *Set) GetType() ds.DataType {
	return ds.Set
}

// MarshalBinary the string to bytes
func (s *Set) MarshalBinary() ([]byte, error) {
	members := s.SMembers()
	return binary.Marshal(members)
}

// UnmarshalBinary the bytes to string
func (s *Set) UnmarshalBinary(data []byte) error {
	var members []string
	err := binary.Unmarshal(data, &members)
	if err != nil {
		return err
	}
	for _, member := range members {
		s.data.Put(member, struct{}{})
	}
	return nil
}
