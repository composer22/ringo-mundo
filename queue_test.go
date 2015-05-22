package ringo

import (
	"os"
	"runtime"
	"runtime/pprof"
	"testing"
)

func TestSimpleQueueSmall(t *testing.T) {
	// Set to one process.
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(prevProcs)

	master := SimplePublishNodeNew(32)
	slave := SimpleConsumeNodeNew(32)
	master.SetDependency(slave.Committed())
	slave.SetDependency(master.Committed())

	done := make(chan bool)

	go func() {
		for i := int64(0); i < 64; i++ {
			slave.Reserve()
			slave.Commit()
		}
		close(done)
	}()

	for i := int64(0); i < 64; i++ {
		master.Reserve()
		master.Commit()
	}
	<-done
}

func TestSimpleQueueLarge(t *testing.T) {
	// Set to one process.
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(prevProcs)

	master := SimplePublishNodeNew(PT64Meg)
	slave := SimpleConsumeNodeNew(PT64Meg)
	master.SetDependency(slave.Committed())
	slave.SetDependency(master.Committed())

	done := make(chan bool)

	go func() {
		for i := int64(0); i < PT64Meg; i++ {
			slave.Reserve()
			slave.Commit()
		}
		close(done)
	}()

	for i := int64(0); i < PT64Meg; i++ {
		master.Reserve()
		master.Commit()
	}
	<-done
}

func TestMultiQueueSmall(t *testing.T) {
	// Set to one process.
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(prevProcs)

	master := MultiPublishNodeNew(32)
	slave := SimpleConsumeNodeNew(32)
	master.SetDependency(slave.Committed())
	slave.SetDependency(master.Committed())

	done := make(chan bool)

	go func() {
		for i := int64(0); i < 64; i++ {
			slave.Reserve()
			slave.Commit()

		}
		close(done)
	}()
	f, _ := os.Create("ringo.prof")
	pprof.StartCPUProfile(f)
	for i := int64(0); i < 64; i++ {
		master.Reserve()
		master.Commit()
	}
	pprof.StopCPUProfile()
	<-done
}

// Simple Queue Benchmark.
// go test -run=XXX -bench .

// MBP 13-inch, Mid 2009
// 2.53 GHz Intel Core 2 Duo
// 4 GB 1067 MHz DDR3
// OSX 10.10.3

// We tried various configurations here, but settled on a flat architecture.
// Dereferencing pointers and passing variables eats considerable CPU.
//
// Result: 98.0 million transactions per second (10.2 ns/op)
//
func BenchmarkSimpleQueue(b *testing.B) {
	// Set to one process.
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(prevProcs)
	interations := int64(b.N)

	master := SimplePublishNodeNew(PT64Meg)
	slave := SimpleConsumeNodeNew(PT64Meg)
	master.SetDependency(slave.Committed())
	slave.SetDependency(master.Committed())

	done := make(chan bool)

	go func() {
		for i := int64(0); i < interations; i++ {
			slave.Reserve()
			slave.Commit()
		}
		close(done)
	}()

	b.ReportAllocs()
	b.ResetTimer()
	for i := int64(0); i < interations; i++ {
		master.Reserve()
		master.Commit()
	}
	b.StopTimer()
	<-done
}

// Multi Queue Benchmark.
// Same spec as above.
// Because this depends on a lock in the master publisher so it can be shared, its slower.
// Result: 32.9 million transactions per second (30.4 ns/op)
//
func BenchmarkMultiQueue(b *testing.B) {
	// Set to one process.
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(prevProcs)
	interations := int64(b.N)

	master := MultiPublishNodeNew(PT64Meg)
	slave := SimpleConsumeNodeNew(PT64Meg)
	master.SetDependency(slave.Committed())
	slave.SetDependency(master.Committed())

	done := make(chan bool)

	go func() {
		for i := int64(0); i < interations; i++ {
			slave.Reserve()
			slave.Commit()
		}
		close(done)
	}()

	b.ReportAllocs()
	b.ResetTimer()
	for i := int64(0); i < interations; i++ {
		master.Reserve()
		master.Commit()
	}

	b.StopTimer()
	<-done
}
