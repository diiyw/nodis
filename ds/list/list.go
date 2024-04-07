package list

import (
	"bytes"
	"time"

	"github.com/diiyw/nodis/ds"
)

type Node struct {
	data []byte
	next *Node
	prev *Node
}

type DoublyLinkedList struct {
	head   *Node
	tail   *Node
	length int64
}

// Type returns the type of the data structure
func (l *DoublyLinkedList) Type() ds.DataType {
	return ds.List
}

// NewDoublyLinkedList returns a new doubly linked list
func NewDoublyLinkedList() *DoublyLinkedList {
	return &DoublyLinkedList{}
}

// LPush adds an element to the head of the list
func (l *DoublyLinkedList) LPush(data ...[]byte) {
	l.lPush(data...)
}

func (l *DoublyLinkedList) lPush(data ...[]byte) {
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
func (l *DoublyLinkedList) RPush(data ...[]byte) {
	l.rPush(data...)
}

func (l *DoublyLinkedList) rPush(data ...[]byte) {
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
func (l *DoublyLinkedList) LPop(count int64) [][]byte {
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
func (l *DoublyLinkedList) RPop(count int64) [][]byte {
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
func (l *DoublyLinkedList) LRange(start, end int64) [][]byte {
	var result [][]byte
	currentNode := l.head
	var index int64 = 0
	if end < 0 {
		end = l.size() + end
	}
	for currentNode != nil {
		if index >= start && index <= end {
			result = append(result, currentNode.data)
		}
		if index > end {
			break
		}
		currentNode = currentNode.next
		index++
	}
	return result
}

func (l *DoublyLinkedList) size() int64 {
	currentNode := l.head
	var length int64 = 0
	for currentNode != nil {
		length++
		currentNode = currentNode.next
	}
	return length
}

// LLen returns the length of the list
func (l *DoublyLinkedList) LLen() int64 {
	return l.length
}

// BLPop removes and returns the first element of the list
func (l *DoublyLinkedList) BLPop(timeout time.Duration) []byte {
	if l.head == nil {
		time.Sleep(timeout)
	}
	result := l.LPop(1)
	if len(result) == 0 {
		return nil
	}
	return result[0]
}

// BRPop removes and returns the last element of the list
func (l *DoublyLinkedList) BRPop(timeout time.Duration) []byte {
	if l.tail == nil {
		time.Sleep(timeout)
	}
	result := l.RPop(1)
	if len(result) == 0 {
		return nil
	}
	return result[0]
}

// LIndex returns the element at index in the list
func (l *DoublyLinkedList) LIndex(index int64) []byte {
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
func (l *DoublyLinkedList) LInsert(pivot, data []byte, before bool) int64 {
	currentNode := l.head
	for currentNode != nil {
		if bytes.Contains(currentNode.data, pivot) {
			newNode := &Node{data: data}
			if before {
				newNode.next = currentNode
				newNode.prev = currentNode.prev
				if currentNode.prev != nil {
					currentNode.prev.next = newNode
				} else {
					l.head = newNode
				}
				currentNode.prev = newNode
			} else {
				newNode.prev = currentNode
				newNode.next = currentNode.next
				if currentNode.next != nil {
					currentNode.next.prev = newNode
				} else {
					l.tail = newNode
				}
				currentNode.next = newNode
			}
			return l.size()
		}
		currentNode = currentNode.next
	}
	return -1
}

// LPushX adds an element to the head of the list if the list exists
func (l *DoublyLinkedList) LPushX(data []byte) int64 {
	if l.head == nil {
		return 0
	}
	l.lPush(data)
	return l.size()
}

// RPushX adds an element to the end of the list if the list exists
func (l *DoublyLinkedList) RPushX(data []byte) int64 {
	if l.tail == nil {
		return 0
	}
	l.rPush(data)
	return l.size()
}

// LRem removes the first count occurrences of elements equal to value from the list
func (l *DoublyLinkedList) LRem(count int64, value []byte) int64 {
	currentNode := l.head
	var removed int64 = 0
	for currentNode != nil {
		if bytes.Equal(currentNode.data, value) {
			if count > 0 && removed == count {
				break
			}
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
			removed++
			l.length--
		}
		currentNode = currentNode.next
	}
	return removed
}

// LSet sets the list element at index to value
func (l *DoublyLinkedList) LSet(index int64, value []byte) bool {
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
func (l *DoublyLinkedList) LTrim(start, end int64) {
	currentNode := l.head
	var index int64 = 0
	for currentNode != nil {
		if index < start || index > end {
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
		}
		currentNode = currentNode.next
		index++
		l.length--
	}
}

// GetValue returns the byte slice of the list
func (l *DoublyLinkedList) GetValue() [][]byte {
	return l.LRange(0, -1)
}

// SetValue restores the list from the byte slice
func (l *DoublyLinkedList) SetValue(list [][]byte) {
	for _, item := range list {
		l.RPush(item)
	}
}
