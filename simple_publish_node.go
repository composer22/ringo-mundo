package ringo

import "time"

// simplePublishNode represents a publisher, a job source who submits entries into the ring buffer.
// There is no locking with this implementation. Only one go routine that acts as a publisher should
// have an instantiate object for publishing events.
type simplePublishNode struct {
	committed  int64 // Write counter and index to the next ring buffer entry.
	cachepad1  [7]int64
	dependency *int64 // The committed register that this object is dependent on to finish.
	buffSize   int64  // Size of the ring buffer.
}

// NewSimplePublishNode is a factory function for returning a new instance of a simplePublishNode.
func NewSimplePublishNode(size int64) *simplePublishNode {
	return &simplePublishNode{
		buffSize: size,
	}
}

// Reserve is used by the publisher to validate it can store a new item on the buffer.
// It returns the next index as a pointer.
func (s *simplePublishNode) Reserve() *int64 {
	for s.committed-*s.dependency == s.buffSize {
		time.Sleep(time.Microsecond)
	}
	return &s.committed
}

// Commit increments the counter to indicate an entry has been stored or read.
func (s *simplePublishNode) Commit() {
	s.committed++
}

// Committed returns a pointer to the committed counter.
func (s *simplePublishNode) Committed() *int64 {
	return &s.committed
}

// SetDependency is a setter for the dependency of this node.
func (s *simplePublishNode) SetDependency(d *int64) {
	s.dependency = d
}
