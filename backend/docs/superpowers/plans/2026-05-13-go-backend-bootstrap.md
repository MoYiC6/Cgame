# Go Backend Bootstrap Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 初始化一个可启动、可测试的 Go 模块化单体后端骨架，包含 API/Worker 入口、平台基础组件、四个业务模块 ping 路由、最小集成测试和工程命令。

**Architecture:** 先搭建工程骨架与配置样例，再按“platform 基础组件 → middleware/bootstrap → 业务模块 → worker 生命周期 → cmd 入口/工程命令”的顺序推进。所有真正的生产代码都遵循 TDD：先写失败测试，再实现刚好通过测试的最小代码；纯目录、配置样例、OpenAPI 占位和 lint 配置属于脚手架输入，不纳入 TDD 红灯要求。

**Tech Stack:** Go 1.26, Gin, log/slog, gopkg.in/yaml.v3, net/http/httptest, golangci-lint

---

## Preconditions

- 当前 `backend/` 目录不是 Git 仓库，因此本计划**不包含强制 commit 步骤**；如果后续初始化 Git，可按每个 Task 的里程碑单独提交。
- 严格限制范围：**不接真实 DB / Redis / MQ，不加 sqlc，不做真实 migration，不接真实 OpenTelemetry exporter，不写真实支付/通知适配器，不实现真实业务流程。**
- 允许为后续能力预留接口，但不提前做整套基础设施实现。
- 本计划默认在 `/Users/chening/Desktop/zuhao/backend` 执行。

## File Structure Lock-in

### Create: startup and docs
- `cmd/api/main.go`
- `cmd/worker/main.go`
- `api/openapi.yaml`

### Create: configs and engineering files
- `configs/config.local.yaml`
- `configs/config.dev.yaml`
- `configs/config.test.yaml`
- `configs/config.prod.yaml`
- `Makefile`
- `.golangci.yml`
- `.env.example`

### Create: platform
- `internal/platform/config/config.go`
- `internal/platform/config/config_test.go`
- `internal/platform/errors/errors.go`
- `internal/platform/errors/errors_test.go`
- `internal/platform/response/response.go`
- `internal/platform/response/response_test.go`
- `internal/platform/observability/context.go`
- `internal/platform/observability/tracer.go`
- `internal/platform/observability/propagator.go`
- `internal/platform/observability/observability_test.go`
- `internal/platform/logger/logger.go`
- `internal/platform/logger/logger_test.go`

### Create: bootstrap
- `internal/bootstrap/app.go`
- `internal/bootstrap/server.go`
- `internal/bootstrap/middleware.go`
- `internal/bootstrap/middleware_test.go`
- `internal/bootstrap/server_test.go`
- `internal/bootstrap/worker.go`
- `internal/bootstrap/worker_test.go`

### Create: modules
- `internal/modules/order/handler.go`
- `internal/modules/order/service.go`
- `internal/modules/order/repository.go`
- `internal/modules/order/dto.go`
- `internal/modules/order/model.go`
- `internal/modules/order/status.go`
- `internal/modules/order/events.go`
- `internal/modules/payment/handler.go`
- `internal/modules/payment/service.go`
- `internal/modules/payment/repository.go`
- `internal/modules/payment/dto.go`
- `internal/modules/payment/model.go`
- `internal/modules/payment/status.go`
- `internal/modules/payment/events.go`
- `internal/modules/inventory/handler.go`
- `internal/modules/inventory/service.go`
- `internal/modules/inventory/repository.go`
- `internal/modules/inventory/dto.go`
- `internal/modules/inventory/model.go`
- `internal/modules/inventory/status.go`
- `internal/modules/inventory/events.go`
- `internal/modules/notification/handler.go`
- `internal/modules/notification/service.go`
- `internal/modules/notification/repository.go`
- `internal/modules/notification/dto.go`
- `internal/modules/notification/model.go`
- `internal/modules/notification/status.go`
- `internal/modules/notification/events.go`

### Create: clients
- `internal/clients/paymentgateway/client.go`
- `internal/clients/sms/client.go`
- `internal/clients/email/client.go`

### Create: test and placeholders
- `test/integration/ping_test.go`
- `test/fixtures/.gitkeep`
- `migrations/.gitkeep`
- `sql/queries/.gitkeep`

---

## Task 1: Create the repository skeleton and placeholder files

**Files:**
- Create: `api/openapi.yaml`
- Create: `configs/config.local.yaml`
- Create: `configs/config.dev.yaml`
- Create: `configs/config.test.yaml`
- Create: `configs/config.prod.yaml`
- Create: `.env.example`
- Create: `.golangci.yml`
- Create: `migrations/.gitkeep`
- Create: `sql/queries/.gitkeep`
- Create: `test/fixtures/.gitkeep`

- [ ] **Step 1: Create the directory structure**

Run:
```bash
mkdir -p cmd/api cmd/worker api configs internal/bootstrap internal/platform/config internal/platform/errors internal/platform/response internal/platform/observability internal/platform/logger internal/modules/order internal/modules/payment internal/modules/inventory internal/modules/notification internal/clients/paymentgateway internal/clients/sms internal/clients/email migrations sql/queries test/integration test/fixtures
```

Expected: all directories exist under `backend/`.

- [ ] **Step 2: Add the baseline dependencies**

Run:
```bash
go get github.com/gin-gonic/gin gopkg.in/yaml.v3 && go mod tidy
```

Expected: `go.mod` and `go.sum` include Gin and YAML support.

- [ ] **Step 3: Create the placeholder engineering and config files**

Write these files exactly:

```yaml
# configs/config.local.yaml
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

```yaml
# configs/config.dev.yaml
app:
  name: backend
  env: dev

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
  topic_prefix: backend.dev
```

```yaml
# configs/config.test.yaml
app:
  name: backend-test
  env: test

server:
  addr: ":18080"

log:
  level: debug

db:
  # 预留，初始化阶段不使用
  driver: postgres
  dsn: "postgres://user:secret@localhost:5432/backend_test"

redis:
  # 预留，初始化阶段不使用
  addr: "127.0.0.1:6379"

mq:
  # 预留，初始化阶段不使用
  driver: in-memory
  topic_prefix: backend.test
```

```yaml
# configs/config.prod.yaml
app:
  name: backend
  env: prod

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
  topic_prefix: backend.prod
```

```yaml
# .golangci.yml
run:
  timeout: 5m

linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - staticcheck
    - unused
```

```env
# .env.example
APP_CONFIG_PATH=configs/config.local.yaml
```

```yaml
# api/openapi.yaml
openapi: 3.0.3
info:
  title: Backend Bootstrap API
  version: 0.1.0
paths:
  /healthz:
    get:
      summary: Process health check
      responses:
        '200':
          description: OK
  /readyz:
    get:
      summary: Dependency readiness check
      responses:
        '200':
          description: OK
