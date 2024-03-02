package set

import (
	"github.com/dolthub/swiss"
	"github.com/kelindar/binary"
	"sync"
)

type Set struct {
	sync.RWMutex
	data    *swiss.Map[string, int]
	members []string
}

// NewSet creates a new set
func NewSet() *Set {
	return &Set{
		data:    swiss.NewMap[string, int](32),
		members: make([]string, 0, 32),
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
		if _, ok := s.data.Get(m); !ok {
			s.members = append(s.members, m)
			s.data.Put(m, len(s.members)-1)
		}
	}
}

// SCard gets the set members count.
func (s *Set) SCard() int {
	s.RLock()
	defer s.RUnlock()
	return len(s.members)
}

// SDiff gets the difference between sets.
func (s *Set) SDiff(sets ...*Set) []string {
	s.RLock()
	defer s.RUnlock()
	diff := make([]string, 0, 32)
	for _, member := range s.members {
		found := false
		for _, set := range sets {
			if _, ok := set.data.Get(member); ok {
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, member)
		}
	}
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
	for _, member := range s.members {
		found := true
		for _, set := range sets {
			if _, ok := set.data.Get(member); !ok {
				found = false
				break
			}
		}
		if found {
			inter = append(inter, member)
		}
	}
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
	return s.members
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
	if index, ok := s.data.Get(member); ok {
		s.data.Delete(member)
		s.members = append(s.members[:index], s.members[index+1:]...)
	}
}

// SUnion gets the union between sets.
func (s *Set) SUnion(sets ...*Set) []string {
	s.RLock()
	defer s.RUnlock()
	union := make([]string, 0, 32)
	union = append(union, s.members...)
	for _, set := range sets {
		for _, member := range set.members {
			if _, ok := s.data.Get(member); !ok {
				union = append(union, member)
			}
		}
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
func (s *Set) GetType() string {
	return "set"
}

// Marshal the string to bytes
func (s *Set) Marshal() ([]byte, error) {
	return binary.Marshal(s.members)
}

// Unmarshal the bytes to string
func (s *Set) Unmarshal(data []byte) error {
	err := binary.Unmarshal(data, &s.members)
	if err != nil {
		return err
	}
	for i, member := range s.members {
		s.data.Put(member, i)
	}
	return nil
}
