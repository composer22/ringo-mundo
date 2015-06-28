package ringo

import (
	"math"
	"time"
)

// simpleNode can be used by a publisher or consumer to track it's processing of ringbuffer entries.
// This entity assumes that is used by only one publisher or consumer. Reserve and Commits can
// only be called by a single go routine.
type simpleNode struct {
	cachepad1  [8]int64 // Cacheline padding.
	cursor     int64    // Tracks the cell id being processed in the ring.
	cachepad2  [7]int64 // Cacheline padding.
	committed  []int32  // Tracks this nodes progress.
	dependency []int32  // Measures a dependent nodes progress.
	barrier    int64    // Used to find the next dependent cell to check.
	mask       int64    // Used in place of modulo for index calculations.
	shift      uint8    // Used to mark a cell with which rotation processed.
}

// NewSimpleNode is a factory function that returns a single simpleNode instance.
func NewSimpleNode(leader bool, size int64) *simpleNode {
	s := &simpleNode{
		cursor:    int64(initSeqValue),
		committed: make([]int32, size),
		mask:      size - 1,
		shift:     uint8(math.Log2(float64(size))),
	}

	if leader {
		s.barrier = size
	}

	for i := int64(0); i < size; i++ {
		s.committed[i] = int32(initSeqValue)
	}
	return s
}

// Reserve is used to allocate a cell in the ring buffer for processing by the publisher or consumer.
// It validates that the cell has been processed by a dependency node and if free returns that index
// for use.
func (s *simpleNode) Reserve() int64 {
	s.cursor++                   // Increment the pointer to the next cell.
	gate := s.cursor - s.barrier // Calculate the dependency marker.

	// Validate that the dependency has completed processing on this cell and it's free for use.
	// If not, wait until it is.
	for s.dependency[s.cursor&s.mask] != int32(gate>>s.shift) {
		time.Sleep(time.Microsecond)
	}
	return s.cursor
}

// Commit marks a cell in the ring status as completed. Any other node that is relying on this
// node to complete it's work can now know it may proceed to use it's corresponding cell.
func (s *simpleNode) Commit(index int64) {
	s.committed[index&s.mask] = int32(index >> s.shift)
}

// Committed is a getter for the commit ring of this node.
func (s *simpleNode) Committed() []int32 {
	return s.committed
}

// SetDependency is a setter for the dependency of this node.
func (s *simpleNode) SetDependency(dep []int32) {
	s.dependency = dep
}
