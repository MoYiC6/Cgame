# Go 后端项目初始化骨架设计

> 设计日期：2026-05-13
> 范围：根据 `go_backend_architecture_payment_order_inventory_notification.md` 与 `go_backend_development_standards.md`，初始化一个可启动的 Go 模块化单体后端骨架。

## 1. 目标

本次初始化仅构建“可启动、可扩展、边界清晰”的工程骨架，不在第一步实现真实业务能力或完整基础设施接入。

完成后应满足以下目标：

- `go run ./cmd/api` 可以启动 HTTP 服务。
- `go run ./cmd/worker` 可以启动 Worker 入口进程。
- 项目目录结构与架构文档、开发规范保持一致。
- 代码依赖方向明确：`cmd -> bootstrap -> modules/clients/platform`。
- 建立最小可复用平台组件入口：`config`、`logger`、`errors`、`response`、`observability`、`bootstrap`。
- 四个核心业务域 `order`、`payment`、`inventory`、`notification` 具备统一模块骨架。
- 提前预留 request/trace 透传、worker 任务注册、后续基础设施扩展的接口边界，避免二次拆骨架。

## 2. 本轮范围

### 2.1 会做的内容

初始化“方案 2：可启动骨架”，包括：

- 标准目录结构创建。
- `cmd/api` 与 `cmd/worker` 启动入口。
- `internal/bootstrap` 应用装配层。
- `internal/platform/config`：强类型配置加载与基础扩展字段预留。
- `internal/platform/logger`：统一日志入口。
- `internal/platform/errors`：统一应用错误模型。
- `internal/platform/response`：统一 API 响应模型。
- `internal/platform/observability`：tracing / propagation / request context 占位接口。
- API 基础路由、健康检查路由、请求 ID / Trace ID 中间件占位。
- Worker 基础注册接口与空任务运行骨架。
- 四个业务模块的最小骨架文件。
- 第三方 client 目录、最小接口约束与占位实现落点。
- 基础配置文件（含 `config.test.yaml`）、OpenAPI 占位文件、工程占位文件。
- 最小 `ping` 集成测试骨架。
- Makefile 与基础 CI 命令占位。

### 2.2 本轮不会做的内容

为了控制初始化阶段复杂度，本轮明确不做：

- 不接真实 PostgreSQL / MySQL。
- 不接 Redis。
- 不接 MQ。
- 不引入 sqlc 生成代码。
- 不实现真实数据库 migration 内容。
- 不落真实 OpenTelemetry exporter、collector 或指标上报逻辑。
- 不实现真实订单、支付、库存、通知业务流程。
- 不接真实支付网关、短信、邮件服务。
- 不搭完整 Docker Compose 基础设施环境。

这些能力只预留目录、边界和落点，后续按计划逐步接入。

## 3. 推荐目录结构

```text
backend/
  cmd/
    api/
      main.go
    worker/
      main.go

  api/
    openapi.yaml

  configs/
    config.local.yaml
    config.dev.yaml
    config.test.yaml
    config.prod.yaml

  internal/
    bootstrap/
      app.go
      server.go
      worker.go
      middleware.go

    platform/
      config/
        config.go
      logger/
        logger.go
      errors/
        errors.go
      response/
        response.go
      observability/
        tracer.go
        propagator.go
        context.go

    modules/
      order/
        handler.go
        service.go
        repository.go
        dto.go
        model.go
        status.go
        events.go
      payment/
        handler.go
        service.go
        repository.go
        dto.go
        model.go
        status.go
        events.go
      inventory/
        handler.go
        service.go
        repository.go
        dto.go
        model.go
        status.go
        events.go
      notification/
        handler.go
        service.go
        repository.go
        dto.go
        model.go
        status.go
        events.go

    clients/
      paymentgateway/
        client.go
      sms/
        client.go
      email/
        client.go

  migrations/
    .gitkeep

  sql/
    queries/
      .gitkeep

  test/
    integration/
      ping_test.go
    fixtures/
      .gitkeep

  Makefile
  .golangci.yml
  .env.example
```

## 4. 架构与装配设计

### 4.1 启动链路

API 启动流程：

1. `cmd/api/main.go` 读取配置。
2. 初始化 logger。
3. 初始化 observability 占位实现。
4. 调用 `bootstrap` 创建应用实例。
5. 构建 Gin Engine。
6. 注册 recovery、request ID、trace context 中间件。
7. 注册健康检查与模块路由。
8. 启动 HTTP Server。

