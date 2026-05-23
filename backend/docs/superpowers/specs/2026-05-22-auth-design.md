# Auth 认证与授权 P1 设计

> 设计日期：2026-05-22
> 范围：基于当前 `backend/` 的 `cmd -> bootstrap -> modules -> platform` 骨架，为 `admin/` 浏览器管理端落地第一版 Auth 认证、授权与会话能力。

## 1. 目标

本轮目标不是实现一个“大而全”的身份系统，而是在当前仓库阶段内，落地一个可用、可追踪、可扩展的认证闭环。

完成后应满足以下目标：

- 面向 `admin/` 浏览器端提供可用的登录态能力。
- 使用短期 `access token` + 服务端可控 `refresh token` 的双 token 模型。
- `refresh token` 通过 `HttpOnly Cookie` 承载，支持轮换、撤销与重放检测。
- 后端具备统一 `Principal`、鉴权中间件和权限校验能力。
- 认证相关错误、日志、审计、`request_id`、`trace_id` 延续现有平台规范。
- 保持 `platform/security`、`modules/auth`、`modules/user` 的清晰边界，为后续 MFA、密码重置、OAuth/SSO 留出扩展空间。

## 2. 本轮范围

### 2.1 会做的内容

本轮采用“方案 B：Auth 为主、User 最小建模、真实数据库落地”的实现路径，交付内容包括：

- 新增 `internal/platform/security`。
- 新增 `internal/modules/auth`。
- 新增最小 `internal/modules/user`，仅承载用户实体、状态和基础查询边界。
- 增加 Auth 相关数据库表、迁移和最小 seed 方案。
- 实现以下接口：
  - `POST /api/v1/auth/login`
  - `POST /api/v1/auth/refresh`
  - `POST /api/v1/auth/logout`
  - `GET /api/v1/auth/me`
- 实现 `AuthMiddleware`。
- 实现 `RequirePermission`。
- 更新 OpenAPI 中的 `bearerAuth` 与 refresh cookie 声明。
- 增加单元测试、handler/middleware 测试以及必要的 repository 集成测试设计位。

### 2.2 本轮不会做的内容

为了控制 P1 范围，本轮明确不做：

- 公开注册。
- 忘记密码 / 重置密码。
- MFA。
- OAuth / SSO / 第三方登录。
- access token denylist。
- 设备管理后台。
- 完整用户管理后台。
- 复杂风控、验证码、人机校验。
- 面向移动端或开放 API 客户端的多终端 token 返回策略。

## 3. 目标范围与模块边界

P1 要交付的是一个“可用于 `admin/` 管理端登录”的最小完整闭环，而不是完整用户中心。

模块边界按以下方式落地：

- `internal/platform/security`
  - 放通用安全原语：`Principal`、context helpers、密码哈希、JWT token manager、随机 token 生成、权限 helper。
  - 该层不依赖 `modules/auth`。
- `internal/modules/auth`
  - 负责认证、会话、刷新、退出、审计、登录尝试与权限装配。
  - 依赖 `platform/security`、`platform/database`、`platform/errors`、`platform/logger`。
- `internal/modules/user`
  - P1 只保留用户实体、状态常量与最小查询仓储。
  - 暂不扩展为完整用户中心模块。

这种拆法的核心目标是：

- `security` 负责跨模块复用能力。
- `auth` 负责登录态和认证流程。
- `user` 保持独立演进空间，但 P1 不把范围拉爆。

## 4. 数据模型与表结构裁剪方案

### 4.1 P1 实际落地表

P1 实际落地以下表：

- `users`
- `roles`
- `permissions`
- `user_roles`
- `role_permissions`
- `auth_sessions`
- `refresh_tokens`
- `login_attempts`
- `audit_logs`

`password_reset_tokens` 在设计中保留，但不进入 P1 迁移。

### 4.2 各表职责

#### `users`

仅承载登录所需最小信息：

- `public_id`
- `email`
- `password_hash`
- `status`
- `password_changed_at`
- `last_login_at`
- `created_at`
- `updated_at`

约束：

- P1 只支持邮箱登录，但字段名仍保留 `identifier` 语义以兼容未来扩展。
- 邮箱入库前统一做小写规范化。
- `password_hash` 不允许为空。
- `status` 至少支持：`active`、`disabled`、`locked`。
- `locked` 在 P1 中仅表示人工锁定状态，不表示自动锁定状态机。

#### `roles` / `permissions` / `user_roles` / `role_permissions`

作为标准 RBAC 基础表直接建齐。

P1 至少 seed：

- `admin`
- `system`

不采用 `admin:all` 这种超级通配，而采用显式权限集，避免授权模型出现双轨逻辑。

#### `auth_sessions`

表示一个受控登录会话：

- `id` 使用随机字符串，而不是自增 ID。
- `status` 至少包含 `active`、`revoked`。
- `expires_at` 必须存在，P1 中默认与当前 refresh token family 的有效期上限保持一致。
- `revoked_at` 必须存在，用于显式撤销语义。
- `last_seen_at` 建议存在，在 `login`、`refresh`、`logout` 这些显式认证事件中按需要更新。
- `user_agent_hash`、`ip_hash` 只用于审计与风控参考，不作为强校验条件。

