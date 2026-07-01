package queue_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"backend/internal/platform/queue"
)

func reset() {
	queue.ResetMemoryBroker()
}

func TestMemoryPublishSubscribe(t *testing.T) {
	reset()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p := queue.NewMemoryProducer()
	c := queue.NewMemoryConsumer()

	var mu sync.Mutex
	var received []string

	go func() {
		_ = c.Subscribe(ctx, "test-topic", func(msg queue.Message) error {
			mu.Lock()
			received = append(received, string(msg.Body))
			mu.Unlock()
			return nil
		})
	}()

	// let subscriber start
	<-time.After(50 * time.Millisecond)

	if err := p.Publish(ctx, "test-topic", queue.Message{ID: "1", Body: []byte("hello")}); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	<-time.After(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 1 || received[0] != "hello" {
		t.Fatalf("received = %v, want [hello]", received)
	}
}

func TestMemoryMultipleSubscribers(t *testing.T) {
	reset()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p := queue.NewMemoryProducer()

	var wg sync.WaitGroup
	var mu sync.Mutex
	counts := make(map[string]int)

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			c := queue.NewMemoryConsumer()
			_ = c.Subscribe(ctx, "test-topic", func(msg queue.Message) error {
				mu.Lock()
				counts[name]++
				mu.Unlock()
				return nil
			})
		}(string(rune('A' + i)))
	}

	// let consumers start
	<-time.After(200 * time.Millisecond)

	if err := p.Publish(ctx, "test-topic", queue.Message{ID: "1", Body: []byte("broadcast")}); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	cancel()
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	for name, count := range counts {
		if count != 1 {
			t.Fatalf("subscriber %s received %d messages, want 1", name, count)
		}
	}
}

func TestMemoryHandlerError(t *testing.T) {
	reset()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p := queue.NewMemoryProducer()
	c := queue.NewMemoryConsumer()

	callCount := 0
	go func() {
		_ = c.Subscribe(ctx, "test-topic", func(msg queue.Message) error {
			callCount++
			return errors.New("handler error")
		})
	}()

	<-time.After(50 * time.Millisecond)

	if err := p.Publish(ctx, "test-topic", queue.Message{ID: "1", Body: []byte("err")}); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	cancel()
	<-time.After(100 * time.Millisecond)
	if callCount != 1 {
		t.Fatalf("handler called %d times, want 1", callCount)
	}
}
