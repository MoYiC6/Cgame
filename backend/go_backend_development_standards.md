# Go 后端开发规范与可复用组件规范

面向可维护、可测试、可排障、可扩展的生产级工程规范

版本：1.0  
日期：2026-05-10

# 1. 规范目标

本规范用于约束 Go 后端项目的代码结构、模块边界、错误处理、日志追踪、数据库访问、测试、CI 和排障方式。目标是让团队成员写出的代码风格一致、组件可复用、业务逻辑清晰、线上问题可以快速定位。

## 1.1 基本原则

- 业务代码必须可读：优先清晰表达业务意图，避免炫技式抽象。
- 组件必须可替换：支付渠道、通知渠道、缓存、消息队列、存储都通过接口隔离。
- 错误必须可解释：任何返回给调用方的错误都应有稳定错误码。
- 日志必须可搜索：关键字段固定，严禁只打印自然语言字符串。
- 测试必须可重复：本地、CI、测试环境结果一致。
- 上线必须可回滚：数据库、配置、代码、消息消费都要考虑回滚和补偿。

## 1.2 强制约定等级

| 等级 | 含义 |
| --- | --- |
| MUST | 必须遵守，不遵守会影响系统可靠性或团队协作。 |
| SHOULD | 原则上遵守，特殊情况需要在 PR 中说明。 |
| MAY | 可选建议，可根据项目复杂度裁剪。 |

# 2. Go 代码风格

## 2.1 格式化与静态检查

- MUST：所有代码提交前执行 `gofmt` 或 `go fmt ./...`。
- MUST：CI 中执行 `go test ./...` 和 `golangci-lint run`。
- MUST：禁止提交未使用代码、未使用变量、调试打印、临时代码。
- SHOULD：复杂函数圈复杂度过高时拆分，优先提取有业务含义的私有函数。

## 2.2 命名规范

| 对象 | 规范 | 示例 |
| --- | --- | --- |
| 包名 | 短小、单数、无下划线。 | `order`、`payment`、`inventory` |
| 接口 | 表达能力，不加 Impl。 | `PaymentGateway`、`OrderRepository` |
| 实现类 | 按实现命名。 | `wechatGateway`、`postgresOrderRepository` |
| DTO | 输入输出明确。 | `CreateOrderRequest`、`OrderResponse` |
| 错误码 | 大写下划线，稳定不随文案变。 | `ORDER_STATUS_INVALID` |
| 事件名 | domain.action.v1。 | `order.paid.v1` |

## 2.3 函数规范

- MUST：函数只做一件事，名字表达意图。
- MUST：有 IO、数据库、Redis、外部调用的函数第一个参数必须是 `context.Context`。
- SHOULD：Service 方法使用命令对象承载输入，避免参数列表过长。
- SHOULD：返回值不要用裸 `map[string]any` 表示业务数据。

```go
// 推荐
func (s *OrderService) CreateOrder(ctx context.Context, cmd CreateOrderCommand) (*OrderDTO, error)

// 不推荐：参数过多、含义不清、难以扩展
func (s *OrderService) CreateOrder(ctx context.Context, userID int64, skuID int64, qty int, address string, coupon string, source string) error
```

# 3. 项目结构规范

## 3.1 标准目录

```text
internal/
  platform/      # 可复用基础组件
  modules/       # 业务模块
  clients/       # 第三方服务适配
cmd/
  api/           # HTTP 服务入口
  worker/        # 异步任务入口
api/             # OpenAPI 契约
migrations/      # 数据库迁移
sql/queries/     # sqlc 查询
test/            # 集成测试和测试夹具
```

## 3.2 包依赖方向

依赖方向必须单向，不能循环引用。业务模块可以依赖 `platform`，但 `platform` 不能依赖任何业务模块。

```text
cmd
 |
 v
bootstrap
 |
 v
modules  --->  clients
 |
 v
platform
```

## 3.3 禁止事项

- MUST NOT：在 `cmd/main.go` 写业务逻辑。
- MUST NOT：在 Handler 直接调用 Repository。
- MUST NOT：在 Repository 中决定订单状态流转。
- MUST NOT：跨模块直接读取其他模块数据表。
- MUST NOT：业务代码直接调用第三方 SDK。

# 4. 分层开发规范

## 4.1 Handler 规范

