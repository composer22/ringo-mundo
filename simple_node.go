package ringo

import "runtime"

// SimpleNode is an optimizes hub for single thread/go routines to track access to a ring buffer.
// Performance has been optimized using a lockless strategy and using minimal CPU operations.
type SimpleNode struct {
	counter    *int64      // Read or Write counter.
	dependency *SimpleNode // Another node that we are waiting to finish work.
	buffSize   int64       // Size of the ring buffer.
	mask       int64       // Used for modulo conversion.
}

// Factory function for returning a new instance of a SimpleQueue.
func SimpleNodeNew(size int64, dep *SimpleNode) *SimpleNode {
	s := &SimpleNode{
		counter:    new(int64),
		dependency: dep,
		buffSize:   size,
		mask:       size - 1,
	}
	return s
}

// PublishReserve is used by the publisher to validate it can store a new item on the buffer.
// It returns the next index.
func (s *SimpleNode) PublishReserve() int64 {
	for *s.counter-*s.dependency.counter == s.buffSize {
		runtime.Gosched()
	}
	return *s.counter
}

// ConsumeReserve is used by the consumer to validate it should read a new item from the buffer.
// It returns the next index.
func (s *SimpleNode) ConsumeReserve() int64 {
	for *s.dependency.counter-*s.counter == 0 {
		runtime.Gosched()
	}
	return *s.counter
}

// Commit increments the counter to indicate an entry has been stored or read.
func (s *SimpleNode) Commit(noop int64) {
	*s.counter++
}

// SetDependency is a setter for the dependency of this sequence.
func (s *SimpleNode) SetDependency(d *SimpleNode) {
	s.dependency = d
}

// Mask is a getter for the index mask which is used for modulo conversions w/ indexes to the ring.
func (s *SimpleNode) Mask() int64 {
	return s.mask
}
