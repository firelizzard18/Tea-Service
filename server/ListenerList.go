package server

import (
	"fmt"
	"sync/atomic"
	"unsafe"
)

type ListenerList struct {
	// lock *sync.RWMutex
	head *ListenerNode
	// tail *ListenerNode
}

func NewListenerList() *ListenerList {
	l := new(ListenerList)
	// l.lock = new(sync.RWMutex)
	l.head = nil
	// l.tail = nil
	return l
}

func (l *ListenerList) Append() *ListenerNode {
	node := newListenerNode(l)
	new := unsafe.Pointer(node)

	// l.lock.Lock()
	// defer l.lock.Unlock()

	for l.head == nil {
		head := (*unsafe.Pointer)(unsafe.Pointer(&l.head))
		if atomic.CompareAndSwapPointer(head, nil, new) {
			return node
		}
	}

	tail := l.head
	for {
		for tail.next != nil {
			tail = tail.next
		}
		next := (*unsafe.Pointer)(unsafe.Pointer(&tail.next))
		if atomic.CompareAndSwapPointer(next, nil, new) {
			return node
		}
	}
}

func (l *ListenerList) Traverse() chan *ListenerNode {
	out := make(chan *ListenerNode)

	go func() {
		// l.lock.RLock()
		// defer l.lock.RUnlock()
		defer close(out)

		for node := l.head; node != nil; node = node.next {
			out <- node
		}
	}()

	return out
}

func (l *ListenerList) Remove(node *ListenerNode) bool {
	old := unsafe.Pointer(node)
	for {
		new := unsafe.Pointer(node.next)

		if node == l.head {
			head := (*unsafe.Pointer)(unsafe.Pointer(&l.head))
			if atomic.CompareAndSwapPointer(head, old, new) {
				fmt.Println("swapped head", head, l.head)
				return true
			}
		} else {
			prev := l.head
			for prev.next != node {
				if prev == nil {
					return false
				}
				if prev.next != node {
					break
				}
				prev = prev.next
			}
			next := (*unsafe.Pointer)(unsafe.Pointer(&prev.next))
			if atomic.CompareAndSwapPointer(next, old, new) {
				fmt.Println("swapped next", next, prev.next)
				return true
			}
		}
	}
}