为避免把 P1 做成“每个受保护请求都写库”，`last_seen_at` 在 P1 中不随着普通受保护接口访问更新。

P1 的 session 生命周期规则：

- session 在创建时写入 `created_at`、`expires_at`。
- `logout`、refresh reuse、管理员主动处理等场景通过 `revoked_at` 结束 session。
- 若既未撤销也未提前失效，则 session 在 `expires_at` 后视为过期。
- 历史 session 清理由后台任务按 `revoked_at` 或 `expires_at` 超过保留期后处理，P1 不要求实现该清理任务，但语义必须一致。

#### `refresh_tokens`

用于服务端控制 refresh token：

- 只存 `token_hash`。
- raw refresh token 使用 `crypto/rand` 生成至少 32 bytes 随机值。
- raw token 采用 `base64.RawURLEncoding` 编码后返回客户端。
- `token_hash` 使用 `SHA-256(raw_token)` 持久化；若后续引入专用 secret，可升级为 `HMAC-SHA-256(secret, raw_token)`，但 P1 保持单一算法。
- `token_hash` 的编码形式在 P1 中固定为十六进制小写字符串，以便索引与排查一致。
- 每次 refresh 都写入新记录。
- 旧 token 使用后标记 `used_at`。
- 使用 `family_id` 支持 token reuse detection 与整族撤销。

P1 中 refresh token 的比对以“先 hash 再查库”为主，不直接做明文 token 比较；若存在应用内二进制摘要比较，必须使用常量时间比较。

#### `login_attempts`

用于：

- 仅记录 login 行为。
- 面向登录风控、限流分析与撞库排查。
- 后续按 IP / identifier / user_id 限流扩展。

`login_attempts`、`audit_logs`、`auth_sessions` 中的客户端 IP 必须复用同一套 IP 提取逻辑：

- 仅信任应用显式配置的 trusted proxy 链。
- 从框架统一解析后的 client IP 读取，不信任任意原始 `X-Forwarded-For`。
- 若无法确认真实 client IP，则记录 `unknown` 并再做 hash，而不是盲目信任 header。

P1 先保证记录能力，不把账号锁定状态机做重。

#### `audit_logs`

用于记录：

- `login_success`
- `login_failed`
- `refresh_success`
- `refresh_reuse_detected`
- `session_revoked`
- `logout_success`

`audit_logs` 在 P1 中作为认证安全事件总线，记录 login / refresh / logout / revoke / reuse detection 等事件；`login_attempts` 不承载 refresh 或 logout 事件。

约束：

- 禁止记录 password、access token、refresh token、cookie 原文、Authorization header。
- `identifier` 只记录 hash 或脱敏值。

P1 中 `audit_logs` 的最小结构化字段集需要固定为：

- `event_type`
- `result`
- `user_public_id` 可空
- `session_id` 可空
- `request_id`
- `trace_id`
- `ip_hash`
- `user_agent_hash`
- `metadata_json`
- `occurred_at`

若事件只在登录前发生，`user_public_id` 允许为空。`session_mismatch=true` 这类附加信息统一放入 `metadata_json`。

`result` 在 P1 中固定枚举为：

- `success`
- `failure`

事件细分语义由 `event_type` 表达，`result` 不再引入第三套状态词。

P1 的认证事件枚举至少包含：

- `login_success`
- `login_failed`
- `refresh_success`
- `refresh_reuse_detected`
- `session_revoked`
- `logout_success`

其中 `session_revoked` 典型用于：

- refresh reuse 触发 family/session 撤销
- logout 主动撤销 session
- `password_changed_at` 导致当前 session/family 作废

### 4.3 关键约束与索引

为避免实现分叉，P1 迁移需要显式声明以下约束与索引：

- `users.public_id`：唯一索引。
- `users.email`：对规范化后邮箱建立唯一索引。
- `roles.code`：唯一索引。
- `permissions.code`：唯一索引。
- `user_roles(user_id, role_id)`：联合主键或联合唯一索引。
- `role_permissions(role_id, permission_id)`：联合主键或联合唯一索引。
- `auth_sessions.id`：主键。
- `refresh_tokens.token_hash`：唯一索引。
- `refresh_tokens.family_id`：普通索引。
- `refresh_tokens.session_id`：普通索引。
- `login_attempts(identifier_hash, created_at)`：组合索引。
- `login_attempts(ip_hash, created_at)`：组合索引。

`login_attempts` 在 P1 中优先存 `identifier_hash`，而不是原始 identifier。

### 4.4 seed 策略

P1 的 seed 采用保守策略：

- migration 中插入基础角色与基础权限。
- 默认管理员账号不写死在 schema migration 中。
- `local/dev` 可通过环境变量驱动创建默认管理员。
- `prod` 不自动创建管理员。

建议约定环境变量：

