package singleton

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

var (
	errBoom  = errors.New("boom")
	errFatal = errors.New("fatal")
)

func okFactory(context.Context) (int, error) { return 1, nil }

// fastOptions keep the retry budget small and its delays short, so a test that
// drives initialization to failure costs milliseconds.
func fastOptions() []Option {
	return []Option{
		WithMaxAttempts(3),
		WithInitializationTimeout(2 * time.Second),
		WithRetryInterval(time.Millisecond, 2*time.Millisecond),
	}
}

func requireInitError(t *testing.T, err error) *InitError {
	t.Helper()

	var initErr *InitError
	if !errors.As(err, &initErr) {
		t.Fatalf("errors.As(%v, *InitError) = false, want true", err)
	}

	return initErr
}

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	cfg := defaultConfig().retry

	tests := map[string]struct{ got, want any }{
		"max attempts":     {got: cfg.MaxAttempts, want: uint(defaultMaxAttempts)},
		"timeout":          {got: cfg.Timeout, want: defaultTimeout},
		"initial interval": {got: cfg.InitialInterval, want: defaultInitialInterval},
		"maximum interval": {got: cfg.MaxInterval, want: defaultMaxInterval},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if test.got != test.want {
				t.Errorf("%s = %v, want %v", name, test.got, test.want)
			}
		})
	}
}

func TestDefaultConfigRegistersNoObserver(t *testing.T) {
	t.Parallel()

	if defaultConfig().retry.Observer != nil {
		t.Error("Observer = non-nil, want nil until WithRetryObserver registers one")
	}
}

