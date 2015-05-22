package ringo

import "runtime"

// SimpleConsumeNode represents a reader, a consumer who processes entries from the ring buffer.
// Each go routine that acts as a consumer should have an instatiate object for tracking results.
type SimpleConsumeNode struct {
	committed  int64  // Read counter and index to the next ring buffer entry.
	dependency *int64 // Pointer to upstream dependency
}

// Factory function for returning a new instance of a SimpleConsumeNode.
func SimpleConsumeNodeNew(size int64) *SimpleConsumeNode {
	return &SimpleConsumeNode{}
}

// Reserve is used by the consumer to validate it should read a new item from the buffer.
// It returns the next index.
func (s *SimpleConsumeNode) Reserve() *int64 {
	for *s.dependency-s.committed == 0 {
		runtime.Gosched()
	}
	return &s.committed
}

// Commit increments the counter to indicate an entry has been read.
func (s *SimpleConsumeNode) Commit() {
	s.committed++
}

// Committed returns a pointer to the committed counter.
func (s *SimpleConsumeNode) Committed() *int64 {
	return &s.committed
}

// SetDependency sets the dependent commit counter of this node.
func (s *SimpleConsumeNode) SetDependency(d *int64) {
	s.dependency = d
}
