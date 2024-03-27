package str

import "testing"

func BenchmarkSet_Set(b *testing.B) {
	s := NewString()
	for i := 0; i < b.N; i++ {
		s.Set([]byte("hello"))
	}
}

// TestSetBit_SetBit tests the SetBit method
func TestSetBit_SetBit(t *testing.T) {
	s := NewString()
	var m = map[int64]int{
		0:  1,
		2:  1,
		10: 1,
		20: 0,
		30: 1,
	}
	for k, v := range m {
		s.SetBit(k, v)
		if s.GetBit(k) != v {
			t.Errorf("expected %d, got %d", v, s.GetBit(k))
		}
	}
}

func TestSetBit_BitCount(t *testing.T) {
	s := NewString()
	var m = map[int64]int{
		0:  1,
		2:  1,
		10: 1,
		20: 0,
		30: 1,
		8:  1,
		15: 1,
	}
	for k, v := range m {
		s.SetBit(k, v)
	}
	if s.BitCount() != 6 {
		t.Errorf("expected 6, got %d", s.BitCount())
	}
}
