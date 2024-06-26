package set

import (
	"encoding/binary"
	"math/rand"
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
	n := 0
	for _, m := range member {
		_, updated := s.data.Set(m, struct{}{})
		if !updated {
			n++
		}
	}
	return int64(n)
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
		_, ok := s.data.Delete(m)
		if ok {
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

// SRandMember gets a random member from the set.
func (s *Set) SRandMember(count int64) []string {
	if count == 0 {
		return nil
	}
	var unique = true
	if count < 0 {
		unique = false
	}
	var kl = s.data.Len()
	if count > 0 && count > int64(kl) {
		count = int64(kl)
	}
	members := make([]string, 0)
	if unique {
		var keys = make(map[string]bool)
		s.data.Scan(func(key string, value struct{}) bool {
			keys[key] = true
			return true
		})
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
			key, _, ok := s.data.GetAt(index)
			if ok {
				members = append(members, key)
				count++
			}
		}
	}
	return members
}

// SClear clears the set.
func (s *Set) SClear() {
	s.data.Clear()
}

// Iter returns an iterator for the set.
func (s *Set) Iter(fn func(member string) bool) {
	s.data.Scan(func(member string, _ struct{}) bool {
		return fn(member)
	})
}

// Type returns the type of the data structure
func (s *Set) Type() ds.ValueType {
	return ds.Set
}

// GetValue the string to bytes
func (s *Set) GetValue() []byte {
	var members = make([]byte, 0, s.data.Len())
	s.data.Scan(func(member string, _ struct{}) bool {
		mLen := len(member)
		var b = make([]byte, mLen+1)
		n := binary.PutVarint(b, int64(mLen))
		copy(b[n:], member)
		members = append(members, b...)
		return true
	})
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
		s.data.Set(member, struct{}{})
	}
}
