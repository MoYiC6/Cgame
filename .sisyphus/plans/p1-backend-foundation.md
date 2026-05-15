# P1 Backend Foundation Plan

## TL;DR

> **Quick Summary**: 在不引入任何真实业务接口或真实业务 schema 的前提下，为 `backend/` 建立可执行、可测试、可观测的 P1 公共底座：统一响应契约、统一请求绑定与校验、事务抽象、真实数据库基础设施、OpenTelemetry traces + log correlation 闭环、四模块骨架统一接缝。
>
> **Deliverables**:
> - 顶层统一响应升级为 `code / message / data / request_id / trace_id`
> - 统一 `BindAndValidate` 请求绑定校验能力（覆盖 JSON + query + path + header）
> - `TxManager.WithinTx` 事务抽象与 repository 协作约定
> - `pgx + sqlc + goose + Testcontainers` 数据库地基
> - OTEL provider / propagation / Gin 入站 traces / logger correlation / DB tracing / graceful shutdown
> - `/readyz` 改为真实依赖 DB readiness，`/healthz` 仅表示进程存活
> - 四模块 `order/payment/inventory/notification` 骨架按统一 Handler/Service/Repository 规范对齐
> - OpenAPI、测试、lint、Makefile 命令与验收链路补齐
>
> **Estimated Effort**: Large
> **Parallel Execution**: YES - 4 implementation waves + 1 final verification wave
> **Critical Path**: 1 → 4 → 8 → 12 → 17 → F1-F4

---

## Context

### Original Request
用户希望把 `backend/todo.md` 中建议的后端基础设施与 P1 规范真正落地在当前 Go 后端工程里，重点是：统一参数校验、事务抽象、Repository/Service 边界、统一响应、健康检查、request_id/trace_id、OpenAPI、基础测试、lint，并进一步明确本次 P1 还要把 `sqlc + pgx + goose + Testcontainers` 和 OpenTelemetry traces + log correlation 主链路闭环一起纳入。

### Interview Summary
**Key Discussions**:
- P1 只做公共 foundation + 四模块统一骨架，不做任何真实业务接口。
- 测试策略采用 **TDD**。
- 统一响应必须升级为顶层 `code / message / data / request_id / trace_id`。
- 总体落地路径采用 **方案 A：平台优先，模块跟随**。
- 以 `backend/todo.md` 为主约束来源，并尽量贴合现有 `cmd -> bootstrap -> modules -> platform` 分层。
- OpenTelemetry 范围锁定为 **traces + log correlation 主链路闭环**，不扩展到 metrics/dashboard/collector 运维编排。
- 数据库地基必须包含 `sqlc + pgx + goose + Testcontainers`，但不能借机引入真实业务 schema / CRUD。
- `/readyz` 必须依赖 DB，可失败；`/healthz` 仅表示进程存活。
- migration 采用 **独立命令执行**，应用启动不自动跑 migration；测试环境可自动跑。
- `TxManager.WithinTx` 嵌套事务语义锁定为：**嵌套复用外层事务**。
- `BindAndValidate` 覆盖范围锁定为：**JSON + query + path + header 一次性统一**。
- 计划后续执行偏好：先走审查直到通过，再交给执行器实施；每完成一个任务点单独提交一次 commit。

**Research Findings**:
- 当前已有 `bootstrap.NewAPIEngine(...)`、`AppError`、`APIResponse`、`RequestIDMiddleware`、`TraceContextMiddleware`、`/healthz`、`/readyz`、四模块 `/api/v1/*/ping`、基础集成测试与 Makefile。
- 当前 `response.APIResponse` 只有顶层 `request_id`，没有顶层 `trace_id`。
- 当前 `database/db.go` 仅有 `DummyDB` 和空的 `RunMigrations()`；`cmd/api/main.go` 仍装配 noop tracer / noop propagator / dummy DB。
- 当前 `observability` 只有 request/trace context 辅助与 noop tracer / propagator，没有真实 provider 与 shutdown。
- 当前 `api/openapi.yaml` 只覆盖 `/healthz`、`/readyz`，没有模块 `/ping` 与统一响应 schema。
- 四模块当前高度同构，适合通过统一契约与构造注入对齐，而不是引入代码生成器或通用业务基类。

### Metis Review
**Identified Gaps** (addressed):
- `full OpenTelemetry` 定义过宽：已收敛为 provider 初始化、Gin 入站 tracing、trace propagation、logger correlation、DB tracing、shutdown flush。
- DB groundwork 易膨胀成真实业务 schema / CRUD：已明确只允许平台级非业务验证载体，不允许 `orders/payments/...` 业务表。
- 四模块统一 skeleton 易演变成 framework/codegen：已明确只做结构和契约统一，不做脚手架系统。
- bind helper、OpenAPI、事务抽象容易超纲：已把范围锁定在当前 P1 公共接缝，不预埋未来业务框架。

### Implementation Decisions
- **DB 最小验证载体**：P1 只允许平台级非业务探针载体，统一命名为 `platform_runtime_probes`（或同语义等价名）。`migrations/` 只创建该类平台表；`sql/queries/` 只允许围绕该表提供最小查询，如 `InsertRuntimeProbe`、`CountRuntimeProbesByRunID`、`DeleteRuntimeProbesByRunID`，用于验证 `goose + sqlc + pgx + transaction + readiness` 链路，禁止出现任何业务表或业务 query。
- **事务传播模型**：`TxManager.WithinTx(ctx, fn)` 在 `platform/database` 内把事务 executor 写入 `context.Context`；`platform/database` 提供 `ExecutorFromContext(ctx)`（或等价 resolver）解析当前 `DBTX`；repository 统一只接收 `context.Context`，内部优先取 context 中的 tx executor，无则回退连接池。禁止 service 显式向下传原始 tx handle。
- **BindAndValidate 绑定规则**：P1 只支持 `JSON body + query + path + header` 四类输入；helper 内部固定调用顺序为 `ShouldBindUri -> ShouldBindQuery -> ShouldBindHeader -> ShouldBindJSON`；body 仅支持 JSON，不扩展 form/multipart/XML；同一 DTO 字段在 P1 中不得同时承担多个输入源，借此消除覆盖歧义；bind/validate 失败统一映射到 `AppError`，默认错误码使用 `INVALID_ARGUMENT`（若实现中沿用现有等价错误码，必须在测试中写死）。
- **OTEL 抽象策略**：保留 `request_id` / `trace_id` 的 context helper，但不强制维持完整 vendor-agnostic tracing façade；P1 可以让 provider / propagator / middleware / shutdown 更贴近 OpenTelemetry 原生接线方式，只要求现有 bootstrap 依赖位和测试契约保持稳定。
- **模块对齐范围**：四模块只对齐 constructor seam、`/ping` 契约、repository 形态、`TxManager` 接入位与统一响应/绑定约定；不得新增真实业务 endpoint、真实业务 DTO、真实业务 CRUD、真实业务表，也不得把 skeleton 对齐升级成 base class / codegen 框架。
- **执行证据定位**：`.sisyphus/evidence/` 中的文件属于执行阶段 QA 产物，不属于源码依赖；缺失 evidence 代表 QA 未完成，而不是要求在代码层新增额外运行时耦合。

---

## Work Objectives

### Core Objective
在当前 `backend/` 工程内，把 P1 所需的公共接缝从占位状态提升为真实可运行状态：统一响应、绑定校验、事务、数据库、可观测性、健康检查、OpenAPI、测试与 lint 同步完成；同时保持业务模块仍停留在统一 skeleton，不跨入真实业务实现。

### Concrete Deliverables
- `internal/platform/response`：顶层 `trace_id` 统一响应支持
- `internal/platform` 新增统一请求绑定/校验能力
- `internal/platform/database`：真实 DB 连接池、事务管理、迁移入口、sqlc 协作抽象
- `internal/platform/observability`：真实 provider / propagator / tracer / shutdown
- `internal/bootstrap`：API / worker 装配切换到真实 platform 依赖，`/readyz` 真实检查 DB
- `cmd/api` 与 `cmd/worker`：接入 config-driven observability / database initialization / graceful shutdown
- `migrations/` 与 `sql/queries/`：最小非业务 schema/query 链路
- `test/integration/` 与 platform/bootstrap/module 测试：覆盖 TDD 验证矩阵
- `api/openapi.yaml`：补齐 `/healthz`、`/readyz`、四模块 `/ping` 与统一响应结构
- `.golangci.yml`、`Makefile`：补齐 P1 要求的 lint / test / migration / sqlc / integration 命令

### Definition of Done
- [ ] `cd backend && go test ./...` 通过
- [ ] `cd backend && make test` 通过
- [ ] `cd backend && make lint` 通过
- [ ] `cd backend && golangci-lint run` 通过
- [ ] `cd backend && make sqlc-generate` 可执行并无未提交生成差异
- [ ] `cd backend && make test-integration` 通过，包含 Testcontainers + migration + readiness 链路
- [ ] `/healthz`、`/readyz`、`/api/v1/{order|payment|inventory|notification}/ping` 都返回统一响应结构，且顶层含 `request_id`、`trace_id`

### Must Have
- 只在 `backend/` 工作目录内定义与验证所有命令
- 所有新增错误都统一映射为 `AppError`
- 所有统一响应都返回顶层 `trace_id`
- `BindAndValidate` 统一处理 JSON/query/path/header，并复用 `validator/v10`
- `TxManager.WithinTx` 支持嵌套复用外层事务，service 使用事务，repository 不管理事务
- `/readyz` 真实依赖 DB ping
- OTEL 至少覆盖 HTTP 入站、trace propagation、logger correlation、DB tracing、provider shutdown
- 数据库链路必须通过 `pgx + sqlc + goose + Testcontainers` 跑通
- 四模块骨架统一接入公共接缝，不增加真实业务 endpoint

### Must NOT Have (Guardrails)
- 不新增任何真实业务 endpoint、真实业务 DTO、真实业务 CRUD、真实业务 schema
- 不在 `cmd/*` 写业务逻辑
- 不把四模块统一做成代码生成器、模板引擎、base service/repository framework
- 不把事务抽象扩展成 Unit of Work / Outbox / Saga
- 不把 OTEL 扩展到 metrics、dashboard、collector 部署、外部平台运维编排
- 不推倒重写现有 `AppError`、`APIResponse`、middleware、ping/health 路由；仅做最小必要增量改造
- 不以未来业务接口作为 DB / tracing 验证载体

---

## Verification Strategy (MANDATORY)

> **ZERO HUMAN INTERVENTION** - 所有验收必须由 agent 自动执行，禁止“人工点一下看看”。

### Test Decision
- **Infrastructure exists**: YES
- **Automated tests**: TDD
- **Framework**: Go `testing` + `httptest` + Testcontainers for Go
- **If TDD**: 每个任务必须先补失败测试，再写最小实现，再回归通过

### QA Policy
每个任务都必须给出 agent-executed QA scenarios，并把证据保存到 `.sisyphus/evidence/`。

- **Frontend/UI**: 本计划无
- **TUI/CLI**: `bash` / `interactive_bash` 运行 `make`、`go test`、`golangci-lint`、`goose`、`sqlc`
- **API/Backend**: `bash` + `go test` + `httptest` + `curl`/HTTP integration assertions
- **Library/Module**: `go test` 针对 platform/bootstrap/modules 同目录测试

Evidence examples:
- `.sisyphus/evidence/task-4-response-tests.txt`
- `.sisyphus/evidence/task-8-db-integration.txt`
- `.sisyphus/evidence/task-12-otel-tests.txt`
- `.sisyphus/evidence/final-qa/health-and-ping.txt`

---

## Execution Strategy

### Parallel Execution Waves

> 以“平台契约 → 基础设施 → 模块接入 → 文档与验证”为主依赖链。每波内部尽量并行。

```text
Wave 1 (Start Immediately - contracts & test scaffolding):
├── Task 1: 锁定配置与命令基线 [quick]
├── Task 2: TDD 搭建统一响应回归测试 [quick]
├── Task 3: TDD 搭建 bind+validate 契约测试 [quick]
├── Task 4: TDD 搭建事务抽象契约测试 [quick]
├── Task 5: TDD 搭建 observability 契约测试 [quick]
└── Task 6: TDD 搭建 DB groundwork 集成测试骨架 [unspecified-high]

Wave 2 (After Wave 1 - platform implementations):
├── Task 7: 实现顶层 trace_id 统一响应 [quick]
├── Task 8: 实现 BindAndValidate 与 validation error 映射 [unspecified-high]
├── Task 9: 实现 TxManager / DBTX / nested transaction contract [deep]
├── Task 10: 实现真实 DB 连接池、migration、sqlc 基座 [unspecified-high]
└── Task 11: 实现 OTEL provider / propagator / tracer / shutdown [deep]

Wave 3 (After Wave 2 - bootstrap wiring & module alignment):
├── Task 12: 接入 bootstrap API/worker 装配与 graceful shutdown [deep]
├── Task 13: 改造 health/readiness 链路依赖真实 DB [quick]
├── Task 14: 四模块 handler/service/repository 统一接缝对齐 [unspecified-high]
├── Task 15: 日志关联与 DB tracing 接入闭环 [unspecified-high]
└── Task 16: 集成测试扩展到顶层 trace_id / readiness / ping [quick]

Wave 4 (After Wave 3 - docs, tooling, openapi):
├── Task 17: 更新 OpenAPI 统一响应与健康/ping 契约 [writing]
├── Task 18: 扩展 Makefile / lint / sqlc / migration / integration 命令 [quick]
├── Task 19: 补齐 .golangci.yml 到 P1 要求 [quick]
└── Task 20: 验证 worker 占位任务与 shutdown/observability 一致性 [quick]

Wave FINAL (After ALL tasks — 4 parallel reviews, then user okay):
├── Task F1: Plan compliance audit (oracle)
├── Task F2: Code quality review (unspecified-high)
├── Task F3: Real manual QA execution (unspecified-high)
└── Task F4: Scope fidelity check (deep)
-> Present results -> Get explicit user okay

Critical Path: 1 → 4 → 9 → 10 → 12 → 17 → F1-F4
Parallel Speedup: ~65% faster than sequential
Max Concurrent: 6
```

### Dependency Matrix
- **1**: Blocked By: None | Blocks: 10, 18
- **2**: Blocked By: None | Blocks: 7, 16
- **3**: Blocked By: None | Blocks: 8, 14
- **4**: Blocked By: None | Blocks: 9, 14
- **5**: Blocked By: None | Blocks: 11, 15
- **6**: Blocked By: None | Blocks: 10, 16
- **7**: Blocked By: 2 | Blocks: 12, 16, 17
- **8**: Blocked By: 3 | Blocks: 12, 14
- **9**: Blocked By: 4 | Blocks: 10, 14, 15
- **10**: Blocked By: 1, 6, 9 | Blocks: 12, 13, 15, 18
- **11**: Blocked By: 5 | Blocks: 12, 15, 20
- **12**: Blocked By: 7, 8, 10, 11 | Blocks: 13, 16, 20
- **13**: Blocked By: 10, 12 | Blocks: 16, 17
- **14**: Blocked By: 3, 8, 9 | Blocks: 16, 17
- **15**: Blocked By: 5, 9, 10, 11 | Blocks: 16, 20
- **16**: Blocked By: 2, 6, 7, 12, 13, 14, 15 | Blocks: F1, F2, F3, F4
- **17**: Blocked By: 7, 13, 14 | Blocks: F1, F4
- **18**: Blocked By: 1, 10 | Blocks: F2, F3
- **19**: Blocked By: None | Blocks: F2
- **20**: Blocked By: 11, 12, 15 | Blocks: F1, F3

