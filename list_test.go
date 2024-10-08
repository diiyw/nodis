package nodis

import (
	"os"
	"testing"
	"time"
)

func TestList_LPush(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{})
	n.LPush("list", []byte("value"))
	n.LPush("list", []byte("value1"))
	n.LPush("list", []byte("value2"))
	n.LPush("list", []byte("value3"))
	if string(n.LPop("list", 1)[0]) != "value3" {
		t.Error("LPush failed")
	}
	if n.LLen("list") != 3 {
		t.Error("LPush failed")
	}
	v := n.LRange("list", 0, -1)
	if string(v[0]) != "value2" || string(v[1]) != "value1" || string(v[2]) != "value" {
		t.Error("LPush failed")
	}
}

func TestList_RPush(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{})
	n.RPush("list", []byte("value"))
	if string(n.RPop("list", 1)[0]) != "value" {
		t.Error("RPush failed")
	}
}

func TestList_LPop(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{})
	n.LPush("list", []byte("value"))
	if string(n.LPop("list", 1)[0]) != "value" {
		t.Error("LPop failed")
	}
}

func TestList_RPop(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{})
	n.RPush("list", []byte("value"))
	if string(n.RPop("list", 1)[0]) != "value" {
		t.Error("RPop failed")
	}
}

func TestList_LLen(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{})
	n.LPush("list", []byte("value"))
	if n.LLen("list") != 1 {
		t.Error("LLen failed")
	}
}

func TestList_LRange(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{})
	n.LPush("list", []byte("value"))
	if string(n.LRange("list", 0, -1)[0]) != "value" {
		t.Error("LRange failed")
	}
}

func TestList_LIndex(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{})
	n.LPush("list", []byte("value"))
	if string(n.LIndex("list", 0)) != "value" {
		t.Error("LIndex failed")
	}
}

func TestList_BLPopNoPush(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{})
	k, _ := n.BLPop(time.Second, "list")
	if k != "" {
		t.Error("BLPop failed excepted value got nil")
	}
}

func TestList_BLPop(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{})
	go func() {
		time.Sleep(time.Second)
		n.LPush("list", []byte("value"))
	}()
	_, v := n.BLPop(time.Second*2, "list")
	if string(v) != "value" {
		t.Error("BLPop failed excepted value got nil")
	}
}

func TestList_BRPopNoPush(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{})
	k, _ := n.BRPop(time.Second, "list")
	if k != "" {
		t.Error("BRPop failed")
	}
}

func TestList_BRPop(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{})
	go func() {
		n.RPush("list", []byte("value"))
	}()
	_, v := n.BRPop(time.Second, "list")
	if string(v) != "value" {
		t.Error("BRPop failed")
	}
}

func TestList_LSet(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{})
	n.LPush("list", []byte("value"))
	n.LSet("list", 0, []byte("new"))
	if string(n.LIndex("list", 0)) != "new" {
		t.Error("LSet failed")
	}
}

func TestList_LTrim(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{})
	n.LPush("list", []byte("value"))
	n.LTrim("list", 0, 0)
	if string(n.LIndex("list", 0)) != "value" {
		t.Error("LTrim failed")
	}
}

func TestList_LRem(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{})
	n.LPush("list", []byte("value"))
	n.LRem("list", []byte("value"), 0)
	if n.LLen("list") != 0 {
		t.Error("LRem failed")
	}
}

func TestList_LInsert(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{})
	n.LPush("list", []byte("value"))
	n.LInsert("list", []byte("value"), []byte("new"), true)
	if string(n.LIndex("list", 0)) != "new" {
		t.Error("LInsert failed")
	}
}

func TestList_LPushX(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{})
	n.LPushX("list", []byte("value"))
	v := n.LPop("list", 1)
	if len(v) > 0 && string(v[0]) == "value" {
		t.Error("LPushX failed excepted value nil")
	}
}

func TestList_RPushX(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{})
	n.RPushX("list", []byte("value"))
	v := n.RPop("list", 1)
	if len(v) > 0 && string(v[0]) == "value" {
		t.Error("RPushX failed excepted value nil")
	}
}

func TestList_LPopRPush(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{})
	n.LPush("list", []byte("value"))
	n.LPush("list2", []byte("value2"))
	if string(n.LPopRPush("list", "list2")) != "value" {
		t.Error("LPopRPush failed")
	}
	if n.LLen("list2") != 2 {
		t.Error("LPopRPush failed expected list length 2 got 1")
	}
}

func TestList_RPopLPush(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{})
	n.RPush("list", []byte("value"))
	n.RPush("list2", []byte("value"))
	if string(n.RPopLPush("list", "list2")) != "value" {
		t.Error("RPopLPush failed")
	}
	if n.LLen("list2") != 2 {
		t.Error("LPopRPush failed expected list length 2 got 1")
	}
}

func TestList_LPopRPush2(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{})
	n.LPush("list", []byte("value"))
	if string(n.LPopRPush("list", "list2")) != "value" {
		t.Error("LPopRPush failed")
	}
	if n.LLen("list2") != 1 {
		t.Error("LPopRPush failed expected list length 1")
	}
}

func TestList_RPopLPush2(t *testing.T) {
	_ = os.RemoveAll("testdata")
	n := Open(&Options{})
	n.RPush("list", []byte("value"))
	if string(n.RPopLPush("list", "list2")) != "value" {
		t.Error("RPopLPush failed")
	}
	if n.LLen("list2") != 1 {
		t.Error("LPopRPush failed expected list length 1")
	}
}
