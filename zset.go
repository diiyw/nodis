package nodis

import (
	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/ds/zset"
	"github.com/diiyw/nodis/pb"
)

func (n *Nodis) newZSet() ds.DataStruct {
	return zset.NewSortedSet()
}

func (n *Nodis) ZAdd(key string, member string, score float64) {
	k, s := n.getDs(key, n.newZSet, 0)
	k.changed.Store(true)
	n.notify(pb.NewOp(pb.OpType_ZAdd, key).Member(member).Score(score))
	s.(*zset.SortedSet).ZAdd(member, score)
}

func (n *Nodis) ZCard(key string) int64 {
	_, s := n.getDs(key, nil, 0)
	if s == nil {
		return 0
	}
	return s.(*zset.SortedSet).ZCard()
}

func (n *Nodis) ZRank(key string, member string) int64 {
	_, s := n.getDs(key, nil, 0)
	if s == nil {
		return 0
	}
	return s.(*zset.SortedSet).ZRank(member)
}

func (n *Nodis) ZRevRank(key string, member string) int64 {
	_, s := n.getDs(key, nil, 0)
	if s == nil {
		return 0
	}
	return s.(*zset.SortedSet).ZRevRank(member)
}

func (n *Nodis) ZScore(key string, member string) float64 {
	_, s := n.getDs(key, nil, 0)
	if s == nil {
		return 0
	}
	return s.(*zset.SortedSet).ZScore(member)
}

func (n *Nodis) ZIncrBy(key string, member string, score float64) float64 {
	k, s := n.getDs(key, n.newZSet, 0)
	k.changed.Store(true)
	n.notify(pb.NewOp(pb.OpType_ZIncrBy, key).Member(member).Score(score))
	return s.(*zset.SortedSet).ZIncrBy(member, score)
}

func (n *Nodis) ZRange(key string, start int64, stop int64) []string {
	_, s := n.getDs(key, nil, 0)
	if s == nil {
		return nil
	}
	els := s.(*zset.SortedSet).ZRange(start, stop)
	members := make([]string, len(els))
	for i, el := range els {
		if el == nil {
			continue
		}
		members[i] = el.Member
	}
	return members
}

func (n *Nodis) ZRangeWithScores(key string, start int64, stop int64) []*zset.Item {
	_, s := n.getDs(key, nil, 0)
	if s == nil {
		return nil
	}
	return s.(*zset.SortedSet).ZRange(start, stop)
}

func (n *Nodis) ZRevRange(key string, start int64, stop int64) []string {
	_, s := n.getDs(key, nil, 0)
	if s == nil {
		return nil
	}
	els := s.(*zset.SortedSet).ZRevRange(start, stop)
	members := make([]string, len(els))
	for i, el := range els {
		if el == nil {
			continue
		}
		members[i] = el.Member
	}
	return members
}

func (n *Nodis) ZRevRangeWithScores(key string, start int64, stop int64) []*zset.Item {
	_, s := n.getDs(key, nil, 0)
	if s == nil {
		return nil
	}
	return s.(*zset.SortedSet).ZRevRange(start, stop)
}

func (n *Nodis) ZRangeByScore(key string, min float64, max float64) []string {
	_, s := n.getDs(key, nil, 0)
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

func (n *Nodis) ZRangeByScoreWithScores(key string, min float64, max float64) []*zset.Item {
	_, s := n.getDs(key, nil, 0)
	if s == nil {
		return nil
	}
	return s.(*zset.SortedSet).ZRangeByScore(min, max)
}

func (n *Nodis) ZRevRangeByScore(key string, min float64, max float64) []string {
	_, s := n.getDs(key, nil, 0)
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

func (n *Nodis) ZRevRangeByScoreWithScores(key string, min float64, max float64) []*zset.Item {
	_, s := n.getDs(key, nil, 0)
	if s == nil {
		return nil
	}
	return s.(*zset.SortedSet).ZRevRangeByScore(min, max)
}

func (n *Nodis) ZRem(key string, members ...string) int64 {
	k, s := n.getDs(key, nil, 0)
	if s == nil {
		return 0
	}
	var removed int64 = 0
	for _, member := range members {
		if s.(*zset.SortedSet).ZRem(member) {
			removed++
		}
	}
	k.changed.Store(removed > 0)
	if removed > 0 {
		n.notify(pb.NewOp(pb.OpType_ZRem, key).Members(members))
	}
	return removed
}

func (n *Nodis) ZRemRangeByRank(key string, start int64, stop int64) int64 {
	k, s := n.getDs(key, nil, 0)
	if s == nil {
		return 0
	}
	removed := s.(*zset.SortedSet).ZRemRangeByRank(start, stop)
	k.changed.Store(removed > 0)
	if removed > 0 {
		n.notify(pb.NewOp(pb.OpType_ZRemRangeByRank, key).Start(start).Stop(stop))
	}
	return removed
}

func (n *Nodis) ZRemRangeByScore(key string, min float64, max float64) int64 {
	k, s := n.getDs(key, nil, 0)
	if s == nil {
		return 0
	}
	removed := s.(*zset.SortedSet).ZRemRangeByScore(min, max)
	k.changed.Store(removed > 0)
	if removed > 0 {
		n.notify(pb.NewOp(pb.OpType_ZRemRangeByScore, key).Min(min).Max(max))
	}
	return removed
}

func (n *Nodis) ZExists(key string, member string) bool {
	_, s := n.getDs(key, nil, 0)
	if s == nil {
		return false
	}
	return s.(*zset.SortedSet).ZExists(member)
}

func (n *Nodis) ZClear(key string) {
	n.Del(key)
}