- `AUTH_BOOTSTRAP_ADMIN_EMAIL`
- `AUTH_BOOTSTRAP_ADMIN_PASSWORD`

## 5. 配置、安全材料与 Token 策略

### 5.1 Config 扩展

在当前 `internal/platform/config.Config` 上新增 `AuthConfig`，并在 `configs/config.*.yaml` 中增加：

```yaml
auth:
  issuer: backend
  audience: admin-api

  access_token_ttl: 15m
  refresh_token_ttl: 168h

  password:
    min_length: 12
    max_length: 128
    argon2_memory_kib: 19456
    argon2_iterations: 2
    argon2_parallelism: 1

  cookie:
    enabled: true
    name: refresh_token
    domain: ""
    path: "/api/v1/auth"
    secure: true
    http_only: true
    same_site: lax

  jwt:
    algorithm: HS256
    key_id: local-dev-key
```

P1 不启用自动账号锁定，因此不在 P1 配置中引入 `login.max_failed_attempts`、`failed_window`、`lock_duration` 这类会误导实现和运维的字段。

### 5.2 密钥与敏感配置来源

P1 采用以下策略：

- YAML 只存非敏感配置。
- `JWT_HMAC_SECRET` 从环境变量读取。
- 可选 `PASSWORD_PEPPER` 从环境变量读取。
- `local/dev` 允许显式受限的开发回退值。
- `prod` 如果缺失关键密钥，启动直接失败。

敏感值不得出现在：

- `MaskedSummary()` 输出。
- 启动日志。
- 错误响应。

P1 的 HMAC key 策略也需要明确：

- P1 只支持单活 HMAC key。
- `kid` 仅为未来轮换预留，不实现多 key keyring 验证。
- 当前 active key 从 `JWT_HMAC_SECRET` 读取。
- token 中 `kid` 缺失或未知时，验证直接失败。

### 5.3 Access Token 策略

P1 使用：

- JWT
- 算法：`HS256`
- TTL：15 分钟
- 载体：`Authorization: Bearer <access_token>`

claims 至少包含：

- `iss`
- `aud`
- `sub`
- `sid`
- `jti`
- `iat`
- `nbf`
- `exp`
- `roles`
- `permissions`

其中 `sub` 的语义在 P1 中必须固定为 `users.public_id`，而不是内部自增 `users.id`。约束如下：

- `sub = users.public_id`
- `UserID` 仅用于服务端数据库查询与关联
- 对外身份语义、token claims、`me` 响应、审计字段优先使用 `PublicID`

`roles` 与 `permissions` 在写入 `Principal`、JWT claims 与 `me` 响应前，必须先去重，并按稳定顺序排序；P1 统一采用字典序升序。

P1 默认假设 admin 用户的角色和权限集合规模有限；若单用户权限集合增长到可能显著扩大 JWT 体积时，应切换到角色压缩或服务端鉴权缓存方案，而不是无限制继续把全量 permissions 塞进 access token。

P1 将 `roles/permissions` 放入 access token，理由是：

- 降低每次受保护请求的回库成本。
- 保持 `AuthMiddleware` 简洁。
- 当前仓库尚未具备成熟缓存或认证网关。

代价也需要在设计中明确：

- 权限变更后，旧 access token 在 TTL 内仍可能继续有效。
- 下一次 refresh 会重新按数据库当前权限签发。
- `disabled` 或 `session revoked` 对 access token 的生效延迟上限等于 `access_token_ttl`，默认 15 分钟。

这项延迟是 P1 的明确取舍：`AuthMiddleware` 仅做无状态 JWT 校验，不在每次请求时回库查询 session 或 user 状态。

### 5.4 Refresh Token 策略

P1 使用：

- opaque 随机字符串
- 浏览器端通过 `HttpOnly Cookie` 承载
- 服务端只存 hash
- 每次 refresh 强制轮换
- 支持 revoke、reuse detection、family revoke

### 5.5 Cookie 策略

`refresh_token` cookie 建议：

- `HttpOnly=true`
- `Secure=true`，`local` 允许按本地开发方式受控放宽
- `SameSite=Lax`
- `Path=/api/v1/auth`

P1 的部署拓扑假设需要写死：

- `admin` 与 `api` 默认必须为 same-site 部署。
- 若不是 same-site，则当前 `SameSite=Lax` 的 refresh cookie 策略不成立。
- 若未来需要 cross-site 部署，则应切换到 `SameSite=None; Secure`，并补充专门的 CSRF 防护，这不属于 P1。

前端约定：

- 受保护接口通过 Bearer token 访问。
- `refresh/logout` 请求必须携带 `credentials: include`。
- CORS 继续使用显式 origin 白名单，不允许 `* + credentials`。

对 `refresh/logout` 这类依赖 cookie 的端点，P1 还必须增加：

- 仅允许受信 Origin。
- 校验 `Origin`，必要时回退校验 `Referer`。
- CORS 白名单不作为 CSRF 防护替代。

## 6. 核心接口与目录结构落地方式

### 6.1 目录结构

