package ringo

import (
	"runtime"
	"testing"
)

// Simple buffer check.
// 49 million transactions per second. (20.5 ns/op)
func BenchmarkSimpleRW(b *testing.B) {
	// Set to one process.
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(prevProcs)

	interations := int64(b.N)

	m := SimpleQueueNew(PT64Meg)
	done := make(chan bool)

	go func(d chan bool) {
		for i := int64(0); i < interations; i++ {
			j := m.Consumer.ConsumeReserve()
			m.Consumer.Commit(j)
		}
		close(d)
	}(done)

	b.ReportAllocs()
	b.ResetTimer()

	for i := int64(0); i < interations; i++ {
		j := m.Publisher.PublishReserve()
		m.Publisher.Commit(j)
	}
	b.StopTimer()
	<-done
}
