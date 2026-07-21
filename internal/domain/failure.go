// Package domain holds the singleton feature's value types.
//
// It depends on nothing outside the standard library: no retry engine, no
// synchronization primitives, no adapters.
package domain

import "fmt"

// FailureReason explains why initialization stopped.
type FailureReason uint8

const (
	// FailurePermanent means the factory returned an error wrapped with
	// [Permanent], so no further attempts were made.
	FailurePermanent FailureReason = iota + 1

	// FailureExhausted means the attempt budget ran out.
	FailureExhausted

	// FailureTimedOut means the initialization deadline elapsed.
	FailureTimedOut

	// FailureCanceled means the initialization context was cancelled.
	FailureCanceled
)

// chainCapacity is how many errors an [InitError] can unwrap to: the factory
// error and the retry policy's stop condition.
const chainCapacity = 2

// String returns a short human-readable description of r.
func (r FailureReason) String() string {
	names := [...]string{
		FailurePermanent: "permanent failure",
		FailureExhausted: "retries exhausted",
		FailureTimedOut:  "initialization timed out",
		FailureCanceled:  "initialization canceled",
	}

	if int(r) >= len(names) || names[r] == "" {
		return "initialization failed"
	}

	return names[r]
}

// InitError reports that shared initialization failed.
//
// Classify a failure with Reason, not with errors.Is. [InitError.Unwrap]
// deliberately exposes both the factory error and the reason retrying stopped,
// so errors.Is answers whether an error appears anywhere in the chain — a
// different question from why initialization stopped. A factory that fails
// with context.DeadlineExceeded on every attempt exhausts its retry budget, so
// Reason is [FailureExhausted] even though errors.Is reports a match for
// context.DeadlineExceeded.
//
// Build one with [NewInitError].
type InitError struct {
	// Reason reports why initialization stopped. It is the authoritative
	// classification.
	Reason FailureReason

	// Err is the last error the factory returned.
	Err error

	chain []error
}

// NewInitError builds an InitError.
//
// cause is the retry policy's own stop condition rather than anything inferred
// from err. It is not exported on the result but stays reachable through
// [InitError.Unwrap], so no retry library leaks into the public API.
func NewInitError(reason FailureReason, err, cause error) *InitError {
	initErr := &InitError{
		Reason: reason,
		Err:    err,
		chain:  nil,
	}

	chain := make([]error, 0, chainCapacity)

	if err != nil {
		chain = append(chain, err)
	}

	if cause != nil {
		chain = append(chain, cause)
	}

	if len(chain) > 0 {
		initErr.chain = chain
	}

	return initErr
}

// Error formats the reason and the factory error as
// "singleton: <reason>: <err>".
func (e *InitError) Error() string {
	return fmt.Sprintf("singleton: %s: %v", e.Reason, e.Err)
}

// Unwrap returns the factory error together with the retry policy's stop
// condition, so errors.Is and errors.As match either.
//
// Because Unwrap returns a slice, the single-error errors.Unwrap returns nil
// for an InitError.
func (e *InitError) Unwrap() []error {
	if e.chain != nil {
		return e.chain
	}

	if e.Err != nil {
		return []error{e.Err}
	}

	return nil
}
