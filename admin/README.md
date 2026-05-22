# Cgame

基于 Art Design Pro 精简版整理的前端工程，作为本仓库的独立管理端目录存在，不放入 `backend/`。

## 技术栈

- Vue 3
- Vite
- TypeScript
- Pinia
- Vue Router
- Element Plus

## 目录边界

- `src/views/auth`：登录、注册、找回密码
- `src/views/dashboard`：控制台首页
- `src/views/system`：用户、角色、菜单、个人中心
- `src/views/result`：结果页
- `src/views/exception`：异常页
- `src/api`：前端接口封装
- `src/utils/http`：Axios 请求封装

## 本地运行

```bash
pnpm install
pnpm dev
```

默认开发端口来自 `.env` 的 `VITE_PORT`，当前为 `3006`。

## 构建

```bash
pnpm build
```

## 后端接口

当前 `.env.development` 仍使用 Art Design Pro 的 Apifox Mock，方便前端独立启动。接入本仓库 Go 后端时，将 `VITE_API_PROXY_URL` 改为本地 API 地址，例如：

```env
VITE_API_PROXY_URL = http://localhost:8080
```
