package nodis

import (
	"cmp"
	"math/rand/v2"
	"slices"
	"testing"
)

func orderedIndexEntries[K any](index *orderedIndex[K]) []indexEntry[K] {
	start := index.bound(func(K) bool { return false })
	cursor := index.span(start, indexPosition{rank: index.len()}).cursor()
	var entries []indexEntry[K]
	for entry, ok := cursor.next(); ok; entry, ok = cursor.next() {
		entries = append(entries, entry)
	}
	return entries
}

// Check the tree, ranks, traversal chain, free list, and arena accounting
// independently so stale unreachable nodes cannot hide behind correct results.
func assertOrderedIndex[K any](t testing.TB, index *orderedIndex[K], compare func(K, K) int) {
	t.Helper()
	seen := make([]bool, len(index.nodes))
	var visit func(uint32, *indexEntry[K], *indexEntry[K]) (int, int)
	visit = func(node uint32, lower, upper *indexEntry[K]) (int, int) {
		if node == 0 {
			return 0, 0
		}
		if int(node) >= len(index.nodes) || seen[node] {
			t.Fatalf("invalid or cyclic tree node %d", node)
		}
		seen[node] = true
		n := &index.nodes[node]
		if lower != nil && compareIndexEntry(n.entry, *lower, compare) <= 0 || upper != nil && compareIndexEntry(n.entry, *upper, compare) >= 0 {
			t.Fatal("tree order violated")
		}
		lh, ls := visit(n.left, lower, &n.entry)
		rh, rs := visit(n.right, &n.entry, upper)
		if lh-rh > 1 || rh-lh > 1 || int(n.height) != 1+max(lh, rh) || int(n.size) != 1+ls+rs {
			t.Fatalf("AVL balance or subtree metadata violated: node=%d lh=%d rh=%d height=%d size=%d want=%d", node, lh, rh, n.height, n.size, 1+ls+rs)
		}
		return 1 + max(lh, rh), 1 + ls + rs
	}
	_, count := visit(index.root, nil, nil)
	if count != index.len() {
		t.Fatal("tree size changed")
	}
	start := index.bound(func(K) bool { return false })
	var previous uint32
	var chainCount int
	for node := start.node; node != 0; node = index.nodes[node].next {
		if int(node) >= len(seen) || !seen[node] || chainCount >= count {
			t.Fatal("invalid or cyclic in-order chain")
		}
		n := &index.nodes[node]
		if n.prev != previous || previous != 0 && compareIndexEntry(index.nodes[previous].entry, n.entry, compare) >= 0 {
			t.Fatal("in-order links do not follow sorted order")
		}
		previous = node
		chainCount++
	}
	if chainCount != count {
		t.Fatal("in-order chain omitted live nodes")
	}
	for free := index.free; free != 0; free = index.nodes[free].next {
		if int(free) >= len(seen) || seen[free] {
			t.Fatal("invalid or cyclic free list")
		}
		seen[free] = true
		n := &index.nodes[free]
		if n.entry.id != 0 || n.left != 0 || n.right != 0 || n.prev != 0 || n.size != 0 || n.height != 0 {
			t.Fatal("free node retains live metadata")
		}
	}
	for node := 1; node < len(seen); node++ {
		if !seen[node] {
			t.Fatalf("arena node %d leaked", node)
		}
	}
}

func TestOrderedIndexRotationsDeletionAndRanks(t *testing.T) {
	for _, keys := range [][]int{
		{3, 2, 1}, {1, 2, 3}, {3, 1, 2}, {1, 3, 2},
		{8, 4, 12, 2, 6, 10, 14, 1, 3, 5, 7, 9, 11, 13, 15},
	} {
		var index orderedIndex[int]
		for _, key := range keys {
			index.add(key, ObjectID(key), cmp.Compare[int])
			assertOrderedIndex(t, &index, cmp.Compare[int])
		}
		// Replacing an existing entry never adds a second object/value pair.
		index.add(keys[0], ObjectID(keys[0]), cmp.Compare[int])
		if index.len() != len(keys) {
			t.Fatal("duplicate tree entry")
		}
		for _, key := range keys {
			index.remove(key, ObjectID(key), cmp.Compare[int])
			assertOrderedIndex(t, &index, cmp.Compare[int])
		}
		index.remove(999, 999, cmp.Compare[int])
		assertOrderedIndex(t, &index, cmp.Compare[int])
		if index.root != 0 {
			t.Fatal("tree did not become empty")
		}
	}
}

func TestOrderedIndexRandomMutationsAgainstSortedReference(t *testing.T) {
	random := rand.New(rand.NewPCG(1987, 1234))
	var index orderedIndex[int]
	var want []indexEntry[int]
	compare := func(a, b indexEntry[int]) int { return compareIndexEntry(a, b, cmp.Compare[int]) }
	for step := range 10000 {
		entry := indexEntry[int]{random.IntN(41) - 20, ObjectID(1 + random.IntN(64))}
		pos, exists := slices.BinarySearchFunc(want, entry, compare)
		if random.IntN(2) == 0 {
			index.add(entry.key, entry.id, cmp.Compare[int])
			if !exists {
				want = slices.Insert(want, pos, entry)
			}
		} else {
			index.remove(entry.key, entry.id, cmp.Compare[int])
			if exists {
				want = slices.Delete(want, pos, pos+1)
			}
		}
		if step%37 == 0 {
			assertOrderedIndex(t, &index, cmp.Compare[int])
			if got := orderedIndexEntries(&index); !slices.Equal(got, want) {
				t.Fatal("tree differs from sorted reference")
			}
			for key := -22; key <= 22; key++ {
				lower := 0
				for lower < len(want) && want[lower].key < key {
					lower++
				}
				upper := lower
				for upper < len(want) && want[upper].key == key {
					upper++
				}
				if index.lower(key, cmp.Compare[int]) != lower || index.upper(key, cmp.Compare[int]) != upper {
					t.Fatal("tree ranks differ from sorted reference")
				}
			}
		}
	}
	assertOrderedIndex(t, &index, cmp.Compare[int])
}

