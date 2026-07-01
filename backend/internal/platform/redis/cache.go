package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type Cache[T any] struct {
	c          *Client
	defaultTTL time.Duration
}

func NewCache[T any](c *Client, defaultTTL time.Duration) *Cache[T] {
	return &Cache[T]{c: c, defaultTTL: defaultTTL}
}

func (c *Cache[T]) Get(ctx context.Context, key string, load func() (T, error)) (T, error) {
	val, err := c.c.rdb.Get(ctx, key).Bytes()
	if err != nil && !errors.Is(err, goredis.Nil) {
		var zero T
		return zero, err
	}

	if len(val) > 0 {
		var result T
		if err := json.Unmarshal(val, &result); err != nil {
			var zero T
			return zero, err
		}
		return result, nil
	}

	data, err := load()
	if err != nil {
		var zero T
		return zero, err
	}

	encoded, err := json.Marshal(data)
	if err != nil {
		return data, nil
	}

	if err := c.c.rdb.Set(ctx, key, encoded, c.defaultTTL).Err(); err != nil {
		return data, nil
	}

	return data, nil
}

func (c *Cache[T]) Set(ctx context.Context, key string, value T, ttl time.Duration) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.c.rdb.Set(ctx, key, encoded, ttl).Err()
}

func (c *Cache[T]) Del(ctx context.Context, key string) error {
	return c.c.rdb.Del(ctx, key).Err()
}
