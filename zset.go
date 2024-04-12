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
	meta := n.writeKey(key, n.newZSet)
	meta.ds.(*zset.SortedSet).ZAdd(member, score)
	meta.commit()
	n.notify(pb.NewOp(pb.OpType_ZAdd, key).Member(member).Score(score))
}

func (n *Nodis) ZAddXX(key string, member string, score float64) int64 {
	meta := n.writeKey(key, n.newZSet)
	if meta.ds.(*zset.SortedSet).ZAddXX(member, score) {
		n.notify(pb.NewOp(pb.OpType_ZAdd, key).Member(member).Score(score))
		meta.commit()
		return 1
	}
	meta.commit()
	return 0
}

func (n *Nodis) ZAddNX(key string, member string, score float64) int64 {
	meta := n.writeKey(key, n.newZSet)
	if meta.ds.(*zset.SortedSet).ZAddNX(member, score) {
		n.notify(pb.NewOp(pb.OpType_ZAdd, key).Member(member).Score(score))
		meta.commit()
		return 1
	}
	meta.commit()
	return 0
}

// ZAddLT add member if score less than the current score
func (n *Nodis) ZAddLT(key string, member string, score float64) int64 {
	meta := n.writeKey(key, n.newZSet)
	if meta.ds.(*zset.SortedSet).ZAddLT(member, score) {
		n.notify(pb.NewOp(pb.OpType_ZAdd, key).Member(member).Score(score))
		meta.commit()
		return 1
	}
	meta.commit()
	return 0
}

// ZAddGT add member if score greater than the current score
func (n *Nodis) ZAddGT(key string, member string, score float64) int64 {
	meta := n.writeKey(key, n.newZSet)
	if meta.ds.(*zset.SortedSet).ZAddGT(member, score) {
		n.notify(pb.NewOp(pb.OpType_ZAdd, key).Member(member).Score(score))
		meta.commit()
		return 1
	}
	meta.commit()
	return 0
}

func (n *Nodis) ZCard(key string) int64 {
	meta := n.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	v := meta.ds.(*zset.SortedSet).ZCard()
	meta.commit()
	return v
}

func (n *Nodis) ZRank(key string, member string) int64 {
	meta := n.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	v := meta.ds.(*zset.SortedSet).ZRank(member)
	meta.commit()
	return v
}

func (n *Nodis) ZRankWithScore(key string, member string) (int64, *zset.Item) {
	meta := n.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return 0, nil
	}
	c, v := meta.ds.(*zset.SortedSet).ZRankWithScore(member)
	meta.commit()
	return c, v
}

func (n *Nodis) ZRevRank(key string, member string) int64 {
	meta := n.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	v := meta.ds.(*zset.SortedSet).ZRevRank(member)
	meta.commit()
	return v
}

func (n *Nodis) ZRevRankWithScore(key string, member string) (int64, *zset.Item) {
	meta := n.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return 0, nil
	}
	c, v := meta.ds.(*zset.SortedSet).ZRevRankWithScore(member)
	meta.commit()
	return c, v
}

func (n *Nodis) ZScore(key string, member string) float64 {
	meta := n.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	v := meta.ds.(*zset.SortedSet).ZScore(member)
	meta.commit()
	return v
}

func (n *Nodis) ZIncrBy(key string, member string, score float64) float64 {
	meta := n.writeKey(key, n.newZSet)
	v := meta.ds.(*zset.SortedSet).ZIncrBy(member, score)
	meta.commit()
	n.notify(pb.NewOp(pb.OpType_ZIncrBy, key).Member(member).Score(score))
	return v
}

func (n *Nodis) ZRange(key string, start int64, stop int64) []string {
	meta := n.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return nil
	}
	els := meta.ds.(*zset.SortedSet).ZRange(start, stop)
	meta.commit()
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
	meta := n.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return nil
	}
	v := meta.ds.(*zset.SortedSet).ZRange(start, stop)
	meta.commit()
	return v
}

func (n *Nodis) ZRevRange(key string, start int64, stop int64) []string {
	meta := n.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return nil
	}
	els := meta.ds.(*zset.SortedSet).ZRevRange(start, stop)
	meta.commit()
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
	meta := n.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return nil
	}
	v := meta.ds.(*zset.SortedSet).ZRevRange(start, stop)
	meta.commit()
	return v
}

func (n *Nodis) ZRangeByScore(key string, min float64, max float64) []string {
	meta := n.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return nil
	}
	els := meta.ds.(*zset.SortedSet).ZRangeByScore(min, max)
	meta.commit()
	members := make([]string, len(els))
	for i, el := range els {
		members[i] = el.Member
	}
	return members
}

func (n *Nodis) ZRangeByScoreWithScores(key string, min float64, max float64) []*zset.Item {
	meta := n.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return nil
	}
	v := meta.ds.(*zset.SortedSet).ZRangeByScore(min, max)
	meta.commit()
	return v
}

func (n *Nodis) ZRevRangeByScore(key string, min float64, max float64) []string {
	meta := n.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return nil
	}
	els := meta.ds.(*zset.SortedSet).ZRevRangeByScore(min, max)
	meta.commit()
	members := make([]string, len(els))
	for i, el := range els {
		members[i] = el.Member
	}
	return members
}

func (n *Nodis) ZRevRangeByScoreWithScores(key string, min float64, max float64) []*zset.Item {
	meta := n.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return nil
	}
	v := meta.ds.(*zset.SortedSet).ZRevRangeByScore(min, max)
	meta.commit()
	return v
}

func (n *Nodis) ZRem(key string, members ...string) int64 {
	meta := n.writeKey(key, nil)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	var removed int64 = 0
	for _, member := range members {
		if meta.ds.(*zset.SortedSet).ZRem(member) {
			removed++
		}
	}
	meta.commit()
	meta.key.changed = removed > 0
	if removed > 0 {
		n.notify(pb.NewOp(pb.OpType_ZRem, key).Members(members))
	}
	return removed
}

func (n *Nodis) ZRemRangeByRank(key string, start int64, stop int64) int64 {
	meta := n.writeKey(key, nil)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	removed := meta.ds.(*zset.SortedSet).ZRemRangeByRank(start, stop)
	meta.commit()
	meta.key.changed = removed > 0
	if removed > 0 {
		n.notify(pb.NewOp(pb.OpType_ZRemRangeByRank, key).Start(start).Stop(stop))
	}
	return removed
}

func (n *Nodis) ZRemRangeByScore(key string, min float64, max float64) int64 {
	meta := n.writeKey(key, nil)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	removed := meta.ds.(*zset.SortedSet).ZRemRangeByScore(min, max)
	meta.commit()
	meta.key.changed = removed > 0
	if removed > 0 {
		n.notify(pb.NewOp(pb.OpType_ZRemRangeByScore, key).Min(min).Max(max))
	}
	return removed
}

func (n *Nodis) ZExists(key string, member string) bool {
	meta := n.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return false
	}
	v := meta.ds.(*zset.SortedSet).ZExists(member)
	meta.commit()
	return v
}

func (n *Nodis) ZClear(key string) {
	n.Del(key)
}
