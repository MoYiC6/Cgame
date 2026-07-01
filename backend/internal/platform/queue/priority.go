package queue

import (
	"context"
	"encoding/json"

	"backend/internal/platform/redis"
	goredis "github.com/redis/go-redis/v9"
)

type PriorityQueue struct {
	rdb      *redis.Client
	producer Producer
}

func NewPriorityQueue(rdb *redis.Client, p Producer) *PriorityQueue {
	return &PriorityQueue{rdb: rdb, producer: p}
}

func (p *PriorityQueue) Push(ctx context.Context, topic string, msg Message, priority int64) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	pipe := p.rdb.Pool().Pipeline()
	pipe.ZAdd(ctx, priorityKey(topic), goredis.Z{
		Score:  float64(priority),
		Member: msg.ID,
	})
	pipe.HSet(ctx, msgKey(msg.ID), "data", data)
	_, err = pipe.Exec(ctx)
	return err
}

func (p *PriorityQueue) Pop(ctx context.Context, topic string) (Message, error) {
	ids, err := p.rdb.Pool().ZRevRange(ctx, priorityKey(topic), 0, 0).Result()
	if err != nil {
		return Message{}, err
	}
	if len(ids) == 0 {
		return Message{}, nil
	}

	id := ids[0]
	data, err := p.rdb.Pool().HGet(ctx, msgKey(id), "data").Result()
	if err != nil {
		return Message{}, err
	}

	var msg Message
	if err := json.Unmarshal([]byte(data), &msg); err != nil {
		return Message{}, err
	}

	p.rdb.Pool().ZRem(ctx, priorityKey(topic), id)
	p.rdb.Pool().HDel(ctx, msgKey(id), "data")
	return msg, nil
}

func priorityKey(topic string) string {
	return "priority:" + topic
}
