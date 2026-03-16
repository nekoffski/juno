package bus

import (
	"context"
	"fmt"
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
}

func New() *MessageBus {
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
