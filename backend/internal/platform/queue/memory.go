package queue

import (
	"context"
	"sync"
)

type memTopic struct {
	subs map[chan Message]struct{}
	mu   sync.RWMutex
}

type memBroker struct {
	topics map[string]*memTopic
	mu     sync.RWMutex
}

var globalBroker = &memBroker{topics: make(map[string]*memTopic)}

type memoryProducer struct{}

func NewMemoryProducer() Producer {
	return &memoryProducer{}
}

func (p *memoryProducer) Publish(ctx context.Context, topic string, msg Message) error {
	t := globalBroker.topic(topic)
	t.mu.RLock()
	defer t.mu.RUnlock()

	for ch := range t.subs {
		select {
		case ch <- msg:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (p *memoryProducer) Close() error { return nil }

type memoryConsumer struct{}

func NewMemoryConsumer() Consumer {
	return &memoryConsumer{}
}

func (c *memoryConsumer) Subscribe(ctx context.Context, topic string, handler func(Message) error) error {
	t := globalBroker.topic(topic)
	ch := make(chan Message, 64)

	t.mu.Lock()
	t.subs[ch] = struct{}{}
	t.mu.Unlock()

	defer func() {
		t.mu.Lock()
		delete(t.subs, ch)
		close(ch)
		t.mu.Unlock()
	}()

	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return nil
			}
			if err := handler(msg); err != nil {
				// 内存模式无重试，handler 自行处理
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (c *memoryConsumer) Close() error { return nil }

func ResetMemoryBroker() {
	globalBroker = &memBroker{topics: make(map[string]*memTopic)}
}

func (b *memBroker) topic(name string) *memTopic {
	b.mu.RLock()
	t, ok := b.topics[name]
	b.mu.RUnlock()
	if ok {
		return t
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	if t, ok = b.topics[name]; ok {
		return t
	}
	t = &memTopic{subs: make(map[chan Message]struct{})}
	b.topics[name] = t
	return t
}
