package redis

import (
	"context"
	"time"

	"backend/internal/platform/config"
	goredis "github.com/redis/go-redis/v9"
)

type Client struct {
	rdb goredis.UniversalClient
}

func New(cfg config.RedisConfig) (*Client, error) {
	rdb := goredis.NewClient(&goredis.Options{Addr: cfg.Addr})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		rdb.Close()
		return nil, err
	}
	return &Client{rdb: rdb}, nil
}

func (c *Client) Close() error {
	return c.rdb.Close()
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.rdb.Get(ctx, key).Result()
}

func (c *Client) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.rdb.Del(ctx, keys...).Err()
}

func (c *Client) Exists(ctx context.Context, keys ...string) (bool, error) {
	n, err := c.rdb.Exists(ctx, keys...).Result()
	return n > 0, err
}

func (c *Client) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return c.rdb.Expire(ctx, key, ttl).Err()
}

func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.rdb.TTL(ctx, key).Result()
}

func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	return c.rdb.Incr(ctx, key).Result()
}

func (c *Client) Decr(ctx context.Context, key string) (int64, error) {
	return c.rdb.Decr(ctx, key).Result()
}
