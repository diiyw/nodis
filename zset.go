package nodis

import (
	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/ds/zset"
	"github.com/diiyw/nodis/patch"
)

func (n *Nodis) newZSet() ds.Value {
	return zset.NewSortedSet()
}

func (n *Nodis) ZAdd(key string, member string, score float64) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, n.newZSet)
		v = meta.value.(*zset.SortedSet).ZAdd(member, score)
		n.signalModifiedKey(key, meta)
		n.notify(func() []patch.Op {
			return []patch.Op{{patch.OpTypeZAdd, &patch.OpZAdd{Key: key, Member: member, Score: score}}}
		})
		return nil
	})
	return v
}

// ZAddXX Only update elements that already exist. Don't add new elements.
func (n *Nodis) ZAddXX(key string, member string, score float64) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, n.newZSet)
		v = meta.value.(*zset.SortedSet).ZAddXX(member, score)
		n.signalModifiedKey(key, meta)
		n.notify(func() []patch.Op {
			return []patch.Op{{patch.OpTypeZAdd, &patch.OpZAdd{Key: key, Member: member, Score: score}}}
		})
		return nil
	})
	return v
}

func (n *Nodis) ZAddNX(key string, member string, score float64) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, n.newZSet)
		v = meta.value.(*zset.SortedSet).ZAddNX(member, score)
		n.signalModifiedKey(key, meta)
		n.notify(func() []patch.Op {
			return []patch.Op{{patch.OpTypeZAdd, &patch.OpZAdd{Key: key, Member: member, Score: score}}}
		})
		return nil
	})
	return v
}

// ZAddLT add member if score less than the current score
func (n *Nodis) ZAddLT(key string, member string, score float64) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, n.newZSet)
		if meta.value.(*zset.SortedSet).ZAddLT(member, score) {
			n.signalModifiedKey(key, meta)
			n.notify(func() []patch.Op {
				return []patch.Op{{patch.OpTypeZAdd, &patch.OpZAdd{Key: key, Member: member, Score: score}}}
			})
			v = 1
		}
		return nil
	})
	return v
}

// ZAddGT add member if score greater than the current score
func (n *Nodis) ZAddGT(key string, member string, score float64) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, n.newZSet)
		if meta.value.(*zset.SortedSet).ZAddGT(member, score) {
			n.signalModifiedKey(key, meta)
			n.notify(func() []patch.Op {
				return []patch.Op{{patch.OpTypeZAdd, &patch.OpZAdd{Key: key, Member: member, Score: score}}}
			})
			v = 1
		}
		return nil
	})
	return v
}

func (n *Nodis) ZCard(key string) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		v = meta.value.(*zset.SortedSet).ZCard()
		return nil
	})
	return v
}

func (n *Nodis) ZRank(key string, member string) (v int64, err error) {
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		v, err = meta.value.(*zset.SortedSet).ZRank(member)
		return nil
	})
	return
}

func (n *Nodis) ZRankWithScore(key string, member string) (int64, *zset.Item) {
	var c int64
	var v *zset.Item
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		c, v = meta.value.(*zset.SortedSet).ZRankWithScore(member)
		return nil
	})
	return c, v
}

func (n *Nodis) ZRevRank(key string, member string) (v int64, err error) {
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		v, err = meta.value.(*zset.SortedSet).ZRevRank(member)
		return nil
	})
	return v, nil
}

func (n *Nodis) ZRevRankWithScore(key string, member string) (int64, *zset.Item) {
	var c int64
	var v *zset.Item
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		c, v = meta.value.(*zset.SortedSet).ZRevRankWithScore(member)
		return nil
	})
	return c, v
}

func (n *Nodis) ZScore(key string, member string) (v float64, err error) {
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		v, err = meta.value.(*zset.SortedSet).ZScore(member)
		return nil
	})
	return
}

func (n *Nodis) ZIncrBy(key string, member string, score float64) float64 {
	var v float64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, n.newZSet)
		v = meta.value.(*zset.SortedSet).ZIncrBy(member, score)
		n.signalModifiedKey(key, meta)
		n.notify(func() []patch.Op {
			return []patch.Op{{patch.OpTypeZIncrBy, &patch.OpZIncrBy{Key: key, Member: member, Score: score}}}
		})
		return nil
	})
	return v
}

