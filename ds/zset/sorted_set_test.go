package zset

import (
	"fmt"
	"math"
	"testing"
)

func TestSortedSet_ZAdd(t *testing.T) {
	ss := NewSortedSet()
	ss.ZAdd("member1", 1.5)
	ss.ZAdd("member2", 2.5)
	ss.ZAdd("member3", 0.5)
	ss.ZAdd("member3", 0.6)

	// Test if the length is correct
	if ss.ZCard() != 3 {
		t.Errorf("Length error")
	}

	// Test if the elements are inserted correctly
	if !ss.ZExists("member1") || !ss.ZExists("member2") || !ss.ZExists("member3") {
		t.Errorf("Insertion error")
	}

	// Test if the scores are correct
	v1, _ := ss.ZScore("member1")
	v2, _ := ss.ZScore("member2")
	v3, _ := ss.ZScore("member3")
	if v1 != 1.5 || v2 != 2.5 || v3 != 0.6 {
		t.Errorf("Score error")
	}
}

func TestSortedSet_ZCard(t *testing.T) {
	ss := NewSortedSet()
	ss.ZAdd("member1", 1.5)
	ss.ZAdd("member2", 2.5)
	ss.ZAdd("member3", 0.5)

	if ss.ZCard() != 3 {
		t.Errorf("Length error")
	}
}

func TestSortedSet_ZExists(t *testing.T) {
	ss := NewSortedSet()
	ss.ZAdd("member1", 1.5)
	ss.ZAdd("member2", 2.5)
	ss.ZAdd("member3", 0.5)

	if !ss.ZExists("member1") || !ss.ZExists("member2") || !ss.ZExists("member3") {
		t.Errorf("Insertion error")
	}
}

func TestSortedSet_ZScore(t *testing.T) {
	ss := NewSortedSet()
	ss.ZAdd("member1", 1.5)
	ss.ZAdd("member2", 2.5)
	ss.ZAdd("member3", 0.5)

	v1, _ := ss.ZScore("member1")
	v2, _ := ss.ZScore("member2")
	v3, _ := ss.ZScore("member3")
	if v1 != 1.5 {
		t.Errorf("Score error expected 1.5 got %f", v1)
	}
	if v2 != 2.5 {
		t.Errorf("Score error expected 2.5 got %f", v2)
	}
	if v3 != 0.5 {
		t.Errorf("Score error expected 0.5 got %f", v3)
	}
}

func TestSortedSet_ZRange(t *testing.T) {
	ss := NewSortedSet()
	ss.ZAdd("member1", 1.5)
	ss.ZAdd("member2", 2.5)
	ss.ZAdd("member3", 0.5)

	// Test if the range is correct
	if len(ss.ZRange(0, 2)) != 3 {
		t.Errorf("Range error expected 2 got %d", len(ss.ZRange(0, 2)))
	}
}

func TestSortedSet_ZRangeByScore(t *testing.T) {
	ss := NewSortedSet()
	ss.ZAdd("member1", 1.5)
	ss.ZAdd("member2", 2.5)
	ss.ZAdd("member3", 0.5)

	// Test if the range is correct
	members := ss.ZRangeByScore(0.5, 2.5, 0, -1, 0)
	if len(members) != 3 {
		t.Errorf("Range error expected 3 got %d", len(members))
	}
}

func TestSortedSet_ZRank(t *testing.T) {
	ss := NewSortedSet()
	ss.ZAdd("member1", 1.5)
	ss.ZAdd("member2", 2.5)
	ss.ZAdd("member3", 0.5)
	v1, _ := ss.ZRank("member1")
	v2, _ := ss.ZRank("member2")
	v3, _ := ss.ZRank("member3")
	// Test if the rank is correct
	if v1 != 1 {
		t.Errorf("Rank error expected 1 got %d", v1)
	}
	if v2 != 2 {
		t.Errorf("Rank error expected 2 got %d", v2)
	}
	if v3 != 0 {
		t.Errorf("Rank error expected 0 got %d", v3)
	}
}

func TestSortedSet_ZRem(t *testing.T) {
	ss := NewSortedSet()
	ss.ZAdd("member1", 1.5)
	ss.ZAdd("member2", 2.5)
	ss.ZAdd("member3", 0.5)

	// Test if the length is correct
	if ss.ZCard() != 3 {
		t.Errorf("Length error")
	}

	// Test if the elements are removed correctly
	ss.ZRem("member1")
	if ss.ZCard() != 2 || ss.ZExists("member1") {
		t.Errorf("Removal error")
	}
}

