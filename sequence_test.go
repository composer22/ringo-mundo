package ringo

import "testing"

const (
	testRingSize = int64(1024)
)

func TestReserveLeader(t *testing.T) {
	m := SimplePairNew(testRingSize)

	// Store a batch of jobs by the Producer.
	for i := 0; i < 8; i++ {
		j := m.Leader.Reserve()
		m.Leader.Commit(j)
	}

	// Pretend we did the work in the Consumer.
	for i := 0; i < 8; i++ {
		m.Follower.committed[i] = 0
	}

	// Now rotate past beginning again w/ more jobs up to the last cell.
	for i := 0; i < 1024; i++ {
		j := m.Leader.Reserve()
		m.Leader.Commit(j)
	}
}

func TestReserveFollower(t *testing.T) {
	m := SimplePairNew(testRingSize)

	// Pretend the Producer committed a ring of jobs.
	for i := int64(0); i < testRingSize; i++ {
		j := m.Leader.Reserve()
		m.Leader.Commit(j)
	}

	// Now test the follower can read those jobs.
	for i := int64(0); i < testRingSize; i++ {
		j := m.Follower.Reserve()
		m.Follower.Commit(j)
	}

	// Pretend the Producer committed another rotation of jobs.
	for i := int64(0); i < testRingSize; i++ {
		j := m.Leader.Reserve()
		m.Leader.Commit(j)
	}

	// Now test the follower can read another rotation of jobs.
	for i := int64(0); i < testRingSize; i++ {
		j := m.Follower.Reserve()
		m.Follower.Commit(j)
	}
}
