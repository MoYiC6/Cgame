# Scheduler 包完整方案

## 范围

`internal/platform/scheduler/` 可扩展到单 Worker → 多 Worker → 分布式调度。
当前先落地 in-memory 单实例，接口和中间件留出后续替换点。

---

## 核心类型

```go
// JobFunc 是任务执行体。返回 error 时按 RetryPolicy 处理。
type JobFunc func(ctx context.Context) error

// Job 描述一个定时任务
type Job struct {
    Name     string
    Schedule string        // 固定间隔 "5s" 或 cron "0 */5 * * * *"
    Job      JobFunc
}

// RetryPolicy 定义失败重试策略
type RetryPolicy struct {
    MaxRetries int
    Backoff    time.Duration  // 线性退避，可后续改为指数
}

// JobMiddleware 是任务包装器，类似 HTTP middleware
// 用于横切关注点：日志、指标、超时、重试
type JobMiddleware func(next JobFunc) JobFunc
```

---

## Scheduler 接口（可替换后端）

```go
// Scheduler 是调度器抽象。实现方负责按 Schedule 触发 JobFunc。
// 当前实现：memoryScheduler（in-memory Ticker + 内置 cron 解析）
// 未来实现：redisScheduler（ZSet 协调 + 分布式锁）、cronScheduler（精确 cron）
type Scheduler interface {
    // Register 注册任务。重复注册同名任务返回错误。
    Register(job Job) error

    // Remove 移除已注册任务。移除后下次 tick 不再触发。
    Remove(name string) error

    // Start 启动调度循环。阻塞直到 ctx 取消或 Stop 被调用。
    Start(ctx context.Context) error

    // Stop 停止调度，等待进行中的任务完成。
    Stop(ctx context.Context) error

    // Use 注册全局中间件。中间件在 Start 前注册有效。
    Use(middlewares ...JobMiddleware)
}
```

---

## 文件职责

| 文件 | 内容 |
|------|------|
| `scheduler.go` | `Job`/`JobFunc`/`JobMiddleware`/`RetryPolicy` 类型 + `Scheduler` 接口 |
| `memory.go` | `memoryScheduler`：time.Ticker 轮询 + 内置 cron 解析（最小 cron） |
| `middleware.go` | 预置中间件：`LoggingMiddleware`、`RecoveryMiddleware`、`TimeoutMiddleware`、`RetryMiddleware` |
| `registry.go` | `JobRegistry`：任务注册表（当前 map，可替换为 Redis） |

---

## 调度模型

```
Register(job) → registry[name] = entry
Start(ctx)     → 启动 ticker（按最小公约数或各 job 独立 ticker）
                ↓ 每 tick
                for each entry: 检查是否到期 → 触发 JobFunc
                ↓ JobFunc 返回 error
                通过 RetryPolicy 决定是否重试
                ↓ ctx.Done()
Stop(ctx)      → 停止 ticker，等待进行中的任务结束
```

### Cron 支持（最小实现）

不引入外部 cron 库，内置一个简单解析器，支持标准 5 字段：
- `*` — 任意值
- `*/n` — 每 n 单位
- `a,b,c` — 枚举

```go
// ParseCron 将 "0 */5 * * * *" 转为 next trigger 时间
// 只支持秒/分/时/日/月/星期（6 字段，秒级精度）
func ParseCron(expr string) (*CronSchedule, error)
```

### 重试机制

```go
func RetryMiddleware(policy RetryPolicy) JobMiddleware {
    return func(next JobFunc) JobFunc {
        return func(ctx context.Context) error {
            var err error
            for attempt := 0; attempt <= policy.MaxRetries; attempt++ {
                err = next(ctx)
                if err == nil {
                    return nil
                }
                if attempt < policy.MaxRetries {
                    time.Sleep(policy.Backoff * time.Duration(attempt+1))
                }
            }
            return fmt.Errorf("job failed after %d retries: %w", policy.MaxRetries, err)
        }
    }
}
```

---

## 中间件链

```
JobFunc
  → LoggingMiddleware（记录开始/结束/耗时）
  → RecoveryMiddleware（panic 恢复，返回 error 不崩溃）
  → TimeoutMiddleware（任务级超时）
  → RetryMiddleware（失败重试）
  → 实际执行
```

中间件通过 `Use()` 注册，执行时按注册顺序从外到内包裹。

---

## 可扩展点

| 扩展方向 | 当前实现 | 未来扩展 |
|---|---|---|
| 调度后端 | in-memory Ticker | Redis ZSet + 分布式锁 |
| Cron 解析 | 最小 5 字段解析器 | 替换为 `robfig/cron` 或更精确的调度 |
| 任务存储 | map（内存） | `JobStore` 接口 → Redis/DB |
| 重试策略 | 线性退避 | 指数退避、jitter、DLQ 集成 |
| 任务持久化 | 无 | 注册/执行日志持久化 |
| 分布式协调 | 无 | Redis lock / etcd leader |
| 任务依赖 | 无 | DAG 调度 |
| 指标/追踪 | 基础日志 | OTel span、prometheus metrics |

### JobStore 接口（预留）

```go
// JobStore 抽象任务持久化。当前实现：memoryStore（map）
// 未来实现：redisStore（Hash/Set）
type JobStore interface {
    Save(ctx context.Context, job Job) error
    Load(ctx context.Context) ([]Job, error)
    Delete(ctx context.Context, name string) error
}
```

Worker 启动时从 `JobStore` 恢复注册，实现优雅重启不丢任务。

---

## 配置

复用 `MQConfig` 的 `TopicPrefix`（可选），或新增：

```go
type SchedulerConfig struct {
    TickInterval time.Duration  // 最小 tick 间隔，默认 1s
    MaxRetries   int             // 默认最大重试，默认 3
    Backoff      time.Duration   // 默认退避，默认 1s
}
```

---

## 测试

- `memory_test.go`：基本定时触发、任务完成、移除任务
- `middleware_test.go`：重试中间件、恢复中间件
- `cron_test.go`：cron 表达式解析正确性
- `registry_test.go`：注册/移除/重复注册
