package ringo

import (
	"math"
	"runtime"
)

// SeqSimple is an optimal hub forsingle thread/go routines to track access to a ring buffer.
// It can be used in a publisher or consumer role.
// Performance has been optimized using a lockless strategy and minimal cycyle of operations.
// You should use this design when your ringbuffer can dedicate one CPU to act as a publisher,
//  and another CPU to act as a consumer.  And each publishes or consumes one unit of work at a time.
type SeqSimple struct {
	cursor     *int64     // Seq number and pointer to the next available slot.
	dependency *SeqSimple // Another sequence committed buffer that we are waiting to finish work.
	leader     bool       // Is the sequence the leader of the network?
	committed  []int32    // Tracks the commit states of the work being performed.
	barrier    int64      // Used to calculate downstream or upstream dependencies.
	mask       int64      // Used for modulo calculations in indexes.
	shift      uint8      // Used for marking commit states in assignments to commited.
}

// Factory function for returning a new instance of a SeqSimple.
func SeqSimpleNew(size int64, dep *SeqSimple, leader bool) *SeqSimple {
	s := &SeqSimple{
		cursor:     new(int64),
		leader:     leader,
		dependency: dep,
		committed:  make([]int32, size),
		mask:       size - 1,
		shift:      uint8(math.Log2(float64(size))),
	}

	// Init the cursor and commit map with default values.
	*s.cursor = SequenceDefault
	if leader {
		s.barrier = size
	}

	// Initialize buffer.
	for i := int64(0); i < size; i++ {
		s.committed[i] = int32(SequenceDefault)
	}
	return s
}

// Reserve incrmenets and returns the next index for an entry to the ring buffer.
func (s *SeqSimple) Reserve() int64 {
	*s.cursor += 1                // Increment the pointer.
	gate := *s.cursor - s.barrier // Calculate the dependency marker.

	// Validate any dependency blocking.
	for s.dependency.committed[*s.cursor&s.mask] != int32(gate>>s.shift) {
		runtime.Gosched()
	}
	return *s.cursor
}

// Commit updates the result map to track that an entry in history has been allocated and used.
func (s *SeqSimple) Commit(index int64) {
	s.committed[index&s.mask] = int32(index >> s.shift)
}

// SetDependency is a setter for the dependency of this sequence.
func (s *SeqSimple) SetDependency(d *SeqSimple) {
	s.dependency = d
}

// Mask is a getter for the index mask.  Used by application to update the application ring buffer.
func (s *SeqSimple) Mask() int64 {
	return s.mask
}
