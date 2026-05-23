# Admin Auth Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让 `admin/` 前端通过 `backend/` 的真实 `login / refresh / logout / me` 接口完成登录态管理，并保留现有前端菜单模式。

**Architecture:** 先把 `backend/internal/modules/user` 与 `backend/internal/modules/auth` 的 repository 从 noop 实现推进到真实 SQL，再修改 `admin/` 的 auth API、user store、HTTP 拦截器和路由守卫，最后通过前后端联调验证真实登录、静默刷新和退出。前端继续使用 `frontend` 菜单模式，不引入后端菜单接口。

**Tech Stack:** Go 1.26, Gin, database/sql, pgx, Vue 3, Pinia, Vue Router, Axios, Vite

---

## Preconditions

- 后端工作目录为 `/Users/chening/Desktop/zuhao/backend`。
- 前端工作目录为 `/Users/chening/Desktop/zuhao/admin`。
- 本计划严格以 [2026-05-22-admin-auth-integration-design.md](/Users/chening/Desktop/zuhao/backend/docs/superpowers/specs/2026-05-22-admin-auth-integration-design.md) 为范围约束。
- 本轮不做后端菜单模式，不修改 `VITE_ACCESS_MODE=frontend`。
- 本轮会碰前后端两个工程；提交时要按任务边界分批提交，避免混杂。

## File Structure Lock-in

### Backend repository and tests
- Modify: `internal/modules/user/repository.go`
- Create: `internal/modules/user/repository_sql_test.go`
- Modify: `internal/modules/auth/repository.go`
- Create: `internal/modules/auth/repository_sql_test.go`
- Modify: `internal/modules/auth/service_test.go`
- Modify: `cmd/api/main.go`
- Modify: `test/integration/test_helpers.go`

### Admin auth integration
- Modify: `admin/.env`
- Modify: `admin/.env.development`
- Modify: `admin/src/api/auth.ts`
- Modify: `admin/src/types/api/api.d.ts`
- Modify: `admin/src/types/common/response.ts`
- Modify: `admin/src/store/modules/user.ts`
- Modify: `admin/src/utils/http/index.ts`
- Modify: `admin/src/utils/http/error.ts`
- Modify: `admin/src/router/guards/beforeEach.ts`
- Modify: `admin/src/router/core/MenuProcessor.ts`
- Create: `admin/src/utils/http/auth-refresh.ts`

### Frontend verification
- Create: `admin/src/api/auth.test.ts` or `admin/src/utils/http/auth-refresh.test.ts` if current toolchain supports local unit tests cleanly
- If no stable frontend unit test harness exists, verification falls back to `pnpm build` + browser/manual flow checks

---

## Task 1: Replace noop user repository with real SQL queries

**Files:**
- Modify: `internal/modules/user/repository.go`
- Create: `internal/modules/user/repository_sql_test.go`

- [ ] **Step 1: Write the failing repository tests**

Add tests to `internal/modules/user/repository_sql_test.go` covering:

```go
func TestRepositoryGetByEmailReturnsNormalizedUser(t *testing.T) {}
func TestRepositoryGetByIDReturnsUser(t *testing.T) {}
```

Each test should:

- create a temporary Postgres test database using the existing integration DB helper pattern
- run `migrations/00002_create_auth_tables.sql`
- insert a `users` row
- call the repository method
- assert returned `User` fields match DB values

- [ ] **Step 2: Run the user repository tests to verify they fail**

Run:

```bash
go test ./internal/modules/user -run 'TestRepository(GetByEmail|GetByID)' -v
```

Expected: FAIL because the repository still returns `nil, nil`.

- [ ] **Step 3: Implement the SQL-backed user repository**

Replace the noop implementation in `internal/modules/user/repository.go` with a concrete repository that:

- keeps the same `Repository` interface
- stores `dbtx database.DBTX`
- exposes:

```go
func NewRepository(dbtx database.DBTX) Repository
```

- uses `database.ExecutorFromContext(ctx, r.dbtx)` inside methods
- queries:

```sql
SELECT id, public_id, email, password_hash, status, password_changed_at, last_login_at, created_at, updated_at
FROM users
WHERE email = $1
```

