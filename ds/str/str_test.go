package str

import "testing"

func BenchmarkSet_Set(b *testing.B) {
	s := NewString()
	for i := 0; i < b.N; i++ {
		s.Set([]byte("hello"))
	}
}
