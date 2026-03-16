package bus

import "fmt"

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
