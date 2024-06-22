package zset

import (
	"encoding/binary"
	"errors"
	"path/filepath"

	"github.com/diiyw/nodis/ds"
	"github.com/tidwall/btree"
)

const (
	MinOpen = 1
	MaxOpen = 2
)

// SortedSet is a set which keys sorted by bound score
type SortedSet struct {
	dict     btree.Map[string, *Item]
	skiplist *skiplist
}

// NewSortedSet makes a new SortedSet
func NewSortedSet() *SortedSet {
	return &SortedSet{
		skiplist: makeSkiplist(),
	}
}

// Type returns the type of the data structure
func (sortedSet *SortedSet) Type() ds.ValueType {
	return ds.ZSet
}

// ZAdd puts member into set,  and returns whether it has inserted new node
func (sortedSet *SortedSet) ZAdd(member string, score float64) int64 {
	return sortedSet.zAdd(member, score)
}

// ZAddXX Only update elements that already exist. Don't add new elements.
func (sortedSet *SortedSet) ZAddXX(member string, score float64) int64 {
	_, ok := sortedSet.dict.Get(member)
	if ok {
		return sortedSet.zAdd(member, score)
	}
	return 0
}

// ZAddNX Only add new elements. Don't update already existing elements.
func (sortedSet *SortedSet) ZAddNX(member string, score float64) int64 {
	_, ok := sortedSet.dict.Get(member)
	if !ok {
		return sortedSet.zAdd(member, score)
	}
	return 0
}

// ZAddLT add member if score less than the current score
func (sortedSet *SortedSet) ZAddLT(member string, score float64) bool {
	element, ok := sortedSet.dict.Get(member)
	if ok && element.Score > score {
		sortedSet.zAdd(member, score)
		return true
	}
	return false
}

// ZAddGT add member if score greater than the current score
func (sortedSet *SortedSet) ZAddGT(member string, score float64) bool {
	element, ok := sortedSet.dict.Get(member)
	if ok && element.Score < score {
		sortedSet.zAdd(member, score)
		return true
	}
	return false
}

// zAdd puts member into set,  and returns whether it has inserted new node
func (sortedSet *SortedSet) zAdd(member string, score float64) int64 {
	element, ok := sortedSet.dict.Get(member)
	sortedSet.dict.Set(member, &Item{
		Member: member,
		Score:  score,
	})
	if ok {
		if score != element.Score {
			sortedSet.skiplist.remove(member, element.Score)
			sortedSet.skiplist.insert(member, score)
		}
		return 0
	}
	sortedSet.skiplist.insert(member, score)
	return 1
}

// ZCard returns number of members in set
func (sortedSet *SortedSet) ZCard() int64 {
	return int64(sortedSet.dict.Len())
}

// ZRem removes the given member from set
func (sortedSet *SortedSet) ZRem(members ...string) int64 {
	var count int64
	for _, member := range members {
		v, ok := sortedSet.dict.Get(member)
		if ok {
			sortedSet.skiplist.remove(member, v.Score)
			sortedSet.dict.Delete(member)
			count++
		}
	}
	return count
}

// getRank returns the rank of the given member, sort by ascending order, rank starts from 0
func (sortedSet *SortedSet) getRank(member string, desc bool) (rank int64) {
	item, ok := sortedSet.dict.Get(member)
	if !ok {
		return -1
	}
	r := sortedSet.skiplist.getRank(member, item.Score)
	if desc {
		r = sortedSet.skiplist.length - r
	} else {
		r--
	}
	return r
}

// ZRank returns the rank of the given member, sort by ascending order, rank starts from 0
func (sortedSet *SortedSet) ZRank(member string) (int64, error) {
	_, ok := sortedSet.dict.Get(member)
	if !ok {
		return 0, errors.New("member not found")
	}
	return sortedSet.getRank(member, false), nil
}

// ZRankWithScore returns the rank of the given member, sort by ascending order, rank starts from 0
func (sortedSet *SortedSet) ZRankWithScore(member string) (int64, *Item) {
	element, ok := sortedSet.dict.Get(member)
	if !ok {
		return 0, nil
	}
	return sortedSet.getRank(member, false), element
}

// ZRevRank returns the rank of the given member, sort by descending order, rank starts from 0
func (sortedSet *SortedSet) ZRevRank(member string) (int64, error) {
	_, ok := sortedSet.dict.Get(member)
	if !ok {
		return 0, errors.New("member not found")
	}
	return sortedSet.getRank(member, true), nil
}

