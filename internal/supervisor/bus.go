package supervisor

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const queueCapacity = 64

var (
	ErrBusNotFound = errors.New("bus not found")
	ErrQueueFull   = errors.New("queue full")
	ErrEmptyName   = errors.New("bus name must not be empty")
)

type Message struct {
	Payload any
	replyTo chan Response
}

func (m Message) Reply(r Response) {
	if m.replyTo != nil {
		m.replyTo <- r
	}
}

type Response struct {
	Payload any
	Err     error
}

type Future struct {
	ch <-chan Response
}

func (f *Future) Await(ctx context.Context) (Response, error) {
	select {
	case r := <-f.ch:
		return r, r.Err
	case <-ctx.Done():
		return Response{}, ctx.Err()
	}
}

func (f *Future) AwaitFor(ctx context.Context, d time.Duration) (Response, error) {
	ctx, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	return f.Await(ctx)
}

type queue struct {
	ch chan Message
}

type queues map[string]*queue

func getQueue(name string, q *queues) *queue {
	if _, exists := (*q)[name]; !exists {
		return nil
	}
	return (*q)[name]
}

type Sender struct {
	queues *queues
}

func (s *Sender) Send(to string, payload any) error {
	if to == "" {
		return ErrEmptyName
	}
	q := getQueue(to, s.queues)
	if q == nil {
		return fmt.Errorf("%w: %q", ErrBusNotFound, to)
	}
	select {
	case q.ch <- Message{Payload: payload}:
		return nil
	default:
		return fmt.Errorf("%w: %q", ErrQueueFull, to)
	}
}

func (s *Sender) Request(to string, payload any) (*Future, error) {
	if to == "" {
		return nil, ErrEmptyName
	}
	q := getQueue(to, s.queues)
	if q == nil {
		return nil, fmt.Errorf("%w: %q", ErrBusNotFound, to)
	}

	replyCh := make(chan Response, 1)

	select {
	case q.ch <- Message{Payload: payload, replyTo: replyCh}:
	default:
		return nil, fmt.Errorf("%w: %q", ErrQueueFull, to)
	}

	return &Future{ch: replyCh}, nil
}

type MessageBus struct {
	queues *queues
}

func NewMessageBus() *MessageBus {
	return &MessageBus{
		queues: &queues{},
	}
}

func (mb *MessageBus) NewSender() *Sender {
	return &Sender{queues: mb.queues}
}

func (mb *MessageBus) RegisterReceiver(ctx context.Context, name string, handler func(Message)) error {
	if name == "" {
		return ErrEmptyName
	}
	if _, exists := (*mb.queues)[name]; exists {
		return fmt.Errorf("receiver %q is already registered", name)
	}
	q := &queue{ch: make(chan Message, queueCapacity)}
	(*mb.queues)[name] = q

	go func() {
		for {
			select {
			case msg := <-q.ch:
				handler(msg)
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}