and

```sql
SELECT id, public_id, email, password_hash, status, password_changed_at, last_login_at, created_at, updated_at
FROM users
WHERE id = $1
```

- returns `nil, nil` on `sql.ErrNoRows`

- [ ] **Step 4: Run the user repository tests again**

Run:

```bash
go test ./internal/modules/user -run 'TestRepository(GetByEmail|GetByID)' -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/modules/user/repository.go internal/modules/user/repository_sql_test.go
git commit -m "实现用户仓储真实 SQL 查询"
```

## Task 2: Replace noop auth repository with real SQL persistence

**Files:**
- Modify: `internal/modules/auth/repository.go`
- Create: `internal/modules/auth/repository_sql_test.go`

- [ ] **Step 1: Write the failing auth repository tests**

Add tests to `internal/modules/auth/repository_sql_test.go` covering at least:

```go
func TestRepositoryListUserRolesAndPermissions(t *testing.T) {}
func TestRepositoryCreateSessionAndGetSessionByID(t *testing.T) {}
func TestRepositoryCreateRefreshTokenAndGetByHash(t *testing.T) {}
func TestRepositoryMarkRefreshTokenUsed(t *testing.T) {}
func TestRepositoryRevokeRefreshTokenFamily(t *testing.T) {}
func TestRepositoryWriteLoginAttemptAndAuditLog(t *testing.T) {}
```

- [ ] **Step 2: Run the auth repository tests to verify they fail**

Run:

```bash
go test ./internal/modules/auth -run 'TestRepository' -v
```

Expected: FAIL because the repository methods are still noop.

- [ ] **Step 3: Implement the SQL-backed auth repository**

In `internal/modules/auth/repository.go`, replace noop logic with SQL-backed methods using `database.DBTX`.

Required queries include:

- roles:

```sql
SELECT r.code
FROM roles r
JOIN user_roles ur ON ur.role_id = r.id
WHERE ur.user_id = $1
ORDER BY r.code
```

- permissions:

```sql
SELECT DISTINCT p.code
FROM permissions p
JOIN role_permissions rp ON rp.permission_id = p.id
JOIN user_roles ur ON ur.role_id = rp.role_id
WHERE ur.user_id = $1
ORDER BY p.code
```

- create session / get session / revoke session
- create refresh token / lookup by token hash `FOR UPDATE`
- atomic use-marking:

```sql
UPDATE refresh_tokens
SET used_at = NOW(), replaced_by_token_id = $2
WHERE id = $1 AND used_at IS NULL AND revoked_at IS NULL
```

- family revoke:

```sql
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE family_id = $1 AND revoked_at IS NULL
```

- update `users.last_login_at`
- insert `login_attempts`
- insert `audit_logs`

Constructor shape:

```go
func NewRepository(dbtx database.DBTX) Repository
```

- [ ] **Step 4: Run the auth repository tests again**

Run:

```bash
go test ./internal/modules/auth -run 'TestRepository' -v
```

Expected: PASS.

- [ ] **Step 5: Re-run service tests against the real repository interface**

Run:

```bash
go test ./internal/modules/auth -run 'TestService' -v
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/modules/auth/repository.go internal/modules/auth/repository_sql_test.go internal/modules/auth/service_test.go
git commit -m "实现认证仓储真实 SQL 持久化"
```

## Task 3: Wire real repositories into backend API composition

**Files:**
- Modify: `cmd/api/main.go`
- Modify: `test/integration/test_helpers.go`

- [ ] **Step 1: Write a failing integration test for real auth composition**

Add a test to `test/integration/auth_me_test.go` or a new file:

```go
func TestAuthLoginWithNoSeededUserReturnsUnauthorized(t *testing.T) {}
```

Use the real `newIntegrationEngine(...)` and assert the route responds through the real stack.

- [ ] **Step 2: Run the integration auth tests to verify current composition is insufficient**

Run:

```bash
go test ./test/integration -run 'TestAuth(Login|Refresh|Me)' -v
```

Expected: FAIL if constructors or SQL dependencies are still wired incorrectly.

