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
	tx := n.writeKey(key, n.newZSet)
	tx.ds.(*zset.SortedSet).ZAdd(member, score)
	tx.commit()
	n.notify(pb.NewOp(pb.OpType_ZAdd, key).Member(member).Score(score))
}

func (n *Nodis) ZAddXX(key string, member string, score float64) int64 {
	tx := n.writeKey(key, n.newZSet)
	if tx.ds.(*zset.SortedSet).ZAddXX(member, score) {
		n.notify(pb.NewOp(pb.OpType_ZAdd, key).Member(member).Score(score))
		tx.commit()
		return 1
	}
	tx.commit()
	return 0
}

func (n *Nodis) ZAddNX(key string, member string, score float64) int64 {
	tx := n.writeKey(key, n.newZSet)
	if tx.ds.(*zset.SortedSet).ZAddNX(member, score) {
		n.notify(pb.NewOp(pb.OpType_ZAdd, key).Member(member).Score(score))
		tx.commit()
		return 1
	}
	tx.commit()
	return 0
}

// ZAddLT add member if score less than the current score
func (n *Nodis) ZAddLT(key string, member string, score float64) int64 {
	tx := n.writeKey(key, n.newZSet)
	if tx.ds.(*zset.SortedSet).ZAddLT(member, score) {
		n.notify(pb.NewOp(pb.OpType_ZAdd, key).Member(member).Score(score))
		tx.commit()
		return 1
	}
	tx.commit()
	return 0
}

// ZAddGT add member if score greater than the current score
func (n *Nodis) ZAddGT(key string, member string, score float64) int64 {
	tx := n.writeKey(key, n.newZSet)
	if tx.ds.(*zset.SortedSet).ZAddGT(member, score) {
		n.notify(pb.NewOp(pb.OpType_ZAdd, key).Member(member).Score(score))
		tx.commit()
		return 1
	}
	tx.commit()
	return 0
}

func (n *Nodis) ZCard(key string) int64 {
	tx := n.readKey(key)
	if !tx.isOk() {
		return 0
	}
	v := tx.ds.(*zset.SortedSet).ZCard()
	tx.commit()
	return v
}

func (n *Nodis) ZRank(key string, member string) int64 {
	tx := n.readKey(key)
	if !tx.isOk() {
		return 0
	}
	v := tx.ds.(*zset.SortedSet).ZRank(member)
	tx.commit()
	return v
}

func (n *Nodis) ZRankWithScore(key string, member string) (int64, *zset.Item) {
	tx := n.readKey(key)
	if !tx.isOk() {
		return 0, nil
	}
	c, v := tx.ds.(*zset.SortedSet).ZRankWithScore(member)
	tx.commit()
	return c, v
}

func (n *Nodis) ZRevRank(key string, member string) int64 {
	tx := n.readKey(key)
	if !tx.isOk() {
		return 0
	}
	v := tx.ds.(*zset.SortedSet).ZRevRank(member)
	tx.commit()
	return v
}

func (n *Nodis) ZRevRankWithScore(key string, member string) (int64, *zset.Item) {
	tx := n.readKey(key)
	if !tx.isOk() {
		return 0, nil
	}
	c, v := tx.ds.(*zset.SortedSet).ZRevRankWithScore(member)
	tx.commit()
	return c, v
}

func (n *Nodis) ZScore(key string, member string) float64 {
	tx := n.readKey(key)
	if !tx.isOk() {
		return 0
	}
	v := tx.ds.(*zset.SortedSet).ZScore(member)
	tx.commit()
	return v
}

func (n *Nodis) ZIncrBy(key string, member string, score float64) float64 {
	tx := n.writeKey(key, n.newZSet)
	v := tx.ds.(*zset.SortedSet).ZIncrBy(member, score)
	tx.commit()
	n.notify(pb.NewOp(pb.OpType_ZIncrBy, key).Member(member).Score(score))
	return v
}

