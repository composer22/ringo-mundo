package ringo

import (
	"runtime"
	"testing"
)

// A simple queue: Publisher <==> Consumer
func TestSimpleQueueSmallType1(t *testing.T) {
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(runtime.NumCPU())
	defer runtime.GOMAXPROCS(prevProcs)

	master := NewSimplePublishNode(32)
	slave := NewSimpleConsumeNode()
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

// A simple queue: Publisher <==> Consumer with larger buffer.
func TestSimpleQueueLargeType1(t *testing.T) {
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(runtime.NumCPU())
	defer runtime.GOMAXPROCS(prevProcs)

	master := NewSimplePublishNode(PT32Meg)
	slave := NewSimpleConsumeNode()
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
func TestMultiQueueSmallType1(t *testing.T) {
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(runtime.NumCPU())
	defer runtime.GOMAXPROCS(prevProcs)

	master := NewMultiPublishNode(32)
	slave := NewSimpleConsumeNode()
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
func TestSimpleQueueSmallType2(t *testing.T) {
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(runtime.NumCPU())
	defer runtime.GOMAXPROCS(prevProcs)

	master := NewSimpleNode(true, 32)
	slave := NewSimpleNode(false, 32)
	master.SetDependency(slave.Committed())
	slave.SetDependency(master.Committed())

	done := make(chan bool)

	go func() {
		for i := int64(0); i < 64; i++ {
			ndx := slave.Reserve()
			slave.Commit(ndx)
		}
		close(done)
	}()

	for i := int64(0); i < 64; i++ {
		ndx := master.Reserve()
		master.Commit(ndx)
	}
	<-done
}

// A simple queue: Publisher <==> Consumer With larger buffer.
func TestSimpleQueueLargeType2(t *testing.T) {
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(runtime.NumCPU())
	defer runtime.GOMAXPROCS(prevProcs)

	master := NewSimpleNode(true, PT32Meg)
	slave := NewSimpleNode(false, PT32Meg)
	master.SetDependency(slave.Committed())
	slave.SetDependency(master.Committed())

	done := make(chan bool)

	go func() {
		for i := int64(0); i < PT64Meg; i++ {
			ndx := slave.Reserve()
			slave.Commit(ndx)
		}
		close(done)
	}()

	for i := int64(0); i < PT64Meg; i++ {
		ndx := master.Reserve()
		master.Commit(ndx)
	}
	<-done
}

// A Multi Publisher queue: n-Publishers <==> 1 Consumer
func TestMultiQueueSmallType2(t *testing.T) {
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(runtime.NumCPU())
	defer runtime.GOMAXPROCS(prevProcs)

	master := NewMultiNode(true, 32)
	slave := NewSimpleNode(false, 32)
	master.SetDependency(slave.Committed())
	slave.SetDependency(master.Committed())

	done := make(chan bool)

	go func() {
		for i := int64(0); i < 64; i++ {
			ndx := slave.Reserve()
			slave.Commit(ndx)

		}
		close(done)
	}()
	for i := int64(0); i < 64; i++ {
		ndx := master.Reserve()
		master.Commit(ndx)
	}
	<-done
}

// BENCHMARKING TESTS
// go test -run=XXX -bench .

func BenchmarkSimpleQueueType1(b *testing.B) {
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(runtime.NumCPU())
	defer runtime.GOMAXPROCS(prevProcs)
	interations := int64(b.N)

	master := NewSimplePublishNode(PT64Meg)
	slave := NewSimpleConsumeNode()
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

func BenchmarkMultiQueueType1(b *testing.B) {
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(runtime.NumCPU())
	defer runtime.GOMAXPROCS(prevProcs)
	interations := int64(b.N)

	master := NewMultiPublishNode(PT64Meg)
	slave := NewSimpleConsumeNode()
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

func BenchmarkSimpleQueueType2(b *testing.B) {
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(runtime.NumCPU())
	defer runtime.GOMAXPROCS(prevProcs)
	interations := int64(b.N)

	master := NewSimpleNode(true, PT64Meg)
	slave := NewSimpleNode(false, PT64Meg)
	master.SetDependency(slave.Committed())
	slave.SetDependency(master.Committed())

	done := make(chan bool)

	go func() {
		for i := int64(0); i < interations; i++ {
			ndx := slave.Reserve()
			slave.Commit(ndx)
		}
		close(done)
	}()

	b.ReportAllocs()
	b.ResetTimer()
	for i := int64(0); i < interations; i++ {
		ndx := master.Reserve()
		master.Commit(ndx)
	}
	b.StopTimer()
	<-done
}

func BenchmarkMultiQueueType2(b *testing.B) {
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(runtime.NumCPU())
	defer runtime.GOMAXPROCS(prevProcs)
	interations := int64(b.N)

	master := NewMultiNode(true, PT64Meg)
	slave := NewSimpleNode(false, PT64Meg)
	master.SetDependency(slave.Committed())
	slave.SetDependency(master.Committed())

	done := make(chan bool)

	go func() {
		for i := int64(0); i < interations; i++ {
			ndx := slave.Reserve()
			slave.Commit(ndx)
		}
		close(done)
	}()

	b.ReportAllocs()
	b.ResetTimer()
	for i := int64(0); i < interations; i++ {
		ndx := master.Reserve()
		master.Commit(ndx)
	}
	b.StopTimer()
	<-done
}

// Baseline queue test using Golang Channel
func BenchmarkChannelCompare(b *testing.B) {
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(runtime.NumCPU())
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
