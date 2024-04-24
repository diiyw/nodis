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
	var m = map[int64]bool{
		0:  true,
		2:  true,
		10: true,
		20: false,
		30: true,
	}
	for k, v := range m {
		s.SetBit(k, v)
		r := 0
		if v {
			r = 1
		}
		if s.GetBit(k) != int64(r) {
			t.Errorf("expected %v, got %v", v, s.GetBit(k))
		}
	}
}

func TestSetBit_BitCount(t *testing.T) {
	s := NewString()
	var m = map[int64]bool{
		0:  true,
		2:  true,
		10: true,
		20: false,
		30: true,
		8:  true,
		15: true,
	}
	for k, v := range m {
		s.SetBit(k, v)
	}
	if s.BitCount(0, 0) != 6 {
		t.Errorf("expected 6, got %d", s.BitCount(0, 0))
	}
}

func TestSet_Incr(t *testing.T) {
	s := NewString()
	s.Set([]byte("100"))
	if s.Incr(1) != 101 {
		t.Errorf("expected 101, got %d", s.Incr(1))
	}
}

func TestSet_Decr(t *testing.T) {
	s := NewString()
	s.Set([]byte("100"))
	if s.Decr(1) != 99 {
		t.Errorf("expected 99, got %d", s.Decr(1))
	}
	s.Set([]byte("0"))
	if s.Decr(1) != -1 {
		t.Errorf("expected -1, got %d", s.Decr(1))
	}
}
