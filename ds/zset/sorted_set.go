package zset

import (
	"sync"

	"github.com/diiyw/nodis/ds"
	"github.com/dolthub/swiss"
	"github.com/kelindar/binary"
)

// SortedSet is a set which keys sorted by bound score
type SortedSet struct {
	sync.RWMutex
	dict     *swiss.Map[string, *Item]
	skiplist *skiplist
}

// NewSortedSet makes a new SortedSet
func NewSortedSet() *SortedSet {
	return &SortedSet{
		dict:     swiss.NewMap[string, *Item](16),
		skiplist: makeSkiplist(),
	}
}

// GetType returns the type of the data structure
func (sortedSet *SortedSet) GetType() ds.DataType {
	return ds.ZSet
}

// ZAdd puts member into set,  and returns whether it has inserted new node
func (sortedSet *SortedSet) ZAdd(member string, score float64) bool {
	sortedSet.Lock()
	defer sortedSet.Unlock()
	return sortedSet.zAdd(member, score)
}

// zAdd puts member into set,  and returns whether it has inserted new node
func (sortedSet *SortedSet) zAdd(member string, score float64) bool {
	element, ok := sortedSet.dict.Get(member)
	sortedSet.dict.Put(member, &Item{
		Member: member,
		Score:  score,
	})
	if ok {
		if score != element.Score {
			sortedSet.skiplist.remove(member, element.Score)
			sortedSet.skiplist.insert(member, score)
		}
		return false
	}
	sortedSet.skiplist.insert(member, score)
	return true
}

// ZCard returns number of members in set
func (sortedSet *SortedSet) ZCard() int64 {
	sortedSet.RLock()
	defer sortedSet.RUnlock()
	return int64(sortedSet.dict.Count())
}

// ZRem removes the given member from set
func (sortedSet *SortedSet) ZRem(member string) bool {
	sortedSet.Lock()
	defer sortedSet.Unlock()
	v, ok := sortedSet.dict.Get(member)
	if ok {
		sortedSet.skiplist.remove(member, v.Score)
		sortedSet.dict.Delete(member)
		return true
	}
	return false
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
	return r + 1
}

// ZRank returns the rank of the given member, sort by ascending order, rank starts from 0
func (sortedSet *SortedSet) ZRank(member string) int64 {
	sortedSet.RLock()
	defer sortedSet.RUnlock()
	return sortedSet.getRank(member, false)
}

// ZRevRank returns the rank of the given member, sort by descending order, rank starts from 0
func (sortedSet *SortedSet) ZRevRank(member string) int64 {
	sortedSet.RLock()
	defer sortedSet.RUnlock()
	return sortedSet.getRank(member, true)
}

// ZScore returns the score of the given member
func (sortedSet *SortedSet) ZScore(member string) float64 {
	sortedSet.RLock()
	defer sortedSet.RUnlock()
	element, ok := sortedSet.dict.Get(member)
	if !ok {
		return 0
	}
	return element.Score
}