func TestOrderedIndexBulkBuildAndArenaReuse(t *testing.T) {
	for _, count := range []int{0, 1, 2, 3, 16, 257} {
		var index orderedIndex[int]
		for key := count; key > 0; key-- {
			index.appendUnsorted(indexEntry[int]{key, ObjectID(key)})
		}
		index.build(cmp.Compare[int])
		assertOrderedIndex(t, &index, cmp.Compare[int])
		if index.len() != count {
			t.Fatal("bulk build lost entries")
		}
		slots := len(index.nodes)
		for key := 1; key <= count; key++ {
			index.remove(key, ObjectID(key), cmp.Compare[int])
		}
		for key := count; key > 0; key-- {
			index.add(-key, ObjectID(key), cmp.Compare[int])
		}
		if len(index.nodes) != slots {
			t.Fatal("freed arena slots were not reused")
		}
		assertOrderedIndex(t, &index, cmp.Compare[int])
	}
}

func TestOrderedIndexTallTreeMutationsAndCursorsZeroAllocations(t *testing.T) {
	const count = 4096
	index := orderedIndex[int]{nodes: make([]indexNode[int], 1, count+1)}
	keys := make([]int, count)
	for i := range count {
		keys[i] = i * 2
		index.add(keys[i], ObjectID(i+1), cmp.Compare[int])
	}
	var step uint64
	if allocs := testing.AllocsPerRun(1000, func() {
		pos := int(step * 11400714819323198485 & (count - 1))
		step++
		index.remove(keys[pos], ObjectID(pos+1), cmp.Compare[int])
		keys[pos] ^= 1
		index.add(keys[pos], ObjectID(pos+1), cmp.Compare[int])
		start := index.bound(func(key int) bool { return key < 100 })
		end := index.bound(func(key int) bool { return key < 120 })
		cursor := index.span(start, end).cursor()
		var found int
		for entry, ok := cursor.next(); ok; entry, ok = cursor.next() {
			if entry.key < 100 || entry.key >= 120 {
				panic("tree cursor escaped range")
			}
			found++
		}
		if found != 10 {
			panic("tree cursor lost entries")
		}
	}); allocs != 0 {
		t.Fatalf("prepared AVL mutation/rank/cursor: %g allocs", allocs)
	}
	if len(index.nodes) != count+1 {
		t.Fatal("constant-cardinality updates grew the arena")
	}
	assertOrderedIndex(t, &index, cmp.Compare[int])
}

func TestOrderedIndexRemovedNodesReleaseBorrowedKeys(t *testing.T) {
	var index orderedIndex[scalarKey]
	for i := range 63 {
		key := scalarKey{kind: KindString, text: string(rune('A' + i))}
		index.add(key, ObjectID(i+1), compareScalarKey)
	}
	root := index.nodes[index.root]
	if root.left == 0 || root.right == 0 {
		t.Fatal("reference release test requires two children")
	}
	index.remove(root.entry.key, root.entry.id, compareScalarKey)
	for _, node := range index.nodes[1:] {
		if node.entry.key.text == root.entry.key.text {
			t.Fatal("successor replacement retained removed key")
		}
	}
	for i := range 63 {
		key := scalarKey{kind: KindString, text: string(rune('A' + i))}
		index.remove(key, ObjectID(i+1), compareScalarKey)
	}
	assertOrderedIndex(t, &index, compareScalarKey)
	for _, node := range index.nodes[1:] {
		if node.entry != (indexEntry[scalarKey]{}) {
			t.Fatal("free arena retains an encoded object key")
		}
	}
}

func FuzzOrderedIndexMutations(f *testing.F) {
	f.Add([]byte{0, 3, 1, 0, 2, 2, 0, 1, 3, 1, 2, 2})
	f.Add([]byte{0, 5, 1, 0, 3, 2, 0, 9, 3, 0, 7, 4, 1, 5, 1})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 2048 {
			t.Skip()
		}
		var index orderedIndex[int]
		reference := make(map[indexEntry[int]]struct{})
		for pos := 0; pos+2 < len(data); pos += 3 {
			entry := indexEntry[int]{int(int8(data[pos+1])), ObjectID(1 + data[pos+2]%16)}
			if data[pos]&1 == 0 {
				index.add(entry.key, entry.id, cmp.Compare[int])
				reference[entry] = struct{}{}
			} else {
				index.remove(entry.key, entry.id, cmp.Compare[int])
				delete(reference, entry)
			}
			if pos%48 == 0 {
				assertOrderedIndex(t, &index, cmp.Compare[int])
			}
		}
		assertOrderedIndex(t, &index, cmp.Compare[int])
		want := make([]indexEntry[int], 0, len(reference))
		for entry := range reference {
			want = append(want, entry)
		}
		slices.SortFunc(want, func(a, b indexEntry[int]) int {
			return compareIndexEntry(a, b, cmp.Compare[int])
		})
		if got := orderedIndexEntries(&index); !slices.Equal(got, want) {
			t.Fatal("fuzzed AVL tree differs from reference")
		}
	})
}