Worker 启动流程：

1. `cmd/worker/main.go` 读取配置。
2. 初始化 logger。
3. 初始化 observability 占位实现。
4. 调用 `bootstrap` 创建 Worker 应用。
5. 注册空任务或 no-op task。
6. 启动 Worker 进程或阻塞等待。

### 4.2 bootstrap 职责

`internal/bootstrap` 只负责装配，不承担业务逻辑：

- 聚合配置、logger、tracer、middleware。
- 初始化模块 service / handler。
- 注册 HTTP 路由。
- 创建 worker 运行入口。
- 聚合 client 接口实例。

后续接入数据库、Redis、MQ 时，应优先扩展 bootstrap 与 platform，而不是把初始化逻辑散落到 `cmd` 或模块内部。

## 5. 平台组件设计

### 5.1 config

配置组件采用强类型结构体，业务代码禁止直接读取环境变量或 Viper。

首轮除了最小可用字段外，保留后续基础设施扩展字段，避免后期因配置拆包或大范围改签名而重新调整骨架。

建议结构：

```go
type Config struct {
    App    AppConfig
    Server ServerConfig
    Log    LogConfig
    DB     DBConfig
    Redis  RedisConfig
    MQ     MQConfig
}

type AppConfig struct {
    Name string
    Env  string
}

type ServerConfig struct {
    Addr string
}

type LogConfig struct {
    Level string
}

type DBConfig struct {
    Driver string
    DSN    string
}

type RedisConfig struct {
    Addr string
}

type MQConfig struct {
    Driver string
    TopicPrefix string
}
```

本轮实现要求：

- `App / Server / Log` 真正可加载。
- `DB / Redis / MQ` 可先作为预留字段存在，不要求完成真实初始化。
- 配置样例中应明确标注这些字段为“预留，初始化阶段不使用”，减少误用。

配置文件例如：

```yaml
app:
  name: backend
  env: local

server:
  addr: ":8080"

log:
  level: info

db:
  # 预留，初始化阶段不使用
  driver: postgres
  dsn: ""

redis:
  # 预留，初始化阶段不使用
  addr: ""

mq:
  # 预留，初始化阶段不使用
  driver: ""
  topic_prefix: backend
```

### 5.2 logger

日志先使用标准库 `log/slog` 做轻量封装。

原因：

- 初始化成本低。
- 可先统一结构化输出入口。
- 后续可平滑替换为 zap，而不影响上层调用方式。

### 5.3 observability

虽然本轮不接真实 OpenTelemetry，但建议在初始化阶段就建立接口边界，让 API 和 Worker 都能依赖抽象，而不是未来再做横切式改造。

这里建议把“span 创建”和“上下文传播”分开建模，而不是揉进同一个接口：

```go
type Tracer interface {
    StartSpan(ctx context.Context, name string) (context.Context, Span)
}

type Span interface {
    End(err error)
}

type Propagator interface {
    Inject(ctx context.Context, carrier Carrier)
    Extract(ctx context.Context, carrier Carrier) context.Context
}

type Carrier interface {
    Get(key string) string
    Set(key string, value string)
    Keys() []string
}
```

同时提供 request / trace 上下文 hook 占位，例如：

- `WithRequestID(ctx, requestID)`
- `RequestIDFromContext(ctx)`
- `WithTraceID(ctx, traceID)`
- `TraceIDFromContext(ctx)`

其中：

- `Tracer` 只负责开始 span，保持职责单一。
- `Propagator` 负责未来 HTTP Header、MQ metadata 的透传占位。
- `Carrier` 采用最小文本键值抽象，兼容 HTTP Header、`map[string]string`、消息 metadata 等载体。
- 本轮不要求真实 trace 上报，但要求 API/Worker 代码已经按接口调用，避免未来大改函数签名。
- 不把全局默认 tracer / propagator 作为核心设计契约；更推荐 bootstrap 显式注入 no-op 实现，避免隐式全局状态蔓延。

### 5.4 errors

先建立统一 `AppError` 模型，满足：

- 稳定错误码。
- HTTP 状态码映射。
- 安全错误信息与底层 cause 分离。

### 5.5 clients 初始化约束

