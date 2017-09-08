package server

import (
	"sync"
)

type ListenerList struct {
	// TODO use CompareAndSwapPointer
	lock *sync.RWMutex
	head *ListenerNode
	tail *ListenerNode
}

func NewListenerList() *ListenerList {
	l := new(ListenerList)
	l.lock = new(sync.RWMutex)
	l.head = nil
	l.tail = nil
	return l
}

func (l *ListenerList) Append() *ListenerNode {
	node := newListenerNode(l)

	l.lock.Lock()
	defer l.lock.Unlock()

	if l.head == nil {
		l.head = node
		l.tail = node
		return node
	}

	tail := l.tail
	tail.next = node
	node.prev = tail
	l.tail = node
	return node
}

func (l *ListenerList) Traverse() chan *ListenerNode {
	out := make(chan *ListenerNode)

	go func() {
		l.lock.RLock()
		defer close(out)
		defer l.lock.RUnlock()

		for node := l.head; node != nil; node = node.next {
			out <- node
		}
	}()

	return out
}
