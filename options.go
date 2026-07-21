package singleton

import (
	"errors"
	"time"

	"github.com/mostafa-khairy-zofirm/singleton/internal/adapters/backoffretry"
)

const (
	defaultMaxAttempts     = 5
	defaultTimeout         = 30 * time.Second
	defaultInitialInterval = 250 * time.Millisecond
	defaultMaxInterval     = 5 * time.Second
)

// Option configures a [Provider].
//
// Its implementation is unexported so callers cannot depend on the retry
// library underneath. The zero value is invalid and is rejected by [New].
type Option struct {
	apply func(*config) error
}

type config struct {
	retry backoffretry.Config
}

func defaultConfig() config {
	return config{
		retry: backoffretry.Config{
			MaxAttempts:     defaultMaxAttempts,
			Timeout:         defaultTimeout,
			InitialInterval: defaultInitialInterval,
			MaxInterval:     defaultMaxInterval,
		},
	}
}

// WithMaxAttempts sets the total number of attempts, including the first, so
// WithMaxAttempts(1) never retries. The default is 5.
//
// [New] reports an error if n is zero.
func WithMaxAttempts(n uint) Option {
	return Option{
		apply: func(c *config) error {
			if n == 0 {
				return errors.New("singleton: max attempts must be greater than zero")
			}

			c.retry.MaxAttempts = n

			return nil
		},
	}
}

// WithInitializationTimeout limits the complete shared initialization, covering
// every attempt and the waits between them. The default is 30 seconds, and a
// zero duration disables the timeout.
//
// Exceeding it reports [FailureTimedOut]. [New] returns an error if timeout is
// negative.
func WithInitializationTimeout(timeout time.Duration) Option {
	return Option{
		apply: func(c *config) error {
			if timeout < 0 {
				return errors.New("singleton: initialization timeout cannot be negative")
			}

			c.retry.Timeout = timeout

			return nil
		},
	}
}

// WithRetryInterval sets the first delay between attempts and the ceiling that
// delay grows toward. Delays grow exponentially from initial, are capped at
// maximum, and carry jitter. The defaults are 250ms and 5s.
//
// [New] reports an error if initial is not positive, or if maximum is less than
// initial.
func WithRetryInterval(initial, maximum time.Duration) Option {
	return Option{
		apply: func(c *config) error {
			if initial <= 0 {
				return errors.New("singleton: initial retry interval must be greater than zero")
			}

			if maximum < initial {
				return errors.New("singleton: maximum retry interval cannot be less than the initial interval")
			}

			c.retry.InitialInterval = initial
			c.retry.MaxInterval = maximum

			return nil
		},
	}
}

// WithRetryObserver registers a function called once per retried attempt.
//
// A run that succeeds on its third attempt delivers two events; the final
// attempt delivers none, because its outcome is the result of Get.
//
// The observer should return quickly. A panic in it is recovered and discarded,
// so instrumentation cannot become the singleton's result. Passing nil disables
// it.
func WithRetryObserver(observer func(RetryEvent)) Option {
	return Option{
		apply: func(c *config) error {
			c.retry.Observer = observer

			return nil
		},
	}
}
