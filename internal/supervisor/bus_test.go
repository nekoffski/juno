package supervisor

import (
	"context"
	"errors"
	"testing"
	"time"
)

type pingPayload struct{ Text string }

func newSetup(t *testing.T) (*MessageBus, *Sender) {
	t.Helper()
	bus := NewMessageBus()
	return bus, bus.NewSender()
}

func registerReceiver(t *testing.T, bus *MessageBus, name string, handler func(Message)) context.CancelFunc {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	if err := bus.RegisterReceiver(ctx, name, handler); err != nil {
		cancel()
		t.Fatalf("RegisterReceiver(%q): %v", name, err)
	}
	return cancel
}

func TestRegisterReceiver_OK(t *testing.T) {
	bus, _ := newSetup(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := bus.RegisterReceiver(ctx, "svc", func(Message) {}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegisterReceiver_EmptyName(t *testing.T) {
	bus, _ := newSetup(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := bus.RegisterReceiver(ctx, "", func(Message) {})
	if !errors.Is(err, ErrEmptyName) {
		t.Fatalf("expected ErrEmptyName, got %v", err)
	}
}

func TestRegisterReceiver_Duplicate(t *testing.T) {
	bus, _ := newSetup(t)
	cancel := registerReceiver(t, bus, "svc", func(Message) {})
	defer cancel()
	ctx, cancel2 := context.WithCancel(context.Background())
	defer cancel2()
	err := bus.RegisterReceiver(ctx, "svc", func(Message) {})
	if err == nil {
		t.Fatal("expected error on duplicate registration, got nil")
	}
}

func TestSend_OK(t *testing.T) {
	bus, sender := newSetup(t)

	received := make(chan Message, 1)
	cancel := registerReceiver(t, bus, "b", func(msg Message) {
		received <- msg
	})
	defer cancel()

	if err := sender.Send("b", pingPayload{Text: "hello"}); err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	select {
	case msg := <-received:
		p, ok := msg.Payload.(pingPayload)
		if !ok {
			t.Fatalf("expected pingPayload, got %T", msg.Payload)
		}
		if p.Text != "hello" {
			t.Errorf("expected Text %q, got %q", "hello", p.Text)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for message")
	}
}

func TestSend_UnknownBus(t *testing.T) {
	_, sender := newSetup(t)
	err := sender.Send("nonexistent", nil)
	if !errors.Is(err, ErrBusNotFound) {
		t.Fatalf("expected ErrBusNotFound, got %v", err)
	}
}

func TestSend_EmptyName(t *testing.T) {
	_, sender := newSetup(t)
	err := sender.Send("", nil)
	if !errors.Is(err, ErrEmptyName) {
		t.Fatalf("expected ErrEmptyName, got %v", err)
	}
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

	if err := sender.Send("b", pingPayload{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	<-entered

	for i := 0; i < queueCapacity; i++ {
		if err := sender.Send("b", pingPayload{}); err != nil {
			t.Fatalf("unexpected error on send %d: %v", i, err)
		}
	}
	if err := sender.Send("b", pingPayload{}); !errors.Is(err, ErrQueueFull) {
		t.Fatalf("expected ErrQueueFull, got %v", err)
	}
}

func TestRequest_Reply(t *testing.T) {
	bus, sender := newSetup(t)

	cancel := registerReceiver(t, bus, "b", func(msg Message) {
		msg.Reply(Response{Payload: "pong"})
	})
	defer cancel()

	future, err := sender.Request("b", pingPayload{})
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	ctx, cancel2 := context.WithTimeout(context.Background(), time.Second)
	defer cancel2()

	resp, err := future.Await(ctx)
	if err != nil {
		t.Fatalf("Await failed: %v", err)
	}
	if resp.Payload != "pong" {
		t.Errorf("expected payload %q, got %q", "pong", resp.Payload)
	}
}

func TestRequest_UnknownBus(t *testing.T) {
	_, sender := newSetup(t)
	_, err := sender.Request("nonexistent", nil)
	if !errors.Is(err, ErrBusNotFound) {
		t.Fatalf("expected ErrBusNotFound, got %v", err)
	}
}

func TestRequest_EmptyName(t *testing.T) {
	_, sender := newSetup(t)
	_, err := sender.Request("", nil)
	if !errors.Is(err, ErrEmptyName) {
		t.Fatalf("expected ErrEmptyName, got %v", err)
	}
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

	if err := sender.Send("b", pingPayload{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	<-entered

	for i := 0; i < queueCapacity; i++ {
		if err := sender.Send("b", pingPayload{}); err != nil {
			t.Fatalf("unexpected error on send %d: %v", i, err)
		}
	}
	if _, err := sender.Request("b", pingPayload{}); !errors.Is(err, ErrQueueFull) {
		t.Fatalf("expected ErrQueueFull, got %v", err)
	}
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
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	ctx, cancelAwait := context.WithCancel(context.Background())
	cancelAwait()

	_, err = future.Await(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestFutureAwaitFor_OK(t *testing.T) {
	bus, sender := newSetup(t)

	cancel := registerReceiver(t, bus, "b", func(msg Message) {
		msg.Reply(Response{Payload: "pong"})
	})
	defer cancel()

	future, err := sender.Request("b", pingPayload{})
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	ctx, cancel2 := context.WithCancel(context.Background())
	defer cancel2()

	resp, err := future.AwaitFor(ctx, time.Second)
	if err != nil {
		t.Fatalf("AwaitFor failed: %v", err)
	}
	if resp.Payload != "pong" {
		t.Errorf("expected payload %q, got %q", "pong", resp.Payload)
	}
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
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	_, err = future.AwaitFor(context.Background(), 10*time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context.DeadlineExceeded, got %v", err)
	}
}

func TestReply_FireAndForget_NoOp(t *testing.T) {
	msg := Message{}
	msg.Reply(Response{Payload: "ignored"})
}

func TestReceiver_ContextCancel(t *testing.T) {
	bus, sender := newSetup(t)

	handled := make(chan struct{}, 1)
	ctx, cancel := context.WithCancel(context.Background())
	if err := bus.RegisterReceiver(ctx, "b", func(msg Message) {
		handled <- struct{}{}
	}); err != nil {
		t.Fatalf("RegisterReceiver failed: %v", err)
	}

	cancel()
	time.Sleep(10 * time.Millisecond)

	if err := sender.Send("b", pingPayload{}); err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	select {
	case <-handled:
		t.Fatal("handler should not be called after context cancel")
	case <-time.After(50 * time.Millisecond):
	}
}