// ZRevRankWithScore returns the rank of the given member, sort by descending order, rank starts from 0
func (sortedSet *SortedSet) ZRevRankWithScore(member string) (int64, *Item) {
	element, ok := sortedSet.dict.Get(member)
	if !ok {
		return 0, nil
	}
	return sortedSet.getRank(member, true), element
}

// ZScore returns the score of the given member
func (sortedSet *SortedSet) ZScore(member string) (float64, error) {
	element, ok := sortedSet.dict.Get(member)
	if !ok {
		return 0, errors.New("member not found")
	}
	return element.Score, nil
}

// forEachByRank visits each member which rank within [start, stop], sort by ascending order, rank starts from 0
func (sortedSet *SortedSet) forEachByRank(start int64, stop int64, desc bool, consumer func(item *Item) bool) {
	size := sortedSet.ZCard()
	if start > size {
		return
	}
	if start == 0 {
		start = 1
	}
	if stop < 0 {
		stop = size + stop + 1
	}
	if stop < start {
		return
	}
	if start < 0 {
		start = size + start
	}
	// stop max is size
	if stop > size {
		stop = size
	}
	// find start node
	var node *node
	if desc {
		node = sortedSet.skiplist.tail
		if start > 1 {
			node = sortedSet.skiplist.getByRank(size - start)
		}
	} else {
		node = sortedSet.skiplist.header.level[0].forward
		if start > 1 {
			node = sortedSet.skiplist.getByRank(start)
		}
	}

	sliceSize := int(stop - start)
	for i := 0; i <= sliceSize; i++ {
		if !consumer(&node.Item) {
			break
		}
		if desc {
			node = node.backward
		} else {
			node = node.level[0].forward
		}
	}
}

// rangeByRank returns members which rank within [start, stop], sort by ascending order, rank starts from 0
func (sortedSet *SortedSet) rangeByRank(start int64, stop int64, desc bool) []*Item {
	slice := make([]*Item, 0)
	sortedSet.forEachByRank(start, stop, desc, func(item *Item) bool {
		slice = append(slice, item)
		return true
	})
	return slice
}

// ZRange returns members which rank within [start, stop], sort by ascending order, rank starts from 0
func (sortedSet *SortedSet) ZRange(start int64, stop int64) []*Item {
	return sortedSet.rangeByRank(start, stop, false)
}

// ZRevRange returns members which rank within [start, stop], sort by descending order, rank starts from 0
func (sortedSet *SortedSet) ZRevRange(start int64, stop int64) []*Item {
	return sortedSet.rangeByRank(start, stop, true)
}

// rangeCount returns the number of  members which score or member within the given border
func (sortedSet *SortedSet) rangeCount(min float64, max float64, mode int) int64 {
	var i int64 = 0
	// ascending order
	sortedSet.forEachByRank(0, sortedSet.ZCard(), false, func(element *Item) bool {
		matchMin := element.Score >= min
		if mode&MinOpen == MinOpen {
			matchMin = element.Score > min
		}
		matchMax := element.Score <= max
		if mode&MaxOpen == MaxOpen {
			matchMax = element.Score < max
		}
		if matchMin && matchMax {
			i++
		}
		return true
	})
	return i
}

// ZCount returns the number of  members which score or member within the given border
func (sortedSet *SortedSet) ZCount(min float64, max float64, mode int) int64 {
	return sortedSet.rangeCount(min, max, mode)
}

// forEach visits members which score or member within the given border
func (sortedSet *SortedSet) forEach(min float64, max float64, offset int64, limit int64, desc bool, consumer func(element *Item) bool) {
	// find start node
	var node *node
	if desc {
		node = sortedSet.skiplist.getLastInRange(min, max)
	} else {
		node = sortedSet.skiplist.getFirstInRange(min, max)
	}

	for node != nil && offset > 0 {
		if desc {
			node = node.backward
		} else {
			node = node.level[0].forward
		}
		offset--
	}

	// A negative limit returns all elements from the offset
	for i := 0; (i < int(limit) || limit < 0) && node != nil; i++ {
		if !consumer(&node.Item) {
			break
		}
		if desc {
			node = node.backward
		} else {
			node = node.level[0].forward
		}
		if node == nil {
			break
		}
		gtMin := min <= node.Item.Score // greater than min
		ltMax := max >= node.Item.Score
		if !gtMin || !ltMax {
			break // break through score border
		}
	}
}

