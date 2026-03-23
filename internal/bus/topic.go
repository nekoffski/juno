package bus

import (
	"fmt"
	"sync"
)

const defaultSubscriberBuffer = 8

type Topic struct {
	mu   sync.Mutex
	subs map[int]chan any
	next int
}

func newTopic() *Topic {
	return &Topic{subs: make(map[int]chan any)}
}

func (t *Topic) publish(v any) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, ch := range t.subs {
		select {
		case ch <- v:
		default:
		}
	}
}

func (t *Topic) subscribe() (events <-chan any, unsubscribe func()) {
	t.mu.Lock()
	defer t.mu.Unlock()
	id := t.next
	t.next++
	ch := make(chan any, defaultSubscriberBuffer)
	t.subs[id] = ch
	return ch, func() {
		t.mu.Lock()
		defer t.mu.Unlock()
		delete(t.subs, id)
		close(ch)
	}
}

type Publisher struct {
	mb *MessageBus
}

func (p *Publisher) Publish(topic string, event any) error {
	t, exists := p.mb.topics[topic]
	if !exists {
		return fmt.Errorf("topic %q not found", topic)
	}
	t.publish(event)
	return nil
}

func RegisterTopic(mb *MessageBus, name string) error {
	if _, exists := mb.topics[name]; exists {
		return fmt.Errorf("topic %q is already registered", name)
	}
	mb.topics[name] = newTopic()
	return nil
}

type Subscriber struct {
	mb     *MessageBus
	ch     chan any
	done   chan struct{}
	unsubs []func()
	wg     sync.WaitGroup
}

func (s *Subscriber) Subscribe(topic string) error {
	t, exists := s.mb.topics[topic]
	if !exists {
		return fmt.Errorf("topic %q not found", topic)
	}
	topicCh, unsub := t.subscribe()
	s.unsubs = append(s.unsubs, unsub)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for ev := range topicCh {
			select {
			case s.ch <- ev:
			case <-s.done:
				return
			}
		}
	}()
	return nil
}

func (s *Subscriber) Events() <-chan any {
	return s.ch
}

func (s *Subscriber) Close() {
	close(s.done)
	for _, unsub := range s.unsubs {
		unsub()
	}
	s.wg.Wait()
	close(s.ch)
}