```text
internal/
  platform/
    security/
      principal.go
      context.go
      password.go
      token.go
      random.go
      permission.go

  modules/
    auth/
      handler.go
      service.go
      repository.go
      dto.go
      model.go
      token.go
      middleware.go
      permissions.go
      audit.go
      errors.go

    user/
      model.go
      repository.go
      status.go
```

P1 不额外把 auth 目录切成大量薄文件，避免在当前仓库阶段制造过多跳转成本。

### 6.2 `platform/security` 核心抽象

#### `Principal`

```go
type Principal struct {
    PublicID    string
    SessionID   string
    Roles       []string
    Permissions []string
}
```

P1 中的 `Principal` 明确定义为“JWT 可还原身份快照”，只能包含 access token claims 可以无回库还原出的字段。将 `PublicID` 放入 `Principal`，避免 `me` 与后续审计场景重复回库。

P1 中不把 `UserID` 和 `Status` 放进 `Principal`，原因是 `AuthMiddleware` 不回库，当前 JWT claims 也不承载这两个字段。需要内部 `UserID` 的业务流程在 login/refresh 等回库场景中自行读取用户实体，不依赖 middleware principal。

`PublicID` 的职责必须保持清晰：

- `PublicID`：对外稳定身份标识，映射到 JWT `sub` 与 `me.user.id`。

#### context helpers

```go
func WithPrincipal(ctx context.Context, p *Principal) context.Context
func PrincipalFromContext(ctx context.Context) (*Principal, bool)
func MustPrincipal(ctx context.Context) *Principal
func WithSessionID(ctx context.Context, sessionID string) context.Context
func SessionIDFromContext(ctx context.Context) (string, bool)
```

#### `PasswordHasher`

```go
type PasswordHasher interface {
    Hash(password string) (string, error)
    Verify(password string, encodedHash string) (bool, error)
}
```

P1 只实现 `Argon2idHasher`，不在仓库尚无历史密码数据时引入 bcrypt 兼容复杂度。

#### `TokenManager`

```go
type AccessToken struct {
    Token     string
    TokenType string
    ExpiresIn int64
    ExpiresAt time.Time
}

type TokenClaims struct {
    TokenID     string
    Subject     string
    SessionID   string
    Issuer      string
    Audience    string
    IssuedAt    time.Time
    NotBefore   time.Time
    ExpiresAt   time.Time
    Roles       []string
    Permissions []string
}

type TokenManager interface {
    IssueAccessToken(ctx context.Context, p *Principal) (*AccessToken, error)
    VerifyAccessToken(ctx context.Context, raw string) (*Principal, *TokenClaims, error)
}
```

`TokenClaims.Subject` 对应 JWT `sub`，其值固定等于 `Principal.PublicID`。

`VerifyAccessToken` 返回的 `Principal` 必须仅由 `TokenClaims` 无回库重建：

- `Principal.PublicID <- TokenClaims.Subject`
- `Principal.SessionID <- TokenClaims.SessionID`
- `Principal.Roles <- TokenClaims.Roles`
- `Principal.Permissions <- TokenClaims.Permissions`

JWT 时间字段校验需要允许固定时钟偏移容忍度。P1 建议固定为 `30s`，用于 `iat`、`nbf`、`exp` 的比较，避免轻微时钟漂移导致误判。

#### 随机 token 生成

```go
type RandomTokenGenerator interface {
    GenerateURLSafe(n int) (string, error)
}
```

P1 中 `GenerateURLSafe(n)` 的 `n` 表示原始随机字节数，refresh token 至少传入 `32`。

#### 权限 helper

```go
func HasPermission(p *Principal, permission string) bool
func HasAnyPermission(p *Principal, permissions ...string) bool
func HasRole(p *Principal, role string) bool
```

### 6.3 `modules/auth` 核心接口

```go
type Service interface {
    Login(ctx context.Context, req *LoginRequest) (*AuthResponse, *RefreshCookie, error)
    Refresh(ctx context.Context, req *RefreshRequest) (*AuthResponse, *RefreshCookie, error)
    Logout(ctx context.Context, req *LogoutRequest) error
    Me(ctx context.Context) (*MeResponse, error)
}
```

`Login/Refresh` 返回 `RefreshCookie`，让 service 决定 cookie 语义，但不直接操作 `gin.Context`。

### 6.4 `modules/user` 最小边界

P1 中 `user` 只保留：

```go
type Repository interface {
    GetByID(ctx context.Context, userID int64) (*User, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
}
```

用户查询在 P1 中统一收口到 `modules/user.Repository`。登录场景下的“按 identifier 查用户”在当前阶段等价为“按规范化 email 查用户”，不再在 `modules/auth.Repository` 中重复定义第二个用户读取入口。

### 6.5 AuthRepository

