package nodis

import (
	"cmp"
	"slices"
)

// orderedIDs is a derived primary index, protected by the database lock. Normal
// ID creation appends; recovery also supports inserting out-of-order records.
type orderedIDs[K cmp.Ordered] []K

func (ids *orderedIDs[K]) add(id K) {
	pos, exists := slices.BinarySearch(*ids, id)
	if exists {
		return
	}
	*ids = append(*ids, id)
	copy((*ids)[pos+1:], (*ids)[pos:len(*ids)-1])
	(*ids)[pos] = id
}

func (ids *orderedIDs[K]) remove(id K) {
	pos, exists := slices.BinarySearch(*ids, id)
	if !exists {
		return
	}
	copy((*ids)[pos:], (*ids)[pos+1:])
	var zero K
	(*ids)[len(*ids)-1] = zero
	*ids = (*ids)[:len(*ids)-1]
}

func (ids orderedIDs[K]) after(id K) []K {
	pos, exists := slices.BinarySearch(ids, id)
	if exists {
		pos++
	}
	return ids[pos:]
}
