package queue_test

import (
	"context"
	"testing"
	"time"

	"backend/internal/platform/queue"
)

func TestDelayedQueueSchedule(t *testing.T) {
	ctx := context.Background()
	rdb := setupRedis(t)

	streamProducer := queue.NewRedisStreamProducer(rdb)
	delayed := queue.NewDelayedQueue(rdb, streamProducer, 200*time.Millisecond)

	topic := "test-delayed"
	msg := queue.Message{ID: "delayed-1", Body: []byte("delayed-hello")}

	if err := delayed.Schedule(ctx, topic, msg, 500*time.Millisecond); err != nil {
		t.Fatalf("Schedule() error = %v", err)
	}

	// Verify message is in delayed ZSet
	count, err := rdb.Pool().ZCard(ctx, "delayed:"+topic).Result()
	if err != nil {
		t.Fatalf("ZCard() error = %v", err)
	}
	if count != 1 {
		t.Fatalf("ZCard() = %d, want 1", count)
	}

	// Start processing and wait for delivery
	done := make(chan struct{})
	go func() {
		_ = delayed.Start(ctx, topic)
	}()

	streamConsumer := queue.NewRedisStreamConsumer(rdb, "delayed-group", "delayed-consumer")
	go func() {
		_ = streamConsumer.Subscribe(ctx, topic, func(msg queue.Message) error {
			if string(msg.Body) == "delayed-hello" {
				close(done)
			}
			return nil
		})
	}()

	select {
	case <-done:
		// message delivered
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for delayed message")
	}
}