Handler 只负责 HTTP 语义，不负责业务决策。

- MUST：绑定请求、校验参数、提取认证信息、调用 Service、返回统一响应。
- MUST：所有错误交给统一错误处理器转换。
- MUST NOT：拼 SQL、开启事务、调用第三方支付 SDK。
- SHOULD：Handler 文件保持薄，每个方法控制在可阅读范围内。

```go
func (h *OrderHandler) Create(c *gin.Context) {
    var req CreateOrderRequest
    if err := bind.Validate(c, &req); err != nil {
        response.Fail(c, err)
        return
    }

    userID := auth.MustUserID(c.Request.Context())
    dto, err := h.orderSvc.CreateOrder(c.Request.Context(), req.ToCommand(userID))
    if err != nil {
        response.Fail(c, err)
        return
    }

    response.Success(c, dto)
}
```

## 4.2 Service 规范

Service 是业务规则的核心位置，负责状态机、事务、幂等、权限、跨模块编排。

- MUST：所有业务状态变更在 Service 中表达。
- MUST：需要一致性的多表写入通过 TxManager 完成。
- MUST：外部调用前后写清楚状态，避免悬挂状态。
- SHOULD：复杂业务拆成私有方法，但不要过度抽象成无意义 helper。

## 4.3 Repository 规范

Repository 只表达数据访问，不表达业务意图之外的规则。

- MUST：所有 SQL 通过 sqlc 或明确封装执行。
- MUST：Repository 方法接收 ctx，支持事务上下文。
- MUST：数据库唯一约束错误要映射为业务错误。
- MUST NOT：Repository 中调用 Service、Client 或 MQ。

## 4.4 Client 规范

Client 封装第三方服务协议，如支付渠道、短信服务、邮件服务、对象存储。

- MUST：统一设置 timeout。
- MUST：外部请求带 trace propagation。
- MUST：记录请求摘要、响应摘要、耗时、错误码；敏感字段脱敏。
- SHOULD：按业务错误映射第三方错误，不把 SDK 错误直接抛给上层。

# 5. 可复用组件规范

## 5.1 组件设计要求

| 要求 | 说明 |
| --- | --- |
| 小而稳 | 每个组件解决一个基础问题，不掺业务。 |
| 显式配置 | 组件构造函数接收 Config，禁止在内部读环境变量。 |
| 生命周期明确 | 需要关闭的组件必须暴露 Close 或注册 shutdown hook。 |
| 接口优先 | 业务依赖接口，基础设施提供实现。 |
| 可测试 | 组件应支持 fake/mock，便于单元测试。 |
| 可观测 | 组件内关键操作要有日志、指标或 trace。 |

## 5.2 推荐基础组件清单

| 组件 | 职责 |
| --- | --- |
| config | 配置加载、校验、脱敏打印。 |
| logger | 结构化日志、ctx 字段提取、敏感字段脱敏。 |
| database | 连接池、健康检查、关闭。 |
| transaction | 统一事务管理。 |
| redis/cache | 缓存、TTL、序列化、key 构造。 |
| idempotency | 幂等执行、请求摘要、响应缓存。 |
| lock | 短期分布式锁，必须有超时和 owner。 |
| eventbus | 事件发布接口，可由 Outbox/MQ 实现。 |
| outbox | 本地事务事件表、发布 Worker、重试。 |
| httpx | 外部 HTTP 客户端，统一超时、重试、日志、trace。 |
| response | 统一 API 响应。 |
| errors | 业务错误码、错误包装、错误映射。 |
| pagination | 分页参数和响应。 |
| observability | OpenTelemetry、metrics、trace id 关联。 |

## 5.3 组件 API 设计示例

```go
type Cache interface {
    Get(ctx context.Context, key string, dest any) error
    Set(ctx context.Context, key string, value any, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
}

type KeyBuilder interface {
    User(id int64) string
    Order(orderNo string) string
    RateLimit(scope string, id string) string
}
```

## 5.4 禁止把组件做成“大工具箱”

不要创建一个 `utils` 包承载所有东西。工具函数必须按领域或能力归类，例如 `money`、`clock`、`idgen`、`hashing`、`pagination`。无法明确归类的工具函数通常意味着抽象还不成熟。

# 6. 错误处理规范

## 6.1 错误模型

