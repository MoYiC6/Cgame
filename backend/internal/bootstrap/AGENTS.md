# BOOTSTRAP KNOWLEDGE BASE

## OVERVIEW
`internal/bootstrap` 负责进程装配与生命周期管理：Gin 引擎、中间件、健康检查、HTTP server、worker、shutdown 聚合都在这里。

## STRUCTURE
```text
internal/bootstrap/
├── app.go
├── server.go
├── http_server.go
├── middleware.go
└── worker.go
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| 共享依赖容器 | `app.go` | `Dependencies`、`Shutdowner`、`App.Shutdown` |
| 路由装配 | `server.go` | `/healthz`、`/readyz`、`/api/v1` group；顺序：RequestID → Trace → **AccessLog** → CORS → Security → RateLimit → Recovery |
| HTTP 服务封装 | `http_server.go` | server 启停包装 |
| 中间件 | `middleware.go` | request id、trace context、**access log**、recovery、rate limit、CORS、security headers |
| Worker 装配 | `worker.go` | task 注册、probe、run/shutdown |

## CONVENTIONS
- bootstrap 只做装配，不做业务决策。
- API 启动链路：加载 config → 初始化 logger / observability / database 依赖 → 组装 `Dependencies` → 注册 routes → 启动 HTTP server。
- Worker 启动链路：加载 config → 初始化 logger → 创建 worker → 注册 tasks → 启动 worker。
- `server.go` 统一挂健康检查和 API 根路由；模块只注册自己相对路径。
- 中间件必须集中维护，尤其是 request id、trace 透传、panic recovery。
- 任何可关闭资源都应实现 `Shutdown(ctx)` 并纳入 `NewApp(...)` 聚合。
- worker 任务注册要考虑 probe、取消和错误聚合；当前 `Probe(ctx)` 需要显式调用，不是 `Run(ctx)` 自动执行。

## ANTI-PATTERNS
- 不要在 bootstrap 写订单/支付等业务规则。
- 不要在不同入口各自复制一套中间件或健康检查逻辑。
- 不要绕过 `Dependencies` 到处散落共享依赖初始化。
- 不要缺少 recovery / trace / request_id；这些属于启动层硬约束。
- 不要让 worker 在 `ctx.Done()` 后继续无界运行。

## TESTING
- `server_test.go`：验证健康检查、路由注册、ready 行为。
- `middleware_test.go`：验证 header 注入、context 透传、panic recovery。
- `app_test.go`：验证 shutdown 聚合和空对象行为。
- `worker_test.go`：验证 task 生命周期、probe、取消后的退出行为。

## NOTES
- `test/integration/ping_test.go` 直接复用这里的装配层构建完整 API 引擎；改 bootstrap 时，优先回归这条集成链路。
