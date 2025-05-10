package inmemory

import (
	"context"
	"sync"

	"VK/internal/subpub"
)

type SubPub struct {
	Subscriptions map[string][]*Subscriber
	closed        chan struct{}
	mu            sync.Mutex
}

func NewSubPub() *SubPub {
	return &SubPub{
		Subscriptions: make(map[string][]*Subscriber),
		closed:        make(chan struct{}),
		mu:            sync.Mutex{},
	}
}

func (s *SubPub) Subscribe(subject string, cb subpub.MessageHandler) (subpub.Subscription, error) {
	select {
	case <-s.closed:
		return nil, subpub.ErrClosed
	default:
		if subject == "" {
			return nil, subpub.ErrEmptySubject
		}

		s.mu.Lock()
		defer s.mu.Unlock()

		sub := NewSubscriber(int64(len(s.Subscriptions[subject]))+1, subject, cb, s)
		sub.StartListening()
		s.Subscriptions[subject] = append(s.Subscriptions[subject], sub)

		return sub, nil
	}
}

func (s *SubPub) Publish(subject string, msg interface{}) error {
	select {
	case <-s.closed:
		return subpub.ErrClosed
	default:
		s.mu.Lock()
		defer s.mu.Unlock()

		subs, ok := s.Subscriptions[subject]
		if !ok || len(subs) == 0 {
			return subpub.ErrNoSubscribers
		}

		for _, sub := range subs {
			sub.MessageQueue <- msg
		}

		return nil
	}
}

func (s *SubPub) Close(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		s.mu.Lock()
		subs := s.Subscriptions
		s.Subscriptions = nil
		s.mu.Unlock()

		for _, list := range subs {
			for _, sub := range list {
				sub.Unsubscribe()
			}
		}

		close(s.closed)

		return nil
	}
}
