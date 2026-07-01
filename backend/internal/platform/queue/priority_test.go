package queue_test

import (
	"context"
	"testing"

	"backend/internal/platform/queue"
)

func TestPriorityQueuePushPop(t *testing.T) {
	ctx := context.Background()
	rdb := setupRedis(t)

	streamProducer := queue.NewRedisStreamProducer(rdb)
	pq := queue.NewPriorityQueue(rdb, streamProducer)

	topic := "test-priority"

	messages := []struct {
		id       string
		body     string
		priority int64
	}{
		{"low", "low-priority", 1},
		{"high", "high-priority", 100},
		{"mid", "mid-priority", 50},
	}

	for _, m := range messages {
		if err := pq.Push(ctx, topic, queue.Message{ID: m.id, Body: []byte(m.body)}, m.priority); err != nil {
			t.Fatalf("Push() error = %v", err)
		}
	}

	expected := []string{"high-priority", "mid-priority", "low-priority"}
	for i, want := range expected {
		msg, err := pq.Pop(ctx, topic)
		if err != nil {
			t.Fatalf("Pop() error = %v", err)
		}
		if string(msg.Body) != want {
			t.Fatalf("Pop() = %q, want %q (step %d)", string(msg.Body), want, i)
		}
	}
}
