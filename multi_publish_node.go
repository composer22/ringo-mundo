package ringo

import (
	"runtime"
	"sync/atomic"
)

// multiPublishNode is shared by multiple thread/go routines for publishing events to the ring buffer.
// Because multiple routines must compete for next index, a single lock is maintained.
type multiPublishNode struct {
	sequence   int64 // Write counter and index to the next ring buffer entry.
	cachepad1  [7]int64
	committed  int64 // Keeps track of the number of written events to the ring.
	cachepad2  [7]int64
	dependency *int64 // The consumer's committed register we are dependent to finish before proceeding.
	buffSize   int64  // Size of the ring buffer.
}

// Factory function for returning a new instance of a multiPublishNode.
func NewMultiPublishNode(size int64) *multiPublishNode {
	return &multiPublishNode{
		buffSize: size,
	}
}

// Reserve returns the next new index.
func (m *multiPublishNode) Reserve() int64 {
	for {
		previous := m.sequence // Get the previous counter.
		// Wait for room in the buffer if it is full.
		for previous-*m.dependency == m.buffSize {
			runtime.Gosched()
		}
		// Try and store the new increment. If it was changed by another routine, loop and try again.
		if atomic.CompareAndSwapInt64(&m.sequence, previous, previous+1) {
			return previous
		}
	}
}

// Commit increments the commit register to indicate an entry has been stored.
func (m *multiPublishNode) Commit() {
	m.committed++
}

// Committed returns a pointer to the committed counter.
func (m *multiPublishNode) Committed() *int64 {
	return &m.committed
}

// SetDependency set the commtted counter that must complete work before we can proceed.
func (m *multiPublishNode) SetDependency(d *int64) {
	m.dependency = d
}
