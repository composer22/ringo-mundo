package ringo

import (
	"runtime"
	"sync/atomic"
)

// MultiPublishNode is shared by multiple thread/go routines for publishing events to the ring buffer.
// Because multiple routines must compete for next index, a single lock is maintained.
type MultiPublishNode struct {
	cachepad1  [8]int64
	sequence   int64 // Write counter and index to the next ring buffer entry.
	cachepad2  [8]int64
	committed  int64 // Keeps track of the number of written events to the ring.
	cachepad3  [8]int64
	dependency *int64 // The consumer's committed register we are dependent to finish before proceeding.
	cachepad4  [8]int64
	buffSize   int64 // Size of the ring buffer.
	cachepad5  [8]int64
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
		// Wait for room in the buffer if it is full.
		for previous-*m.dependency == m.buffSize {
			runtime.Gosched()
		}
		// Try and store the new increment. If it was changed by another routine, loop and try again.
		if atomic.CompareAndSwapInt64(&m.sequence, previous, next) {
			return previous
		}
	}
}

// Commit increments the commit register to indicate an entry has been stored.
func (m *MultiPublishNode) Commit() {
	m.committed++
}

// Committed returns a pointer to the committed counter.
func (m *MultiPublishNode) Committed() *int64 {
	return &m.committed
}

// SetDependency set the commtted counter that must complete work before we can proceed.
func (m *MultiPublishNode) SetDependency(d *int64) {
	m.dependency = d
}