### Agent Dispatch Summary
- **Wave 1**: T1-T5 → `quick`; T6 → `unspecified-high`
- **Wave 2**: T7 → `quick`; T8/T10 → `unspecified-high`; T9/T11 → `deep`
- **Wave 3**: T12 → `deep`; T13/T16 → `quick`; T14/T15 → `unspecified-high`
- **Wave 4**: T17 → `writing`; T18/T19/T20 → `quick`
- **FINAL**: F1 → `oracle`; F2/F3 → `unspecified-high`; F4 → `deep`

---

## TODOs

> 实现 + 测试 = 一个任务。每个任务都必须包含明确 QA 场景。以下先写入任务批次；后续分批追加到本 section。

- [x] 1. 锁定配置与命令基线

  **What to do**:
  - 扩展 `backend/internal/platform/config/config.go` 的强类型配置，补齐 P1 需要的 DB 与 OTEL 配置段，并保持“只有 config 包读取环境变量”的规则不变。
  - 同步更新 `backend/configs/config.local.yaml`、`config.dev.yaml`、`config.test.yaml`、`config.prod.yaml`，为后续 `pgx/sqlc/goose/OTEL` 接入预留一致字段。
  - 扩展 `backend/internal/platform/config/config_test.go`，覆盖配置文件加载、环境覆盖、缺失必填项、脱敏摘要，特别是新增的 DB/OTEL 字段。
  - 为后续任务固定命令名：`make test`、`make lint`、`make sqlc-generate`、`make migrate-up`、`make test-integration`，但本任务只锁定命名和配置契约，不实现数据库逻辑。

  **Must NOT do**:
  - 不在本任务里初始化真实 DB 或 OTEL provider。
  - 不把 DSN、exporter endpoint、token 等敏感值直接打印到日志或摘要。
  - 不在 `config` 包外直接读取环境变量。

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 以配置模型和测试为主，改动面集中且反馈快。
  - **Skills**: [`test-driven-development`, `verification-before-completion`]
    - `test-driven-development`: 先写失败配置测试，防止后续基础设施任务踩空字段。
    - `verification-before-completion`: 确保配置命令与测试输出都有证据。
  - **Skills Evaluated but Omitted**:
    - `feature-research`: 现有 `config` 模式清晰，无需额外调研。

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 2, 3, 4, 5, 6)
  - **Blocks**: 10, 18
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `backend/internal/platform/config/config.go:12-171` - 当前强类型配置、env override、Validate、MaskedSummary 的基线实现。
  - `backend/internal/platform/config/config_test.go:10-123` - 现有配置测试风格；后续新增字段与校验应沿用这里的断言结构。
  - `backend/Makefile:1-13` - 目前只有 `run-api/run-worker/test/lint`，后续命令扩展必须兼容这里的入口风格。
  - `backend/configs/` - 当前多环境配置文件目录；新增字段必须四套环境文件一致对齐。

  **API/Type References**:
  - `backend/internal/bootstrap/app.go:13-19` - `Dependencies` 目前只依赖 `Config/Logger/Tracer/Propagator/DB`，新增配置要服务于这些依赖装配。

  **Test References**:
  - `backend/internal/platform/config/config_test.go:10-123` - 直接复用当前测试组织方式，不要发明另一套配置测试框架。

  **External References**:
  - `https://github.com/spf13/viper` - `backend/todo.md` 指向的配置方案来源；计划应兼容强类型配置收口。

  **WHY Each Reference Matters**:
  - 当前配置系统已经具备“文件 + env override + validate + mask”全链路；本任务应是在其上扩字段，而不是新起一套配置读取机制。

  **Acceptance Criteria**:
  - [ ] `backend/internal/platform/config/config_test.go` 新增 DB/OTEL 配置契约测试，并先失败后通过。
  - [ ] `cd backend && go test ./internal/platform/config -count=1` → PASS。
  - [ ] `MaskedSummary()` 仍不会暴露任何完整 DSN / endpoint credential。
  - [ ] 四套 `configs/config.*.yaml` 都包含后续 P1 所需字段，命名一致。

  **QA Scenarios (MANDATORY)**:

  ```text
  Scenario: 配置 happy path - 多环境配置可加载
    Tool: Bash (go test)
    Preconditions: backend/configs 下四套环境配置已补齐新增字段
    Steps:
      1. 运行 `cd backend && go test ./internal/platform/config -run 'TestLoadFromFile|TestLoadConfigAppliesEnvironmentOverrides|TestLoadConfigUsesAPPENVWhenArgumentEmpty' -count=1`
      2. 断言输出包含 `ok`，且返回码为 0
      3. 保存测试输出到 `.sisyphus/evidence/task-1-config-happy.txt`
    Expected Result: 配置文件加载、环境覆盖、默认环境回退全部通过
    Failure Indicators: 任一测试 FAIL；输出出现未识别字段、必填项缺失、未脱敏摘要
    Evidence: .sisyphus/evidence/task-1-config-happy.txt

  Scenario: 配置 failure path - 非法配置被拒绝
    Tool: Bash (go test)
    Preconditions: 已添加针对缺失新增必填字段的负向测试
    Steps:
      1. 运行 `cd backend && go test ./internal/platform/config -run 'TestLoadRejectsMissingAppName|TestLoadRejectsInvalidDBOrObservabilityConfig' -count=1`
      2. 断言测试进程退出 0，表示“正确拒绝非法配置”的断言成立
      3. 保存输出到 `.sisyphus/evidence/task-1-config-failure.txt`
    Expected Result: 非法配置场景被测试明确拒绝，且 safe failure 由测试验证
    Failure Indicators: 测试未覆盖非法配置；错误信息未指向缺失字段；返回码非 0
    Evidence: .sisyphus/evidence/task-1-config-failure.txt
  ```

  **Evidence to Capture**:
  - [ ] `.sisyphus/evidence/task-1-config-happy.txt`
  - [ ] `.sisyphus/evidence/task-1-config-failure.txt`

  **Commit**: YES
  - Message: `test(config): 锁定数据库与观测配置契约`
  - Files: `backend/internal/platform/config/*`, `backend/configs/*`
  - Pre-commit: `cd backend && go test ./internal/platform/config -count=1`

- [x] 2. 完成统一响应顶层 trace_id 契约

  **What to do**:
  - 修改 `backend/internal/platform/response/response.go`，让 `APIResponse` 顶层统一包含 `trace_id`，并同时在 `Success` / `Fail` 中自动从 context 提取。
  - 扩展 `backend/internal/platform/response/response_test.go`，覆盖成功响应、错误响应、缺失 trace 时的兼容行为。
  - 保持 `AppError` 作为失败响应唯一错误来源；只改响应壳，不改变业务错误码体系。

  **Must NOT do**:
  - 不把 `trace_id` 再塞回模块 payload 里作为主要返回通道。
  - 不修改 `AppError` 的语义或新增第二套错误响应模型。

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 改动集中在 response package 与单元测试，反馈很快。
  - **Skills**: [`test-driven-development`, `verification-before-completion`]
    - `test-driven-development`: 先写 `trace_id` 缺失的失败用例，再补实现。
    - `verification-before-completion`: 强制核对成功/失败两类响应都被覆盖。
  - **Skills Evaluated but Omitted**:
    - `subagent-driven-development`: 单包改动，不需要任务拆分。

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 3, 4, 5, 6)
  - **Blocks**: 7, 16
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `backend/internal/platform/response/response.go:11-26` - 现有统一响应入口，顶层 `trace_id` 需要从这里注入。
  - `backend/internal/platform/response/response_test.go:14-69` - 现有成功/失败响应测试基线。
  - `backend/internal/platform/observability/context.go:12-28` - `RequestIDFromContext` / `TraceIDFromContext` 的取值方式。
  - `backend/internal/bootstrap/server_test.go:17-52` - health 路由当前通过 `response.APIResponse` 解码，是响应壳回归的关键消费者。

  **API/Type References**:
  - `backend/internal/platform/errors/errors.go:8-68` - 失败响应仍须由 `AppError` 驱动 `code/status/message`。
  - `backend/internal/modules/order/dto.go:14-17` - 当前模块 payload 内自带 `trace_id` 示例字段，后续任务要逐步去除其“顶层职责”。

  **Test References**:
  - `backend/internal/platform/response/response_test.go:14-69` - 直接扩展现有测试文件，不另起响应契约测试目录。

  **External References**:
  - `https://spec.openapis.org/oas/v3.0.3` - 后续 OpenAPI 响应 schema 需要与这里的顶层字段保持一致。

  **WHY Each Reference Matters**:
  - 响应壳已经是稳定入口；只要这里正确抽取 `trace_id`，bootstrap、health、module ping、OpenAPI 和 integration tests 都能围绕同一契约回归。

  **Acceptance Criteria**:
  - [ ] `APIResponse` 顶层新增 `trace_id` 字段。
  - [ ] `Success` 与 `Fail` 都会从 request context 提取 `request_id` 和 `trace_id`。
  - [ ] `cd backend && go test ./internal/platform/response -count=1` → PASS。
  - [ ] 所有响应测试都不再依赖 `data.trace_id` 才能判断 trace 是否存在。

  **QA Scenarios (MANDATORY)**:

  ```text
  Scenario: 响应 happy path - 成功响应返回顶层 trace_id
    Tool: Bash (go test)
    Preconditions: response 包测试已覆盖成功响应
    Steps:
      1. 运行 `cd backend && go test ./internal/platform/response -run TestSuccessWritesRequestIDAndTraceID -count=1`
      2. 断言输出包含 `ok` 且返回码为 0
      3. 保存输出到 `.sisyphus/evidence/task-2-response-happy.txt`
    Expected Result: 成功响应顶层同时包含 `request_id` 与 `trace_id`
    Failure Indicators: `trace_id` 为空、字段仍只在 `data` 中、测试 FAIL
    Evidence: .sisyphus/evidence/task-2-response-happy.txt

  Scenario: 响应 failure path - 错误响应也返回顶层 trace_id
    Tool: Bash (go test)
    Preconditions: 已为 `Fail(...)` 增加负向测试
    Steps:
      1. 运行 `cd backend && go test ./internal/platform/response -run TestFailUsesAppErrorMetadataAndTraceID -count=1`
      2. 断言输出包含 `ok` 且返回码为 0
      3. 保存输出到 `.sisyphus/evidence/task-2-response-failure.txt`
    Expected Result: 400/500 类错误响应顶层仍包含 `trace_id`
    Failure Indicators: 错误响应缺少 `trace_id`；错误码/消息被破坏；返回码非 0
    Evidence: .sisyphus/evidence/task-2-response-failure.txt
  ```

  **Evidence to Capture**:
  - [ ] `.sisyphus/evidence/task-2-response-happy.txt`
  - [ ] `.sisyphus/evidence/task-2-response-failure.txt`

  **Commit**: YES
  - Message: `feat(response): 统一顶层 trace_id 响应契约`
  - Files: `backend/internal/platform/response/*`
  - Pre-commit: `cd backend && go test ./internal/platform/response -count=1`

- [x] 3. 实现统一 BindAndValidate 契约

  **What to do**:
  - 在 `backend/internal/platform/` 下新增请求绑定/校验 helper（默认放在 `internal/platform/httpx/` 或等价聚合位置），提供 `BindAndValidate[T any](c *gin.Context) (*T, error)` 一类统一入口。
  - 该 helper 必须统一覆盖 JSON body、query、path param、header 四类输入，并复用 `go-playground/validator/v10`。
  - 设计 validation error 到 `AppError` 的稳定映射约定，为后续 handler 统一使用铺路。
  - 为 helper 编写同目录测试，覆盖合法输入、缺失必填、query/path/header 绑定、错误 content-type 或格式错误。

  **Must NOT do**:
  - 不在每个 handler 手写重复绑定/校验代码。
  - 不把 helper 发展成新的 Web 框架层。
  - 不在本任务中修改真实业务 endpoint。

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: 同时涉及 Gin 绑定、validator、泛型 helper 和错误映射，复杂度高于普通单包改动。
  - **Skills**: [`test-driven-development`, `verification-before-completion`]
    - `test-driven-development`: 需要先锁 query/path/header/JSON 多输入契约。
    - `verification-before-completion`: 防止只覆盖 JSON 忽略其他三类输入。
  - **Skills Evaluated but Omitted**:
    - `feature-research`: Gin/validator 范围明确，主要工作是契约收口，不是技术选型。

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2, 4, 5, 6)
  - **Blocks**: 8, 14
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `backend/internal/bootstrap/server.go:15-37` - 当前 Gin 装配入口；helper 必须与现有 Gin 使用方式兼容。
  - `backend/internal/modules/order/handler.go:20-34` - 当前 handler 入口样式；后续模块要在这里替换手工读取/拼接逻辑。
  - `backend/internal/modules/order/dto.go:3-17` - 当前 DTO 放置位置；后续校验 tag 应按该文件组织。

  **API/Type References**:
  - `backend/internal/platform/errors/errors.go:8-68` - validation 失败最终必须映射成统一 `AppError`。
  - `backend/internal/platform/response/response.go:11-26` - helper 的错误输出最终要走统一响应壳，而不是直接 `c.JSON(...)`。

  **Test References**:
  - `backend/internal/bootstrap/server_test.go:17-52` - Gin + httptest 测试风格基线。
  - `backend/test/integration/ping_test.go:22-88` - 当前 integration test 装配方式；后续若为 helper 补集成路径应沿用这里的 engine 组装方式。

  **External References**:
  - `https://gin-gonic.com/en/docs/examples/binding-and-validation/` - Gin 绑定/校验行为参考。
  - `https://pkg.go.dev/github.com/go-playground/validator/v10` - validator tag 与错误结构参考。

  **WHY Each Reference Matters**:
  - 本项目已经确定 Gin 只做入口层，helper 应该是对 Gin 现有绑定能力的统一封装，而不是重造 request lifecycle。

  **Acceptance Criteria**:
  - [ ] 新增统一 helper，签名与行为足以支持 JSON + query + path + header。
  - [ ] helper 内部绑定顺序固定为 `ShouldBindUri -> ShouldBindQuery -> ShouldBindHeader -> ShouldBindJSON`，并有测试覆盖该顺序约束。
  - [ ] DTO 字段不依赖多输入源覆盖；若出现重复绑定歧义，测试会失败并要求拆分字段。
  - [ ] `cd backend && go test ./internal/platform/... -run 'TestBindAndValidate' -count=1` → PASS。
  - [ ] 至少有 1 个负向测试验证缺失必填字段返回统一错误。
  - [ ] 至少有 1 个负向测试验证错误格式输入不会泄露底层错误实现细节。

  **QA Scenarios (MANDATORY)**:

  ```text
  Scenario: 绑定 happy path - JSON + query/path/header 一次性绑定成功
    Tool: Bash (go test)
    Preconditions: 新增 helper 测试夹具路由与 DTO 校验 tag
    Steps:
      1. 运行 `cd backend && go test ./internal/platform/... -run TestBindAndValidateBindsJSONQueryPathAndHeader -count=1`
      2. 断言输出包含 `ok` 且返回码为 0
      3. 保存输出到 `.sisyphus/evidence/task-3-bind-happy.txt`
    Expected Result: 单次调用 helper 即可得到完整 DTO，且 validator 通过
    Failure Indicators: 任一输入源未绑定；校验未生效；返回码非 0
    Evidence: .sisyphus/evidence/task-3-bind-happy.txt

  Scenario: 绑定 failure path - 非法输入被统一拒绝
    Tool: Bash (go test)
    Preconditions: 已添加非法 JSON / 缺失必填 / 错误 header 的负向测试
    Steps:
      1. 运行 `cd backend && go test ./internal/platform/... -run 'TestBindAndValidateRejectsInvalidPayload|TestBindAndValidateRejectsMissingRequiredField' -count=1`
      2. 断言输出包含 `ok` 且返回码为 0
      3. 保存输出到 `.sisyphus/evidence/task-3-bind-failure.txt`
    Expected Result: helper 返回统一错误类型，测试显式验证 safe failure
    Failure Indicators: helper 吞错、泄露底层解析细节、未触发 validator、返回码非 0
    Evidence: .sisyphus/evidence/task-3-bind-failure.txt
  ```

  **Evidence to Capture**:
  - [ ] `.sisyphus/evidence/task-3-bind-happy.txt`
  - [ ] `.sisyphus/evidence/task-3-bind-failure.txt`

  **Commit**: YES
  - Message: `feat(platform): 落地统一请求绑定与参数校验`
  - Files: `backend/internal/platform/httpx/*`（或最终选定的 helper 目录）, `backend/internal/modules/*/dto.go`
  - Pre-commit: `cd backend && go test ./internal/platform/... -run TestBindAndValidate -count=1`

