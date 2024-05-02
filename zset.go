package nodis

import (
	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/ds/zset"
	"github.com/diiyw/nodis/pb"
)

func (n *Nodis) newZSet() ds.DataStruct {
	return zset.NewSortedSet()
}

func (n *Nodis) ZAdd(key string, member string, score float64) int64 {
	meta := n.store.writeKey(key, n.newZSet)
	v := meta.ds.(*zset.SortedSet).ZAdd(member, score)
	meta.commit()
	n.notify(pb.NewOp(pb.OpType_ZAdd, key).Member(member).Score(score))
	return v
}

// ZAddXX Only update elements that already exist. Don't add new elements.
func (n *Nodis) ZAddXX(key string, member string, score float64) int64 {
	meta := n.store.writeKey(key, n.newZSet)
	v := meta.ds.(*zset.SortedSet).ZAddXX(member, score)
	n.notify(pb.NewOp(pb.OpType_ZAdd, key).Member(member).Score(score))
	meta.commit()
	return v
}

func (n *Nodis) ZAddNX(key string, member string, score float64) int64 {
	meta := n.store.writeKey(key, n.newZSet)
	v := meta.ds.(*zset.SortedSet).ZAddNX(member, score)
	n.notify(pb.NewOp(pb.OpType_ZAdd, key).Member(member).Score(score))
	meta.commit()
	return v
}

// ZAddLT add member if score less than the current score
func (n *Nodis) ZAddLT(key string, member string, score float64) int64 {
	meta := n.store.writeKey(key, n.newZSet)
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
	meta := n.store.writeKey(key, n.newZSet)
	if meta.ds.(*zset.SortedSet).ZAddGT(member, score) {
		n.notify(pb.NewOp(pb.OpType_ZAdd, key).Member(member).Score(score))
		meta.commit()
		return 1
	}
	meta.commit()
	return 0
}

func (n *Nodis) ZCard(key string) int64 {
	meta := n.store.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	v := meta.ds.(*zset.SortedSet).ZCard()
	meta.commit()
	return v
}

func (n *Nodis) ZRank(key string, member string) int64 {
	meta := n.store.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	v := meta.ds.(*zset.SortedSet).ZRank(member)
	meta.commit()
	return v
}

func (n *Nodis) ZRankWithScore(key string, member string) (int64, *zset.Item) {
	meta := n.store.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return 0, nil
	}
	c, v := meta.ds.(*zset.SortedSet).ZRankWithScore(member)
	meta.commit()
	return c, v
}

func (n *Nodis) ZRevRank(key string, member string) int64 {
	meta := n.store.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	v := meta.ds.(*zset.SortedSet).ZRevRank(member)
	meta.commit()
	return v
}

func (n *Nodis) ZRevRankWithScore(key string, member string) (int64, *zset.Item) {
	meta := n.store.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return 0, nil
	}
	c, v := meta.ds.(*zset.SortedSet).ZRevRankWithScore(member)
	meta.commit()
	return c, v
}

func (n *Nodis) ZScore(key string, member string) float64 {
	meta := n.store.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	v := meta.ds.(*zset.SortedSet).ZScore(member)
	meta.commit()
	return v
}

func (n *Nodis) ZIncrBy(key string, member string, score float64) float64 {
	meta := n.store.writeKey(key, n.newZSet)
	v := meta.ds.(*zset.SortedSet).ZIncrBy(member, score)
	meta.commit()
	n.notify(pb.NewOp(pb.OpType_ZIncrBy, key).Member(member).Score(score))
	return v
}

func (n *Nodis) ZRange(key string, start int64, stop int64) []string {
	meta := n.store.readKey(key)
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
	meta := n.store.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return nil
	}
	v := meta.ds.(*zset.SortedSet).ZRange(start, stop)
	meta.commit()
	return v
}

func (n *Nodis) ZRevRange(key string, start int64, stop int64) []string {
	meta := n.store.readKey(key)
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
	meta := n.store.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return nil
	}
	v := meta.ds.(*zset.SortedSet).ZRevRange(start, stop)
	meta.commit()
	return v
}

func (n *Nodis) ZRangeByScore(key string, min float64, max float64, offset, count int64, mode int) []string {
	meta := n.store.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return nil
	}
	els := meta.ds.(*zset.SortedSet).ZRangeByScore(min, max, offset, count, mode)
	meta.commit()
	members := make([]string, len(els))
	for i, el := range els {
		members[i] = el.Member
	}
	return members
}

func (n *Nodis) ZRangeByScoreWithScores(key string, min float64, max float64, offset, count int64, mode int) []*zset.Item {
	meta := n.store.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return nil
	}
	v := meta.ds.(*zset.SortedSet).ZRangeByScore(min, max, offset, count, mode)
	meta.commit()
	return v
}

