package set

import (
	"encoding/binary"
	"math/rand"
	"path/filepath"

	"github.com/diiyw/nodis/ds"
)

type Set struct {
	data map[string]struct{}
}

// NewSet creates a new set
func NewSet() *Set {
	return &Set{
		data: make(map[string]struct{}),
	}
}

// SAdd adds a member to the set
func (s *Set) SAdd(member ...string) int64 {
	return s.sAdd(member...)
}

func (s *Set) sAdd(member ...string) int64 {
	n := 0
	for _, m := range member {
		_, updated := s.data[m]
		s.data[m] = struct{}{}
		if !updated {
			n++
		}
	}
	return int64(n)
}

// SCard gets the set members count.
func (s *Set) SCard() int64 {
	return int64(len(s.data))
}

// SDiff gets the difference between sets.
func (s *Set) SDiff(sets ...*Set) []string {
	diff := make([]string, 0, 32)
	for member := range s.data {
		found := false
		for _, set := range sets {
			_, found = set.data[member]
			if found {
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
	for _, member := range diff {
		destination.sAdd(member)
	}
}

// SInter gets the intersection between sets.
func (s *Set) SInter(sets ...*Set) []string {
	inter := make([]string, 0, 32)
	for member := range s.data {
		found := true
		for _, set := range sets {
			_, found = set.data[member]
			if !found {
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
	for _, member := range inter {
		destination.sAdd(member)
	}
}

// SMembers gets the set members.
func (s *Set) SMembers() []string {
	members := make([]string, 0, len(s.data))
	for member := range s.data {
		members = append(members, member)
	}
	return members
}

// SIsMember checks if a member is in the set.
func (s *Set) SIsMember(member string) bool {
	_, ok := s.data[member]
	return ok
}

// SRem removes a member from the set.
func (s *Set) SRem(member ...string) int64 {
	var removed int64 = 0
	for _, m := range member {
		_, ok := s.data[m]
		if ok {
			delete(s.data, m)
			removed++
		}
	}
	return removed
}

// SPop removes and returns a random member from the set.
func (s *Set) SPop(count int64) []string {
	if count <= 0 {
		return nil
	}
	if count > int64(len(s.data)) {
		count = int64(len(s.data))
	}
	members := make([]string, 0, count)
	for member := range s.data {
		if count > 0 {
			delete(s.data, member)
			members = append(members, member)
			count--
		}
	}
	return members
}

// SUnion gets the union between sets.
func (s *Set) SUnion(sets ...*Set) []string {
	union := s.SMembers()
	for _, set := range sets {
		for member := range set.data {
			if _, ok := s.data[member]; !ok {
				union = append(union, member)
			}
		}
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
	if cursor >= int64(len(s.data)) {
		return 0, nil
	}
	var i int64 = 0
	for member := range s.data {
		if count > 0 && int64(len(keys)) >= count {
			break
		}
		if matched, err := filepath.Match(match, member); matched && err == nil {
			keys = append(keys, member)
		}
		i++
	}
	return cursor, keys
}

// SRandMember gets a random member from the set.
func (s *Set) SRandMember(count int64) []string {
	if count == 0 {
		return nil
	}
	var unique = true
	if count < 0 {
		unique = false
	}
	var kl = len(s.data)
	if count > 0 && count > int64(kl) {
		count = int64(kl)
	}
	members := make([]string, 0)
	if unique {
		var keys = make(map[string]bool)
		for key := range s.data {
			keys[key] = true
		}
		for m := range keys {
			if count == 0 {
				break
			}
			members = append(members, m)
			count--
		}
	} else {
		for count < 0 {
			index := rand.Intn(kl)
			var i int
			for key := range s.data {
				if i == index {
					members = append(members, key)
					break
				}
				i++
			}
			count++
		}
	}
	return members
}

// SClear clears the set.
func (s *Set) SClear() {
	s.data = make(map[string]struct{})
}

// Iter returns an iterator for the set.
func (s *Set) Iter(fn func(member string) bool) {
	for member := range s.data {
		if !fn(member) {
			break
		}
	}
}

// Type returns the type of the data structure
func (s *Set) Type() ds.ValueType {
	return ds.Set
}

// GetValue the string to bytes
func (s *Set) GetValue() []byte {
	var members = make([]byte, 0, len(s.data))
	for member := range s.data {
		mLen := len(member)
		var b = make([]byte, mLen+1)
		n := binary.PutVarint(b, int64(mLen))
		copy(b[n:], member)
		members = append(members, b...)
	}
	return members
}

// SetValue the bytes to string
func (s *Set) SetValue(members []byte) {
	for {
		if len(members) == 0 {
			break
		}
		mLen, n := binary.Varint(members)
		members = members[n:]
		member := string(members[:mLen])
		members = members[mLen:]
		s.data[member] = struct{}{}
	}
}