- [x] 4. 锁定 TxManager / repository 协作事务契约

  **What to do**:
  - 在 `backend/internal/platform/database/` 内定义 `TxManager`、`DBTX`（或等价 executor interface）、事务上下文传递方式以及嵌套事务复用外层事务的规则。
  - 事务传播固定采用 `WithinTx(ctx, fn)` + `ExecutorFromContext(ctx)`（或等价 resolver）模式：事务 executor 只允许由 `platform/database` 写入和读取 context，repository 统一只接收 `context.Context`。
  - 为事务抽象编写契约测试：commit、rollback、panic rollback、context cancel、nested reuse outer transaction。
  - 规定 service 如何调用 `WithinTx(ctx, fn)`，repository 如何只依赖 executor/CRUD 接口而不主动管理事务。
  - 如果需要 context key 传递 tx handle，必须把 key 封装在 `platform/database`，禁止业务层手动拼装。

  **Must NOT do**:
  - 不引入 Unit of Work / Outbox / Saga。
  - 不让 repository 自己 `Begin/Commit/Rollback`。
  - 不在本任务中创建真实业务 repository CRUD。

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 先锁定契约测试与接口形态，再进入真实 DB 实现，范围仍较集中。
  - **Skills**: [`test-driven-development`, `verification-before-completion`]
    - `test-driven-development`: 事务语义必须先有失败测试，否则实现极易漂移。
    - `verification-before-completion`: 核对 commit/rollback/panic/context/nested 五类场景全部覆盖。
  - **Skills Evaluated but Omitted**:
    - `feature-research`: 语义已由用户锁死，重点是契约落地。

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2, 3, 5, 6)
  - **Blocks**: 9, 14
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `backend/internal/platform/database/db.go:1-21` - 当前 database 包非常薄，适合在此扩展事务抽象而不是重建新包。
  - `backend/internal/modules/order/service.go:5-22` - 当前 service 只调用 repo，后续要在 service 层接入 `WithinTx`。
  - `backend/internal/modules/order/repository.go:5-17` - 当前 repository 只有 noop；后续 repository 只应接 executor，不管理事务生命周期。

  **API/Type References**:
  - 用户已确认事务语义：`type TxManager interface { WithinTx(ctx context.Context, fn func(ctx context.Context) error) error }`。
  - `backend/internal/platform/errors/errors.go:8-68` - 事务失败最终仍需映射到统一错误语义。

  **Test References**:
  - 项目内暂无事务测试，需在 `backend/internal/platform/database/*_test.go` 新建契约测试，沿用 Go 原生 `testing`。

  **External References**:
  - `https://pkg.go.dev/context` - context cancel / deadline 行为需要被测试明确验证。

  **WHY Each Reference Matters**:
  - 当前仓库没有任何真实事务协作基线；必须先把“service 开事务、repository 不管理事务”的边界通过契约测试写死，后续真实 pgx 实现才不会越界。

  **Acceptance Criteria**:
  - [ ] `platform/database` 内存在 `TxManager`、`DBTX` 与 `ExecutorFromContext(ctx)`（或等价 resolver）契约。
  - [ ] repository 统一只通过 `context.Context` 解析 executor，不要求 service 显式下传 tx handle。
  - [ ] 嵌套 `WithinTx` 会复用外层事务而不是开启第二层事务。
  - [ ] `cd backend && go test ./internal/platform/database -run 'TestTxManager' -count=1` → PASS。
  - [ ] 至少覆盖 commit / rollback / panic rollback / context cancel / nested reuse 五类测试。

  **QA Scenarios (MANDATORY)**:

  ```text
  Scenario: 事务 happy path - 成功提交
    Tool: Bash (go test)
    Preconditions: 已为 TxManager 编写成功提交测试
    Steps:
      1. 运行 `cd backend && go test ./internal/platform/database -run TestTxManagerCommitsOnSuccess -count=1`
      2. 断言输出包含 `ok` 且返回码为 0
      3. 保存输出到 `.sisyphus/evidence/task-4-tx-happy.txt`
    Expected Result: callback 无错误时事务提交，测试显式断言提交次数与状态
    Failure Indicators: 未提交、错误 rollback、测试 FAIL
    Evidence: .sisyphus/evidence/task-4-tx-happy.txt

  Scenario: 事务 failure path - 错误/ panic / context cancel 触发回滚
    Tool: Bash (go test)
    Preconditions: 已添加 rollback、panic、cancel、nested 负向测试
    Steps:
      1. 运行 `cd backend && go test ./internal/platform/database -run 'TestTxManagerRollsBackOnError|TestTxManagerRollsBackOnPanic|TestTxManagerHandlesContextCancel|TestTxManagerNestedReusesOuterTransaction' -count=1`
      2. 断言输出包含 `ok` 且返回码为 0
      3. 保存输出到 `.sisyphus/evidence/task-4-tx-failure.txt`
    Expected Result: 失败场景全部触发正确的回滚或 nested 复用行为
    Failure Indicators: panic 未回滚；context cancel 未终止；nested 开启了新事务；返回码非 0
    Evidence: .sisyphus/evidence/task-4-tx-failure.txt
  ```

  **Evidence to Capture**:
  - [ ] `.sisyphus/evidence/task-4-tx-happy.txt`
  - [ ] `.sisyphus/evidence/task-4-tx-failure.txt`

  **Commit**: YES
  - Message: `test(database): 锁定事务管理器协作契约`
  - Files: `backend/internal/platform/database/*`, `backend/internal/modules/*/service.go`, `backend/internal/modules/*/repository.go`（如仅接口签名占位）
  - Pre-commit: `cd backend && go test ./internal/platform/database -run TestTxManager -count=1`

- [x] 5. 锁定 observability 契约与 OTEL 最小闭环测试

  **What to do**:
  - 扩展 `backend/internal/platform/observability/` 契约，明确 tracer / propagator / shutdown 所需接口与配置。
  - 为入站无 trace context、带 traceparent/trace header、provider 未配置或 exporter 不可达等场景写失败测试。
  - 保留现有 `request_id` / `trace_id` context helper，但为未来 OTEL provider 接入写出稳定测试护栏。
  - 提前定义 API 响应顶层 `trace_id`、日志 `trace_id`、传播器 extract/inject 三者的协同预期。

  **Must NOT do**:
  - 不在本任务中接入外部 metrics/dashboard。
  - 不把 observability 逻辑散落到业务模块。

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 本任务先锁测试和接口预期，还不进入真实 provider 装配实现。
  - **Skills**: [`test-driven-development`, `verification-before-completion`]
    - `test-driven-development`: 先定义 OTEL 闭环测试护栏，避免后续实现范围失控。
    - `verification-before-completion`: 检查 header propagation、response trace、shutdown 行为都被测试涉及。
  - **Skills Evaluated but Omitted**:
    - `feature-research`: 边界已经由用户与 Oracle 锁死。

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2, 3, 4, 6)
  - **Blocks**: 11, 15
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `backend/internal/platform/observability/context.go:1-28` - request/trace context 现有最小契约。
  - `backend/internal/platform/observability/propagator.go:5-51` - 当前 `noopPropagator` 的 extract/inject 路径。
  - `backend/internal/platform/observability/tracer.go:5-25` - 当前 `Tracer` / `Span` 抽象。
  - `backend/internal/platform/observability/observability_test.go:8-36` - 现有 context round-trip 与 noop propagator 测试基线。
  - `backend/internal/bootstrap/middleware.go:15-43` - 当前 request_id / trace_id 注入链路。

  **API/Type References**:
  - `backend/internal/platform/logger/logger.go:65-85` - 日志关联目前通过 `WithContext` 读取 request_id/trace_id，需要与真实 provider 行为兼容。
  - `backend/internal/platform/response/response.go:11-26` - 顶层 `trace_id` 响应与 observability context 必须共用同一来源。

  **Test References**:
  - `backend/internal/platform/observability/observability_test.go:8-36` - 直接扩展此文件或同目录测试集合。

  **External References**:
  - `https://opentelemetry.io/docs/languages/go/` - Go OTEL SDK 初始化与 provider 生命周期参考。

  **WHY Each Reference Matters**:
  - 当前 observability 只有 noop 抽象；如果不先通过测试定义“什么叫 P1 闭环”，后续实现极易要么过轻（只保留 header），要么过重（引入整套运维体系）。

  **Acceptance Criteria**:
  - [ ] observability 测试覆盖 request/trace context round-trip、inject/extract、missing trace、自定义 header 或 W3C trace 传播预期、shutdown 行为。
  - [ ] `cd backend && go test ./internal/platform/observability -count=1` → PASS。
  - [ ] 至少有 1 个测试验证 provider/exporter 不可用时服务行为符合“可降级、不中断启动”的约定（若采用该约定）。

  **QA Scenarios (MANDATORY)**:

  ```text
  Scenario: 观测 happy path - trace 注入与提取闭环
    Tool: Bash (go test)
    Preconditions: observability 契约测试已补齐 inject/extract 与 context round-trip
    Steps:
      1. 运行 `cd backend && go test ./internal/platform/observability -run 'TestRequestAndTraceContextRoundTrip|TestPropagatorInjectExtract' -count=1`
      2. 断言输出包含 `ok` 且返回码为 0
      3. 保存输出到 `.sisyphus/evidence/task-5-observability-happy.txt`
    Expected Result: trace/request 能在 context 和 carrier 间闭环传播
    Failure Indicators: trace 丢失、inject/extract 不一致、返回码非 0
    Evidence: .sisyphus/evidence/task-5-observability-happy.txt

  Scenario: 观测 failure path - 缺失/异常 provider 被安全处理
    Tool: Bash (go test)
    Preconditions: 已添加 exporter 不可达或 provider 配置缺失的行为测试
    Steps:
      1. 运行 `cd backend && go test ./internal/platform/observability -run 'TestObservabilityHandlesMissingTraceContext|TestObservabilityProviderDegradesGracefully' -count=1`
      2. 断言输出包含 `ok` 且返回码为 0
      3. 保存输出到 `.sisyphus/evidence/task-5-observability-failure.txt`
    Expected Result: 服务不会因为观测依赖异常而破坏既定降级策略
    Failure Indicators: provider 初始化失败直接 panic；trace 丢失未被测试覆盖；返回码非 0
    Evidence: .sisyphus/evidence/task-5-observability-failure.txt
  ```

  **Evidence to Capture**:
  - [ ] `.sisyphus/evidence/task-5-observability-happy.txt`
  - [ ] `.sisyphus/evidence/task-5-observability-failure.txt`

  **Commit**: YES
  - Message: `test(observability): 锁定链路追踪最小闭环契约`
  - Files: `backend/internal/platform/observability/*`, `backend/internal/bootstrap/middleware.go`（如仅测试夹具需要）
  - Pre-commit: `cd backend && go test ./internal/platform/observability -count=1`