- [ ] **Step 3: Update backend composition to use SQL repositories**

In `cmd/api/main.go`:

- keep existing `security.NewArgon2idHasher(...)`
- keep `security.NewHMACTokenManager(...)`
- replace:

```go
user.NewRepository()
auth.NewRepository()
```

with:

```go
user.NewRepository(nil)
auth.NewRepository(nil)
```

or pass the concrete DB executor if available from the chosen DB layer.

In `test/integration/test_helpers.go`, make the same constructor change so integration uses the real repository implementation shape.

- [ ] **Step 4: Run integration auth tests again**

Run:

```bash
go test ./test/integration -run 'TestAuth(Login|Refresh|Me)' -v
```

Expected: PASS for the current minimal auth integration suite.

- [ ] **Step 5: Commit**

```bash
git add cmd/api/main.go test/integration/test_helpers.go test/integration/auth_*.go
git commit -m "调整后端装配使用真实认证仓储"
```

## Task 4: Switch admin auth API contract to backend `/api/v1/auth/*`

**Files:**
- Modify: `admin/src/api/auth.ts`
- Modify: `admin/src/types/api/api.d.ts`
- Modify: `admin/src/types/common/response.ts`

- [ ] **Step 1: Write the failing frontend contract checks**

If a frontend unit test harness is already practical, add a focused test file. If not, use a static grep checkpoint first:

Run:

```bash
rg -n "/api/auth/login|/api/user/info|refreshToken|userName: string" admin/src/api/auth.ts admin/src/types/api/api.d.ts admin/src/store/modules/user.ts
```

Expected: hits show old auth contract is still present.

- [ ] **Step 2: Update auth API wrappers**

In `admin/src/api/auth.ts`:

- map login request to backend payload:

```ts
export function fetchLogin(params: Api.Auth.LoginParams) {
  return request.post<Api.Auth.LoginResponse>({
    url: '/api/v1/auth/login',
    params: {
      identifier: params.userName,
      password: params.password
    }
  })
}
```

- add:

```ts
export function fetchRefreshToken() {
  return request.post<Api.Auth.LoginResponse>({
    url: '/api/v1/auth/refresh'
  })
}

export function fetchLogout() {
  return request.post({
    url: '/api/v1/auth/logout'
  })
}

export function fetchGetUserInfo() {
  return request.get<Api.Auth.UserInfo>({
    url: '/api/v1/auth/me'
  })
}
```

- [ ] **Step 3: Update auth-related types**

In `admin/src/types/api/api.d.ts`, replace mock auth shapes with backend-aligned ones:

```ts
interface LoginResponse {
  access_token: string
  token_type: string
  expires_in: number
  user: {
    id: string
    roles: string[]
    permissions?: string[]
  }
}

interface UserInfo {
  id: string
  roles: string[]
  permissions: string[]
  session_id: string
  userId?: string
  userName?: string
  avatar?: string
}
```

Keep `BaseResponse` unchanged except ensuring both `message` and `msg` remain supported.

- [ ] **Step 4: Run a frontend contract grep**

Run:

```bash
rg -n "/api/v1/auth/login|/api/v1/auth/me|/api/v1/auth/refresh|/api/v1/auth/logout" admin/src/api/auth.ts
```

Expected: all four real backend endpoints are present.

- [ ] **Step 5: Commit**

```bash
git add admin/src/api/auth.ts admin/src/types/api/api.d.ts admin/src/types/common/response.ts
git commit -m "切换前端认证 API 到真实后端契约"
```

## Task 5: Remove frontend refresh token storage and normalize user snapshot mapping

**Files:**
- Modify: `admin/src/store/modules/user.ts`

- [ ] **Step 1: Write the failing store assertions**

If frontend unit tests are available, add tests. Otherwise use static checkpoints:

```bash
rg -n "refreshToken|setToken\(|userId: number|userName" admin/src/store/modules/user.ts
```

Expected: hits show `refreshToken` is still locally stored.

- [ ] **Step 2: Update the user store**

In `admin/src/store/modules/user.ts`:

