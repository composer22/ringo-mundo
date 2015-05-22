package ringo

import (
	"runtime"
	"sync/atomic"
)

// MultiPublishNode is used by multiple thread/go routines for publishing events to the ring buffer.
type MultiPublishNode struct {
	sequence   int64  // Write counter and index to the next ringbuffer entry.
	committed  int64  // Keeps track of the number of written events to the ring.
	dependency *int64 // The consumer who we are dependent on to read from the buffer.
	buffSize   int64  // Size of the ring buffer.
}

// Factory function for returning a new instance of a MultiPublishNode.
func MultiPublishNodeNew(size int64) *MultiPublishNode {
	return &MultiPublishNode{
		buffSize: size,
	}
}

// Reserve returns the next new index.
func (m *MultiPublishNode) Reserve() int64 {
	for {
		previous := m.sequence // Get the previous counter.
		next := previous + 1   // and increment it. Hold it for later.
		// WaitFor room in the buffer.
		for previous-*m.dependency == m.buffSize {
			runtime.Gosched()
		}
		// Try and store the increment. If it has changed loop and try again.
		if atomic.CompareAndSwapInt64(&m.sequence, previous, next) {
			return previous
		}
	}
}

// Commit increments the comms to indicate an entry has been stored.
func (m *MultiPublishNode) Commit() {
	m.committed++
}

// Committed returns a pointer to the committed counter.
func (m *MultiPublishNode) Committed() *int64 {
	return &m.committed
}

// SetDependency set for the dependency commtted counter in this node.
func (m *MultiPublishNode) SetDependency(d *int64) {
	m.dependency = d
}