- [x] 6. 搭建 DB groundwork 的 TDD 集成测试骨架

  **What to do**:
  - 在 `backend/test/integration/` 或紧邻 database 包的集成测试目录中，新增 Testcontainers 驱动的 Postgres 集成测试骨架。
  - 该骨架必须覆盖：启动容器、应用 migration、执行最小平台级非业务 query、验证 readiness 链路将来可依赖真实 DB。
  - 为避免 scope creep，只允许引入“平台级探针表/查询”（例如 `app_runtime_probe` 或同级别非业务载体），禁止创建订单/支付等领域表。
  - 本任务以失败测试和测试 harness 为主，不在这里完成全部 DB 实现。

  **Must NOT do**:
  - 不创建真实业务 schema / CRUD / seed data。
  - 不让 `make test` 默认强制拉起容器，除非后续明确拆分 test target。

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: 涉及容器生命周期、migration 运行、集成测试组织，复杂度较高。
  - **Skills**: [`test-driven-development`, `verification-before-completion`]
    - `test-driven-development`: 先写失败的 integration harness，明确最终 DB 验收链路。
    - `verification-before-completion`: 需要严格区分 unit tests 与容器 integration tests 命令边界。
  - **Skills Evaluated but Omitted**:
    - `using-git-worktrees`: 当前仅写计划，不涉及隔离实现。

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2, 3, 4, 5)
  - **Blocks**: 10, 16
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `backend/test/integration/ping_test.go:22-88` - 当前 integration test 直接通过 bootstrap 组装 engine；DB 集成测试也应优先复用这种工程风格。
  - `backend/migrations/` - 目前为空目录，仅 `.gitkeep`；后续 migration 文件将从这里进入。
  - `backend/sql/queries/` - 目前为空目录，仅 `.gitkeep`；后续 sqlc 查询文件从这里进入。
  - `backend/docker-compose.yml:1-27` - 现有容器化思路，但集成测试不应依赖 compose；应由 Testcontainers 自主管理。

  **API/Type References**:
  - `backend/internal/platform/database/db.go:1-21` - 后续真实 DB 接口要满足这里的测试 harness。
  - `backend/internal/bootstrap/server.go:19-30` - readiness 当前已依赖 `deps.DB.Ping(...)`；集成测试需要最终覆盖这一链路。

  **Test References**:
  - `backend/test/integration/ping_test.go:22-88` - 可直接沿用 Gin `httptest` 与 response decode 风格。

  **External References**:
  - `https://golang.testcontainers.org/` - Testcontainers for Go 用法参考。

  **WHY Each Reference Matters**:
  - 当前项目已经用 integration tests 直接验证 bootstrap；DB groundwork 也应沿用同一测试哲学，而不是单独做难维护的外部 shell 脚本测试。

  **Acceptance Criteria**:
  - [ ] 存在可失败后再转绿的 Testcontainers 集成测试骨架。
  - [ ] 集成测试只使用平台级探针 schema/query，不引入业务表。
  - [ ] `cd backend && go test ./test/integration -run TestDatabaseGroundwork -count=1` 最终可通过。
  - [ ] 已明确 `make test` 与 `make test-integration` 的分工边界。

  **QA Scenarios (MANDATORY)**:

  ```text
  Scenario: DB 集成 happy path - 容器、迁移、探针查询跑通
    Tool: Bash (go test)
    Preconditions: Testcontainers 测试骨架、migration fixture、probe query fixture 已存在
    Steps:
      1. 运行 `cd backend && go test ./test/integration -run TestDatabaseGroundwork -count=1`
      2. 断言输出包含 `ok` 且返回码为 0
      3. 保存输出到 `.sisyphus/evidence/task-6-db-happy.txt`
    Expected Result: 容器启动、迁移执行、最小 probe query、清理容器全部成功
    Failure Indicators: Docker 启动失败未被测试捕获；migration 未执行；query 不通；返回码非 0
    Evidence: .sisyphus/evidence/task-6-db-happy.txt

  Scenario: DB 集成 failure path - migration 或容器异常被明确暴露
    Tool: Bash (go test)
    Preconditions: 已添加 migration 失败或容器依赖异常的负向断言
    Steps:
      1. 运行 `cd backend && go test ./test/integration -run 'TestDatabaseGroundworkFailsOnBrokenMigration|TestDatabaseGroundworkHandlesContainerDependencyFailure' -count=1`
      2. 断言测试进程退出 0，表示负向断言本身通过
      3. 保存输出到 `.sisyphus/evidence/task-6-db-failure.txt`
    Expected Result: 异常场景被测试准确暴露，不是静默跳过
    Failure Indicators: 失败原因模糊；测试直接 skip；返回码非 0
    Evidence: .sisyphus/evidence/task-6-db-failure.txt
  ```

  **Evidence to Capture**:
  - [ ] `.sisyphus/evidence/task-6-db-happy.txt`
  - [ ] `.sisyphus/evidence/task-6-db-failure.txt`

  **Commit**: YES
  - Message: `test(integration): 搭建数据库地基容器化回归骨架`
  - Files: `backend/test/integration/*`, `backend/migrations/*`, `backend/sql/queries/*`（如仅测试夹具）
  - Pre-commit: `cd backend && go test ./test/integration -run TestDatabaseGroundwork -count=1`

- [x] 7. 落地统一响应顶层 trace_id 实现并回归 bootstrap 消费方

  **What to do**:
  - 按 Task 2 的失败测试，把 `backend/internal/platform/response/response.go` 真正改为顶层输出 `trace_id`。
  - 回归 `backend/internal/bootstrap/server_test.go`、`backend/test/integration/ping_test.go` 等消费者，使断言从顶层读取 `trace_id`。
  - 为兼容过渡期，可保留 payload 中现有 `trace_id` 字段直到模块对齐任务完成，但顶层 `trace_id` 必须成为唯一强契约。

  **Must NOT do**:
  - 不在此任务中完成全部模块 payload 清理。
  - 不破坏现有 `request_id` 行为。

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 主要是将已锁定的响应契约真正落地到代码与测试消费者。
  - **Skills**: [`test-driven-development`, `verification-before-completion`]
    - `test-driven-development`: 先确保消费者测试因缺失顶层 `trace_id` 失败，再补实现。
    - `verification-before-completion`: 强制检查 platform 与 bootstrap / integration 两层消费方都回归通过。
  - **Skills Evaluated but Omitted**:
    - `subagent-driven-development`: 任务范围仍可控。

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 8, 9, 10, 11)
  - **Blocks**: 12, 16, 17
  - **Blocked By**: 2

  **References**:
  - `backend/internal/platform/response/response.go:11-26`
  - `backend/internal/platform/response/response_test.go:14-69`
  - `backend/internal/bootstrap/server_test.go:17-52`
  - `backend/test/integration/ping_test.go:22-88`
  - `backend/internal/modules/order/dto.go:14-17`

  **WHY Each Reference Matters**:
  - 响应契约一旦变更，最先受影响的是 bootstrap / integration 的解码断言；必须同步回归，避免“单元测绿，组装链路红”。

  **Acceptance Criteria**:
  - [ ] `response.Success/Fail` 真实输出顶层 `trace_id`。
  - [ ] `cd backend && go test ./internal/platform/response ./internal/bootstrap ./test/integration -count=1` → PASS。
  - [ ] `server_test.go` 与 `ping_test.go` 都显式断言顶层 `trace_id`。

  **QA Scenarios (MANDATORY)**:

  ```text
  Scenario: 响应落地 happy path - platform 与 bootstrap 同步转绿
    Tool: Bash (go test)
    Preconditions: Task 2 失败测试已存在
    Steps:
      1. 运行 `cd backend && go test ./internal/platform/response ./internal/bootstrap -count=1`
      2. 断言输出包含 `ok` 且返回码为 0
      3. 保存输出到 `.sisyphus/evidence/task-7-response-bootstrap-happy.txt`
    Expected Result: 顶层 `trace_id` 在 response 与 health/readiness 消费方均生效
    Failure Indicators: bootstrap 解码失败；顶层字段为空；返回码非 0
    Evidence: .sisyphus/evidence/task-7-response-bootstrap-happy.txt

  Scenario: 响应落地 failure path - 集成测试不再依赖 payload trace_id
    Tool: Bash (go test)
    Preconditions: `ping_test.go` 已更新为顶层断言优先
    Steps:
      1. 运行 `cd backend && go test ./test/integration -run TestBootstrapRegistersAllPingRoutes -count=1`
      2. 断言输出包含 `ok` 且返回码为 0
      3. 保存输出到 `.sisyphus/evidence/task-7-response-bootstrap-failure.txt`
    Expected Result: 即使 payload 未来移除 `trace_id`，集成测试仍能通过顶层字段验证链路
    Failure Indicators: 测试仍依赖 `data.trace_id`；返回码非 0
    Evidence: .sisyphus/evidence/task-7-response-bootstrap-failure.txt
  ```

  **Evidence to Capture**:
  - [ ] `.sisyphus/evidence/task-7-response-bootstrap-happy.txt`
  - [ ] `.sisyphus/evidence/task-7-response-bootstrap-failure.txt`

  **Commit**: YES
  - Message: `feat(response): 顶层 trace_id 接入组装链路`
  - Files: `backend/internal/platform/response/*`, `backend/internal/bootstrap/*`, `backend/test/integration/*`
  - Pre-commit: `cd backend && go test ./internal/platform/response ./internal/bootstrap ./test/integration -count=1`

- [x] 8. 落地 BindAndValidate 与 validation error 统一映射

  **What to do**:
  - 在选定的 platform helper 目录中实现 `BindAndValidate[T any](c *gin.Context) (*T, error)`。
  - 把 JSON/query/path/header 绑定顺序、优先级、validator 调用、unknown field 策略写死在实现与测试中。
  - 提供将 bind/validate 失败映射为 `AppError` 的 helper，供后续模块 handler 直接复用。
  - 若需要中间辅助类型或 decoder，限制在 platform 目录内，不向业务层泄漏底层细节。

  **Must NOT do**:
  - 不让 handler 继续复制散落的绑定逻辑。
  - 不在本任务里引入新的 router abstraction。

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: 要把多输入源绑定、validator、错误映射和 Gin 行为细节一次性落实到代码里。
  - **Skills**: [`test-driven-development`, `verification-before-completion`]
    - `test-driven-development`: 绑定顺序和错误映射都适合先写精确失败用例。
    - `verification-before-completion`: 防止只让 happy path 通过而忽略格式错误、字段缺失等失败路径。
  - **Skills Evaluated but Omitted**:
    - `writing-plans`: 已在本计划内，无需再次生成子计划。

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 7, 9, 10, 11)
  - **Blocks**: 12, 14
  - **Blocked By**: 3

  **References**:
  - `backend/internal/bootstrap/server.go:15-37`
  - `backend/internal/modules/order/handler.go:20-34`
  - `backend/internal/modules/order/dto.go:3-17`
  - `backend/internal/platform/errors/errors.go:8-68`
  - `backend/internal/platform/response/response.go:11-26`

  **WHY Each Reference Matters**:
  - helper 最终必须接回现有 Gin handler 与统一错误/响应壳；如果只在 platform 自测通过、却无法自然融入 handler，任务就不算完成。

  **Acceptance Criteria**:
  - [ ] 统一 helper 实现完成，测试全部转绿。
  - [ ] `cd backend && go test ./internal/platform/... -run TestBindAndValidate -count=1` → PASS。
  - [ ] validation/bind 错误能够稳定映射到 `AppError`，且 safe message 可用于统一响应。

  **QA Scenarios (MANDATORY)**:

  ```text
  Scenario: helper 落地 happy path - 统一绑定成功
    Tool: Bash (go test)
    Preconditions: Task 3 测试已存在
    Steps:
      1. 运行 `cd backend && go test ./internal/platform/... -run TestBindAndValidateBindsJSONQueryPathAndHeader -count=1`
      2. 断言输出包含 `ok` 且返回码为 0
      3. 保存输出到 `.sisyphus/evidence/task-8-bind-happy.txt`
    Expected Result: helper 实现后 happy path 测试由红转绿
    Failure Indicators: query/path/header 任一未绑定；validator 未触发；返回码非 0
    Evidence: .sisyphus/evidence/task-8-bind-happy.txt

  Scenario: helper 落地 failure path - 错误统一映射
    Tool: Bash (go test)
    Preconditions: 负向测试已存在
    Steps:
      1. 运行 `cd backend && go test ./internal/platform/... -run 'TestBindAndValidateRejectsInvalidPayload|TestBindAndValidateRejectsMissingRequiredField' -count=1`
      2. 断言输出包含 `ok` 且返回码为 0
      3. 保存输出到 `.sisyphus/evidence/task-8-bind-failure.txt`
    Expected Result: 所有负向测试验证 `AppError` 映射与 safe failure 行为
    Failure Indicators: 返回原始解析错误给外部；错误码不稳定；返回码非 0
    Evidence: .sisyphus/evidence/task-8-bind-failure.txt
  ```

  **Evidence to Capture**:
  - [ ] `.sisyphus/evidence/task-8-bind-happy.txt`
  - [ ] `.sisyphus/evidence/task-8-bind-failure.txt`

  **Commit**: YES
  - Message: `feat(platform): 实现统一绑定校验助手`
  - Files: `backend/internal/platform/httpx/*`（或最终目录）, `backend/internal/platform/errors/*`
  - Pre-commit: `cd backend && go test ./internal/platform/... -run TestBindAndValidate -count=1`

- [x] 9. 落地 TxManager、DBTX 与嵌套事务复用实现

  **What to do**:
  - 在 `backend/internal/platform/database/` 内实现真实 `TxManager`、`DBTX`/executor 抽象、context 事务句柄传递与 nested reuse outer transaction 逻辑。
  - 为后续 pgx 接入定义最小公共接口，使 repository 可以依赖“连接池或事务”的统一执行器，而不需要感知事务生命周期。
  - 实现 panic rollback、context cancel、普通 error rollback、success commit 的统一逻辑。
  - 保持 API 足够薄，只支持当前 P1 所需事务协作，不预留过度抽象。

  **Must NOT do**:
  - 不引入业务级 repository/CRUD。
  - 不在 service 层之外公开事务内部细节。

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: 涉及事务语义、context 传播、executor 抽象和后续真实 DB 接入的核心边界，是本计划最关键接缝之一。
  - **Skills**: [`test-driven-development`, `verification-before-completion`]
    - `test-driven-development`: 事务语义必须对照 Task 4 契约逐项转绿。
    - `verification-before-completion`: 提交前要核对五类事务行为全部有证据。
  - **Skills Evaluated but Omitted**:
    - `systematic-debugging`: 这是正常实现路径，不是失败后补救。

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 7, 8, 10, 11)
  - **Blocks**: 10, 14, 15
  - **Blocked By**: 4

  **References**:
  - `backend/internal/platform/database/db.go:1-21`
  - `backend/internal/modules/order/service.go:5-22`
  - `backend/internal/modules/order/repository.go:5-17`
  - 用户确认的 `WithinTx` 语义：嵌套复用外层事务

  **WHY Each Reference Matters**:
  - service / repository 当前都很薄，正好可以在不引入真实业务逻辑的前提下，把事务边界真正嵌进去。

  **Acceptance Criteria**:
  - [ ] `TxManager` 实现通过 Task 4 的所有事务契约测试。
  - [ ] `cd backend && go test ./internal/platform/database -run TestTxManager -count=1` → PASS。
  - [ ] repository 可通过统一 executor 接口运行，事务控制仍由 service/TxManager 持有。

  **QA Scenarios (MANDATORY)**:

  ```text
  Scenario: 事务实现 happy path - commit 与 nested reuse 正常
    Tool: Bash (go test)
    Preconditions: Task 4 事务契约测试已存在
    Steps:
      1. 运行 `cd backend && go test ./internal/platform/database -run 'TestTxManagerCommitsOnSuccess|TestTxManagerNestedReusesOuterTransaction' -count=1`
      2. 断言输出包含 `ok` 且返回码为 0
      3. 保存输出到 `.sisyphus/evidence/task-9-tx-happy.txt`
    Expected Result: 成功提交与嵌套复用全部通过
    Failure Indicators: nested 开新事务；提交未发生；返回码非 0
    Evidence: .sisyphus/evidence/task-9-tx-happy.txt

  Scenario: 事务实现 failure path - error/panic/cancel 正确回滚
    Tool: Bash (go test)
    Preconditions: 负向测试已存在
    Steps:
      1. 运行 `cd backend && go test ./internal/platform/database -run 'TestTxManagerRollsBackOnError|TestTxManagerRollsBackOnPanic|TestTxManagerHandlesContextCancel' -count=1`
      2. 断言输出包含 `ok` 且返回码为 0
      3. 保存输出到 `.sisyphus/evidence/task-9-tx-failure.txt`
    Expected Result: 三类失败路径均触发正确回滚
    Failure Indicators: 任一失败场景未 rollback；panic 泄漏；返回码非 0
    Evidence: .sisyphus/evidence/task-9-tx-failure.txt
  ```

  **Evidence to Capture**:
  - [ ] `.sisyphus/evidence/task-9-tx-happy.txt`
  - [ ] `.sisyphus/evidence/task-9-tx-failure.txt`

  **Commit**: YES
  - Message: `feat(database): 实现事务管理器与统一执行器`
  - Files: `backend/internal/platform/database/*`, `backend/internal/modules/*/service.go`, `backend/internal/modules/*/repository.go`
  - Pre-commit: `cd backend && go test ./internal/platform/database -run TestTxManager -count=1`

