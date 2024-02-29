package list

import "testing"

func TestList_LPush(t *testing.T) {
	l := NewDoublyLinkedList()
	for i := 0; i < 100; i++ {
		l.LPush(i)
	}
	if l.LLen() != 100 {
		t.Errorf("len error")
	}
}

func TestList_RPush(t *testing.T) {
	l := NewDoublyLinkedList()
	for i := 0; i < 100; i++ {
		l.RPush(i)
	}
	if l.LLen() != 100 {
		t.Errorf("len error")
	}
}

func TestList_LPop(t *testing.T) {
	l := NewDoublyLinkedList()
	for i := 0; i < 100; i++ {
		l.LPush(i)
	}
	for i := 0; i < 100; i++ {
		v := l.LPop()
		if i == 0 {
			if v != 99 {
				t.Errorf("pop error expect 0 go %d", v)
			}
		}
		if v != 99-i {
			t.Errorf("pop error")
		}
	}
}

func TestList_RPop(t *testing.T) {
	l := NewDoublyLinkedList()
	for i := 0; i < 10; i++ {
		l.LPush(i)
	}
	v := l.RPop()
	if v != 0 {
		t.Errorf("pop error expect 0 go %d", v)
	}
}

func TestList_LLen(t *testing.T) {
	l := NewDoublyLinkedList()
	for i := 0; i < 10; i++ {
		l.LPush(i)
	}
	if l.LLen() != 10 {
		t.Errorf("len error")
	}
}

func TestList_LIndex(t *testing.T) {
	l := NewDoublyLinkedList()
	for i := 0; i < 10; i++ {
		l.LPush(i)
	}
	v, ok := l.LIndex(0)
	if !ok || v != 9 {
		t.Errorf("index error expect 9 go %d", v)
	}
}

func TestList_LRange(t *testing.T) {
	l := NewDoublyLinkedList()
	for i := 0; i < 10; i++ {
		l.LPush(i)
	}
	v := l.LRange(0, 5)
	if len(v) != 6 {
		t.Errorf("range error")
	}
}

func TestList_LInsert(t *testing.T) {
	l := NewDoublyLinkedList()
	for i := 0; i < 10; i++ {
		l.LPush(i)
	}
	l.LInsert(5, 100, false)
	v, ok := l.LIndex(5)
	if !ok || v != 100 {
		t.Errorf("insert error")
	}
}

func TestList_LPushx(t *testing.T) {
	l := NewDoublyLinkedList()
	l.LPushx(100)
	if l.LLen() != 0 {
		t.Errorf("pushx error")
	}
}

func TestList_RPushx(t *testing.T) {
	l := NewDoublyLinkedList()
	l.RPushx(100)
	if l.LLen() != 0 {
		t.Errorf("pushx error")
	}
}

func TestList_LRem(t *testing.T) {
	l := NewDoublyLinkedList()
	for i := 0; i < 10; i++ {
		l.LPush(i)
	}
	l.LRem(5, 0)
	if l.LLen() != 9 {
		t.Errorf("rem error expect 9 go %d", l.LLen())
	}
	v, ok := l.LIndex(0)
	if !ok || v != 9 {
		t.Errorf("rem error expect 9 go %d", v)
	}
}

func TestList_LSet(t *testing.T) {
	l := NewDoublyLinkedList()
	for i := 0; i < 10; i++ {
		l.LPush(i)
	}
	l.LSet(5, 100)
	v, ok := l.LIndex(5)
	if !ok || v != 100 {
		t.Errorf("set error")
	}
}

func TestList_LTrim(t *testing.T) {
	l := NewDoublyLinkedList()
	for i := 0; i < 10; i++ {
		l.LPush(i)
	}
	l.LTrim(0, 5)
	if l.LLen() != 6 {
		t.Errorf("trim error")
	}
}