```go
type Repository interface {
    ListUserRoles(ctx context.Context, userID int64) ([]string, error)
    ListUserPermissions(ctx context.Context, userID int64) ([]string, error)

    CreateSession(ctx context.Context, session *AuthSession) error
    GetSessionByID(ctx context.Context, sessionID string) (*AuthSession, error)
    RevokeSession(ctx context.Context, sessionID string) error

    CreateRefreshToken(ctx context.Context, token *RefreshToken) error
    GetRefreshTokenByHashForUpdate(ctx context.Context, tokenHash string) (*RefreshToken, error)
    MarkRefreshTokenUsed(ctx context.Context, tokenID int64, replacedByTokenID int64) (bool, error)
    RevokeRefreshTokenFamily(ctx context.Context, familyID string) error
    RevokeRefreshTokensBySessionID(ctx context.Context, sessionID string) error

    UpdateLastLoginAt(ctx context.Context, userID int64, at time.Time) error
    CreateLoginAttempt(ctx context.Context, attempt *LoginAttempt) error
    CreateAuditLog(ctx context.Context, log *AuditLog) error
}
```

事务边界统一复用当前 `platform/database.TxManager`，不在 auth 内部引入第二套事务抽象。

其中 refresh 路径必须满足“旧 token 消费权只能被一个请求抢到一次”的原子性约束，不能依赖先查再改的宽松时序。

`RevokeUserSessions` 不进入 P1 仓储接口。整用户会话撤销属于后续“改密码”“管理员踢下线”“禁用用户联动清会话”等场景的扩展能力，当前 P1 只实现针对 session、token family 的局部撤销。

## 7. `login / refresh / logout / me` 详细流程与错误语义

### 7.1 `POST /api/v1/auth/login`

请求：

```json
{
  "identifier": "admin@example.com",
  "password": "secret"
}
```

P1 只支持邮箱登录，但字段名保留 `identifier` 以兼容未来手机号或用户名扩展。

流程：

1. handler 绑定请求并做基础校验。
2. service 规范化 `identifier`，邮箱统一小写、去空白。
3. handler 或 service 通过统一 IP 提取器拿到 client IP，并在当前流程中复用到 session、login attempt、audit log。
4. 通过 `modules/user.Repository` 按规范化 email 查用户。
5. 若用户不存在，使用 dummy hash 做一次 `Verify`。
6. 写 `login_attempts(success=false)` 与 `audit_logs(login_failed)`，内部 `reason=invalid_credentials`。
7. 返回统一 `AUTH_INVALID_CREDENTIALS`。
8. 若用户存在，校验密码。
9. 若密码错误，写 `login_attempts(success=false)` 与 `audit_logs(login_failed)`，内部 `reason=invalid_credentials`。
10. 返回统一 `AUTH_INVALID_CREDENTIALS`。
11. 校验用户状态，只允许 `active`。
12. 若用户为 `disabled` 或人工 `locked`，写 `login_attempts(success=false)` 与 `audit_logs(login_failed)`，内部 `reason` 分别记录为 `account_disabled` 或 `account_locked`。
13. 读取用户 roles 和 permissions。
14. 开启事务。
15. 创建 `auth_sessions`，同时写入 `expires_at`、`user_agent_hash`、`ip_hash`。
16. 生成 raw refresh token，计算 hash，写入 `refresh_tokens`。
17. 签发 access token。
18. 更新 `users.last_login_at`。
19. 写 `login_attempts(success=true)` 与 `audit_logs(login_success)`。
20. 提交事务。
21. 返回 access token，由 handler 写 refresh cookie。

硬约束：access token 和 refresh cookie 只能在事务成功提交后写出；任何事务失败都必须丢弃本次签发结果，不能向客户端返回 token 或 cookie。

失败语义：

- 用户不存在或密码错误：
  - `401 AUTH_INVALID_CREDENTIALS`
  - 文案固定为“账号或密码错误”
- 用户禁用：
  - `403 AUTH_ACCOUNT_DISABLED`
- 用户锁定：
  - `423 AUTH_ACCOUNT_LOCKED`

P1 中：

- `AUTH_ACCOUNT_LOCKED` 仅用于人工锁定的用户状态。
- 不实现基于失败次数的自动锁定或自动解锁。
- 暴力破解防护的最小执行策略为：沿用现有 `RateLimitMiddleware` 做全局限流，同时记录 `login_attempts` 供审计和后续专用限流扩展。
- 所有登录尝试都必须写入 `login_attempts`；所有认证失败都必须写入 `audit_logs(login_failed)`，即使用户不存在也不例外。

### 7.2 `GET /api/v1/auth/me`

流程：

1. 路由经过 `AuthMiddleware`。
2. 从 context 读取 `Principal`。
3. 返回快照型身份信息：`user.id(public_id)`、`roles`、`permissions`、`session_id`。

P1 不为 `me` 额外回库，直接使用 principal 快照。

这意味着 `me` 在 P1 中不是“当前真实用户资料”接口，而是“当前 access token 身份快照”接口；若用户邮箱、显示名、状态在签发后发生变化，要等重新 refresh 或重新 login 后才会反映到返回值中。

`password_changed_at` 在 P1 中不参与 `me` 返回值的实时判定。

### 7.3 `POST /api/v1/auth/refresh`

