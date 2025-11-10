package pool

import "sync"

type Resetter interface{ Reset() }

type Pool[T Resetter] struct {
	pool sync.Pool
}

func New[T Resetter](newFunc func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{New: func() any { return newFunc() }},
	}
}

func (p *Pool[T]) Get() T {
	v := p.pool.Get()
	if v == nil {
		var zero T
		return zero
	}
	return v.(T)
}

func (p *Pool[T]) Put(x T) {
	if any(x) == nil {
		p.pool.Put(x)
		return
	}
	x.Reset()
	p.pool.Put(x)
}
