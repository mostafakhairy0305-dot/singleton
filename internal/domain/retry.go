package domain

import "time"

// RetryEvent describes a failed attempt that will be retried.
type RetryEvent struct {
	// Attempt is the 1-based number of the attempt that just failed.
	Attempt uint

	// Err is why that attempt failed.
	Err error

	// NextDelay is how long the policy waits before the next attempt.
	NextDelay time.Duration
}

// RetryObserver receives one event per retried attempt.
type RetryObserver func(RetryEvent)

// Safe returns an observer that recovers from panics in o, or nil if o is nil.
//
// Instrumentation must never become the singleton's result. Without this, a
// panicking observer is caught by the recover that guards factory panics, and
// every later call for the life of the process re-panics with it. The
// composition root applies Safe before handing an observer to any adapter, so
// no adapter can violate the invariant.
func (o RetryObserver) Safe() RetryObserver {
	if o == nil {
		return nil
	}

	return func(event RetryEvent) {
		defer func() {
			_ = recover()
		}()

		o(event)
	}
}
