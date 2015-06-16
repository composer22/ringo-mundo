package ringo

import (
	"runtime"
	"testing"
)

// A simple queue: Publisher <==> Consumer
//
func TestSimpleQueueSmall(t *testing.T) {
	// Set to one process.
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(runtime.NumCPU())
	defer runtime.GOMAXPROCS(prevProcs)

	master := SimplePublishNodeNew(32)
	slave := SimpleConsumeNodeNew()
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

// A simple queue: Publisher <==> Consumer
// With larger buffer.
func TestSimpleQueueLarge(t *testing.T) {
	// Set to one process.
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(runtime.NumCPU())
	defer runtime.GOMAXPROCS(prevProcs)

	master := SimplePublishNodeNew(PT64Meg)
	slave := SimpleConsumeNodeNew()
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

// A Multi Publisher queue: n-Publishers <==> 1 Consumer
func TestMultiQueueSmall(t *testing.T) {
	// Set to one process.
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(runtime.NumCPU())
	defer runtime.GOMAXPROCS(prevProcs)

	master := MultiPublishNodeNew(32)
	slave := SimpleConsumeNodeNew()
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

// Simple Queue Benchmark.
// go test -run=XXX -bench .

func BenchmarkSimpleQueue(b *testing.B) {
	// Set to one process.
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(1) // runtime.NumCPU()
	defer runtime.GOMAXPROCS(prevProcs)
	interations := int64(b.N)

	master := SimplePublishNodeNew(PT64Meg)
	slave := SimpleConsumeNodeNew()
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

// Multi publisher Queue Benchmark.
// Because this depends on a lock in the master publisher so it can be shared, its slower.
func BenchmarkMultiQueue(b *testing.B) {
	// Set to one process.
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(1) // runtime.NumCPU()
	defer runtime.GOMAXPROCS(prevProcs)
	interations := int64(b.N)

	master := MultiPublishNodeNew(PT64Meg)
	slave := SimpleConsumeNodeNew()
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

// Baseline queue test using Golang Channel
//
func BenchmarkChannelCompare(b *testing.B) {
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(1) //runtime.NumCPU()
	defer runtime.GOMAXPROCS(prevProcs)
	interations := int64(b.N)

	queue := make(chan bool, interations)
	done := make(chan bool)

	go func(q chan bool, d chan bool) {
		i := int64(0)
		for i < interations {
			select {
			case <-q:
				i++
			default:
			}
			runtime.Gosched()
		}
		d <- true
	}(queue, done)

	b.ReportAllocs()
	b.ResetTimer()
	for i := int64(0); i < interations; i++ {
		queue <- true
	}

	b.StopTimer()
	<-done
}
