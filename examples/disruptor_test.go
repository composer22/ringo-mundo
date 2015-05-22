package ringo

// The following is an example on how to wire up a simple disruptor pattern.
// In a disruptor pattern, multiple publishers write events or to the ring buffer.
// Consumers waiting on the Publisher multicast these events to various  sources,
// For example, one consumer might act as a journaler; another as a feed to an external system.
// Trailing behind this layer of consumers is a barrier, which acts as a monitor over these consumers
// an records the lowest sequence.
// Finally, connected to the barrier is Consumer which acts as the application and is dependent on the
// previous consumers to have finished performing their reads.
// The original Publisher uses this Application Consumer as a dependency to know whether it can
// publish more events into the ring e.g. whether it is full.
// For this example, the topology will be:
// MultiPublisher(3 goroutines put) *=> 2 MultiConsumers => Barrier => 1 MultiConsumer (app)
//         ||                                                                ^^
//         -------------------------- <== dependency ------------------------||
func DisruptorWireupExample() {
	ringSize := ringo.PT16Meg
	mask := ringSize - 1
	ringBuffer := make([]int, ringSize) // A big ring of ints all ready to go.

	// Build the components
	publisher := ringo.MultiPublishNodeNew(ringSize) // Publisher to share in incoming go routines.
	consumer1 := ringo.MultiConsumeNodeNew(ringSize) // Journaler go routine use.
	consumer2 := ringo.MultiConsumeNodeNew(ringSize) // Send to external system go routine use.
	consumer3 := ringo.MultiConsumeNodeNew(ringSize) // App consumer dependent on above for go routine.
	barrier := ringo.BarrierNew(ringSize)            // Barrier to watch consumer 1 and 2.

	// Link the committed counter dependencies together.
	consumer1.SetDependency(publisher.Committed())
	consumer2.SetDependency(publisher.Committed())
	barrier.AddDependency(consumer1.Committed())
	barrier.AddDependency(consumer2.Committed())
	consumer2.SetDependency(barrier.Committed())
	publisher.SetDependency(consumer3.Committed())

}
