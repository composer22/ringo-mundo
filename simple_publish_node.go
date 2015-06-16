package ringo

import "runtime"

// SimplePublishNode represents a publisher, a job source who submits entries into the ring buffer.
// There is no locking with this implementation. Only one go routine that acts as a publisher should
// have an instantiate object for publishing events.
type SimplePublishNode struct {
	committed  int64 // Write counter and index to the next ring buffer entry.
	cachepad2  [7]int64
	dependency *int64 // The committed register that this object is dependent on to finish.
	buffSize   int64  // Size of the ring buffer.
	cachepad3  [6]int64
}

// Factory function for returning a new instance of a SimplePublishNode.
func SimplePublishNodeNew(size int64) *SimplePublishNode {
	return &SimplePublishNode{
		buffSize: size,
	}
}

// Reserve is used by the publisher to validate it can store a new item on the buffer.
// It returns the next index as a pointer.
func (s *SimplePublishNode) Reserve() *int64 {
	for s.committed-*s.dependency == s.buffSize {
		runtime.Gosched()
	}
	return &s.committed
}

// Commit increments the counter to indicate an entry has been stored or read.
func (s *SimplePublishNode) Commit() {
	s.committed++
}

// Committed returns a pointer to the committed counter.
func (s *SimplePublishNode) Committed() *int64 {
	return &s.committed
}

// SetDependency is a setter for the dependency of this node.
func (s *SimplePublishNode) SetDependency(d *int64) {
	s.dependency = d
}
