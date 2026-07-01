# Redis 客户端设计

## 范围

`internal/platform/redis/` 平铺三文件：client.go + lock.go + cache.go。无子包。

## 组件

### client.go — go-redis v9 基础封装

```go
func New(cfg config.RedisConfig) (*Client, error)
func (c *Client) Close() error
// Get/Set/Del/Exists/Expire/TTL/Incr/Decr
```

`Client` 包装 `redis.UniversalClient`，支持单机模式和后续哨兵/集群切换。

### lock.go — 分布式锁

```go
func NewLock(client rediser, key string, ttl time.Duration) *Lock
func (l *Lock) TryLock(ctx context.Context) (bool, error)  // SET NX, 非阻塞
func (l *Lock) Lock(ctx context.Context) error              // 重试轮询直到超时
func (l *Lock) Unlock(ctx context.Context) error            // Lua EVAL 原子释放
```

直接使用 `redis.UniversalClient` 接口。测试用 testcontainers 起真实 Redis（复用项目现有模式）。

### cache.go — 缓存模板

```go
func NewCache(client rediser, defaultTTL time.Duration) *Cache
func (c *Cache) Get(ctx, key string, load func() (any, error)) (any, error)  // read-through
func (c *Cache) Set(ctx, key string, value any, ttl time.Duration) error
func (c *Cache) Del(ctx, key string) error
```

`Get` 流程：查 Redis → 命中返回 → 未命中调 `load` → 写回 Redis → 返回。

## 配置

复用 `config.RedisConfig{Addr string}`，当前只配 Addr，后续需要时加 Password/DB。

## 错误处理

- 锁获取超时返回裸 error（`"timed out waiting for lock: <key>"`）
- 基础操作（Get/Set/Del）返回裸 Go error，统一由调用方上层决定是否包装 AppError
- 连接失败在 `New()` 时通过 `Ping` 验证，失败返回带 cause 的 error

## 测试

- `client_test.go` — 集成测试连本地 Redis（testcontainers 或用配置的 DSN）
- `lock_test.go` — 基本锁定/超时/释放/竞争
- `cache_test.go` — read-through/写入/删除/TTL 行为
