package redis_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"backend/internal/platform/config"
	pkg "backend/internal/platform/redis"

	goredis "github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestClientSetGet(t *testing.T) {
	ctx := context.Background()
	rdb := setupRedis(t)
	defer rdb.Close()

	if err := rdb.Set(ctx, "test:key", "hello", 0); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	val, err := rdb.Get(ctx, "test:key")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if val != "hello" {
		t.Fatalf("Get() = %q, want %q", val, "hello")
	}
}

func TestClientGetMiss(t *testing.T) {
	ctx := context.Background()
	rdb := setupRedis(t)
	defer rdb.Close()

	_, err := rdb.Get(ctx, "test:missing")
	if !errors.Is(err, goredis.Nil) {
		t.Fatalf("Get() expected goredis.Nil, got %v", err)
	}
}

func TestClientDel(t *testing.T) {
	ctx := context.Background()
	rdb := setupRedis(t)
	defer rdb.Close()

	rdb.Set(ctx, "test:del", "x", 0)
	if err := rdb.Del(ctx, "test:del"); err != nil {
		t.Fatalf("Del() error = %v", err)
	}

	ok, err := rdb.Exists(ctx, "test:del")
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if ok {
		t.Fatal("key still exists after Del")
	}
}

func TestExists(t *testing.T) {
	ctx := context.Background()
	rdb := setupRedis(t)
	defer rdb.Close()

	ok, err := rdb.Exists(ctx, "test:exists:nope")
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if ok {
		t.Fatal("Exists() = true for missing key")
	}

	rdb.Set(ctx, "test:exists:yes", "v", 0)
	ok, err = rdb.Exists(ctx, "test:exists:yes")
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if !ok {
		t.Fatal("Exists() = false for existing key")
	}
}

func TestExpireTTL(t *testing.T) {
	ctx := context.Background()
	rdb := setupRedis(t)
	defer rdb.Close()

	rdb.Set(ctx, "test:ttl", "v", 0)
	rdb.Expire(ctx, "test:ttl", 10*time.Second)

	d, err := rdb.TTL(ctx, "test:ttl")
	if err != nil {
		t.Fatalf("TTL() error = %v", err)
	}
	if d <= 0 || d > 10*time.Second {
		t.Fatalf("TTL() = %v, want ~10s", d)
	}
}

func TestIncrDecr(t *testing.T) {
	ctx := context.Background()
	rdb := setupRedis(t)
	defer rdb.Close()

	n, err := rdb.Incr(ctx, "test:counter")
	if err != nil {
		t.Fatalf("Incr() error = %v", err)
	}
	if n != 1 {
		t.Fatalf("Incr() = %d, want 1", n)
	}

	n, err = rdb.Decr(ctx, "test:counter")
	if err != nil {
		t.Fatalf("Decr() error = %v", err)
	}
	if n != 0 {
		t.Fatalf("Decr() = %d, want 0", n)
	}
}

func TestLockTryLock(t *testing.T) {
	ctx := context.Background()
	rdb := setupRedis(t)
	defer rdb.Close()

	lock := pkg.NewLock(rdb, "test:lock:1", 5*time.Second)
	ok, err := lock.TryLock(ctx)
	if err != nil {
		t.Fatalf("TryLock() error = %v", err)
	}
	if !ok {
		t.Fatal("TryLock() = false, want true")
	}
	defer lock.Unlock(ctx)

	lock2 := pkg.NewLock(rdb, "test:lock:1", 5*time.Second)
	ok, err = lock2.TryLock(ctx)
	if err != nil {
		t.Fatalf("TryLock() error = %v", err)
	}
	if ok {
		t.Fatal("TryLock() = true on already held lock")
	}
}

func TestLockUnlock(t *testing.T) {
	ctx := context.Background()
	rdb := setupRedis(t)
	defer rdb.Close()

	lock := pkg.NewLock(rdb, "test:lock:2", 5*time.Second)
	lock.TryLock(ctx)
	lock.Unlock(ctx)

	lock2 := pkg.NewLock(rdb, "test:lock:2", 5*time.Second)
	ok, err := lock2.TryLock(ctx)
	if err != nil {
		t.Fatalf("TryLock after unlock error = %v", err)
	}
	if !ok {
		t.Fatal("TryLock after unlock = false, want true")
	}
	lock2.Unlock(ctx)
}

