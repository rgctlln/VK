package subpub

import (
	"context"
	"errors"
)

var (
	ErrEmptySubject  = errors.New("empty subject")
	ErrNoSubscribers = errors.New("no subscribers on such subject")
	ErrClosed        = errors.New("closed")
)

type MessageHandler func(msg interface{})

type Subscription interface {
	Unsubscribe()
}

type SubPub interface {
	Subscribe(subject string, cb MessageHandler) (Subscription, error)

	Publish(subject string, msg interface{}) error

	Close(ctx context.Context) error
}
