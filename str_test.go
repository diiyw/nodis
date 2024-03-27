package nodis

import (
	"testing"
	"time"
)

func TestStr_Set(t *testing.T) {
	n := Open(&Options{
		Path:            "testdata",
		RecycleDuration: 60 * time.Second,
	})
	n.Set("a", []byte("b"), 0)
	v := n.Get("a")
	if string(v) != "b" {
		t.Errorf("Set failed expected b got `%s`", string(v))
	}
	n.Set("a", []byte("b"), 0)
	v = n.Get("a")
	if string(v) != "b" {
		t.Errorf("Set failed expected b got `%s`", string(v))
	}
	n.Set("a", []byte("b"), 1)
	v = n.Get("a")
	if string(v) != "b" {
		t.Errorf("Set failed expected b got `%s`", string(v))
	}
}

func TestStr_SetBit(t *testing.T) {
	n := Open(&Options{
		Path:            "testdata",
		RecycleDuration: 60 * time.Second,
	})
	n.SetBit("a", 0, 1)
	if n.GetBit("a", 0) != 1 {
		t.Errorf("SetBit failed expected 1 got %d", n.GetBit("a", 0))
	}
	n.SetBit("a", 0, 0)
	if n.GetBit("a", 0) != 0 {
		t.Errorf("SetBit failed expected 0 got %d", n.GetBit("a", 0))
	}
}

func TestStr_BitCount(t *testing.T) {
	n := Open(&Options{
		Path:            "testdata",
		RecycleDuration: 60 * time.Second,
	})
	n.SetBit("a", 0, 1)
	n.SetBit("a", 1, 1)
	n.SetBit("a", 2, 1)
	n.SetBit("a", 3, 1)
	n.SetBit("a", 4, 1)
	n.SetBit("a", 5, 1)
	n.SetBit("a", 6, 1)
	n.SetBit("a", 7, 1)
	if n.BitCount("a") != 8 {
		t.Errorf("BitCount failed expected 8 got %d", n.BitCount("a"))
	}
}