所有业务错误使用稳定错误码。底层错误使用 `Cause` 包装，返回给客户端时只暴露安全信息。

```go
var ErrInventoryNotEnough = errors.NewAppError(
    "INVENTORY_NOT_ENOUGH",
    "库存不足",
    http.StatusConflict,
)

if affected == 0 {
    return errors.WithFields(ErrInventoryNotEnough, "sku_id", skuID)
}
```

## 6.2 错误处理要求

- MUST：错误码稳定，不能因为文案调整改变。
- MUST：日志中记录 cause，响应中只返回安全 message。
- MUST：不要吞掉错误，除非明确记录并有补偿策略。
- MUST：第三方错误映射为系统错误码，如 `PAYMENT_CHANNEL_TIMEOUT`。
- MUST NOT：使用字符串比较判断错误类型。

## 6.3 panic 使用规范

- MUST NOT：业务流程中使用 panic 表示错误。
- MAY：启动阶段配置缺失、依赖初始化失败可以直接失败退出。
- MUST：HTTP 服务必须有 recovery middleware，并记录 request_id、trace_id、堆栈摘要。

# 7. API 与 HTTP 规范

## 7.1 路由规范

| 动作 | 方法 | 示例 |
| --- | --- | --- |
| 创建资源 | POST | `POST /api/v1/orders` |
| 查询资源 | GET | `GET /api/v1/orders/{order_no}` |
| 更新局部字段 | PATCH | `PATCH /api/v1/orders/{order_no}` |
| 执行业务动作 | POST | `POST /api/v1/orders/{order_no}/cancel` |
| 删除资源 | DELETE | 谨慎使用；多数业务使用状态关闭。 |

## 7.2 请求规范

- MUST：写接口支持 `Idempotency-Key`，至少订单创建、支付、退款接口必须支持。
- MUST：请求体有明确 DTO，不直接暴露数据库 Model。
- MUST：参数校验失败返回 `INVALID_ARGUMENT`，并指出字段。
- SHOULD：金额使用整数最小单位，不使用 float。

## 7.3 响应规范

```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "order_no": "ord_202605100001"
  },
  "request_id": "req_01HXYZ"
}
```

- MUST：所有响应包含 `code`、`message`、`request_id`。
- MUST：错误响应不能泄露 SQL、堆栈、密钥、渠道原始错误。
- SHOULD：分页响应统一使用 `items`、`page`、`page_size`、`total` 或 cursor 格式。

# 8. 数据库与 SQL 规范

## 8.1 Migration 规范

- MUST：所有表结构变更通过 migration 提交。
- MUST：migration 文件名带序号和语义，例如 `000012_add_payment_callbacks.up.sql`。
- MUST：上线前评审 migration 的锁表风险、耗时、回滚方案。
- MUST NOT：手动在线上执行未入库的结构变更 SQL。

## 8.2 SQL 规范

- MUST：核心查询写在 `sql/queries` 并由 sqlc 生成代码。
- MUST：涉及状态更新时，WHERE 条件必须包含当前状态，避免错误流转。
- MUST：库存扣减必须使用条件更新或乐观锁，防止负库存。
- SHOULD：查询只返回需要字段，不用 `SELECT *`。
- SHOULD：复杂查询配合 explain 分析索引。

```sql
-- name: MarkOrderPaid :execrows
UPDATE orders
SET status = 'PAID',
    paid_at = now(),
    updated_at = now()
WHERE order_no = $1
  AND status = 'PENDING_PAYMENT';
```

## 8.3 事务规范

- MUST：一个业务动作内的强一致写入放在同一个事务。
- MUST：事务中避免耗时外部 HTTP 调用。
- MUST：事务中写业务数据和 Outbox 事件，事务外发布消息。
- MUST：事务函数内返回错误即回滚。

## 8.4 索引规范

- MUST：业务唯一标识加唯一索引，如 order_no、payment_no、event_id。
- MUST：幂等键、回调事件 ID、消息消费记录加唯一索引。
- SHOULD：高频查询按 `user_id + created_at`、`status + updated_at` 等组合索引设计。
- SHOULD：低基数字段不单独建索引，结合查询条件设计组合索引。

# 9. 状态机与业务规则规范

## 9.1 状态流转