func (n *Nodis) ZRange(key string, start int64, stop int64) []string {
	tx := n.readKey(key)
	if !tx.isOk() {
		return nil
	}
	els := tx.ds.(*zset.SortedSet).ZRange(start, stop)
	tx.commit()
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
	tx := n.readKey(key)
	if !tx.isOk() {
		return nil
	}
	v := tx.ds.(*zset.SortedSet).ZRange(start, stop)
	tx.commit()
	return v
}

func (n *Nodis) ZRevRange(key string, start int64, stop int64) []string {
	tx := n.readKey(key)
	if !tx.isOk() {
		return nil
	}
	els := tx.ds.(*zset.SortedSet).ZRevRange(start, stop)
	tx.commit()
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
	tx := n.readKey(key)
	if !tx.isOk() {
		return nil
	}
	v := tx.ds.(*zset.SortedSet).ZRevRange(start, stop)
	tx.commit()
	return v
}

func (n *Nodis) ZRangeByScore(key string, min float64, max float64) []string {
	tx := n.readKey(key)
	if !tx.isOk() {
		return nil
	}
	els := tx.ds.(*zset.SortedSet).ZRangeByScore(min, max)
	tx.commit()
	members := make([]string, len(els))
	for i, el := range els {
		members[i] = el.Member
	}
	return members
}

func (n *Nodis) ZRangeByScoreWithScores(key string, min float64, max float64) []*zset.Item {
	tx := n.readKey(key)
	if !tx.isOk() {
		return nil
	}
	v := tx.ds.(*zset.SortedSet).ZRangeByScore(min, max)
	tx.commit()
	return v
}

func (n *Nodis) ZRevRangeByScore(key string, min float64, max float64) []string {
	tx := n.readKey(key)
	if !tx.isOk() {
		return nil
	}
	els := tx.ds.(*zset.SortedSet).ZRevRangeByScore(min, max)
	tx.commit()
	members := make([]string, len(els))
	for i, el := range els {
		members[i] = el.Member
	}
	return members
}

func (n *Nodis) ZRevRangeByScoreWithScores(key string, min float64, max float64) []*zset.Item {
	tx := n.readKey(key)
	if !tx.isOk() {
		return nil
	}
	v := tx.ds.(*zset.SortedSet).ZRevRangeByScore(min, max)
	tx.commit()
	return v
}

func (n *Nodis) ZRem(key string, members ...string) int64 {
	tx := n.writeKey(key, nil)
	if !tx.isOk() {
		return 0
	}
	var removed int64 = 0
	for _, member := range members {
		if tx.ds.(*zset.SortedSet).ZRem(member) {
			removed++
		}
	}
	tx.commit()
	tx.key.changed = removed > 0
	if removed > 0 {
		n.notify(pb.NewOp(pb.OpType_ZRem, key).Members(members))
	}
	return removed
}

func (n *Nodis) ZRemRangeByRank(key string, start int64, stop int64) int64 {
	tx := n.writeKey(key, nil)
	if !tx.isOk() {
		return 0
	}
	removed := tx.ds.(*zset.SortedSet).ZRemRangeByRank(start, stop)
	tx.commit()
	tx.key.changed = removed > 0
	if removed > 0 {
		n.notify(pb.NewOp(pb.OpType_ZRemRangeByRank, key).Start(start).Stop(stop))
	}
	return removed
}

func (n *Nodis) ZRemRangeByScore(key string, min float64, max float64) int64 {
	tx := n.writeKey(key, nil)
	if !tx.isOk() {
		return 0
	}
	removed := tx.ds.(*zset.SortedSet).ZRemRangeByScore(min, max)
	tx.commit()
	tx.key.changed = removed > 0
	if removed > 0 {
		n.notify(pb.NewOp(pb.OpType_ZRemRangeByScore, key).Min(min).Max(max))
	}
	return removed
}

func (n *Nodis) ZExists(key string, member string) bool {
	tx := n.readKey(key)
	if !tx.isOk() {
		return false
	}
	v := tx.ds.(*zset.SortedSet).ZExists(member)
	tx.commit()
	return v
}

func (n *Nodis) ZClear(key string) {
	n.Del(key)
}