```

Also create empty placeholder files:
```text
migrations/.gitkeep
sql/queries/.gitkeep
test/fixtures/.gitkeep
```

- [ ] **Step 4: Verify the scaffold files exist**

Run:
```bash
test -f configs/config.test.yaml && test -f api/openapi.yaml && test -f .golangci.yml && test -f .env.example && test -f migrations/.gitkeep && test -f sql/queries/.gitkeep && test -f test/fixtures/.gitkeep
```

Expected: command exits with code 0 and no output.

---

## Task 2: Build the typed config component with tests first

**Files:**
- Create: `internal/platform/config/config_test.go`
- Create: `internal/platform/config/config.go`

- [ ] **Step 1: Write the failing config tests**

Create `internal/platform/config/config_test.go`:

```go
package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadFromFile(t *testing.T) {
	cfg, err := Load(filepath.Join("..", "..", "..", "configs", "config.test.yaml"))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.App.Name != "backend-test" {
		t.Fatalf("expected app name backend-test, got %q", cfg.App.Name)
	}
	if cfg.Server.Addr != ":18080" {
		t.Fatalf("expected server addr :18080, got %q", cfg.Server.Addr)
	}
	if cfg.Log.Level != "debug" {
		t.Fatalf("expected log level debug, got %q", cfg.Log.Level)
	}
	if cfg.DB.Driver != "postgres" {
		t.Fatalf("expected DB driver postgres, got %q", cfg.DB.Driver)
	}
	if cfg.MQ.TopicPrefix != "backend.test" {
		t.Fatalf("expected topic prefix backend.test, got %q", cfg.MQ.TopicPrefix)
	}
}

func TestLoadRejectsMissingAppName(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	content := []byte("app:\n  name: \"\"\nserver:\n  addr: \":8080\"\nlog:\n  level: info\n")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected Load to fail when app.name is empty")
	}
	if !strings.Contains(err.Error(), "app.name") {
		t.Fatalf("expected error to mention app.name, got %v", err)
	}
}

func TestMaskedSummaryHidesSecrets(t *testing.T) {
	cfg := Config{
		App: AppConfig{Name: "backend", Env: "test"},
		Server: ServerConfig{Addr: ":18080"},
		Log: LogConfig{Level: "debug"},
		DB: DBConfig{Driver: "postgres", DSN: "postgres://user:secret@localhost:5432/backend_test"},
		Redis: RedisConfig{Addr: "127.0.0.1:6379"},
		MQ: MQConfig{Driver: "in-memory", TopicPrefix: "backend.test"},
	}

	summary := cfg.MaskedSummary()
	if strings.Contains(summary["db_dsn"], "secret") {
		t.Fatalf("expected masked db_dsn, got %q", summary["db_dsn"])
	}
	if summary["app_name"] != "backend" {
		t.Fatalf("expected app_name backend, got %q", summary["app_name"])
	}
}
```

- [ ] **Step 2: Run the config tests to verify they fail**

Run:
```bash
go test ./internal/platform/config -run 'TestLoadFromFile|TestLoadRejectsMissingAppName|TestMaskedSummaryHidesSecrets' -v
```

Expected: FAIL because `Load`, `Config`, or `MaskedSummary` do not exist yet.

- [ ] **Step 3: Write the minimal config implementation**

Create `internal/platform/config/config.go`:

```go
package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App    AppConfig    `yaml:"app"`
	Server ServerConfig `yaml:"server"`
	Log    LogConfig    `yaml:"log"`
	DB     DBConfig     `yaml:"db"`
	Redis  RedisConfig  `yaml:"redis"`
	MQ     MQConfig     `yaml:"mq"`
}

type AppConfig struct {
	Name string `yaml:"name"`
	Env  string `yaml:"env"`
}

type ServerConfig struct {
	Addr string `yaml:"addr"`
}

type LogConfig struct {
	Level string `yaml:"level"`
}

type DBConfig struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

type RedisConfig struct {
	Addr string `yaml:"addr"`
}

type MQConfig struct {
	Driver      string `yaml:"driver"`
	TopicPrefix string `yaml:"topic_prefix"`
}

func Load(path string) (Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	if cfg.App.Env == "" {
		cfg.App.Env = "local"
	}
	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.App.Name) == "" {
		return fmt.Errorf("app.name is required")
	}
	if strings.TrimSpace(c.Server.Addr) == "" {
		return fmt.Errorf("server.addr is required")
	}
	return nil
}

func (c Config) MaskedSummary() map[string]string {
	return map[string]string{
		"app_name": c.App.Name,
		"app_env":  c.App.Env,
		"server":   c.Server.Addr,
		"log_level": c.Log.Level,
		"db_driver": c.DB.Driver,
		"db_dsn":    maskSecret(c.DB.DSN),
		"redis":     maskSecret(c.Redis.Addr),
		"mq_driver": c.MQ.Driver,
		"mq_topic":  c.MQ.TopicPrefix,
	}
}

func maskSecret(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 4 {
		return "****"
	}
	return value[:2] + strings.Repeat("*", len(value)-4) + value[len(value)-2:]
}
```

- [ ] **Step 4: Run the config tests again**

Run:
```bash
go test ./internal/platform/config -run 'TestLoadFromFile|TestLoadRejectsMissingAppName|TestMaskedSummaryHidesSecrets' -v
```

Expected:
```text
=== RUN   TestLoadFromFile
--- PASS: TestLoadFromFile
=== RUN   TestLoadRejectsMissingAppName
--- PASS: TestLoadRejectsMissingAppName
=== RUN   TestMaskedSummaryHidesSecrets
--- PASS: TestMaskedSummaryHidesSecrets
PASS
```

---

## Task 3: Add AppError and unified response helpers

**Files:**
- Create: `internal/platform/errors/errors_test.go`
- Create: `internal/platform/errors/errors.go`
- Create: `internal/platform/response/response_test.go`
- Create: `internal/platform/response/response.go`

- [ ] **Step 1: Write the failing AppError tests**

Create `internal/platform/errors/errors_test.go`:

```go
package apperrors

import (
	stderrors "errors"
	"net/http"
	"testing"
)

func TestWrapPreservesMetadata(t *testing.T) {
	base := NewAppError("INTERNAL_ERROR", "internal error", http.StatusInternalServerError)
	wrapped := Wrap(base, stderrors.New("database down"))

	if wrapped.Code != "INTERNAL_ERROR" {
		t.Fatalf("expected code INTERNAL_ERROR, got %q", wrapped.Code)
	}
	if wrapped.HTTPStatus != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", wrapped.HTTPStatus)
	}
	if wrapped.Cause == nil {
		t.Fatal("expected wrapped cause to be present")
	}
}

