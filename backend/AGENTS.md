# PROJECT KNOWLEDGE BASE

**Generated:** 2026-07-01 Asia/Shanghai
**Branch:** main

## OVERVIEW
这是一个 Go 1.26 后端工程，采用“模块化单体 + API/Worker 双入口”。代码边界按 `cmd -> bootstrap -> modules/clients/platform` 组织。

## STRUCTURE
```text
backend/
├── cmd/                # API / worker 进程入口
├── internal/
│   ├── bootstrap/      # 启动装配、路由、中间件、生命周期
│   ├── modules/        # order / payment / inventory / notification
│   ├── platform/       # config / logger / errors / response / observability / database
│   └── clients/        # 第三方服务适配器
├── configs/            # config.local/dev/test/prod.yaml
├── api/                # OpenAPI 契约
├── migrations/         # 数据库迁移占位
├── sql/queries/        # SQL / sqlc 查询占位
└── test/integration/   # 集成测试
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| API 入口 | `cmd/api/main.go` | 读取配置、组装依赖、注册四个模块、启动 HTTP 服务 |
| Worker 入口 | `cmd/worker/main.go` | 读取配置、初始化 worker、注册任务 |
| 路由挂载 | `internal/bootstrap/server.go` | `/healthz`、`/readyz`、`/api/v1/*` |
| 生命周期 / shutdown | `internal/bootstrap/app.go` | 聚合 `Shutdowner` |
| 配置规则 | `internal/platform/config/config.go` | `APP_CONFIG_PATH` 优先，其次 `configs/config.<env>.yaml` |
| 统一错误 / 响应 | `internal/platform/errors`、`internal/platform/response` | 稳定错误码 + 统一 APIResponse |
| 第三方适配边界 | `internal/clients/AGENTS.md` | paymentgateway / sms / email 适配器规范 |
| 集成测试样例 | `test/integration/ping_test.go` | 验证 bootstrap + modules 注册链路 |
| 强约束来源 | `go_backend_development_standards.md`、`go_backend_architecture_payment_order_inventory_notification.md` | 写代码前先对照 |

## CODE MAP
| Symbol | Type | Location | Role |
|--------|------|----------|------|
| `main` | function | `cmd/api/main.go` | API 装配入口 |
| `NewAPIEngine` | function | `internal/bootstrap/server.go` | 创建 Gin 引擎并注册健康检查、业务路由 |
| `Dependencies` | struct | `internal/bootstrap/app.go` | 启动层共享依赖容器 |
| `Config` | struct | `internal/platform/config/config.go` | 强类型配置模型 |
| `LoadConfig` | function | `internal/platform/config/config.go` | 读取配置、环境变量覆盖、校验 |
| `AppError` | struct | `internal/platform/errors/errors.go` | 统一业务错误模型（stack + WithCause + HTTP status） |
| `Code` | const | `internal/platform/errors/codes.go` | 9 个全局错误码，启动时唯一性校验 |
| `Handler.RegisterRoutes` | method | `internal/modules/*/handler.go` | 模块向 `/api/v1` 挂路由 |
| `AccessLogMiddleware` | func | `internal/bootstrap/middleware.go` | 记录每个请求的方法/路径/状态/耗时 |
| `TxOption` | struct | `internal/platform/database/tx.go` | 事务选项（ReadOnly / Timeout） |

## CONVENTIONS
- 只在 `cmd/*/main.go` 做装配：读配置、构造依赖、注册 handler、启动进程。
- 业务模块可以依赖 `platform`，但 `platform` 不能依赖任何业务模块。
- Handler 只负责 HTTP 语义；业务状态流转只能放在 Service。
- Repository 只表达数据访问，不做业务决策，不调用 Service/Client/MQ。
- 事件命名遵循 `domain.action.v1`。
- 错误码稳定，不随文案改动。

## ANTI-PATTERNS (THIS PROJECT)
- 不要在 `cmd/*` 写业务逻辑。
- 不要在 Handler 直接调 Repository。
- 不要在 Repository 决定订单/支付/库存/通知状态流转。
- 不要跨模块直接读对方数据表。
- 不要让业务代码直接碰第三方 SDK；走 `internal/clients` 或接口抽象。
- 不要把 `platform` 做成隐式全局工具箱；按能力分包，不要堆 `utils`。

## UNIQUE STYLES
- 配置分环境文件管理：`local` / `dev` / `test` / `prod`。
- `config.test.yaml` 使用测试端口、测试 DSN、`mq.driver: in-memory`。
- 集成测试直接组装 `bootstrap.NewAPIEngine(...)`，不是通过 CLI 黑盒启动进程。
- 当前四个业务域文件骨架一致，优先抽象共享模块规则，而不是为每个域单独写规范。

## COMMANDS
```bash
make run-api
make run-worker
make test
make lint
go test ./...
golangci-lint run
docker compose up api worker
```

## NOTES
- 当前仓库仍处于 bootstrap 阶段：`migrations/`、`sql/queries/`、`test/fixtures/` 以占位为主。
- 未发现仓库内 CI 工作流文件；`Makefile` 是主要本地命令入口，`docker-compose.yml` 提供容器化启动入口。
- 进入子目录工作前，继续读取对应子层级的 AGENTS.md。
- 三地基层已完成改造：**错误处理**（stack + codes + auto-map）、**日志**（Debug + JSON/Text + cfg 驱动）、**数据库**（单池 + TxOption）。
- 新增代码文件：`internal/platform/errors/codes.go`（全局错误码）、`internal/platform/redis/`（Redis 客户端/锁/缓存）。
- `response.Fail` 签名改为 `(c, err error)` 自动 `errors.As` 映射，不再要求调用方传 `AppError`。
