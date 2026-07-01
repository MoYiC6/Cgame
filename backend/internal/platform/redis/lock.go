package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const unlockScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0
`

var unlockCmd = goredis.NewScript(unlockScript)

type Lock struct {
	c   *Client
	key string
	val string
	ttl time.Duration
}

func NewLock(c *Client, key string, ttl time.Duration) *Lock {
	return &Lock{c: c, key: key, val: fmt.Sprintf("%d", time.Now().UnixNano()), ttl: ttl}
}

func (l *Lock) TryLock(ctx context.Context) (bool, error) {
	ok, err := l.c.rdb.SetNX(ctx, l.key, l.val, l.ttl).Result()
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (l *Lock) Lock(ctx context.Context, timeout time.Duration) error {
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		ok, err := l.TryLock(ctx)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline.C:
			return fmt.Errorf("timed out waiting for lock: %s", l.key)
		case <-ticker.C:
		}
	}
}

func (l *Lock) Unlock(ctx context.Context) error {
	return unlockCmd.Run(ctx, l.c.rdb, []string{l.key}, l.val).Err()
}
