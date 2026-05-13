# 支付、订单、库存、通知后端架构设计文档

模块化单体起步，支持事件驱动、幂等、可观测和后续服务拆分

版本：1.0  
日期：2026-05-10

# 1. 文档目标与设计原则

本文档用于设计一个以 Go 为主的业务后端系统，覆盖支付、订单、库存、通知四个核心业务域。目标不是把框架堆满，而是建立一套长期可维护的工程地基：组件可复用、代码边界清晰、核心链路幂等、问题可观测、故障可追溯。

## 1.1 适用范围

- 适用于单体优先、后续可演进到微服务的 Go 后端。
- 适用于电商、会员订阅、虚拟商品、库存扣减、支付回调、通知发送等常见业务。
- 默认后端技术栈：`Gin + sqlc/pgx + PostgreSQL/MySQL + Redis + MQ + OpenTelemetry`。
- 如果项目早期规模较小，可以先用同步流程和单库事务；只要保留 Outbox、幂等和模块边界，后期可以平滑拆分。

## 1.2 核心目标

| 目标 | 落地要求 |
| --- | --- |
| 组件可复用 | 基础能力沉淀为 config、logger、database、transaction、cache、idempotency、eventbus、notifier、payment client 等组件。 |
| 代码优雅 | Handler、Service、Repository、Client 分层明确，禁止跨层访问；业务规则集中在 Service/Domain 层。 |
| 出问题可查 | 每个请求都有 request_id、trace_id；关键链路有结构化日志、指标、Trace、审计记录。 |
| 数据一致 | 订单、支付、库存通过事务、幂等、状态机、Outbox 和补偿任务保障最终一致。 |
| 可扩展 | 支付渠道、通知渠道、存储实现、消息队列实现通过接口抽象替换。 |

## 1.3 设计原则

- 单一职责：每个组件只解决一个明确问题。
- 依赖倒置：业务层依赖接口，不直接依赖第三方 SDK、数据库驱动或 MQ 客户端。
- 显式优先：少用全局变量和隐式状态，依赖通过构造函数注入。
- 失败可恢复：外部调用默认可能失败，必须有超时、重试、幂等、补偿。
- 状态机优先：订单、支付、库存、通知都必须有明确状态流转，不允许靠布尔字段拼凑业务状态。
- 日志结构化：日志字段稳定，方便检索、聚合和告警。
- 先单体后拆分：早期用模块化单体降低复杂度，边界按未来服务拆分方式设计。

> **要点：** 本设计推荐“模块化单体 + 事件驱动 + Outbox”的方式起步。它比一开始上复杂微服务更稳，也比随意 CRUD 更适合长期演进。

# 2. 推荐技术栈

## 2.1 基础技术选型

| 层 | 推荐 | 说明 |
| --- | --- | --- |
| HTTP | Gin | 负责路由、中间件、请求绑定、响应输出。 |
| 数据库访问 | sqlc + pgx / database/sql | 核心业务推荐写 SQL 并生成类型安全代码；复杂查询更可控。 |
| 数据库迁移 | goose 或 golang-migrate | 所有 schema 变更必须版本化、可回滚、可审计。 |
| 缓存 | Redis | 用于幂等、缓存、限流、分布式锁、热点数据。 |
| 消息 | Kafka / RabbitMQ / NATS / Redis Stream | 业务量小时可先用表驱动任务，后续替换为 MQ。 |
| 日志 | slog 或 zap | 统一结构化日志，推荐封装 logger 接口。 |
| 配置 | Viper + 强类型 Config | 读取配置后转换为结构体，业务代码禁止直接调用 viper。 |
| 可观测性 | OpenTelemetry | Trace、Metrics、日志关联、关键链路耗时追踪。 |
| 测试 | testing + testify + Testcontainers | 核心链路跑真实数据库和 Redis 容器。 |
| 接口契约 | OpenAPI | 接口先约定，再实现；也便于联调和生成客户端。 |

## 2.2 建议先不上或谨慎使用

