package ringo

type POCProducer struct {
	committed int64
	cachepad2 [7]int64
}

func (p *POCProducer) Commit() {
	p.committed++
}

type POCConsumer struct {
	dependency *int64
}

func (p *POCConsumer) Read() int64 {
	return *p.dependency
}