func TestMetadataHelpersHandleGenericErrors(t *testing.T) {
	err := stderrors.New("boom")
	if Code(err) != "INTERNAL_ERROR" {
		t.Fatalf("expected INTERNAL_ERROR fallback, got %q", Code(err))
	}
	if Status(err) != http.StatusInternalServerError {
		t.Fatalf("expected 500 fallback, got %d", Status(err))
	}
	if SafeMessage(err) != "internal error" {
		t.Fatalf("expected generic safe message, got %q", SafeMessage(err))
	}
}
```

- [ ] **Step 2: Run the AppError tests to verify they fail**

Run:
```bash
go test ./internal/platform/errors -run 'TestWrapPreservesMetadata|TestMetadataHelpersHandleGenericErrors' -v
```

Expected: FAIL because `NewAppError`, `Wrap`, `Code`, `Status`, and `SafeMessage` do not exist yet.

- [ ] **Step 3: Implement the AppError type**

Create `internal/platform/errors/errors.go`:

```go
package apperrors

import (
	stderrors "errors"
	"net/http"
)

type AppError struct {
	Code       string
	Message    string
	HTTPStatus int
	Cause      error
}

func NewAppError(code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

func Wrap(base *AppError, cause error) *AppError {
	if base == nil {
		return &AppError{Code: "INTERNAL_ERROR", Message: "internal error", HTTPStatus: http.StatusInternalServerError, Cause: cause}
	}
	clone := *base
	clone.Cause = cause
	return &clone
}

func (e *AppError) Error() string {
	if e == nil {
		return "internal error"
	}
	return e.Message
}

func Status(err error) int {
	var appErr *AppError
	if stderrors.As(err, &appErr) {
		return appErr.HTTPStatus
	}
	return http.StatusInternalServerError
}

func Code(err error) string {
	var appErr *AppError
	if stderrors.As(err, &appErr) {
		return appErr.Code
	}
	return "INTERNAL_ERROR"
}

func SafeMessage(err error) string {
	var appErr *AppError
	if stderrors.As(err, &appErr) {
		return appErr.Message
	}
	return "internal error"
}
```

- [ ] **Step 4: Write the failing response tests**

Create `internal/platform/response/response_test.go`:

```go
package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	apperrors "backend/internal/platform/errors"
	"backend/internal/platform/observability"
	"github.com/gin-gonic/gin"
)

func TestSuccessWritesRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	request = request.WithContext(observability.WithRequestID(request.Context(), "req-123"))
	ctx.Request = request

	Success(ctx, gin.H{"module": "health"})

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var body APIResponse[map[string]any]
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	if body.RequestID != "req-123" {
		t.Fatalf("expected request_id req-123, got %q", body.RequestID)
	}
	if body.Code != "OK" {
		t.Fatalf("expected code OK, got %q", body.Code)
	}
}

func TestFailUsesAppErrorMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	request := httptest.NewRequest(http.MethodGet, "/panic", nil)
	request = request.WithContext(observability.WithRequestID(request.Context(), "req-500"))
	ctx.Request = request

	Fail(ctx, apperrors.NewAppError("INVALID_ARGUMENT", "invalid input", http.StatusBadRequest))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}

	var body APIResponse[map[string]any]
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	if body.Code != "INVALID_ARGUMENT" {
		t.Fatalf("expected INVALID_ARGUMENT, got %q", body.Code)
	}
	if body.Message != "invalid input" {
		t.Fatalf("expected invalid input, got %q", body.Message)
	}
	if body.RequestID != "req-500" {
		t.Fatalf("expected request_id req-500, got %q", body.RequestID)
	}
}
```

- [ ] **Step 5: Run the response tests to verify they fail**

Run:
```bash
go test ./internal/platform/response -run 'TestSuccessWritesRequestID|TestFailUsesAppErrorMetadata' -v
```

Expected: FAIL because `Success`, `Fail`, and `APIResponse` do not exist yet.

- [ ] **Step 6: Implement the response helpers**

Create `internal/platform/response/response.go`:

```go
package response

import (
	"net/http"

	apperrors "backend/internal/platform/errors"
	"backend/internal/platform/observability"
	"github.com/gin-gonic/gin"
)

type APIResponse[T any] struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Data      T      `json:"data,omitempty"`
	RequestID string `json:"request_id"`
}

func Success(c *gin.Context, data any) {
	requestID, _ := observability.RequestIDFromContext(c.Request.Context())
	c.JSON(http.StatusOK, APIResponse[any]{
		Code:      "OK",
		Message:   "success",
		Data:      data,
		RequestID: requestID,
	})
}

func Fail(c *gin.Context, err error) {
	requestID, _ := observability.RequestIDFromContext(c.Request.Context())
	c.JSON(apperrors.Status(err), APIResponse[any]{
		Code:      apperrors.Code(err),
		Message:   apperrors.SafeMessage(err),
		RequestID: requestID,
	})
}
```

- [ ] **Step 7: Run the error and response tests again**

Run:
```bash
go test ./internal/platform/errors ./internal/platform/response -v
```

Expected:
```text
ok  	backend/internal/platform/errors
ok  	backend/internal/platform/response
```

---

## Task 4: Build observability helpers and the slog seam

**Files:**
- Create: `internal/platform/observability/observability_test.go`
- Create: `internal/platform/observability/context.go`
- Create: `internal/platform/observability/tracer.go`
- Create: `internal/platform/observability/propagator.go`
- Create: `internal/platform/logger/logger_test.go`
- Create: `internal/platform/logger/logger.go`

- [ ] **Step 1: Write the failing observability tests**

Create `internal/platform/observability/observability_test.go`:

```go
package observability

import (
	"context"
	"testing"
)

func TestRequestAndTraceContextRoundTrip(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-42")
	ctx = WithTraceID(ctx, "trace-42")

	requestID, ok := RequestIDFromContext(ctx)
	if !ok || requestID != "req-42" {
		t.Fatalf("expected request id req-42, got %q", requestID)
	}
	traceID, ok := TraceIDFromContext(ctx)
	if !ok || traceID != "trace-42" {
		t.Fatalf("expected trace id trace-42, got %q", traceID)
	}
}

func TestNoopPropagatorInjectExtract(t *testing.T) {
	propagator := NewNoopPropagator()
	carrier := MapCarrier{}
	ctx := WithTraceID(context.Background(), "trace-prop")

	propagator.Inject(ctx, carrier)
	newCtx := propagator.Extract(context.Background(), carrier)

	traceID, ok := TraceIDFromContext(newCtx)
	if !ok || traceID != "trace-prop" {
		t.Fatalf("expected extracted trace id trace-prop, got %q", traceID)
	}
}
```

- [ ] **Step 2: Run the observability tests to verify they fail**

Run:
```bash
go test ./internal/platform/observability -run 'TestRequestAndTraceContextRoundTrip|TestNoopPropagatorInjectExtract' -v
```

Expected: FAIL because the context helpers and no-op tracer/propagator do not exist yet.

- [ ] **Step 3: Implement the observability package**

Create `internal/platform/observability/context.go`:

```go
package observability

import "context"

type contextKey string