| 内容 | 建议 |
| --- | --- |
| 复杂微服务框架 | 不是不能用，而是不建议第一天就把业务拆散。先做好模块边界。 |
| 自研 ORM / 自研 DI 框架 | 投入大、收益低，容易变成维护负担。 |
| 全局单例 | 除 logger、metrics 等基础入口外，业务依赖应通过构造函数传递。 |
| 分布式事务 | 优先用本地事务 + Outbox + 幂等补偿；除非业务真的需要强一致。 |

# 3. 业务域划分

系统按业务能力划分为四个核心域：订单、支付、库存、通知。每个域拥有自己的 Handler、Service、Repository、Model、DTO 和事件定义。跨域调用必须通过 Service 接口或事件完成，禁止直接访问其他域的数据表。

## 3.1 订单域 Order

- 负责订单创建、确认、取消、关闭、支付成功后流转、售后关联。
- 维护订单主状态和订单项，不直接操作支付渠道，不直接发送通知。
- 订单是业务编排中心，但不应该变成所有逻辑的大杂烩。库存和支付由各自组件执行，并通过状态和事件协作。

## 3.2 支付域 Payment

- 负责支付单创建、支付渠道请求、回调验签、支付状态确认、退款单管理。
- 必须支持幂等：同一个业务订单只能有受控数量的支付单，同一个渠道回调只能处理一次。
- 支付回调必须先落库，再执行业务状态流转，防止回调丢失。

## 3.3 库存域 Inventory

- 负责库存查询、库存预占、库存确认扣减、库存释放、库存流水。
- 库存变更必须有流水表，任何库存数量变化都要可追踪。
- 高并发场景优先使用数据库条件更新、乐观锁、库存预占表或 Redis 辅助削峰。

## 3.4 通知域 Notification

- 负责短信、邮件、站内信、Webhook、Push 等通知任务。
- 通知不应阻塞订单核心链路，建议通过事件或 Outbox 异步发送。
- 通知任务必须记录发送状态、失败原因、重试次数和渠道响应。

# 4. 总体架构

## 4.1 模块化单体架构图

```text
Client / Admin / Partner
        |
        v
+----------------------------+
| Gin HTTP Server            |
| middleware: auth, trace,   |
| recovery, request log      |
+-------------+--------------+
              |
              v
+----------------------------+
| Handler Layer              |
| bind / validate / response |
+-------------+--------------+
              |
              v
+----------------------------+
| Service Layer              |
| business rules / tx /      |
| state machine / idempotency|
+------+------+-------+------+
       |      |       |
       v      v       v
   Order   Payment  Inventory
       \      |       /
        \     v      /
         +----Events----+
              |
              v
     Outbox / MQ / Worker
              |
              v
       Notification / Audit
```

## 4.2 分层职责

| 层 | 职责 | 禁止事项 |
| --- | --- | --- |
| Handler | 解析请求、参数校验、调用 Service、统一响应。 | 禁止写业务规则、禁止直接访问数据库。 |
| Service | 业务编排、事务、状态机、权限、幂等、领域规则。 | 禁止拼 SQL、禁止依赖 HTTP 细节。 |
| Repository | 数据库读写、SQL 查询、数据映射。 | 禁止决定业务状态流转。 |
| Client | 第三方服务调用，如支付、短信、邮件、对象存储。 | 禁止让业务层知道第三方协议细节。 |
| Worker | 消费事件、执行异步任务、重试补偿。 | 禁止无幂等消费。 |
| Platform | 日志、配置、数据库、缓存、消息、追踪等基础设施。 | 禁止耦合具体业务。 |

# 5. 推荐项目目录结构

```text
your-app/
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
    config.prod.yaml

  migrations/
    000001_create_orders.up.sql
    000001_create_orders.down.sql

  sql/
    queries/
      orders.sql
      payments.sql
      inventories.sql
      notifications.sql

  internal/
    bootstrap/
      app.go
      server.go
      worker.go

    platform/
      config/
      logger/
      database/
      transaction/
      redis/
      idempotency/
      lock/
      eventbus/
      outbox/
      observability/
      httpx/
      errors/
      response/

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
      inventory/
      notification/

    clients/
      paymentgateway/
      sms/
      email/

  test/
    integration/
    fixtures/

  Makefile
  Dockerfile
  docker-compose.yaml
  .golangci.yml
  .env.example
  README.md
```

