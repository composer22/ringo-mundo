package ringo

// SimpleQueue builds a simple queue for communication between two go routines.
// SimplePair is useful for situation where you need a queue between one publisher and one
// consumer without the performance degredation of locking.
type SimpleQueue struct {
	Publisher *SimpleNode
	Consumer  *SimpleNode
}

// SimpleQueueNew wires up and returns a simple lockless queue.
// SimpleNode(publisher) <== 1:1 ==> SimpleNode(consumer)
func SimpleQueueNew(size int64) *SimpleQueue {
	s := &SimpleQueue{
		Publisher: SimpleNodeNew(size, nil),
		Consumer:  SimpleNodeNew(size, nil),
	}
	s.Publisher.SetDependency(s.Consumer)
	s.Consumer.SetDependency(s.Publisher)
	return s
}
