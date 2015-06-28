package ringo

import "runtime"

// consumeBarrier acts as a collector of committed states, setting itself to reflect the lowest found.
// This is used to setup dependencies between multiple components.
// For example two Consumers are running in parallel.  A third Consumer must wait for both
// to complete their work on a cell in a ring buffer. A Barrier would watch the first two Consumers
// and record the lowest completed result. The third Consumer would check the barrier to see if it
// can proceed to read the next cell.
type consumeBarrier struct {
	cachepad1    [8]int64
	committed    int64 // Lowest committed cell value from the dependencies.
	cachepad2    [7]int64
	dependencies []*int64 // A list of committed registers for upstream activity.
	running      bool     // Is this Barrier chasing the dependencies in a Run() loop?
}

// Factory function for returning a new instance of a consumeBarrier.
func NewConsumeBarrier() *consumeBarrier {
	return &consumeBarrier{
		dependencies: make([]*int64, 0),
	}
}

// Run continually updates the current count by chasing the multiple dependencies.
func (b *consumeBarrier) Run() {
	var lowest int64
	b.running = true
	for b.running {
		lowest = sequenceMax
		for _, d := range b.dependencies {
			if *d < lowest {
				lowest = *d
			}
		}
		b.committed = lowest
		runtime.Gosched()
	}
}

// Stop breaks the loop cycle of the run.
func (b *consumeBarrier) Stop() {
	b.running = false
}

// Running returns the state of the running flag.
func (b *consumeBarrier) Running() bool {
	return b.running
}

// Committed returns a pointer to the committed counter.
func (b *consumeBarrier) Committed() *int64 {
	return &b.committed
}

// AddDependency is a setter for a dependency of this barrier.
func (b *consumeBarrier) AddDependency(d *int64) {
	b.dependencies = append(b.dependencies, d)
}