## 5.1 目录规则

- `internal/platform` 只放可复用基础组件，不允许写业务规则。
- `internal/modules` 按业务域组织，模块内部自包含。
- `internal/clients` 放第三方服务适配器，业务层只依赖接口。
- `sql/queries` 放 sqlc 查询文件；`migrations` 放数据库版本变更。
- `cmd/api` 和 `cmd/worker` 只是装配入口，禁止堆业务逻辑。

# 6. 基础组件设计

## 6.1 Config 组件

配置组件负责加载文件配置、环境变量覆盖、默认值、必填校验和敏感字段脱敏。业务代码只能依赖强类型 Config，不允许直接调用 Viper。

```go
type Config struct {
    Env     string
    Server  ServerConfig
    DB      DBConfig
    Redis   RedisConfig
    MQ      MQConfig
    Log     LogConfig
    OTEL    OTELConfig
    Payment PaymentConfig
}

func Load(path string) (*Config, error) {
    // read yaml, bind env, validate required fields
    // mask secrets when printing boot logs
    return cfg, nil
}
```

## 6.2 Logger 组件

日志组件统一输出 JSON 结构化日志。所有日志都应尽量带 `request_id`、`trace_id`、`user_id`、`order_id`、`payment_id` 等关键字段。

```go
type Logger interface {
    Debug(ctx context.Context, msg string, fields ...Field)
    Info(ctx context.Context, msg string, fields ...Field)
    Warn(ctx context.Context, msg string, fields ...Field)
    Error(ctx context.Context, msg string, fields ...Field)
}
```

## 6.3 Database 与 Transaction 组件

数据库组件只负责连接池、健康检查和关闭。事务组件负责把事务上下文传入 Repository，避免 Service 到处手写 Begin/Commit/Rollback。

```go
type TxManager interface {
    WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

err := txManager.WithinTx(ctx, func(ctx context.Context) error {
    if err := orderRepo.Create(ctx, order); err != nil {
        return err
    }
    if err := inventorySvc.Reserve(ctx, reserveCmd); err != nil {
        return err
    }
    return outboxRepo.Append(ctx, event)
})
```

## 6.4 Idempotency 组件

幂等组件用于防止重复创建订单、重复处理支付回调、重复消费消息、重复发送通知。幂等必须同时依赖业务唯一键、数据库唯一索引和幂等记录表，不能只靠 Redis。

| 场景 | 幂等键 | 落库策略 |
| --- | --- | --- |
| 创建订单 | `user_id + client_order_no` 或 `Idempotency-Key` | 订单表唯一索引；重复请求返回原订单。 |
| 支付回调 | `channel + channel_trade_no + event_type` | 回调日志表唯一索引；已处理则直接返回成功。 |
| 消息消费 | `event_id + consumer_name` | 消费记录表唯一索引；重复消息跳过。 |
| 通知发送 | `biz_type + biz_id + channel + template` | 通知任务唯一键；重复任务合并或返回已有任务。 |

## 6.5 Outbox 组件

Outbox 用于把业务状态变更和事件发布放进同一个本地事务。业务写库成功后，Worker 扫描 outbox_event 表并发布消息，避免“数据库成功但消息丢失”。

```sql
CREATE TABLE outbox_events (
  id              BIGSERIAL PRIMARY KEY,
  event_id        VARCHAR(64) NOT NULL UNIQUE,
  topic           VARCHAR(128) NOT NULL,
  aggregate_type  VARCHAR(64) NOT NULL,
  aggregate_id    VARCHAR(64) NOT NULL,
  payload         JSONB NOT NULL,
  status          VARCHAR(32) NOT NULL,
  retry_count     INT NOT NULL DEFAULT 0,
  next_retry_at   TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL,
  updated_at      TIMESTAMPTZ NOT NULL
);
```

