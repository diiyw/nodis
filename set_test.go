package nodis

import (
	"testing"
	"time"
)

func TestSet_Set(t *testing.T) {
	n := Open(Options{
		Path:         "testdata",
		SyncInterval: 60 * time.Second,
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