func (n *Nodis) ZRange(key string, start int64, stop int64) []string {
	var v []string
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		els := meta.value.(*zset.SortedSet).ZRange(start, stop)
		v = make([]string, len(els))
		for i, el := range els {
			if el == nil {
				continue
			}
			v[i] = el.Member
		}
		return nil
	})
	return v
}

func (n *Nodis) ZRangeWithScores(key string, start int64, stop int64) []*zset.Item {
	var v []*zset.Item
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {

			return nil
		}
		v = meta.value.(*zset.SortedSet).ZRange(start, stop)
		return nil
	})
	return v
}

func (n *Nodis) ZRevRange(key string, start int64, stop int64) []string {
	var v []string
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		els := meta.value.(*zset.SortedSet).ZRevRange(start, stop)

		v = make([]string, len(els))
		for i, el := range els {
			if el == nil {
				continue
			}
			v[i] = el.Member
		}
		return nil
	})
	return v
}

func (n *Nodis) ZRevRangeWithScores(key string, start int64, stop int64) []*zset.Item {
	var v []*zset.Item
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		v = meta.value.(*zset.SortedSet).ZRevRange(start, stop)
		return nil
	})
	return v
}

func (n *Nodis) ZRangeByScore(key string, min float64, max float64, offset, count int64, mode int) []string {
	var v []string
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		els := meta.value.(*zset.SortedSet).ZRangeByScore(min, max, offset, count, mode)
		v = make([]string, len(els))
		for i, el := range els {
			v[i] = el.Member
		}
		return nil
	})
	return v
}

func (n *Nodis) ZRangeByScoreWithScores(key string, min float64, max float64, offset, count int64, mode int) []*zset.Item {
	var v []*zset.Item
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {

			return nil
		}
		v = meta.value.(*zset.SortedSet).ZRangeByScore(min, max, offset, count, mode)
		return nil
	})
	return v
}

func (n *Nodis) ZRevRangeByScore(key string, min float64, max float64, offset, count int64, mode int) []string {
	var v []string
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		els := meta.value.(*zset.SortedSet).ZRevRangeByScore(min, max, offset, count, mode)
		v = make([]string, len(els))
		for i, el := range els {
			v[i] = el.Member
		}
		return nil
	})
	return v
}

func (n *Nodis) ZRevRangeByScoreWithScores(key string, min float64, max float64, offset, count int64, mode int) []*zset.Item {
	var v []*zset.Item
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {

			return nil
		}
		v = meta.value.(*zset.SortedSet).ZRevRangeByScore(min, max, offset, count, mode)
		return nil
	})
	return v
}

func (n *Nodis) ZRem(key string, members ...string) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, nil)
		if !meta.isOk() {
			return nil
		}
		v = meta.value.(*zset.SortedSet).ZRem(members...)
		if v > 0 {
			n.signalModifiedKey(key, meta)
			n.notify(func() []patch.Op {
				return []patch.Op{{patch.OpTypeZRem, &patch.OpZRem{Key: key, Members: members}}}
			})
		}
		return nil
	})
	return v
}

func (n *Nodis) ZRemRangeByRank(key string, start int64, stop int64) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, nil)
		if !meta.isOk() {
			return nil
		}
		v = meta.value.(*zset.SortedSet).ZRemRangeByRank(start, stop)
		if v > 0 {
			n.signalModifiedKey(key, meta)
			n.notify(func() []patch.Op {
				return []patch.Op{{patch.OpTypeZRemRangeByRank, &patch.OpZRemRangeByRank{Key: key, Start: start, Stop: stop}}}
			})
		}
		return nil
	})
	return v
}

func (n *Nodis) ZRemRangeByScore(key string, min float64, max float64, mode int) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, nil)
		if !meta.isOk() {
			return nil
		}
		v = meta.value.(*zset.SortedSet).ZRemRangeByScore(min, max, mode)
		if v > 0 {
			n.signalModifiedKey(key, meta)
			n.notify(func() []patch.Op {
				return []patch.Op{{patch.OpTypeZRemRangeByScore, &patch.OpZRemRangeByScore{Key: key, Min: min, Max: max, Mode: int64(mode)}}}
			})
		}
		return nil
	})
	return v
}

func (n *Nodis) ZExists(key string, member string) bool {
	var v bool
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		v = meta.value.(*zset.SortedSet).ZExists(member)
		return nil
	})
	return v
}

func (n *Nodis) ZClear(key string) {
	n.Del(key)
}

