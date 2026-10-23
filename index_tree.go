package nodis

import (
	"math"
	"slices"
)

// Nodes live in a reusable arena. Index zero is the absent child; freed nodes
// use next as a free-list link. Tree rotations never change the in-order chain.
type indexNode[K any] struct {
	entry                   indexEntry[K]
	left, right, prev, next uint32
	size                    uint32
	height                  uint8
}

// orderedIndex is an AVL tree augmented with subtree sizes and an in-order
// chain. Mutations and rank bounds take O(log n); traversing k entries is O(k).
// The database lock protects all nodes, cursors, and borrowed key references.
type orderedIndex[K any] struct {
	nodes      []indexNode[K]
	root, free uint32
}

func (index *orderedIndex[K]) len() int {
	return int(index.size(index.root))
}

func (index *orderedIndex[K]) size(node uint32) uint32 {
	if node == 0 {
		return 0
	}
	return index.nodes[node].size
}

func (index *orderedIndex[K]) height(node uint32) uint8 {
	if node == 0 {
		return 0
	}
	return index.nodes[node].height
}

func (index *orderedIndex[K]) refresh(node uint32) {
	n := &index.nodes[node]
	n.size = 1 + index.size(n.left) + index.size(n.right)
	n.height = 1 + max(index.height(n.left), index.height(n.right))
}

func (index *orderedIndex[K]) rotateLeft(node uint32) uint32 {
	right := index.nodes[node].right
	index.nodes[node].right = index.nodes[right].left
	index.nodes[right].left = node
	index.refresh(node)
	index.refresh(right)
	return right
}

func (index *orderedIndex[K]) rotateRight(node uint32) uint32 {
	left := index.nodes[node].left
	index.nodes[node].left = index.nodes[left].right
	index.nodes[left].right = node
	index.refresh(node)
	index.refresh(left)
	return left
}

func (index *orderedIndex[K]) balance(node uint32) uint32 {
	index.refresh(node)
	left, right := index.nodes[node].left, index.nodes[node].right
	difference := int(index.height(left)) - int(index.height(right))
	if difference > 1 {
		if index.height(index.nodes[left].left) < index.height(index.nodes[left].right) {
			index.nodes[node].left = index.rotateLeft(left)
		}
		return index.rotateRight(node)
	}
	if difference < -1 {
		if index.height(index.nodes[right].right) < index.height(index.nodes[right].left) {
			index.nodes[node].right = index.rotateRight(right)
		}
		return index.rotateLeft(node)
	}
	return node
}

func (index *orderedIndex[K]) allocate(entry indexEntry[K]) uint32 {
	if index.free != 0 {
		node := index.free
		index.free = index.nodes[node].next
		index.nodes[node] = indexNode[K]{entry: entry, height: 1, size: 1}
		return node
	}
	if len(index.nodes) == 0 {
		index.nodes = make([]indexNode[K], 1, 2)
	}
	// Never silently truncate an arena reference, including on 64-bit hosts.
	if uint64(len(index.nodes)) > math.MaxUint32 {
		panic("nodis: index node capacity exhausted")
	}
	node := uint32(len(index.nodes))
	index.nodes = append(index.nodes, indexNode[K]{entry: entry, height: 1, size: 1})
	return node
}

func (index *orderedIndex[K]) release(node uint32) {
	previous, next := index.nodes[node].prev, index.nodes[node].next
	if previous != 0 {
		index.nodes[previous].next = next
	}
	if next != 0 {
		index.nodes[next].prev = previous
	}
	// Clear the key so a free node cannot retain an old object's encoded bytes.
	index.nodes[node] = indexNode[K]{next: index.free}
	index.free = node
}

func (index *orderedIndex[K]) add(key K, id ObjectID, compare func(K, K) int) {
	index.root = index.insert(index.root, indexEntry[K]{key, id}, 0, 0, compare)
}

func (index *orderedIndex[K]) insert(node uint32, entry indexEntry[K], previous, next uint32, compare func(K, K) int) uint32 {
	if node == 0 {
		inserted := index.allocate(entry)
		index.nodes[inserted].prev, index.nodes[inserted].next = previous, next
		if previous != 0 {
			index.nodes[previous].next = inserted
		}
		if next != 0 {
			index.nodes[next].prev = inserted
		}
		return inserted
	}
	order := compareIndexEntry(entry, index.nodes[node].entry, compare)
	if order == 0 {
		// Equal values may reference a newly committed object's byte buffer.
		index.nodes[node].entry = entry
		return node
	}
	if order < 0 {
		child := index.insert(index.nodes[node].left, entry, previous, node, compare)
		// allocate may have grown the arena: do not retain node pointers across it.
		index.nodes[node].left = child
	} else {
		child := index.insert(index.nodes[node].right, entry, node, next, compare)
		index.nodes[node].right = child
	}
	return index.balance(node)
}