// zRange returns members which score or member within the given border
// param limit: <0 means no limit
func (sortedSet *SortedSet) zRange(min float64, max float64, offset int64, limit int64, desc bool, mode int) []*Item {
	if limit == 0 || offset < 0 {
		return make([]*Item, 0)
	}
	slice := make([]*Item, 0)
	sortedSet.forEach(min, max, offset, limit, desc, func(element *Item) bool {
		if mode&MinOpen == MinOpen && element.Score == min {
			return true
		}
		if mode&MaxOpen == MaxOpen && element.Score == max {
			return true
		}
		slice = append(slice, element)
		return true
	})
	return slice
}

// removeRange removes members which score or member within the given border
func (sortedSet *SortedSet) removeRange(min float64, max float64, mode int) int64 {
	removed := sortedSet.skiplist.removeRange(min, max, 0, mode)
	for _, element := range removed {
		sortedSet.dict.Delete(element.Member)
	}
	return int64(len(removed))
}

// ZRemRangeByScore removes members which score or member within the given border
func (sortedSet *SortedSet) ZRemRangeByScore(min float64, max float64, mode int) int64 {
	return sortedSet.removeRange(min, max, mode)
}

// ZRemRangeByRank removes member ranking within [start, stop]
// sort by ascending order and rank starts from 0
func (sortedSet *SortedSet) ZRemRangeByRank(start int64, stop int64) int64 {
	if stop < 0 {
		stop = sortedSet.ZCard() + stop
	}
	if start < 0 {
		start = sortedSet.ZCard() + start
	}
	if start >= stop || start < 0 {
		return 0
	}
	removed := sortedSet.skiplist.removeRangeByRank(start, stop)
	for _, element := range removed {
		sortedSet.dict.Delete(element.Member)
	}
	return int64(len(removed))
}

// ZExists returns whether the given member exists in set
func (sortedSet *SortedSet) ZExists(member string) bool {
	_, ok := sortedSet.dict.Get(member)
	return ok
}

// ZRangeByScore returns members which score or member within the given border
func (sortedSet *SortedSet) ZRangeByScore(min float64, max float64, offset, count int64, mode int) []*Item {
	return sortedSet.zRange(min, max, offset, count, false, mode)
}

// ZRevRangeByScore returns members which score or member within the given border
func (sortedSet *SortedSet) ZRevRangeByScore(min float64, max float64, offset, count int64, mode int) []*Item {
	return sortedSet.zRange(min, max, offset, count, true, mode)
}

// ZIncrBy increases the score of the given member
func (sortedSet *SortedSet) ZIncrBy(member string, score float64) float64 {
	element, ok := sortedSet.dict.Get(member)
	if ok {
		score += element.Score
	}
	sortedSet.zAdd(member, score)
	return score
}

// ZMax returns the member with the highest score
func (sortedSet *SortedSet) ZMax() *Item {
	return &sortedSet.skiplist.tail.Item
}

// ZMin returns the member with the lowest score
func (sortedSet *SortedSet) ZMin() *Item {
	return &sortedSet.skiplist.header.level[0].forward.Item
}

// ZScan returns members which score or member within the given border
func (sortedSet *SortedSet) ZScan(cursor int64, match string, count int64) (int64, []*Item) {
	var items = make([]*Item, 0)
	if count == 0 {
		count = sortedSet.ZCard()
	}
	if match == "" {
		match = "*"
	}
	sortedSet.forEachByRank(cursor, cursor+count, false, func(element *Item) bool {
		if matched, _ := filepath.Match(match, element.Member); matched {
			items = append(items, element)
		}
		return true
	})
	return cursor + int64(len(items)), nil
}

func (sortedSet *SortedSet) GetValue() []byte {
	var keyScores = make([]byte, 0, sortedSet.dict.Len())
	sortedSet.dict.Scan(func(_ string, v *Item) bool {
		data := v.encode()
		dataLen := len(data)
		var b = make([]byte, 8+dataLen)
		n := binary.PutVarint(b, int64(dataLen))
		copy(b[n:], data)
		keyScores = append(keyScores, b[:n+dataLen]...)
		return true
	})
	return keyScores
}

func (sortedSet *SortedSet) SetValue(keyScores []byte) {
	sortedSet.skiplist = makeSkiplist()
	for {
		if len(keyScores) == 0 {
			break
		}
		dataLen, n := binary.Varint(keyScores)
		if n <= 0 {
			break
		}
		end := n + int(dataLen)
		item := decodeItem(keyScores[n:end])
		sortedSet.zAdd(item.Member, item.Score)
		keyScores = keyScores[end:]
	}
}
