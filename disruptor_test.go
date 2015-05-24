package ringo

import (
	"runtime"
	"testing"
)

// The following is an example on how to wire up a simple disruptor pattern.
// In a disruptor pattern, multiple publishers write events to the ring buffer.
// Consumers waiting on the Publisher submission, multicast these events to various  sources,
// For example, one consumer might act as a journaler; another as a feed to an external system.
// Trailing behind this layer of consumers is a barrier, which acts as a monitor to records the lowest id.
// Finally, connected to the barrier is a Consumer which acts as the application and is dependent on the
// previous consumers to have finished performing their reads before it does it's finalization.
// The original Publisher uses this Application Consumer as a dependency to know whether it can
// publish more events into the ring e.g. whether the ring is full.
// For this example, the topology will be:
// MultiPublisher(3 goroutines put) *=> 2 MultiConsumers => Barrier => 1 MultiConsumer (app)
//         ||                                                                ^^
//         -------------------------- <== dependency ------------------------||
//
func TestDisruptorSmall(t *testing.T) {
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(prevProcs)

	// Build the components
	publisher := MultiPublishNodeNew(32) // Publisher to share in incoming go routines.
	consumer1 := SimpleConsumeNodeNew()  // Consumer 1: Journaler
	consumer2 := SimpleConsumeNodeNew()  // Consumer 2: Send to external system go routine use.
	barrier := BarrierNew()              // Barrier to watch consumer 1 and 2.
	consumer3 := SimpleConsumeNodeNew()  // Consumer 3: App consumer dependent on above for go routine.

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

// go test -run=XXX -bench=BenchmarkDisruptor
// 23.6 million transactions per second (42.3 ns/op)
func BenchmarkDisruptor(b *testing.B) {
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(prevProcs)
	interations := int64(b.N)

	// Build the components
	publisher := MultiPublishNodeNew(PT64Meg) // Publisher to share in incoming go routines.
	consumer1 := SimpleConsumeNodeNew()       // Consumer 1: Journaler
	consumer2 := SimpleConsumeNodeNew()       // Consumer 2: Send to external system go routine use.
	barrier := BarrierNew()                   // Barrier to watch consumer 1 and 2.
	consumer3 := SimpleConsumeNodeNew()       // Consumer 3: App consumer dependent on above for go routine.

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