func TestLockTimeout(t *testing.T) {
	ctx := context.Background()
	rdb := setupRedis(t)
	defer rdb.Close()

	lock := pkg.NewLock(rdb, "test:lock:3", 5*time.Second)
	lock.TryLock(ctx)
	defer lock.Unlock(ctx)

	lock2 := pkg.NewLock(rdb, "test:lock:3", 5*time.Second)
	err := lock2.Lock(ctx, 200*time.Millisecond)
	if err == nil {
		t.Fatal("Lock() expected timeout error")
	}
}

func TestCacheGetReadThrough(t *testing.T) {
	ctx := context.Background()
	rdb := setupRedis(t)
	defer rdb.Close()

	cache := pkg.NewCache[string](rdb, 10*time.Second)
	loadCount := 0

	val, err := cache.Get(ctx, "test:cache:rt", func() (string, error) {
		loadCount++
		return "loaded", nil
	})
	if err != nil {
		t.Fatalf("Cache.Get() error = %v", err)
	}
	if val != "loaded" {
		t.Fatalf("Cache.Get() = %q, want %q", val, "loaded")
	}
	if loadCount != 1 {
		t.Fatalf("load called %d times, want 1", loadCount)
	}

	val, err = cache.Get(ctx, "test:cache:rt", func() (string, error) {
		loadCount++
		return "loaded", nil
	})
	if err != nil {
		t.Fatalf("Cache.Get() error = %v", err)
	}
	if val != "loaded" {
		t.Fatalf("Cache.Get() = %q, want %q", val, "loaded")
	}
	if loadCount != 1 {
		t.Fatal("load called again on cached key")
	}
}

func TestCacheSetGet(t *testing.T) {
	ctx := context.Background()
	rdb := setupRedis(t)
	defer rdb.Close()

	cache := pkg.NewCache[string](rdb, 10*time.Second)

	if err := cache.Set(ctx, "test:cache:sg", "hello", 0); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	val, err := cache.Get(ctx, "test:cache:sg", func() (string, error) {
		return "", errors.New("should not be called")
	})
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if val != "hello" {
		t.Fatalf("Get() = %q, want %q", val, "hello")
	}
}

func TestCacheDel(t *testing.T) {
	ctx := context.Background()
	rdb := setupRedis(t)
	defer rdb.Close()

	cache := pkg.NewCache[string](rdb, 10*time.Second)
	cache.Set(ctx, "test:cache:del", "x", 0)

	if err := cache.Del(ctx, "test:cache:del"); err != nil {
		t.Fatalf("Del() error = %v", err)
	}

	ok, err := rdb.Exists(ctx, "test:cache:del")
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if ok {
		t.Fatal("key still exists after cache Del")
	}
}

func TestCacheStruct(t *testing.T) {
	ctx := context.Background()
	rdb := setupRedis(t)
	defer rdb.Close()

	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	cache := pkg.NewCache[User](rdb, 10*time.Second)
	loadCount := 0

	u, err := cache.Get(ctx, "test:cache:user", func() (User, error) {
		loadCount++
		return User{ID: 1, Name: "alice"}, nil
	})
	if err != nil {
		t.Fatalf("Cache.Get() error = %v", err)
	}
	if u.ID != 1 || u.Name != "alice" {
		t.Fatalf("Cache.Get() = %+v, want {1 alice}", u)
	}

	u, err = cache.Get(ctx, "test:cache:user", func() (User, error) {
		loadCount++
		return User{}, nil
	})
	if err != nil {
		t.Fatalf("Cache.Get() error = %v", err)
	}
	if u.ID != 1 || u.Name != "alice" {
		t.Fatalf("Cache.Get() = %+v, want {1 alice} after cache hit", u)
	}
	if loadCount != 1 {
		t.Fatal("load called again on cached key")
	}
}

func setupRedis(t *testing.T) *pkg.Client {
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

	cfg := config.RedisConfig{Addr: fmt.Sprintf("localhost:%s", port.Port())}
	rdb, err := pkg.New(cfg)
	if err != nil {
		t.Fatalf("pkg.New: %v", err)
	}
	return rdb
}