- MUST：订单、支付、库存预占、通知任务都必须定义状态枚举。
- MUST：状态流转必须集中在 Service 或 Domain 方法中。
- MUST：数据库更新状态时带上原状态条件。
- MUST：非法状态流转返回明确错误码。

```go
func (o *Order) MarkPaid(now time.Time) error {
    if o.Status != OrderStatusPendingPayment {
        return ErrOrderStatusInvalid
    }
    o.Status = OrderStatusPaid
    o.PaidAt = &now
    return nil
}
```

## 9.2 金额规范

- MUST：金额使用整数最小单位，例如分、cent。
- MUST：金额计算集中封装，禁止散落在 Handler 中。
- MUST：支付回调金额必须与本地支付单金额一致。
- MUST NOT：使用 float64 表示金额。

# 10. 幂等、重试与补偿规范

## 10.1 幂等规范

- MUST：外部入口和消息消费都要幂等。
- MUST：幂等使用数据库唯一约束兜底。
- MUST：同一幂等键但请求摘要不同，应返回 `IDEMPOTENCY_CONFLICT`。
- SHOULD：幂等记录保存响应摘要，重复请求直接返回原结果。

## 10.2 重试规范

| 错误类型 | 是否重试 | 说明 |
| --- | --- | --- |
| 参数错误 | 否 | 重试不会成功。 |
| 认证/权限错误 | 否 | 需要修正身份或权限。 |
| 库存不足 | 否 | 业务失败。 |
| 支付渠道超时 | 是 | 需要查询或补偿，不能盲目重复扣款。 |
| 网络瞬断 | 是 | 使用有限次数和退避策略。 |
| 数据库唯一键冲突 | 视情况 | 通常转为幂等返回或业务冲突。 |

## 10.3 补偿任务规范

- MUST：补偿任务要可重复执行。
- MUST：补偿任务记录执行次数、最后错误、下一次执行时间。
- MUST：支付类补偿优先查询渠道状态，再决定本地状态。
- MUST：超过最大重试后进入人工处理列表。

# 11. 消息与 Outbox 规范

## 11.1 事件发布

- MUST：业务事务中只写 Outbox，不直接发布 MQ。
- MUST：事件包含 event_id、event_type、aggregate_id、occurred_at、trace_id。
- MUST：事件 payload 不能放敏感信息。
- SHOULD：事件版本化，例如 `order.paid.v1`。

## 11.2 消息消费

- MUST：消费者按 `event_id + consumer_name` 做幂等。
- MUST：消费失败要记录错误和重试次数。
- MUST：消费逻辑不能依赖消息只投递一次。
- SHOULD：消费者只处理自己关心的字段，忽略未知字段，便于事件演进。

```go
func (c *OrderPaidConsumer) Handle(ctx context.Context, event Event) error {
    return c.consumerIdem.Do(ctx, event.ID, "notification.order_paid", func(ctx context.Context) error {
        return c.notificationSvc.CreatePaymentSuccessTask(ctx, event.Payload)
    })
}
```

# 12. 日志、Trace 与指标规范

## 12.1 日志等级

| 等级 | 使用场景 |
| --- | --- |
| DEBUG | 本地调试、详细变量，生产默认关闭或采样。 |
| INFO | 关键业务节点成功，如订单创建、支付回调处理成功。 |
| WARN | 可恢复异常，如渠道超时后进入补偿、消息消费重试。 |
| ERROR | 请求失败、任务最终失败、不可恢复异常。 |

## 12.2 必须记录的业务日志

- 订单创建成功/失败。
- 库存预占、确认扣减、释放。
- 支付单创建、渠道请求、支付回调验签、支付成功/失败。
- 订单取消、关闭、退款。
- Outbox 发布失败、消息消费失败。
- 通知发送成功、失败、最终失败。

## 12.3 日志字段规范

```go
logger.Info(ctx, "order created",
    field.String("order_no", order.OrderNo),
    field.Int64("user_id", order.UserID),
    field.Int64("amount_payable", order.AmountPayable),
    field.String("request_id", requestid.FromContext(ctx)),
    field.String("trace_id", traceid.FromContext(ctx)),
)
```

## 12.4 Trace 规范

- MUST：HTTP 入口创建 trace。
- MUST：DB、Redis、外部 HTTP 调用要挂到当前 trace。
- MUST：日志中输出 trace_id，方便从日志跳到链路。
- SHOULD：核心业务步骤创建子 span，例如 reserve_inventory、create_payment、publish_outbox。