- [x] 10. 接入真实 DB groundwork：pgx + goose + sqlc + Testcontainers 目标实现

  **What to do**:
  - 用 `pgx` 落地真实数据库连接池与 `Ping(ctx)`，替换 `DummyDB` 占位实现。
  - 在 `backend/migrations/` 中新增唯一允许的最小平台级非业务 migration，围绕 `platform_runtime_probes`（或等价同语义命名）建表；该表只用于验证 migration/sqlc/transaction/readiness，不承载业务语义。
  - 在 `backend/sql/queries/` 中新增与 `platform_runtime_probes` 配套的最小查询，并配置 `sqlc` 生成代码链路；查询范围限定为 probe insert/count/delete（或等价最小集合）。
  - 新增 `sqlc.yaml`（若当前不存在）以及 migration / sqlc / integration 所需的最小配置文件。
  - 使 Task 6 的 Testcontainers 测试真正通过：启动 Postgres、执行 migration、执行 probe query、清理资源。

  **Must NOT do**:
  - 不创建 `orders`、`payments`、`inventory`、`notifications` 等业务表。
  - 不在应用启动路径自动执行 migration。
  - 不把 sqlc 扩展成每模块一套空 query 包。

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: 涉及多工具协作（pgx/goose/sqlc/Testcontainers）与真实 DB 运行链路。
  - **Skills**: [`test-driven-development`, `verification-before-completion`]
    - `test-driven-development`: 先让 Task 6 的 DB integration harness 失败，再逐步补齐链路。
    - `verification-before-completion`: 确保 migration/sqlc/integration 三条命令都各自有证据。
  - **Skills Evaluated but Omitted**:
    - `feature-research`: 技术栈已由 `backend/todo.md` 与用户锁定。

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 7, 8, 9, 11)
  - **Blocks**: 12, 13, 15, 18
  - **Blocked By**: 1, 6, 9

  **References**:
  - `backend/internal/platform/database/db.go:1-21` - 当前 DB 抽象起点。
  - `backend/migrations/` - 当前空目录，后续 migration 必须放这里。
  - `backend/sql/queries/` - 当前空目录，后续 query fixture 必须放这里。
  - `backend/test/integration/ping_test.go:22-88` - readiness 集成测试会复用相同装配方式。
  - `backend/Makefile:1-13` - 需要扩展 sqlc/migration/integration 命令入口。
  - `backend/configs/config.test.yaml` - Testcontainers / readiness 会依赖测试配置。

  **API/Type References**:
  - `backend/internal/bootstrap/server.go:22-30` - readiness 最终以 `deps.DB.Ping(ctx)` 为入口。
  - `backend/internal/bootstrap/app.go:13-19` - 新 DB 需要纳入共享依赖容器与 shutdown 生命周期。

  **Test References**:
  - Task 6 产出的 Testcontainers harness 是本任务首要验收入口。

  **External References**:
  - `https://github.com/jackc/pgx`
  - `https://github.com/pressly/goose`
  - `https://sqlc.dev/`
  - `https://golang.testcontainers.org/`

  **WHY Each Reference Matters**:
  - 这是 P1 中“基础设施从占位到真实”的关键任务，但必须始终锚定“平台级非业务验证载体”，否则很容易滑向真实领域建模。

  **Acceptance Criteria**:
  - [ ] `database.DB` 不再只是 `DummyDB`，具备真实 `Ping` 与 shutdown 能力。
  - [ ] migration 仅创建 `platform_runtime_probes`（或等价平台级 probe 表）及其最小辅助对象，不出现任何业务表。
  - [ ] `sql/queries/` 仅包含 probe insert/count/delete（或等价最小集合），不出现业务 query。
  - [ ] `cd backend && make sqlc-generate` → PASS。
  - [ ] `cd backend && make migrate-up` → PASS（针对目标环境或测试环境约定）。
  - [ ] `cd backend && go test ./test/integration -run TestDatabaseGroundwork -count=1` → PASS。
  - [ ] 整个链路只使用平台级 probe schema/query，无真实业务表。

  **QA Scenarios (MANDATORY)**:

  ```text
  Scenario: DB groundwork happy path - migration + sqlc + probe query 全链路通过
    Tool: Bash
    Preconditions: pgx 连接池、migration、sqlc 配置、probe 表/query 已实现
    Steps:
      1. 运行 `cd backend && make sqlc-generate`
      2. 运行 `cd backend && go test ./test/integration -run TestDatabaseGroundwork -count=1`
      3. 断言两条命令都返回 0，并保存组合输出到 `.sisyphus/evidence/task-10-db-happy.txt`
    Expected Result: 代码生成成功，容器集成测试通过，probe query 可执行
    Failure Indicators: sqlc 生成失败；migration 不可执行；query 不通；返回码非 0
    Evidence: .sisyphus/evidence/task-10-db-happy.txt

  Scenario: DB groundwork failure path - 错误 migration 或连接异常可被检测
    Tool: Bash (go test)
    Preconditions: 已存在 migration/container 负向测试
    Steps:
      1. 运行 `cd backend && go test ./test/integration -run 'TestDatabaseGroundworkFailsOnBrokenMigration|TestDatabaseGroundworkHandlesContainerDependencyFailure' -count=1`
      2. 断言测试返回 0，表明负向断言成立
      3. 保存输出到 `.sisyphus/evidence/task-10-db-failure.txt`
    Expected Result: 失败原因被测试准确检测，而不是被忽略
    Failure Indicators: 异常场景未覆盖；测试静默 skip；返回码非 0
    Evidence: .sisyphus/evidence/task-10-db-failure.txt
  ```

  **Evidence to Capture**:
  - [ ] `.sisyphus/evidence/task-10-db-happy.txt`
  - [ ] `.sisyphus/evidence/task-10-db-failure.txt`

  **Commit**: YES
  - Message: `feat(database): 接入真实数据库地基能力`
  - Files: `backend/internal/platform/database/*`, `backend/migrations/*`, `backend/sql/queries/*`, `backend/sqlc.yaml`, `backend/configs/*`
  - Pre-commit: `cd backend && make sqlc-generate && go test ./test/integration -run TestDatabaseGroundwork -count=1`

- [x] 11. 接入 OTEL provider / propagator / tracer / shutdown 真实实现

  **What to do**:
  - 在 `backend/internal/platform/observability/` 中把 noop tracer / propagator 升级为可配置的真实 OTEL provider 与 propagator，实现 provider 初始化、propagator 注册、tracer 获取、shutdown flush。
  - 保留当前 `request_id` / `trace_id` context helper 仍可作为日志与响应层读取来源，但不强制维持完整 vendor-agnostic tracing façade；优先采用贴近 OpenTelemetry 原生接线方式的 provider / middleware / propagator 组织。
  - 明确无 exporter 配置时的降级行为；保证服务不会因观测后端不可达而非预期崩溃。
  - 为 HTTP、DB、worker 路径后续接入 tracer 预留统一入口，但不在本任务里扩展 metrics/dashboard。

  **Must NOT do**:
  - 不引入 metrics/dashboard/collector 部署。
  - 不把 exporter endpoint 或 credential 硬编码进代码。

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: OTEL provider 生命周期与降级行为属于长期架构接缝，错误代价高。
  - **Skills**: [`test-driven-development`, `verification-before-completion`]
    - `test-driven-development`: 依据 Task 5 的 observability 契约测试逐项实现。
    - `verification-before-completion`: 提交前必须确认 init/propagation/shutdown/degrade 四类场景都有证据。
  - **Skills Evaluated but Omitted**:
    - `systematic-debugging`: 当前仍是正向实现阶段。

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 7, 8, 9, 10)
  - **Blocks**: 12, 15, 20
  - **Blocked By**: 5

  **References**:
  - `backend/internal/platform/observability/context.go:1-28`
  - `backend/internal/platform/observability/propagator.go:5-51`
  - `backend/internal/platform/observability/tracer.go:5-25`
  - `backend/internal/platform/observability/observability_test.go:8-36`
  - `backend/internal/platform/logger/logger.go:65-85`
  - `backend/cmd/api/main.go:28-35` - 当前 API 入口仍装配 `NewNoopTracer()` / `NewNoopPropagator()`。

  **WHY Each Reference Matters**:
  - 当前 API/worker 启动链路已经预留了 observability 依赖位，但还是占位实现；这是在不破坏现有装配结构的前提下升级到真实 provider 的最佳切点。

  **Acceptance Criteria**:
  - [ ] `cmd/api` / `cmd/worker` 可通过新 provider/propagator 初始化路径工作（具体装配在后续 bootstrap task 完成）。
  - [ ] `cd backend && go test ./internal/platform/observability -count=1` → PASS。
  - [ ] shutdown 能调用 provider flush / close，且有自动化测试覆盖。
  - [ ] exporter 缺失或不可达时行为符合既定降级策略。

  **QA Scenarios (MANDATORY)**:

  ```text
  Scenario: OTEL happy path - provider 初始化与 shutdown 闭环
    Tool: Bash (go test)
    Preconditions: observability 真实实现已接入测试
    Steps:
      1. 运行 `cd backend && go test ./internal/platform/observability -run 'TestObservabilityProviderInitializes|TestObservabilityShutdownFlushesProvider' -count=1`
      2. 断言输出包含 `ok` 且返回码为 0
      3. 保存输出到 `.sisyphus/evidence/task-11-otel-happy.txt`
    Expected Result: provider 能初始化并在 shutdown 时 flush/close
    Failure Indicators: init 失败；shutdown 未执行；返回码非 0
    Evidence: .sisyphus/evidence/task-11-otel-happy.txt

  Scenario: OTEL failure path - exporter 异常时安全降级
    Tool: Bash (go test)
    Preconditions: 已添加 provider degrade 负向测试
    Steps:
      1. 运行 `cd backend && go test ./internal/platform/observability -run TestObservabilityProviderDegradesGracefully -count=1`
      2. 断言输出包含 `ok` 且返回码为 0
      3. 保存输出到 `.sisyphus/evidence/task-11-otel-failure.txt`
    Expected Result: exporter 不可达或未配置时服务按约定降级，而非崩溃
    Failure Indicators: panic；测试无法验证降级；返回码非 0
    Evidence: .sisyphus/evidence/task-11-otel-failure.txt
  ```

  **Evidence to Capture**:
  - [ ] `.sisyphus/evidence/task-11-otel-happy.txt`
  - [ ] `.sisyphus/evidence/task-11-otel-failure.txt`

  **Commit**: YES
  - Message: `feat(observability): 接入可关闭的链路追踪提供者`
  - Files: `backend/internal/platform/observability/*`, `backend/internal/platform/config/*`
  - Pre-commit: `cd backend && go test ./internal/platform/observability -count=1`

- [x] 12. 接入 API / worker bootstrap 装配与 graceful shutdown 闭环

  **What to do**:
  - 更新 `backend/cmd/api/main.go` 和 `backend/cmd/worker/main.go`，不再装配 noop tracer / dummy DB，而是使用真实 config-driven platform 初始化结果。
  - 确保 `backend/internal/bootstrap/app.go` 聚合 DB、OTEL provider、HTTP server、worker 的 shutdown 顺序与错误聚合。
  - 保持 `cmd/*` 只做装配，不新增业务逻辑。
  - 如有必要，扩展 `internal/bootstrap` 测试，验证启动依赖注入和 shutdown 聚合行为。

  **Must NOT do**:
  - 不在 `cmd/api` 或 `cmd/worker` 写业务规则。
  - 不在此任务自动执行 migration。

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: 启动装配是所有 platform 能力汇合点，改坏会影响 API、worker、shutdown 全链路。
  - **Skills**: [`test-driven-development`, `verification-before-completion`]
    - `test-driven-development`: 先让 bootstrap/app/server/worker 相关测试暴露缺失依赖注入或 shutdown 聚合问题。
    - `verification-before-completion`: 提交前必须核对 API 与 worker 两个入口都能构造并关闭资源。
  - **Skills Evaluated but Omitted**:
    - `using-git-worktrees`: 仅写计划，不涉及实际隔离执行。

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 13, 14, 15, 16)
  - **Blocks**: 13, 16, 20
  - **Blocked By**: 7, 8, 10, 11

  **References**:
  - `backend/cmd/api/main.go:22-64`
  - `backend/cmd/worker/main.go:27-53`
  - `backend/internal/bootstrap/app.go:13-51`
  - `backend/internal/bootstrap/server.go:15-37`
  - `backend/internal/bootstrap/worker.go:11-99`

  **WHY Each Reference Matters**:
  - 当前两个入口都已经是“真实装配位置”；P1 目标不是重构入口结构，而是把真实 platform 依赖塞进现有装配骨架，并补齐关闭链路。

  **Acceptance Criteria**:
  - [ ] API/worker 入口都使用真实 DB / OTEL 依赖初始化路径。
  - [ ] `cd backend && go test ./internal/bootstrap ./cmd/... -count=1` → PASS（如 cmd 不适合直接测，则至少 bootstrap tests PASS 且有构造测试覆盖）。
  - [ ] shutdown 聚合可关闭 DB / provider / HTTP server / worker，错误可被 join 返回。

  **QA Scenarios (MANDATORY)**:

  ```text
  Scenario: 装配 happy path - API/worker 使用真实依赖构造成功
    Tool: Bash (go test)
    Preconditions: bootstrap/cmd 装配测试已存在
    Steps:
      1. 运行 `cd backend && go test ./internal/bootstrap -count=1`
      2. 若为入口增加可测试构造函数，再运行相应测试命令
      3. 保存输出到 `.sisyphus/evidence/task-12-bootstrap-happy.txt`
    Expected Result: bootstrap 与依赖注入测试全部通过
    Failure Indicators: 入口仍装配 noop/dummy；shutdowner 未纳入聚合；返回码非 0
    Evidence: .sisyphus/evidence/task-12-bootstrap-happy.txt

  Scenario: 装配 failure path - shutdown 聚合正确暴露错误
    Tool: Bash (go test)
    Preconditions: 已为 `App.Shutdown` / worker / provider 失败场景补测试
    Steps:
      1. 运行 `cd backend && go test ./internal/bootstrap -run 'TestAppShutdownAggregatesErrors|TestWorkerShutdownHandlesFailures' -count=1`
      2. 断言输出包含 `ok` 且返回码为 0
      3. 保存输出到 `.sisyphus/evidence/task-12-bootstrap-failure.txt`
    Expected Result: 多个 shutdown 错误被正确聚合和返回
    Failure Indicators: shutdown 错误被吞掉；worker 资源泄漏；返回码非 0
    Evidence: .sisyphus/evidence/task-12-bootstrap-failure.txt
  ```

  **Evidence to Capture**:
  - [ ] `.sisyphus/evidence/task-12-bootstrap-happy.txt`
  - [ ] `.sisyphus/evidence/task-12-bootstrap-failure.txt`

  **Commit**: YES
  - Message: `feat(bootstrap): 接入真实平台依赖装配链路`
  - Files: `backend/cmd/api/main.go`, `backend/cmd/worker/main.go`, `backend/internal/bootstrap/*`
  - Pre-commit: `cd backend && go test ./internal/bootstrap -count=1`

