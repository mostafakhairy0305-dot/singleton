package domain

// PermanentError marks a factory error as non-retriable.
//
// A retry adapter detects it with errors.As and stops immediately.
type PermanentError struct {
	// Err is the wrapped, non-retriable error.
	Err error
}

// Error returns the wrapped error's message.
func (e *PermanentError) Error() string {
	return e.Err.Error()
}

// Unwrap returns the wrapped error.
func (e *PermanentError) Unwrap() error {
	return e.Err
}

// Permanent marks a factory error as non-retriable.
//
// Permanent(nil) returns nil.
func Permanent(err error) error {
	if err == nil {
		return nil
	}

	return &PermanentError{Err: err}
}