## 6.6 Error 与 Response 组件

系统错误必须统一转换为 API 错误。Handler 不应该直接返回底层数据库错误、第三方错误或 panic 信息。

```go
type AppError struct {
    Code       string
    Message    string
    HTTPStatus int
    Cause      error
    Fields     map[string]any
}

type APIResponse[T any] struct {
    Code      string `json:"code"`
    Message   string `json:"message"`
    Data      T      `json:"data,omitempty"`
    RequestID string `json:"request_id"`
}
```

# 7. 数据库模型设计

## 7.1 通用字段规范

| 字段 | 说明 |
| --- | --- |
| id | 内部主键，建议 BIGINT/UUID。 |
| biz_id | 业务可见 ID，避免暴露自增主键。 |
| status | 状态机字段，必须枚举化。 |
| created_at / updated_at | 所有业务表必须有。 |
| deleted_at | 需要软删除时才加，不要默认滥用。 |
| version | 需要乐观锁时使用。 |
| metadata | JSON 扩展字段，谨慎使用，不能替代核心字段。 |

## 7.2 核心表清单

| 表 | 用途 | 关键约束 |
| --- | --- | --- |
| orders | 订单主表 | `order_no` 唯一；`user_id + client_order_no` 可选唯一。 |
| order_items | 订单明细 | `order_id + sku_id` 或行号约束。 |
| payment_orders | 支付单 | `payment_no` 唯一；`order_id + status` 需按业务限制。 |
| payment_callback_logs | 支付回调日志 | `channel + channel_event_id` 唯一。 |
| inventories | SKU 库存 | `sku_id + warehouse_id` 唯一；支持 version。 |
| inventory_reservations | 库存预占记录 | `reservation_no` 唯一；关联 order_id。 |
| stock_flows | 库存流水 | 每次库存变动一条流水；`flow_no` 唯一。 |
| notification_tasks | 通知任务 | `task_no` 唯一；支持重试。 |
| outbox_events | 待发布事件 | `event_id` 唯一；Worker 扫描发布。 |
| idempotency_records | 幂等记录 | `scope + idempotency_key` 唯一。 |

## 7.3 订单表核心字段

| 字段 | 类型示例 | 说明 |
| --- | --- | --- |
| order_no | varchar(64) | 业务订单号，外部可见。 |
| user_id | bigint | 下单用户。 |
| status | varchar(32) | PENDING_PAYMENT / PAID / CANCELLED 等。 |
| amount_total | bigint | 订单总金额，单位分。 |
| amount_payable | bigint | 应付金额，单位分。 |
| currency | varchar(8) | 币种，如 CNY/USD。 |
| client_order_no | varchar(128) | 客户端幂等号。 |
| paid_at | timestamp | 支付完成时间。 |
| closed_at | timestamp | 关闭时间。 |

## 7.4 支付表核心字段

| 字段 | 类型示例 | 说明 |
| --- | --- | --- |
| payment_no | varchar(64) | 内部支付单号。 |
| order_no | varchar(64) | 关联订单号。 |
| channel | varchar(32) | 支付渠道，如 wechat/alipay/stripe。 |
| channel_trade_no | varchar(128) | 渠道交易号。 |
| status | varchar(32) | INIT / PAYING / SUCCEEDED / FAILED / CLOSED / REFUNDED。 |
| amount | bigint | 支付金额，单位分。 |
| request_payload | jsonb | 请求渠道参数，注意脱敏。 |
| response_payload | jsonb | 渠道响应，注意脱敏。 |

## 7.5 库存表核心字段

| 字段 | 类型示例 | 说明 |
| --- | --- | --- |
| sku_id | bigint | 商品 SKU。 |
| warehouse_id | bigint | 仓库或库存池。 |
| available_qty | int | 可售库存。 |
| reserved_qty | int | 预占库存。 |
| sold_qty | int | 已售库存。 |
| version | int | 乐观锁版本。 |

