# P1 Backend Foundation - Learnings

## 项目基线

- **工作目录**: `backend/` — 所有 Go 命令必须在此执行
- **Go module root**: `backend/go.mod`
- **分层**: `cmd -> bootstrap -> modules -> platform`

## 关键约定

- `APIResponse` 只有 `code/message/data/request_id`，无 `trace_id` — Task 2/7 修复点
- DB: `DummyDB` 占位，`RunMigrations` 空 — Task 6/10 修复点
- Observability: 全 noop — Task 5/11 修复点
- 四模块 handler 把 `trace_id` 手写到 payload — Task 14 清除此双轨

## 配置体系

- 强类型 Config struct + YAML + env override
- `Validate()` 只校验 `app.name` / `server.addr`
- `MaskedSummary()` 脱敏 db_dsn/redis

## 测试风格

- Unit tests 与 integration tests 分离
- Integration tests 在 `test/integration/` 使用 httptest + bootstrap engine
- Testcontainers 用于数据库集成测试

## Wave 1 执行说明

Wave 1 中 Task 1-6 全部无前置依赖，设计为并行执行，但为避免并发文件冲突，顺序委派：
1. T1 (config) — 先完成，为后续 DB/OTEL 字段打基础
2. T2 (response trace_id) — 独立修改 response 包
3. T3 (BindAndValidate) — 独立新建 platform/httpx 包
4. T4 (TxManager contract) — 独立扩展 platform/database 包
5. T5 (observability contract) — 独立扩展 observability 包
6. T6 (DB groundwork skeleton) — 独立新建集成测试骨架


## Task 1 配置基线补充（config）

- `Config` 新增 `Observability` 强类型段，字段包含 exporter type/endpoint 与 service 元信息。
- `DBConfig` 新增连接池参数：`max_open_conns`、`max_idle_conns`、`conn_max_lifetime_secs`，并支持 env override。
- `applyEnvOverrides` 新增覆盖键：`DB_MAX_OPEN_CONNS`、`DB_MAX_IDLE_CONNS`、`DB_CONN_MAX_LIFETIME_SECS`、`OTEL_TRACE_EXPORTER_TYPE`、`OTEL_TRACE_EXPORTER_ENDPOINT`、`OTEL_SERVICE_NAME`、`OTEL_SERVICE_VERSION`、`OTEL_ENVIRONMENT`。
- `Validate()` 扩展最小契约：连接池参数非负、idle 不大于 open、`trace_exporter_type` 仅允许 `none|otlp`、`otlp` 场景必须有 endpoint、`observability.service_name` 必填。
- `MaskedSummary()` 继续脱敏：`db_dsn` 与 `otel_exporter_endpoint` 均通过 `maskSecret` 输出，避免明文泄露。
- 四套配置文件已补齐 DB 连接池 + observability 段，`config.test.yaml` 默认 `otlp` 用于测试加载契约。
- 按 TDD 执行：先新增 `TestLoadRejectsInvalidDBOrObservabilityConfig` 并验证 RED，再实现到 GREEN。


## Task 3 统一 BindAndValidate 契约（httpx）

- 新增 `internal/platform/httpx.BindAndValidate[T any](c *gin.Context) (*T, error)`，固定顺序为 `uri -> query -> header -> json`。
- 为避免 Gin v1.12 `ShouldBind*` 每步都触发 validator 的重复校验问题，helper 采用“先纯绑定、后统一 `binding.Validator.ValidateStruct` 校验”的方式。
- URI / query 复用 `binding.MapFormWithTag`；header 通过反射收集 DTO 上声明的 `header` tag 后再映射，兼容 Gin header binder 的 canonical key 行为。
- JSON body 仅支持 JSON；缺省空 body 时跳过解析，解析错误与统一校验错误都映射为 `AppError{Code: INVALID_ARGUMENT, HTTPStatus: 400}` 并保留 cause。
- 契约测试覆盖 happy path（path/query/header/json 四源聚合）以及 failure path（缺失 header 的 validation error、坏 JSON 的 bind error）。

## Task 6 DB groundwork skeleton

- `test/integration/db_groundwork_test.go` 采用与 `ping_test.go` 一致的集成测试风格：直接在 Go 测试中启动依赖、执行业务外 probe 验证、并通过 `bootstrap.NewAPIEngine` 回归 `/readyz`。
- DB 载体严格限制为 `platform_runtime_probes`，配套 fixture 只放在 `test/integration/testdata/db_groundwork/`，未引入任何业务表或业务 endpoint。
- 迁移 fixture 用 goose SQL 文件表达 happy/broken 两条路径；probe query fixture 保持 sqlc 风格标记，但在 Task 6 内仅作为“待接入 sqlc”的最小解析输入，由测试内轻量解析执行，避免提前跨入 Task 10 的真实 sqlc 生成实现。
- Testcontainers 在当前环境需要显式兼容 Colima：测试内若发现 `~/.colima/default/docker.sock`，则设置 `DOCKER_HOST=unix://...` 与 `TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE=/var/run/docker.sock`，否则容易报 `rootless Docker not found`。
- failure evidence 通过 `TestDatabaseGroundworkFailsOnBrokenMigration` 固化“broken migration 会被准确暴露但测试本身仍转绿”的负向断言，符合计划要求的 failure-path QA 语义。

## Task 7 追踪响应回归

- `response.APIResponse` 的顶层 `trace_id` 已生效后，bootstrap 侧测试应直接断言 `APIResponse.TraceID`，避免继续依赖 `data.trace_id` 的历史双轨字段。
- `ping_test.go` 的回归重点是顶层响应封装，不是业务 payload；业务 `data` 仅保留 `module` 断言，`trace_id` 统一从 `APIResponse.TraceID` 读取。

## 2026-05-14 Task 9
- SQLTxManager 使用 Beginner + sqlTx 接口解耦 *sql.Tx concrete type，便于 pure Go fake 单测。
- 嵌套事务判定必须先调用 ExecutorFromContext(ctx, nil)，已有 executor 时直接复用 ctx，不再 BeginTx。

- Task 10: sqlc v2 配置使用 ，生成的  依赖  风格 （Exec/Query/QueryRow 带 context），因此 integration test 需直接基于  使用生成代码，而 goose 迁移仍可通过独立  + pgx stdlib 执行。
- Task 10: 为兼容 macOS Colima，保留 ；本地 Docker 通过  成功跑通 Postgres + Ryuk。

- Task 10（修正记录）: 上一条 Task 10 学习记录因 shell 反引号插值被污染；以下为准。
- Task 10: `sqlc` v2 配置应使用 `sql_package: pgx/v5`，生成的 `dbgen.Queries` 依赖 `pgx` 风格 `DBTX`（`Exec/Query/QueryRow` 均携带 `context.Context`），因此 integration test 需直接基于 `*pgxpool.Pool` 调用生成查询，而 goose 迁移仍通过独立 `database/sql` + pgx stdlib 跑通即可。
- Task 10: 在 macOS + Colima 环境下，保留 `configureTestcontainersDockerEnvironment` 很关键；本次验证通过 `unix:///Users/chening/.colima/default/docker.sock` 成功跑通 Postgres 与 Ryuk。

## Task 11 - OTEL provider
- observability provider now initializes from config, keeps noop fallback for trace_exporter_type=none, and returns a usable noop provider plus observable error when OTLP endpoint is missing.
- Propagator preserves X-Trace-ID compatibility while preferring W3C traceparent on extract and injecting traceparent for valid W3C trace IDs.
- go mod tidy is required when importing OTLP trace gRPC exporter packages so go.mod/go.sum stay consistent.
