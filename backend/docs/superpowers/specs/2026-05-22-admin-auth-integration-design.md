# Admin 与 Backend Auth 真实对接设计

> 设计日期：2026-05-22
> 范围：将 `admin/` 前端管理端的认证链路从当前 mock/旧契约切换到 `backend/` 的真实 Auth API，仅覆盖登录态，不包含后端菜单权限模式切换。

## 1. 目标与范围

本轮目标是把前端真正接到后端的 `login / refresh / logout / me`，让浏览器端可以用真实账号完成登录、续期和退出。

本轮完成后应满足：

- 登录页调用后端真实登录接口。
- 浏览器通过 `HttpOnly refresh cookie` 维持会话。
- 前端基于后端 `me` 接口恢复当前身份快照。
- access token 过期时，前端会先尝试静默 refresh，再决定是否退出登录。
- 退出登录时，同时清除前端状态与后端会话。
- 前端保留当前静态菜单/前端权限模式，不依赖后端菜单接口。

本轮明确不做：

- 后端动态菜单接口。
- 将前端权限模式切换到 `backend`。
- 基于后端权限码改写整套路由注册逻辑。
- 完整用户资料接口（昵称、头像、邮箱展示等扩展字段）。

## 2. 当前差距

当前前端和后端的 Auth 契约不一致，主要体现在：

- 前端 `admin/src/api/auth.ts` 仍调用旧接口：
  - `POST /api/auth/login`
  - `GET /api/user/info`
- 后端真实接口是：
  - `POST /api/v1/auth/login`
  - `POST /api/v1/auth/refresh`
  - `POST /api/v1/auth/logout`
  - `GET /api/v1/auth/me`
- 前端 `user store` 仍保存 `refreshToken`，与后端当前的 cookie 模式冲突。
- 前端路由守卫将“用户信息获取”和“菜单初始化”耦合在一起，而后端本轮只提供身份快照，不提供菜单。
- 后端当前虽已接入 auth handler，但 `user/auth repository` 仍是 noop，实现上还不能真正查询数据库。

因此，真实对接必须同时修改前端请求链和后端 repository 持久化实现。

## 3. 推荐方案

本轮采用“兼容式真实对接”方案：

- 前端保留当前静态菜单与 `frontend` 权限模式。
- 只替换认证相关 API 与会话管理逻辑。
- 后端先补齐真实 `user/auth repository`，确保 `login / me / refresh / logout` 可访问数据库。
- 前端停止保存 `refreshToken`，仅通过 cookie 管理 refresh token。
- access token 仍可暂时保存在前端 store 和本地持久化中，以减少改动面。

这样做的原因：

- 与当前后端能力匹配，不强行引入菜单接口。
- 改动集中在认证边界，易于验证。
- 后续若要切换到后端菜单模式，只需继续扩展 `menu` 和 `permission` 契约，不需要再推翻本轮登录态实现。

## 4. 前端对接设计

### 4.1 API 契约切换

前端统一切到以下真实接口：

- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- `GET /api/v1/auth/me`

字段映射规则：

- 登录请求：
  - 前端 `userName` 输入映射到后端 `identifier`
  - `password` 保持不变
- 登录响应：
  - 后端 `data.access_token` -> 前端 `accessToken`
  - 后端不再返回 `refresh_token`
- `me` 响应：
  - 后端返回的是身份快照，不是完整用户资料
  - 前端只可靠使用：`id`、`roles`、`permissions`、`session_id`

### 4.2 user store 调整

`admin/src/store/modules/user.ts` 需要做这些收敛：

- 删除 `refreshToken` 状态及相关读写逻辑。
- `setToken()` 只接收 `accessToken`。
- `logOut()` 不再尝试处理本地 refresh token。
- `setUserInfo()` 接受新的后端快照结构，并在前端补齐当前页面依赖的兼容字段。

由于现有界面还可能读取 `userName`、`userId`、`avatar` 等字段，本轮采用兼容映射：

- `userId`：先映射为后端 `user.id` 字符串，前端类型改为兼容字符串。
- `userName`：若后端未返回，则默认用 `user.id` 或空字符串占位。
- `avatar`：本轮不从后端获取，前端允许为空或继续使用本地默认图。

本轮不要求把整个前端用户类型收缩到最小，只要求不再假定后端一定返回旧 mock 字段。

### 4.3 HTTP 拦截器与静默刷新

`admin/src/utils/http/index.ts` 需要改成真实会话模式：

- 请求时继续自动带 `Authorization: Bearer <access_token>`。
- `withCredentials` 在开发和生产都按配置启用，以便浏览器携带 refresh cookie。
- 当响应为 `401` 时：
  1. 若当前请求不是 `refresh`/`logout`，先触发一次 `POST /api/v1/auth/refresh`
  2. refresh 成功后更新 `accessToken`
  3. 重放原始请求一次
  4. 若 refresh 失败，再执行 `logOut()`

约束：