虽然本轮不接真实第三方服务，但 `internal/clients` 的接口设计应从第一天遵守以下约束：

- 外部请求必须预留 timeout 控制点。
- 外部请求必须预留 trace propagation 注入点。
- 必须预留请求摘要、响应摘要、耗时和错误码记录位置。
- 敏感字段必须可脱敏。

初始化阶段不要求真实实现这些能力，但要求接口与装配方式不要阻断后续接入。

### 5.6 response

统一 API 响应结构，确保健康检查和模块 ping 接口都遵循同一格式，便于后续 Handler 统一输出。

建议响应写入逻辑能够从上下文中读取 `request_id`，并为后续挂入 `trace_id` 保留空间。

## 6. 中间件与请求上下文占位

初始化阶段应提供最小中间件占位，保证请求上下文约定从第一天建立，而不是后面再补。

建议包括：

- `RequestIDMiddleware`：生成或透传 `X-Request-ID`。
- `TraceContextMiddleware`：为后续 trace 接入保留上下文写入点。
- `RecoveryMiddleware`：统一捕获 panic 并返回标准错误响应。

目标不是完整 observability，而是让 Handler/Service 从第一天就能从 `context.Context` 中读取 request / trace 信息。

## 7. Worker 扩展骨架设计

Worker 不能只是一个“阻塞等待”的空壳，建议在初始化阶段就建立任务注册接口，后续无论接队列消费、重试任务还是定时任务，都可以复用同一挂载方式。

建议最小接口：

```go
type Worker interface {
    RegisterTask(name string, handler func(ctx context.Context) error)
    Run(ctx context.Context) error
}
```

对于健康检查，不建议把 `Status(ctx context.Context) error` 强行放进 `Worker` 核心接口，否则会把运行器职责和任务探针职责耦合在一起。更合适的做法是：

- 保持 `Worker` 接口最小化。
- 为后续任务探针预留可选扩展接口，例如 `TaskProbe`。
- 在实现计划中为 no-op task 或注册任务保留健康检查挂点。

可选扩展示例：

```go
type TaskProbe interface {
    Probe(ctx context.Context) error
}
```

如果后续确实需要更丰富的状态对象，再单独增加 `TaskStatusProvider` 一类接口，而不是在骨架阶段提前塞进核心契约。

初始化时可以注册一个 no-op task 或 health task，重点是建立：

- worker 生命周期入口
- 任务注册点
- 后续接 MQ / cron 的扩展边界
- 任务健康探针的可扩展挂点

## 8. 业务模块骨架设计

四个业务域本轮只做“最小统一骨架”，不做真实业务：

- `handler.go`：定义 Handler 与 `Ping` 接口。
- `service.go`：定义 Service 接口与基础实现。
- `repository.go`：定义 Repository 接口占位。
- `dto.go`：定义最小 DTO。
- `model.go`：定义最小领域模型。
- `status.go`：定义状态类型与常量。
- `events.go`：定义事件名常量。

这样可以保证：

- 四个业务域从第一天开始边界统一。
- 后续扩展时不需要重新调整目录与职责。
- Handler 不直接访问 Repository，符合开发规范。

### 8.1 模块占位示例约定

为了保证团队后续扩展时风格统一，建议在每个模块的占位文件中就体现最小约定。例如：

```go
// internal/modules/order/dto.go
type CreateOrderRequest struct {
    UserID int64  `json:"user_id"`
    SKU    string `json:"sku"`
    Qty    int    `json:"qty"`
}

type OrderResponse struct {
    OrderNo string `json:"order_no"`
    Status  string `json:"status"`
}
```

```go
// internal/modules/order/status.go
const (
    OrderPending   = "pending"
    OrderPaid      = "paid"
    OrderCancelled = "cancelled"
)
```

```go
// internal/modules/order/events.go
const (
    EventOrderCreated = "order.created.v1"
    EventOrderPaid    = "order.paid.v1"
)
```

这里的重点不是提前完整定义业务状态或完整 DTO，而是明确：

- DTO、状态与事件使用集中定义。
- 命名风格从一开始就一致。
- 新人能快速理解 `Handler -> Service -> Repository` 的占位调用链。
- 后续扩展时不会把状态散落进 handler / service。

## 9. clients 接口约束

`internal/clients` 下的第三方适配器目录不应只是空目录，建议在设计上明确每个 client 对外暴露接口而不是直接暴露 SDK。

