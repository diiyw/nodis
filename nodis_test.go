package nodis

import "testing"

func TestNodis_Open(t *testing.T) {
	opt := Options{
		Path:         "testdata",
		SyncInterval: 10,
	}
	got := Open(opt)
	if got == nil {
		t.Errorf("Open() = %v, want %v", got, "Nodis{}")
	}
}

func TestNodis_Sync(t *testing.T) {
	opt := Options{
		Path:         "testdata",
		SyncInterval: 10,
	}
	n := Open(opt)
	n.Set("test", "test1", 0)
	err := n.Sync()
	if err != nil {
		t.Errorf("Sync() = %v, want %v", err, nil)
	}
}
