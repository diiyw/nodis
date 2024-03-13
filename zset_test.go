package nodis

import (
	"os"
	"testing"
)

func TestZSet_ZAdd(t *testing.T) {
	opt := DefaultOptions
	opt.Path = "testdata"
	os.RemoveAll("testdata")
	n := Open(opt)
	defer n.Close()
	n.ZAdd("zset", "a", 1)
	n.ZAdd("zset", "b", 2)
	n.ZAdd("zset", "c", 3)
	if n.ZCard("zset") != 3 {
		t.Errorf("ZCard() = %v, want %v", n.ZCard("zset"), 3)
	}
}

func TestZSet_ZRange(t *testing.T) {
	opt := DefaultOptions
	opt.Path = "testdata"
	os.RemoveAll("testdata")
	n := Open(opt)
	defer n.Close()
	n.ZAdd("zset", "a", 1)
	n.ZAdd("zset", "b", 2)
	n.ZAdd("zset", "c", 3)
	range1 := n.ZRange("zset", 0, 1)
	if len(range1) != 1 {
		t.Errorf("ZRange() = %v, want %v", len(range1), 1)
	}
	if range1[0] != "a" {
		t.Errorf("ZRange() = %v, want %v", range1[0], "a")
	}
	range2 := n.ZRevRange("zset", 0, 1)
	if len(range2) != 1 {
		t.Errorf("ZRevRange() = %v, want %v", len(range2), 1)
	}
	if range2[0] != "c" {
		t.Errorf("ZRevRange() = %v, want %v", range2[0], "c")
	}
	range3 := n.ZRevRange("zset", 1, 2)
	if len(range3) != 2 {
		t.Errorf("ZRevRange() = %v, want %v", len(range3), 2)
	}
	if range3[0] != "c" {
		t.Errorf("ZRevRange() = %v, want %v", range3[0], "c")
	}
	if range3[1] != "b" {
		t.Errorf("ZRevRange() = %v, want %v", range3[1], "b")
	}
}

func TestZSet_ZRangeByScore(t *testing.T) {
	opt := DefaultOptions
	opt.Path = "testdata"
	os.RemoveAll("testdata")
	n := Open(opt)
	defer n.Close()
	n.ZAdd("zset", "a", 1)
	n.ZAdd("zset", "b", 2)
	n.ZAdd("zset", "c", 3)
	range1 := n.ZRangeByScore("zset", 1, 2)
	if len(range1) != 2 {
		t.Errorf("ZRangeByScore() = %v, want %v", len(range1), 2)
	}
	if range1[0] != "a" {
		t.Errorf("ZRangeByScore() = %v, want %v", range1[0], "a")
	}
	if range1[1] != "b" {
		t.Errorf("ZRangeByScore() = %v, want %v", range1[1], "b")
	}
}

func TestZSet_ZRank(t *testing.T) {
	opt := DefaultOptions
	opt.Path = "testdata"
	os.RemoveAll("testdata")
	n := Open(opt)
	defer n.Close()
	n.ZAdd("zset", "a", 1)
	n.ZAdd("zset", "b", 2)
	n.ZAdd("zset", "c", 3)
	if n.ZRank("zset", "a") != 1 {
		t.Errorf("ZRank() = %v, want %v", n.ZRank("zset", "a"), 1)
	}
	if n.ZRank("zset", "b") != 2 {
		t.Errorf("ZRank() = %v, want %v", n.ZRank("zset", "b"), 2)
	}
	if n.ZRank("zset", "c") != 3 {
		t.Errorf("ZRank() = %v, want %v", n.ZRank("zset", "c"), 3)
	}
}

