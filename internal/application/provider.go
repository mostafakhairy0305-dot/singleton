// Package application implements the singleton feature's use cases.
//
// It depends on the domain and on outbound ports, never on an adapter: no
// retry engine is imported here.
package application

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/mostafa-khairy-zofirm/singleton/internal/ports"
)

// Provider lazily initializes and returns one shared value.
//
// Build one with [NewProvider]. The zero value is not usable and a Provider
// must not be copied after first use. It is safe for concurrent use.
type Provider[T any] struct {
	factory ports.Operation[T]
	retrier ports.Retrier[T]

	mu      sync.Mutex
	current atomic.Pointer[state[T]]
}

// NewProvider wires a factory to the retry policy that will drive it.
func NewProvider[T any](
	factory ports.Operation[T],
	retrier ports.Retrier[T],
) *Provider[T] {
	return &Provider[T]{
		factory: factory,
		retrier: retrier,
	}
}

// Get waits for and returns the shared value, starting initialization on the
// first call and blocking until it settles.
//
// The caller's context cancels only this caller's wait. It does not cancel the
// shared initialization, so one short-lived request cannot poison the
// singleton for the entire process. A failed initialization is cached and
// returned to every later caller until [Provider.Reset] is called.
//
// Get returns the zero T with context.Cause(ctx) if the caller's context ends
// before initialization settles, and re-panics with the factory's panic value
// if the factory panicked. It panics if ctx is nil or if the Provider is the
// zero value.
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

// Reset discards a failed or panicked initialization so the next
// [Provider.Get] starts a new one.
//
// It does nothing while initialization is in progress, so callers already
// waiting are never left on a discarded state, and nothing after a success,
// because a live value other goroutines already hold must not be torn down.
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