const (
	requestIDKey contextKey = "request_id"
	traceIDKey   contextKey = "trace_id"
)

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

func RequestIDFromContext(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(requestIDKey).(string)
	return value, ok && value != ""
}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

func TraceIDFromContext(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(traceIDKey).(string)
	return value, ok && value != ""
}
```

Create `internal/platform/observability/tracer.go`:

```go
package observability

import "context"

type Tracer interface {
	StartSpan(ctx context.Context, name string) (context.Context, Span)
}

type Span interface {
	End(err error)
}

type noopTracer struct{}

type noopSpan struct{}

func NewNoopTracer() Tracer {
	return noopTracer{}
}

func (noopTracer) StartSpan(ctx context.Context, name string) (context.Context, Span) {
	return ctx, noopSpan{}
}

func (noopSpan) End(err error) {}
```

Create `internal/platform/observability/propagator.go`:

```go
package observability

import "context"

type Propagator interface {
	Inject(ctx context.Context, carrier Carrier)
	Extract(ctx context.Context, carrier Carrier) context.Context
}

type Carrier interface {
	Get(key string) string
	Set(key string, value string)
	Keys() []string
}

type MapCarrier map[string]string

type noopPropagator struct{}

func NewNoopPropagator() Propagator {
	return noopPropagator{}
}

func (m MapCarrier) Get(key string) string {
	return m[key]
}

func (m MapCarrier) Set(key string, value string) {
	m[key] = value
}

func (m MapCarrier) Keys() []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

func (noopPropagator) Inject(ctx context.Context, carrier Carrier) {
	if traceID, ok := TraceIDFromContext(ctx); ok {
		carrier.Set("X-Trace-ID", traceID)
	}
}

func (noopPropagator) Extract(ctx context.Context, carrier Carrier) context.Context {
	if traceID := carrier.Get("X-Trace-ID"); traceID != "" {
		return WithTraceID(ctx, traceID)
	}
	return ctx
}
```

- [ ] **Step 4: Write the failing logger test**

Create `internal/platform/logger/logger_test.go`:

```go
package logger

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"backend/internal/platform/observability"
)

func TestWithContextAddsRequestAndTraceIDs(t *testing.T) {
	var buffer bytes.Buffer
	base := New("debug", &buffer)
	ctx := context.Background()
	ctx = observability.WithRequestID(ctx, "req-log")
	ctx = observability.WithTraceID(ctx, "trace-log")

	WithContext(ctx, base).Info("boot ok", "component", "api")

	output := buffer.String()
	if !strings.Contains(output, "req-log") {
		t.Fatalf("expected output to contain request id, got %s", output)
	}
	if !strings.Contains(output, "trace-log") {
		t.Fatalf("expected output to contain trace id, got %s", output)
	}
}
```

- [ ] **Step 5: Run the logger test to verify it fails**

Run:
```bash
go test ./internal/platform/logger -run TestWithContextAddsRequestAndTraceIDs -v
```

Expected: FAIL because `New` and `WithContext` do not exist yet.

- [ ] **Step 6: Implement the logger package**

Create `internal/platform/logger/logger.go`:

```go
package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"

	"backend/internal/platform/observability"
)

func New(level string, writer io.Writer) *slog.Logger {
	if writer == nil {
		writer = os.Stdout
	}

	var slogLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		slogLevel = slog.LevelDebug
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: slogLevel})
	return slog.New(handler)
}

func WithContext(ctx context.Context, base *slog.Logger) *slog.Logger {
	if base == nil {
		base = New("info", nil)
	}

	fields := make([]any, 0, 4)
	if requestID, ok := observability.RequestIDFromContext(ctx); ok {
		fields = append(fields, "request_id", requestID)
	}
	if traceID, ok := observability.TraceIDFromContext(ctx); ok {
		fields = append(fields, "trace_id", traceID)
	}
	if len(fields) == 0 {
		return base
	}
	return base.With(fields...)
}
```

- [ ] **Step 7: Run the observability and logger tests again**

Run:
```bash
go test ./internal/platform/observability ./internal/platform/logger -v
```

Expected:
```text
ok  	backend/internal/platform/observability
ok  	backend/internal/platform/logger
```

---

## Task 5: Add middleware and HTTP bootstrap around health routes

**Files:**
- Create: `internal/bootstrap/middleware_test.go`
- Create: `internal/bootstrap/middleware.go`
- Create: `internal/bootstrap/app.go`
- Create: `internal/bootstrap/server.go`
- Create: `internal/bootstrap/server_test.go`

- [ ] **Step 1: Write the failing middleware tests**

Create `internal/bootstrap/middleware_test.go`:

```go
package bootstrap

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/platform/logger"
	"backend/internal/platform/observability"
	"backend/internal/platform/response"
	"github.com/gin-gonic/gin"
)

func TestRequestAndTraceMiddlewarePopulateHeadersAndContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(RequestIDMiddleware(), TraceContextMiddleware(observability.NewNoopPropagator()))
	engine.GET("/ping", func(c *gin.Context) {
		traceID, _ := observability.TraceIDFromContext(c.Request.Context())
		response.Success(c, gin.H{"trace_id": traceID})
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/ping", nil)
	request.Header.Set("X-Request-ID", "req-mw")
	request.Header.Set("X-Trace-ID", "trace-mw")

	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if recorder.Header().Get("X-Request-ID") != "req-mw" {
		t.Fatalf("expected response request id req-mw, got %q", recorder.Header().Get("X-Request-ID"))
	}
	if recorder.Header().Get("X-Trace-ID") != "trace-mw" {
		t.Fatalf("expected response trace id trace-mw, got %q", recorder.Header().Get("X-Trace-ID"))
	}

	var body response.APIResponse[map[string]any]
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	if body.RequestID != "req-mw" {
		t.Fatalf("expected request_id req-mw, got %q", body.RequestID)
	}
	if body.Data["trace_id"] != "trace-mw" {
		t.Fatalf("expected trace_id trace-mw, got %#v", body.Data["trace_id"])
	}
}

func TestRecoveryMiddlewareConvertsPanicToJSONError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(RequestIDMiddleware(), RecoveryMiddleware(logger.New("debug", io.Discard)))
	engine.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/panic", nil)
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
}
```

- [ ] **Step 2: Run the middleware tests to verify they fail**

Run:
```bash
go test ./internal/bootstrap -run 'TestRequestAndTraceMiddlewarePopulateHeadersAndContext|TestRecoveryMiddlewareConvertsPanicToJSONError' -v
```

Expected: FAIL because the middleware functions do not exist yet.

- [ ] **Step 3: Implement the bootstrap dependency types and middleware**

Create `internal/bootstrap/app.go`:

```go
package bootstrap

import (
	"log/slog"

	"backend/internal/platform/config"
	"backend/internal/platform/observability"
)

type Dependencies struct {
	Config     config.Config
	Logger     *slog.Logger
	Tracer     observability.Tracer
	Propagator observability.Propagator
}

