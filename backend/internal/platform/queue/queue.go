package queue

import (
	"context"
	"time"
)

type Message struct {
	ID        string
	Body      []byte
	Headers   map[string]string
	CreatedAt time.Time
}

type Producer interface {
	Publish(ctx context.Context, topic string, msg Message) error
	Close() error
}

type Consumer interface {
	Subscribe(ctx context.Context, topic string, handler func(Message) error) error
	Close() error
}
