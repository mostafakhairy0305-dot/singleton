package singleton

import (
	"context"
	"errors"

	"github.com/mostafa-khairy-zofirm/singleton/internal/adapters/backoffretry"
	"github.com/mostafa-khairy-zofirm/singleton/internal/application"
	"github.com/mostafa-khairy-zofirm/singleton/internal/domain"
	"github.com/mostafa-khairy-zofirm/singleton/internal/ports"
)

type (
	Factory[T any]  = ports.Operation[T]
	Provider[T any] = application.Provider[T]
	InitError       = domain.InitError
	FailureReason   = domain.FailureReason
	RetryEvent      = domain.RetryEvent
)

type Interface[T any] interface {
	Get(ctx context.Context) (T, error)
	Reset()
}

const (
	FailurePermanent = domain.FailurePermanent
	FailureExhausted = domain.FailureExhausted
	FailureTimedOut  = domain.FailureTimedOut
	FailureCanceled  = domain.FailureCanceled
)

func Permanent(err error) error {
	return domain.Permanent(err)
}

func New[T any](
	factory Factory[T],
	options ...Option,
) (*Provider[T], error) {
	if factory == nil {
		return nil, errors.New("singleton: factory is nil")
	}

	cfg := defaultConfig()

	for _, option := range options {
		if option.apply == nil {
			return nil, errors.New("singleton: invalid option")
		}

		if err := option.apply(&cfg); err != nil {
			return nil, err
		}
	}

	cfg.retry.Observer = cfg.retry.Observer.Safe()

	return application.NewProvider(factory, backoffretry.New[T](cfg.retry)), nil
}

func MustNew[T any](
	factory Factory[T],
	options ...Option,
) *Provider[T] {
	provider, err := New(factory, options...)
	if err != nil {
		panic(err)
	}

	return provider
}

var _ Interface[int] = (*Provider[int])(nil)
