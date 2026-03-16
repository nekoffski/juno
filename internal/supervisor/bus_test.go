package supervisor

import (
	"context"
	"errors"
	"testing"
	"time"
)

func setup(t *testing.T, names ...string) (*MessageBusManager, map[string]*MessageBus) {
	t.Helper()
	mbm := NewMessageBusManager()
	buses := make(map[string]*MessageBus, len(names))
	for _, name := range names {
		bus, err := mbm.RegisterBus(name)
		if err != nil {
			t.Fatalf("RegisterBus(%q): %v", name, err)
		}
		buses[name] = bus
	}
	return mbm, buses
}

func TestRegisterBus_OK(t *testing.T) {
	mbm := NewMessageBusManager()
	bus, err := mbm.RegisterBus("svc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bus == nil {
		t.Fatal("expected non-nil MessageBus")
	}
}

func TestRegisterBus_EmptyName(t *testing.T) {
	mbm := NewMessageBusManager()
	_, err := mbm.RegisterBus("")
	if !errors.Is(err, ErrEmptyName) {
		t.Fatalf("expected ErrEmptyName, got %v", err)
	}
}

func TestRegisterBus_Duplicate(t *testing.T) {
	mbm := NewMessageBusManager()
	if _, err := mbm.RegisterBus("svc"); err != nil {
		t.Fatalf("first registration failed: %v", err)
	}
	_, err := mbm.RegisterBus("svc")
	if err == nil {
		t.Fatal("expected error on duplicate registration, got nil")
	}
}

func TestSend_OK(t *testing.T) {
	_, buses := setup(t, "a", "b")

	msg := Message{ID: "1", Type: "ping", Payload: "hello"}
	if err := buses["a"].Send("b", msg); err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	select {
	case received := <-buses["b"].Receive():
		if received.ID != msg.ID {
			t.Errorf("expected ID %q, got %q", msg.ID, received.ID)
		}
		if received.Type != msg.Type {
			t.Errorf("expected Type %q, got %q", msg.Type, received.Type)
		}
		if received.Payload != msg.Payload {
			t.Errorf("expected Payload %q, got %q", msg.Payload, received.Payload)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for message")
	}
}

func TestSend_UnknownBus(t *testing.T) {
	_, buses := setup(t, "a")

	err := buses["a"].Send("nonexistent", Message{})
	if !errors.Is(err, ErrBusNotFound) {
		t.Fatalf("expected ErrBusNotFound, got %v", err)
	}
}

func TestSend_EmptyName(t *testing.T) {
	_, buses := setup(t, "a")

	err := buses["a"].Send("", Message{})
	if !errors.Is(err, ErrEmptyName) {
		t.Fatalf("expected ErrEmptyName, got %v", err)
	}
}

func TestSend_QueueFull(t *testing.T) {
	_, buses := setup(t, "a", "b")

	msg := Message{Type: "ping"}
	for i := 0; i < queueCapacity; i++ {
		if err := buses["a"].Send("b", msg); err != nil {
			t.Fatalf("unexpected error on send %d: %v", i, err)
		}
	}
	err := buses["a"].Send("b", msg)
	if !errors.Is(err, ErrQueueFull) {
		t.Fatalf("expected ErrQueueFull, got %v", err)
	}
}

func TestRequest_Reply(t *testing.T) {
	_, buses := setup(t, "a", "b")

	go func() {
		msg := <-buses["b"].Receive()
		msg.Reply(Response{Payload: "pong"})
	}()

	future, err := buses["a"].Request("b", Message{ID: "1", Type: "ping"})
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := future.Await(ctx)
	if err != nil {
		t.Fatalf("Await failed: %v", err)
	}
	if resp.Payload != "pong" {
		t.Errorf("expected payload %q, got %q", "pong", resp.Payload)
	}
}

func TestRequest_UnknownBus(t *testing.T) {
	_, buses := setup(t, "a")

	_, err := buses["a"].Request("nonexistent", Message{})
	if !errors.Is(err, ErrBusNotFound) {
		t.Fatalf("expected ErrBusNotFound, got %v", err)
	}
}

func TestRequest_EmptyName(t *testing.T) {
	_, buses := setup(t, "a")

	_, err := buses["a"].Request("", Message{})
	if !errors.Is(err, ErrEmptyName) {
		t.Fatalf("expected ErrEmptyName, got %v", err)
	}
}

func TestFutureAwait_ContextCancelled(t *testing.T) {
	_, buses := setup(t, "a", "b")

	future, err := buses["a"].Request("b", Message{ID: "1", Type: "ping"})
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = future.Await(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestFutureAwaitFor_OK(t *testing.T) {
	_, buses := setup(t, "a", "b")

	go func() {
		msg := <-buses["b"].Receive()
		msg.Reply(Response{Payload: "pong"})
	}()

	future, err := buses["a"].Request("b", Message{ID: "1", Type: "ping"})
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	resp, err := future.AwaitFor(time.Second)
	if err != nil {
		t.Fatalf("AwaitFor failed: %v", err)
	}
	if resp.Payload != "pong" {
		t.Errorf("expected payload %q, got %q", "pong", resp.Payload)
	}
}

func TestFutureAwaitFor_Timeout(t *testing.T) {
	_, buses := setup(t, "a", "b")

	future, err := buses["a"].Request("b", Message{ID: "1", Type: "ping"})
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	_, err = future.AwaitFor(10 * time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context.DeadlineExceeded, got %v", err)
	}
}

func TestReply_FireAndForget_NoOp(t *testing.T) {
	msg := Message{ID: "1", Type: "event"}
	msg.Reply(Response{Payload: "ignored"})
}

func TestRequest_QueueFull(t *testing.T) {
	_, buses := setup(t, "a", "b")

	msg := Message{Type: "flood"}
	for i := 0; i < queueCapacity; i++ {
		if err := buses["a"].Send("b", msg); err != nil {
			t.Fatalf("unexpected error on send %d: %v", i, err)
		}
	}

	_, err := buses["a"].Request("b", Message{Type: "ping"})
	if !errors.Is(err, ErrQueueFull) {
		t.Fatalf("expected ErrQueueFull, got %v", err)
	}
}
