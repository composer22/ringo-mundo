package ringo

import (
	"runtime"
	"testing"
)

// The following tests are an example on how to wire up a simple disruptor pattern.
// In a disruptor pattern, multiple publishers write events to the ring buffer.
// Consumers waiting on the Publisher submission, multicast these events to various resources,
// For example, one consumer might act as a journaler; another as a feed to an external system.
// Trailing behind this layer of consumers is a barrier, which acts as a monitor to records the
// lowest completed index.
// Finally, connected to the barrier is a Consumer which acts as the application and is dependent on the
// previous consumers to have finished performing their reads before it does it's own finalization.
// The original Publisher uses this Application Consumer as a dependency to know whether it can
// publish more events into the ring e.g. whether the ring is full.
// For this example, the topology will be:
// 1 MultiPublishNode(3 goroutines) *=> 2 SimpleConsumeNodes(journal/send) => Barrier => 1 SimpleConsumeNode(app)
//         ^^                                                                                    VV
//         ||--------------------------------------- <== dependency <== -------------------------||
//
func TestDisruptorSmallType1(t *testing.T) {
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(runtime.NumCPU())
	defer runtime.GOMAXPROCS(prevProcs)

	// Build the components
	publisher := NewMultiPublishNode(32) // Publisher to share in incoming go routines.
	consumer1 := NewSimpleConsumeNode()  // Consumer 1: Journaler.
	consumer2 := NewSimpleConsumeNode()  // Consumer 2: Send to external system go routine use.
	barrier := NewConsumeBarrier()       // Barrier to watch consumer 1 and 2 committed counts.
	consumer3 := NewSimpleConsumeNode()  // Consumer 3: App consumer dependent on above for go routine.

	// Link the committed counter dependencies together.
	consumer1.SetDependency(publisher.Committed()) // Watch publisher for work.
	consumer2.SetDependency(publisher.Committed()) // Watch publisher for work.
	barrier.AddDependency(consumer1.Committed())   // Watch consumer for complete.
	barrier.AddDependency(consumer2.Committed())   // Watch consumer for complete.
	consumer3.SetDependency(barrier.Committed())   // Watch barrier for lowest complete.
	publisher.SetDependency(consumer3.Committed()) // Watch the last consumer for complete.

	done := make(chan bool)

	go func() {
		barrier.Run()
	}()

	go func() {
		for i := int64(0); i < 64; i++ {
			consumer1.Reserve()
			consumer1.Commit()
		}
	}()

	go func() {
		for i := int64(0); i < 64; i++ {
			consumer2.Reserve()
			consumer2.Commit()
		}
	}()

	go func() {
		for i := int64(0); i < 64; i++ {
			consumer3.Reserve()
			consumer3.Commit()
		}
		done <- true
	}()

	for i := int64(0); i < 64; i++ {
		publisher.Reserve()
		publisher.Commit()
	}

	<-done
	barrier.Stop()
}

func TestDisruptorSmallType2(t *testing.T) {
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(runtime.NumCPU())
	defer runtime.GOMAXPROCS(prevProcs)

	// Build the components
	publisher := NewSimpleNode(true, 32)  // Publisher for one incoming go routine.
	consumer1 := NewSimpleNode(false, 32) // Consumer 1: Journaler
	consumer2 := NewSimpleNode(false, 32) // Consumer 2: Send to external system go routine use.
	barrier := NewNodeBarrier(32)         // Barrier to watch consumer 1 and 2.
	consumer3 := NewSimpleNode(false, 32) // Consumer 3: App consumer dependent on above for go routine.

	// Link the committed counter dependencies together.
	consumer1.SetDependency(publisher.Committed())
	consumer2.SetDependency(publisher.Committed())
	barrier.AddDependency(consumer1.Committed())
	barrier.AddDependency(consumer2.Committed())
	consumer3.SetDependency(barrier.Committed())
	publisher.SetDependency(consumer3.Committed())

	done := make(chan bool)

	go func() {
		for i := int64(0); i < 64; i++ {
			ndx := consumer1.Reserve()
			consumer1.Commit(ndx)
		}
	}()

	go func() {
		for i := int64(0); i < 64; i++ {
			ndx := consumer2.Reserve()
			consumer2.Commit(ndx)
		}
	}()

	go func() {
		barrier.Run()
	}()

	go func() {
		for i := int64(0); i < 64; i++ {
			ndx := consumer3.Reserve()
			consumer3.Commit(ndx)
		}
		done <- true
	}()

	for i := int64(0); i < 64; i++ {
		ndx := publisher.Reserve()
		publisher.Commit(ndx)
	}

	<-done
	barrier.Stop()
}

