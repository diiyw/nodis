package nodis

import (
	"os"
	"testing"
)

func TestSet_SAdd(t *testing.T) {
	opt := DefaultOptions
	opt.Path = "testdata"
	os.RemoveAll("testdata")
	n := Open(opt)
	defer n.Close()
	n.SAdd("set", "a", "b", "c")
	if n.SCard("set") != 3 {
		t.Errorf("SCard() = %v, want %v", n.SCard("set"), 3)
	}
}

func TestSet_SDiff(t *testing.T) {
	opt := DefaultOptions
	opt.Path = "testdata"
	os.RemoveAll("testdata")
	n := Open(opt)
	defer n.Close()
	n.SAdd("set1", "a", "b", "c")
	n.SAdd("set2", "b", "c", "d")
	diff := n.SDiff("set1", "set2")
	if len(diff) != 1 {
		t.Errorf("SDiff() = %v, want %v", len(diff), 1)
	}
	if diff[0] != "a" {
		t.Errorf("SDiff() = %v, want %v", diff[0], "a")
	}
}

func TestSet_SInter(t *testing.T) {
	opt := DefaultOptions
	opt.Path = "testdata"
	os.RemoveAll("testdata")
	n := Open(opt)
	defer n.Close()
	n.SAdd("set1", "a", "b", "c")
	n.SAdd("set2", "b", "c", "d")
	inter := n.SInter("set1", "set2")
	if len(inter) != 2 {
		t.Errorf("SInter() = %v, want %v", len(inter), 2)
	}
	if inter[0] != "b" {
		t.Errorf("SInter() = %v, want %v", inter[0], "b")
	}
	if inter[1] != "c" {
		t.Errorf("SInter() = %v, want %v", inter[1], "c")
	}
}

func TestSet_SIsMember(t *testing.T) {
	opt := DefaultOptions
	opt.Path = "testdata"
	os.RemoveAll("testdata")
	n := Open(opt)
	defer n.Close()
	n.SAdd("set", "a", "b", "c")
	if !n.SIsMember("set", "a") {
		t.Errorf("SIsMember() = %v, want %v", n.SIsMember("set", "a"), true)
	}
}

func TestSet_SMembers(t *testing.T) {
	opt := DefaultOptions
	opt.Path = "testdata"
	os.RemoveAll("testdata")
	n := Open(opt)
	defer n.Close()
	n.SAdd("set", "a", "b", "c")
	members := n.SMembers("set")
	if len(members) != 3 {
		t.Errorf("SMembers() = %v, want %v", len(members), 3)
	}
}