# 8. 状态机设计

## 8.1 订单状态

| 状态 | 含义 | 允许流转 |
| --- | --- | --- |
| CREATED | 订单已创建，尚未锁库存或待确认。 | PENDING_PAYMENT / CANCELLED |
| PENDING_PAYMENT | 库存已预占，等待支付。 | PAID / CANCELLED / CLOSED |
| PAID | 支付成功，订单生效。 | FULFILLING / REFUNDING |
| FULFILLING | 履约中。 | COMPLETED / REFUNDING |
| COMPLETED | 订单完成。 | REFUNDING |
| CANCELLED | 用户主动取消或业务取消。 | 终态 |
| CLOSED | 超时未支付关闭。 | 终态 |
| REFUNDING | 退款处理中。 | REFUNDED / PAID |
| REFUNDED | 已退款。 | 终态 |

## 8.2 支付状态

| 状态 | 含义 | 说明 |
| --- | --- | --- |
| INIT | 支付单已创建。 | 尚未请求渠道或刚创建。 |
| PAYING | 已发起支付。 | 等待用户支付或渠道确认。 |
| SUCCEEDED | 支付成功。 | 只能由渠道回调或主动查询确认。 |
| FAILED | 支付失败。 | 可根据业务允许重新支付。 |
| CLOSED | 支付关闭。 | 订单取消或超时关闭。 |
| REFUNDING | 退款中。 | 等待渠道退款确认。 |
| REFUNDED | 已退款。 | 退款完成。 |

## 8.3 库存预占状态

| 状态 | 含义 | 说明 |
| --- | --- | --- |
| RESERVED | 库存已预占。 | 订单待支付。 |
| CONFIRMED | 预占转已售。 | 支付成功后确认。 |
| RELEASED | 预占释放。 | 订单取消或超时关闭。 |
| EXPIRED | 预占过期。 | 后台任务释放。 |

## 8.4 通知任务状态

| 状态 | 含义 | 说明 |
| --- | --- | --- |
| PENDING | 待发送。 | 任务刚创建。 |
| SENDING | 发送中。 | Worker 正在处理。 |
| SUCCEEDED | 发送成功。 | 终态。 |
| FAILED_RETRYABLE | 失败可重试。 | 达到最大次数前继续重试。 |
| FAILED_FINAL | 最终失败。 | 需要人工或补偿。 |

# 9. 核心业务流程

## 9.1 创建订单

1. 客户端传入商品、数量、收货信息和 `Idempotency-Key`。
2. Handler 校验参数后调用 `OrderService.CreateOrder`。
3. Service 查询 SKU、价格、库存等基础信息。
4. 进入本地事务：创建订单、创建订单明细、预占库存、写库存流水、写 Outbox 事件。
5. 事务提交后返回订单号和待支付金额。
6. 异步 Worker 发布 `order.created` 事件，触发通知或风控等非核心流程。

```go
func (s *OrderService) CreateOrder(ctx context.Context, cmd CreateOrderCommand) (*OrderDTO, error) {
    idemKey := idempotency.Key("order.create", cmd.UserID, cmd.ClientOrderNo)

    return s.idem.Do(ctx, idemKey, func(ctx context.Context) (*OrderDTO, error) {
        var result *OrderDTO
        err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
            order, err := s.buildOrder(ctx, cmd)
            if err != nil { return err }

            if err := s.orderRepo.Create(ctx, order); err != nil { return err }
            if err := s.inventorySvc.Reserve(ctx, ReserveCommand{OrderNo: order.OrderNo}); err != nil { return err }
            if err := s.outbox.Append(ctx, NewOrderCreatedEvent(order)); err != nil { return err }

            result = ToOrderDTO(order)
            return nil
        })
        return result, err
    })
}
```

## 9.2 发起支付

1. 客户端请求支付，传入订单号和支付渠道。
2. OrderService 校验订单归属、金额、状态必须为 `PENDING_PAYMENT`。
3. PaymentService 创建或复用支付单。
4. 调用支付渠道适配器创建预支付参数。
5. 保存渠道请求和响应的脱敏信息。
6. 返回支付参数给客户端。

