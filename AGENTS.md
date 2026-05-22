# PROJECT KNOWLEDGE BASE

**Generated:** 2026-05-13 Asia/Shanghai
**Commit:** cdd0e0e
**Branch:** main

## OVERVIEW
仓库根目录承载多个工程目录。Go 后端工程位于 `backend/`；前端管理端位于 `admin/`。进入具体工程前先切到对应目录。

## STRUCTURE
```text
./
├── backend/   # 实际项目根；Go module、命令入口、内部模块、配置、测试都在这里
├── admin/     # 前端管理端；Vue/Vite/TypeScript/Element Plus
└── .git/      # Git 元数据
```

_上图只保留关键结构；`.idea/`、`.gitignore` 等编辑器/元数据文件省略。_

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| 理解项目全貌 | `backend/AGENTS.md` | 先读这份，再进入具体子层级 |
| 启动前端管理端 | `admin/` | 独立 Node/Vite 工程，执行 `pnpm dev` |
| 启动 API / Worker | `backend/cmd/` | 入口只做装配，不写业务 |
| 查业务模块约束 | `backend/internal/modules/AGENTS.md` | 四个业务域共享同一骨架 |
| 查基础设施规范 | `backend/internal/platform/AGENTS.md` | config/logger/errors/response/observability/database |
| 查启动装配规范 | `backend/internal/bootstrap/AGENTS.md` | Gin、middleware、shutdown、worker |

## CONVENTIONS
- 根目录不是 Go module 根；不要在仓库根执行 Go 改动时假设 `go.mod` 在这里。
- 默认工作目录应切到 `backend/` 再执行 `go test ./...`、`make test`、`make lint`。
- 前端工作目录应切到 `admin/` 再执行 `pnpm install`、`pnpm dev`、`pnpm build`。

## ANTI-PATTERNS (THIS PROJECT)
- 不要把仓库根当成业务代码入口。
- 不要把 `admin/` 前端代码放进 `backend/`；两个工程保持独立目录。

## COMMANDS
```bash
cd backend && make run-api
cd backend && make run-worker
cd backend && make test
cd backend && make lint
cd admin && pnpm install
cd admin && pnpm dev
cd admin && pnpm build
```

## NOTES
- `admin/` 是基于 Art Design Pro 精简版整理的前端工程，默认仍使用 Mock 接口，后续接入 Go 后端时优先调整 `admin/.env.development`。
