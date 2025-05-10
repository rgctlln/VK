package inmemory

import "VK/internal/subpub"

const maxMessages = 1024

type Subscriber struct {
	ID           int64
	Subject      string
	Handler      subpub.MessageHandler
	MessageQueue chan interface{}
	closed       chan struct{}
	pub          *SubPub
}

func NewSubscriber(id int64, subject string, cb subpub.MessageHandler, pub *SubPub) *Subscriber {
	return &Subscriber{
		ID:           id,
		Subject:      subject,
		Handler:      cb,
		MessageQueue: make(chan interface{}, maxMessages),
		closed:       make(chan struct{}),
		pub:          pub,
	}
}

func (s *Subscriber) StartListening() {
	go func() {
		for {
			select {
			case <-s.closed:
				return

			case msg, ok := <-s.MessageQueue:
				if !ok {
					return
				}
				s.Handler(msg)
			}
		}
	}()
}

func (s *Subscriber) StopListening() {
	close(s.closed)
	close(s.MessageQueue)
}

func (s *Subscriber) Unsubscribe() {
	s.pub.mu.Lock()
	defer s.pub.mu.Unlock()

	subs, ok := s.pub.Subscriptions[s.Subject]
	if !ok {
		return
	}

	newSubs := make([]*Subscriber, 0, len(subs))
	for _, candidate := range subs {
		if candidate != s {
			newSubs = append(newSubs, candidate)
		}
	}

	if len(newSubs) == 0 {
		delete(s.pub.Subscriptions, s.Subject)
	} else {
		s.pub.Subscriptions[s.Subject] = newSubs
	}

	s.StopListening()
}
