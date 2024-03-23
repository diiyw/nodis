package nodis

import (
	"os"
	"testing"
	"time"
)

func TestKey_Expire(t *testing.T) {
	_ = os.RemoveAll("testdata")
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

func TestKey_ExpireAt(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{
		Path:            "testdata",
		RecycleDuration: 60 * time.Second,
	})
	n.Set("test", []byte("test1"), 1)
	n.ExpireAt("test", time.Now().Add(2*time.Second))
	time.Sleep(2 * time.Second)
	v := n.Get("test")
	if v != nil {
		t.Errorf("Get() = %v, want %v", true, false)
	}
}

func TestKey_TTL(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{
		Path:            "testdata",
		RecycleDuration: 60 * time.Second,
	})
	n.Set("test", []byte("test1"), 1)
	n.Expire("test", 300)
	v := n.TTL("test")
	if v <= 0 {
		t.Errorf("TTL() = %v, want > %v", v, 0)
	}
	time.Sleep(2 * time.Second)
	v = n.TTL("test")
	if v < 0 {
		t.Errorf("TTL() = %v, want > %v", v, 0)
	}
}

func TestKey_Rename(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{
		Path:            "testdata",
		RecycleDuration: 60 * time.Second,
	})
	n.Set("test", []byte("test1"), 1)
	n.Rename("test", "test2")
	v := n.Get("test")
	if v != nil {
		t.Errorf("Get() = %v, want %v", true, false)
	}
	v = n.Get("test2")
	if v == nil {
		t.Errorf("Get() = %v, want %v", false, true)
	}
}

func TestKey_Keys(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{
		Path:            "testdata",
		RecycleDuration: 60 * time.Second,
	})
	n.Set("test1", []byte("test1"), 1)
	n.Set("test2", []byte("test2"), 1)
	n.Set("test3", []byte("test3"), 1)
	keys := n.Keys("test*")
	if len(keys) != 3 {
		t.Errorf("Keys() = %v, want %v", len(keys), 3)
	}
}

func TestKey_Type(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{
		Path:            "testdata",
		RecycleDuration: 60 * time.Second,
	})
	n.Set("test1", []byte("test1"), 0)
	n.LPush("test2", []byte("test2"))
	n.ZAdd("test3", "test3", 10)
	n.SAdd("test4", "test4")
	n.HSet("test5", "test5", []byte("test5"))
	if n.Type("test1") != "string" {
		t.Errorf("Type() = %v, want %v", n.Type("test1"), "string")
	}
	if n.Type("test2") != "list" {
		t.Errorf("Type() = %v, want %v", n.Type("test2"), "int")
	}
	if n.Type("test3") != "zset" {
		t.Errorf("Type() = %v, want %v", n.Type("test3"), "float64")
	}
	if n.Type("test4") != "set" {
		t.Errorf("Type() = %v, want %v", n.Type("test4"), "string")
	}
	if n.Type("test5") != "hash" {
		t.Errorf("Type() = %v, want %v", n.Type("test5"), "bool")
	}
	err := n.Close()
	if err != nil {
		t.Fatalf("Close() = %v, want %v", err, nil)
	}
	n = Open(&Options{
		Path:            "testdata",
		RecycleDuration: 60 * time.Second,
	})
	if n.Type("test1") != "string" {
		t.Errorf("Type() = %v, want %v", n.Type("test1"), "string")
	}
	if n.Type("test2") != "list" {
		t.Errorf("Type() = %v, want %v", n.Type("test2"), "int")
	}
	if n.Type("test3") != "zset" {
		t.Errorf("Type() = %v, want %v", n.Type("test3"), "float64")
	}
	if n.Type("test4") != "set" {
		t.Errorf("Type() = %v, want %v", n.Type("test4"), "string")
	}
	if n.Type("test5") != "hash" {
		t.Errorf("Type() = %v, want %v", n.Type("test5"), "bool")
	}
}

func TestKey_Scan(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{
		Path:            "testdata",
		RecycleDuration: 60 * time.Second,
	})
	n.Set("test1", []byte("test1"), 1)
	n.Set("test2", []byte("test2"), 1)
	n.Set("test3", []byte("test3"), 1)
	n.Set("test4", []byte("test4"), 1)
	n.Set("test5", []byte("test5"), 1)
	n.Set("test6", []byte("test6"), 1)
	n.Set("test7", []byte("test7"), 1)
	n.Set("test8", []byte("test8"), 1)
	n.Set("test9", []byte("test9"), 1)
	n.Set("test10", []byte("test10"), 1)
	n.Set("test11", []byte("test11"), 1)
	n.Set("test12", []byte("test12"), 1)
	n.Set("test13", []byte("test13"), 1)
	n.Set("test14", []byte("test14"), 1)
	n.Set("test15", []byte("test15"), 1)
	n.Set("test16", []byte("test16"), 1)
	n.Set("test17", []byte("test17"), 1)
	n.Set("test18", []byte("test18"), 1)
	n.Set("test19", []byte("test19"), 1)
	n.Set("test20", []byte("test20"), 1)
	n.Set("test21", []byte("test21"), 1)
	n.Set("test22", []byte("test22"), 1)
	n.Set("test23", []byte("test23"), 1)
	n.Set("test24", []byte("test24"), 1)
	n.Set("test25", []byte("test25"), 1)
	n.Set("test26", []byte("test26"), 1)
	n.Set("test27", []byte("test27"), 1)
	n.Set("test28", []byte("test28"), 1)
	n.Set("test29", []byte("test29"), 1)
	n.Set("test30", []byte("test30"), 1)
	n.Set("test31", []byte("test31"), 1)
	_, result := n.Scan(0, "test*", 10)
	if len(result) != 10 {
		t.Errorf("Scan() = %v, want %v", len(result), 10)
	}
	_, result = n.Scan(32, "test*", 10)
	if len(result) != 0 {
		t.Errorf("Scan() = %v, want %v", len(result), 0)
	}
	_, result = n.Scan(0, "test*", 32)
	if len(result) != 31 {
		t.Errorf("Scan() = %v, want %v", len(result), 31)
	}
	_, result = n.Scan(23, "test*", 10)
	if len(result) != 8 {
		t.Errorf("Scan() = %v, want %v", len(result), 8)
	}
}

func TestKey_Exists(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{
		Path:            "testdata",
		RecycleDuration: 60 * time.Second,
	})
	n.Set("test", []byte("test1"), 1)
	if !n.Exists("test") {
		t.Errorf("Exists() = %v, want %v", n.Exists("test"), true)
	}
	n.Del("test")
	if n.Exists("test") {
		t.Errorf("Exists() = %v, want %v", n.Exists("test"), false)
	}
}