// ZCount returns the number of elements in the sorted set at key with a score between min and max.
func (n *Nodis) ZCount(key string, min, max float64, mode int) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		v = meta.value.(*zset.SortedSet).ZCount(min, max, mode)
		return nil
	})
	return v
}

// ZMax returns the member with the highest score in the sorted set at key.
func (n *Nodis) ZMax(key string) *zset.Item {
	var v *zset.Item
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		v = meta.value.(*zset.SortedSet).ZMax()
		return nil
	})
	return v
}

// ZMin returns the member with the lowest score in the sorted set at key.
func (n *Nodis) ZMin(key string) *zset.Item {
	var v *zset.Item
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		v = meta.value.(*zset.SortedSet).ZMin()
		return nil
	})
	return v
}

func (n *Nodis) ZUnion(keys []string, weights []float64, aggregate string) []*zset.Item {
	var v []*zset.Item
	_ = n.exec(func(tx *Tx) error {
		var items = make(map[string]float64)
		for i, key := range keys {
			m := tx.readKey(key)
			if !m.isOk() {
				continue
			}
			var weight float64 = 1
			if i < len(weights) {
				weight = weights[i]
			}
			zs := m.value.(*zset.SortedSet).ZRange(0, -1)
			for _, z := range zs {
				if aggregate == "SUM" || aggregate == "" {
					if _, ok := items[z.Member]; !ok {
						items[z.Member] = z.Score * weight
					} else {
						items[z.Member] = items[z.Member]*weight + weight*z.Score
					}
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
		}
		for member, score := range items {
			v = append(v, &zset.Item{Member: member, Score: score})
		}
		return nil
	})
	return v
}

// ZUnionStore computes the union of numkeys sorted sets given by the specified keys, and stores the result in destination.
func (n *Nodis) ZUnionStore(destination string, keys []string, weights []float64, aggregate string) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(destination, n.newZSet)
		if !meta.isOk() {
			return nil
		}
		items := n.ZUnion(keys, weights, aggregate)
		if len(items) == 0 {
			return nil
		}
		for _, item := range items {
			meta.value.(*zset.SortedSet).ZAdd(item.Member, item.Score)
		}
		n.signalModifiedKey(destination, meta)
		n.notify(func() []patch.Op {
			return []patch.Op{{patch.OpTypeZUnionStore, &patch.OpZUnionStore{Key: destination, Keys: keys, Weights: weights, Aggregate: aggregate}}}
		})
		v = int64(len(items))
		return nil
	})
	return v
}

func (n *Nodis) ZInter(keys []string, weights []float64, aggregate string) []*zset.Item {
	var v []*zset.Item
	_ = n.exec(func(tx *Tx) error {
		var items = make(map[string]float64)
		for i, key := range keys {
			m := tx.readKey(key)
			if !m.isOk() {
				return nil
			}
			var weight float64 = 1
			if i < len(weights) {
				weight = weights[i]
			}
			zs := m.value.(*zset.SortedSet).ZRange(0, -1)
			for _, z := range zs {
				var found = true
				for j, otherKey := range keys {
					otherZ := tx.readKey(otherKey)
					if i == j {
						continue
					}
					if !otherZ.value.(*zset.SortedSet).ZExists(z.Member) {
						found = false
						break
					}
				}
				if found {
					if aggregate == "SUM" || aggregate == "" {
						if _, ok := items[z.Member]; !ok {
							items[z.Member] = z.Score * weight
						} else {
							items[z.Member] = items[z.Member]*weight + weight*z.Score
						}
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
			}
		}
		for member, score := range items {
			v = append(v, &zset.Item{Member: member, Score: score})
		}
		return nil
	})
	return v
}

// ZInterStore computes the intersection of numkeys sorted sets given by the specified keys, and stores the result in destination.
func (n *Nodis) ZInterStore(destination string, keys []string, weights []float64, aggregate string) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		items := n.ZInter(keys, weights, aggregate)
		meta := tx.writeKey(destination, n.newZSet)
		if !meta.isOk() {
			return nil
		}
		if len(items) == 0 {
			return nil
		}
		for _, item := range items {
			meta.value.(*zset.SortedSet).ZAdd(item.Member, item.Score)
		}
		n.signalModifiedKey(destination, meta)
		n.notify(func() []patch.Op {
			return []patch.Op{{patch.OpTypeZInterStore, &patch.OpZInterStore{Key: destination, Keys: keys, Weights: weights, Aggregate: aggregate}}}
		})
		v = int64(len(items))
		return nil
	})
	return v
}