type RouteRegistrar interface {
	RegisterRoutes(group any)
}
```

Create `internal/bootstrap/middleware.go`:

```go
package bootstrap

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	apperrors "backend/internal/platform/errors"
	"backend/internal/platform/observability"
	"backend/internal/platform/response"
	"github.com/gin-gonic/gin"
)

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("req-%d", time.Now().UnixNano())
		}

		ctx := observability.WithRequestID(c.Request.Context(), requestID)
		c.Request = c.Request.WithContext(ctx)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func TraceContextMiddleware(propagator observability.Propagator) gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = fmt.Sprintf("trace-%d", time.Now().UnixNano())
		}

		carrier := observability.MapCarrier{"X-Trace-ID": traceID}
		ctx := propagator.Extract(c.Request.Context(), carrier)
		ctx = observability.WithTraceID(ctx, traceID)
		c.Request = c.Request.WithContext(ctx)
		c.Header("X-Trace-ID", traceID)
		c.Next()
	}
}

func RecoveryMiddleware(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				if log != nil {
					log.Error("panic recovered", "panic", recovered, "path", c.Request.URL.Path)
				}
				response.Fail(c, apperrors.Wrap(
					apperrors.NewAppError("INTERNAL_ERROR", "internal error", http.StatusInternalServerError),
					fmt.Errorf("panic: %v", recovered),
				))
				c.Abort()
			}
		}()

		c.Next()
	}
}
```

- [ ] **Step 4: Write the failing server bootstrap test**

Create `internal/bootstrap/server_test.go`:

```go
package bootstrap

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/platform/config"
	"backend/internal/platform/logger"
	"backend/internal/platform/observability"
	"backend/internal/platform/response"
	"github.com/gin-gonic/gin"
)

func TestNewAPIEngineRegistersHealthRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	deps := Dependencies{
		Config: config.Config{
			App:    config.AppConfig{Name: "backend-test", Env: "test"},
			Server: config.ServerConfig{Addr: ":18080"},
			Log:    config.LogConfig{Level: "debug"},
		},
		Logger:     logger.New("debug", io.Discard),
		Tracer:     observability.NewNoopTracer(),
		Propagator: observability.NewNoopPropagator(),
	}

	engine := NewAPIEngine(deps)

	for _, path := range []string{"/healthz", "/readyz"} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, path, nil)
		engine.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200 for %s, got %d", path, recorder.Code)
		}

		var body response.APIResponse[map[string]any]
		if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
			t.Fatalf("json.Unmarshal returned error: %v", err)
		}
		if body.Code != "OK" {
			t.Fatalf("expected code OK for %s, got %q", path, body.Code)
		}
		if body.RequestID == "" {
			t.Fatalf("expected request_id for %s", path)
		}
	}
}
```

- [ ] **Step 5: Run the server test to verify it fails**

Run:
```bash
go test ./internal/bootstrap -run TestNewAPIEngineRegistersHealthRoutes -v
```

Expected: FAIL because `NewAPIEngine` does not exist yet.

- [ ] **Step 6: Implement the API engine**

Create `internal/bootstrap/server.go`:

```go
package bootstrap

import (
	"backend/internal/platform/response"
	"github.com/gin-gonic/gin"
)

type HTTPRouteRegistrar interface {
	RegisterRoutes(group *gin.RouterGroup)
}

func NewAPIEngine(deps Dependencies, registrars ...HTTPRouteRegistrar) *gin.Engine {
	engine := gin.New()
	engine.Use(RequestIDMiddleware(), TraceContextMiddleware(deps.Propagator), RecoveryMiddleware(deps.Logger))

	engine.GET("/healthz", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "ok"})
	})
	engine.GET("/readyz", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "ok", "dependencies": "skipped"})
	})

	api := engine.Group("/api/v1")
	for _, registrar := range registrars {
		registrar.RegisterRoutes(api)
	}

	return engine
}
```

- [ ] **Step 7: Run the bootstrap tests again**

Run:
```bash
go test ./internal/bootstrap -run 'TestRequestAndTraceMiddlewarePopulateHeadersAndContext|TestRecoveryMiddlewareConvertsPanicToJSONError|TestNewAPIEngineRegistersHealthRoutes' -v
```

Expected:
```text
ok  	backend/internal/bootstrap
```

---

## Task 6: Add the four domain module skeletons, clients, and ping integration test

**Files:**
- Create: `test/integration/ping_test.go`
- Create: all files under `internal/modules/order/`
- Create: all files under `internal/modules/payment/`
- Create: all files under `internal/modules/inventory/`
- Create: all files under `internal/modules/notification/`
- Create: `internal/clients/paymentgateway/client.go`
- Create: `internal/clients/sms/client.go`
- Create: `internal/clients/email/client.go`

- [ ] **Step 1: Write the failing integration test for all ping routes**

Create `test/integration/ping_test.go`:

```go
package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/bootstrap"
	"backend/internal/modules/inventory"
	"backend/internal/modules/notification"
	"backend/internal/modules/order"
	"backend/internal/modules/payment"
	"backend/internal/platform/config"
	"backend/internal/platform/logger"
	"backend/internal/platform/observability"
	"backend/internal/platform/response"
	"github.com/gin-gonic/gin"
)

func TestBootstrapRegistersAllPingRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	deps := bootstrap.Dependencies{
		Config: config.Config{
			App:    config.AppConfig{Name: "backend-test", Env: "test"},
			Server: config.ServerConfig{Addr: ":18080"},
			Log:    config.LogConfig{Level: "debug"},
		},
		Logger:     logger.New("debug", io.Discard),
		Tracer:     observability.NewNoopTracer(),
		Propagator: observability.NewNoopPropagator(),
	}

	engine := bootstrap.NewAPIEngine(
		deps,
		order.NewHandler(order.NewService(order.NewRepository())),
		payment.NewHandler(payment.NewService(payment.NewRepository())),
		inventory.NewHandler(inventory.NewService(inventory.NewRepository())),
		notification.NewHandler(notification.NewService(notification.NewRepository())),
	)

	tests := []struct {
		name   string
		path   string
		module string
	}{
		{name: "order", path: "/api/v1/orders/ping", module: "order"},
		{name: "payment", path: "/api/v1/payments/ping", module: "payment"},
		{name: "inventory", path: "/api/v1/inventories/ping", module: "inventory"},
		{name: "notification", path: "/api/v1/notifications/ping", module: "notification"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)
			request.Header.Set("X-Trace-ID", "trace-int")

			engine.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusOK {
				t.Fatalf("expected 200 for %s, got %d", tt.path, recorder.Code)
			}
			if recorder.Header().Get("X-Request-ID") == "" {
				t.Fatalf("expected X-Request-ID header for %s", tt.path)
			}
			if recorder.Header().Get("X-Trace-ID") != "trace-int" {
				t.Fatalf("expected X-Trace-ID trace-int for %s, got %q", tt.path, recorder.Header().Get("X-Trace-ID"))
			}

			var body response.APIResponse[map[string]any]
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatalf("json.Unmarshal returned error: %v", err)
			}
			if body.Data["module"] != tt.module {
				t.Fatalf("expected module %s, got %#v", tt.module, body.Data["module"])
			}
			if body.Data["trace_id"] != "trace-int" {
				t.Fatalf("expected trace_id trace-int, got %#v", body.Data["trace_id"])
			}
		})
	}
}
```

- [ ] **Step 2: Run the integration test to verify it fails**

Run:
```bash
go test ./test/integration -run TestBootstrapRegistersAllPingRoutes -v
```

Expected: FAIL because the four module packages and constructors do not exist yet.

- [ ] **Step 3: Implement the order module**

Create these files exactly:

```go
// internal/modules/order/dto.go
package order

