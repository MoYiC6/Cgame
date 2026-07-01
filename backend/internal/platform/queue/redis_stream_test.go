package queue_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"backend/internal/platform/config"
	"backend/internal/platform/queue"
	"backend/internal/platform/redis"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupRedis(t *testing.T) *redis.Client {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("testcontainers: %v", err)
	}
	t.Cleanup(func() { container.Terminate(ctx) })

	port, err := container.MappedPort(ctx, "6379")
	if err != nil {
		t.Fatalf("MappedPort: %v", err)
	}

	cfg := config.RedisConfig{Addr: "localhost:" + port.Port()}
	rdb, err := redis.New(cfg)
	if err != nil {
		t.Fatalf("redis.New: %v", err)
	}
	t.Cleanup(func() { rdb.Close() })
	return rdb
}

func TestRedisStreamPublishSubscribe(t *testing.T) {
	ctx := context.Background()
	rdb := setupRedis(t)

	p := queue.NewRedisStreamProducer(rdb)
	c := queue.NewRedisStreamConsumer(rdb, "test-group", "consumer-1")

	var received []string
	done := make(chan struct{})

	go func() {
		_ = c.Subscribe(ctx, "test-stream", func(msg queue.Message) error {
			received = append(received, string(msg.Body))
			close(done)
			return nil
		})
	}()

	// let consumer group start
	<-time.After(200 * time.Millisecond)

	if err := p.Publish(ctx, "test-stream", queue.Message{ID: "msg-1", Body: []byte("stream-hello")}); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	select {
	case <-done:
		// message received
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for message")
	}

	if len(received) != 1 || received[0] != "stream-hello" {
		t.Fatalf("received = %v, want [stream-hello]", received)
	}
}

func TestRedisStreamMultipleMessages(t *testing.T) {
	ctx := context.Background()
	rdb := setupRedis(t)

	p := queue.NewRedisStreamProducer(rdb)
	c := queue.NewRedisStreamConsumer(rdb, "test-group-2", "consumer-2")

	var mu sync.Mutex
	var received []string
	done := make(chan struct{})

	go func() {
		_ = c.Subscribe(ctx, "test-stream-2", func(msg queue.Message) error {
			mu.Lock()
			received = append(received, string(msg.Body))
			if len(received) == 3 {
				close(done)
			}
			mu.Unlock()
			return nil
		})
	}()

	<-time.After(200 * time.Millisecond)

	for i := 0; i < 3; i++ {
		body := []byte("msg-" + string(rune('0'+i)))
		if err := p.Publish(ctx, "test-stream-2", queue.Message{ID: string(rune('m' + i)), Body: body}); err != nil {
			t.Fatalf("Publish() error = %v", err)
		}
	}

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for messages")
	}

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 3 {
		t.Fatalf("received %d messages, want 3", len(received))
	}
}