## 12.5 Metrics 规范

- MUST：HTTP 请求量、错误率、延迟。
- MUST：支付回调成功/失败数量。
- MUST：Outbox 待发布数量和发布失败数量。
- MUST：Worker 消费失败和重试数量。
- SHOULD：按渠道、状态、业务类型打标签，但避免高基数字段如 order_no。

# 13. 配置与密钥规范

## 13.1 配置文件

```text
configs/
  config.local.yaml
  config.dev.yaml
  config.test.yaml
  config.prod.yaml
.env.example
```

## 13.2 规则

- MUST：配置启动时校验，缺失关键配置直接失败退出。
- MUST：密钥通过环境变量或密钥系统注入，禁止提交到 Git。
- MUST：启动日志可以打印配置摘要，但必须脱敏。
- SHOULD：配置结构体按组件分组，如 Server、DB、Redis、Payment。

# 14. 安全规范

## 14.1 认证与授权

- MUST：所有用户写操作都需要认证。
- MUST：后台接口需要角色或权限校验。
- MUST：Service 层再次校验资源归属，不只依赖前端传参。
- SHOULD：权限表达为 `resource:action`，例如 `order:refund`。

## 14.2 支付安全

- MUST：支付回调用原始 body 验签。
- MUST：回调时间戳和 nonce 防重放。
- MUST：金额、币种、订单号、渠道交易号与本地记录比对。
- MUST：渠道密钥只存在配置或密钥系统，不落日志。

## 14.3 敏感信息

- MUST：密码只保存强哈希，不可逆加密也不允许。
- MUST：token、密钥、验证码、银行卡、身份证、手机号、邮箱默认脱敏。
- MUST：生产日志禁止打印完整请求体，除非明确白名单和脱敏。

# 15. 测试规范

## 15.1 测试命名

- MUST：测试文件命名 `xxx_test.go`。
- SHOULD：测试函数命名表达场景，例如 `TestCreateOrder_WhenInventoryNotEnough_ReturnsConflict`。
- SHOULD：使用表驱动测试覆盖多组状态流转。

## 15.2 单元测试要求

- MUST：状态机、金额计算、错误映射、幂等键生成必须有单元测试。
- MUST：Service 复杂业务分支必须覆盖正常、异常、重复请求、非法状态。
- SHOULD：使用 fake repository 测 Service，避免所有测试都依赖数据库。

## 15.3 集成测试要求

- MUST：Repository 使用真实数据库测试。
- MUST：订单创建、支付回调、取消订单、库存扣减跑集成测试。
- SHOULD：使用 Testcontainers 启动数据库和 Redis。
- SHOULD：测试结束清理数据或使用事务回滚。

## 15.4 必测清单

| 模块 | 必须覆盖 |
| --- | --- |
| 订单 | 重复创建、库存不足、取消、超时关闭、非法状态流转。 |
| 支付 | 发起支付、重复回调、验签失败、金额不匹配、渠道超时。 |
| 库存 | 并发预占、释放、确认扣减、流水一致性。 |
| 通知 | 发送成功、失败重试、最终失败、重复消息消费。 |
| Outbox | 事务内写事件、发布失败重试、重复发布不重复消费。 |

# 16. Git、CI 与评审规范

## 16.1 分支与提交

- SHOULD：分支名包含类型和需求号，例如 `feature/order-create`、`fix/payment-callback-idempotency`。
- SHOULD：提交信息表达意图，例如 `order: add idempotent create order`。
- MUST：一个 PR 聚焦一个主题，避免大杂烩。

## 16.2 CI 必跑项

- `go test ./...`
- `golangci-lint run`
- `sqlc generate` 后检查无未提交变更。
- migration 命名检查。
- OpenAPI lint 或文档生成检查。

## 16.3 Code Review 清单

- 是否破坏分层依赖？
- 是否有统一错误码？
- 是否有 request_id、trace_id 和关键业务字段日志？
- 是否有幂等和唯一约束？
- 是否有事务边界？
- 是否会在事务里调用外部服务？
- 是否泄露敏感信息？
- 是否补充必要测试？
- 是否更新 OpenAPI 和文档？

# 17. 文档规范

## 17.1 必须维护的文档

