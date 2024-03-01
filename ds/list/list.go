package list

import (
	"bytes"
	"sync"
	"time"

	"github.com/kelindar/binary"
)

type Node struct {
	data []byte
	next *Node
	prev *Node
}

type DoublyLinkedList struct {
	sync.RWMutex
	head *Node
	tail *Node
}

// GetType returns the type of the data structure
func (l *DoublyLinkedList) GetType() string {
	return "list"
}

// NewDoublyLinkedList returns a new doubly linked list
func NewDoublyLinkedList() *DoublyLinkedList {
	return &DoublyLinkedList{}
}

// LPush adds an element to the head of the list
func (l *DoublyLinkedList) LPush(data []byte) {
	l.Lock()
	defer l.Unlock()
	newNode := &Node{data: data}
	if l.head == nil {
		l.head = newNode
		l.tail = newNode
	} else {
		newNode.next = l.head
		l.head.prev = newNode
		l.head = newNode
	}
}

// RPush adds an element to the end of the list
func (l *DoublyLinkedList) RPush(data []byte) {
	l.Lock()
	defer l.Unlock()
	newNode := &Node{data: data}
	if l.head == nil {
		l.head = newNode
		l.tail = newNode
	} else {
		l.tail.next = newNode
		newNode.prev = l.tail
		l.tail = newNode
	}
}

// LPop returns the first element of the list
func (l *DoublyLinkedList) LPop() []byte {
	l.Lock()
	defer l.Unlock()
	if l.head == nil {
		return nil // 链表为空
	}
	data := l.head.data
	l.head = l.head.next
	if l.head != nil {
		l.head.prev = nil
	} else {
		l.tail = nil
	}
	return data
}

// RPop returns the last element of the list
func (l *DoublyLinkedList) RPop() []byte {
	l.Lock()
	defer l.Unlock()
	if l.tail == nil {
		return nil // 链表为空
	}
	data := l.tail.data
	l.tail = l.tail.prev
	if l.tail != nil {
		l.tail.next = nil
	} else {
		l.head = nil
	}
	return data
}

// LRange returns a range of elements from the list
func (l *DoublyLinkedList) LRange(start, end int) [][]byte {
	l.RLock()
	defer l.RUnlock()
	var result [][]byte
	currentNode := l.head
	index := 0
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

func (l *DoublyLinkedList) size() int {
	currentNode := l.head
	length := 0
	for currentNode != nil {
		length++
		currentNode = currentNode.next
	}
	return length
}

// LLen returns the length of the list
func (l *DoublyLinkedList) LLen() int {
	l.RLock()
	defer l.RUnlock()
	return l.size()
}

// BLPop removes and returns the first element of the list
func (l *DoublyLinkedList) BLPop(timeout time.Duration) []byte {
	if l.head == nil {
		time.Sleep(timeout)
	}
	return l.LPop()
}

// BRPop removes and returns the last element of the list
func (l *DoublyLinkedList) BRPop(timeout time.Duration) []byte {
	if l.tail == nil {
		time.Sleep(timeout)
	}
	return l.RPop()
}

// LIndex returns the element at index in the list
func (l *DoublyLinkedList) LIndex(index int) ([]byte, bool) {
	l.RLock()
	defer l.RUnlock()
	currentNode := l.head
	currentIndex := 0
	for currentNode != nil {
		if currentIndex == index {
			return currentNode.data, true
		}
		currentNode = currentNode.next
		currentIndex++
	}
	return nil, false
}

// LInsert inserts the element before or after the pivot element
func (l *DoublyLinkedList) LInsert(pivot, data []byte, before bool) int {
	l.Lock()
	defer l.Unlock()
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
	return 0
}

// LPushX adds an element to the head of the list if the list exists
func (l *DoublyLinkedList) LPushX(data []byte) int {
	l.Lock()
	defer l.Unlock()
	if l.head == nil {
		return 0
	}
	l.LPush(data)
	return l.size()
}

// RPushX adds an element to the end of the list if the list exists
func (l *DoublyLinkedList) RPushX(data []byte) int {
	l.Lock()
	defer l.Unlock()
	if l.tail == nil {
		return 0
	}
	l.RPush(data)
	return l.size()
}

// LRem removes the first count occurrences of elements equal to value from the list
func (l *DoublyLinkedList) LRem(count int, value []byte) int {
	l.Lock()
	defer l.Unlock()
	currentNode := l.head
	removed := 0
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
		}
		currentNode = currentNode.next
	}
	return removed
}

// LSet sets the list element at index to value
func (l *DoublyLinkedList) LSet(index int, value []byte) bool {
	l.Lock()
	defer l.Unlock()
	currentNode := l.head
	currentIndex := 0
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
func (l *DoublyLinkedList) LTrim(start, end int) {
	l.Lock()
	defer l.Unlock()
	currentNode := l.head
	index := 0
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
	}
}

// Marshal returns the byte slice of the list
func (l *DoublyLinkedList) Marshal() ([]byte, error) {
	return binary.Marshal(l.LRange(0, -1))
}

// Unmarshal restores the list from the byte slice
func (l *DoublyLinkedList) Unmarshal(data []byte) error {
	var list [][]byte
	if err := binary.Unmarshal(data, &list); err != nil {
		return err
	}
	for _, item := range list {
		l.RPush(item)
	}
	return nil
}
