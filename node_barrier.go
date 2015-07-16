package ringo

import (
	"math"
	"time"
)

// nodeBarrier acts as a collector of committed states, setting itself to reflect the lowest found.
// For example two Consumers are running in parallel.  A third Consumer must wait for both
// to complete their work on a cell in a ring buffer. A Barrier would watch the first two Consumers
// and record when both complete the same cell. The third Consumer would check the barrier to see if it
// can also proceed to read the next cell.
type nodeBarrier struct {
	cachepad1    [8]int64  // Cacheline padding.
	cursor       int64     // Tracks the cell id being processed in the ring.
	cachepad2    [7]int64  // Cacheline padding.
	committed    []int32   // Tracks this nodes progress.
	dependencies [][]int32 // Measures multiple dependent node progress.
	mask         int64     // Used in place of modulo for index calculations.
	shift        uint8     // Used to mark a cell with which rotation processed.
	running      bool      // Is this Barrier chasing the dependencies in a Run() loop?
}

// Factory function for returning a new instance of a nodeBarrier.
func NewNodeBarrier(size int64) *nodeBarrier {
	n := &nodeBarrier{
		cursor:       int64(initSeqValue),
		committed:    make([]int32, size),
		dependencies: make([][]int32, 0),
		mask:         size - 1,
		shift:        uint8(math.Log2(float64(size))),
	}

	for i := int64(0); i < size; i++ {
		n.committed[i] = int32(initSeqValue)
	}
	return n
}

// Run continually updates the commt status by chasing the multiple dependencies.
func (n *nodeBarrier) Run() {
	n.running = true
	for n.running {
		n.cursor++ // Increment pointer.

		// Wait for all dependencies to complete.
		for _, dep := range n.dependencies {
			for dep[n.cursor&n.mask] != int32(n.cursor>>n.shift) {
				time.Sleep(time.Microsecond)
			}
		}

		// Mark and continue.
		n.committed[n.cursor&n.mask] = int32(n.cursor >> n.shift)
	}
}

// Stop breaks the loop cycle of the run.
func (n *nodeBarrier) Stop() {
	n.running = false
}

// Running returns the state of the running flag.
func (n *nodeBarrier) Running() bool {
	return n.running
}

// AddDependency is a setter for a dependency of this barrier.
func (n *nodeBarrier) AddDependency(dep []int32) {
	n.dependencies = append(n.dependencies, dep)
}

// Committed is a getter for the commit ring of this node.
func (n *nodeBarrier) Committed() []int32 {
	return n.committed
}
