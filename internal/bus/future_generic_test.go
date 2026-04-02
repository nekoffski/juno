package bus

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeFutureWith(payload any, err error) *Future {
	ch := make(chan Response, 1)
	ch <- Response{Payload: payload, Err: err}
	return &Future{ch: ch}
}

func TestAwaitGeneric_OK(t *testing.T) {
	f := makeFutureWith("hello", nil)
	v, err := Await[string](context.Background(), f)
	require.NoError(t, err)
	assert.Equal(t, "hello", v)
}

func TestAwaitGeneric_WrongType(t *testing.T) {
	f := makeFutureWith(42, nil)
	_, err := Await[string](context.Background(), f)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected payload type")
}

func TestAwaitGeneric_PropagatesError(t *testing.T) {
	sentinel := errors.New("downstream failure")
	f := makeFutureWith(nil, sentinel)
	_, err := Await[string](context.Background(), f)
	assert.ErrorIs(t, err, sentinel)
}

func TestAwaitGeneric_ContextCancelled(t *testing.T) {
	ch := make(chan Response)
	f := &Future{ch: ch}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Await[string](ctx, f)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestAwaitForGeneric_OK(t *testing.T) {
	f := makeFutureWith(99, nil)
	v, err := AwaitFor[int](context.Background(), f, time.Second)
	require.NoError(t, err)
	assert.Equal(t, 99, v)
}

func TestAwaitForGeneric_Timeout(t *testing.T) {
	ch := make(chan Response)
	f := &Future{ch: ch}
	_, err := AwaitFor[int](context.Background(), f, 10*time.Millisecond)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}
