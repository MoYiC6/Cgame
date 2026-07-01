package queue

import (
	"context"
	"fmt"
	"time"

	"backend/internal/platform/redis"
	goredis "github.com/redis/go-redis/v9"
)

type redisStreamProducer struct {
	rdb *redis.Client
}

func NewRedisStreamProducer(rdb *redis.Client) Producer {
	return &redisStreamProducer{rdb: rdb}
}

func (p *redisStreamProducer) Publish(ctx context.Context, topic string, msg Message) error {
	_, err := p.rdb.Pool().XAdd(ctx, &goredis.XAddArgs{
		Stream: topic,
		Values: map[string]interface{}{
			"id":    msg.ID,
			"body":  string(msg.Body),
			"ts":    msg.CreatedAt.UnixMilli(),
		},
	}).Result()
	return err
}

func (p *redisStreamProducer) Close() error {
	return p.rdb.Close()
}

type redisStreamConsumer struct {
	rdb       *redis.Client
	group     string
	consumer  string
	readCount int64
}

func NewRedisStreamConsumer(rdb *redis.Client, group, consumer string) Consumer {
	return &redisStreamConsumer{
		rdb:      rdb,
		group:    group,
		consumer: consumer,
	}
}

func (c *redisStreamConsumer) Subscribe(ctx context.Context, topic string, handler func(Message) error) error {
	if err := c.ensureGroup(ctx, topic); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		streams, err := c.rdb.Pool().XReadGroup(ctx, &goredis.XReadGroupArgs{
			Group:    c.group,
			Consumer: c.consumer,
			Streams:  []string{topic, ">"},
			Count:    c.readCount,
			Block:    5 * time.Second,
		}).Result()

		if err != nil {
			if err == goredis.Nil {
				continue
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			time.Sleep(500 * time.Millisecond)
			continue
		}

		for _, stream := range streams {
			for _, xmsg := range stream.Messages {
				msg := toMessage(xmsg)
				if err := handler(msg); err != nil {
					// 不 ACK，让消息留在 PEL 中，供下次重试
					continue
				}
				c.rdb.Pool().XAck(ctx, topic, c.group, xmsg.ID)
			}
		}
	}
}

func (c *redisStreamConsumer) ensureGroup(ctx context.Context, topic string) error {
	err := c.rdb.Pool().XGroupCreateMkStream(ctx, topic, c.group, "0").Err()
	if err == nil || err == goredis.Nil {
		return nil
	}
	if err.Error() == "BUSYGROUP Consumer Group already exists" {
		return nil
	}
	return err
}

func (c *redisStreamConsumer) Close() error {
	return c.rdb.Close()
}

func toMessage(xmsg goredis.XMessage) Message {
	body := []byte(xmsg.Values["body"].(string))
	ts := time.Now()
	if v, ok := xmsg.Values["ts"].(string); ok {
		if ms, err := parseInt64(v); err == nil {
			ts = time.UnixMilli(ms)
		}
	}
	return Message{
		ID:        xmsg.ID,
		Body:      body,
		Headers:   nil,
		CreatedAt: ts,
	}
}

func parseInt64(s string) (int64, error) {
	var n int64
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}
