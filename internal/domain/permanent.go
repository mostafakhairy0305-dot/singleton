package domain

type PermanentError struct {
	Err error
}

func (e *PermanentError) Error() string {
	return e.Err.Error()
}

func (e *PermanentError) Unwrap() error {
	return e.Err
}

func Permanent(err error) error {
	if err == nil {
		return nil
	}

	return &PermanentError{Err: err}
}
