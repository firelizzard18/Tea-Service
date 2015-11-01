package main

import (
   "sync"
)

type ListenerList struct {
   lock *sync.RWMutex
   head *ListenerNode
   tail *ListenerNode
}

type ListenerNode struct {
   sink chan []byte

   list *ListenerList
   prev *ListenerNode
   next *ListenerNode
}

func NewListenerList() *ListenerList {
   l := new(ListenerList)
   l.lock = new(sync.RWMutex)
   l.head = nil
   l.tail = nil
   return l
}

func (l *ListenerList) Append() *ListenerNode {
   node := new(ListenerNode)
   node.sink = make(chan []byte)
   node.list = l
   node.prev = nil
   node.next = nil

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

func (n *ListenerNode) Read() chan []byte {
   out := make(chan []byte)

   go func() {
      for b := range n.sink {
         out <- b
      }
      close(out)
   }()

   return out
}

func (n *ListenerNode) Remove() {
   n.list.lock.Lock()
   defer n.list.lock.Unlock()

   prev := n.prev
   next := n.next

   if (n.list.head == n) {
      n.list.head = next
   } else {
      prev.next = next
   }

   if (n.list.tail == n) {
      n.list.tail = prev
   } else {
      next.prev = prev
   }

   close(n.sink)
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