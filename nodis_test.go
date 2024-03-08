package nodis

import (
	"os"
	"strconv"
	"testing"
	"time"
)

func TestNodis_Open(t *testing.T) {
	opt := Options{
		Path:         "testdata",
		TidyDuration: 60 * time.Second,
	}
	got := Open(&opt)
	if got == nil {
		t.Errorf("Open() = %v, want %v", got, "Nodis{}")
	}
}

func TestNodis_Sync(t *testing.T) {
	opt := Options{
		Path:         "testdata",
		TidyDuration: 60 * time.Second,
	}
	n := Open(&opt)
	defer func() {
		_, _ = os.ReadFile("testdata/nodis.db")
		_, _ = os.ReadFile("testdata/nodis.meta")
	}()
	n.Set("test", []byte("test1"), 0)
	err := n.sync()
	if err != nil {
		t.Errorf("Sync() = %v, want %v", err, nil)
	}
}

func TestNodis_OpenAndSync(t *testing.T) {
	_ = os.RemoveAll("testdata")
	opt := DefaultOptions
	opt.Path = "testdata"
	n := Open(opt)
	n.Set("set", []byte("set"), 0)
	n.ZAdd("zset", "zset", 1)
	n.HSet("hset", "hset", []byte("hset"))
	n.LPush("lpush", []byte("lpush"))
	err := n.sync()
	if err != nil {
		t.Errorf("Sync() = %v, want %v", err, nil)
	}
	n.Close()
	n = Open(opt)
	v := n.Get("set")
	if v == nil {
		t.Errorf("Get() = %s, want %v", v, "set")
	}
	s := n.ZScore("zset", "zset")
	if s != 1 {
		t.Errorf("ZScore() = %v, want %v", v, 1)
	}
	v = n.HGet("hset", "hset")
	if v == nil {
		t.Errorf("HGet() = %s, want %v", v, "hset")
	}
	v = n.LPop("lpush")
	if v == nil {
		t.Errorf("LPop() = %s, want %v", v, "lpush")
	}
}

func TestNodis_OpenAndSyncBigdata10000(t *testing.T) {
	_ = os.Remove("testdata/nodis.db")
	_ = os.Remove("testdata/nodis.meta")
	opt := Options{
		Path:         "testdata",
		TidyDuration: 60 * time.Second,
	}
	n := Open(&opt)
	for i := 0; i < 10000; i++ {
		is := strconv.Itoa(i)
		n.Set(is, []byte(is), 0)
	}
	for i := 10000; i < 20000; i++ {
		n.ZAdd("zset", strconv.Itoa(i), float64(i))
	}
	for i := 20000; i < 30000; i++ {
		n.HSet("hset", strconv.Itoa(i), []byte(strconv.Itoa(i)))
	}
	for i := 30000; i < 40000; i++ {
		n.LPush("lpush", []byte(strconv.Itoa(i)))
	}
	err := n.sync()
	if err != nil {
		t.Errorf("Sync() = %v, want %v", err, nil)
	}
	n = Open(&opt)
	for i := 0; i < 10000; i++ {
		is := strconv.Itoa(i)
		v := n.Get(is)
		if string(v) != is {
			t.Errorf("Get() = %s, want %v", v, is)
		}
	}
	for i := 10000; i < 20000; i++ {
		v := n.ZScore("zset", strconv.Itoa(i))
		if v != float64(i) {
			t.Errorf("ZScore() = %v, want %v", v, i)
		}
	}

	for i := 20000; i < 30000; i++ {
		v := n.HGet("hset", strconv.Itoa(i))
		if string(v) != strconv.Itoa(i) {
			t.Errorf("HGet() = %s, want %v", v, strconv.Itoa(i))
		}
	}
	for i := 0; i < 10000; i++ {
		v := n.LPop("lpush")
		if string(v) != strconv.Itoa(39999-i) {
			t.Errorf("LPop() = %s, want %v", v, strconv.Itoa(9999-i))
		}
	}
}
