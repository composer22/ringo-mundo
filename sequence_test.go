package ringo

import (
	"runtime"
	"testing"
)

const (
	testRingSize = int64(1024)
)

//func TestReserveLeader(t *testing.T) {
//	m := SimpleQueueNew(testRingSize)

//	// Store a batch of jobs by the Producer.
//	for i := 0; i < 8; i++ {
//		j := m.Publisher.Reserve()
//		m.Publisher.Commit(j)
//	}

//	// Pretend we did the work in the Consumer.
//	for i := 0; i < 8; i++ {
//		j := m.Consumer.Reserve()
//		m.Consumer.Commit(j)
//	}

//	// Now rotate past beginning again w/ more jobs up to the last cell.
//	for i := 0; i < 1022; i++ {
//		j := m.Publisher.Reserve()
//		m.Publisher.Commit(j)
//	}
//}

func TestBasicsSmall(t *testing.T) {
	// Set to one process.
	prevProcs := runtime.GOMAXPROCS(-1)
	runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(prevProcs)

	m := SimpleQueueNew(32)
	done := make(chan bool)

	go func(d chan bool) {
		for i := int64(0); i < 64; i++ {
			j := m.Consumer.ConsumeReserve()
			m.Consumer.Commit(j)
		}
		close(d)
	}(done)

	for i := int64(0); i < 64; i++ {
		j := m.Publisher.PublishReserve()
		m.Publisher.Commit(j)
	}
	<-done
}

//func TestBasicsLarge(t *testing.T) {
//	return
//	// Set to one process.
//	prevProcs := runtime.GOMAXPROCS(-1)
//	runtime.GOMAXPROCS(1)
//	defer runtime.GOMAXPROCS(prevProcs)

//	m := SimpleQueueNew(PT16Meg)
//	done := make(chan bool)

//	go func(d chan bool) {
//		for i := int64(0); i < PT64Meg; i++ {
//			j := m.Consumer.ConsumeReserve()
//			m.Consumer.Commit(j)
//		}
//		close(d)
//	}(done)

//	for i := int64(0); i < PT64Meg; i++ {
//		j := m.Publisher.PublishReserve()
//		m.Publisher.Commit(j)

//	}
//	<-done
//}