func TestSortedSet_ZRemRangeByRank(t *testing.T) {
	ss := NewSortedSet()
	ss.ZAdd("member1", 1.5)
	ss.ZAdd("member2", 2.5)
	ss.ZAdd("member3", 0.5)

	// Test if the length is correct
	if ss.ZCard() != 3 {
		t.Errorf("Length error")
	}

	// Test if the elements are removed correctly
	ss.ZRemRangeByRank(0, 1)
	if ss.ZCard() != 1 {
		t.Errorf("Removal error expected 2 got %d", ss.ZCard())
	}
}

func TestSortedSet_ZRemRangeByScore(t *testing.T) {
	ss := NewSortedSet()
	ss.ZAdd("member1", 1.5)
	ss.ZAdd("member2", 2.5)
	ss.ZAdd("member3", 0.5)

	// Test if the length is correct
	if ss.ZCard() != 3 {
		t.Errorf("Length error expected 3 got %d", ss.ZCard())
	}

	// Test if the elements are removed correctly
	removed := ss.ZRemRangeByScore(1, 2, 0)
	if ss.ZCard() != 2 {
		t.Errorf("Removal error expected 2 got %d removed %d", ss.ZCard(), removed)
	}
}

func TestSortedSet_ZRevRange(t *testing.T) {
	ss := NewSortedSet()
	ss.ZAdd("member1", 1.5)
	ss.ZAdd("member2", 2.5)
	ss.ZAdd("member3", 0.5)

	// Test if the range is correct
	if len(ss.ZRevRange(0, 1)) != 1 {
		t.Errorf("Range error expected 0 got %d", len(ss.ZRevRange(0, 1)))
	}
	if len(ss.ZRevRange(0, 2)) != 2 {
		t.Errorf("Range error expected 2 got %d", len(ss.ZRevRange(0, 2)))
	}
}

func TestSortedSet_ZRevRangeByScore(t *testing.T) {
	ss := NewSortedSet()
	ss.ZAdd("member1", 1.5)
	ss.ZAdd("member2", 2.5)
	ss.ZAdd("member3", 0.5)

	// Test if the range is correct
	if len(ss.ZRevRangeByScore(1, 2, 0, -1, 0)) != 1 {
		t.Errorf("Range error expected 1 got %d", len(ss.ZRevRangeByScore(1, 2, 0, -1, 0)))
	}
	if len(ss.ZRevRangeByScore(0, 3, 0, -1, 0)) != 3 {
		t.Errorf("Range error expected 3 got %d", len(ss.ZRevRangeByScore(0, 3, 0, -1, 0)))
	}
}

func TestSortedSet_ZRevRank(t *testing.T) {
	ss := NewSortedSet()
	ss.ZAdd("member1", 1.5)
	ss.ZAdd("member2", 2.5)
	ss.ZAdd("member3", 0.5)

	// Test if the rank is correct
	v1, _ := ss.ZRevRank("member2")
	if v1 != 0 {
		t.Errorf("Rank error expected 1 got %d", v1)
	}
	v2, _ := ss.ZRank("member2")
	if v2 != 2 {
		t.Errorf("Rank error expected 3 got %d", v2)
	}
}

func TestSortedSet_ZCount(t *testing.T) {
	ss := NewSortedSet()
	ss.ZAdd("member1", 1.5)
	ss.ZAdd("member2", 2.5)
	ss.ZAdd("member3", 0.5)
	ss.ZAdd("member4", 3.5)

	// Test if the count is correct
	if ss.ZCount(1, 3, 0) != 2 {
		t.Errorf("Count error expected 3 got %d", ss.ZCount(1, 3, 0))
	}
	if ss.ZCount(2, 4, 0) != 2 {
		t.Errorf("Count error expected 2 got %d", ss.ZCount(2, 4, 0))
	}
}

func TestGetSetValue(t *testing.T) {
	ss := NewSortedSet()
	ss.ZAdd("member1", 1.5)
	ss.ZAdd("member2", 2.5)
	ss.ZAdd("member3", math.MaxFloat64)

	v := ss.GetValue()
	ss2 := NewSortedSet()
	ss2.SetValue(v)
	if ss2.dict.Len() != 3 {
		t.Errorf("Value error expected 3 got %d", ss2.dict.Len())
	}
	for _, item := range ss.ZRange(0, -1) {
		fmt.Println(item.Member, item.Score)
	}
}
