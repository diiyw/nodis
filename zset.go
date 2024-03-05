package nodis

import (
	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/ds/zset"
)

func (n *Nodis) newZSet() ds.DataStruct {
	return zset.NewSortedSet()
}

func (n *Nodis) ZAdd(key string, member string, score float64) {
	s := n.getDs(key, n.newZSet, 0)
	s.(*zset.SortedSet).ZAdd(member, score)
}

func (n *Nodis) ZCard(key string) int64 {
	s := n.getDs(key, nil, 0)
	if s == nil {
		return 0
	}
	return s.(*zset.SortedSet).ZCard()
}

func (n *Nodis) ZRank(key string, member string) int64 {
	s := n.getDs(key, nil, 0)
	if s == nil {
		return 0
	}
	return s.(*zset.SortedSet).ZRank(member)
}

func (n *Nodis) ZRevRank(key string, member string) int64 {
	s := n.getDs(key, nil, 0)
	if s == nil {
		return 0
	}
	return s.(*zset.SortedSet).ZRevRank(member)
}

func (n *Nodis) ZScore(key string, member string) float64 {
	s := n.getDs(key, nil, 0)
	if s == nil {
		return 0
	}
	return s.(*zset.SortedSet).ZScore(member)
}

func (n *Nodis) ZIncrBy(key string, score float64, member string) float64 {
	s := n.getDs(key, n.newZSet, 0)
	return s.(*zset.SortedSet).ZIncrBy(member, score)
}

func (n *Nodis) ZRange(key string, start int64, stop int64) []string {
	s := n.getDs(key, nil, 0)
	if s == nil {
		return nil
	}
	els := s.(*zset.SortedSet).ZRange(start, stop)
	members := make([]string, len(els))
	for i, el := range els {
		members[i] = el.Member
	}
	return members
}

func (n *Nodis) ZRangeWithScores(key string, start int64, stop int64) []*zset.Element {
	s := n.getDs(key, nil, 0)
	if s == nil {
		return nil
	}
	return s.(*zset.SortedSet).ZRange(start, stop)
}

func (n *Nodis) ZRevRange(key string, start int64, stop int64) []string {
	s := n.getDs(key, nil, 0)
	if s == nil {
		return nil
	}
	els := s.(*zset.SortedSet).ZRevRange(start, stop)
	members := make([]string, len(els))
	for i, el := range els {
		members[i] = el.Member
	}
	return members
}

func (n *Nodis) ZRevRangeWithScores(key string, start int64, stop int64) []*zset.Element {
	s := n.getDs(key, nil, 0)
	if s == nil {
		return nil
	}
	return s.(*zset.SortedSet).ZRevRange(start, stop)
}

func (n *Nodis) ZRangeByScore(key string, min float64, max float64) []string {
	s := n.getDs(key, nil, 0)
	if s == nil {
		return nil
	}
	els := s.(*zset.SortedSet).ZRangeByScore(min, max)
	members := make([]string, len(els))
	for i, el := range els {
		members[i] = el.Member
	}
	return members
}

func (n *Nodis) ZRangeByScoreWithScores(key string, min float64, max float64) []*zset.Element {
	s := n.getDs(key, nil, 0)
	if s == nil {
		return nil
	}
	return s.(*zset.SortedSet).ZRangeByScore(min, max)
}

func (n *Nodis) ZRevRangeByScore(key string, min float64, max float64) []string {
	s := n.getDs(key, nil, 0)
	if s == nil {
		return nil
	}
	els := s.(*zset.SortedSet).ZRevRangeByScore(min, max)
	members := make([]string, len(els))
	for i, el := range els {
		members[i] = el.Member
	}
	return members
}

func (n *Nodis) ZRevRangeByScoreWithScores(key string, min float64, max float64) []*zset.Element {
	s := n.getDs(key, nil, 0)
	if s == nil {
		return nil
	}
	return s.(*zset.SortedSet).ZRevRangeByScore(min, max)
}

func (n *Nodis) ZRem(key string, members ...string) int64 {
	s := n.getDs(key, nil, 0)
	if s == nil {
		return 0
	}
	var i int64 = 0
	for _, member := range members {
		if s.(*zset.SortedSet).ZRem(member) {
			i++
		}
	}
	return i
}

func (n *Nodis) ZRemRangeByRank(key string, start int64, stop int64) int64 {
	s := n.getDs(key, nil, 0)
	if s == nil {
		return 0
	}
	return s.(*zset.SortedSet).ZRemRangeByRank(start, stop)
}

func (n *Nodis) ZRemRangeByScore(key string, min float64, max float64) int64 {
	s := n.getDs(key, nil, 0)
	if s == nil {
		return 0
	}
	return s.(*zset.SortedSet).ZRemRangeByScore(min, max)
}

func (n *Nodis) ZExists(key string, member string) bool {
	s := n.getDs(key, nil, 0)
	if s == nil {
		return false
	}
	return s.(*zset.SortedSet).ZExists(member)
}

func (n *Nodis) ZClear(key string) {
	n.Clear(key)
}
