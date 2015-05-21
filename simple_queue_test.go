package ringo

import (
	"runtime"
	"testing"
)

func TestQueueSmall(t *testing.T) {
	// Set to one process.
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(prevProcs)

	m := SimpleQueueNew(32)
	done := make(chan bool)

	go func() {
		for i := int64(0); i < 64; i++ {
			m.Slave.Reserve()
			m.Slave.Commit()
		}
		close(done)
	}()

	for i := int64(0); i < 64; i++ {
		m.Master.Reserve()
		m.Master.Commit()
	}
	<-done
}

func TestQueueLarge(t *testing.T) {
	// Set to one process.
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(prevProcs)

	m := SimpleQueueNew(PT64Meg)
	done := make(chan bool)

	go func() {
		for i := int64(0); i < PT64Meg; i++ {
			m.Slave.Reserve()
			m.Slave.Commit()
		}
		close(done)
	}()

	for i := int64(0); i < PT64Meg; i++ {
		m.Master.Reserve()
		m.Master.Commit()
	}
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
// Result: 68.0 million transactions per second (14.7 ns/op)
//
func BenchmarkSimpleQueue(b *testing.B) {
	// Set to one process.
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(prevProcs)

	interations := int64(b.N)

	q := SimpleQueueNew(PT64Meg)
	done := make(chan bool)

	go func() {
		for i := int64(0); i < interations; i++ {
			q.Slave.Reserve()
			q.Slave.Commit()
		}
		close(done)
	}()

	b.ReportAllocs()
	b.ResetTimer()
	for i := int64(0); i < interations; i++ {
		q.Master.Reserve()
		q.Master.Commit()
	}
	b.StopTimer()
	<-done
}
