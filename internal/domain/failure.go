package domain

import "fmt"

type FailureReason uint8

const (
	FailurePermanent FailureReason = iota + 1
	FailureExhausted
	FailureTimedOut
	FailureCanceled
)

func (r FailureReason) String() string {
	switch r {
	case FailurePermanent:
		return "permanent failure"
	case FailureExhausted:
		return "retries exhausted"
	case FailureTimedOut:
		return "initialization timed out"
	case FailureCanceled:
		return "initialization canceled"
	default:
		return "initialization failed"
	}
}

type InitError struct {
	Reason FailureReason
	Err    error
	chain  []error
}

func NewInitError(reason FailureReason, err, cause error) *InitError {
	initErr := &InitError{
		Reason: reason,
		Err:    err,
	}

	chain := make([]error, 0, 2)

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

func (e *InitError) Error() string {
	return fmt.Sprintf("singleton: %s: %v", e.Reason, e.Err)
}

func (e *InitError) Unwrap() []error {
	if e.chain != nil {
		return e.chain
	}

	if e.Err != nil {
		return []error{e.Err}
	}

	return nil
}