- [x] 13. 把 /readyz 改造成真实 DB readiness，保持 /healthz 仅表明进程存活

  **What to do**:
  - 在 `backend/internal/bootstrap/server.go` 中落实：`/healthz` 只检查进程活着，不依赖 DB；`/readyz` 通过真实 `deps.DB.Ping(ctx)` 判定 readiness。
  - 调整 `readyz` 的成功 payload / 失败 payload，使其继续走统一响应结构，并带顶层 `request_id`、`trace_id`。
  - 扩展 `backend/internal/bootstrap/server_test.go` 与集成测试，覆盖 DB 可用与不可用两类 readiness 结果。

  **Must NOT do**:
  - 不把 `/healthz` 与 `/readyz` 混成同一语义。
  - 不在健康检查里引入业务级检查。

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 变更集中在 bootstrap health 路径与测试，但语义很关键。
  - **Skills**: [`test-driven-development`, `verification-before-completion`]
    - `test-driven-development`: 先写 DB 可用/不可用两类健康检查测试。
    - `verification-before-completion`: 强制核对 health 与 readiness 语义没有倒置。
  - **Skills Evaluated but Omitted**:
    - `feature-research`: 语义已由用户锁死。

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 12, 14, 15, 16)
  - **Blocks**: 16, 17
  - **Blocked By**: 10, 12

  **References**:
  - `backend/internal/bootstrap/server.go:19-30`
  - `backend/internal/bootstrap/server_test.go:17-52`
  - `backend/internal/platform/response/response.go:11-26`
  - `backend/test/integration/ping_test.go:22-88`

  **WHY Each Reference Matters**:
  - 当前 `readyz` 已经走 `deps.DB.Ping`，但 payload 仍写着 `dependencies: skipped`；P1 需要把这条路径从占位语义变成真实 readiness 契约。

  **Acceptance Criteria**:
  - [ ] `/healthz` 与 `/readyz` 语义分离清晰。
  - [ ] DB 不可用时 `/readyz` 返回 503 且统一响应顶层含 `request_id`、`trace_id`。
  - [ ] `cd backend && go test ./internal/bootstrap -run TestNewAPIEngineRegistersHealthRoutes -count=1` → PASS。

  **QA Scenarios (MANDATORY)**:

  ```text
  Scenario: readiness happy path - DB 可用时返回 200
    Tool: Bash (go test)
    Preconditions: 已为 DB 可用场景补 server 测试
    Steps:
      1. 运行 `cd backend && go test ./internal/bootstrap -run TestReadyzReturnsOKWhenDBHealthy -count=1`
      2. 断言输出包含 `ok` 且返回码为 0
      3. 保存输出到 `.sisyphus/evidence/task-13-readyz-happy.txt`
    Expected Result: `/readyz` 在 DB healthy 时返回 200 且顶层字段完整
    Failure Indicators: readiness 仍返回 skipped；trace/request 字段缺失；返回码非 0
    Evidence: .sisyphus/evidence/task-13-readyz-happy.txt

  Scenario: readiness failure path - DB 不可用时返回 503
    Tool: Bash (go test)
    Preconditions: 已为 DB ping 失败场景补测试 stub
    Steps:
      1. 运行 `cd backend && go test ./internal/bootstrap -run TestReadyzReturnsServiceUnavailableWhenDBFails -count=1`
      2. 断言输出包含 `ok` 且返回码为 0
      3. 保存输出到 `.sisyphus/evidence/task-13-readyz-failure.txt`
    Expected Result: `/readyz` 返回 503 且统一错误响应带 `request_id`/`trace_id`
    Failure Indicators: readiness 仍返回 200；错误响应不统一；返回码非 0
    Evidence: .sisyphus/evidence/task-13-readyz-failure.txt
  ```

  **Evidence to Capture**:
  - [ ] `.sisyphus/evidence/task-13-readyz-happy.txt`
  - [ ] `.sisyphus/evidence/task-13-readyz-failure.txt`

  **Commit**: YES
  - Message: `feat(bootstrap): 区分健康检查与数据库就绪语义`
  - Files: `backend/internal/bootstrap/server.go`, `backend/internal/bootstrap/server_test.go`, `backend/test/integration/*`
  - Pre-commit: `cd backend && go test ./internal/bootstrap -run 'TestReadyz|TestHealthz' -count=1`

- [x] 14. 四模块骨架统一接入 handler/service/repository 新契约

  **What to do**:
  - 在 `backend/internal/modules/{order,payment,inventory,notification}/` 中统一接入新的 handler/service/repository 构造风格：handler 用统一 bind/validate 与 response；service 持有 repository + `TxManager`；repository 仅依赖 DB executor/CRUD 接口。
  - 对四个模块都做对称更新，禁止只改一个模块再“口头继承”。
  - 去掉模块对 payload 中 `trace_id` 的顶层职责依赖，让顶层响应壳承担主 trace contract；如果 payload 内暂保留字段，仅能作为兼容过渡。
  - repository 形态仅提升到“兼容统一 executor/transaction seam 的最小骨架”，不得借此扩成真实 CRUD 仓储。
  - 为模块层补最薄必要测试，证明四模块都接上新契约，但不扩成真实业务流程测试矩阵。

  **Must NOT do**:
  - 不新增真实业务 endpoint。
  - 不把四模块抽成通用 base class / 生成器。
  - 不在 repository 引入业务逻辑。

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: 要在四个同构模块上做一致化修改，易出现漏改与不对称。
  - **Skills**: [`test-driven-development`, `verification-before-completion`]
    - `test-driven-development`: 先写至少一组模块对称性测试或 table-driven regression。
    - `verification-before-completion`: 提交前逐个模块核对 handler/service/repository 签名与行为一致。
  - **Skills Evaluated but Omitted**:
    - `dispatching-parallel-agents`: 执行时可用，但本计划中任务本身仍保持单一职责描述。

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 12, 13, 15, 16)
  - **Blocks**: 16, 17
  - **Blocked By**: 3, 8, 9

  **References**:
  - `backend/internal/modules/order/handler.go:12-34`
  - `backend/internal/modules/order/service.go:5-22`
  - `backend/internal/modules/order/repository.go:5-17`
  - `backend/internal/modules/payment/handler.go:12-34`
  - `backend/internal/modules/inventory/handler.go:12-34`
  - `backend/internal/modules/notification/handler.go:12-34`
  - `backend/internal/modules/AGENTS.md` - 明确模块依赖顺序与反模式。

  **WHY Each Reference Matters**:
  - 四模块目前已经高度同构；这既是优势，也是风险点。必须用统一契约把这种同构固化，而不是放任它们逐渐漂移。

  **Acceptance Criteria**:
  - [ ] 四个模块都已接入统一构造风格与公共接缝。
  - [ ] `cd backend && go test ./internal/modules/... -count=1` → PASS（如存在对应测试）。
  - [ ] 任一模块不再需要自行决定 trace 顶层输出。
  - [ ] 未新增任何真实业务 endpoint 或 CRUD。

  **QA Scenarios (MANDATORY)**:

  ```text
  Scenario: 模块对齐 happy path - 四模块骨架全部接入新契约
    Tool: Bash (go test)
    Preconditions: 已为模块骨架补 table-driven 或对称性测试
    Steps:
      1. 运行 `cd backend && go test ./internal/modules/... -count=1`
      2. 断言输出包含 `ok` 且返回码为 0
      3. 保存输出到 `.sisyphus/evidence/task-14-modules-happy.txt`
    Expected Result: 四模块 handler/service/repository 对齐测试通过
    Failure Indicators: 任一模块漏改；签名不一致；返回码非 0
    Evidence: .sisyphus/evidence/task-14-modules-happy.txt

  Scenario: 模块对齐 failure path - 不允许单模块漂移
    Tool: Bash (go test)
    Preconditions: 已添加用于检测不对称行为的表驱动测试或 compile-time 契约测试，测试命名统一包含 `Symmetry` 或 `Contract`
    Steps:
      1. 运行 `cd backend && go test ./internal/modules/... -run 'Test.*Symmetry|Test.*Contract' -count=1`
      2. 断言输出包含 `ok` 且返回码为 0
      3. 保存输出到 `.sisyphus/evidence/task-14-modules-failure.txt`
    Expected Result: 测试能防止只改单模块、其他模块未同步的情况
    Failure Indicators: 测试无法发现模块间漂移；契约测试未覆盖四模块；返回码非 0
    Evidence: .sisyphus/evidence/task-14-modules-failure.txt
  ```

  **Evidence to Capture**:
  - [ ] `.sisyphus/evidence/task-14-modules-happy.txt`
  - [ ] `.sisyphus/evidence/task-14-modules-failure.txt`

  **Commit**: YES
  - Message: `refactor(modules): 统一四模块骨架接缝`
  - Files: `backend/internal/modules/order/*`, `backend/internal/modules/payment/*`, `backend/internal/modules/inventory/*`, `backend/internal/modules/notification/*`
  - Pre-commit: `cd backend && go test ./internal/modules/... -count=1`

- [ ] 15. 打通日志关联与 DB tracing 主链路闭环

  **What to do**:
  - 让 `backend/internal/platform/logger/logger.go` 与真实 observability provider 协同，把 request_id / trace_id 稳定打入日志字段。
  - 在 DB 层为关键 query / transaction 生命周期挂 span 或基础 attributes，满足 P1“traces + log correlation”闭环。
  - 保证 API 响应顶层 `trace_id`、日志 `trace_id`、DB tracing 使用的是同一 trace lineage。
  - 补充 platform/bootstrap/integration 测试，验证日志关联与 trace context 不脱节。

  **Must NOT do**:
  - 不做自定义 metrics。
  - 不在业务模块里手写散落 tracing 逻辑。

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: 涉及 logger、observability、database 三层协作，需要跨包验证同一 trace lineage。
  - **Skills**: [`test-driven-development`, `verification-before-completion`]
    - `test-driven-development`: 先写 context/log/tracing 协作测试，再落实现。
    - `verification-before-completion`: 防止只做日志字段不做 DB span，或只做 DB span 不做日志关联。
  - **Skills Evaluated but Omitted**:
    - `feature-research`: 目标范围已锁定。

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 12, 13, 14, 16)
  - **Blocks**: 16, 20
  - **Blocked By**: 5, 9, 10, 11

  **References**:
  - `backend/internal/platform/logger/logger.go:32-85`
  - `backend/internal/platform/observability/context.go:12-28`
  - `backend/internal/platform/observability/tracer.go:5-25`
  - `backend/internal/platform/database/db.go:1-21`
  - `backend/cmd/api/main.go:28-35`

  **WHY Each Reference Matters**:
  - 目前 logger 已能从 context 读 request/trace，但底层 trace provider 与 DB tracing 还没接上；这是把三者串成同一主链路的关键一步。

  **Acceptance Criteria**:
  - [ ] `logger.WithContext(...)` 在真实请求链路中稳定输出 request_id / trace_id。
  - [ ] DB 操作带基础 tracing 证据或可测试属性。
  - [ ] `cd backend && go test ./internal/platform/... ./internal/bootstrap -count=1` → PASS。

  **QA Scenarios (MANDATORY)**:

  ```text
  Scenario: 关联 happy path - 日志与 DB tracing 共享同一 trace 上下文
    Tool: Bash (go test)
    Preconditions: 已添加 logger/DB tracing 协作测试
    Steps:
      1. 运行 `cd backend && go test ./internal/platform/... -run 'TestLoggerWithContextIncludesRequestAndTraceID|TestDatabaseTracingUsesRequestContext' -count=1`
      2. 断言输出包含 `ok` 且返回码为 0
      3. 保存输出到 `.sisyphus/evidence/task-15-correlation-happy.txt`
    Expected Result: request trace 能同时出现在日志字段与 DB tracing 链路中
    Failure Indicators: trace 只出现在一层；context 丢失；返回码非 0
    Evidence: .sisyphus/evidence/task-15-correlation-happy.txt

  Scenario: 关联 failure path - 缺失上下文时行为可预测
    Tool: Bash (go test)
    Preconditions: 已添加缺失 trace/request context 的负向测试
    Steps:
      1. 运行 `cd backend && go test ./internal/platform/... -run 'TestLoggerWithContextHandlesMissingMetadata|TestDatabaseTracingHandlesMissingTraceContext' -count=1`
      2. 断言输出包含 `ok` 且返回码为 0
      3. 保存输出到 `.sisyphus/evidence/task-15-correlation-failure.txt`
    Expected Result: 缺失上下文时行为符合既定降级策略，不 panic、不写脏数据
    Failure Indicators: panic；trace 伪造不稳定；返回码非 0
    Evidence: .sisyphus/evidence/task-15-correlation-failure.txt
  ```

  **Evidence to Capture**:
  - [ ] `.sisyphus/evidence/task-15-correlation-happy.txt`
  - [ ] `.sisyphus/evidence/task-15-correlation-failure.txt`

  **Commit**: YES
  - Message: `feat(observability): 打通日志与数据库追踪闭环`
  - Files: `backend/internal/platform/logger/*`, `backend/internal/platform/observability/*`, `backend/internal/platform/database/*`
  - Pre-commit: `cd backend && go test ./internal/platform/... -count=1`

