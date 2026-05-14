你这个思路是对的：**从一开始把地基打好，不是把所有框架都堆上去，而是把“项目边界、初始化、错误、日志、配置、数据库、鉴权、可观测性、测试、工程规范”先定好。**

我建议你从一开始用这套：

```text
Gin
+ sqlc/pgx 或 GORM
+ slog 或 zap
+ Viper
+ 手写初始化 / Fx
+ OpenTelemetry
+ goose 或 golang-migrate
+ OpenAPI
+ golangci-lint
+ Testcontainers
```

但注意：**同一层能力不要同时上两个替代品。** 比如 `GORM/sqlc` 二选一，`zap/slog` 二选一，`wire/fx/手写初始化` 三选一。

---

# 1. 推荐技术选型

## Web 层：Gin

用 **Gin** 负责 HTTP 路由、中间件、请求绑定、参数校验、REST API。Gin 官方定位就是高性能 Go HTTP Web 框架，适合 REST API、Web 应用和微服务场景；Gin 的绑定和验证也直接支持 JSON、XML、YAML、表单等请求体，并使用 `go-playground/validator/v10` 做校验。([Gin Web Framework][1])

推荐使用：

```text
Gin
+ 自己封装 middleware
+ 自己封装 response/error
+ 自己封装 validation 错误返回
```

Gin 只做“入口层”，不要把业务逻辑写在 handler 里。

---

## 数据库层：优先 sqlc + pgx，或者 GORM

如果你想把地基打稳，我更推荐：

```text
PostgreSQL/MySQL
+ sqlc
+ pgx / database/sql
+ goose 或 golang-migrate
```

`sqlc` 的核心方式是：你自己写 SQL，它根据 SQL 生成类型安全的 Go 代码；这样 SQL 可控、类型清晰、适合长期维护。([sqlc.dev][2])

如果用 PostgreSQL，`pgx` 是 Go 的 PostgreSQL driver/toolkit，支持标准 `database/sql` 适配，也提供 PostgreSQL 特有能力。([GitHub][3])

GORM 适合后台系统、CRUD 多、开发速度优先的项目；它是完整 ORM，支持关联、事务、迁移、预加载、Hook 等能力。([GORM][4])

我的建议：

```text
严肃业务系统 / 想练扎实：sqlc + pgx
快速后台 / 管理系统：GORM
复杂项目：核心交易链路用 sqlc，简单后台模块可用 GORM，但要分包隔离
```

---

## 数据库迁移：goose 或 golang-migrate

这个必须从第一天就加。

不要手动改数据库结构，不要直接在线上执行零散 SQL。

推荐二选一：

```text
goose
或
golang-migrate
```

`golang-migrate` 支持 CLI 和 Go library，会按顺序把 migration 应用到数据库。`goose` 也支持 CLI/library，可以用 SQL migration 或 Go migration 管理数据库 schema。([GitHub][5])

目录可以这样：

```text
migrations/
  000001_create_users.up.sql
  000001_create_users.down.sql
  000002_create_orders.up.sql
  000002_create_orders.down.sql
```

---

## 日志：新项目优先 slog，高性能场景用 zap

Go 标准库的 `log/slog` 提供结构化日志，日志记录包含 message、level 和 key-value attributes。([Go Packages][6])

`zap` 是 Uber 的高性能结构化分级日志库，官方文档强调它适合日志在 hot path 的场景，使用较少反射和分配来降低开销。([Go Packages][7])

推荐：

```text
普通新项目：slog
高性能服务 / 公司已有 zap 基建：zap
```

不要业务代码里到处直接 `slog.Info()` 或 `zap.L()`，最好封一层 logger interface 或统一初始化，让日志字段规范一致。

---

## 配置：Viper + 强类型 Config

`Viper` 是 Go 的配置解决方案，支持默认值、显式设置、配置文件、环境变量、远程配置等。([GitHub][8])

推荐做法：