type CreateOrderRequest struct {
	UserID int64  `json:"user_id"`
	SKU    string `json:"sku"`
	Qty    int    `json:"qty"`
}

type OrderResponse struct {
	OrderNo string `json:"order_no"`
	Status  string `json:"status"`
}

type PingResponse struct {
	Module  string `json:"module"`
	TraceID string `json:"trace_id,omitempty"`
}
```

```go
// internal/modules/order/model.go
package order

type Order struct {
	OrderNo string
	UserID  int64
	SKU     string
	Qty     int
	Status  string
}
```

```go
// internal/modules/order/status.go
package order

const (
	OrderPending   = "pending"
	OrderPaid      = "paid"
	OrderCancelled = "cancelled"
)
```

```go
// internal/modules/order/events.go
package order

const (
	EventOrderCreated = "order.created.v1"
	EventOrderPaid    = "order.paid.v1"
)
```

```go
// internal/modules/order/repository.go
package order

import "context"

type Repository interface {
	Ping(ctx context.Context) error
}

type noopRepository struct{}

func NewRepository() Repository {
	return noopRepository{}
}

func (noopRepository) Ping(ctx context.Context) error {
	return nil
}
```

```go
// internal/modules/order/service.go
package order

import "context"

type Service interface {
	Ping(ctx context.Context) (PingResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Ping(ctx context.Context) (PingResponse, error) {
	if err := s.repo.Ping(ctx); err != nil {
		return PingResponse{}, err
	}
	return PingResponse{Module: "order"}, nil
}
```

```go
// internal/modules/order/handler.go
package order

import (
	"backend/internal/platform/observability"
	"backend/internal/platform/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/orders/ping", h.Ping)
}

func (h *Handler) Ping(c *gin.Context) {
	payload, err := h.service.Ping(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	if traceID, ok := observability.TraceIDFromContext(c.Request.Context()); ok {
		payload.TraceID = traceID
	}
	response.Success(c, payload)
}
```

- [ ] **Step 4: Implement the payment module**

Create these files exactly:

```go
// internal/modules/payment/dto.go
package payment

type CreatePaymentRequest struct {
	OrderNo string `json:"order_no"`
	Amount  int64  `json:"amount"`
	Channel string `json:"channel"`
}

type PaymentResponse struct {
	PaymentNo string `json:"payment_no"`
	Status    string `json:"status"`
}

type PingResponse struct {
	Module  string `json:"module"`
	TraceID string `json:"trace_id,omitempty"`
}
```

```go
// internal/modules/payment/model.go
package payment

type PaymentOrder struct {
	PaymentNo string
	OrderNo   string
	Amount    int64
	Status    string
}
```

```go
// internal/modules/payment/status.go
package payment

const (
	PaymentInit      = "init"
	PaymentSucceeded = "succeeded"
	PaymentFailed    = "failed"
)
```

```go
// internal/modules/payment/events.go
package payment

const (
	EventPaymentCreated   = "payment.created.v1"
	EventPaymentSucceeded = "payment.succeeded.v1"
)
```

```go
// internal/modules/payment/repository.go
package payment

import "context"

type Repository interface {
	Ping(ctx context.Context) error
}

type noopRepository struct{}

func NewRepository() Repository {
	return noopRepository{}
}

func (noopRepository) Ping(ctx context.Context) error {
	return nil
}
```

```go
// internal/modules/payment/service.go
package payment

import "context"

type Service interface {
	Ping(ctx context.Context) (PingResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Ping(ctx context.Context) (PingResponse, error) {
	if err := s.repo.Ping(ctx); err != nil {
		return PingResponse{}, err
	}
	return PingResponse{Module: "payment"}, nil
}
```

```go
// internal/modules/payment/handler.go
package payment

import (
	"backend/internal/platform/observability"
	"backend/internal/platform/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/payments/ping", h.Ping)
}

func (h *Handler) Ping(c *gin.Context) {
	payload, err := h.service.Ping(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	if traceID, ok := observability.TraceIDFromContext(c.Request.Context()); ok {
		payload.TraceID = traceID
	}
	response.Success(c, payload)
}
```

- [ ] **Step 5: Implement the inventory module**

Create these files exactly:

```go
// internal/modules/inventory/dto.go
package inventory

type InventoryResponse struct {
	SKU       string `json:"sku"`
	Available int    `json:"available"`
}

type PingResponse struct {
	Module  string `json:"module"`
	TraceID string `json:"trace_id,omitempty"`
}
```

```go
// internal/modules/inventory/model.go
package inventory

type Inventory struct {
	SKU       string
	Available int
	Reserved  int
	Status    string
}
```

```go
// internal/modules/inventory/status.go
package inventory

const (
	InventoryReady    = "ready"
	InventoryReserved = "reserved"
	InventoryReleased = "released"
)
```

```go
// internal/modules/inventory/events.go
package inventory

const (
	EventInventoryReserved = "inventory.reserved.v1"
	EventInventoryReleased = "inventory.released.v1"
)
```

```go
// internal/modules/inventory/repository.go
package inventory

import "context"

type Repository interface {
	Ping(ctx context.Context) error
}

type noopRepository struct{}

func NewRepository() Repository {
	return noopRepository{}
}

func (noopRepository) Ping(ctx context.Context) error {
	return nil
}
```

```go
// internal/modules/inventory/service.go
package inventory

import "context"

type Service interface {
	Ping(ctx context.Context) (PingResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Ping(ctx context.Context) (PingResponse, error) {
	if err := s.repo.Ping(ctx); err != nil {
		return PingResponse{}, err
	}
	return PingResponse{Module: "inventory"}, nil
}
```

```go
// internal/modules/inventory/handler.go
package inventory

import (
	"backend/internal/platform/observability"
	"backend/internal/platform/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/inventories/ping", h.Ping)
}

func (h *Handler) Ping(c *gin.Context) {
	payload, err := h.service.Ping(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	if traceID, ok := observability.TraceIDFromContext(c.Request.Context()); ok {
		payload.TraceID = traceID
	}
	response.Success(c, payload)
}
```

- [ ] **Step 6: Implement the notification module**

Create these files exactly:

```go
// internal/modules/notification/dto.go
package notification

type SendNotificationRequest struct {
	Channel string `json:"channel"`
	Target  string `json:"target"`
	Body    string `json:"body"`
}

type NotificationResponse struct {
	TaskNo  string `json:"task_no"`
	Status  string `json:"status"`
	Channel string `json:"channel"`
}

type PingResponse struct {
	Module  string `json:"module"`
	TraceID string `json:"trace_id,omitempty"`
}
```

```go
// internal/modules/notification/model.go
package notification

type NotificationTask struct {
	TaskNo   string
	Channel  string
	Target   string
	Body     string
	Status   string
}
```

```go
// internal/modules/notification/status.go
package notification

const (
	NotificationPending = "pending"
	NotificationSent    = "sent"
	NotificationFailed  = "failed"
)
```

```go
// internal/modules/notification/events.go
package notification

const (
	EventNotificationCreated = "notification.created.v1"
	EventNotificationFailed  = "notification.failed.v1"
)
```

```go
// internal/modules/notification/repository.go
package notification

import "context"

type Repository interface {
	Ping(ctx context.Context) error
}

type noopRepository struct{}

func NewRepository() Repository {
	return noopRepository{}
}

func (noopRepository) Ping(ctx context.Context) error {
	return nil
}
```

```go
// internal/modules/notification/service.go
package notification

import "context"

type Service interface {
	Ping(ctx context.Context) (PingResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Ping(ctx context.Context) (PingResponse, error) {
	if err := s.repo.Ping(ctx); err != nil {
		return PingResponse{}, err
	}
	return PingResponse{Module: "notification"}, nil
}
```

```go
// internal/modules/notification/handler.go
package notification

import (
	"backend/internal/platform/observability"
	"backend/internal/platform/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/notifications/ping", h.Ping)
}

func (h *Handler) Ping(c *gin.Context) {
	payload, err := h.service.Ping(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	if traceID, ok := observability.TraceIDFromContext(c.Request.Context()); ok {
		payload.TraceID = traceID
	}
	response.Success(c, payload)
}
```

- [ ] **Step 7: Create the client interfaces**

Create these files exactly:

```go
// internal/clients/paymentgateway/client.go
package paymentgateway

import "context"

type CreatePaymentRequest struct {
	OrderNo  string
	Amount   int64
	Currency string
}

type CreatePaymentResponse struct {
	PaymentNo string
	Status    string
}

type Client interface {
	CreatePayment(ctx context.Context, req CreatePaymentRequest) (*CreatePaymentResponse, error)
}
```

```go
// internal/clients/sms/client.go
package sms

import "context"

type Client interface {
	Send(ctx context.Context, to string, message string) error
}
```

```go
// internal/clients/email/client.go
package email

import "context"

type Client interface {
	Send(ctx context.Context, to string, subject string, body string) error
}
```

- [ ] **Step 8: Run the integration test again**

Run:
```bash
go test ./test/integration -run TestBootstrapRegistersAllPingRoutes -v
```

Expected:
```text
=== RUN   TestBootstrapRegistersAllPingRoutes
=== RUN   TestBootstrapRegistersAllPingRoutes/order
=== RUN   TestBootstrapRegistersAllPingRoutes/payment
=== RUN   TestBootstrapRegistersAllPingRoutes/inventory
=== RUN   TestBootstrapRegistersAllPingRoutes/notification
--- PASS: TestBootstrapRegistersAllPingRoutes
PASS
```

---

## Task 7: Implement the worker lifecycle and optional probe recognition

**Files:**
- Create: `internal/bootstrap/worker_test.go`
- Create: `internal/bootstrap/worker.go`

- [ ] **Step 1: Write the failing worker tests**

Create `internal/bootstrap/worker_test.go`:

```go
package bootstrap

import (
	"context"
	"io"
	"testing"
	"time"

	"backend/internal/platform/logger"
)

type stubTask struct{}

func (stubTask) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (stubTask) Probe(ctx context.Context) error {
	return nil
}

func TestWorkerRunStopsOnContextCancel(t *testing.T) {
	worker := NewWorker(logger.New("debug", io.Discard))
	worker.RegisterTask("placeholder", func(ctx context.Context) error {
		<-ctx.Done()
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- worker.Run(ctx)
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("expected nil error on cancellation, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("worker.Run did not exit after cancellation")
	}
}

func TestWorkerProbeRecognizesOptionalTaskProbe(t *testing.T) {
	worker := NewWorker(logger.New("debug", io.Discard))
	worker.RegisterRunnable("probeable", stubTask{})

	if err := worker.Probe(context.Background()); err != nil {
		t.Fatalf("expected Probe to succeed, got %v", err)
	}
}
```

- [ ] **Step 2: Run the worker tests to verify they fail**

Run:
```bash
go test ./internal/bootstrap -run 'TestWorkerRunStopsOnContextCancel|TestWorkerProbeRecognizesOptionalTaskProbe' -v
```

Expected: FAIL because `NewWorker`, `RegisterRunnable`, `Run`, and `Probe` do not exist yet.

- [ ] **Step 3: Implement the worker runtime**

Create `internal/bootstrap/worker.go`:

```go
package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

type Worker interface {
	RegisterTask(name string, handler func(ctx context.Context) error)
	Run(ctx context.Context) error
}

type RunnableTask interface {
	Run(ctx context.Context) error
}

type TaskProbe interface {
	Probe(ctx context.Context) error
}

type InMemoryWorker struct {
	log    *slog.Logger
	mu     sync.Mutex
	tasks  map[string]func(ctx context.Context) error
	probes map[string]TaskProbe
}

func NewWorker(log *slog.Logger) *InMemoryWorker {
	return &InMemoryWorker{
		log:    log,
		tasks:  make(map[string]func(ctx context.Context) error),
		probes: make(map[string]TaskProbe),
	}
}

func (w *InMemoryWorker) RegisterTask(name string, handler func(ctx context.Context) error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.tasks[name] = handler
}

func (w *InMemoryWorker) RegisterRunnable(name string, task RunnableTask) {
	w.RegisterTask(name, task.Run)
	if probe, ok := task.(TaskProbe); ok {
		w.mu.Lock()
		w.probes[name] = probe
		w.mu.Unlock()
	}
}

func (w *InMemoryWorker) Probe(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	for name, probe := range w.probes {
		if err := probe.Probe(ctx); err != nil {
			return fmt.Errorf("probe %s: %w", name, err)
		}
	}
	return nil
}

func (w *InMemoryWorker) Run(ctx context.Context) error {
	w.mu.Lock()
	copied := make(map[string]func(ctx context.Context) error, len(w.tasks))
	for name, task := range w.tasks {
		copied[name] = task
	}
	w.mu.Unlock()

	var wg sync.WaitGroup
	errCh := make(chan error, len(copied))

	for name, task := range copied {
		wg.Add(1)
		go func(name string, task func(ctx context.Context) error) {
			defer wg.Done()
			if err := task(ctx); err != nil && err != context.Canceled {
				errCh <- fmt.Errorf("task %s: %w", name, err)
			}
		}(name, task)
	}

	<-ctx.Done()
	wg.Wait()

	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}
```

- [ ] **Step 4: Run the worker tests again**

Run:
```bash
go test ./internal/bootstrap -run 'TestWorkerRunStopsOnContextCancel|TestWorkerProbeRecognizesOptionalTaskProbe' -v
```

Expected:
```text
ok  	backend/internal/bootstrap
```

---

## Task 8: Wire the API/worker entrypoints and final engineering commands

**Files:**
- Create: `cmd/api/main.go`
- Create: `cmd/worker/main.go`
- Create: `Makefile`

- [ ] **Step 1: Prove the entrypoints do not exist yet**

Run:
```bash
go run ./cmd/api
```

Expected: FAIL because `cmd/api/main.go` does not exist yet.

Then run:
```bash
go run ./cmd/worker
```

Expected: FAIL because `cmd/worker/main.go` does not exist yet.

- [ ] **Step 2: Implement the API main package**

Create `cmd/api/main.go`:

```go
package main

import (
	"log"
	"os"

	"backend/internal/bootstrap"
	"backend/internal/modules/inventory"
	"backend/internal/modules/notification"
	"backend/internal/modules/order"
	"backend/internal/modules/payment"
	"backend/internal/platform/config"
	"backend/internal/platform/logger"
	"backend/internal/platform/observability"
)

func main() {
	configPath := os.Getenv("APP_CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.local.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	appLogger := logger.New(cfg.Log.Level, os.Stdout)
	deps := bootstrap.Dependencies{
		Config:     cfg,
		Logger:     appLogger,
		Tracer:     observability.NewNoopTracer(),
		Propagator: observability.NewNoopPropagator(),
	}

	engine := bootstrap.NewAPIEngine(
		deps,
		order.NewHandler(order.NewService(order.NewRepository())),
		payment.NewHandler(payment.NewService(payment.NewRepository())),
		inventory.NewHandler(inventory.NewService(inventory.NewRepository())),
		notification.NewHandler(notification.NewService(notification.NewRepository())),
	)

	appLogger.Info("api starting", "addr", cfg.Server.Addr, "config", cfg.MaskedSummary())
	if err := engine.Run(cfg.Server.Addr); err != nil {
		appLogger.Error("api stopped", "error", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 3: Implement the worker main package**

Create `cmd/worker/main.go`:

```go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"backend/internal/bootstrap"
	"backend/internal/platform/config"
	"backend/internal/platform/logger"
)

type placeholderTask struct{}

func (placeholderTask) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (placeholderTask) Probe(ctx context.Context) error {
	return nil
}

func main() {
	configPath := os.Getenv("APP_CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.local.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	appLogger := logger.New(cfg.Log.Level, os.Stdout)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	worker := bootstrap.NewWorker(appLogger)
	worker.RegisterRunnable("placeholder", placeholderTask{})

	appLogger.Info("worker starting", "config", cfg.MaskedSummary())
	if err := worker.Run(ctx); err != nil {
		appLogger.Error("worker stopped", "error", err)
		os.Exit(1)
	}
	appLogger.Info("worker stopped cleanly")
}
```

- [ ] **Step 4: Add the Makefile**

Create `Makefile`:

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

- [ ] **Step 5: Run formatting and the full test suite**

Run:
```bash
go fmt ./... && go test ./...
```

Expected:
```text
ok  	backend/internal/platform/config
ok  	backend/internal/platform/errors
ok  	backend/internal/platform/response
ok  	backend/internal/platform/observability
ok  	backend/internal/platform/logger
ok  	backend/internal/bootstrap
ok  	backend/test/integration
```

- [ ] **Step 6: Smoke-test the API process**

Run:
```bash
APP_CONFIG_PATH=configs/config.local.yaml go run ./cmd/api >/tmp/backend-api.log 2>&1 & pid=$!; sleep 2; curl -i http://127.0.0.1:8080/healthz; kill $pid; wait $pid
```

Expected:
- `HTTP/1.1 200 OK`
- body contains `"code":"OK"`
- body contains `"request_id":`

- [ ] **Step 7: Smoke-test the worker process**

Run:
```bash
APP_CONFIG_PATH=configs/config.local.yaml go run ./cmd/worker >/tmp/backend-worker.log 2>&1 & pid=$!; sleep 2; kill -INT $pid; wait $pid
```

Expected:
- process exits cleanly
- `/tmp/backend-worker.log` contains `worker starting`
- `/tmp/backend-worker.log` contains `worker stopped cleanly`

---

## Spec Coverage Checklist

- `cmd/api` + `cmd/worker` entrypoint: covered by **Task 8**.
- 强类型配置 + `config.test.yaml` + 脱敏摘要：covered by **Task 1** and **Task 2**.
- `errors` / `response` 平台组件：covered by **Task 3**.
- `Tracer` / `Propagator` / `Carrier` / request-trace context：covered by **Task 4**.
- `slog` logger seam：covered by **Task 4**.
- request_id / trace / recovery middleware：covered by **Task 5**.
- `/healthz` and `/readyz`: covered by **Task 5**.
- 四个模块的 `handler/service/repository/dto/model/status/events`: covered by **Task 6**.
- clients 接口落点 + timeout/trace 扩展位：covered by **Task 6**.
- worker `RegisterTask` / `Run(ctx)` / optional `TaskProbe`: covered by **Task 7**.
- `test/integration/ping_test.go`: covered by **Task 6**.
- `Makefile`, `go test ./...`, `golangci-lint run`: covered by **Task 8**.

## Placeholder Scan

- No `TBD`
- No `TODO`
- No “similar to Task N” references
- No undefined file paths outside this plan

## Type Consistency Check

- `bootstrap.NewAPIEngine` expects `HTTPRouteRegistrar`; each module `Handler` provides `RegisterRoutes(group *gin.RouterGroup)`.
- `bootstrap.NewWorker` returns `*InMemoryWorker`; core `Worker` interface remains `RegisterTask` + `Run` only.
- `TaskProbe` is optional and not embedded into the `Worker` interface.
- `response.Success` / `response.Fail` always emit `request_id`.

## Final Verification Gate

Before declaring the bootstrap complete, re-run exactly:

```bash
go fmt ./...
go test ./...
golangci-lint run
APP_CONFIG_PATH=configs/config.local.yaml go run ./cmd/api
APP_CONFIG_PATH=configs/config.local.yaml go run ./cmd/worker
```

The work is complete only when:
- all tests pass
- lint passes
- API starts successfully
- worker starts and exits cleanly on signal
- no DB / Redis / MQ / exporter / real business flow slipped into this first pass