## 9.3 支付回调

1. 支付网关回调 `/payments/{channel}/callback`。
2. Payment Handler 读取原始 body 并验签。
3. 回调日志先落库，利用唯一索引防重复。
4. 查询支付单和订单，校验金额、币种、状态。
5. 进入本地事务：支付单置为 `SUCCEEDED`，订单置为 `PAID`，库存预占转确认，写库存流水，写 Outbox 事件。
6. 事务提交后立即向支付渠道返回成功。
7. 后续通知由 Outbox/MQ 异步处理。

> **注意：** 支付回调的外部返回必须尽快、稳定、幂等。不要在回调同步发送短信、邮件、Webhook 或做耗时统计。

## 9.4 取消订单

1. 只有 `CREATED` 或 `PENDING_PAYMENT` 状态允许取消。
2. 进入事务：订单置为 `CANCELLED`，支付单关闭，库存预占释放，写库存流水，写 Outbox 事件。
3. 如果支付渠道已创建支付单，事务外异步调用渠道关闭支付，失败则进入补偿任务。
4. 重复取消应返回当前状态，不应报内部错误。

## 9.5 退款

1. 校验订单状态、可退金额和售后规则。
2. 创建退款单，状态为 `REFUNDING`。
3. 调用支付渠道发起退款。
4. 渠道回调或主动查询确认退款结果。
5. 退款成功后更新退款单、支付单、订单状态，并按业务规则处理库存。

## 9.6 通知发送

1. 订单支付成功后写出 `order.paid` 事件。
2. 通知 Worker 消费事件，创建 notification_task。
3. 根据用户偏好、模板、渠道能力生成通知内容。
4. 调用短信、邮件、站内信或 Webhook Client。
5. 失败时按指数退避重试，超过次数后进入 `FAILED_FINAL`，记录失败原因。

# 10. 一致性、幂等与并发控制

## 10.1 一致性策略

| 场景 | 策略 |
| --- | --- |
| 订单创建 + 库存预占 | 同库时使用本地事务；跨库时使用库存服务接口 + 预占记录 + 补偿。 |
| 支付成功 + 订单变更 | 回调先落库，再在事务内更新支付、订单和库存状态。 |
| 订单状态变更 + 通知 | 业务事务内写 Outbox，通知异步消费。 |
| 消息发布失败 | Outbox 保留事件，Worker 重试发布。 |
| Worker 消费失败 | 消费记录 + 重试次数 + 死信/人工处理。 |

## 10.2 幂等原则

- 任何外部可重试入口必须幂等：创建订单、发起支付、支付回调、取消订单、退款回调、通知 Webhook。
- 幂等不能只靠缓存，必须有数据库唯一约束兜底。
- 幂等记录要保存请求摘要和响应摘要，重复请求可以返回一致结果。
- 幂等键应有 scope，避免不同业务共用一个 key 造成误判。

## 10.3 库存扣减并发控制

库存扣减推荐使用条件更新，确保库存不会扣成负数。

```sql
UPDATE inventories
SET available_qty = available_qty - $1,
    reserved_qty = reserved_qty + $1,
    version = version + 1,
    updated_at = now()
WHERE sku_id = $2
  AND warehouse_id = $3
  AND available_qty >= $1;
```

更新影响行数为 0 时，说明库存不足或版本冲突，应返回明确业务错误 `INVENTORY_NOT_ENOUGH`。

# 11. API 设计

## 11.1 统一响应

```json
{
  "code": "OK",
  "message": "success",
  "data": {},
  "request_id": "req_01H..."
}
```