```text
config.yaml
.env
环境变量
↓
Viper 读取
↓
解析到强类型 Config struct
↓
传给各模块
```

不要在业务代码里到处 `viper.GetString()`，这样会导致配置依赖散落全项目。

应该这样：

```go
type Config struct {
    Server ServerConfig
    DB     DBConfig
    Redis  RedisConfig
    JWT    JWTConfig
    Log    LogConfig
}
```

---

## 依赖注入：优先手写初始化；复杂项目用 Fx；不建议新项目强依赖 Wire

这点很重要。

以前很多人推荐 `wire`，但 Google Wire 仓库已经在 **2025-08-25** 被归档，并且 README 标注 “This project is no longer maintained”。Wire 的思想仍然好：通过代码生成连接组件，避免运行时反射和全局变量；但新项目不建议强依赖它。([GitHub][9])

推荐：

```text
小中型项目：手写初始化
复杂生命周期项目：Fx
不推荐新项目：Wire
```

`Fx` 是 Uber 的 Go 依赖注入和应用生命周期框架，支持 startup/shutdown hooks，适合需要管理多个长期运行组件的服务。([Uber Go][10])

我个人建议你一开始这样：

```text
先手写 bootstrap
等模块很多、启动/关闭流程复杂，再考虑 Fx
```

---

## 可观测性：OpenTelemetry 从一开始接入

OpenTelemetry 是可观测性标准。Go 里需要初始化 OpenTelemetry SDK，然后用 API 对代码进行 instrumentation，应用会输出 traces、metrics 等 telemetry 数据。([OpenTelemetry][11])

Gin 有官方生态里的 `otelgin` instrumentation，可以用 middleware 追踪进入 Gin 路由的请求。([Go Packages][12])

从第一天接入的好处是：

```text
request_id / trace_id
日志关联
接口耗时
数据库耗时
外部请求耗时
错误链路
```

这些东西后面补会很痛。

---

## API 文档：OpenAPI

OpenAPI 是描述 HTTP API 的标准，可以让人和工具理解 API 能力，也可以生成客户端代码、测试和文档。([OpenAPI Initiative][13])

推荐：

```text
先写 OpenAPI spec
或
从代码生成 OpenAPI
```

Go 里可以考虑：

```text
oapi-codegen
swaggo
```

`oapi-codegen` 可以根据 OpenAPI 3 spec 生成 Go server/client/model 代码。([GitHub][14])

---

## 测试与质量：testing + Testcontainers + golangci-lint

Go 官方 Wiki 推荐表驱动测试，因为一个测试表可以覆盖多组输入和预期输出。([Go][15])

`Testcontainers for Go` 可以在测试里启动和清理容器依赖，适合做数据库、Redis、MQ 的集成测试。([Go Testcontainers][16])

`golangci-lint` 是 Go 常用 lint runner，支持并行运行、缓存、YAML 配置、IDE/CI 集成和大量 linters。([GitHub][17])

---

# 2. 最推荐的初始组合

如果你想从一开始就稳，我建议用这个：

```text
Gin
+ PostgreSQL
+ sqlc
+ pgx
+ goose
+ slog
+ Viper
+ OpenTelemetry
+ 手写初始化
+ OpenAPI
+ golangci-lint
+ Testcontainers
```

也就是：

```text
HTTP：Gin
数据库访问：sqlc + pgx
数据库迁移：goose
日志：slog
配置：Viper
链路追踪：OpenTelemetry
依赖管理：手写 bootstrap
接口文档：OpenAPI
测试：testing + Testcontainers
代码质量：golangci-lint
```

GORM 不是不能用，但如果你的目标是“把地基打扎实”，我更建议先学会自己写 SQL，再用 sqlc 生成类型安全代码。

---

# 3. 你需要开发的底层功能

下面这些才是真正的“后端地基”。

---

## 第一层：项目启动基建

你需要封装这些：

```text
配置加载
日志初始化
数据库连接
Redis 连接
HTTP Server 初始化
路由注册
中间件注册
优雅关闭
依赖组装
```