P1 采用 Cookie 模式，默认从 cookie 中读取 refresh token，而不是 body。

流程：

1. handler 读取 refresh cookie。
2. 若不存在，返回 `401 AUTH_REFRESH_INVALID`。
3. service 对 raw refresh token 做 hash。
4. 开启事务。
5. repository 以行锁语义读取旧 `refresh_tokens` 记录，例如 `SELECT ... FOR UPDATE`。
6. 若不存在，返回 `401 AUTH_REFRESH_INVALID`。
7. 校验 `expires_at`、`revoked_at`。
8. 若 `used_at` 已存在，判定 reuse attack。
9. reuse 时在同一事务内撤销该 `family_id` 的全部 refresh token，并撤销当前 session。
10. 写 `audit_logs(refresh_reuse_detected)`。
11. handler 清理 refresh cookie。
12. 返回 `401 AUTH_REFRESH_REUSED`。
13. 正常路径下，查询 session 与 user 状态。
14. 若 session 已撤销或 user 非 `active`，handler 清理 refresh cookie 后统一返回 `401 AUTH_REFRESH_INVALID`。
15. 若 session 创建时间或 refresh token family 签发时间早于 `users.password_changed_at`，则拒绝刷新、撤销当前 session / token family、记录 `session_revoked`，并清理 refresh cookie；对外统一返回 `401 AUTH_REFRESH_INVALID`。
16. 读取最新 roles 和 permissions。
17. 生成新的 raw refresh token 与 hash。
18. 以原子更新方式消费旧 token，要求 `MarkRefreshTokenUsed` 仅在 `used_at IS NULL AND revoked_at IS NULL` 时成功。
19. 若未成功消费旧 token，说明发生并发竞争，handler 清理 refresh cookie 后映射为 `401 AUTH_REFRESH_REUSED`。
20. 写入新 token，沿用同一 `family_id`。
21. 签发新 access token。
22. 写审计日志 `refresh_success`。
23. 提交事务。
24. 返回新的 access token，由 handler 覆盖 refresh cookie。

硬约束：refresh 生成出的新 access token 和新 refresh cookie 只能在事务成功提交后写出；任何事务失败都必须丢弃本次签发结果与 cookie 覆盖动作。

P1 在 refresh 时始终以数据库当前权限重新签发 access token，而不是沿用旧权限快照。

这意味着：并发的两个 refresh 请求中，只有成功抢到旧 token 消费权的那一个可以继续签发新 token；另一个请求必须失败，不能出现双成功刷新。

### 7.4 `POST /api/v1/auth/logout`

P1 将 `logout` 设计为 cookie 驱动接口：不强制要求 Bearer access token，允许仅凭 refresh cookie 完成服务端撤销。若 bearer token 同时存在，则用于补充 session 上下文和审计信息。

流程：

1. handler 读取 refresh cookie，并尝试从 Bearer token 中解析 principal；Bearer 缺失或过期不阻断 logout 主流程。
2. 若存在 refresh cookie，则 hash 后找到对应 token 所属 session。
3. 若 refresh cookie session 与 principal session 都存在且不一致，则同时撤销两个 session，并在审计日志中记录 `session_mismatch=true`。
4. 若只存在 refresh cookie session，则撤销该 session / token family。
5. 若只有 principal 而没有 refresh cookie，至少撤销当前 session。
6. 若 logout 实际触发了 session 或 token family 撤销，同时写 `session_revoked` 事件；`logout_success` 用于表达接口语义成功。
7. 即使 refresh cookie 无效、查不到、已过期或已撤销，也统一清理浏览器端 refresh cookie。
8. 写 `audit_logs(logout_success)`。
9. handler 清除 refresh cookie。
10. 返回成功。

若 logout 过程中发生事务失败或撤销异常，则不得向客户端返回“已成功退出”的响应。

P1 不引入 access token denylist。`logout` 的语义为：

- refresh 路径立即失效。
- session 立即撤销。
- 已签发 access token 通过短 TTL 自然过期。

对于需要“禁用账号或撤销 session 后立即拒绝所有 access token”的场景，当前 P1 设计不满足，需要在后续版本引入 denylist、session version 或每请求回库校验。

refresh / logout 相关 cookie 处理规则：

- `AUTH_REFRESH_INVALID`：清理 refresh cookie。
- `AUTH_REFRESH_REUSED`：清理 refresh cookie。
- `logout` 成功：清理 refresh cookie。
- `logout` 虽未命中有效 session 但收到 cookie：仍清理 refresh cookie。
- 清 cookie 时必须与设置 cookie 时使用相同的 `name`、`domain`、`path`、`sameSite`、`secure`。

### 7.5 错误语义

中间件和 handler 对外统一使用以下错误：

- `AUTH_INVALID_CREDENTIALS`
- `AUTH_TOKEN_MISSING`
- `AUTH_TOKEN_INVALID`
- `AUTH_TOKEN_EXPIRED`
- `AUTH_REFRESH_INVALID`
- `AUTH_REFRESH_REUSED`
- `AUTH_UNAUTHORIZED`
- `AUTH_FORBIDDEN`
- `AUTH_ACCOUNT_DISABLED`
- `AUTH_ACCOUNT_LOCKED`

