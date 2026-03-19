package bus

import (
	"fmt"

	"github.com/nekoffski/juno/internal/core"
)

type Sender struct {
	queues *queues
}

func (s *Sender) Send(to string, payload any) error {
	if to == "" {
		return core.ErrEmptyName
	}
	q := getQueue(to, s.queues)
	if q == nil {
		return fmt.Errorf("%w: %q", core.ErrBusNotFound, to)
	}
	select {
	case q.ch <- Message{Payload: payload}:
		return nil
	default:
		return fmt.Errorf("%w: %q", core.ErrQueueFull, to)
	}
}

func (s *Sender) Request(to string, payload any) (*Future, error) {
	if to == "" {
		return nil, core.ErrEmptyName
	}
	q := getQueue(to, s.queues)
	if q == nil {
		return nil, fmt.Errorf("%w: %q", core.ErrBusNotFound, to)
	}

	replyCh := make(chan Response, 1)

	select {
	case q.ch <- Message{Payload: payload, replyTo: replyCh}:
	default:
		return nil, fmt.Errorf("%w: %q", core.ErrQueueFull, to)
	}

	return &Future{ch: replyCh}, nil
}