// go test -run=XXX -bench=BenchmarkDisruptor

// Simplified Disruptor Pattern - Single publishing source.
func BenchmarkDisruptorSimpleType1(b *testing.B) {
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(runtime.NumCPU()) // runtime.NumCPU()
	defer runtime.GOMAXPROCS(prevProcs)
	interations := int64(b.N)

	// Build the components
	publisher := NewSimplePublishNode(PT64Meg) // Publisher for one incoming go routine.
	consumer1 := NewSimpleConsumeNode()        // Consumer 1: Journaler
	consumer2 := NewSimpleConsumeNode()        // Consumer 2: Send to external system go routine use.
	barrier := NewConsumeBarrier()             // Barrier to watch consumer 1 and 2.
	consumer3 := NewSimpleConsumeNode()        // Consumer 3: App consumer dependent on above for go routine.

	// Link the committed counter dependencies together.
	consumer1.SetDependency(publisher.Committed())
	consumer2.SetDependency(publisher.Committed())
	barrier.AddDependency(consumer1.Committed())
	barrier.AddDependency(consumer2.Committed())
	consumer3.SetDependency(barrier.Committed())
	publisher.SetDependency(consumer3.Committed())

	done := make(chan bool)

	go func() {
		barrier.Run()
	}()

	go func() {
		for i := int64(0); i < interations; i++ {
			consumer1.Reserve()
			consumer1.Commit()
		}
	}()

	go func() {
		for i := int64(0); i < interations; i++ {
			consumer2.Reserve()
			consumer2.Commit()
		}
	}()

	go func() {
		for i := int64(0); i < interations; i++ {
			consumer3.Reserve()
			consumer3.Commit()
		}
		done <- true
	}()

	b.ReportAllocs()
	b.ResetTimer()
	for i := int64(0); i < interations; i++ {
		publisher.Reserve()
		publisher.Commit()
	}

	b.StopTimer()

	<-done
	barrier.Stop()
}

// Multiple publishers - standard Disruptor pattern.
func BenchmarkDisruptorMultiType1(b *testing.B) {
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(runtime.NumCPU())
	defer runtime.GOMAXPROCS(prevProcs)
	interations := int64(b.N)

	// Build the components
	publisher := NewMultiPublishNode(PT64Meg) // Publisher to share in incoming go routines.
	consumer1 := NewSimpleConsumeNode()       // Consumer 1: Journaler
	consumer2 := NewSimpleConsumeNode()       // Consumer 2: Send to external system go routine use.
	barrier := NewConsumeBarrier()            // Barrier to watch consumer 1 and 2.
	consumer3 := NewSimpleConsumeNode()       // Consumer 3: App consumer dependent on above for go routine.

	// Link the committed counter dependencies together.
	consumer1.SetDependency(publisher.Committed())
	consumer2.SetDependency(publisher.Committed())
	barrier.AddDependency(consumer1.Committed())
	barrier.AddDependency(consumer2.Committed())
	consumer3.SetDependency(barrier.Committed())
	publisher.SetDependency(consumer3.Committed())

	done := make(chan bool)

	go func() {
		barrier.Run()
	}()

	go func() {
		for i := int64(0); i < interations; i++ {
			consumer1.Reserve()
			consumer1.Commit()
		}
	}()

	go func() {
		for i := int64(0); i < interations; i++ {
			consumer2.Reserve()
			consumer2.Commit()
		}
	}()

	go func() {
		for i := int64(0); i < interations; i++ {
			consumer3.Reserve()
			consumer3.Commit()
		}
		done <- true
	}()

	b.ReportAllocs()
	b.ResetTimer()
	for i := int64(0); i < interations; i++ {
		publisher.Reserve()
		publisher.Commit()
	}

	b.StopTimer()

	<-done
	barrier.Stop()
}

