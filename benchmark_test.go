package ringo

import (
	"runtime"
	"testing"
)

// Simple buffer check.  Seems to do well with only 1 proc on my machine.
// 36.6 ns/op = about 27.3 million ops per second
func BenchmarkSimpleWriteRead(b *testing.B) {
	// Set to one process.
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(prevProcs)

	interations := int64(b.N)

	m := SimplePairNew(PT64Meg)
	done := make(chan bool)

	go func() {
		for i := int64(0); i < interations; i++ {
			j := m.Follower.Reserve()
			m.Follower.Commit(j)
		}
		close(done)
	}()

	b.ReportAllocs()
	b.ResetTimer()

	for i := int64(0); i < interations; i++ {
		j := m.Leader.Reserve()
		m.Leader.Commit(j)

	}
	b.StopTimer()
	<-done
}
