package bus

import (
	"context"
	"testing"
	"time"

	"github.com/nekoffski/juno/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type pingPayload struct{ Text string }

func newSetup(t *testing.T) (*MessageBus, *Sender) {
	t.Helper()
	bus := New()
	return bus, bus.NewSender()
}

func registerReceiver(t *testing.T, bus *MessageBus, name string, handler func(Message)) context.CancelFunc {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	err := bus.RegisterReceiver(ctx, name, handler)
	require.NoError(t, err, "RegisterReceiver(%q)", name)
	return cancel
}

func TestRegisterReceiver_OK(t *testing.T) {
	bus, _ := newSetup(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := bus.RegisterReceiver(ctx, "svc", func(Message) {})
	require.NoError(t, err)
}

func TestRegisterReceiver_EmptyName(t *testing.T) {
	bus, _ := newSetup(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := bus.RegisterReceiver(ctx, "", func(Message) {})
	assert.ErrorIs(t, err, core.ErrEmptyName)
}

func TestRegisterReceiver_Duplicate(t *testing.T) {
	bus, _ := newSetup(t)
	cancel := registerReceiver(t, bus, "svc", func(Message) {})
	defer cancel()
	ctx, cancel2 := context.WithCancel(context.Background())
	defer cancel2()
	err := bus.RegisterReceiver(ctx, "svc", func(Message) {})
	assert.Error(t, err)
}

func TestSend_OK(t *testing.T) {
	bus, sender := newSetup(t)

	received := make(chan Message, 1)
	cancel := registerReceiver(t, bus, "b", func(msg Message) {
		received <- msg
	})
	defer cancel()

	require.NoError(t, sender.Send("b", pingPayload{Text: "hello"}))

	select {
	case msg := <-received:
		p, ok := msg.Payload.(pingPayload)
		require.True(t, ok, "expected pingPayload, got %T", msg.Payload)
		assert.Equal(t, "hello", p.Text)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for message")
	}
}

func TestSend_UnknownBus(t *testing.T) {
	_, sender := newSetup(t)
	assert.ErrorIs(t, sender.Send("nonexistent", nil), core.ErrBusNotFound)
}

func TestSend_EmptyName(t *testing.T) {
	_, sender := newSetup(t)
	assert.ErrorIs(t, sender.Send("", nil), core.ErrEmptyName)
}

func TestSend_QueueFull(t *testing.T) {
	bus, sender := newSetup(t)

	entered := make(chan struct{})
	block := make(chan struct{})

	cancel := registerReceiver(t, bus, "b", func(msg Message) {
		select {
		case entered <- struct{}{}:
		default:
		}
		<-block
	})
	defer func() {
		close(block)
		cancel()
	}()

	require.NoError(t, sender.Send("b", pingPayload{}))
	<-entered

	for i := 0; i < queueCapacity; i++ {
		require.NoError(t, sender.Send("b", pingPayload{}), "send %d", i)
	}
	assert.ErrorIs(t, sender.Send("b", pingPayload{}), core.ErrQueueFull)
}

func TestRequest_Reply(t *testing.T) {
	bus, sender := newSetup(t)

	cancel := registerReceiver(t, bus, "b", func(msg Message) {
		msg.Reply(Response{Payload: "pong"})
	})
	defer cancel()

	future, err := sender.Request("b", pingPayload{})
	require.NoError(t, err)

	ctx, cancel2 := context.WithTimeout(context.Background(), time.Second)
	defer cancel2()

	resp, err := future.Await(ctx)
	require.NoError(t, err)
	assert.Equal(t, "pong", resp)
}

func TestRequest_UnknownBus(t *testing.T) {
	_, sender := newSetup(t)
	_, err := sender.Request("nonexistent", nil)
	assert.ErrorIs(t, err, core.ErrBusNotFound)
}

func TestRequest_EmptyName(t *testing.T) {
	_, sender := newSetup(t)
	_, err := sender.Request("", nil)
	assert.ErrorIs(t, err, core.ErrEmptyName)
}

func TestRequest_QueueFull(t *testing.T) {
	bus, sender := newSetup(t)

	entered := make(chan struct{})
	block := make(chan struct{})

	cancel := registerReceiver(t, bus, "b", func(msg Message) {
		select {
		case entered <- struct{}{}:
		default:
		}
		<-block
	})
	defer func() {
		close(block)
		cancel()
	}()

	require.NoError(t, sender.Send("b", pingPayload{}))
	<-entered

	for i := 0; i < queueCapacity; i++ {
		require.NoError(t, sender.Send("b", pingPayload{}), "send %d", i)
	}
	_, err := sender.Request("b", pingPayload{})
	assert.ErrorIs(t, err, core.ErrQueueFull)
}

func TestFutureAwait_ContextCancelled(t *testing.T) {
	bus, sender := newSetup(t)

	block := make(chan struct{})
	cancel := registerReceiver(t, bus, "b", func(msg Message) { <-block })
	defer func() {
		close(block)
		cancel()
	}()

	future, err := sender.Request("b", pingPayload{})
	require.NoError(t, err)

	ctx, cancelAwait := context.WithCancel(context.Background())
	cancelAwait()

	_, err = future.Await(ctx)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestFutureAwaitFor_OK(t *testing.T) {
	bus, sender := newSetup(t)

	cancel := registerReceiver(t, bus, "b", func(msg Message) {
		msg.Reply(Response{Payload: "pong"})
	})
	defer cancel()

	future, err := sender.Request("b", pingPayload{})
	require.NoError(t, err)

	ctx, cancel2 := context.WithCancel(context.Background())
	defer cancel2()

	resp, err := future.AwaitFor(ctx, time.Second)
	require.NoError(t, err)
	assert.Equal(t, "pong", resp)
}

func TestFutureAwaitFor_Timeout(t *testing.T) {
	bus, sender := newSetup(t)

	block := make(chan struct{})
	cancel := registerReceiver(t, bus, "b", func(msg Message) { <-block })
	defer func() {
		close(block)
		cancel()
	}()

	future, err := sender.Request("b", pingPayload{})
	require.NoError(t, err)

	_, err = future.AwaitFor(context.Background(), 10*time.Millisecond)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestReply_FireAndForget_NoOp(t *testing.T) {
	msg := Message{}
	msg.Reply(Response{Payload: "ignored"})
}

func TestReceiver_ContextCancel(t *testing.T) {
	bus, sender := newSetup(t)

	handled := make(chan struct{}, 1)
	ctx, cancel := context.WithCancel(context.Background())
	require.NoError(t, bus.RegisterReceiver(ctx, "b", func(msg Message) {
		handled <- struct{}{}
	}))

	cancel()
	time.Sleep(10 * time.Millisecond)

	require.NoError(t, sender.Send("b", pingPayload{}))

	select {
	case <-handled:
		t.Fatal("handler should not be called after context cancel")
	case <-time.After(50 * time.Millisecond):
	}
}
