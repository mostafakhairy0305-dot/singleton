package domain

import "time"

type RetryEvent struct {
	Attempt   uint
	Err       error
	NextDelay time.Duration
}

type RetryObserver func(RetryEvent)

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
