package list

type NodeG[T any] struct {
	data T
	next *NodeG[T]
	prev *NodeG[T]
}

// Value returns the value of the node
func (n *NodeG[T]) Value() T {
	return n.data
}

// LinkedListG is a doubly linked list
type LinkedListG[T any] struct {
	head   *NodeG[T]
	tail   *NodeG[T]
	length int64
}

// NewLinkedListG returns a new doubly linked list
func NewLinkedListG[T any]() *LinkedListG[T] {
	return &LinkedListG[T]{}
}

// LPush adds an element to the head of the list
func (l *LinkedListG[T]) LPush(data ...T) {
	for _, datum := range data {
		newNode := &NodeG[T]{data: datum}
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

// ForRange iterates over the list
func (l *LinkedListG[T]) ForRange(fn func(T) bool) {
	for n := l.head; n != nil; n = n.next {
		if !fn(n.data) {
			break
		}
	}
}

// ForRangeNode iterates over the list
func (l *LinkedListG[T]) ForRangeNode(fn func(*NodeG[T]) bool) {
	for n := l.head; n != nil; n = n.next {
		if !fn(n) {
			break
		}
	}
}

// RemoveNode removes a node from the list
func (l *LinkedListG[T]) RemoveNode(n *NodeG[T]) {
	if n.prev == nil {
		l.head = n.next
	} else {
		n.prev.next = n.next
	}
	if n.next == nil {
		l.tail = n.prev
	} else {
		n.next.prev = n.prev
	}
	l.length--
}
