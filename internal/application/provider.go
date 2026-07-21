package application

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/mostafa-khairy-zofirm/singleton/internal/ports"
)

type Provider[T any] struct {
	factory ports.Operation[T]
	retrier ports.Retrier[T]

	mu      sync.Mutex
	current atomic.Pointer[state[T]]
}

func NewProvider[T any](
	factory ports.Operation[T],
	retrier ports.Retrier[T],
) *Provider[T] {
	return &Provider[T]{
		factory: factory,
		retrier: retrier,
	}
}

func (p *Provider[T]) Get(ctx context.Context) (T, error) {
	if ctx == nil {
		panic("singleton: nil context")
	}

	s := p.load()

	if s.settled.Load() {
		return s.result()
	}

	select {
	case <-s.done:
		return s.result()

	case <-ctx.Done():
		if s.settled.Load() {
			return s.result()
		}

		var zero T

		return zero, context.Cause(ctx)
	}
}

func (p *Provider[T]) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()

	s := p.current.Load()
	if s == nil {
		return
	}

	if !s.settled.Load() {
		return
	}

	if s.panicked || s.err != nil {
		p.current.Store(nil)
	}
}

func (p *Provider[T]) load() *state[T] {
	if s := p.current.Load(); s != nil {
		return s
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if s := p.current.Load(); s != nil {
		return s
	}

	if p.factory == nil {
		panic("singleton: Provider must be created with New or MustNew")
	}

	s := &state[T]{
		done: make(chan struct{}),
	}

	p.current.Store(s)

	go p.initialize(s)

	return s
}

type state[T any] struct {
	settled atomic.Bool
	done    chan struct{}

	value T
	err   error

	panicked   bool
	panicValue any
}

func (s *state[T]) result() (T, error) {
	if s.panicked {
		panic(s.panicValue)
	}

	return s.value, s.err
}

func (p *Provider[T]) initialize(s *state[T]) {
	defer func() {
		if value := recover(); value != nil {
			s.panicked = true
			s.panicValue = value
		}

		s.settled.Store(true)
		close(s.done)
	}()

	value, err := p.retrier.Do(context.Background(), p.factory)
	if err != nil {
		s.err = err

		return
	}

	s.value = value
}
