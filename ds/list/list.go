package list

import (
	"time"

	"github.com/kelindar/binary"
)

type Node struct {
	data string
	next *Node
	prev *Node
}

type DoublyLinkedList struct {
	head *Node
	tail *Node
}

// NewDoublyLinkedList returns a new doubly linked list
func NewDoublyLinkedList() *DoublyLinkedList {
	return &DoublyLinkedList{}
}

// LPush adds an element to the head of the list
func (list *DoublyLinkedList) LPush(data string) {
	newNode := &Node{data: data}
	if list.head == nil {
		list.head = newNode
		list.tail = newNode
	} else {
		newNode.next = list.head
		list.head.prev = newNode
		list.head = newNode
	}
}

// RPush adds an element to the end of the list
func (list *DoublyLinkedList) RPush(data string) {
	newNode := &Node{data: data}
	if list.head == nil {
		list.head = newNode
		list.tail = newNode
	} else {
		list.tail.next = newNode
		newNode.prev = list.tail
		list.tail = newNode
	}
}

// LPop returns the first element of the list
func (list *DoublyLinkedList) LPop() any {
	if list.head == nil {
		return nil // 链表为空
	}
	data := list.head.data
	list.head = list.head.next
	if list.head != nil {
		list.head.prev = nil
	} else {
		list.tail = nil
	}
	return data
}

// RPop returns the last element of the list
func (list *DoublyLinkedList) RPop() any {
	if list.tail == nil {
		return -1 // 链表为空
	}
	data := list.tail.data
	list.tail = list.tail.prev
	if list.tail != nil {
		list.tail.next = nil
	} else {
		list.head = nil
	}
	return data
}

// LRange returns a range of elements from the list
func (list *DoublyLinkedList) LRange(start, end int) []any {
	var result []any
	currentNode := list.head
	index := 0
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

// LLen returns the length of the list
func (list *DoublyLinkedList) LLen() int {
	currentNode := list.head
	length := 0
	for currentNode != nil {
		length++
		currentNode = currentNode.next
	}
	return length
}

// BLPop removes and returns the first element of the list
func (list *DoublyLinkedList) BLPop(timeout time.Duration) any {
	if list.head == nil {
		time.Sleep(timeout)
	}
	return list.LPop()
}

// BRPop removes and returns the last element of the list
func (list *DoublyLinkedList) BRPop(timeout time.Duration) any {
	if list.tail == nil {
		time.Sleep(timeout)
	}
	return list.RPop()
}

// LIndex returns the element at index in the list
func (list *DoublyLinkedList) LIndex(index int) (any, bool) {
	currentNode := list.head
	currentIndex := 0
	for currentNode != nil {
		if currentIndex == index {
			return currentNode.data, true
		}
		currentNode = currentNode.next
		currentIndex++
	}
	return 0, false
}

// LInsert inserts the element before or after the pivot element
func (list *DoublyLinkedList) LInsert(pivot, data string, before bool) int {
	currentNode := list.head
	for currentNode != nil {
		if currentNode.data == pivot {
			newNode := &Node{data: data}
			if before {
				newNode.next = currentNode
				newNode.prev = currentNode.prev
				if currentNode.prev != nil {
					currentNode.prev.next = newNode
				} else {
					list.head = newNode
				}
				currentNode.prev = newNode
			} else {
				newNode.prev = currentNode
				newNode.next = currentNode.next
				if currentNode.next != nil {
					currentNode.next.prev = newNode
				} else {
					list.tail = newNode
				}
				currentNode.next = newNode
			}
			return list.LLen()
		}
		currentNode = currentNode.next
	}
	return 0
}

// LPushX adds an element to the head of the list if the list exists
func (list *DoublyLinkedList) LPushX(data string) int {
	if list.head == nil {
		return 0
	}
	list.LPush(data)
	return list.LLen()
}

// RPushX adds an element to the end of the list if the list exists
func (list *DoublyLinkedList) RPushX(data string) int {
	if list.tail == nil {
		return 0
	}
	list.RPush(data)
	return list.LLen()
}

// LRem removes the first count occurrences of elements equal to value from the list
func (list *DoublyLinkedList) LRem(count int, value any) int {
	currentNode := list.head
	removed := 0
	for currentNode != nil {
		if currentNode.data == value {
			if count > 0 && removed == count {
				break
			}
			if currentNode.prev != nil {
				currentNode.prev.next = currentNode.next
			} else {
				list.head = currentNode.next
			}
			if currentNode.next != nil {
				currentNode.next.prev = currentNode.prev
			} else {
				list.tail = currentNode.prev
			}
			removed++
		}
		currentNode = currentNode.next
	}
	return removed
}

// LSet sets the list element at index to value
func (list *DoublyLinkedList) LSet(index int, value string) bool {
	currentNode := list.head
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
func (list *DoublyLinkedList) LTrim(start, end int) {
	currentNode := list.head
	index := 0
	for currentNode != nil {
		if index < start || index > end {
			if currentNode.prev != nil {
				currentNode.prev.next = currentNode.next
			} else {
				list.head = currentNode.next
			}
			if currentNode.next != nil {
				currentNode.next.prev = currentNode.prev
			} else {
				list.tail = currentNode.prev
			}
		}
		currentNode = currentNode.next
		index++
	}
}

// Marshal returns the byte slice of the list
func (list *DoublyLinkedList) Marshal() ([]byte, error) {
	return binary.Marshal(list)
}

// Unmarshal restores the list from the byte slice
func (list *DoublyLinkedList) Unmarshal(data []byte) error {
	return binary.Unmarshal(data, list)
}
