package list

import (
	"bytes"
	"encoding/binary"

	"github.com/diiyw/nodis/ds"
)

type Node struct {
	data []byte
	next *Node
	prev *Node
}

// LinkedList is a doubly linked list
type LinkedList struct {
	head   *Node
	tail   *Node
	length int64
}

// Type returns the type of the data structure
func (l *LinkedList) Type() ds.ValueType {
	return ds.List
}

// NewLinkedList returns a new doubly linked list
func NewLinkedList() *LinkedList {
	return &LinkedList{}
}

// LPush adds an element to the head of the list
func (l *LinkedList) LPush(data ...[]byte) {
	for _, datum := range data {
		newNode := &Node{data: datum}
		if l.head == nil {
			l.head = newNode
			l.tail = newNode
		} else {
			newNode.next = l.head
			l.head.prev = newNode
			l.head = newNode
		}
		l.length++
	}
}

// RPush adds an element to the end of the list
func (l *LinkedList) RPush(data ...[]byte) {
	for _, datum := range data {
		newNode := &Node{data: datum}
		if l.head == nil {
			l.head = newNode
			l.tail = newNode
		} else {
			l.tail.next = newNode
			newNode.prev = l.tail
			l.tail = newNode
		}
		l.length++
	}
}

// LPop returns the first element of the list
func (l *LinkedList) LPop(count int64) [][]byte {
	if l.head == nil {
		return nil // 链表为空
	}
	var result [][]byte
	for i := int64(0); i < count; i++ {
		if l.head == nil {
			break
		}
		result = append(result, l.head.data)
		l.head = l.head.next
		if l.head != nil {
			l.head.prev = nil
		} else {
			l.tail = nil
		}
		l.length--
	}
	return result
}

// RPop returns the last element of the list
func (l *LinkedList) RPop(count int64) [][]byte {
	if l.tail == nil {
		return nil
	}
	var result [][]byte
	for i := int64(0); i < count; i++ {
		if l.tail == nil {
			break
		}
		result = append(result, l.tail.data)
		l.tail = l.tail.prev
		if l.tail != nil {
			l.tail.next = nil
		} else {
			l.head = nil
		}
		l.length--
	}
	return result
}

// LRange returns a range of elements from the list
func (l *LinkedList) LRange(start, end int64) [][]byte {
	var result [][]byte
	l.forEach(start, end, func(v []byte) {
		result = append(result, v)
	})
	return result
}

// LRange returns a range of elements from the list
func (l *LinkedList) forEach(start, end int64, fn func(v []byte)) {
	currentNode := l.head
	var index int64 = 0
	if start != 0 && start >= end {
		return
	}
	if start < 0 {
		start = l.size() + start
	}
	if end < 0 {
		end = l.size() + end
	}
	for currentNode != nil {
		if index >= start && index <= end {
			fn(currentNode.data)
		}
		if index > end {
			break
		}
		currentNode = currentNode.next
		index++
	}
}

func (l *LinkedList) size() int64 {
	currentNode := l.head
	var length int64 = 0
	for currentNode != nil {
		length++
		currentNode = currentNode.next
	}
	return length
}

// LLen returns the length of the list
func (l *LinkedList) LLen() int64 {
	return l.length
}

// LIndex returns the element at index in the list
func (l *LinkedList) LIndex(index int64) []byte {
	if index < 0 {
		index = l.length + index
	}
	currentNode := l.head
	var currentIndex int64 = 0
	for currentNode != nil {
		if currentIndex == index {
			return currentNode.data
		}
		currentNode = currentNode.next
		currentIndex++
	}
	return nil
}

// LInsert inserts the element before or after the pivot element
// the list length after a successful insert operation.
// 0 when the key doesn't exist.
// -1 when the pivot wasn't found.
func (l *LinkedList) LInsert(pivot, data []byte, before bool) int64 {
	currentNode := l.head
	for currentNode != nil {
		if bytes.Equal(currentNode.data, pivot) {
			newNode := &Node{data: data}
			if before {
				if currentNode.prev != nil {
					currentNode.prev.next = newNode
					newNode.prev = currentNode.prev
				} else {
					l.head = newNode
				}
				newNode.next = currentNode
				currentNode.prev = newNode
			} else {
				if currentNode.next != nil {
					currentNode.next.prev = newNode
					newNode.next = currentNode.next
				} else {
					l.tail = newNode
				}
				newNode.prev = currentNode
				currentNode.next = newNode
			}
			l.length++
			return l.length
		}
		currentNode = currentNode.next
	}
	return -1
}