| 文档 | 说明 |
| --- | --- |
| README | 本地启动、依赖、常用命令。 |
| OpenAPI | HTTP 接口契约。 |
| 架构文档 | 模块边界、核心流程、状态机。 |
| 错误码文档 | 错误码、HTTP 状态、说明、处理建议。 |
| 事件文档 | 事件名、payload、生产者、消费者。 |
| Runbook | 常见故障排查步骤。 |

## 17.2 更新规则

- MUST：新增接口时更新 OpenAPI。
- MUST：新增事件时更新事件文档。
- MUST：新增错误码时更新错误码文档。
- MUST：新增关键链路时补充 Runbook。

# 18. 排障规范

## 18.1 通用排障顺序

1. 拿到 `request_id`、`trace_id`、`order_no`、`payment_no` 或 `event_id`。
2. 先查入口日志，确认请求是否到达、参数是否正确、响应错误码是什么。
3. 用 trace 查看慢在哪一段：Handler、Service、DB、Redis、外部服务。
4. 查业务表状态：订单、支付单、库存预占、通知任务。
5. 查 Outbox 和 Worker 消费记录，确认事件是否发布和消费。
6. 查第三方渠道日志或本地 callback log。
7. 确认是否需要补偿、重试或人工修复。

## 18.2 常见故障 Runbook

| 问题 | 优先检查 |
| --- | --- |
| 用户已支付但订单未支付 | 支付回调日志、验签结果、金额比对、支付单状态、订单状态、事务错误。 |
| 库存少了但订单失败 | 库存预占记录、stock_flows、事务是否回滚、补偿任务。 |
| 订单重复 | Idempotency-Key、client_order_no 唯一索引、订单创建日志。 |
| 通知没收到 | notification_tasks 状态、渠道响应、重试次数、模板变量。 |
| 接口突然变慢 | trace、数据库慢查询、Redis 延迟、外部服务耗时、连接池耗尽。 |
| Outbox 积压 | Worker 是否运行、MQ 是否可用、发布错误、重试退避配置。 |

# 19. 性能规范

- MUST：所有外部调用设置超时。
- MUST：数据库连接池配置合理，并暴露指标。
- MUST：高频查询有索引和分页。
- MUST：列表接口限制 page_size 最大值。
- SHOULD：热点缓存设置随机 TTL，避免雪崩。
- SHOULD：批量处理时控制并发，避免打爆数据库或第三方服务。
- MUST NOT：在请求链路中同步执行大量通知发送或批处理任务。

# 20. 发布规范

## 20.1 上线前

- 功能开关准备好，必要时可灰度。
- migration 已评审并在测试环境跑过。
- 核心链路集成测试通过。
- 错误率、延迟、支付失败、Outbox 积压告警已配置。
- 回滚方案明确。

## 20.2 上线后

- 观察 HTTP 错误率和延迟。
- 观察支付回调成功率。
- 观察订单状态分布是否异常。
- 观察 Outbox 积压和 Worker 重试。
- 观察数据库慢查询和连接池指标。

# 21. 推荐 PR 模板

```markdown
## 变更内容
- 

## 影响范围
- API:
- DB:
- 事件:
- 配置:

## 可靠性检查
- [ ] 是否有幂等
- [ ] 是否有事务边界
- [ ] 是否有错误码
- [ ] 是否有日志和 trace 字段
- [ ] 是否有测试
- [ ] 是否更新 OpenAPI/文档

## 上线与回滚
- 上线步骤:
- 回滚方案:
```

# 22. 最小落地版本

为了避免过度设计，第一版可以只实现以下基础能力，然后随着业务复杂度逐步补齐。

| 阶段 | 必须具备 |
| --- | --- |
| MVP | 配置、日志、错误、响应、数据库、migration、事务、订单创建、库存预占、支付回调、基础通知。 |
| 上线前 | 幂等、Outbox、Worker、Trace、Metrics、告警、集成测试、Runbook。 |
| 增长期 | 多支付渠道、多通知渠道、退款、补偿任务、库存对账、运营后台。 |
| 规模化 | 模块拆服务、事件治理、权限系统、审计中心、财务对账、数据分析。 |

> **结论：** 优雅不是抽象越多越好，而是每一层职责清楚、错误能定位、变更有边界、故障能恢复。
