package ringo

import "runtime"

// simpleConsumeNode represents a reader, a consumer who processes entries from the ring buffer.
// Each go routine that acts as a consumer should have an instantiated object for tracking it's results.
type simpleConsumeNode struct {
	cachepad1  [8]int64
	committed  int64 // Read counter and index to the next ring buffer entry.
	cachepad2  [7]int64
	dependency *int64 // The committed register that this object is dependent on to finish.
}

// NewSimpleConsumeNode is a factory function for returning a new instance of a simpleConsumeNode.
func NewSimpleConsumeNode() *simpleConsumeNode {
	return &simpleConsumeNode{}
}

// Reserve is used by the consumer to validate it should read a new item from the buffer.
// It returns the next index as a pointer.
func (s *simpleConsumeNode) Reserve() *int64 {
	for *s.dependency-s.committed == 0 {
		runtime.Gosched()
	}
	return &s.committed
}

// Commit increments the counter to indicate an entry has been read.
func (s *simpleConsumeNode) Commit() {
	s.committed++
}

// Committed returns a pointer to the committed counter.
func (s *simpleConsumeNode) Committed() *int64 {
	return &s.committed
}

// SetDependency sets the dependent commit counter of this node.
func (s *simpleConsumeNode) SetDependency(d *int64) {
	s.dependency = d
}
