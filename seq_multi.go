package ringo

import (
	"math"
	"runtime"
	"sync/atomic"
)

// THIS CODE IS A WIP AND SHOULD BE CONSIDERED DEPRECATED OR UNUSABLE

// SeqMulti is used by multiple thread/go routines for tracking a Ring Buffer.
type SeqMulti struct {
	cursor     *int64    // Seq number and pointer to the next available pub/con slot.
	dependency *SeqMulti // Another sequence committed buffer that we are waiting on finishing work.
	leader     bool      // Is the sequence a follower (or a leader?
	buffSize   int64     // The length of the ringbuffer and the committed map.
	committed  []int32   // Tracks the commit states of the work being performed.
	barrier    int64     // Used to calculate downstream or upstream dependencies.
	mask       int64     // Used for modulo calculations in indexes.
	shift      uint8     // Used for marking commit states in assignments to commited.
}

// Factory function for returning a new instance of a SeqMulti.
func SeqMultiNew(size int64, dep *SeqMulti, leader bool) *SeqMulti {
	s := &SeqMulti{
		cursor:     new(int64),
		leader:     leader,
		dependency: dep,
		committed:  make([]int32, size),
		buffSize:   size,
		mask:       size - 1,
		shift:      uint8(math.Log2(float64(size))),
	}

	// Init the cursor and barrier adjustment with values.
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

// Reserve returns the upper most index for a segment of cells requested by "size".
func (s *SeqMulti) Reserve(count int64) int64 {
	// Loop and allocate
	for {
		previous := atomic.LoadInt64(s.cursor) // Get the previous pointer.
		upper := previous + count              // Increment it to get the upper bounds of the chunk.
		lower := previous + 1
		gate := lower - s.barrier // Calculate the dependency barrier

		// Check dependency based on first cell in the series.  If has not been processed
		// in the last rotation, wait
		for s.dependency.committed[lower&s.mask] != int32(gate>>s.shift) {
			runtime.Gosched()
		}

		// Update the new sequence number
		if atomic.CompareAndSwapInt64(s.cursor, previous, upper) {
			return upper
		}
	}
}

// Commit updates the committed map to track that a segment in the ring buffer
// has been allocated and used.
func (s *SeqMulti) Commit(lower, upper int64) {
	for ; upper >= lower; upper-- {
		s.committed[upper&s.mask] = int32(upper >> s.shift)
	}
}

// SetDependency is a setter for the dependency of this sequence.
func (s *SeqMulti) SetDependency(d *SeqMulti) {
	s.dependency = d
}

// Mask is a getter for the index mask.
func (s *SeqMulti) Mask() int64 {
	return s.mask
}
