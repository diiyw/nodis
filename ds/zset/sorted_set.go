package zset

import (
	"strconv"
	"sync"

	"github.com/dolthub/swiss"
	"github.com/kelindar/binary"
)

// SortedSet is a set which keys sorted by bound score
type SortedSet struct {
	sync.RWMutex
	dict     *swiss.Map[string, *Element]
	skiplist *skiplist
}

// NewSortedSet makes a new SortedSet
func NewSortedSet() *SortedSet {
	return &SortedSet{
		dict:     swiss.NewMap[string, *Element](16),
		skiplist: makeSkiplist(),
	}
}

// GetType returns the type of the data structure
func (sortedSet *SortedSet) GetType() string {
	return "zset"
}

// ZAdd puts member into set,  and returns whether it has inserted new node
func (sortedSet *SortedSet) ZAdd(member string, score float64) bool {
	element, ok := sortedSet.dict.Get(member)
	sortedSet.dict.Put(member, &Element{
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
	return int64(sortedSet.dict.Count())
}

// ZRem removes the given member from set
func (sortedSet *SortedSet) ZRem(member string) bool {
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
	element, ok := sortedSet.dict.Get(member)
	if !ok {
		return -1
	}
	r := sortedSet.skiplist.getRank(member, element.Score)
	if desc {
		r = sortedSet.skiplist.length - r
	} else {
		r--
	}
	return r + 1
}

// ZRank returns the rank of the given member, sort by ascending order, rank starts from 0
func (sortedSet *SortedSet) ZRank(member string) int64 {
	return sortedSet.getRank(member, false)
}

// ZRevRank returns the rank of the given member, sort by descending order, rank starts from 0
func (sortedSet *SortedSet) ZRevRank(member string) int64 {
	return sortedSet.getRank(member, true)
}

// ZScore returns the score of the given member
func (sortedSet *SortedSet) ZScore(member string) float64 {
	element, ok := sortedSet.dict.Get(member)
	if !ok {
		return 0
	}
	return element.Score
}

// forEachByRank visits each member which rank within [start, stop), sort by ascending order, rank starts from 0
func (sortedSet *SortedSet) forEachByRank(start int64, stop int64, desc bool, consumer func(element *Element) bool) {
	size := sortedSet.ZCard()
	if start < 0 || start >= size {
		panic("illegal start " + strconv.FormatInt(start, 10))
	}
	if stop < start || stop > size {
		panic("illegal end " + strconv.FormatInt(stop, 10))
	}

	// find start node
	var node *node
	if desc {
		node = sortedSet.skiplist.tail
		if start > 0 {
			node = sortedSet.skiplist.getByRank(size - start)
		}
	} else {
		node = sortedSet.skiplist.header.level[0].forward
		if start > 0 {
			node = sortedSet.skiplist.getByRank(start + 1)
		}
	}

	sliceSize := int(stop - start)
	for i := 0; i < sliceSize; i++ {
		if !consumer(&node.Element) {
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
func (sortedSet *SortedSet) rangeByRank(start int64, stop int64, desc bool) []*Element {
	sliceSize := int(stop - start)
	slice := make([]*Element, sliceSize)
	i := 0
	sortedSet.forEachByRank(start, stop, desc, func(element *Element) bool {
		slice[i] = element
		i++
		return true
	})
	return slice
}

// ZRangeByRank returns members which rank within [start, stop), sort by ascending order, rank starts from 0
func (sortedSet *SortedSet) ZRange(start int64, stop int64) []*Element {
	return sortedSet.rangeByRank(start, stop, false)
}

// ZRevRangeByRank returns members which rank within [start, stop), sort by descending order, rank starts from 0
func (sortedSet *SortedSet) ZRevRange(start int64, stop int64) []*Element {
	return sortedSet.rangeByRank(start, stop, true)
}

// rangeCount returns the number of  members which score or member within the given border
func (sortedSet *SortedSet) rangeCount(min Border, max Border) int64 {
	var i int64 = 0
	// ascending order
	sortedSet.forEachByRank(0, sortedSet.ZCard(), false, func(element *Element) bool {
		gtMin := min.less(element) // greater than min
		if !gtMin {
			// has not into range, continue foreach
			return true
		}
		ltMax := max.greater(element) // less than max
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
	return sortedSet.rangeCount(&ScoreBorder{Value: min, Exclude: false}, &ScoreBorder{Value: max, Exclude: false})
}

// forEach visits members which score or member within the given border
func (sortedSet *SortedSet) forEach(min Border, max Border, offset int64, limit int64, desc bool, consumer func(element *Element) bool) {
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
		if !consumer(&node.Element) {
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
		gtMin := min.less(&node.Element) // greater than min
		ltMax := max.greater(&node.Element)
		if !gtMin || !ltMax {
			break // break through score border
		}
	}
}

// zrange returns members which score or member within the given border
// param limit: <0 means no limit
func (sortedSet *SortedSet) zrange(min Border, max Border, offset int64, limit int64, desc bool) []*Element {
	if limit == 0 || offset < 0 {
		return make([]*Element, 0)
	}
	slice := make([]*Element, 0)
	sortedSet.forEach(min, max, offset, limit, desc, func(element *Element) bool {
		slice = append(slice, element)
		return true
	})
	return slice
}

// removeRange removes members which score or member within the given border
func (sortedSet *SortedSet) removeRange(min Border, max Border) int64 {
	removed := sortedSet.skiplist.removeRange(min, max, 0)
	for _, element := range removed {
		sortedSet.dict.Delete(element.Member)
	}
	return int64(len(removed))
}

// ZRemRangeByScore removes members which score or member within the given border
func (sortedSet *SortedSet) ZRemRangeByScore(min float64, max float64) int64 {
	return sortedSet.removeRange(&ScoreBorder{Value: min, Exclude: false}, &ScoreBorder{Value: max, Exclude: false})
}

// ZRemRangeByRank removes member ranking within [start, stop)
// sort by ascending order and rank starts from 0
func (sortedSet *SortedSet) ZRemRangeByRank(start int64, stop int64) int64 {
	removed := sortedSet.skiplist.removeRangeByRank(start+1, stop+1)
	for _, element := range removed {
		sortedSet.dict.Delete(element.Member)
	}
	return int64(len(removed))
}

func (sortedSet *SortedSet) ZPopMin(count int) []*Element {
	first := sortedSet.skiplist.getFirstInRange(scoreNegativeInfBorder, scorePositiveInfBorder)
	if first == nil {
		return nil
	}
	border := &ScoreBorder{
		Value:   first.Score,
		Exclude: false,
	}
	removed := sortedSet.skiplist.removeRange(border, scorePositiveInfBorder, count)
	for _, element := range removed {
		sortedSet.dict.Delete(element.Member)
	}
	return removed
}

func (sortedSet *SortedSet) ZPopMax(count int) []*Element {
	last := sortedSet.skiplist.getLastInRange(scoreNegativeInfBorder, scorePositiveInfBorder)
	if last == nil {
		return nil
	}
	border := &ScoreBorder{
		Value:   last.Score,
		Exclude: false,
	}
	removed := sortedSet.skiplist.removeRange(scoreNegativeInfBorder, border, count)
	for _, element := range removed {
		sortedSet.dict.Delete(element.Member)
	}
	return removed
}

// ZExists returns whether the given member exists in set
func (sortedSet *SortedSet) ZExists(member string) bool {
	_, ok := sortedSet.dict.Get(member)
	return ok
}

// ZRangeByScore returns members which score or member within the given border
func (sortedSet *SortedSet) ZRangeByScore(min float64, max float64) []*Element {
	return sortedSet.zrange(&ScoreBorder{Value: min, Exclude: false}, &ScoreBorder{Value: max, Exclude: false}, 0, -1, false)
}

// ZRevRangeByScore returns members which score or member within the given border
func (sortedSet *SortedSet) ZRevRangeByScore(min float64, max float64) []*Element {
	return sortedSet.zrange(&ScoreBorder{Value: min, Exclude: false}, &ScoreBorder{Value: max, Exclude: false}, 0, -1, true)
}

// ZIncrBy increases the score of the given member
func (sortedSet *SortedSet) ZIncrBy(member string, score float64) float64 {
	element, ok := sortedSet.dict.Get(member)
	if !ok {
		return 0
	}
	sortedSet.ZAdd(member, element.Score+score)
	return element.Score + score
}

func (sortedSet *SortedSet) Marshal() ([]byte, error) {
	var m = make(map[string]*Element)
	sortedSet.dict.Iter(func(key string, value *Element) bool {
		m[key] = value
		return false
	})
	return binary.Marshal(m)
}

func (sortedSet *SortedSet) Unmarshal(data []byte) error {
	var m = make(map[string]*Element)
	err := binary.Unmarshal(data, &m)
	if err != nil {
		return err
	}
	sortedSet.skiplist = makeSkiplist()
	sortedSet.dict = swiss.NewMap[string, *Element](16)
	for k, v := range m {
		sortedSet.ZAdd(k, v.Score)
	}
	return nil
}
