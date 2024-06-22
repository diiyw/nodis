package list

import (
	"strconv"
	"testing"
)

func TestList_LPush(t *testing.T) {
	l := NewLinkedList()
	for i := 0; i < 100; i++ {
		is := strconv.Itoa(i)
		l.LPush([]byte(is))
	}
	if l.LLen() != 100 {
		t.Errorf("len error")
	}
}

func TestList_RPush(t *testing.T) {
	l := NewLinkedList()
	for i := 0; i < 100; i++ {
		is := strconv.Itoa(i)
		l.LPush([]byte(is))
	}
	if l.LLen() != 100 {
		t.Errorf("len error")
	}
}

func TestList_LPop(t *testing.T) {
	l := NewLinkedList()
	for i := 0; i < 100; i++ {
		is := strconv.Itoa(i)
		l.LPush([]byte(is))
	}
	for i := 0; i < 100; i++ {
		v := l.LPop(1)
		if i == 0 {
			if string(v[0]) != "99" {
				t.Errorf("pop error expect 0 go %d", v)
			}
		}
		is := strconv.Itoa(99 - i)
		if string(v[0]) != is {
			t.Errorf("pop error")
		}
	}
}

func TestList_RPop(t *testing.T) {
	l := NewLinkedList()
	for i := 0; i < 10; i++ {
		is := strconv.Itoa(i)
		l.LPush([]byte(is))
	}
	v := l.RPop(1)
	if string(v[0]) != "0" {
		t.Errorf("pop error expect 0 go %d", v)
	}
}

func TestList_LLen(t *testing.T) {
	l := NewLinkedList()
	for i := 0; i < 10; i++ {
		is := strconv.Itoa(i)
		l.LPush([]byte(is))
	}
	if l.LLen() != 10 {
		t.Errorf("len error")
	}
}

func TestList_LIndex(t *testing.T) {
	l := NewLinkedList()
	for i := 0; i < 10; i++ {
		is := strconv.Itoa(i)
		l.LPush([]byte(is))
	}
	v := l.LIndex(0)
	if string(v) != "9" {
		t.Errorf("index error expect 9 go %d", v)
	}
}

func TestList_LRange(t *testing.T) {
	l := NewLinkedList()
	for i := 0; i < 10; i++ {
		is := strconv.Itoa(i)
		l.LPush([]byte(is))
	}
	v := l.LRange(0, 5)
	if len(v) != 6 {
		t.Errorf("range error")
	}
}

func TestList_LInsert(t *testing.T) {
	l := NewLinkedList()
	for i := 0; i < 10; i++ {
		is := strconv.Itoa(i)
		l.LPush([]byte(is))
	}
	l.LInsert([]byte("5"), []byte("100"), false)
	v := l.LIndex(5)
	if string(v) != "100" {
		t.Errorf("insert error")
	}
}

func TestList_LRem(t *testing.T) {
	l := NewLinkedList()
	for i := 0; i < 10; i++ {
		is := strconv.Itoa(i)
		l.LPush([]byte(is))
	}
	l.LRem(5, []byte("5"))
	if l.LLen() != 9 {
		t.Errorf("rem error expect 9 go %d", l.LLen())
	}
	v := l.LIndex(0)
	if string(v) != "9" {
		t.Errorf("rem error expect 9 go %d", v)
	}
}

func TestList_LSet(t *testing.T) {
	l := NewLinkedList()
	for i := 0; i < 10; i++ {
		is := strconv.Itoa(i)
		l.LPush([]byte(is))
	}
	l.LSet(5, []byte("100"))
	v := l.LIndex(5)
	if string(v) != "100" {
		t.Errorf("set error")
	}
}

func TestList_LTrim(t *testing.T) {
	l := NewLinkedList()
	for i := 0; i < 10; i++ {
		is := strconv.Itoa(i)
		l.LPush([]byte(is))
	}
	l.LTrim(0, 5)
	if l.LLen() != 6 {
		t.Errorf("trim error")
	}
}

func BenchmarkDoublyLinkedList_LPush(b *testing.B) {
	l := NewLinkedList()
	for i := 0; i < b.N; i++ {
		is := strconv.Itoa(i)
		l.LPush([]byte(is))
	}
}

func BenchmarkDoublyLinkedList_LPop(b *testing.B) {
	l := NewLinkedList()
	for i := 0; i < b.N; i++ {
		is := strconv.Itoa(i)
		l.LPush([]byte(is))
	}
	for i := 0; i < b.N; i++ {
		l.LPop(1)
	}
}

func TestGetSetValue(t *testing.T) {
	l := NewLinkedList()
	for i := 0; i < 10; i++ {
		is := strconv.Itoa(i)
		l.LPush([]byte(is))
	}
	v := l.GetValue()
	l2 := NewLinkedList()
	l2.SetValue(v)
	if l.LLen() != l2.LLen() {
		t.Errorf("get set value error")
	}
	l2Pop := l2.LPop(10)
	for i, it := range l.LPop(10) {
		if string(it) != string(l2Pop[i]) {
			t.Errorf("get set value error")
		}
	}
}