// Simplified Disruptor Pattern - Single publishing source.
func BenchmarkDisruptorSimpleType2(b *testing.B) {
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(runtime.NumCPU()) // runtime.NumCPU()
	defer runtime.GOMAXPROCS(prevProcs)
	interations := int64(b.N)

	// Build the components
	publisher := NewSimpleNode(true, PT64Meg)  // Publisher for one incoming go routine.
	consumer1 := NewSimpleNode(false, PT64Meg) // Consumer 1: Journaler
	consumer2 := NewSimpleNode(false, PT64Meg) // Consumer 2: Send to external system go routine use.
	barrier := NewNodeBarrier(PT64Meg)         // Barrier to watch consumer 1 and 2.
	consumer3 := NewSimpleNode(false, PT64Meg) // Consumer 3: App consumer dependent on above for go routine.

	// Link the committed counter dependencies together.
	consumer1.SetDependency(publisher.Committed())
	consumer2.SetDependency(publisher.Committed())
	barrier.AddDependency(consumer1.Committed())
	barrier.AddDependency(consumer2.Committed())
	consumer3.SetDependency(barrier.Committed())
	publisher.SetDependency(consumer3.Committed())

	done := make(chan bool)

	go func() {
		for i := int64(0); i < interations; i++ {
			ndx := consumer1.Reserve()
			consumer1.Commit(ndx)
		}
	}()

	go func() {
		for i := int64(0); i < interations; i++ {
			ndx := consumer2.Reserve()
			consumer2.Commit(ndx)
		}
	}()

	go func() {
		barrier.Run()
	}()

	go func() {
		for i := int64(0); i < interations; i++ {
			ndx := consumer3.Reserve()
			consumer3.Commit(ndx)
		}
		done <- true
	}()

	b.ReportAllocs()
	b.ResetTimer()
	for i := int64(0); i < interations; i++ {
		ndx := publisher.Reserve()
		publisher.Commit(ndx)
	}

	b.StopTimer()

	<-done
	barrier.Stop()
}

// Multiple publishers - standard Disruptor pattern.
func BenchmarkDisruptorMultiType2(b *testing.B) {
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(runtime.NumCPU())
	defer runtime.GOMAXPROCS(prevProcs)
	interations := int64(b.N)

	// Build the components
	publisher := NewMultiNode(true, PT64Meg)   // Publisher for one incoming go routine.
	consumer1 := NewSimpleNode(false, PT64Meg) // Consumer 1: Journaler
	consumer2 := NewSimpleNode(false, PT64Meg) // Consumer 2: Send to external system go routine use.
	barrier := NewNodeBarrier(PT64Meg)         // Barrier to watch consumer 1 and 2.
	consumer3 := NewSimpleNode(false, PT64Meg) // Consumer 3: App consumer dependent on above for go routine.

	// Link the committed counter dependencies together.
	consumer1.SetDependency(publisher.Committed())
	consumer2.SetDependency(publisher.Committed())
	barrier.AddDependency(consumer1.Committed())
	barrier.AddDependency(consumer2.Committed())
	consumer3.SetDependency(barrier.Committed())
	publisher.SetDependency(consumer3.Committed())

	done := make(chan bool)

	go func() {
		for i := int64(0); i < interations; i++ {
			ndx := consumer1.Reserve()
			consumer1.Commit(ndx)
		}
	}()

	go func() {
		for i := int64(0); i < interations; i++ {
			ndx := consumer2.Reserve()
			consumer2.Commit(ndx)
		}
	}()

	go func() {
		barrier.Run()
	}()

	go func() {
		for i := int64(0); i < interations; i++ {
			ndx := consumer3.Reserve()
			consumer3.Commit(ndx)
		}
		done <- true
	}()

	b.ReportAllocs()
	b.ResetTimer()
	for i := int64(0); i < interations; i++ {
		ndx := publisher.Reserve()
		publisher.Commit(ndx)
	}

	b.StopTimer()

	<-done
	barrier.Stop()
}
