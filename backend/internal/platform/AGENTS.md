# PLATFORM KNOWLEDGE BASE

## OVERVIEW
`internal/platform` 是共享基础设施层。这里只放可复用能力，不放业务规则。

## STRUCTURE
```text
internal/platform/
├── config/
├── database/
├── errors/
├── logger/
├── observability/
├── redis/
├── response/
└── security/
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| 配置加载 | `config/config.go` | `LoadConfig`、env override、校验、脱敏摘要；`LogConfig.Format` 控制 JSON/text |
| 统一日志 | `logger/logger.go` | `Logger` 接口（Info/Warn/Error/Debug）、`New(cfg)` 配置驱动、`NewText()` 测试用、JSON/Text 切换 |
| 错误模型 | `errors/errors.go` | `AppError`（stack + code + cause + HTTP status）、`WithCause`/`StackTrace`、全局 `codes.go` |
| API 响应 | `response/response.go` | `Success`/`Fail(c, err)` 自动 `errors.As` 映射 |
| 追踪上下文 | `observability/*.go` | request_id、trace_id、propagator、tracer |
| 数据库抽象 | `database/db.go` | `PgxPool` 唯一池 + `SQLDB` 派生、`TxOption{ReadOnly,Timeout}`、`WithinTx` 事务 |
| Redis | `redis/client.go` | go-redis v9 封装、分布式锁（`lock.go`）、泛型缓存模板（`cache.go`） |

## CONVENTIONS
- 平台组件不得在内部隐式读取环境变量；需要运行时配置的组件应由调用方显式传入所需参数或配置，`config` 包除外。
- 平台 API 要稳定、强类型、可测试；避免返回裸 `map[string]any` 作为核心接口。
- 当前配置摘要通过 `MaskedSummary()` 脱敏；新增日志字段或响应字段时，必须避免直接输出 DSN、token、credential 等敏感值。
- 错误统一走 `AppError` 及其包装函数；对外响应只暴露安全 message。当前 9 个全局错误码（`codes.go`），启动时唯一性校验。
- 可观测性是平台职责：request_id、trace_id、传播器、中间件协作都应围绕 `observability` 保持一致。
- 平台包可以被 modules/bootstrap 依赖，但绝不能反向依赖业务模块。
- 数据库连接池统一：`cmd/*/main.go` 只创建 `PgxPool`，`SQLDB` 通过 `Pool.SQLDB()` 派生，不新建第二个池。

## ANTI-PATTERNS
- 不要在 `platform` 中写订单、支付、库存、通知等业务规则。
- 不要新增 `utils` 大杂烩包；按能力拆分小包。
- 不要让业务代码直接读环境变量、自己拼响应 JSON、自己定义另一套错误码体系。
- 不要吞错误；保留 cause，并通过统一 API 映射。
- 不要在平台层引入对具体模块实现的 import。
- 不要开启第二个独立数据库连接池；所有 query 走同一个 `PgxPool` 派生。
- 不要在 handler 里手动构造 `AppError`；让 `response.Fail` 做 auto-map。

## TESTING
- 优先写组件级单元测试和契约测试，例如 `config_test.go`、`logger_contract_test.go`。
- 配置测试要覆盖：文件加载、环境覆盖、必填校验、脱敏摘要。
- 错误/响应测试要覆盖：稳定错误码、HTTP 状态、safe message、stack trace 格式。
- 日志测试验证 JSON/Text 两种输出格式正确。
- 数据库测试验证 `TxOption.ReadOnly` 和 `TxOption.Timeout` 生效。

## NOTES
- 目前 `config/config.go` 是符号最密集的共享组件之一；改动它通常会影响 API、worker、测试和启动日志。
- 错误处理层已完成改造：`errors.go` 重写（加 stack）+ 新增 `codes.go` + `response.Fail(c, err)` 全局自动映射。
- 日志层已完成改造：`New(cfg)` 配置驱动（非 `New(level, writer)`）、Debug 级别、JSON/Text 切换、`AccessLogMiddleware` 记录请求。
- 数据库层已完成改造：`PgxPool.SQLDB()` 避免双池、`TxOption` 支持只读和超时。
