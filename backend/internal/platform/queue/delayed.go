package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"backend/internal/platform/redis"
	goredis "github.com/redis/go-redis/v9"
)

type DelayedQueue struct {
	rdb          *redis.Client
	producer     Producer
	pollInterval time.Duration
}

func NewDelayedQueue(rdb *redis.Client, p Producer, poll time.Duration) *DelayedQueue {
	if poll == 0 {
		poll = time.Second
	}
	return &DelayedQueue{rdb: rdb, producer: p, pollInterval: poll}
}

func (d *DelayedQueue) Schedule(ctx context.Context, topic string, msg Message, delay time.Duration) error {
	msg.CreatedAt = time.Now()
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	pipe := d.rdb.Pool().Pipeline()
	pipe.ZAdd(ctx, delayedKey(topic), goredis.Z{
		Score:  float64(time.Now().Add(delay).UnixMilli()),
		Member: msg.ID,
	})
	pipe.HSet(ctx, msgKey(msg.ID), "data", data)
	_, err = pipe.Exec(ctx)
	return err
}

func (d *DelayedQueue) Start(ctx context.Context, topic string) error {
	ticker := time.NewTicker(d.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			d.processExpired(ctx, topic)
		}
	}
}

func (d *DelayedQueue) processExpired(ctx context.Context, topic string) {
	now := time.Now().UnixMilli()
	ids, err := d.rdb.Pool().ZRangeByScore(ctx, delayedKey(topic), &goredis.ZRangeBy{
		Min: "-inf",
		Max: fmt.Sprintf("%d", now),
	}).Result()

	if err != nil || len(ids) == 0 {
		return
	}

	for _, id := range ids {
		data, err := d.rdb.Pool().HGet(ctx, msgKey(id), "data").Result()
		if err != nil {
			continue
		}

		var msg Message
		if err := json.Unmarshal([]byte(data), &msg); err != nil {
			continue
		}

		if err := d.producer.Publish(ctx, topic, msg); err != nil {
			continue
		}

		d.rdb.Pool().ZRem(ctx, delayedKey(topic), id)
		d.rdb.Pool().HDel(ctx, msgKey(id), "data")
	}
}

func delayedKey(topic string) string {
	return "delayed:" + topic
}

func msgKey(id string) string {
	return "msg:" + id
}
