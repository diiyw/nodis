package nodis

import (
	"fmt"
	"testing"
	"unsafe"
)

// Each case mutates a prepared index at a constant cardinality. Capacity growth
// is excluded so both the running time and allocation regressions are visible.
func BenchmarkSecondaryIndexMutation(b *testing.B) {
	for _, count := range []int{1024, 16384, 131072} {
		for _, mode := range []string{"middle-insert-remove", "first-remove-add", "random-replace"} {
			b.Run(fmt.Sprintf("%d/%s", count, mode), func(b *testing.B) {
				index := orderedIndex[scalarKey]{nodes: make([]indexNode[scalarKey], 1, count+2)}
				keys := make([]scalarKey, count)
				for i := range count {
					keys[i] = scalarKey{kind: KindInt, num: uint64(i * 2)}
					index.add(keys[i], ObjectID(i+1), compareScalarKey)
				}
				middle := scalarKey{kind: KindInt, num: uint64(count)}
				index.add(middle, ObjectID(count+1), compareScalarKey)
				index.remove(middle, ObjectID(count+1), compareScalarKey)
				var step uint64
				b.ReportAllocs()
				b.ResetTimer()
				for b.Loop() {
					switch mode {
					case "middle-insert-remove":
						index.add(middle, ObjectID(count+1), compareScalarKey)
						index.remove(middle, ObjectID(count+1), compareScalarKey)
					case "first-remove-add":
						index.remove(keys[0], 1, compareScalarKey)
						index.add(keys[0], 1, compareScalarKey)
					case "random-replace":
						// A multiplicative permutation visits the prepared keys.
						pos := int((step * 11400714819323198485) & uint64(count-1))
						step++
						index.remove(keys[pos], ObjectID(pos+1), compareScalarKey)
						keys[pos].num ^= 1
						index.add(keys[pos], ObjectID(pos+1), compareScalarKey)
					}
				}
				b.ReportMetric(float64(unsafe.Sizeof(indexNode[scalarKey]{})), "node-B")
			})
		}
	}
}