启动流程建议固定成这样：

```text
main.go
↓
LoadConfig()
↓
NewLogger()
↓
NewDB()
↓
NewRedis()
↓
NewRepositories()
↓
NewServices()
↓
NewHandlers()
↓
NewRouter()
↓
Start HTTP Server
↓
Graceful Shutdown
```

Go 的 `context.Context` 用来在 API 边界之间传递 deadline、取消信号和请求级值；进入请求时创建 context，调用下游服务时继续传递 context，这是所有 IO 操作的基础。([Go Packages][18])

---

## 第二层：HTTP 基础能力

你要开发这些 middleware：

```text
Recovery middleware
Request ID middleware
Access log middleware
Timeout middleware
CORS middleware
Auth middleware
Rate limit middleware
Trace middleware
Body size limit middleware
```

其中最重要的是：

```text
Recovery
Request ID
Access Log
Timeout
Auth
Trace
```

每个请求都应该带这些字段：

```text
request_id
trace_id
method
path
status
latency
client_ip
user_agent
user_id
error
```

这样线上出问题时，你能通过一条日志追到完整链路。

---

## 第三层：统一响应格式

建议所有 HTTP API 返回统一结构。

例如：

```json
{
  "code": "OK",
  "message": "success",
  "data": {},
  "request_id": "req_xxx"
}
```

错误时：

```json
{
  "code": "USER_NOT_FOUND",
  "message": "用户不存在",
  "request_id": "req_xxx"
}
```

你需要封装：

```text
Success(c, data)
Fail(c, err)
BindAndValidate(c, req)
Pagination response
```

不要在 handler 里到处手写：

```go
c.JSON(200, gin.H{...})
```

而是统一走响应层。

---

## 第四层：统一错误模型

这是非常核心的地基。

你需要设计自己的业务错误：

```go
type AppError struct {
    Code       string
    Message    string
    HTTPStatus int
    Cause      error
}
```

错误应该分层：

```text
参数错误
认证错误
权限错误
资源不存在
业务冲突
限流错误
外部服务错误
数据库错误
系统内部错误
```

例如：

```text
INVALID_ARGUMENT
UNAUTHORIZED
FORBIDDEN
NOT_FOUND
CONFLICT
RATE_LIMITED
EXTERNAL_ERROR
INTERNAL_ERROR
```

数据库错误、第三方错误、panic 错误，最后都应该被转换成统一 API 错误返回。

---

## 第五层：数据库基础能力

数据库层至少要有这些：

```text
连接池配置
migration
事务封装
Repository 规范
SQL 超时控制
慢查询日志
分页查询规范
ID 生成策略
created_at / updated_at
软删除策略，按需
```

重点是事务封装。

建议定义一个 transaction manager：

```go
type TxManager interface {
    WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
```

业务层这样用：

```go
err := txManager.WithinTx(ctx, func(ctx context.Context) error {
    // 创建订单
    // 扣库存
    // 写流水
    return nil
})
```

不要在 service 里直接到处写：

```go
tx, err := db.Begin()
```

否则项目大了以后事务会很乱。

---

## 第六层：鉴权与权限

如果你的系统有用户登录，从一开始就要设计：

```text
用户表
角色表
权限表
登录
注册
密码加密
JWT / Session
Refresh Token
Auth middleware
RBAC 权限判断
登录态失效
审计日志
```

最低配可以先做：

```text
users
roles
user_roles
permissions
role_permissions
```

权限可以先简单一点：

```text
admin
user
guest
```

后面再演进到：

```text
resource + action
```

例如：

```text
order:read
order:create
order:cancel
user:disable
```

---

## 第七层：配置与环境管理

建议从第一天区分：

```text
local
dev
test
staging
prod
```

配置文件可以这样：

```text
configs/
  config.local.yaml
  config.dev.yaml
  config.prod.yaml
```

环境变量用于覆盖敏感信息：

