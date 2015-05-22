package ringo

import "runtime"

// SimpleConsumeNode represents a publisher, a job source who submits entries into the ring buffer.
// Only one go routine that acts as a publisher should have an instatiate object for publishing events.
// This is a lock-less implementation of a publisher, since it is only one source.
type SimplePublishNode struct {
	committed  int64  // Write counter and index to the next ringbuffer entry.
	dependency *int64 // The consumer who we are dependent on to read.
	buffSize   int64  // Size of the ring buffer.
}

// Factory function for returning a new instance of a SimplePublishNode.
func SimplePublishNodeNew(size int64) *SimplePublishNode {
	return &SimplePublishNode{
		buffSize: size,
	}
}

// Reserve is used by the publisher to validate it can store a new item on the buffer.
// It returns the next index.
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
