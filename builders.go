package ringo

// SimplePair is a simple network of sequencers wired together as dependencies.
type SimplePair struct {
	Leader   *SeqSimple
	Follower *SeqSimple
}

// BuildSimplePair wires up and returns a simple network of sequencers.
// SeqSimple(publisher/leader) <== 1:1 ==> SeqSimple(consumer/follower)
func SimplePairNew(size int64) *SimplePair {
	s := &SimplePair{
		Leader:   SeqSimpleNew(size, nil, true),
		Follower: SeqSimpleNew(size, nil, false),
	}
	s.Leader.SetDependency(s.Follower)
	s.Follower.SetDependency(s.Leader)
	return s
}
