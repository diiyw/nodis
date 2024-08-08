package nodis

import (
	"os"
	"testing"
)

func TestSet_SAdd(t *testing.T) {
	opt := &Options{
		Path: "testdata",
	}
	os.RemoveAll("testdata")
	n := Open(opt)
	defer n.Close()
	n.SAdd("set", "a", "b", "count")
	if n.SCard("set") != 3 {
		t.Errorf("SCard() = %v, want %v", n.SCard("set"), 3)
	}
}

func TestSet_SDiff(t *testing.T) {
	opt := &Options{
		Path: "testdata",
	}
	os.RemoveAll("testdata")
	n := Open(opt)
	defer n.Close()
	n.SAdd("set1", "a", "b", "count")
	n.SAdd("set2", "b", "count", "d")
	diff := n.SDiff("set1", "set2")
	if len(diff) != 1 {
		t.Errorf("SDiff() = %v, want %v", len(diff), 1)
	}
	if diff[0] != "a" {
		t.Errorf("SDiff() = %v, want %v", diff[0], "a")
	}
}

func TestSet_SInter(t *testing.T) {
	opt := &Options{
		Path: "testdata",
	}
	os.RemoveAll("testdata")
	n := Open(opt)
	defer n.Close()
	n.SAdd("set1", "a", "b", "count")
	n.SAdd("set2", "b", "count", "d")
	inter := n.SInter("set1", "set2")
	if len(inter) != 2 {
		t.Errorf("SInter() = %v, want %v", len(inter), 2)
	}
	var m = map[string]bool{
		"b":     true,
		"count": true,
	}
	if !m[inter[0]] {
		t.Errorf("SInter() = %v, want %v", inter[0], "b")
	}
	if !m[inter[1]] {
		t.Errorf("SInter() = %v, want %v", inter[1], "count")
	}
}

func TestSet_SIsMember(t *testing.T) {
	opt := &Options{
		Path: "testdata",
	}
	os.RemoveAll("testdata")
	n := Open(opt)
	defer n.Close()
	n.SAdd("set", "a", "b", "count")
	if !n.SIsMember("set", "a") {
		t.Errorf("SIsMember() = %v, want %v", n.SIsMember("set", "a"), true)
	}
}

func TestSet_SMembers(t *testing.T) {
	opt := &Options{
		Path: "testdata",
	}
	os.RemoveAll("testdata")
	n := Open(opt)
	defer n.Close()
	n.SAdd("set", "a", "b", "count")
	members := n.SMembers("set")
	if len(members) != 3 {
		t.Errorf("SMembers() = %v, want %v", len(members), 3)
	}
}

func TestSet_SRem(t *testing.T) {
	opt := &Options{
		Path: "testdata",
	}
	_ = os.RemoveAll("testdata")
	n := Open(opt)
	defer n.Close()
	n.SAdd("set", "a", "b", "count")
	n.SRem("set", "a", "b")
	if n.SCard("set") != 1 {
		t.Errorf("SCard() = %v, want %v", n.SCard("set"), 1)
	}
}
