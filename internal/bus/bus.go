package bus

import (
	"context"
	"fmt"

	"github.com/nekoffski/juno/internal/core"
)

const queueCapacity = 64

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
	queues *queues
	topics map[string]*Topic
}

func New() *MessageBus {
	return &MessageBus{
		queues: &queues{},
		topics: make(map[string]*Topic),
	}
}

func (mb *MessageBus) NewSender() *Sender {
	return &Sender{queues: mb.queues}
}

func (mb *MessageBus) NewPublisher() *Publisher {
	return &Publisher{mb: mb}
}

func (mb *MessageBus) NewSubscriber() *Subscriber {
	return &Subscriber{
		mb:   mb,
		ch:   make(chan any, defaultSubscriberBuffer),
		done: make(chan struct{}),
	}
}

func (mb *MessageBus) RegisterReceiver(ctx context.Context, name string, handler func(Message)) error {
	if name == "" {
		return core.ErrEmptyName
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
