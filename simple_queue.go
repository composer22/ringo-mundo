package ringo

// SimpleQueue builds a simple queue for communication between two go routines.
// SimpleQueue is useful for situation where you need a queue between one publisher and one
// consumer without the performance degredation of locking.
type SimpleQueue struct {
	Master *SimplePublishNode // Publisher
	Slave  *SimpleConsumeNode // Consumer
}

// SimpleQueueNew wires up and returns a simple lockless queue.
// SimplePublishNode(publisher/master) <== 1:1 ==> SimpleConsumeNode(consumer/slave)
func SimpleQueueNew(size int64) *SimpleQueue {
	s := &SimpleQueue{
		Master: SimplePublishNodeNew(size),
		Slave:  SimpleConsumeNodeNew(size),
	}
	s.Master.SetDependency(s.Slave)
	s.Slave.SetDependency(s.Master)
	return s
}