## 11.2 核心接口清单

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | /api/v1/orders | 创建订单，要求 Idempotency-Key。 |
| GET | /api/v1/orders/{order_no} | 查询订单详情。 |
| POST | /api/v1/orders/{order_no}/cancel | 取消订单。 |
| POST | /api/v1/orders/{order_no}/payments | 发起支付。 |
| POST | /api/v1/payments/{channel}/callback | 支付渠道回调。 |
| POST | /api/v1/refunds | 发起退款。 |
| POST | /api/v1/refunds/{channel}/callback | 退款渠道回调。 |
| GET | /api/v1/inventories/{sku_id} | 查询库存。 |
| POST | /api/v1/notifications/test | 测试通知模板和渠道。 |
| GET | /healthz | 进程健康检查。 |
| GET | /readyz | 依赖健康检查。 |
| GET | /metrics | 指标暴露。 |

## 11.3 API 错误码示例

| 错误码 | HTTP | 说明 |
| --- | --- | --- |
| INVALID_ARGUMENT | 400 | 参数错误。 |
| UNAUTHORIZED | 401 | 未认证。 |
| FORBIDDEN | 403 | 无权限。 |
| ORDER_NOT_FOUND | 404 | 订单不存在。 |
| ORDER_STATUS_INVALID | 409 | 订单状态不允许当前操作。 |
| INVENTORY_NOT_ENOUGH | 409 | 库存不足。 |
| PAYMENT_SIGNATURE_INVALID | 400 | 支付回调验签失败。 |
| PAYMENT_AMOUNT_MISMATCH | 409 | 支付金额不匹配。 |
| IDEMPOTENCY_CONFLICT | 409 | 同一幂等键请求内容不一致。 |
| EXTERNAL_SERVICE_ERROR | 502 | 第三方服务异常。 |
| INTERNAL_ERROR | 500 | 系统内部错误。 |

# 12. 事件与消息设计

## 12.1 事件命名

事件名称采用 `domain.action.v1` 格式，版本号用于兼容演进。事件内容必须包含 event_id、occurred_at、trace_id、aggregate_id。

| 事件 | 触发时机 | 消费者 |
| --- | --- | --- |
| order.created.v1 | 订单创建成功。 | 通知、风控、数据统计。 |
| order.cancelled.v1 | 订单取消成功。 | 通知、库存审计。 |
| payment.succeeded.v1 | 支付成功确认。 | 订单、通知、财务。 |
| inventory.reserved.v1 | 库存预占成功。 | 订单审计。 |
| inventory.released.v1 | 库存释放。 | 数据统计。 |
| notification.failed.v1 | 通知最终失败。 | 告警、人工处理。 |

## 12.2 事件示例

```json
{
  "event_id": "evt_01HXYZ",
  "event_type": "order.paid.v1",
  "aggregate_type": "order",
  "aggregate_id": "ord_202605100001",
  "occurred_at": "2026-05-10T10:00:00Z",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "payload": {
    "order_no": "ord_202605100001",
    "user_id": 123,
    "amount_payable": 9900,
    "currency": "CNY"
  }
}
```

# 13. 可观测性与排障设计

## 13.1 日志字段规范

| 字段 | 说明 |
| --- | --- |
| time | 日志时间。 |
| level | DEBUG / INFO / WARN / ERROR。 |
| msg | 稳定、可搜索的英文或中文短句。 |
| request_id | 请求唯一 ID。 |
| trace_id / span_id | 链路追踪 ID。 |
| user_id | 当前用户。 |
| order_no | 订单号。 |
| payment_no | 支付单号。 |
| event_id | 事件 ID。 |
| latency_ms | 耗时。 |
| error_code | 业务错误码。 |
| error | 错误摘要，不输出敏感信息。 |

## 13.2 指标设计

| 指标 | 类型 | 说明 |
| --- | --- | --- |
| http_requests_total | counter | HTTP 请求数，按 path/status/method 聚合。 |
| http_request_duration_seconds | histogram | HTTP 请求耗时。 |
| order_created_total | counter | 订单创建数量。 |
| payment_callback_total | counter | 支付回调数量，按 channel/status 聚合。 |
| inventory_reserve_fail_total | counter | 库存预占失败数。 |
| notification_send_total | counter | 通知发送数量，按 channel/status 聚合。 |
| outbox_pending_total | gauge | 待发布事件积压数量。 |
| worker_retry_total | counter | Worker 重试次数。 |

