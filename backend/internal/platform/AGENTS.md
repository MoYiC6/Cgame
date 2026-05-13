# PLATFORM KNOWLEDGE BASE

## OVERVIEW
`internal/platform` 是共享基础设施层。这里只放可复用能力，不放业务规则。

## STRUCTURE
```text
internal/platform/
├── config/
├── logger/
├── errors/
├── response/
├── observability/
└── database/
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| 配置加载 | `config/config.go` | `LoadConfig`、env override、校验、脱敏摘要 |
| 统一日志 | `logger/logger.go` | 结构化日志、字段助手、trace/request 关联 |
| 错误模型 | `errors/errors.go` | `AppError`、`Code`、`Status`、`SafeMessage` |
| API 响应 | `response/response.go` | 成功/失败响应结构 |
| 追踪上下文 | `observability/*.go` | request_id、trace_id、propagator、tracer |
| 数据库抽象 | `database/db.go` | DB 接口与健康检查基础能力 |

## CONVENTIONS
- 平台组件不得在内部隐式读取环境变量；需要运行时配置的组件应由调用方显式传入所需参数或配置，`config` 包除外。
- 平台 API 要稳定、强类型、可测试；避免返回裸 `map[string]any` 作为核心接口。
- 当前配置摘要通过 `MaskedSummary()` 脱敏；新增日志字段或响应字段时，必须避免直接输出 DSN、token、credential 等敏感值。
- 错误统一走 `AppError` 及其包装函数；对外响应只暴露安全 message。
- 可观测性是平台职责：request_id、trace_id、传播器、中间件协作都应围绕 `observability` 保持一致。
- 平台包可以被 modules/bootstrap 依赖，但绝不能反向依赖业务模块。

## ANTI-PATTERNS
- 不要在 `platform` 中写订单、支付、库存、通知等业务规则。
- 不要新增 `utils` 大杂烩包；按能力拆分小包。
- 不要让业务代码直接读环境变量、自己拼响应 JSON、自己定义另一套错误码体系。
- 不要吞错误；保留 cause，并通过统一 API 映射。
- 不要在平台层引入对具体模块实现的 import。

## TESTING
- 优先写组件级单元测试和契约测试，例如 `config_test.go`、`logger_contract_test.go`。
- 配置测试要覆盖：文件加载、环境覆盖、必填校验、脱敏摘要。
- 错误/响应测试要覆盖：稳定错误码、HTTP 状态、safe message。

## NOTES
- 目前 `config/config.go` 是符号最密集的共享组件之一；改动它通常会影响 API、worker、测试和启动日志。
