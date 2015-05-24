package ringo

import "runtime"

// Barrier acts as a collector of committed states, setting itself to reflect the lowest found.
// This is used to setup dependencies between multiple readers.
type Barrier struct {
	committed    int64    // Read counter and index to the next ring buffer entry.
	dependencies []*int64 // A list of commit registers for upstream activity.
	running      bool     // Is the Barrier chasing the dependencies in a Run() loop?
}

// Factory function for returning a new instance of a Barrier.
func BarrierNew() *Barrier {
	return &Barrier{
		dependencies: make([]*int64, 0),
	}
}

// Run continually updates the current count by chasing the multiple dependencies.
func (b *Barrier) Run() {
	var lowest int64
	b.running = true
	for b.running {
		lowest = SequenceMax
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
func (b *Barrier) Stop() {
	b.running = false
}

// Running returns the state of the running flag.
func (b *Barrier) Running() bool {
	return b.running
}

// SetDependency is a setter for the dependencies of this barrier.
func (b *Barrier) AddDependency(d *int64) {
	b.dependencies = append(b.dependencies, d)
}

// Committed returns a pointer to the committed counter.
func (b *Barrier) Committed() *int64 {
	return &b.committed
}
