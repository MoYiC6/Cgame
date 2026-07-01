# Queue 包设计

## 范围

`internal/platform/queue/` 六文件，无子包。直接复用 redis 包的 `*redis.Client`。

## 接口

```go
type Message struct {
    ID        string
    Body      []byte
    Headers   map[string]string  // 可选，retry 计数等放这里
    CreatedAt time.Time
}

type Producer interface {
    Publish(ctx context.Context, topic string, msg Message) error
    Close() error
}

type Consumer interface {
    Subscribe(ctx context.Context, topic string, handler func(Message) error) error
    Close() error
}
```

## 文件职责

| 文件 | 类型 | 说明 |
|------|------|------|
| `queue.go` | 接口 | `Message` + `Producer`/`Consumer` 接口 |
| `memory.go` | 实现 | 基于 channel 的 in-memory Producer/Consumer |
| `redis_stream.go` | 实现 | 基于 Redis Stream XADD/XREADGROUP/XACK 的持久化驱动 |
| `delayed.go` | 增强 | `DelayedQueue` 包装 Producer，用 ZSet 按时间排序，poll 到期消息转投 stream |
| `priority.go` | 增强 | `PriorityQueue` 包装 Producer，ZSet 按 score 排序，Pop 取最高优先 |
| `deadletter.go` | 增强 | `DLQConsumer` 包装 Consumer，失败时检查 Headers["retry"]，超限则转存 dlq:{topic} stream |

## 延迟队列

```go
type DelayedQueue struct {
    rdb          *redis.Client
    producer     Producer
    pollInterval time.Duration  // 默认 1s
}

func NewDelayedQueue(rdb *redis.Client, p Producer, poll time.Duration) *DelayedQueue
func (d *DelayedQueue) Schedule(ctx, topic, msg, delay) error
func (d *DelayedQueue) Start(ctx, handler) error  // 后台 goroutine 轮询
```

实现：`ZADD delayed:{topic} <unix_ms> <msg_id>`，Start 内 ticker 每秒 `ZRANGEBYSCORE -inf <now>` 取出过期 ID，查对应消息体，XADD 到 stream，ZREM。

## 优先级队列

```go
type PriorityQueue struct {
    rdb      *redis.Client
    producer Producer
}

func NewPriorityQueue(rdb *redis.Client, p Producer) *PriorityQueue
func (p *PriorityQueue) Push(ctx, topic, msg, priority int64) error
func (p *PriorityQueue) Pop(ctx, topic) (Message, error)  // ZREVRANGE 取最高分
```

实现：`ZADD priority:{topic} <score> <msg_id>`，Pop 时 `ZREVRANGE` 取最高，查消息体，`ZREM`。

> 注：优先级与消费并发天然冲突（高优先 starvation 问题），本实现单 Pop 取最高，适合并发低的场景。高并发场景后续改多队列。

## 死信队列

```go
type DLQConfig struct {
    MaxRetries int
    DLQTopic   string
}

type DLQConsumer struct {
    next   Consumer
    dlq    Producer
    config DLQConfig
}

func NewDLQConsumer(next Consumer, dlq Producer, cfg DLQConfig) *DLQConsumer
func (d *DLQConsumer) Subscribe(ctx, topic string, handler func(Message) error) error
```

实现：Subscribe 委托给 next，handler 返回 error 时：
- `retry := int(msg.Headers["retry"]) + 1`
- `retry < MaxRetries` → Headers["retry"]=retry，重新 Publish 原 topic
- `retry >= MaxRetries` → Publish 到 `DLQTopic`

## 配置

复用 `MQConfig`（`config/config.go` 已有），暂只使用 `TopicPrefix` 字段。

## 测试

- `memory_test.go`：基本 Publish/Subscribe + 并发
- `redis_stream_test.go`：testcontainers 起 redis:7，验证 XADD/XREADGROUP 完整流程
- `delayed_test.go`：Schedule 短延迟后验证消息到达
- `priority_test.go`：Push 不同 priority，Pop 按序
- `deadletter_test.go`：handler 失败 N 次后验证进 DLQ
