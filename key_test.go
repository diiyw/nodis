package nodis

import (
	"testing"
	"time"
)

func TestKey_Expire(t *testing.T) {
	n := Open(&Options{
		Path:            "testdata",
		RecycleDuration: 60 * time.Second,
	})
	n.Set("test", []byte("test1"), 1)
	time.Sleep(2 * time.Second)
	v := n.Get("test")
	if v != nil {
		t.Errorf("Get() = %v, want %v", true, false)
	}
	n.Set("test", []byte("test1"), 2)
	n.Expire("test", 300)
	time.Sleep(2 * time.Second)
	v = n.Get("test")
	if v == nil {
		t.Errorf("Get() = %v, want %v", false, true)
	}
}