```text
DB_PASSWORD
JWT_SECRET
REDIS_PASSWORD
OTEL_EXPORTER_ENDPOINT
```

不要把密钥写进 Git。

配置模块需要做：

```text
默认值
必填校验
类型解析
环境变量覆盖
启动时打印非敏感配置
敏感字段脱敏
```

---

## 第八层：日志基础能力

日志不是简单打印字符串。

你需要规定日志格式：

```json
{
  "time": "...",
  "level": "INFO",
  "msg": "request completed",
  "request_id": "...",
  "trace_id": "...",
  "method": "GET",
  "path": "/api/v1/users",
  "status": 200,
  "latency_ms": 12,
  "user_id": 123
}
```

需要封装：

```text
业务日志
请求日志
错误日志
慢查询日志
第三方调用日志
panic 日志
敏感字段脱敏
```

尤其注意不要打印：

```text
密码
token
银行卡
身份证
手机号完整值
邮箱完整值
```

---

## 第九层：可观测性

OpenTelemetry 建议从第一天接入：

```text
trace
metrics
logs correlation
```

你至少需要：

```text
HTTP 请求 trace
DB 查询 trace
Redis trace
外部 HTTP 调用 trace
错误 trace
trace_id 写入日志
```

同时加这些接口：

```text
GET /healthz
GET /readyz
GET /metrics
```

区别：

```text
/healthz：进程活着即可
/readyz：数据库、Redis 等依赖正常，服务可以接流量
/metrics：暴露指标
```

---

## 第十层：缓存基础能力

如果用 Redis，不要直接在业务里乱拼 key。

要封装：

```text
Redis Client
Key Builder
TTL 规范
缓存穿透处理
缓存击穿处理
缓存雪崩处理
分布式锁，按需
```

key 命名建议：

```text
app:{env}:user:{id}
app:{env}:order:{id}
app:{env}:rate_limit:{user_id}
```

缓存层建议统一封装：

```go
type Cache interface {
    Get(ctx context.Context, key string, dest any) error
    Set(ctx context.Context, key string, value any, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
}
```

---

## 第十一层：外部服务调用基础能力

调用第三方 HTTP API 时，不要直接到处 `http.Get()`。

要封装统一 client：

```text
timeout
retry
circuit breaker，按需
request log
response log
trace propagation
error mapping
```

每个外部服务单独一个 client：

```text
payment.Client
sms.Client
email.Client
storage.Client
```

业务层不要知道 HTTP 细节。

---

## 第十二层：异步任务与消息

一开始可以不接 MQ，但要预留边界。

后面常见能力：

```text
Job worker
消息队列
延迟任务
重试
死信队列
Outbox pattern
幂等消费
```

如果业务里有支付、订单、库存、通知，建议很早就考虑：

```text
幂等键
事件表
消息状态表
重试次数
失败原因
```

---

## 第十三层：幂等与防重复提交

这个在真实项目很重要。

例如：

```text
创建订单
支付回调
发优惠券
扣库存
提交表单
```

都需要幂等。

你可以做：

```text
Idempotency-Key 请求头
业务唯一索引
请求记录表
Redis 短期锁
状态机校验
```

例如：

```text
POST /orders
Idempotency-Key: xxx
```

同一个 key 重复请求，返回同一个结果。

---

## 第十四层：分页、排序、过滤规范

不要每个接口各写各的分页。

统一成：

```text
page
page_size
sort
order
```

或游标分页：

```text
cursor
limit
```

统一响应：

```json
{
  "items": [],
  "page": 1,
  "page_size": 20,
  "total": 100
}
```

大数据量接口优先考虑 cursor pagination。

---

## 第十五层：文件上传与存储，按需

如果项目涉及头像、附件、图片、视频，需要抽象 storage：

```text
LocalStorage
S3Storage
OSSStorage
COSStorage
MinIOStorage
```

业务层只依赖：