// forEachByRank visits each member which rank within [start, stop), sort by ascending order, rank starts from 1
func (sortedSet *SortedSet) forEachByRank(start int64, stop int64, desc bool, consumer func(item *Item) bool) {
	size := sortedSet.ZCard()
	if start < 0 || start > size {
		return
	}
	if stop < start {
		return
	}
	// stop max is size
	if stop > size {
		stop = size
	}
	if start == 0 {
		start = 1
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

// rangeByRank returns members which rank within [start, stop), sort by ascending order, rank starts from 0
func (sortedSet *SortedSet) rangeByRank(start int64, stop int64, desc bool) []*Item {
	sliceSize := int(stop - start + 1)
	slice := make([]*Item, 0, sliceSize)
	sortedSet.forEachByRank(start, stop, desc, func(item *Item) bool {
		slice = append(slice, item)
		return true
	})
	return slice
}

// ZRange returns members which rank within [start, stop), sort by ascending order, rank starts from 0
func (sortedSet *SortedSet) ZRange(start int64, stop int64) []*Item {
	sortedSet.RLock()
	defer sortedSet.RUnlock()
	return sortedSet.rangeByRank(start, stop, false)
}

// ZRevRange returns members which rank within [start, stop), sort by descending order, rank starts from 0
func (sortedSet *SortedSet) ZRevRange(start int64, stop int64) []*Item {
	sortedSet.RLock()
	defer sortedSet.RUnlock()
	return sortedSet.rangeByRank(start, stop, true)
}

// rangeCount returns the number of  members which score or member within the given border
func (sortedSet *SortedSet) rangeCount(min float64, max float64) int64 {
	var i int64 = 0
	// ascending order
	sortedSet.forEachByRank(0, sortedSet.ZCard(), false, func(element *Item) bool {
		gtMin := min < element.Score // greater than min
		if !gtMin {
			// has not into range, continue foreach
			return true
		}
		ltMax := max > element.Score // less than max
		if !ltMax {
			// break through score border, break foreach
			return false
		}
		// gtMin && ltMax
		i++
		return true
	})
	return i
}

// ZCount returns the number of  members which score or member within the given border
func (sortedSet *SortedSet) ZCount(min float64, max float64) int64 {
	sortedSet.RLock()
	defer sortedSet.RUnlock()
	return sortedSet.rangeCount(min, max)
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
func (sortedSet *SortedSet) zRange(min float64, max float64, offset int64, limit int64, desc bool) []*Item {
	if limit == 0 || offset < 0 {
		return make([]*Item, 0)
	}
	slice := make([]*Item, 0)
	sortedSet.forEach(min, max, offset, limit, desc, func(element *Item) bool {
		slice = append(slice, element)
		return true
	})
	return slice
}

// removeRange removes members which score or member within the given border
func (sortedSet *SortedSet) removeRange(min float64, max float64) int64 {
	removed := sortedSet.skiplist.removeRange(min, max, 0)
	for _, element := range removed {
		sortedSet.dict.Delete(element.Member)
	}
	return int64(len(removed))
}

// ZRemRangeByScore removes members which score or member within the given border
func (sortedSet *SortedSet) ZRemRangeByScore(min float64, max float64) int64 {
	sortedSet.Lock()
	defer sortedSet.Unlock()
	return sortedSet.removeRange(min, max)
}

// ZRemRangeByRank removes member ranking within [start, stop)
// sort by ascending order and rank starts from 0
func (sortedSet *SortedSet) ZRemRangeByRank(start int64, stop int64) int64 {
	sortedSet.Lock()
	defer sortedSet.Unlock()
	removed := sortedSet.skiplist.removeRangeByRank(start+1, stop+1)
	for _, element := range removed {
		sortedSet.dict.Delete(element.Member)
	}
	return int64(len(removed))
}

// ZExists returns whether the given member exists in set
func (sortedSet *SortedSet) ZExists(member string) bool {
	sortedSet.RLock()
	defer sortedSet.RUnlock()
	_, ok := sortedSet.dict.Get(member)
	return ok
}

// ZRangeByScore returns members which score or member within the given border
func (sortedSet *SortedSet) ZRangeByScore(min float64, max float64) []*Item {
	sortedSet.RLock()
	defer sortedSet.RUnlock()
	return sortedSet.zRange(min, max, 0, -1, false)
}

// ZRevRangeByScore returns members which score or member within the given border
func (sortedSet *SortedSet) ZRevRangeByScore(min float64, max float64) []*Item {
	sortedSet.RLock()
	defer sortedSet.RUnlock()
	return sortedSet.zRange(min, max, 0, -1, true)
}

// ZIncrBy increases the score of the given member
func (sortedSet *SortedSet) ZIncrBy(member string, score float64) float64 {
	sortedSet.Lock()
	defer sortedSet.Unlock()
	element, ok := sortedSet.dict.Get(member)
	if !ok {
		return 0
	}
	sortedSet.zAdd(member, element.Score+score)
	return element.Score + score
}

func (sortedSet *SortedSet) MarshalBinary() ([]byte, error) {
	var m = make(map[string]*Item)
	sortedSet.RLock()
	defer sortedSet.RUnlock()
	sortedSet.dict.Iter(func(key string, value *Item) bool {
		m[key] = value
		return false
	})
	return binary.Marshal(m)
}

func (sortedSet *SortedSet) UnmarshalBinary(data []byte) error {
	var m = make(map[string]*Item)
	err := binary.Unmarshal(data, &m)
	if err != nil {
		return err
	}
	sortedSet.skiplist = makeSkiplist()
	sortedSet.dict = swiss.NewMap[string, *Item](16)
	for k, v := range m {
		sortedSet.zAdd(k, v.Score)
	}
	return nil
}