func TestZSet_ZRevRange(t *testing.T) {
	opt := DefaultOptions
	opt.Path = "testdata"
	os.RemoveAll("testdata")
	n := Open(opt)
	defer n.Close()
	n.ZAdd("zset", "a", 1)
	n.ZAdd("zset", "b", 2)
	n.ZAdd("zset", "c", 3)
	range1 := n.ZRevRange("zset", 0, 2)
	if len(range1) != 2 {
		t.Errorf("ZRevRange() = %v, want %v", len(range1), 2)
	}
	if range1[0] != "c" {
		t.Errorf("ZRevRange() = %v, want %v", range1[0], "c")
	}
	if range1[1] != "b" {
		t.Errorf("ZRevRange() = %v, want %v", range1[1], "b")
	}
}

func TestZSet_ZRevRangeByScore(t *testing.T) {
	opt := DefaultOptions
	opt.Path = "testdata"
	os.RemoveAll("testdata")
	n := Open(opt)
	defer n.Close()
	n.ZAdd("zset", "a", 1)
	n.ZAdd("zset", "b", 2)
	n.ZAdd("zset", "c", 3)
	range1 := n.ZRevRangeByScore("zset", 1, 2)
	if len(range1) != 2 {
		t.Errorf("ZRevRangeByScore() = %v, want %v", len(range1), 2)
	}
	if range1[0] != "b" {
		t.Errorf("ZRevRangeByScore() = %v, want %v", range1[0], "b")
	}
	if range1[1] != "a" {
		t.Errorf("ZRevRangeByScore() = %v, want %v", range1[1], "a")
	}
}

func TestZSet_ZRevRank(t *testing.T) {
	opt := DefaultOptions
	opt.Path = "testdata"
	os.RemoveAll("testdata")
	n := Open(opt)
	defer n.Close()
	n.ZAdd("zset", "a", 1)
	n.ZAdd("zset", "b", 2)
	n.ZAdd("zset", "c", 3)
	if n.ZRevRank("zset", "a") != 3 {
		t.Errorf("ZRevRank() = %v, want %v", n.ZRevRank("zset", "a"), 3)
	}
	if n.ZRevRank("zset", "b") != 2 {
		t.Errorf("ZRevRank() = %v, want %v", n.ZRevRank("zset", "b"), 2)
	}
	if n.ZRevRank("zset", "c") != 1 {
		t.Errorf("ZRevRank() = %v, want %v", n.ZRevRank("zset", "c"), 1)
	}
}

func TestZSet_ZScore(t *testing.T) {
	opt := DefaultOptions
	opt.Path = "testdata"
	os.RemoveAll("testdata")
	n := Open(opt)
	defer n.Close()
	n.ZAdd("zset", "a", 1)
	n.ZAdd("zset", "b", 2)
	n.ZAdd("zset", "c", 3)
	if n.ZScore("zset", "a") != 1 {
		t.Errorf("ZScore() = %v, want %v", n.ZScore("zset", "a"), 1)
	}
	if n.ZScore("zset", "b") != 2 {
		t.Errorf("ZScore() = %v, want %v", n.ZScore("zset", "b"), 2)
	}
	if n.ZScore("zset", "c") != 3 {
		t.Errorf("ZScore() = %v, want %v", n.ZScore("zset", "c"), 3)
	}
}

func TestZSet_ZIncrBy(t *testing.T) {
	opt := DefaultOptions
	opt.Path = "testdata"
	os.RemoveAll("testdata")
	n := Open(opt)
	defer n.Close()
	n.ZAdd("zset", "a", 1)
	n.ZIncrBy("zset", "a", 1)
	if n.ZScore("zset", "a") != 2 {
		t.Errorf("ZIncrBy() = %v, want %v", n.ZScore("zset", "a"), 2)
	}
}

func TestZSet_ZRem(t *testing.T) {
	opt := DefaultOptions
	opt.Path = "testdata"
	os.RemoveAll("testdata")
	n := Open(opt)
	defer n.Close()
	n.ZAdd("zset", "a", 1)
	n.ZAdd("zset", "b", 2)
	n.ZAdd("zset", "c", 3)
	n.ZRem("zset", "a")
	if n.ZCard("zset") != 2 {
		t.Errorf("ZCard() = %v, want %v", n.ZCard("zset"), 2)
	}
}