func (n *Nodis) ZRevRangeByScore(key string, min float64, max float64, offset, count int64, mode int) []string {
	meta := n.store.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return nil
	}
	els := meta.ds.(*zset.SortedSet).ZRevRangeByScore(min, max, offset, count, mode)
	meta.commit()
	members := make([]string, len(els))
	for i, el := range els {
		members[i] = el.Member
	}
	return members
}

func (n *Nodis) ZRevRangeByScoreWithScores(key string, min float64, max float64, offset, count int64, mode int) []*zset.Item {
	meta := n.store.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return nil
	}
	v := meta.ds.(*zset.SortedSet).ZRevRangeByScore(min, max, offset, count, mode)
	meta.commit()
	return v
}

func (n *Nodis) ZRem(key string, members ...string) int64 {
	meta := n.store.writeKey(key, nil)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	var removed int64 = meta.ds.(*zset.SortedSet).ZRem(members...)
	meta.key.changed = removed > 0
	meta.commit()
	if removed > 0 {
		n.notify(pb.NewOp(pb.OpType_ZRem, key).Members(members))
	}
	return removed
}

func (n *Nodis) ZRemRangeByRank(key string, start int64, stop int64) int64 {
	meta := n.store.writeKey(key, nil)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	removed := meta.ds.(*zset.SortedSet).ZRemRangeByRank(start, stop)
	meta.key.changed = removed > 0
	meta.commit()
	if removed > 0 {
		n.notify(pb.NewOp(pb.OpType_ZRemRangeByRank, key).Start(start).Stop(stop))
	}
	return removed
}

func (n *Nodis) ZRemRangeByScore(key string, min float64, max float64) int64 {
	meta := n.store.writeKey(key, nil)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	removed := meta.ds.(*zset.SortedSet).ZRemRangeByScore(min, max)
	meta.key.changed = removed > 0
	meta.commit()
	if removed > 0 {
		n.notify(pb.NewOp(pb.OpType_ZRemRangeByScore, key).Min(min).Max(max))
	}
	return removed
}

func (n *Nodis) ZExists(key string, member string) bool {
	meta := n.store.readKey(key)
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

// zCount returns the number of elements in the sorted set at key with a score between min and max.
func (n *Nodis) ZCount(key string, min, max float64, mode int) int64 {
	meta := n.store.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	v := meta.ds.(*zset.SortedSet).ZCount(min, max, mode)
	meta.commit()
	return v
}

// ZMax returns the member with the highest score in the sorted set at key.
func (n *Nodis) ZMax(key string) *zset.Item {
	meta := n.store.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return nil
	}
	v := meta.ds.(*zset.SortedSet).ZMax()
	meta.commit()
	return v
}

// ZMin returns the member with the lowest score in the sorted set at key.
func (n *Nodis) ZMin(key string) *zset.Item {
	meta := n.store.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return nil
	}
	v := meta.ds.(*zset.SortedSet).ZMin()
	meta.commit()
	return v
}

// ZUnionStore computes the union of numkeys sorted sets given by the specified keys, and stores the result in destination.
func (n *Nodis) ZUnionStore(destination string, keys []string, weights []float64, aggregate string) int64 {
	meta := n.store.writeKey(destination, n.newZSet)
	if !meta.isOk() {
		meta.commit()
		return 0
	}
	lockedKeys := make([]*metadata, 0, len(keys))
	for _, key := range keys {
		ds := n.store.readKey(key)
		if !ds.isOk() {
			ds.commit()
			continue
		}
		lockedKeys = append(lockedKeys, ds)
	}
	if len(lockedKeys) == 0 {
		meta.commit()
		return 0
	}
	var items = make(map[string]float64)
	for i, m := range lockedKeys {
		zs := m.ds.(*zset.SortedSet).ZRange(0, -1)
		for _, z := range zs {
			var weight float64 = 1
			if i < len(weights) {
				weight = weights[i]
			}
			if aggregate == "SUM" || aggregate == "" {
				items[z.Member] = items[z.Member]*weight + weight*z.Score
			}
			if aggregate == "MIN" {
				if _, ok := items[z.Member]; !ok || z.Score < items[z.Member] {
					items[z.Member] = z.Score * weight
				}
			}
			if aggregate == "MAX" {
				if _, ok := items[z.Member]; !ok || z.Score > items[z.Member] {
					items[z.Member] = z.Score * weight
				}
			}
		}
		m.commit()
	}
	for member, score := range items {
		meta.ds.(*zset.SortedSet).ZAdd(member, score)
	}
	meta.commit()
	n.notify(pb.NewOp(pb.OpType_ZUnionStore, destination).Keys(keys).Weights(weights).Aggregate(aggregate))
	return int64(len(items))
}