- [ ] 16. 扩展集成测试矩阵到顶层 trace_id、readiness、ping 与 DB 主链路

  **What to do**:
  - 扩展 `backend/test/integration/ping_test.go`，必要时拆分为 `backend/test/integration/health_test.go`、`backend/test/integration/readiness_test.go`、`backend/test/integration/ping_test.go`，分别覆盖 `/healthz`、`/readyz`、四模块 `/ping`、顶层 `request_id/trace_id`、header 透传、DB readiness。
  - 把集成测试明确拆成两类：无需容器的 bootstrap/in-process 集成测试，以及依赖 Testcontainers 的 DB groundwork 集成测试；并与 `make test` / `make test-integration` 保持一致。
  - 所有集成测试都必须只验证平台与模块骨架契约，不依赖任何真实业务表或真实业务流程。

  **Must NOT do**:
  - 不把所有断言继续塞在单个超大测试文件里。
  - 不要求先手工启动服务再跑测试。

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 主要是扩展现有 integration test 组织与断言矩阵。
  - **Skills**: [`test-driven-development`, `verification-before-completion`]
    - `test-driven-development`: 先让新增集成断言失败，再补齐上游实现。
    - `verification-before-completion`: 确保 health/readiness/ping/trace/header 四类链路都被覆盖。
  - **Skills Evaluated but Omitted**:
    - `dispatching-parallel-agents`: 改动集中在 integration 层，单任务即可完成。

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 12, 13, 14, 15)
  - **Blocks**: F1, F2, F3, F4
  - **Blocked By**: 2, 6, 7, 12, 13, 14, 15

  **References**:
  - `backend/test/integration/ping_test.go:22-88` - 当前 integration test 基线，后续要按职责拆分并扩展断言。
  - `backend/internal/bootstrap/server_test.go:17-52` - health/readiness 单元侧基线；集成侧应与此语义对齐。
  - `backend/internal/bootstrap/server.go:19-37` - 实际 health/readiness 路由与响应入口。
  - `backend/internal/platform/response/response.go:11-26` - 集成测试最终解码的统一响应壳。

  **WHY Each Reference Matters**:
  - 当前 integration test 已经是最接近真实 API 装配链路的验证方式；P1 最终是否成立，取决于这里能否完整覆盖公共契约而不越界到业务流程。

  **Acceptance Criteria**:
  - [ ] `cd backend && go test ./test/integration -count=1` → PASS。
  - [ ] 集成测试显式断言顶层 `trace_id`、`request_id`、header 透传、`/readyz` DB 语义。
  - [ ] 无任何集成测试依赖真实业务 schema。

  **QA Scenarios (MANDATORY)**:

  ```text
  Scenario: 集成 happy path - health/readiness/ping 公共契约通过
    Tool: Bash (go test)
    Preconditions: integration tests 已扩展完成
    Steps:
      1. 运行 `cd backend && go test ./test/integration -count=1`
      2. 断言输出包含 `ok` 且返回码为 0
      3. 保存输出到 `.sisyphus/evidence/task-16-integration-happy.txt`
    Expected Result: `/healthz`、`/readyz`、四模块 `/ping` 全部通过统一响应和 tracing 断言
    Failure Indicators: 顶层 `trace_id` 缺失；header 未透传；readiness 语义错误；返回码非 0
    Evidence: .sisyphus/evidence/task-16-integration-happy.txt

  Scenario: 集成 failure path - DB 不可用时 readiness 正确失败
    Tool: Bash (go test)
    Preconditions: 已有 DB failure stub 或容器异常场景的 readiness 集成测试
    Steps:
      1. 运行 `cd backend && go test ./test/integration -run TestReadyzFailsWhenDBUnavailable -count=1`
      2. 断言输出包含 `ok` 且返回码为 0
      3. 保存输出到 `.sisyphus/evidence/task-16-integration-failure.txt`
    Expected Result: readiness 失败路径被集成测试覆盖
    Failure Indicators: 测试仍返回 200；错误响应字段不完整；返回码非 0
    Evidence: .sisyphus/evidence/task-16-integration-failure.txt
  ```

  **Evidence to Capture**:
  - [ ] `.sisyphus/evidence/task-16-integration-happy.txt`
  - [ ] `.sisyphus/evidence/task-16-integration-failure.txt`

  **Commit**: YES
  - Message: `test(integration): 扩展平台公共契约回归矩阵`
  - Files: `backend/test/integration/*`
  - Pre-commit: `cd backend && go test ./test/integration -count=1`

- [ ] 17. 更新 OpenAPI：统一响应结构 + health/readiness + 四模块 ping 契约

  **What to do**:
  - 更新 `backend/api/openapi.yaml`，补齐统一响应 schema，顶层明确 `code / message / data / request_id / trace_id`。
  - 为 `/healthz`、`/readyz`、`/api/v1/order/ping`、`/api/v1/payment/ping`、`/api/v1/inventory/ping`、`/api/v1/notification/ping` 建立最小但完整的 P1 文档示例。
  - 对 `/readyz` 明确 200 与 503 两类响应语义。
  - 严格只描述当前已有公共接口，不预埋未来真实业务接口 schema。

  **Must NOT do**:
  - 不为未来 CRUD/业务流程预先建模。
  - 不让 OpenAPI 脱离当前真实实现。

  **Recommended Agent Profile**:
  - **Category**: `writing`
    - Reason: 以契约文档精确表达为主，需要高一致性和边界克制。
  - **Skills**: [`verification-before-completion`]
    - `verification-before-completion`: 文档必须逐项对照真实路由与响应结构。
  - **Skills Evaluated but Omitted**:
    - `test-driven-development`: 文档任务通过对照实现与测试结果完成即可。

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with Tasks 18, 19, 20)
  - **Blocks**: F1, F4
  - **Blocked By**: 7, 13, 14

  **References**:
  - `backend/api/openapi.yaml:1-17` - 当前 OpenAPI 基线，只有 `/healthz` 与 `/readyz`。
  - `backend/internal/bootstrap/server.go:19-37` - health/readiness 实际路由来源。
  - `backend/internal/modules/order/handler.go:20-34`
  - `backend/internal/modules/payment/handler.go:20-34`
  - `backend/internal/modules/inventory/handler.go:20-34`
  - `backend/internal/modules/notification/handler.go:20-34`
  - `backend/internal/platform/response/response.go:11-26` - 统一响应 schema 的真实来源。

  **WHY Each Reference Matters**:
  - OpenAPI 当前极度精简；P1 要求它成为“当前公共接口真实契约”，而不是未来业务接口的空壳模板。

  **Acceptance Criteria**:
  - [ ] `api/openapi.yaml` 覆盖 health/readiness/四模块 ping。
  - [ ] 所有响应 schema 顶层含 `request_id`、`trace_id`。
  - [ ] `/readyz` 文档中同时描述 200 和 503。
  - [ ] 文档内容与当前测试过的真实接口行为一致。

  **QA Scenarios (MANDATORY)**:

  ```text
  Scenario: OpenAPI happy path - 文档覆盖当前全部公共接口
    Tool: Bash + Read + Grep
    Preconditions: openapi.yaml 已更新
    Steps:
      1. 使用 `read` 读取 `backend/api/openapi.yaml`
      2. 使用 `grep` 检查文档中存在 `/healthz`、`/readyz`、`/api/v1/order/ping`、`/api/v1/payment/ping`、`/api/v1/inventory/ping`、`/api/v1/notification/ping`
      3. 使用 `grep` 检查统一响应 schema 中存在 `request_id`、`trace_id`，并检查 `/readyz` 同时描述 `200` 与 `503`
      4. 使用 `bash` 汇总检查结果并保存到 `.sisyphus/evidence/task-17-openapi-happy.txt`
    Expected Result: health/readiness/四模块 ping 全部被文档覆盖，响应壳字段完整，`/readyz` 的 200/503 语义齐全
    Failure Indicators: 漏路径；漏 `503`；缺少 `trace_id`；schema 与真实契约不一致
    Evidence: .sisyphus/evidence/task-17-openapi-happy.txt

  Scenario: OpenAPI failure path - 拒绝未来业务接口越界建模
    Tool: Bash + Read + Grep
    Preconditions: 仅允许记录当前公共接口
    Steps:
      1. 使用 `read` 读取 `backend/api/openapi.yaml`
      2. 使用 `grep` 检查文档中不存在未来 CRUD 路径模式，如 `/orders`、`/payments`、`/inventory/items`、`/notifications/send`
      3. 使用 `grep` 检查文档中不存在真实业务 schema 关键字，如 `CreateOrderRequest`、`PaymentDetail`、`InventoryItem`
      4. 使用 `bash` 汇总检查结果并保存到 `.sisyphus/evidence/task-17-openapi-failure.txt`
    Expected Result: 文档严格停留在当前 P1 公共接口
    Failure Indicators: 出现任何未来业务接口或真实业务 schema；越界项无法自动命中
    Evidence: .sisyphus/evidence/task-17-openapi-failure.txt
  ```

  **Evidence to Capture**:
  - [ ] `.sisyphus/evidence/task-17-openapi-happy.txt`
  - [ ] `.sisyphus/evidence/task-17-openapi-failure.txt`

  **Commit**: YES
  - Message: `docs(openapi): 补齐健康检查与模块 ping 契约`
  - Files: `backend/api/openapi.yaml`
  - Pre-commit: 记录路由对照证据后再提交

- [ ] 18. 扩展 Makefile 与显式命令入口

  **What to do**:
  - 在 `backend/Makefile` 中补齐 `sqlc-generate`、`migrate-up`、`migrate-down`、`test-integration` 等显式命令入口。
  - 保持 `make test` 只覆盖快速回归，`make test-integration` 覆盖容器化集成路径，命令语义清晰不混淆。
  - 如需要辅助命令，优先保持 Makefile 自解释，不引入复杂脚本系统。

  **Must NOT do**:
  - 不让 `make test` 强制依赖 Docker。
  - 不把 migration 自动绑到 `run-api` 或 `run-worker`。

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 主要是工具入口扩展与命令语义收口。
  - **Skills**: [`verification-before-completion`]
    - `verification-before-completion`: 每个新增命令都必须跑过并有输出证据。
  - **Skills Evaluated but Omitted**:
    - `test-driven-development`: Makefile 任务以命令验证为主。

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with Tasks 17, 19, 20)
  - **Blocks**: F2, F3
  - **Blocked By**: 1, 10

  **References**:
  - `backend/Makefile:1-13` - 当前命令入口过薄，是本任务直接修改对象。
  - `backend/migrations/` - migration 命令目标目录。
  - `backend/sql/queries/` - sqlc 生成链路的查询来源目录。
  - `backend/test/integration/` - `test-integration` 目标来源。

  **WHY Each Reference Matters**:
  - P1 的很多能力如果没有稳定命令入口，后续执行和审查都会非常脆弱。

  **Acceptance Criteria**:
  - [ ] `make sqlc-generate`、`make migrate-up`、`make test-integration` 可执行。
  - [ ] `make test` 与 `make test-integration` 分工清晰。
  - [ ] `cd backend && make test && make lint` 仍可通过。

  **QA Scenarios (MANDATORY)**:

  ```text
  Scenario: 命令 happy path - 新增 Make 入口全部可运行
    Tool: Bash
    Preconditions: Makefile 已扩展完成
    Steps:
      1. 运行 `cd backend && make test`
      2. 运行 `cd backend && make lint`
      3. 运行 `cd backend && make sqlc-generate`
      4. 运行 `cd backend && make test-integration`
      5. 保存组合输出到 `.sisyphus/evidence/task-18-make-happy.txt`
    Expected Result: 所有新增命令成功执行，返回码均为 0
    Failure Indicators: 命令缺失；语义混乱；依赖 Docker 的命令错误绑到 `make test`
    Evidence: .sisyphus/evidence/task-18-make-happy.txt

  Scenario: 命令 failure path - 错误 target 配置能被及时发现
    Tool: Bash + Read
    Preconditions: 已接入实际命令实现；若本地环境不适合直接执行 `make migrate-up`，则需至少保证 target 指向存在的脚本/命令并在测试环境中可执行
    Steps:
      1. 运行 `cd backend && make -n migrate-up`
      2. 运行 `cd backend && make -n migrate-down`
      3. 读取 `backend/Makefile`，核对 `migrate-up`、`migrate-down`、`sqlc-generate`、`test-integration` 的 target 与实际目录/工具链一致
      4. 若环境允许，再实际运行 `cd backend && make migrate-up` 以暴露 target 断链问题；若环境不允许，则在证据中明确原因并保留 dry-run 输出
      5. 保存检查结果到 `.sisyphus/evidence/task-18-make-failure.txt`
    Expected Result: Makefile target 与实际工具链一一对应，不存在指向不存在路径/命令的 target
    Failure Indicators: target 指向不存在文件/命令；dry-run 暴露错误；实际执行断链且未被记录
    Evidence: .sisyphus/evidence/task-18-make-failure.txt
  ```

  **Evidence to Capture**:
  - [ ] `.sisyphus/evidence/task-18-make-happy.txt`
  - [ ] `.sisyphus/evidence/task-18-make-failure.txt`

  **Commit**: YES
  - Message: `chore(tooling): 补齐数据库与集成测试命令入口`
  - Files: `backend/Makefile`
  - Pre-commit: `cd backend && make test && make lint`

- [ ] 19. 补齐 .golangci.yml 到 P1 要求并验证 lint 可跑

  **What to do**:
  - 将 `backend/.golangci.yml` 从当前 `errcheck/govet/staticcheck/unused` 扩展到至少包含用户要求的 `govet`、`errcheck`、`gofmt`、`gosimple`、`staticcheck`、`structcheck`。
  - 如果 `structcheck` 因 golangci-lint 版本原因不可用，必须选择最接近的等价方案，并在配置注释或计划证据中明确说明，而不是静默略过。
  - 保持 lint 配置与当前项目规模匹配，不一次性堆入大量与 P1 无关的 linter。

  **Must NOT do**:
  - 不为了“看起来严格”而加入大量无关 linter。
  - 不保留无效 linter 名称导致 `golangci-lint run` 无法执行。

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 配置文件单点修改，但必须兼顾工具版本兼容与可执行性。
  - **Skills**: [`verification-before-completion`]
    - `verification-before-completion`: 必须实际跑 `golangci-lint run` 并留证据。
  - **Skills Evaluated but Omitted**:
    - `test-driven-development`: lint 配置更适合直接用命令验证。

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with Tasks 17, 18, 20)
  - **Blocks**: F2
  - **Blocked By**: None

  **References**:
  - `backend/.golangci.yml:1-11` - 当前 lint 配置基线。
  - 用户原始 P1 规范 - 至少要求启用 `govet/errcheck/gofmt/gosimple/staticcheck/structcheck`。

  **WHY Each Reference Matters**:
  - 当前 lint 配置明显低于用户目标；但补齐方式必须兼顾实际工具版本，否则配置文件本身会失效。

  **Acceptance Criteria**:
  - [ ] `.golangci.yml` 覆盖 P1 要求的 linter 集。
  - [ ] `cd backend && golangci-lint run` → PASS。
  - [ ] 如存在版本兼容替代项，已在证据中说明原因与替代关系。

  **QA Scenarios (MANDATORY)**:

  ```text
  Scenario: lint happy path - 配置可执行且覆盖 P1 要求
    Tool: Bash
    Preconditions: `.golangci.yml` 已更新
    Steps:
      1. 运行 `cd backend && golangci-lint run`
      2. 断言返回码为 0
      3. 保存输出到 `.sisyphus/evidence/task-19-lint-happy.txt`
    Expected Result: lint 配置可执行，并覆盖约定的 P1 核心规则
    Failure Indicators: 未识别 linter；配置报错；返回码非 0
    Evidence: .sisyphus/evidence/task-19-lint-happy.txt

  Scenario: lint failure path - 无效 linter 配置能被识别并修正
    Tool: Bash
    Preconditions: 需要核对 linter 名称与版本兼容性
    Steps:
      1. 检查 `golangci-lint run` 输出是否包含未知 linter 或弃用警告
      2. 若有兼容性替代，记录替代说明
      3. 保存结果到 `.sisyphus/evidence/task-19-lint-failure.txt`
    Expected Result: 不存在静默失效的 linter 配置
    Failure Indicators: 未知 linter 未处理；配置假通过但实际未启用目标能力
    Evidence: .sisyphus/evidence/task-19-lint-failure.txt
  ```

  **Evidence to Capture**:
  - [ ] `.sisyphus/evidence/task-19-lint-happy.txt`
  - [ ] `.sisyphus/evidence/task-19-lint-failure.txt`

  **Commit**: YES
  - Message: `chore(lint): 补齐 P1 阶段静态检查规则`
  - Files: `backend/.golangci.yml`
  - Pre-commit: `cd backend && golangci-lint run`

