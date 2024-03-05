package zset

import (
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
	if ss.ZScore("member1") != 1.5 || ss.ZScore("member2") != 2.5 || ss.ZScore("member3") != 0.6 {
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

	if ss.ZScore("member1") != 1.5 {
		t.Errorf("Score error expected 1.5 got %f", ss.ZScore("member1"))
	}
	if ss.ZScore("member2") != 2.5 {
		t.Errorf("Score error expected 2.5 got %f", ss.ZScore("member2"))
	}
	if ss.ZScore("member3") != 0.5 {
		t.Errorf("Score error expected 0.5 got %f", ss.ZScore("member3"))
	}
}

func TestSortedSet_ZRange(t *testing.T) {
	ss := NewSortedSet()
	ss.ZAdd("member1", 1.5)
	ss.ZAdd("member2", 2.5)
	ss.ZAdd("member3", 0.5)

	// Test if the range is correct
	if len(ss.ZRange(0, 2)) != 2 {
		t.Errorf("Range error")
	}
}

func TestSortedSet_ZRangeByScore(t *testing.T) {
	ss := NewSortedSet()
	ss.ZAdd("member1", 1.5)
	ss.ZAdd("member2", 2.5)
	ss.ZAdd("member3", 0.5)

	// Test if the range is correct
	if len(ss.ZRangeByScore(1, 2)) != 1 {
		t.Errorf("Range error")
	}
}

func TestSortedSet_ZRank(t *testing.T) {
	ss := NewSortedSet()
	ss.ZAdd("member1", 1.5)
	ss.ZAdd("member2", 2.5)
	ss.ZAdd("member3", 0.5)

	// Test if the rank is correct
	if ss.ZRank("member1") != 2 {
		t.Errorf("Rank error expected 2 got %d", ss.ZRank("member1"))
	}
	if ss.ZRank("member2") != 3 {
		t.Errorf("Rank error expected 3 got %d", ss.ZRank("member2"))
	}
	if ss.ZRank("member3") != 1 {
		t.Errorf("Rank error expected 1 got %d", ss.ZRank("member3"))
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
	if ss.ZCard() != 2 {
		t.Errorf("Removal error")
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
	removed := ss.ZRemRangeByScore(1, 2)
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
	if len(ss.ZRevRangeByScore(1, 2)) != 1 {
		t.Errorf("Range error expected 1 got %d", len(ss.ZRevRangeByScore(1, 2)))
	}
	if len(ss.ZRevRangeByScore(0, 3)) != 3 {
		t.Errorf("Range error expected 3 got %d", len(ss.ZRevRangeByScore(0, 3)))
	}
}

func TestSortedSet_ZRevRank(t *testing.T) {
	ss := NewSortedSet()
	ss.ZAdd("member1", 1.5)
	ss.ZAdd("member2", 2.5)
	ss.ZAdd("member3", 0.5)

	// Test if the rank is correct
	if ss.ZRevRank("member2") != 1 {
		t.Errorf("Rank error expected 1 got %d", ss.ZRevRank("member2"))
	}
	if ss.ZRank("member2") != 3 {
		t.Errorf("Rank error expected 3 got %d", ss.ZRevRank("member2"))
	}
}

func TestSortedSet_ZCount(t *testing.T) {
	ss := NewSortedSet()
	ss.ZAdd("member1", 1.5)
	ss.ZAdd("member2", 2.5)
	ss.ZAdd("member3", 0.5)
	ss.ZAdd("member4", 3.5)

	// Test if the count is correct
	if ss.ZCount(1, 3) != 2 {
		t.Errorf("Count error expected 3 got %d", ss.ZCount(1, 3))
	}
	if ss.ZCount(2, 4) != 2 {
		t.Errorf("Count error expected 2 got %d", ss.ZCount(2, 4))
	}
}
