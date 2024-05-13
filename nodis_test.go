package nodis

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/diiyw/nodis/pb"
)

func TestNodis_Open(t *testing.T) {
	opt := Options{
		Path:         "testdata",
		TidyDuration: 60 * time.Second,
		FileSize:     FileSizeGB,
	}
	got := Open(&opt)
	if got == nil {
		t.Errorf("Open() = %v, want %v", got, "Nodis{}")
	}
}

func TestNodis_OpenAndCloseBigdata10000(t *testing.T) {
	_ = os.RemoveAll("testdata")
	opt := Options{
		Path:         "testdata",
		TidyDuration: 60 * time.Second,
		FileSize:     FileSizeGB,
	}
	n := Open(&opt)
	for i := 0; i < 10000; i++ {
		is := strconv.Itoa(i)
		n.Set(is, []byte(is))
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
	err := n.Close()
	if err != nil {
		t.Errorf("Close() = %v, want %v", err, nil)
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
		v, _ := n.ZScore("zset", strconv.Itoa(i))
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
		v := n.LPop("lpush", 1)[0]
		if string(v) != strconv.Itoa(39999-i) {
			t.Errorf("LPop() = %s, want %v", v, strconv.Itoa(9999-i))
		}
	}
}

func TestNodis_Snapshot(t *testing.T) {
	_ = os.RemoveAll("testdata")
	opt := &Options{
		Path: "testdata",
	}
	n := Open(opt)
	n.Set("test", []byte("test"))
	n.Snapshot("testdata")
	_ = n.Close()
}

func TestNodis_SnapshotChanged(t *testing.T) {
	_ = os.RemoveAll("testdata")
	opt := &Options{
		Path: "testdata",
	}
	n := Open(opt)
	n.Set("test", []byte("test"))
	n.Snapshot(opt.Path)
	time.Sleep(time.Second)
	n.Set("test", []byte("test_new"))
	n.Snapshot(opt.Path)
	_ = n.Close()
}

func TestNodis_Tidy(t *testing.T) {
	_ = os.RemoveAll("testdata")
	opt := &Options{
		Path:         "testdata",
		TidyDuration: time.Second,
	}
	n := Open(opt)
	n.SetEX("test", []byte("test"), 1)
	time.Sleep(2 * time.Second)
	v := n.Get("test")
	if v != nil {
		t.Errorf("Get() = %v, want %v", v, nil)
	}
	n.SetEX("test", []byte("test"), 5)
	// load from storage
	time.Sleep(2 * time.Second)
	v = n.Get("test")
	if v == nil {
		t.Errorf("Get() = %v, want %v", v, "test")
	}
	_ = n.Close()
}

func TestNodis_Clear(t *testing.T) {
	_ = os.RemoveAll("testdata")
	opt := &Options{
		Path: "testdata",
	}
	n := Open(opt)
	n.Set("test", []byte("test"))
	n.Clear()
	v := n.Get("test")
	if v != nil {
		t.Errorf("Get() = %v, want %v", v, nil)
	}
	_ = n.Close()
}

func TestNodis_Patch(t *testing.T) {
	_ = os.RemoveAll("testdata")
	opt := &Options{
		Path: "testdata",
	}
	var ops = []*pb.Op{
		{
			Operation: &pb.Operation{
				Type:  pb.OpType_Set,
				Key:   "string",
				Value: []byte("string"),
			},
		},
		{
			Operation: &pb.Operation{
				Type:   pb.OpType_ZAdd,
				Key:    "zset",
				Member: "zset",
				Score:  10,
			},
		},
		{
			Operation: &pb.Operation{
				Type:   pb.OpType_LPush,
				Key:    "lpush",
				Values: [][]byte{[]byte("lpush3"), []byte("lpush2"), []byte("lpush")},
			},
		},
		{
			Operation: &pb.Operation{
				Type:    pb.OpType_SAdd,
				Key:     "set",
				Members: []string{"member", "member2", "member3"},
			},
		},
		{
			Operation: &pb.Operation{
				Type:  pb.OpType_HSet,
				Key:   "hash",
				Field: "field",
				Value: []byte("value"),
			},
		},
	}
	n := Open(opt)
	n.Patch(ops...)
	v := n.Get("string")
	if string(v) != "string" {
		t.Errorf("Get() = %v, want %v", v, "test_new")
	}
	s, _ := n.ZScore("zset", "zset")
	if s != 10 {
		t.Errorf("ZScore() = %v, want %v", s, 10)
	}
	v = n.LPop("lpush", 1)[0]
	if string(v) != "lpush" {
		t.Errorf("LPop() = %s, want %v", v, "lpush")
	}
	b := n.SIsMember("set", "member")
	if !b {
		t.Errorf("SIsMember() = %v, want %v", b, true)
	}
	v = n.HGet("hash", "field")
	if string(v) != "value" {
		t.Errorf("HGet() = %v, want %v", v, "value")
	}
}

func TestNodis_SetEntity(t *testing.T) {
	_ = os.RemoveAll("testdata")
	opt := &Options{
		Path: "testdata",
	}
	n := Open(opt)
	n.Set("test", []byte("test"))
	data := n.GetEntry("test")
	n.Clear()
	v := n.Get("test")
	if v != nil {
		t.Errorf("Get() = %v, want %v", v, nil)
		return
	}
	err := n.SetEntry(data)
	if err != nil {
		t.Errorf("SetEntity() = %v, want %v", err, nil)
	}
	v = n.Get("test")
	if string(v) != "test" {
		t.Errorf("Get() = %s, want %v", v, "test")
	}
}

func TestNodis_Stick(t *testing.T) {
	_ = os.RemoveAll("testdata")
	opt := &Options{
		Path: "testdata",
	}
	n := Open(opt)
	n.Set("test", []byte("test"))
	n.Stick([]string{"test"}, func(op *pb.Operation) {
		if op == nil || string(op.Value) != "test_new" {
			t.Errorf("Stick() = %v, want %v", op, "test_new")
		}
	})
	n.Set("test", []byte("test_new"))
	time.Sleep(time.Second)
}

func TestNodis_UnStick(t *testing.T) {
	_ = os.RemoveAll("testdata")
	opt := &Options{
		Path: "testdata",
	}
	n := Open(opt)
	n.Set("test", []byte("test"))
	id := n.Stick([]string{"test"}, func(op *pb.Operation) {
		if op == nil || string(op.Value) != "test_new" {
			t.Errorf("Stick() = %v, want %v", op, "test_new")
		}
	})
	n.UnStick(id)
	n.Set("test", []byte("test_new"))
	time.Sleep(time.Second)
}

func TestNodis_OpenAndClose(t *testing.T) {
	_ = os.RemoveAll("testdata")
	opt := &Options{
		Path: "testdata",
	}
	n := Open(opt)
	n.Set("str", []byte("set"))
	n.ZAdd("zset", "zset", 1)
	n.HSet("hset", "hset", []byte("hset"))
	n.LPush("lpush", []byte("lpush"))
	n.SAdd("set", "set")
	_ = n.Close()
	n = Open(opt)
	v := n.Get("str")
	if v == nil {
		t.Errorf("Get() = %s, want %v", v, "set")
	}
	s, _ := n.ZScore("zset", "zset")
	if s != 1 {
		t.Errorf("ZScore() = %v, want %v", v, 1)
	}
	v = n.HGet("hset", "hset")
	if v == nil {
		t.Errorf("HGet() = %s, want %v", v, "hset")
	}
	if !n.SIsMember("set", "set") {
		t.Errorf("SIsMember() = %v, want %v", false, true)
	}
	v = n.LPop("lpush", 1)[0]
	if v == nil {
		t.Errorf("LPop() = %s, want %v", v, "lpush")
	}
}
