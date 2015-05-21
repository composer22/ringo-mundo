package ringo

import "runtime"

type SimpleConsumeNode struct {
	sequence   int64              // Read counter and index to the next rinbuffer entry.
	dependency *SimplePublishNode // The publisher node who we are dependent on for posted items.
	buffSize   int64              // Size of the ring buffer.
}

// Factory function for returning a new instance of a SimpleConsumeNode.
func SimpleConsumeNodeNew(size int64) *SimpleConsumeNode {
	return &SimpleConsumeNode{
		buffSize: size,
	}
}

// Reserve is used by the consumer to validate it should read a new item from the buffer.
// It returns the next index.
func (s *SimpleConsumeNode) Reserve() *int64 {
	for s.dependency.sequence-s.sequence == 0 {
		runtime.Gosched()
	}
	return &s.sequence
}

// Commit increments the counter to indicate an entry has been stored or read.
func (s *SimpleConsumeNode) Commit() {
	s.sequence++
}

// SetDependency is a setter for the dependency of this node.
func (s *SimpleConsumeNode) SetDependency(d *SimplePublishNode) {
	s.dependency = d
}
