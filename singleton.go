// Package singleton provides lazy, retryable, process-local singletons.
//
// A [Provider] initializes one shared value on first use and returns that same
// value to every later caller. Initialization runs on its own goroutine under a
// context owned by this package, while Get waits under the caller's context.
// Cancelling a caller therefore stops only that caller waiting: a request that
// times out cannot poison the singleton for the rest of the process.
//
// Failed attempts are retried with exponential backoff and jitter until the
// attempt budget or the initialization timeout is spent, after which the
// failure is cached. Wrap an error with [Permanent] to stop retrying at once,
// and call Reset to discard a failed initialization so the next Get starts a
// new one.
//
//	var client = singleton.MustNew(func(ctx context.Context) (*redis.Client, error) {
//		c := redis.NewClient(options)
//
//		if err := c.Ping(ctx).Err(); err != nil {
//			_ = c.Close()
//
//			return nil, fmt.Errorf("ping redis: %w", err)
//		}
//
//		return c, nil
//	})
//
//	func Client(ctx context.Context) (*redis.Client, error) {
//		return client.Get(ctx)
//	}
//
// This package requires Go 1.24 or later.
package singleton

import (
	"context"
	"errors"

	"github.com/mostafakhairy0305-dot/singleton/internal/adapters/backoffretry"
	"github.com/mostafakhairy0305-dot/singleton/internal/application"
	"github.com/mostafakhairy0305-dot/singleton/internal/domain"
	"github.com/mostafakhairy0305-dot/singleton/internal/ports"
)

// Factory creates the singleton value.
//
// The context belongs to this package: it carries the deadline set by
// [WithInitializationTimeout], not any caller's deadline.
//
// A factory that returns an error must release whatever it has already
// acquired. Every failed attempt's return value is discarded, so a factory that
// dials a connection and then fails validation leaks one connection per
// attempt. Wrap the error with [Permanent] when retrying cannot help.
//
// Factory is an alias for func(context.Context) (T, error).
type Factory[T any] = ports.Operation[T]

// Provider lazily initializes and returns one shared value.
//
// Create one with [New] or [MustNew]. The zero value is not usable and a
// Provider must not be copied after first use. It is safe for concurrent use.
//
// Provider is an alias for an internal type, so its methods are not listed
// below:
//
//	func (p *Provider[T]) Get(ctx context.Context) (T, error)
//	func (p *Provider[T]) Reset()
//
// Get waits for and returns the shared value, starting initialization on the
// first call and blocking until it settles. It returns the zero T and an
// [InitError] if initialization failed, or the zero T and an error wrapping
// context.Cause(ctx) if the caller's context ended first, which errors.Is
// matches against the cause. It re-panics with the factory's panic value
// if the factory panicked, and panics if ctx is nil or if the Provider is the
// zero value. A failed initialization is cached and returned to every later
// caller until Reset is called.
//
// Reset discards a failed or panicked initialization so the next Get starts a
// new one. It does nothing after a success, nothing while initialization is in
// flight, and nothing before the first Get.
type Provider[T any] = application.Provider[T]

// InitError reports that shared initialization failed.
//
// Every initialization failure returned by Get is an *InitError, so
// errors.As has no silent false branch. A failure that is not an *InitError
// came from the caller's own context rather than from initialization.
//
// InitError is an alias for an internal type, so its fields and methods are not
// listed below:
//
//	type InitError struct {
//		Reason FailureReason // why initialization stopped
//		Err    error         // the last error the factory returned
//	}
//
//	func (e *InitError) Error() string
//	func (e *InitError) Unwrap() []error
//
// Error formats as "singleton: <reason>: <err>".
//
// Unwrap returns the factory error together with the retry policy's stop
// condition, so errors.Is matches either. Because it returns a slice, the
// single-error errors.Unwrap returns nil for an InitError.
//
// Classify a failure with Reason, not with errors.Is: a factory that fails with
// context.DeadlineExceeded on every attempt exhausts its retry budget, so
// Reason is [FailureExhausted] even though errors.Is reports a match for
// context.DeadlineExceeded.
type InitError = domain.InitError

// FailureReason explains why initialization stopped.
//
// Its String method returns a short human-readable description: "permanent
// failure", "retries exhausted", "initialization timed out" or "initialization
// canceled".
type FailureReason = domain.FailureReason

const (
	// FailurePermanent means the factory returned an error wrapped with
	// [Permanent], so no further attempts were made.
	FailurePermanent = domain.FailurePermanent

	// FailureExhausted means the attempt budget set by [WithMaxAttempts] ran
	// out.
	FailureExhausted = domain.FailureExhausted

	// FailureTimedOut means the deadline set by [WithInitializationTimeout]
	// elapsed.
	FailureTimedOut = domain.FailureTimedOut

	// FailureCanceled means the initialization context was cancelled.
	//
	// No current code path produces it: the initialization context descends
	// from context.Background with only a timeout, and nothing cancels it. It
	// is reserved for a future shutdown hook, so do not write a code path that
	// depends on receiving it.
	FailureCanceled = domain.FailureCanceled
)

// RetryEvent describes a failed attempt that will be retried. It is delivered
// to the observer registered with [WithRetryObserver].
//
// RetryEvent is an alias for an internal type, so its fields are not listed
// below:
//
//	type RetryEvent struct {
//		Attempt   uint          // 1-based number of the attempt that just failed
//		Err       error         // why that attempt failed
//		NextDelay time.Duration // wait before the next attempt
//	}
type RetryEvent = domain.RetryEvent

var (
	errNilFactory    = errors.New("singleton: factory is nil")
	errInvalidOption = errors.New("singleton: invalid option")
)

// Interface is the behaviour [Provider] implements.
//
// Depend on it in consumers that need to substitute a fake.
type Interface[T any] interface {
	// Get waits for and returns the shared value.
	Get(ctx context.Context) (T, error)

	// Reset discards a failed initialization so the next Get starts a new one.
	Reset()
}

// Permanent marks a factory error as non-retriable. Initialization stops at
// that attempt and reports [FailurePermanent].
//
// Permanent(nil) returns nil. The wrapped error stays reachable through
// errors.Is and errors.As, but the concrete type returned is internal and
// cannot be type-asserted from outside this package.
func Permanent(err error) error {
	if err == nil {
		return nil
	}

	return &domain.PermanentError{Err: err}
}

// New creates a lazy singleton provider.
//
// It does not call factory; the first Get does.
//
// New returns an error if factory is nil, if any option is invalid, or if a
// zero-value [Option] is passed. It never reports factory failures, which
// surface from Get instead.
func New[T any](
	factory Factory[T],
	options ...Option,
) (*Provider[T], error) {
	if factory == nil {
		return nil, errNilFactory
	}

	cfg := defaultConfig()

	for _, option := range options {
		if option.apply == nil {
			return nil, errInvalidOption
		}

		err := option.apply(&cfg)
		if err != nil {
			return nil, err
		}
	}

	cfg.retry.Observer = cfg.retry.Observer.Safe()

	return application.NewProvider(factory, backoffretry.New[T](cfg.retry)), nil
}

// MustNew is like [New] but panics instead of returning an error, which suits
// package-level declarations.
//
// It panics only for invalid construction options, never for factory failures.
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
