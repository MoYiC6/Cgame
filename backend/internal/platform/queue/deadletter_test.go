package queue_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"backend/internal/platform/queue"
)

func TestDLQRetriesThenDeadLetter(t *testing.T) {
	reset()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dlqProducer := queue.NewMemoryProducer()
	next := queue.NewMemoryConsumer()
	dlq := queue.NewDLQConsumer(next, dlqProducer, queue.DLQConfig{
		MaxRetries: 3,
		DLQTopic:   "dlq:orders",
	})

	var dlqReceived []queue.Message
	dlqDone := make(chan struct{})

	// Start DLQ consumer BEFORE publishing
	dlqConsumer := queue.NewMemoryConsumer()
	go func() {
		_ = dlqConsumer.Subscribe(ctx, "dlq:orders", func(msg queue.Message) error {
			if msg.ID == "order-1" {
				dlqReceived = append(dlqReceived, msg)
				close(dlqDone)
			}
			return nil
		})
	}()

	go func() {
		_ = dlq.Subscribe(ctx, "orders", func(msg queue.Message) error {
			return errors.New("always fail")
		})
	}()

	p := queue.NewMemoryProducer()

	// let consumers start
	<-time.After(100 * time.Millisecond)

	// Publish a message that will fail
	if err := p.Publish(ctx, "orders", queue.Message{ID: "order-1", Body: []byte("fail-me")}); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	select {
	case <-dlqDone:
		if len(dlqReceived) != 1 {
			t.Fatalf("DLQ received %d messages, want 1", len(dlqReceived))
		}
		if dlqReceived[0].Headers["dlq_reason"] == "" {
			t.Fatal("DLQ message missing dlq_reason header")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for DLQ message")
	}
}