- remove `refreshToken` state
- make `setToken(newAccessToken: string)` only store access token
- update `setUserInfo()` to map backend `me` snapshot into current UI-compatible shape
- ensure `logOut()` clears only local user info, login flag, lock state, and access token

Recommended mapping inside `setUserInfo`:

```ts
const setUserInfo = (newInfo: Api.Auth.UserInfo) => {
  info.value = {
    ...newInfo,
    userId: newInfo.userId || newInfo.id,
    userName: newInfo.userName || newInfo.id,
    avatar: newInfo.avatar || ''
  }
}
```

- [ ] **Step 3: Run the store grep again**

Run:

```bash
rg -n "refreshToken" admin/src/store/modules/user.ts
```

Expected: no matches.

- [ ] **Step 4: Commit**

```bash
git add admin/src/store/modules/user.ts
git commit -m "移除前端 refreshToken 本地存储"
```

## Task 6: Implement silent refresh and request replay in admin HTTP client

**Files:**
- Create: `admin/src/utils/http/auth-refresh.ts`
- Modify: `admin/src/utils/http/index.ts`
- Modify: `admin/src/utils/http/error.ts`

- [ ] **Step 1: Write the failing auth refresh flow checks**

If a test harness exists, add targeted tests. Otherwise create explicit review checkpoints in code comments and verify through build/manual flow later.

- [ ] **Step 2: Extract refresh coordination helper**

Create `admin/src/utils/http/auth-refresh.ts` with a single-flight refresh coordinator:

```ts
let refreshPromise: Promise<string> | null = null

export async function ensureFreshAccessToken(refreshFn: () => Promise<Api.Auth.LoginResponse>) {
  if (!refreshPromise) {
    refreshPromise = refreshFn()
      .then((data) => data.access_token)
      .finally(() => {
        refreshPromise = null
      })
  }
  return refreshPromise
}
```

- [ ] **Step 3: Update HTTP interceptor**

In `admin/src/utils/http/index.ts`:

- keep `withCredentials: VITE_WITH_CREDENTIALS === 'true'`
- set auth header as:

```ts
if (accessToken) request.headers.set('Authorization', `Bearer ${accessToken}`)
```

- on 401:
  - skip refresh for `/api/v1/auth/refresh` and `/api/v1/auth/logout`
  - call `fetchRefreshToken()` through `ensureFreshAccessToken(...)`
  - update `userStore.setToken(newAccessToken)`
  - replay the original request once
  - if refresh fails, call `userStore.logOut()`

- [ ] **Step 4: Keep error semantics stable**

Ensure `admin/src/utils/http/error.ts` still handles network errors, but does not override auth refresh logic implemented at interceptor level.

- [ ] **Step 5: Run frontend build to verify the request chain compiles**

Run:

```bash
cd /Users/chening/Desktop/zuhao/admin && pnpm build
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add admin/src/utils/http/auth-refresh.ts admin/src/utils/http/index.ts admin/src/utils/http/error.ts
git commit -m "实现前端静默刷新与请求重放"
```

## Task 7: Decouple `me` from menu initialization in route guard

**Files:**
- Modify: `admin/src/router/guards/beforeEach.ts`
- Modify: `admin/src/router/core/MenuProcessor.ts`

- [ ] **Step 1: Write a failing route-flow checkpoint**

Use a static review checkpoint:

```bash
rg -n "fetchGetUserInfo|getMenuList|routeInitFailed|routeInitInProgress" admin/src/router/guards/beforeEach.ts
```

Expected: current flow shows user info fetch and menu initialization tightly coupled.

- [ ] **Step 2: Update the guard logic**

Modify `fetchUserInfo()` in `beforeEach.ts` so it maps the backend `me` response shape into the user store and does not assume old mock fields exist.

Keep dynamic route registration flow, but ensure it still uses `MenuProcessor` in frontend mode without depending on backend menu APIs.

Required outcome:

- user identity comes from `/api/v1/auth/me`
- menu still comes from `asyncRoutes`
- route init does not fail only because backend menu APIs do not exist

- [ ] **Step 3: Run frontend build again**

Run:

```bash
cd /Users/chening/Desktop/zuhao/admin && pnpm build
```

Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add admin/src/router/guards/beforeEach.ts admin/src/router/core/MenuProcessor.ts
git commit -m "解耦前端 me 获取与菜单初始化"
```

## Task 8: Update admin development env and perform real end-to-end smoke validation

**Files:**
- Modify: `admin/.env`
- Modify: `admin/.env.development`

- [ ] **Step 1: Update env files for backend auth cookies**

Apply these changes:

In `admin/.env`:

```env
VITE_ACCESS_MODE = frontend
VITE_WITH_CREDENTIALS = true
```

In `admin/.env.development`:

```env
VITE_API_URL = /
VITE_API_PROXY_URL = http://127.0.0.1:8080
```

If local backend still uses `configs/config.local.yaml`, the default API port is `:8080`; only change this when local config differs.

- [ ] **Step 2: Run backend tests and frontend build as pre-flight**

Run:

```bash
cd /Users/chening/Desktop/zuhao/backend && go test ./...
cd /Users/chening/Desktop/zuhao/admin && pnpm build
```

Expected: both PASS.

- [ ] **Step 3: Start backend API**

Run:

```bash
cd /Users/chening/Desktop/zuhao/backend && make run-api
```

Expected: backend starts and listens on the configured local port.

- [ ] **Step 4: Start admin dev server**

Run:

```bash
cd /Users/chening/Desktop/zuhao/admin && pnpm dev
```

Expected: admin starts on `http://127.0.0.1:3006` or the configured Vite port.

- [ ] **Step 5: Verify the real auth flow in browser**

Using Browser or Playwright, verify:

- login page submits to the real backend
- login success sets the refresh cookie
- page refresh still preserves login state
- logout clears local state and cookie-backed session

- [ ] **Step 6: Commit**

```bash
git add admin/.env admin/.env.development
git commit -m "调整前端开发环境对接真实认证后端"
```

## Task 9: Final verification and regression pass

**Files:**
- Modify only files implicated by failing verification commands

- [ ] **Step 1: Run backend full verification**

Run:

```bash
cd /Users/chening/Desktop/zuhao/backend && go test ./... && golangci-lint run
```

Expected: PASS.

- [ ] **Step 2: Run frontend full verification**

Run:

```bash
cd /Users/chening/Desktop/zuhao/admin && pnpm build
```

Expected: PASS.

- [ ] **Step 3: Fix smallest possible regressions**

If any verification fails, patch only the implicated files, rerun the exact failing command, then rerun the broader suite.

- [ ] **Step 4: Commit verification fixes**

```bash
git add .
git commit -m "收口前后端真实认证对接验证问题"
```

## Spec Coverage Check

- 只做真实 `login / refresh / logout / me`：covered by **Task 2**, **Task 4**, **Task 6**, **Task 7**, **Task 8**.
- 前端保持 `frontend` 菜单模式：covered by **Task 7**, **Task 8**.
- 删除前端 `refreshToken` 本地存储：covered by **Task 5**.
- 401 静默 refresh + 请求重放：covered by **Task 6**.
- 后端 repository 从 noop 切到真实 SQL：covered by **Task 1**, **Task 2**, **Task 3**.
- `me` 仅作为身份快照：covered by **Task 4**, **Task 5**, **Task 7**.
- 联调与浏览器验收：covered by **Task 8**, **Task 9**.

## Placeholder Scan

- 没有 `TODO`、`TBD`、`implement later` 这类占位。
- 每个任务都给出精确文件路径、验证命令和预期结果。
- 对无法立即依赖前端单测框架的部分，计划明确使用 `pnpm build` 和真实浏览器联调作为当前仓库阶段的验证策略。

## Type Consistency Check

- 后端真实 auth 路由始终使用 `/api/v1/auth/*`。
- 前端 `LoginResponse` 始终对齐后端 `access_token / token_type / expires_in / user`。
- 前端 `UserInfo` 始终围绕后端 `me` 快照结构映射，而不是回退到旧 mock 契约。
- `refreshToken` 只存在于后端 cookie，不再存在于前端 store 类型和状态里。
