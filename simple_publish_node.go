package ringo

import "runtime"

type SimplePublishNode struct {
	sequence   int64              // Write counter and index to the next ringbuffer entry.
	dependency *SimpleConsumeNode // The consumer who we are dependent on to read.
	buffSize   int64              // Size of the ring buffer.
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
	for s.sequence-s.dependency.sequence == s.buffSize {
		runtime.Gosched()
	}
	return &s.sequence
}

// Commit increments the counter to indicate an entry has been stored or read.
func (s *SimplePublishNode) Commit() {
	s.sequence++
}

// SetDependency is a setter for the dependency of this node.
func (s *SimplePublishNode) SetDependency(d *SimpleConsumeNode) {
	s.dependency = d
}
