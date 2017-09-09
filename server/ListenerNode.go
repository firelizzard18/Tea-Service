package server

type ListenerNode struct {
	sink chan []byte

	list *ListenerList
	// prev *ListenerNode
	next *ListenerNode
}

func newListenerNode(list *ListenerList) *ListenerNode {
	node := new(ListenerNode)

	node.sink = make(chan []byte)
	node.list = list
	// node.prev = nil
	node.next = nil

	return node
}

func (n *ListenerNode) Read() chan []byte {
	out := make(chan []byte)

	go func() {
		// n.list.lock.RLock()
		// defer n.list.lock.RUnlock()
		defer close(out)

		for b := range n.sink {
			out <- b
		}
	}()

	return out
}

func (n *ListenerNode) Write(data []byte) {
	// n.list.lock.RLock()
	// defer n.list.lock.RUnlock()
	n.sink <- data
}

func (n *ListenerNode) Remove() bool {
	if !n.list.Remove(n) {
		return false
	}

	close(n.sink)
	return true
}
