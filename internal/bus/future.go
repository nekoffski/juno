package bus

import (
	"context"
	"fmt"
	"time"
)

type Future struct {
	ch <-chan Response
}

func (f *Future) Await(ctx context.Context) (any, error) {
	select {
	case r := <-f.ch:
		return r.Payload, r.Err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (f *Future) AwaitFor(ctx context.Context, d time.Duration) (any, error) {
	ctx, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	return f.Await(ctx)
}

func Await[T any](ctx context.Context, f *Future) (T, error) {
	var zero T
	payload, err := f.Await(ctx)
	if err != nil {
		return zero, err
	}
	v, ok := payload.(T)
	if !ok {
		return zero, fmt.Errorf("unexpected payload type: expected %T, got %T", zero, payload)
	}
	return v, nil
}

func AwaitFor[T any](ctx context.Context, f *Future, d time.Duration) (T, error) {
	ctx, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	return Await[T](ctx, f)
}