## 13.3 排障入口

- 用户反馈订单异常：先查 `request_id` 或 `order_no`，再查订单状态、支付单状态、库存预占、Outbox 事件。
- 支付成功但订单未变更：查支付回调日志是否落库，回调验签是否失败，事务是否失败，Outbox 是否积压。
- 库存不一致：查 stock_flows，按 sku_id、order_no、reservation_no 回放库存变更。
- 通知未收到：查 notification_tasks 的状态、渠道响应、重试次数和失败原因。
- 接口慢：通过 trace 查看 Handler、Service、DB、Redis、外部服务耗时。

## 13.4 日志示例

```json
{
  "time": "2026-05-10T10:00:00.123Z",
  "level": "INFO",
  "msg": "payment callback processed",
  "request_id": "req_01HXYZ",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "channel": "wechat",
  "order_no": "ord_202605100001",
  "payment_no": "pay_202605100001",
  "status": "SUCCEEDED",
  "latency_ms": 42
}
```

# 14. 安全设计

- 所有写接口必须认证，后台接口必须鉴权。
- 支付回调必须使用原始 body 验签，禁止先解析再验签。
- 支付金额、币种、订单号必须与本地记录比对。
- 敏感字段必须脱敏：token、密码、手机号、邮箱、身份证、银行卡、渠道密钥。
- 后台操作必须写审计日志，包括操作者、操作对象、前后状态、原因。
- 外部 Webhook 必须校验签名、时间戳和重放窗口。
- 接口必须设置超时、body size limit、限流策略。

# 15. 测试策略

## 15.1 测试分层

| 测试类型 | 覆盖内容 |
| --- | --- |
| 单元测试 | 状态机、金额计算、幂等键生成、错误映射。 |
| Repository 测试 | SQL 查询、唯一约束、事务回滚、库存条件更新。 |
| Service 测试 | 创建订单、支付回调、取消订单、退款、通知任务。 |
| HTTP 测试 | 参数校验、认证鉴权、响应格式、错误码。 |
| 集成测试 | 真实数据库、Redis、MQ 或 Outbox Worker。 |
| 回归测试 | 支付回调重复、库存不足、订单超时关闭、Worker 重试。 |

## 15.2 必测场景

- 重复提交创建订单，只生成一个订单。
- 库存不足时订单创建失败，库存不变。
- 支付回调重复，只处理一次。
- 支付金额不匹配，不更新订单为已支付。
- 订单取消释放库存，重复取消不破坏数据。
- Outbox 发布失败后可重试，不丢事件。
- 通知发送失败后按策略重试，最终失败可查询。

# 16. 上线与运维要求

## 16.1 上线前检查

- 数据库 migration 已评审，可回滚或有修复方案。
- OpenAPI 文档与实现一致。
- 所有核心接口有超时、日志、Trace 和错误码。
- 支付回调、退款回调在测试环境完成全链路验证。
- 告警规则已配置：错误率、延迟、Outbox 积压、Worker 失败、支付回调失败。
- 准备好回滚方案和数据修复脚本。

## 16.2 健康检查

| 接口 | 含义 | 检查内容 |
| --- | --- | --- |
| /healthz | 进程存活。 | 只检查应用是否运行。 |
| /readyz | 服务是否可接流量。 | 检查数据库、Redis、MQ、关键依赖。 |
| /metrics | 指标抓取。 | 暴露 Prometheus 格式指标。 |

# 17. 演进路线

1. 第一阶段：模块化单体，完成订单、支付、库存、通知核心闭环。
2. 第二阶段：引入 Outbox Worker、通知异步化、支付补偿任务、库存对账。
3. 第三阶段：将通知、支付网关适配、库存服务按边界拆分为独立服务。
4. 第四阶段：引入更完整的风控、审计、数据仓库、财务对账和运营后台。

> **结论：** 不要以技术拆分代替业务边界。真正该优先做的是状态机、幂等、事务边界、事件边界和可观测性。