func (index *orderedIndex[K]) remove(key K, id ObjectID, compare func(K, K) int) {
	index.root = index.erase(index.root, indexEntry[K]{key, id}, compare)
}

func (index *orderedIndex[K]) erase(node uint32, entry indexEntry[K], compare func(K, K) int) uint32 {
	if node == 0 {
		return 0
	}
	order := compareIndexEntry(entry, index.nodes[node].entry, compare)
	if order < 0 {
		index.nodes[node].left = index.erase(index.nodes[node].left, entry, compare)
	} else if order > 0 {
		index.nodes[node].right = index.erase(index.nodes[node].right, entry, compare)
	} else {
		left, right := index.nodes[node].left, index.nodes[node].right
		if left == 0 || right == 0 {
			index.release(node)
			if left != 0 {
				return left
			}
			return right
		}
		// Move the immediate successor's entry, then unlink its original slot.
		// This preserves both the BST order and the threaded traversal chain.
		successor := index.nodes[index.nodes[node].next].entry
		index.nodes[node].entry = successor
		index.nodes[node].right = index.erase(right, successor, compare)
	}
	return index.balance(node)
}

type indexPosition struct {
	node uint32
	rank int
}

// bound returns the first entry for which before is false and its exact rank.
// before must be monotone in key order. The end position has node zero.
func (index *orderedIndex[K]) bound(before func(K) bool) indexPosition {
	var position indexPosition
	for node := index.root; node != 0; {
		n := &index.nodes[node]
		if before(n.entry.key) {
			position.rank += 1 + int(index.size(n.left))
			node = n.right
		} else {
			position.node = node
			node = n.left
		}
	}
	return position
}

func (index *orderedIndex[K]) lower(key K, compare func(K, K) int) int {
	return index.bound(func(value K) bool { return compare(value, key) < 0 }).rank
}

func (index *orderedIndex[K]) upper(key K, compare func(K, K) int) int {
	return index.bound(func(value K) bool { return compare(value, key) <= 0 }).rank
}

// Recovery appends unsorted nodes, sorts once, then builds a balanced tree and
// traversal links in linear time. It does not insert each historical mutation.
func (index *orderedIndex[K]) appendUnsorted(entry indexEntry[K]) {
	index.allocate(entry)
}

func (index *orderedIndex[K]) build(compare func(K, K) int) {
	if len(index.nodes) <= 1 {
		return
	}
	slices.SortFunc(index.nodes[1:], func(a, b indexNode[K]) int {
		return compareIndexEntry(a.entry, b.entry, compare)
	})
	index.root, index.free = index.buildRange(1, len(index.nodes)), 0
}

func (index *orderedIndex[K]) buildRange(start, end int) uint32 {
	if start == end {
		return 0
	}
	middle := start + (end-start)/2
	node := uint32(middle)
	index.nodes[node].left = index.buildRange(start, middle)
	index.nodes[node].right = index.buildRange(middle+1, end)
	index.nodes[node].prev = node - 1
	index.nodes[node].next = 0
	if middle+1 < len(index.nodes) {
		index.nodes[node].next = node + 1
	}
	index.refresh(node)
	return node
}

type indexSpan[K any] struct {
	index  *orderedIndex[K]
	first  uint32
	length int
}

func (index *orderedIndex[K]) span(start, end indexPosition) indexSpan[K] {
	return indexSpan[K]{index: index, first: start.node, length: end.rank - start.rank}
}

type indexCursor[K any] struct {
	index     *orderedIndex[K]
	node      uint32
	remaining int
}

func (span indexSpan[K]) cursor() indexCursor[K] {
	return indexCursor[K]{span.index, span.first, span.length}
}

func (cursor *indexCursor[K]) next() (indexEntry[K], bool) {
	if cursor.remaining == 0 {
		return indexEntry[K]{}, false
	}
	node := &cursor.index.nodes[cursor.node]
	cursor.node = node.next
	cursor.remaining--
	return node.entry, true
}
