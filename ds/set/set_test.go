package set

import (
	"fmt"
	"testing"
)

func TestSet_SAdd(t *testing.T) {
	s := NewSet()
	s.SAdd("hello")
	if !s.SIsMember("hello") {
		t.Errorf("SAdd failed")
	}
}

func TestSet_SRem(t *testing.T) {
	s := NewSet()
	s.SAdd("hello")
	s.SRem("hello")
	if s.SIsMember("hello") {
		t.Errorf("SRem failed")
	}
}

func TestSet_SMembers(t *testing.T) {
	s := NewSet()
	s.SAdd("hello")
	s.SAdd("world")
	m := s.SMembers()
	if len(m) != 2 {
		t.Errorf("SMembers failed")
	}
}

func TestSet_SCard(t *testing.T) {
	s := NewSet()
	s.SAdd("hello")
	s.SAdd("world")
	if s.SCard() != 2 {
		t.Errorf("SCard failed")
	}
}

func TestSet_SDiff(t *testing.T) {
	s1 := NewSet()
	s1.SAdd("hello")
	s1.SAdd("world")
	s2 := NewSet()
	s2.SAdd("world")
	diff := s1.SDiff(s2)
	if len(diff) != 1 {
		t.Errorf("SDiff failed")
	}
}

func TestSet_SDiffStore(t *testing.T) {
	s1 := NewSet()
	s1.SAdd("hello")
	s1.SAdd("world")
	s2 := NewSet()
	s2.SAdd("world")
	s3 := NewSet()
	s1.SDiffStore(s3, s2)
	if !s3.SIsMember("hello") {
		t.Errorf("SDiffStore failed")
	}
}

func TestSet_SInter(t *testing.T) {
	s1 := NewSet()
	s1.SAdd("hello")
	s1.SAdd("world")
	s2 := NewSet()
	s2.SAdd("world")
	inter := s1.SInter(s2)
	if len(inter) != 1 {
		t.Errorf("SInter failed excepted 1 got %d", len(inter))
	}
}

func TestSet_SIsMember(t *testing.T) {
	s := NewSet()
	s.SAdd("hello")
	if !s.SIsMember("hello") {
		t.Errorf("SIsMember failed")
	}
}

func TestSet_SInterStore(t *testing.T) {
	s1 := NewSet()
	s1.SAdd("hello")
	s1.SAdd("world")
	s2 := NewSet()
	s2.SAdd("world")
	s3 := NewSet()
	s1.SInterStore(s3, s2)
	if !s3.SIsMember("world") {
		t.Errorf("SInterStore failed")
	}
}

func TestSet_SUnion(t *testing.T) {
	s1 := NewSet()
	s1.SAdd("hello")
	s1.SAdd("world")
	s2 := NewSet()
	s2.SAdd("world")
	union := s1.SUnion(s2)
	if len(union) != 2 {
		t.Errorf("SUnion failed")
	}
}

func TestSet_SUnionStore(t *testing.T) {
	s1 := NewSet()
	s1.SAdd("hello")
	s1.SAdd("world")
	s2 := NewSet()
	s2.SAdd("world")
	s3 := NewSet()
	s1.SUnionStore(s3, s2)
	if !s3.SIsMember("world") {
		t.Errorf("SUnionStore failed")
	}
}

func TestGetSetValue(t *testing.T) {
	s := NewSet()
	s.SAdd("hello")
	s.SAdd("world")
	s.SAdd("worldxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
	s.SAdd("helloxx")
	v := s.GetValue()
	s2 := NewSet()
	s2.SetValue(v)
	if s.SCard() != s2.SCard() {
		t.Errorf("GetValue failed")
	}
	for _, s3 := range s2.SMembers() {
		fmt.Println(s3)
	}
}