```go
type Storage interface {
    Put(ctx context.Context, key string, r io.Reader) error
    GetURL(ctx context.Context, key string) (string, error)
    Delete(ctx context.Context, key string) error
}
```

不要业务代码直接依赖某个云厂商 SDK。

---

## 第十六层：测试基建

从第一天就建立：

```text
单元测试
集成测试
HTTP handler 测试
Repository 测试
Service 测试
Mock 外部服务
测试数据库
测试数据工厂
```

目录可以这样：

```text
internal/user/service_test.go
internal/user/repository_test.go
test/integration/user_test.go
test/fixtures/
```

推荐：

```text
testing
testify
gomock/mockery，按需
Testcontainers
```

数据库相关测试不要只 mock，关键链路要跑真实数据库容器。

---

## 第十七层：CI/CD 与本地开发命令

你需要一个 `Makefile`：

```makefile
run:
	go run ./cmd/api

test:
	go test ./...

lint:
	golangci-lint run

fmt:
	go fmt ./...

generate:
	sqlc generate

migrate-up:
	goose up

migrate-down:
	goose down
```

还需要：

```text
Dockerfile
docker-compose.yaml
.env.example
.golangci.yml
README.md
```

本地一键启动：

```bash
docker compose up -d
make migrate-up
make run
```

---

# 4. 推荐项目目录结构

可以这样开局：

```text
your-app/
  cmd/
    api/
      main.go

  configs/
    config.local.yaml
    config.dev.yaml
    config.prod.yaml

  migrations/
    000001_create_users.up.sql
    000001_create_users.down.sql

  api/
    openapi.yaml

  internal/
    bootstrap/
      app.go
      server.go

    config/
      config.go

    logger/
      logger.go

    observability/
      otel.go
      metrics.go

    database/
      db.go
      tx.go
      migrate.go

    redis/
      redis.go

    http/
      router.go
      response.go
      error.go
      middleware/
        request_id.go
        recovery.go
        access_log.go
        timeout.go
        auth.go
        cors.go

    auth/
      handler.go
      service.go
      repository.go
      model.go

    user/
      handler.go
      service.go
      repository.go
      model.go
      dto.go

    order/
      handler.go
      service.go
      repository.go
      model.go
      dto.go

    platform/
      email/
      sms/
      storage/
      payment/

  pkg/
    如果真的要给外部项目复用，才放这里

  test/
    integration/
    fixtures/

  sql/
    queries/
      users.sql
      orders.sql
    schema/

  sqlc.yaml
  go.mod
  Makefile
  Dockerfile
  docker-compose.yaml
  .golangci.yml
  .env.example
  README.md
```

重点：

```text
cmd/api：程序入口
internal：业务和基础设施代码
migrations：数据库迁移
sql：sqlc SQL 文件
api/openapi.yaml：API 契约
configs：配置
test：集成测试
```

---

# 5. 分层建议

我建议采用这种分层：

```text
Handler
↓
Service
↓
Repository
↓
Database
```

职责分别是：

```text
Handler：解析请求、校验参数、返回响应
Service：业务逻辑、事务、权限、状态流转
Repository：数据库读写
Database：连接池、事务、底层 DB 能力
```

不要让 handler 直接操作数据库。

不要让 repository 写业务规则。

不要让 model 变成什么都塞的大杂烩。

---

# 6. 你第一阶段必须完成的基础模块

按优先级来。

## P0：第一天就要有

```text
项目目录结构
配置加载
日志初始化
数据库连接
数据库 migration
统一错误
统一响应
Gin router
基础 middleware
优雅关闭
Makefile
Docker Compose
```

## P1：第一个业务接口前要有

```text
参数校验
事务封装
Repository 规范
Service 规范
OpenAPI 文档
健康检查
request_id
trace_id
基础测试
golangci-lint
```

## P2：上线前要有

```text
鉴权
权限
审计日志
OpenTelemetry
metrics
慢查询日志
限流
CORS
安全 header
CI
集成测试
错误告警
```

