package queue

import (
	"context"
	"strconv"
)

type DLQConfig struct {
	MaxRetries int
	DLQTopic   string
}

type DLQConsumer struct {
	next   Consumer
	dlq    Producer
	config DLQConfig
}

func NewDLQConsumer(next Consumer, dlq Producer, cfg DLQConfig) *DLQConsumer {
	return &DLQConsumer{next: next, dlq: dlq, config: cfg}
}

func (d *DLQConsumer) Subscribe(ctx context.Context, topic string, handler func(Message) error) error {
	return d.next.Subscribe(ctx, topic, func(msg Message) error {
		err := handler(msg)
		if err == nil {
			return nil
		}

		retry := 0
		if v, ok := msg.Headers["retry"]; ok {
			retry, _ = strconv.Atoi(v)
		}
		retry++

		if retry >= d.config.MaxRetries {
			msg.Headers = ensureHeaders(msg.Headers)
			msg.Headers["dlq_reason"] = err.Error()
			msg.Headers["retry"] = strconv.Itoa(retry)
			_ = d.dlq.Publish(ctx, d.config.DLQTopic, msg)
			return nil
		}

		msg.Headers = ensureHeaders(msg.Headers)
		msg.Headers["retry"] = strconv.Itoa(retry)
		_ = d.dlq.Publish(ctx, topic, msg)
		return nil
	})
}

func ensureHeaders(headers map[string]string) map[string]string {
	if headers == nil {
		return make(map[string]string)
	}
	return headers
}
