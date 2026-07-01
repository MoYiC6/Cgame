# Go 后端完整架构大纲

> 基于对 Java 后端（~70 表、119 Controller、93 Entity、163 Service）和当前 Go 后端的完整分析。

---

## 一、进程与入口层

```
cmd/
├── api/          # HTTP API 进程
├── worker/       # 后台任务进程（定时器、队列消费）
├── migrate/      # 数据库迁移独立入口
└── cli/          # 运维命令（种子数据、手动对账等）
```

**原则：CQRS 进程分离。** API 进程管请求响应，Worker 进程管后台处理。各自独立扩缩容。

**当前差距：** Worker 骨架有了，缺 migrate/cli 入口；Worker 只有 placeholder。

---

## 二、启动装配层（Bootstrap）

```
internal/bootstrap/
├── app.go          # Dependencies 容器 + Shutdown 聚合
├── server.go       # Gin 引擎创建、路由装配、健康检查
├── http_server.go  # HTTP Server 生命周期（优雅启停）
├── middleware.go    # 全局中间件链
├── worker.go       # Worker 生命周期 + 任务注册
└── module.go       # 模块注册接口
```

**职责：** 零业务逻辑，只做装配。统一管理中间件顺序、优雅关闭、信号处理。

**当前差距：** 基本完整，但缺 `module.go` 统一注册模式（当前用 `...HTTPRouteRegistrar` 散装配）。

---

## 三、基础设施层（Platform）—— 地基核心

```
internal/platform/
├── config/          # 强类型配置加载 + 环境覆盖 + 校验 + hot-reload
├── logger/          # 结构化日志（JSON 输出、级别、context 注入、采样）
├── errors/          # 统一错误模型（code + message + cause + stack + http_status）
├── response/        # 统一 API 响应格式 + 错误 -> HTTP 映射
├── database/        # 连接池 + 事务管理 + sqlc 生成代码 + 健康检查
├── redis/           # Redis 客户端抽象 + 分布式锁 + 缓存模板
├── queue/           # 消息队列抽象（支持 in-memory / Redis Stream / RabbitMQ）
├── security/        # JWT + 密码哈希 + Principal + RBAC 权限
├── httpx/           # 请求绑定/校验 + 响应写入 + 中间件工具
├── observability/   # 追踪 + 指标 + 健康检查指标
├── scheduler/       # 定时任务框架
└── migration/       # 数据库迁移运行器
```

**当前差距：** config 缺 hot-reload；logger 缺采样。

---

## 四、业务模块层（Modules）

```
internal/modules/
├── auth/             # 认证（完整）
├── user/             # 用户（基础）
├── order/            # 订单（骨架）
├── payment/          # 支付（骨架）
├── inventory/        # 库存/商品（骨架）
├── notification/     # 通知（骨架）
├── admin/            # 管理后台 API
├── teacher/          # 选手模块
├── chat/             # 聊天 / WebSocket
├── game/             # 游戏房间
├── finance/          # 财务 / 对账 / 提现
├── kook/             # KOOK 机器人集成
├── file/             # 文件上传
├── review/           # 评价模块
├── coupon/           # 优惠券
└── analytics/        # 统计 / 访客追踪
```

### 每个模块的标准结构

```
internal/modules/<domain>/
├── handler.go       # HTTP handler + 路由注册（薄，只做 HTTP 语义）
├── service.go       # 业务逻辑（含事务边界、领域事件发布）
├── repository.go    # 数据访问（只做 CRUD，不做业务决策）
├── dto.go           # 请求/响应 DTO
├── model.go         # 领域模型
├── status.go        # 状态枚举/常量
├── errors.go        # 模块专属错误码 sentinel
├── events.go        # 领域事件定义
└── *_test.go        # 单元测试
```

### 模块间依赖规则

- ✅ 允许：`handler → service → repository`
- ✅ 允许：任何模块 → `platform` 层
- ❌ 禁止：业务模块之间直接 import model/table
- ✅ 允许：通过接口解耦（如 `auth` 的 `UserReader` 接口）
- ✅ 允许：通过领域事件跨模块协作

**当前差距：** 目前只实现了 auth（完整）和 user（基础），其余 4 个是 Ping。Java 后端有 15+ 业务域。

---

## 五、第三方适配器层（Clients）

```
internal/clients/
├── paymentgateway/    # 微信支付 + 支付宝适配器接口
├── sms/               # 短信发送
├── email/             # 邮件发送
├── storage/           # 七牛云 / OSS 文件存储
├── oauth/             # 微信 OAuth、小程序登录
├── faceid/            # 腾讯云人脸核身
└── kook/              # KOOK 机器人 API
```

**原则：** 每个第三方都抽象为接口，业务代码只依赖接口，不直接依赖 SDK。

**当前差距：** 有 sms/email/paymentgateway 接口骨架，但没实现。缺 storage/oauth/faceid/kook。

---

## 六、API 契约层

```
api/
├── openapi.yaml       # OpenAPI 3.0 完整规格
├── postman/           # Postman collection（可选）
└── gen/               # 从 OpenAPI 生成的客户端代码（可选）
```

**原则：** OpenAPI 作为前后端契约，API 先于代码。当前仅在 auth 和 ping 路由有文档。

**当前差距：** 文档只覆盖现有路由，远未覆盖完整 API。

---

## 七、数据层

```
migrations/            # Goose 数据库迁移
├── 00001_xxx.sql
├── 00002_xxx.sql
└── ...

sql/queries/           # sqlc 查询定义
├── auth.sql
├── user.sql
├── order.sql
└── ...

internal/platform/database/generated/   # sqlc 生成代码
├── db.go
├── models.go
└── *.sql.go
```

**原则：**
- 迁移文件必须可回滚、幂等
- sqlc 查询按模块分文件
- 生成代码不手动修改

**当前差距：** 只有 2 个迁移文件和 2 个查询文件。生产需要管理 ~70 个迁移。

---

## 八、测试结构

```
test/
├── integration/        # 集成测试（testcontainers）
│   ├── test_helpers.go
│   ├── auth_login_test.go
│   ├── health_test.go
│   └── ...
└── fixtures/           # 测试数据（YAML/JSON）

internal/modules/*/*_test.go   # 单元测试（与源码同目录）
internal/platform/*/*_test.go  # 基础设施层单元测试
```

**当前差距：** 集成测试框架有了，但覆盖度低（只有 auth + health）。单元测试在 platform 层覆盖较好，modules 层只有 auth 有测试。

---

## 九、运维与部署

```
├── Dockerfile            # 多阶段构建
├── docker-compose.yml    # 本地开发编排
├── Makefile              # build/test/lint/migrate/sqlc
├── .golangci.yml         # golangci-lint 配置
├── configs/              # 分环境配置
│   ├── config.local.yaml
│   ├── config.dev.yaml
│   ├── config.test.yaml
│   └── config.prod.yaml
└── scripts/              # 部署脚本
```

**当前差距：** 大部分已有，但 Makefile 缺 `build`/`fmt`/`tidy`/`check` 等常用命令。

---

## 当前 Go 地基完整度评分

| 层级 | 完整度 | 优先级 |
|------|--------|--------|
| 入口层 | 60% | 低 |
| 装配层 | 80% | 低 |
| **基础设施** | **92%** | **高** |
| 业务模块 | 15% | 高 |
| 第三方适配器 | 10% | 中 |
| API 契约 | 25% | 中 |
| 数据层 | 10% | 高 |
| 测试 | 35% | 中 |
| 运维 | 70% | 低 |