要求：

- 外部 message 必须安全、稳定。
- 不区分“账号不存在”和“密码错误”。
- 权限不足不暴露资源是否存在。

## 8. 中间件、路由装配、OpenAPI 与 Worker/System Principal 接入

### 8.1 路由装配

Auth 模块完全复用现有 `bootstrap.NewAPIEngine(...registrars)` 链路，不额外起特殊 server。

`modules/auth/handler.go` 挂载：

- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- `GET /api/v1/auth/me`

其中：

- `login`、`refresh` 为公开接口。
- `logout` 为 cookie 驱动接口，不通过 `AuthMiddleware` 强制保护。
- `me` 通过路由组挂 `AuthMiddleware`。

### 8.2 中间件策略

保留现有全局中间件顺序：

- `RequestIDMiddleware`
- `TraceContextMiddleware`
- `CORSMiddleware`
- `SecurityHeadersMiddleware`
- `RateLimitMiddleware`
- `RecoveryMiddleware`

新增策略：

- `AuthMiddleware` 不作为全局中间件，只按路由组挂载。
- `RequirePermission` 按具体业务路由挂载。

### 8.3 `AuthMiddleware`

行为：

1. 只从 `Authorization: Bearer` 读取 access token。
2. 不从 refresh cookie 中推导 principal。
3. 验证签名、`alg`、`iss`、`aud`、`nbf`、`iat`、`exp`。
4. 解析 `Principal`。
5. 写入 `context.Context`。
6. 保留现有 `request_id` / `trace_id`，不覆盖。

### 8.4 `RequirePermission`

中间件职责：

- 无 principal -> `401 AUTH_UNAUTHORIZED`
- 有 principal 但权限不足 -> `403 AUTH_FORBIDDEN`

使用方式示例：

```go
protected := group.Group("/orders")
protected.Use(auth.AuthMiddleware(tokenManager))
protected.GET(":id", auth.RequirePermission("order:read"), h.GetOrder)
```

中间件负责函数级权限，service 负责对象级权限。

### 8.5 OpenAPI

更新现有 `api/openapi.yaml`：

- 新增 `components.securitySchemes.bearerAuth`
- 新增 `components.securitySchemes.refreshCookie`
- 新增 `LoginRequest`、`AuthUser`、`AuthTokens`、`AuthResponseData`、`MeResponseData`
- `POST /api/v1/auth/refresh` 明确声明依赖 `refreshCookie`
- `POST /api/v1/auth/logout` 明确声明依赖 `refreshCookie`
- 在 `GET /api/v1/auth/me` 上声明 `bearerAuth`

OpenAPI 还需要明确这些契约：

- `POST /api/v1/auth/login` 与 `POST /api/v1/auth/refresh` 会返回 `Set-Cookie` 写入 refresh token。
- `POST /api/v1/auth/logout` 会返回清理 refresh token 的 `Set-Cookie`。
- `refreshCookie` 的名字固定为 `refresh_token`，并在接口描述中说明 `credentials: include`。
- cookie 的 `Path`、`SameSite`、`Secure` 语义在接口描述或安全说明中体现，避免前端只能依赖口头约定。

接口级错误码映射也需要在 OpenAPI 契约中明确：

- `POST /api/v1/auth/login`
  - `401 AUTH_INVALID_CREDENTIALS`
  - `403 AUTH_ACCOUNT_DISABLED`
  - `423 AUTH_ACCOUNT_LOCKED`
- `POST /api/v1/auth/refresh`
  - `401 AUTH_REFRESH_INVALID`
  - `401 AUTH_REFRESH_REUSED`
- `POST /api/v1/auth/logout`
  - 成功通常返回 `200`
  - 若撤销事务失败，则返回统一服务端错误
- `GET /api/v1/auth/me`
  - `401 AUTH_TOKEN_MISSING`
  - `401 AUTH_TOKEN_INVALID`
  - `401 AUTH_TOKEN_EXPIRED`

Auth 响应继续复用当前统一 `APIResponse` 外壳。

### 8.6 `admin/` 前端协同契约

- `login` 成功：
  - body 返回 `access_token`、`token_type`、`expires_in`、`user`
  - response set-cookie 写 refresh token
- `refresh` 成功：
  - body 返回新的 `access_token`、`token_type`、`expires_in`、`user`
  - response set-cookie 覆盖 refresh token
- `logout` 成功：
  - 清理 refresh cookie

### 8.7 Worker / System Principal

虽然 P1 不实现 worker 登录，但需要预留 system principal 能力。建议在 `platform/security` 中提供：

```go
func NewSystemPrincipal(sessionID string, permissions ...string) *Principal
```

后续 worker 可注入：

- `Roles: []string{"system"}`
- `Permissions: [...]`
- `SessionID: "worker:<task>"`