## P3：业务复杂后再加

```text
Redis 缓存
消息队列
异步任务
分布式锁
Outbox
幂等框架
文件存储抽象
多租户
配置中心
服务发现
```

---

# 7. 不建议一开始做的东西

这些容易过度设计：

```text
自己写 ORM
自己写 DI 框架
自己写日志框架
自己写配置中心
自己写通用 CRUD 生成器
过早上微服务
过早上复杂 DDD
过早搞多租户
过早搞分布式事务
```

你真正应该做的是：

```text
边界清楚
错误统一
日志可查
配置可控
数据库可迁移
请求可追踪
测试能跑
部署可重复
```

这才叫地基好。

---

# 8. 最终建议

你的初始地基可以定成：

```text
Gin
+ sqlc
+ pgx
+ goose
+ slog
+ Viper
+ OpenTelemetry
+ 手写初始化
+ OpenAPI
+ Testcontainers
+ golangci-lint
```

然后自己开发这些底层模块：

```text
config
logger
database
transaction
migration
response
error
middleware
auth
rbac
validation
pagination
cache
observability
healthcheck
external client
testing
```

最核心的一句话：

```text
框架只是工具，真正的地基是你自己封装出来的一套工程规范。
```

第一版不要追求“大而全”，但必须做到：**统一错误、统一响应、统一日志、统一配置、统一数据库事务、统一中间件、统一测试方式。**

[1]: https://gin-gonic.com/en/docs/?utm_source=chatgpt.com "Documentation | Gin Web Framework"
[2]: https://sqlc.dev/?utm_source=chatgpt.com "Compile SQL to type-safe code | sqlc.dev"
[3]: https://github.com/jackc/pgx?utm_source=chatgpt.com "GitHub - jackc/pgx: PostgreSQL driver and toolkit for Go"
[4]: https://gorm.io/zh_CN/docs/index.html?utm_source=chatgpt.com "GORM 指南 | GORM - The fantastic ORM library for Golang ..."
[5]: https://github.com/golang-migrate/migrate?utm_source=chatgpt.com "golang-migrate/migrate: Database migrations. CLI and Golang library."
[6]: https://pkg.go.dev/log/slog?utm_source=chatgpt.com "slog package - log/slog - Go Packages"
[7]: https://pkg.go.dev/go.uber.org/zap?utm_source=chatgpt.com "zap package - go.uber.org/zap - Go Packages"
[8]: https://github.com/spf13/viper?utm_source=chatgpt.com "GitHub - spf13/viper: Go configuration with fangs"
[9]: https://github.com/google/wire "GitHub - google/wire: Compile-time Dependency Injection for Go · GitHub"
[10]: https://uber-go.github.io/fx/lifecycle.html?utm_source=chatgpt.com "Lifecycle - Fx - GitHub Pages"
[11]: https://opentelemetry.io/docs/languages/go/instrumentation/?utm_source=chatgpt.com "Instrumentation - OpenTelemetry"
[12]: https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin?utm_source=chatgpt.com "otelgin package - go.opentelemetry.io/contrib/instrumentation/github ..."
[13]: https://www.openapis.org/?utm_source=chatgpt.com "OpenAPI Initiative – The OpenAPI Initiative provides an open source ..."
[14]: https://github.com/oapi-codegen/oapi-codegen/?utm_source=chatgpt.com "GitHub - oapi-codegen/oapi-codegen: Generate Go client and server ..."
[15]: https://go.dev/wiki/TableDrivenTests?utm_source=chatgpt.com "Go Wiki: TableDrivenTests - The Go Programming Language"
[16]: https://golang.testcontainers.org/?utm_source=chatgpt.com "Testcontainers for Go"
[17]: https://github.com/golangci/golangci-lint?utm_source=chatgpt.com "GitHub - golangci/golangci-lint: Fast linters runner for Go"
[18]: https://pkg.go.dev/context?utm_source=chatgpt.com "context package - context - Go Packages"