- 必须防止多个并发 401 同时触发多次 refresh。
- 必须避免 refresh 请求本身再次进入 refresh 死循环。
- refresh 失败后，前端需要接受后端清 cookie 的结果并清空本地状态。

### 4.4 路由守卫与菜单初始化

本轮不切后端菜单模式，所以守卫逻辑需要解耦：

- `fetchUserInfo()` 改为调用后端 `me`。
- `MenuProcessor` 仍按 `frontend` 模式从 `asyncRoutes` 生成菜单。
- 路由初始化不再依赖后端菜单接口是否存在。

这意味着：

- 登录成功后，用户身份来自后端。
- 菜单仍然来自前端静态定义。
- 角色字段可以继续用于前端静态菜单过滤，但以后端 `roles` 快照为准。

### 4.5 环境配置

`admin/.env` 与 `admin/.env.development` 需要做这些调整：

- `VITE_WITH_CREDENTIALS=true`
- `VITE_API_URL=/`
- `VITE_API_PROXY_URL=http://127.0.0.1:<backend-port>`
- `VITE_ACCESS_MODE=frontend` 继续保持不变

开发态建议通过 Vite 代理转发到 Go backend，避免直接跨域联调带来 cookie 问题。

## 5. 后端对接设计

### 5.1 必须完成的真实能力

后端要让前端真实登录，本轮至少要补齐：

- `user.Repository` 真实查询：
  - `GetByEmail`
  - `GetByID`
- `auth.Repository` 真实读写：
  - roles / permissions 查询
  - session 创建与撤销
  - refresh token 创建 / 查询 / 使用 / family revoke
  - login_attempts / audit_logs 写入
  - `last_login_at` 更新

如果 repository 仍是 noop，前端虽然能命中路由，但真实登录不会成立。

### 5.2 Cookie 与 CORS 约束

本轮浏览器联调必须满足：

- 后端 CORS 不可使用任意 origin + credentials。
- 允许 `admin` 开发域名或端口的 origin。
- `Allow-Credentials=true`。
- refresh cookie 的 `path` 至少覆盖 `/api/v1/auth`。
- 本地开发环境下 `cookie.secure=false`，否则浏览器不会在 http 下写 cookie。

### 5.3 me 返回值现实约束

当前后端 `me` 是身份快照接口，不是完整用户资料接口。

因此本轮前端对接必须接受：

- `me` 只能稳定提供当前登录主体的最小信息。
- 不应要求它同时承担用户中心资料接口职责。

如果前端确实需要更多展示字段，应作为后续 `user profile` 能力单独补接口，而不是在本轮把 `me` 扩展成大而全接口。

## 6. 错误语义与交互要求

前后端联调时，对这些错误语义要保持一致：

- 登录失败：`AUTH_INVALID_CREDENTIALS`
- refresh 无效：`AUTH_REFRESH_INVALID`
- refresh 重放：`AUTH_REFRESH_REUSED`
- access token 缺失：`AUTH_TOKEN_MISSING`
- access token 无效：`AUTH_TOKEN_INVALID`
- access token 过期：`AUTH_TOKEN_EXPIRED`

前端行为要求：

- 登录页显示后端安全 message。
- `401 + token expired/invalid` 优先走静默 refresh。
- `refresh invalid/reused` 直接退出登录并跳回登录页。
- logout 永远清空前端 store，即使后端返回的是“无效 cookie 已清理”。

## 7. 测试与验收

### 7.1 前端验收

至少验证：

- 登录页可使用真实账号登录。
- 刷新浏览器页面后仍保持登录态。
- 人工让 access token 失效后，页面请求能自动 refresh。
- logout 后刷新页面应回到未登录状态。
- 前端本地状态中不再保存 `refreshToken`。

### 7.2 后端验收

至少验证：

- `POST /api/v1/auth/login` 能查询真实用户并写 session / refresh token。
- `GET /api/v1/auth/me` 能在真实 token 下返回身份快照。
- `POST /api/v1/auth/refresh` 能在 cookie 存在时完成续期。
- `POST /api/v1/auth/logout` 能撤销当前 refresh session 并清 cookie。

### 7.3 联调验收

最终可接受标准：

- `admin/` 不再依赖 mock auth 接口。
- `admin/` 登录、刷新、退出全部通过 `backend/` 真实 Auth API 完成。
- 菜单继续走前端模式，但不阻塞真实登录态。

## 8. 第一批实施顺序

推荐实现顺序：

1. 后端补真实 `user/auth repository`
2. 后端补 repository 级测试与必要 integration
3. 前端改 `src/api/auth.ts` 与类型定义
4. 前端改 `user store`，移除 `refreshToken`
5. 前端改 `utils/http`，实现静默 refresh 和请求重放
6. 前端改路由守卫，将 `me` 与菜单初始化解耦
7. 调整 `.env.development` 与代理
8. 启动前后端联调验证登录、刷新、退出

这套顺序的核心是先让后端具备真实能力，再让前端接真实契约；否则前端改完也只能对着 noop repository 空转。
