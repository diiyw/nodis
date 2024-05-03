package nodis

import (
	"os"
	"testing"
)

func TestHash_HSet(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{Path: "testdata"})
	n.HSet("hash", "field", []byte("value"))
	if n.HGet("hash", "field") == nil {
		t.Error("HSet failed")
	}
}

func TestHash_HGet(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{Path: "testdata"})
	n.HSet("hash", "field", []byte("value"))
	if string(n.HGet("hash", "field")) != "value" {
		t.Error("HGet failed")
	}
}

func TestHash_HDel(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{Path: "testdata"})
	n.HSet("hash", "field", []byte("value"))
	n.HDel("hash", "field")
	if n.HGet("hash", "field") != nil {
		t.Error("HDel failed")
	}
}

func TestHash_HLen(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{Path: "testdata"})
	n.HSet("hash", "field", []byte("value"))
	if n.HLen("hash") != 1 {
		t.Error("HLen failed")
	}
}

func TestHash_HKeys(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{Path: "testdata"})
	n.HSet("hash", "field", []byte("value"))
	if n.HKeys("hash")[0] != "field" {
		t.Error("HKeys failed")
	}
}

func TestHash_HExists(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{Path: "testdata"})
	n.HSet("hash", "field", []byte("value"))
	if !n.HExists("hash", "field") {
		t.Error("HExists failed")
	}
}

func TestHash_HGetAll(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{Path: "testdata"})
	n.HSet("hash", "field", []byte("value"))
	if string(n.HGetAll("hash")["field"]) != "value" {
		t.Error("HGetAll failed")
	}
}

func TestHash_HSetNX(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{Path: "testdata"})
	n.HSetNX("hash", "field", []byte("value"))
	if v := n.HSetNX("hash", "field", []byte("value")); v == 1 {
		t.Error("HSetNX failed")
	}
}

func TestHash_HMSet(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{Path: "testdata"})
	n.HMSet("hash", map[string][]byte{"field": []byte("value")})
	if string(n.HGet("hash", "field")) != "value" {
		t.Error("HMSet failed")
	}
}

func TestHash_HMGet(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{Path: "testdata"})
	n.HMSet("hash", map[string][]byte{"field": []byte("value")})
	if string(n.HMGet("hash", "field")[0]) != "value" {
		t.Error("HMGet failed expected value got", string(n.HMGet("hash", "field")[0]))
	}
}

func TestHash_HIncrBy(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{Path: "testdata"})
	n.HSet("hash", "field", []byte("1"))
	if v, err := n.HIncrBy("hash", "field", 1); v != 2 || err != nil {
		t.Error("HIncrBy failed")
	}
}

func TestHash_HIncrByFloat(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{Path: "testdata"})
	n.HSet("hash", "field", []byte("1"))
	if v, err := n.HIncrByFloat("hash", "field", 1.1); v != 2.1 || err != nil {
		t.Error("HIncrByFloat failed")
	}
}

func TestHash_HClear(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{Path: "testdata"})
	n.HSet("hash", "field", []byte("value"))
	n.HClear("hash")
	if n.HGet("hash", "field") != nil {
		t.Error("HClear failed")
	}
}

func TestHash_HScan(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{Path: "testdata"})
	n.HSet("hash", "field", []byte("value"))
	l, data := n.HScan("hash", 0, "*", 0)
	if l != 1 {
		t.Error("HScan failed expected 1 got", l)
	}
	if string(data["field"]) != "value" {
		t.Error("HScan failed expected value got", string(data["field"]))
	}
}

func TestHash_HVals(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{Path: "testdata"})
	n.HSet("hash", "field", []byte("value"))
	if len(n.HVals("hash")) != 1 {
		t.Error("HVars failed")
	}
}
