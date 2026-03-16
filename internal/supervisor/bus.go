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

type MessageType string

type Message struct {
	ID      string
	Type    MessageType
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

func (f *Future) AwaitFor(d time.Duration) (Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), d)
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

type MessageBus struct {
	name   string
	self   *queue
	queues *queues
}

func (mb *MessageBus) Send(to string, msg Message) error {
	if to == "" {
		return ErrEmptyName
	}
	q := getQueue(to, mb.queues)
	if q == nil {
		return fmt.Errorf("%w: %q", ErrBusNotFound, to)
	}
	select {
	case q.ch <- msg:
		return nil
	default:
		return fmt.Errorf("%w: %q", ErrQueueFull, to)
	}
}

func (mb *MessageBus) Receive() <-chan Message {
	return mb.self.ch
}

func (mb *MessageBus) Request(to string, msg Message) (*Future, error) {
	if to == "" {
		return nil, ErrEmptyName
	}
	q := getQueue(to, mb.queues)
	if q == nil {
		return nil, fmt.Errorf("%w: %q", ErrBusNotFound, to)
	}

	replyCh := make(chan Response, 1)
	msg.replyTo = replyCh

	select {
	case q.ch <- msg:
	default:
		return nil, fmt.Errorf("%w: %q", ErrQueueFull, to)
	}

	return &Future{ch: replyCh}, nil
}

type MessageBusManager struct {
	queues *queues
}

func NewMessageBusManager() *MessageBusManager {
	return &MessageBusManager{
		queues: &queues{},
	}
}

func (mbm *MessageBusManager) RegisterBus(name string) (*MessageBus, error) {
	if name == "" {
		return nil, ErrEmptyName
	}
	if _, exists := (*mbm.queues)[name]; exists {
		return nil, fmt.Errorf("bus %q is already registered", name)
	}
	q := &queue{ch: make(chan Message, queueCapacity)}
	(*mbm.queues)[name] = q
	return &MessageBus{
		name:   name,
		self:   q,
		queues: mbm.queues,
	}, nil
}