func TestZSet_ZRemRangeByRank(t *testing.T) {
	opt := DefaultOptions
	opt.Path = "testdata"
	os.RemoveAll("testdata")
	n := Open(opt)
	defer n.Close()
	n.ZAdd("zset", "a", 1)
	n.ZAdd("zset", "b", 2)
	n.ZAdd("zset", "c", 3)
	n.ZRemRangeByRank("zset", 0, 1)
	if n.ZCard("zset") != 2 {
		t.Errorf("ZCard() = %v, want %v", n.ZCard("zset"), 2)
	}
}

func TestZSet_ZRemRangeByScore(t *testing.T) {
	opt := DefaultOptions
	opt.Path = "testdata"
	os.RemoveAll("testdata")
	n := Open(opt)
	defer n.Close()
	n.ZAdd("zset", "a", 1)
	n.ZAdd("zset", "b", 2)
	n.ZAdd("zset", "c", 3)
	n.ZRemRangeByScore("zset", 1, 2)
	if n.ZCard("zset") != 1 {
		t.Errorf("ZCard() = %v, want %v", n.ZCard("zset"), 1)
	}
}

func TestZSet_ZClear(t *testing.T) {
	opt := DefaultOptions
	opt.Path = "testdata"
	os.RemoveAll("testdata")
	n := Open(opt)
	defer n.Close()
	n.ZAdd("zset", "a", 1)
	n.ZAdd("zset", "b", 2)
	n.ZAdd("zset", "c", 3)
	n.ZClear("zset")
	if n.ZCard("zset") != 0 {
		t.Errorf("ZCard() = %v, want %v", n.ZCard("zset"), 0)
	}
}

func TestZSet_ZExists(t *testing.T) {
	opt := DefaultOptions
	opt.Path = "testdata"
	os.RemoveAll("testdata")
	n := Open(opt)
	defer n.Close()
	n.ZAdd("zset", "a", 1)
	if !n.ZExists("zset", "a") {
		t.Errorf("ZExists() = %v, want %v", n.ZExists("zset", "a"), true)
	}
}

func TestZSet_ZRangeWithScores(t *testing.T) {
	opt := DefaultOptions
	opt.Path = "testdata"
	os.RemoveAll("testdata")
	n := Open(opt)
	defer n.Close()
	n.ZAdd("zset", "a", 1)
	n.ZAdd("zset", "b", 2)
	n.ZAdd("zset", "c", 3)
	range1 := n.ZRangeWithScores("zset", 1, 2)
	if len(range1) != 2 {
		t.Errorf("ZRangeWithScores() = %v, want %v", len(range1), 2)
	}
	if range1[0].Member != "a" {
		t.Errorf("ZRangeWithScores() = %v, want %v", range1[0].Member, "a")
	}
	if range1[1].Member != "b" {
		t.Errorf("ZRangeWithScores() = %v, want %v", range1[1].Member, "b")
	}
}

func TestZSet_ZRevRangeWithScores(t *testing.T) {
	opt := DefaultOptions
	opt.Path = "testdata"
	os.RemoveAll("testdata")
	n := Open(opt)
	defer n.Close()
	n.ZAdd("zset", "a", 1)
	n.ZAdd("zset", "b", 2)
	n.ZAdd("zset", "c", 3)
	range1 := n.ZRevRangeWithScores("zset", 1, 2)
	if len(range1) != 2 {
		t.Errorf("ZRevRangeWithScores() = %v, want %v", len(range1), 2)
	}
	if range1[0].Member != "c" {
		t.Errorf("ZRevRangeWithScores() = %v, want %v", range1[0].Member, "c")
	}
	if range1[1].Member != "b" {
		t.Errorf("ZRevRangeWithScores() = %v, want %v", range1[1].Member, "b")
	}
}
