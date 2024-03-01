package nodis

import (
	"os"
	"testing"
	"time"
)

func TestNodis_Open(t *testing.T) {
	opt := Options{
		Path:         "testdata",
		SyncInterval: 60 * time.Second,
	}
	got := Open(opt)
	if got == nil {
		t.Errorf("Open() = %v, want %v", got, "Nodis{}")
	}
}

func TestNodis_Sync(t *testing.T) {
	opt := Options{
		Path:         "testdata",
		SyncInterval: 60 * time.Second,
	}
	n := Open(opt)
	defer func() {
		os.ReadFile("testdata/nodis.db")
		os.ReadFile("testdata/nodis.meta")
	}()
	n.Set("test", []byte("test1"), 0)
	err := n.Sync()
	if err != nil {
		t.Errorf("Sync() = %v, want %v", err, nil)
	}
}

func TestNodis_OpenAndSync(t *testing.T) {
	opt := Options{
		Path:         "testdata",
		SyncInterval: 60 * time.Second,
	}
	n := Open(opt)
	n.Set("set", []byte("set"), 0)
	n.ZAdd("zset", "zset", 1)
	n.HSet("hset", "hset", []byte("hset"))
	n.LPush("lpush", []byte("lpush"))
	err := n.Sync()
	if err != nil {
		t.Errorf("Sync() = %v, want %v", err, nil)
	}
	n = Open(opt)
	v := n.Get("set")
	if v == nil {
		t.Errorf("Get() = %s, want %v", v, "set")
	}
	s := n.ZScore("zset", "zset")
	if s != 1 {
		t.Errorf("ZScore() = %v, want %v", v, 1)
	}
	v, ok := n.HGet("hset", "hset")
	if !ok || v == nil {
		t.Errorf("HGet() = %s, want %v", v, "hset")
	}
	v = n.LPop("lpush")
	if !ok || v == nil {
		t.Errorf("LPop() = %s, want %v", v, "lpush")
	}
}
