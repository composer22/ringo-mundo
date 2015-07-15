package ringo

import (
	"math"
	"sync/atomic"
	"time"
)

// multiNode can be used by a publisher or consumer to track it's processing of ringbuffer entries.
// multiNode works in the same way as simpleNode except multiNode allows multiple threads to access
// it's functions concurrently. Due to it's need for the use of locks, it is slower than simpleNode.
// If you do not need concurrent access to the same node, use simpleNode.
type multiNode struct {
	cursor     int64    // Tracks the cell id being processed in the ring.
	cachepad1  [7]int64 // Cacheline padding.
	committed  []int32  // Tracks this nodes progress.
	dependency []int32  // Measures a dependent nodes progress.
	barrier    int64    // Used to find the next dependent cell to check.
	mask       int64    // Used in place of modulo for index calculations.
	shift      uint8    // Used to mark a cell with which rotation processed.
}

// NewMultiNode is a factory function that returns a single multiNode instance.
func NewMultiNode(leader bool, size int64) *multiNode {
	m := &multiNode{
		cursor:    int64(initSeqValue),
		committed: make([]int32, size),
		mask:      size - 1,
		shift:     uint8(math.Log2(float64(size))),
	}

	if leader {
		m.barrier = size
	}

	for i := int64(0); i < size; i++ {
		m.committed[i] = int32(initSeqValue)
	}
	return m
}

// Reserve is used to allocate a cell in the ring buffer for processing by the publisher or consumer.
// It validates that the cell has been processed by a dependency node and if free returns that index
// for use.
func (m *multiNode) Reserve() int64 {
	var previous, next, gate int64

	// Loop and allocate
	for {
		previous = atomic.LoadInt64(&m.cursor) // Get the previous pointer.
		next = previous + 1                    // Increment to get next index.
		gate = next - m.barrier                // Calculate the dependency marker.

		// Validate that the dependency has completed processing on this cell and it's free for use.
		// If not, wait until it is.
		for m.dependency[next&m.mask] != int32(gate>>m.shift) {
			time.Sleep(time.Microsecond)
		}

		// Try and update the new sequence number. If successful, then return,
		// otherwise loop and try this process again (some other caller got the index first).
		if atomic.CompareAndSwapInt64(&m.cursor, previous, next) {
			return next
		}
	}
}

// Commit marks a cell in the ring status as completed. Any other node that is relying on this
// node to complete it's work can now know it may proceed to use it's corresponding cell.
func (m *multiNode) Commit(index int64) {
	m.committed[index&m.mask] = int32(index >> m.shift)
}

// Committed is a getter for the commit ring of this node.
func (m *multiNode) Committed() []int32 {
	return m.committed
}

// SetDependency is a setter for the dependency of this node.
func (m *multiNode) SetDependency(dep []int32) {
	m.dependency = dep
}
