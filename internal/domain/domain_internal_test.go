package domain

import (
	"errors"
	"slices"
	"testing"
	"time"
)

// unknownReason is what [FailureReason.String] falls back to.
const unknownReason = "initialization failed"

var (
	errFactory = errors.New("factory failed")
	errStop    = errors.New("policy stopped")
)

// sameErrors compares two unwrap chains positionally.
func sameErrors(got, want []error) bool {
	return slices.EqualFunc(got, want, errors.Is)
}

func TestFailureReasonString(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		reason FailureReason
		want   string
	}{
		"permanent":         {reason: FailurePermanent, want: "permanent failure"},
		"exhausted":         {reason: FailureExhausted, want: "retries exhausted"},
		"timed out":         {reason: FailureTimedOut, want: "initialization timed out"},
		"canceled":          {reason: FailureCanceled, want: "initialization canceled"},
		"the unnamed zero":  {reason: 0, want: unknownReason},
		"one past the last": {reason: FailureCanceled + 1, want: unknownReason},
		"far past the last": {reason: 200, want: unknownReason},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := test.reason.String(); got != test.want {
				t.Errorf("String() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestNewInitErrorBuildsTheUnwrapChain(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		err   error
		cause error
		want  []error
	}{
		"factory error and stop condition": {
			err:   errFactory,
			cause: errStop,
			want:  []error{errFactory, errStop},
		},
		"factory error only":  {err: errFactory, cause: nil, want: []error{errFactory}},
		"stop condition only": {err: nil, cause: errStop, want: []error{errStop}},
		"neither":             {err: nil, cause: nil, want: nil},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			initErr := NewInitError(FailureExhausted, test.err, test.cause)

			if initErr.Reason != FailureExhausted {
				t.Errorf("Reason = %v, want %v", initErr.Reason, FailureExhausted)
			}

			if !errors.Is(initErr.Err, test.err) {
				t.Errorf("Err = %v, want %v", initErr.Err, test.err)
			}

			if got := initErr.Unwrap(); !sameErrors(got, test.want) {
				t.Errorf("Unwrap() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestInitErrorUnwrapFallsBackToErrWithoutAChain(t *testing.T) {
	t.Parallel()

	// NewInitError always fills the chain, so a bare struct literal is the only
	// way to reach the fallback.
	initErr := &InitError{Reason: FailureExhausted, Err: errFactory, chain: nil}

	if got := initErr.Unwrap(); !sameErrors(got, []error{errFactory}) {
		t.Errorf("Unwrap() = %v, want [%v]", got, errFactory)
	}

	empty := &InitError{Reason: FailureExhausted, Err: nil, chain: nil}

	if got := empty.Unwrap(); got != nil {
		t.Errorf("Unwrap() = %v, want nil", got)
	}
}

func TestInitErrorErrorFormatsReasonAndCause(t *testing.T) {
	t.Parallel()

	initErr := NewInitError(FailurePermanent, errFactory, errStop)

	const want = "singleton: permanent failure: factory failed"

	if got := initErr.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}

	if !errors.Is(initErr, errFactory) {
		t.Errorf("errors.Is(initErr, errFactory) = false, want true")
	}

	if !errors.Is(initErr, errStop) {
		t.Errorf("errors.Is(initErr, errStop) = false, want true")
	}

	// Unwrap returns a slice, so the single-error form finds nothing.
	unwrapped := errors.Unwrap(error(initErr))
	if unwrapped != nil {
		t.Errorf("errors.Unwrap(initErr) = %v, want nil", unwrapped)
	}
}

func TestPermanentReturnsNilForNil(t *testing.T) {
	t.Parallel()

	got := Permanent(nil)
	if got != nil {
		t.Errorf("Permanent(nil) = %v, want nil", got)
	}
}

func TestPermanentWrapsTheError(t *testing.T) {
	t.Parallel()

	got := Permanent(errFactory)

	var permanent *PermanentError
	if !errors.As(got, &permanent) {
		t.Fatalf("errors.As(%v, *PermanentError) = false, want true", got)
	}

	if !errors.Is(permanent.Err, errFactory) {
		t.Errorf("Err = %v, want %v", permanent.Err, errFactory)
	}

	if got.Error() != errFactory.Error() {
		t.Errorf("Error() = %q, want %q", got.Error(), errFactory.Error())
	}

	if !errors.Is(got, errFactory) {
		t.Errorf("errors.Is(got, errFactory) = false, want true")
	}
}

func TestRetryObserverSafeReturnsNilForNil(t *testing.T) {
	t.Parallel()

	var observer RetryObserver

	if observer.Safe() != nil {
		t.Error("Safe() on a nil observer = non-nil, want nil")
	}
}

func TestRetryObserverSafeDeliversTheEvent(t *testing.T) {
	t.Parallel()

	var got RetryEvent

	observer := RetryObserver(func(event RetryEvent) { got = event })
	want := RetryEvent{Attempt: 2, Err: errFactory, NextDelay: time.Second}

	observer.Safe()(want)

	if got != want {
		t.Errorf("observed %+v, want %+v", got, want)
	}
}

func TestRetryObserverSafeRecoversFromAPanic(t *testing.T) {
	t.Parallel()

	called := false

	observer := RetryObserver(func(RetryEvent) {
		called = true

		panic("observer exploded")
	})

	observer.Safe()(RetryEvent{Attempt: 1, Err: errFactory, NextDelay: 0})

	if !called {
		t.Error("the observer was never called")
	}
}
