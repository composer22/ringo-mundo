package ringo

import "runtime"

type MultiConsumeNode struct {
	committed  int64  // Read counter and index to the next ring buffer entry.
	dependency *int64 // Pointer to upstream dependency
	buffSize   int64  // Size of the ring buffer.
}

// Factory function for returning a new instance of a MultiConsumeNode.
func MultiConsumeNodeNew(size int64) *MultiConsumeNode {
	return &MultiConsumeNode{
		buffSize: size,
	}
}

// Reserve is used by the consumer to validate it should read a new item from the buffer.
// It returns the next index.
func (m *MultiConsumeNode) Reserve() *int64 {
	for *m.dependency-m.committed == 0 {
		runtime.Gosched()
	}
	return &m.committed
}

// Commit increments the counter to indicate an entry has been read.
func (m *MultiConsumeNode) Commit() {
	m.committed++
}

// Committed returns a pointer to the committed counter.
func (m *MultiConsumeNode) Committed() *int64 {
	return &m.committed
}

// SetDependency sets the dependent commit counter of this node.
func (m *MultiConsumeNode) SetDependency(d *int64) {
	m.dependency = d
}