示例：

```go
type SMSClient interface {
    Send(ctx context.Context, to string, msg string) error
}
```

```go
type EmailClient interface {
    Send(ctx context.Context, to string, subject string, body string) error
}
```

```go
type PaymentGateway interface {
    CreatePayment(ctx context.Context, req CreatePaymentRequest) (*CreatePaymentResponse, error)
}
```

初始化阶段只要求：

- 明确接口定义位置。
- 明确未来实现位于 `clients/<provider>`。
- 业务模块依赖接口，不依赖具体 SDK。

## 10. 初始 HTTP 路由设计

初始化阶段只保留最小验证路由：

- `GET /healthz`
- `GET /readyz`
- `GET /api/v1/orders/ping`
- `GET /api/v1/payments/ping`
- `GET /api/v1/inventories/ping`
- `GET /api/v1/notifications/ping`

这些接口的目的不是承载业务，而是验证：

- 路由组织方式是否合理。
- bootstrap 是否正确装配模块。
- 统一响应结构是否可用。
- request / trace 上下文是否已贯通。
- 后续模块接入是否遵循相同模式。

建议统一响应结构如下：

```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "module": "order"
  },
  "request_id": ""
}
```

## 11. 测试骨架与工程命令

### 11.1 测试骨架

除了 `test/integration/` 目录外，初始化阶段建议直接加入一个最小 `ping` 集成测试，验证：

- server 能被构建。
- 路由已注册。
- handler / response / bootstrap 协作正常。
- `RequestIDMiddleware` 已把 request_id 写入上下文与响应。
- `TraceContextMiddleware` 已具备最小透传占位，至少能验证 trace 信息进入请求上下文。
- worker bootstrap 能注册 placeholder task，并在 context 取消时可退出。
- 如果 task 实现了 `TaskProbe`，其探针挂点可被识别。

例如：

- `test/integration/ping_test.go`
- 通过 `httptest` 或最小集成方式调用 `/healthz` 和某个模块的 `/ping`
- 断言 HTTP 状态码与统一响应结构
- 断言 request_id 存在，trace context 占位已贯通
- 对 worker 进行最小生命周期验证，不引入真实队列、重试或调度语义

这样能保证初始化不是“只有目录没有验证”。

### 11.2 Makefile / CI 占位

文档中提到的 `Makefile` 不应只作为文件名占位，建议初始化阶段直接提供最小命令：

```makefile
run-api:
	go run ./cmd/api

run-worker:
	go run ./cmd/worker

test:
	go test ./...

lint:
	golangci-lint run
```

这样后续接 CI/CD 时，可以直接复用本地工程命令，而不是重新设计脚本入口。

## 12. 工程边界与约束

初始化实现必须遵守以下约束：

- `cmd` 层只作为装配入口，不写业务逻辑。
- Handler 不直接调用 Repository。
- 不创建 `utils` 大杂烩包。
- 业务模块不依赖具体第三方 SDK。
- `platform` 不依赖业务模块。
- 不用伪造数据库、伪造支付流程来“凑完整”。
- 不在初始化阶段提前做与当前目标无关的抽象。
- 允许为未来扩展预留接口，但不允许借“预留”之名提前实现整套基础设施。

## 13. 验收标准

本轮初始化完成时，应满足：

1. 目录结构已按设计落地。
2. `go run ./cmd/api` 可启动。
3. `go run ./cmd/worker` 可启动。
4. `go test ./...` 通过。
5. 代码可通过 `gofmt`。
6. API 已具备 request ID / trace 占位中间件。
7. Worker 已具备任务注册接口。
8. 初始化内容与“可启动骨架”范围一致，没有提前实现过量业务。

## 14. 后续衔接

在本设计通过后，下一阶段应进入实现计划编写，而不是直接大范围自由编码。实现计划需要把初始化拆成小步任务，例如：

- 创建目录与工程文件。
- 写配置组件测试与实现。
- 写 logger 组件测试与实现。
- 写 observability / middleware 占位测试与实现。
- 写 response / errors 组件测试与实现。
- 写 bootstrap 与 server/worker 启动链路。
- 为四个模块建立统一 ping handler。
- 编写最小集成测试。
- 补齐 Makefile 与工程命令。
- 验证启动与测试通过。

这样可以确保初始化过程可执行、可验证、可回滚。