- [ ] 20. 校准 worker 占位任务与 observability / shutdown / config 一致性

  **What to do**:
  - 更新 `backend/cmd/worker/main.go` 与 `backend/internal/bootstrap/worker.go`，让 worker 路径在配置、日志、OTEL provider、DB（如需要 probe 或 future hook）、shutdown 约定上与 API 入口保持一致。
  - 保持 worker 当前仍为 placeholder task，不引入真实异步业务，但要确保其生命周期管理符合 P1 foundation 标准。
  - 扩展 worker 测试，覆盖 task 注册、cancel 退出、probe、shutdown 与 observability/context 协作。

  **Must NOT do**:
  - 不新增真实异步任务。
  - 不让 worker 路径绕开统一 logger / observability / config 初始化。

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 变更集中在 worker 装配与测试一致性校准。
  - **Skills**: [`test-driven-development`, `verification-before-completion`]
    - `test-driven-development`: 先让 worker 生命周期与 observability 相关测试失败，再补实现。
    - `verification-before-completion`: 确认 worker 路径没有成为 P1 foundation 的遗漏角落。
  - **Skills Evaluated but Omitted**:
    - `feature-research`: worker 结构已很清晰。

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with Tasks 17, 18, 19)
  - **Blocks**: F1, F3
  - **Blocked By**: 11, 12, 15

  **References**:
  - `backend/cmd/worker/main.go:16-53` - 当前 worker 仍是 placeholder task 与基础 shutdown。
  - `backend/internal/bootstrap/worker.go:11-99` - worker 注册、probe、run/shutdown 的实现基线。
  - `backend/internal/bootstrap/app.go:25-51` - shutdown 聚合语义。
  - `backend/internal/platform/logger/logger.go:65-85` - worker 日志同样需要 context 关联。

  **WHY Each Reference Matters**:
  - 仓库是 API/Worker 双入口；如果只让 API 达到 P1 foundation 标准，worker 仍停留在旧占位装配，将形成后续隐患。

  **Acceptance Criteria**:
  - [ ] worker 保持 placeholder 语义，但装配路径与 shutdown/observability/config 约定与 API 一致。
  - [ ] `cd backend && go test ./internal/bootstrap -run 'TestWorker' -count=1` → PASS。
  - [ ] worker 取消退出、probe、shutdown 错误聚合均有测试覆盖。

  **QA Scenarios (MANDATORY)**:

  ```text
  Scenario: worker happy path - placeholder 任务可注册、运行、退出
    Tool: Bash (go test)
    Preconditions: worker tests 已扩展
    Steps:
      1. 运行 `cd backend && go test ./internal/bootstrap -run TestWorker -count=1`
      2. 断言输出包含 `ok` 且返回码为 0
      3. 保存输出到 `.sisyphus/evidence/task-20-worker-happy.txt`
    Expected Result: worker 任务注册、run/cancel、probe 基本行为全部通过
    Failure Indicators: 任务未退出；probe 不工作；返回码非 0
    Evidence: .sisyphus/evidence/task-20-worker-happy.txt

  Scenario: worker failure path - shutdown / probe 错误被聚合暴露
    Tool: Bash (go test)
    Preconditions: 已有 worker 失败场景测试
    Steps:
      1. 运行 `cd backend && go test ./internal/bootstrap -run 'TestWorkerPropagatesProbeFailure|TestAppShutdownAggregatesErrors' -count=1`
      2. 断言输出包含 `ok` 且返回码为 0
      3. 保存输出到 `.sisyphus/evidence/task-20-worker-failure.txt`
    Expected Result: worker 错误不会被吞掉，shutdown 聚合符合约定
    Failure Indicators: 错误被静默忽略；返回码非 0
    Evidence: .sisyphus/evidence/task-20-worker-failure.txt
  ```

  **Evidence to Capture**:
  - [ ] `.sisyphus/evidence/task-20-worker-happy.txt`
  - [ ] `.sisyphus/evidence/task-20-worker-failure.txt`

  **Commit**: YES
  - Message: `refactor(worker): 对齐占位任务的基础设施契约`
  - Files: `backend/cmd/worker/main.go`, `backend/internal/bootstrap/worker.go`, `backend/internal/bootstrap/*_test.go`
  - Pre-commit: `cd backend && go test ./internal/bootstrap -run TestWorker -count=1`

---

## Final Verification Wave (MANDATORY — after ALL implementation tasks)

> 4 个复核任务并行执行，全部通过后把结果展示给用户，等待用户明确 okay。
> **仍遵循 ZERO HUMAN INTERVENTION**：本波次所有检查必须由 agent 自动执行；`Real Manual QA` 在此处特指“agent 按步骤亲自操作与断言”，不是要求用户手动测试。

- [ ] F1. **Plan Compliance Audit** — `oracle`
  逐项核对本计划中的 Must Have / Must NOT Have / Deliverables / Evidence files，输出 `APPROVE/REJECT`。

  **QA Scenarios (MANDATORY)**:

  ```text
  Scenario: 合规 happy path - Must Have / Deliverables / Evidence 全部满足
    Tool: Bash + Read + Grep
    Preconditions: 所有任务已完成，`.sisyphus/evidence/` 中存在对应证据文件
    Steps:
      1. 读取 `.sisyphus/plans/p1-backend-foundation.md` 中的 Must Have、Must NOT Have、Deliverables
      2. 用 `grep` / `read` 核对 `backend/` 中对应实现是否存在，用 `read` 核对 `.sisyphus/evidence/` 证据清单是否齐全
      3. 输出汇总到 `.sisyphus/evidence/final-qa/f1-plan-compliance.txt`
    Expected Result: Must Have 全满足、Must NOT Have 未命中、Deliverables 与 evidence 均可对应
    Failure Indicators: 任一 Must Have 缺失；命中禁止项；证据缺失；结果无法归因到文件/命令
    Evidence: .sisyphus/evidence/final-qa/f1-plan-compliance.txt

  Scenario: 合规 failure path - 越界或缺项能被自动拒绝
    Tool: Bash + Grep
    Preconditions: 已定义禁止项关键词与路径边界检查规则
    Steps:
      1. 对 `backend/` 运行针对真实业务表、真实 CRUD、真实业务 endpoint、MQ/outbox 等越界项的搜索
      2. 记录命中结果或空结果到 `.sisyphus/evidence/final-qa/f1-plan-compliance-failure-check.txt`
    Expected Result: 无越界命中；若有命中可直接给出文件路径与拒绝理由
    Failure Indicators: 搜索规则模糊；越界项无法自动判定；结果不能落到具体文件
    Evidence: .sisyphus/evidence/final-qa/f1-plan-compliance-failure-check.txt
  ```

- [ ] F2. **Code Quality Review** — `unspecified-high`
  运行 `cd backend && go test ./...`、`cd backend && make lint`，并用自动化搜索检查 Go 世界里的坏味道：未处理错误、空错误分支、调试日志泄漏、注释掉代码、未使用导入、无意义抽象。

  **QA Scenarios (MANDATORY)**:

  ```text
  Scenario: 质量 happy path - 测试与 lint 全绿
    Tool: Bash
    Preconditions: 所有实现任务已完成
    Steps:
      1. 运行 `cd backend && go test ./...`
      2. 运行 `cd backend && make lint`
      3. 保存组合输出到 `.sisyphus/evidence/final-qa/f2-quality-happy.txt`
    Expected Result: 测试和 lint 返回码都为 0
    Failure Indicators: 任一命令失败；存在未修复静态检查错误；返回码非 0
    Evidence: .sisyphus/evidence/final-qa/f2-quality-happy.txt

  Scenario: 质量 failure path - 坏味道扫描可自动判定
    Tool: Bash + Grep + Read
    Preconditions: 已约定扫描模式，例如 `fmt.Println(`、`panic(`、`TODO`、注释掉代码片段、空错误处理模式
    Steps:
      1. 使用 `grep` 在 `backend/` 中对目标模式运行内容搜索
      2. 使用 `read` 逐项读取命中文件上下文，区分允许用法与违规用法
      3. 使用 `bash` 汇总审查结果到 `.sisyphus/evidence/final-qa/f2-quality-smells.txt`
    Expected Result: 无未解释的坏味道残留；若有命中，结果能归类为 APPROVE/REJECT 并附文件位置
    Failure Indicators: 仅凭主观判断不给规则；命中后无法解释；扫描遗漏显著模式
    Evidence: .sisyphus/evidence/final-qa/f2-quality-smells.txt
  ```

- [ ] F3. **Agent-Executed QA Replay** — `unspecified-high`
  按各任务 QA scenarios 逐项自动执行，保存证据到 `.sisyphus/evidence/final-qa/`；重点覆盖 health/readiness、顶层 trace_id、DB readiness、migration/sqlc/Testcontainers、worker shutdown。

  **QA Scenarios (MANDATORY)**:

  ```text
  Scenario: 回放 happy path - 关键公共链路逐项自动重放成功
    Tool: Bash
    Preconditions: 各任务定义的命令已可执行
    Steps:
      1. 运行 `cd backend && go test ./internal/platform/... -count=1`
      2. 运行 `cd backend && go test ./internal/bootstrap -count=1`
      3. 运行 `cd backend && go test ./test/integration -count=1`
      4. 运行 `cd backend && make sqlc-generate`
      5. 运行 `cd backend && make migrate-up`
      6. 运行 `cd backend && make test-integration`
      7. 记录每条命令的返回码与关键输出，并保存到 `.sisyphus/evidence/final-qa/f3-agent-replay-happy.txt`
    Expected Result: 关键公共链路全部按计划自动回放成功
    Failure Indicators: 任一关键命令失败；输出与计划期望不一致；证据缺失
    Evidence: .sisyphus/evidence/final-qa/f3-agent-replay-happy.txt

  Scenario: 回放 failure path - 关键失败场景自动重放成功
    Tool: Bash
    Preconditions: 已存在 readiness failure、migration/container failure、worker shutdown failure 等负向测试
    Steps:
      1. 运行 `cd backend && go test ./internal/bootstrap -run 'TestReadyzReturnsServiceUnavailableWhenDBFails|TestAppShutdownAggregatesErrors|TestWorkerShutdownHandlesFailures' -count=1`
      2. 运行 `cd backend && go test ./test/integration -run 'TestDatabaseGroundworkFailsOnBrokenMigration|TestDatabaseGroundworkHandlesContainerDependencyFailure|TestReadyzFailsWhenDBUnavailable' -count=1`
      3. 验证这些命令本身返回 0，证明“失败场景被正确断言”而不是被跳过
      4. 保存到 `.sisyphus/evidence/final-qa/f3-agent-replay-failure.txt`
    Expected Result: 失败路径均被自动化测试覆盖且断言成立
    Failure Indicators: 负向测试缺失；测试被 skip；失败原因不明确；返回码非 0
    Evidence: .sisyphus/evidence/final-qa/f3-agent-replay-failure.txt
  ```

- [ ] F4. **Scope Fidelity Check** — `deep`
  核对所有改动是否严格停留在 foundation + unified skeleton，不包含真实业务 schema、真实 CRUD、真实业务 endpoint、MQ/outbox、鉴权等越界项。

  **QA Scenarios (MANDATORY)**:

  ```text
  Scenario: 范围 happy path - 改动严格停留在 foundation + unified skeleton
    Tool: Bash + Read + Grep
    Preconditions: 所有实现任务已完成
    Steps:
      1. 使用 `read` 对照计划中的 IN / OUT / Guardrails 读取关键目录说明：`cmd/`、`internal/bootstrap/`、`internal/platform/`、`internal/modules/`、`migrations/`、`sql/queries/`
      2. 使用 `grep` 检查是否只出现允许的 platform/runtime probe 改动，而未出现业务越界模式
      3. 使用 `bash` 汇总结论到 `.sisyphus/evidence/final-qa/f4-scope-happy.txt`
    Expected Result: 所有改动均能映射到 foundation 或 unified skeleton 范围内
    Failure Indicators: 存在无法归属的改动；出现真实业务实现；范围说明与实际不符
    Evidence: .sisyphus/evidence/final-qa/f4-scope-happy.txt

  Scenario: 范围 failure path - 越界项会被自动命中并拒绝
    Tool: Bash + Grep
    Preconditions: 已定义越界关键模式：真实业务 endpoint、业务表、CRUD、MQ/outbox、auth 等
    Steps:
      1. 使用 `grep` 对 `backend/` 执行越界模式搜索
      2. 使用 `bash` 将命中结果或空结果保存到 `.sisyphus/evidence/final-qa/f4-scope-failure-check.txt`
    Expected Result: 无越界命中；若有命中则可直接定位并拒绝
    Failure Indicators: 越界搜索规则不明确；命中后无法给出拒绝依据
    Evidence: .sisyphus/evidence/final-qa/f4-scope-failure-check.txt
  ```

---

## Commit Strategy

- 每个任务点单独提交一次 commit，提交信息使用中文 Conventional Commits，例如：
  - `test(platform): 补充统一响应顶层 trace_id 失败用例`
  - `feat(database): 接入 pgx 连接池与事务管理器`
  - `chore(tooling): 补齐 sqlc 与 migration 命令`
- 每次 commit 前必须执行该任务的最小验证命令；不得跨任务打包提交。

---

## Success Criteria

### Verification Commands
```bash
cd backend && go test ./...
cd backend && make test
cd backend && make lint
cd backend && golangci-lint run
cd backend && make sqlc-generate
cd backend && make migrate-up
cd backend && make test-integration
```

### Final Checklist
- [ ] 顶层响应结构统一包含 `trace_id`
- [ ] `BindAndValidate` 统一覆盖 JSON/query/path/header
- [ ] `TxManager` 嵌套事务复用外层事务且有自动化测试
- [ ] `pgx + sqlc + goose + Testcontainers` 链路真实可跑
- [ ] `/readyz` 依赖 DB，`/healthz` 不依赖 DB
- [ ] OTEL traces + log correlation 主链路闭环完成
- [ ] 四模块骨架统一接入公共接缝，未新增真实业务 endpoint
- [ ] OpenAPI / Makefile / lint / tests 与代码实现同步
