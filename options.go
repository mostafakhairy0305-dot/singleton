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

func WithRetryObserver(observer func(RetryEvent)) Option {
	return Option{
		apply: func(c *config) error {
			c.retry.Observer = observer

			return nil
		},
	}
}