// LRem removes the first count occurrences of elements equal to value from the list
// Removes the first count occurrences of elements equal to element from the list stored at key. The count argument influences the operation in the following ways:
// count > 0: Remove elements equal to element moving from head to tail.
// count < 0: Remove elements equal to element moving from tail to head.
// count = 0: Remove all elements equal to element.
func (l *LinkedList) LRem(count int64, value []byte) int64 {
	var removed int64
	if count > 0 {
		removed = l.lRem(count, value)
	} else if count < 0 {
		removed = l.lRevRem(-count, value)
	} else {
		removed = l.lRemAll(value)
	}
	return removed
}

// lRemAll removes all elements equal to value from the list
func (l *LinkedList) lRemAll(value []byte) int64 {
	var removed int64
	currentNode := l.head
	for currentNode != nil {
		if bytes.Equal(currentNode.data, value) {
			if currentNode.prev != nil {
				currentNode.prev.next = currentNode.next
			} else {
				l.head = currentNode.next
			}
			if currentNode.next != nil {
				currentNode.next.prev = currentNode.prev
			} else {
				l.tail = currentNode.prev
			}
			l.length--
			removed++
		}
		currentNode = currentNode.next
	}
	return removed
}

func (l *LinkedList) lRem(count int64, value []byte) int64 {
	var removed int64
	currentNode := l.head
	for currentNode != nil {
		if bytes.Equal(currentNode.data, value) {
			if count == 0 || removed < count {
				if currentNode.prev != nil {
					currentNode.prev.next = currentNode.next
				} else {
					l.head = currentNode.next
				}
				if currentNode.next != nil {
					currentNode.next.prev = currentNode.prev
				} else {
					l.tail = currentNode.prev
				}
				l.length--
				removed++
			}
		}
		currentNode = currentNode.next
	}
	return removed
}

func (l *LinkedList) lRevRem(count int64, value []byte) int64 {
	var removed int64
	currentNode := l.tail
	for currentNode != nil {
		if bytes.Equal(currentNode.data, value) {
			if count == 0 || removed < count {
				if currentNode.next != nil {
					currentNode.next.prev = currentNode.prev
				} else {
					l.tail = currentNode.prev
				}
				if currentNode.prev != nil {
					currentNode.prev.next = currentNode.next
				} else {
					l.head = currentNode.next
				}
				l.length--
				removed++
			}
		}
		currentNode = currentNode.prev
	}
	return removed
}

// LSet sets the list element at index to value
func (l *LinkedList) LSet(index int64, value []byte) bool {
	currentNode := l.head
	var currentIndex int64 = 0
	for currentNode != nil {
		if currentIndex == index {
			currentNode.data = value
			return true
		}
		currentNode = currentNode.next
		currentIndex++
	}
	return false
}

// LTrim trims an existing list so that it will contain only the specified range of elements specified
// For example: LTRIM foobar 0 2 will modify the list stored at foobar so that only the first three elements of the list will remain.
// start and end can also be negative numbers indicating offsets from the end of the list, where -1 is the last element of the list, -2 the penultimate element and so on.
// Out of range indexes will not produce an error: if start is larger than the end of the list, or start > end, the result will be an empty list (which causes key to be removed). If end is larger than the end of the list, Redis will treat it like the last element of the list.
func (l *LinkedList) LTrim(start, end int64) {
	currentNode := l.head
	var currentIndex int64 = 0
	if end < 0 {
		end = l.size() + end
	}
	for currentNode != nil {
		if currentIndex < start || currentIndex > end {
			if currentNode.prev != nil {
				currentNode.prev.next = currentNode.next
			} else {
				l.head = currentNode.next
			}
			if currentNode.next != nil {
				currentNode.next.prev = currentNode.prev
			} else {
				l.tail = currentNode.prev
			}
			l.length--
		}
		currentNode = currentNode.next
		currentIndex++
	}
}

// GetValue returns the byte slice of the list
func (l *LinkedList) GetValue() []byte {
	var list []byte
	l.forEach(0, -1, func(v []byte) {
		var b = make([]byte, len(v)+8)
		var vLen = len(v)
		n := binary.PutVarint(b, int64(vLen))
		copy(b[n:], v)
		list = append(list, b[:n+vLen]...)
	})
	return list
}

// SetValue restores the list from the byte slice
func (l *LinkedList) SetValue(list []byte) {
	for {
		if len(list) == 0 {
			break
		}
		vLen, n := binary.Varint(list)
		if n == 0 {
			break
		}
		v := list[n : n+int(vLen)]
		list = list[n+int(vLen):]
		l.RPush(v)
	}
}
