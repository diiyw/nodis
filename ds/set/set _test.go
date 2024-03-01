package set

import "testing"

func BenchmarkSet_Set(b *testing.B) {
	s := NewSet()
	for i := 0; i < b.N; i++ {
		s.Set([]byte("hello"))
	}
}
