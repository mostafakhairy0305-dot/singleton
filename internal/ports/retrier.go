package ports

import "context"

type Operation[T any] func(context.Context) (T, error)

type Retrier[T any] interface {
	Do(ctx context.Context, op Operation[T]) (T, error)
}
