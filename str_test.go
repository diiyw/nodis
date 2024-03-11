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
