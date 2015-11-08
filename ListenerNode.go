package main

type ListenerNode struct {
   sink chan []byte

   list *ListenerList
   prev *ListenerNode
   next *ListenerNode
}

func newListenerNode(list *ListenerList) *ListenerNode {
   node := new(ListenerNode)

   node.sink = make(chan []byte)
   node.list = list
   node.prev = nil
   node.next = nil

   return node
}

func (n *ListenerNode) Read() chan []byte {
   out := make(chan []byte)

   go func() {
      n.list.lock.RLock()
      defer close(out)
      defer n.list.lock.RUnlock()

      for b := range n.sink {
         out <- b
      }
   }()

   return out
}

func (n *ListenerNode) Write(data []byte) {
   n.list.lock.RLock()
   defer n.list.lock.RUnlock()
   n.sink <- data
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