func TestNewRejectsInvalidConstruction(t *testing.T) {
	t.Parallel()

	var zeroOption Option

	tests := map[string]struct {
		factory Factory[int]
		options []Option
		want    error
	}{
		"nil factory": {factory: nil, options: nil, want: errNilFactory},
		"zero option": {
			factory: okFactory,
			options: []Option{zeroOption},
			want:    errInvalidOption,
		},
		"zero max attempts": {
			factory: okFactory,
			options: []Option{WithMaxAttempts(0)},
			want:    errZeroMaxAttempts,
		},
		"negative timeout": {
			factory: okFactory,
			options: []Option{WithInitializationTimeout(-time.Second)},
			want:    errNegativeTimeout,
		},
		"zero initial interval": {
			factory: okFactory,
			options: []Option{WithRetryInterval(0, time.Second)},
			want:    errZeroInitialInterval,
		},
		"maximum below initial": {
			factory: okFactory,
			options: []Option{WithRetryInterval(2*time.Second, time.Second)},
			want:    errMaxIntervalBelowInitial,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			provider, err := New(test.factory, test.options...)
			if provider != nil {
				t.Errorf("New() provider = %v, want nil", provider)
			}

			if !errors.Is(err, test.want) {
				t.Errorf("New() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestNewAppliesEveryOption(t *testing.T) {
	t.Parallel()

	options := append(
		fastOptions(),
		WithRetryObserver(func(RetryEvent) {}),
		// A nil observer is allowed and simply disables the callback.
		WithRetryObserver(nil),
		// A zero timeout disables the deadline.
		WithInitializationTimeout(0),
	)

	provider, err := New(okFactory, options...)
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}

	got, err := provider.Get(context.Background())
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}

	if got != 1 {
		t.Errorf("Get() = %d, want 1", got)
	}
}

func TestMustNewReturnsAProvider(t *testing.T) {
	t.Parallel()

	provider := MustNew(okFactory, fastOptions()...)

	got, err := provider.Get(context.Background())
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}

	if got != 1 {
		t.Errorf("Get() = %d, want 1", got)
	}
}

func TestMustNewPanicsOnInvalidConstruction(t *testing.T) {
	t.Parallel()

	var recovered any

	func() {
		defer func() { recovered = recover() }()

		_ = MustNew[int](nil)
	}()

	err, ok := recovered.(error)
	if !ok {
		t.Fatalf("recovered %v, want an error", recovered)
	}

	if !errors.Is(err, errNilFactory) {
		t.Errorf("recovered error = %v, want %v", err, errNilFactory)
	}
}

func TestPermanentReturnsNilForNil(t *testing.T) {
	t.Parallel()

	got := Permanent(nil)
	if got != nil {
		t.Errorf("Permanent(nil) = %v, want nil", got)
	}
}

func TestPermanentStopsRetryingAtTheFirstAttempt(t *testing.T) {
	t.Parallel()

	var calls atomic.Int64

	provider := MustNew(func(context.Context) (int, error) {
		calls.Add(1)

		return 0, Permanent(errFatal)
	}, fastOptions()...)

	_, err := provider.Get(context.Background())

	initErr := requireInitError(t, err)
	if initErr.Reason != FailurePermanent {
		t.Errorf("Reason = %v, want %v", initErr.Reason, FailurePermanent)
	}

	if !errors.Is(err, errFatal) {
		t.Errorf("Get() error = %v, want it to wrap %v", err, errFatal)
	}

	if calls.Load() != 1 {
		t.Errorf("the factory ran %d times, want 1", calls.Load())
	}
}

func TestGetReportsAnExhaustedBudget(t *testing.T) {
	t.Parallel()

	provider := MustNew(func(context.Context) (int, error) {
		return 0, errBoom
	}, fastOptions()...)

	_, err := provider.Get(context.Background())

	initErr := requireInitError(t, err)
	if initErr.Reason != FailureExhausted {
		t.Errorf("Reason = %v, want %v", initErr.Reason, FailureExhausted)
	}

	if !errors.Is(err, errBoom) {
		t.Errorf("Get() error = %v, want it to wrap %v", err, errBoom)
	}
}

func TestResetStartsANewInitialization(t *testing.T) {
	t.Parallel()

	var calls atomic.Int64

	// Depending on the exported interface is how a consumer substitutes a fake.
	var provider Interface[int] = MustNew(func(context.Context) (int, error) {
		if calls.Add(1) <= 3 {
			return 0, errBoom
		}

		return 8, nil
	}, fastOptions()...)

	_, err := provider.Get(context.Background())
	if err == nil {
		t.Fatal("Get() error = nil, want the exhausted budget")
	}

	provider.Reset()

	got, err := provider.Get(context.Background())
	if err != nil {
		t.Fatalf("Get() after Reset error = %v, want nil", err)
	}

	if got != 8 {
		t.Errorf("Get() after Reset = %d, want 8", got)
	}
}

func TestGetReportsTheInitializationTimeout(t *testing.T) {
	t.Parallel()

	provider := MustNew(func(ctx context.Context) (int, error) {
		<-ctx.Done()

		return 0, errBoom
	},
		WithMaxAttempts(5),
		WithInitializationTimeout(20*time.Millisecond),
		WithRetryInterval(time.Millisecond, 2*time.Millisecond),
	)

	_, err := provider.Get(context.Background())

	initErr := requireInitError(t, err)
	if initErr.Reason != FailureTimedOut {
		t.Errorf("Reason = %v, want %v", initErr.Reason, FailureTimedOut)
	}
}

func TestNewMakesTheRetryObserverPanicSafe(t *testing.T) {
	t.Parallel()

	var events atomic.Int64

	provider := MustNew(func(context.Context) (int, error) {
		return 0, errBoom
	}, append(fastOptions(), WithRetryObserver(func(RetryEvent) {
		events.Add(1)

		panic("observer exploded")
	}))...)

	// A panicking observer must not become the singleton's result.
	_, err := provider.Get(context.Background())

	initErr := requireInitError(t, err)
	if initErr.Reason != FailureExhausted {
		t.Errorf("Reason = %v, want %v", initErr.Reason, FailureExhausted)
	}

	if events.Load() != 2 {
		t.Errorf("the observer saw %d events, want 2", events.Load())
	}
}