避免内部任务绕过 service 权限与状态机规则。

## 9. 测试策略、迁移/seed、实施顺序与风险控制

### 9.1 测试策略

#### `platform/security` 单元测试

必须覆盖：

- `Principal` context helpers
- `HasPermission / HasRole`
- `Argon2idHasher.Hash / Verify`
- `TokenManager.IssueAccessToken / VerifyAccessToken`
- 错误 `alg` / 错误 `iss` / 错误 `aud` / 过期 token / 非法 token

#### `modules/auth` service 单元测试

至少覆盖：

- `Login` 成功
- `Login` 账号不存在
- `Login` 密码错误
- `Login` 用户禁用
- `Login` 人工锁定用户
- `Refresh` 成功
- `Refresh` token 不存在
- `Refresh` token 过期
- `Refresh` token reused
- `Refresh` 并发竞争下仅一个请求成功
- `Refresh` family revoke 后最新 token 也失效
- `Logout` 与 `Refresh` 并发时结果可预期
- `Logout` 成功
- `Logout` 时 principal session 与 cookie session 不一致
- `Me` 返回 principal 快照

#### handler / middleware 测试

至少覆盖：

- `POST /auth/login` 参数错误
- `POST /auth/login` 登录成功并写 cookie
- `POST /auth/login` 登录失败
- `POST /auth/refresh` 无 cookie
- `POST /auth/refresh` 成功轮换 cookie
- `POST /auth/refresh` 非受信 Origin 被拒绝
- `POST /auth/logout` 未登录
- `POST /auth/logout` principal session 与 cookie session 不一致
- `GET /auth/me` 未登录
- `GET /auth/me` 已登录
- `AuthMiddleware` 缺 token / 非法 / 过期 / 有效写 context
- `RequirePermission` 无 principal / 权限不足 / 放行
- `request_id / trace_id` 在成功与失败响应中保持可见

#### repository 集成测试

建议覆盖：

- 用户唯一约束
- 创建 session
- 创建 refresh token
- 按 hash 加锁查询 refresh token
- 标记 token 已使用
- family revoke
- 并发 refresh 仅一个事务成功消费旧 token
- `logout` 与 `refresh` 并发时 session / token 状态一致
- 查询用户 roles/permissions
- 审计日志写入

### 9.2 迁移与 seed 策略

- `migrations/` 仅承载 schema 变更。
- 新增 auth 初始迁移，创建 P1 所需全部表。
- 基础角色与权限可随 migration 初始化。
- 默认管理员账号不写死在 migration 中。
- `local/dev` 仅可通过显式 bootstrap/init 命令创建管理员，不在应用每次启动时自动检查创建。
- `prod` 不自动创建默认管理员。

bootstrap admin 的幂等语义需要固定：

- 仅在 `local/dev` 环境允许执行。
- 仅在显式 bootstrap/init 命令下执行。
- 若 email 已存在，则跳过，不重置密码，不覆盖角色。
- bootstrap 行为必须写结构化启动日志，便于排查。

### 9.3 实施顺序

建议实施顺序：

1. 扩展 `config`，加入 `AuthConfig` 与密钥校验。
2. 落 `platform/security`。
3. 编写 migration 与最小 seed 方案。
4. 落 `modules/user` 最小模型与查询仓储。
5. 落 `modules/auth` model / repository / service。
6. 落 `AuthMiddleware` 与 `RequirePermission`。
7. 接入 `cmd/api/main.go` 与 `bootstrap.NewAPIEngine(...)`。
8. 更新 `api/openapi.yaml`。
9. 补测试。
10. 运行 `go test ./...`。
11. 运行 `golangci-lint run`。

### 9.4 风险控制

P1 的控制原则：

- 不改动现有四个业务模块的既有行为，只新增 auth 能力与可复用中间件。
- 不把 Auth 作为全局 mandatory middleware。
- 不在 P1 引入 Redis 或 DB access token denylist。
- 明确接受 `disabled`、`locked`、`session revoked` 对已签发 access token 的延迟生效窗口，最长为 `access_token_ttl`。
- 不把注册、找回密码、邮件/短信链路一起拉进来。
- 不信任前端传来的 role / permission / user_id。
- 不把 CORS 白名单误当作 CSRF 防护，cookie 端点仍需校验 `Origin/Referer`。

## 10. 验收标准

P1 完成时至少满足：

- `platform/security` 落地。
- `modules/auth` 落地。
- 最小 `modules/user` 落地。
- 登录接口可用。
- access token 可签发和验证。
- refresh token 可哈希存储、刷新、轮换、撤销。
- `AuthMiddleware` 可保护接口。
- `RequirePermission` 可做权限校验。
- `Principal` 可从 context 读取。
- 登录失败不暴露账号是否存在。
- 密码使用 Argon2id 存储。
- 日志不包含密码或 token 原文。
- `request_id` / `trace_id` 贯穿 auth handler / service / log / audit。
- OpenAPI 声明 `bearerAuth`。
- `go test ./...` 通过。
- `golangci-lint run` 通过。
