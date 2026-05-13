# PROJECT KNOWLEDGE BASE

**Generated:** 2026-05-13 Asia/Shanghai
**Commit:** cdd0e0e
**Branch:** main

## OVERVIEW
仓库根目录只是外壳。真正的 Go 后端工程位于 `backend/`，日常开发、测试、启动、架构约束都以该目录为准。

## STRUCTURE
```text
./
├── backend/   # 实际项目根；Go module、命令入口、内部模块、配置、测试都在这里
└── .git/      # Git 元数据
```

_上图只保留关键结构；`.idea/`、`.gitignore` 等编辑器/元数据文件省略。_

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| 理解项目全貌 | `backend/AGENTS.md` | 先读这份，再进入具体子层级 |
| 启动 API / Worker | `backend/cmd/` | 入口只做装配，不写业务 |
| 查业务模块约束 | `backend/internal/modules/AGENTS.md` | 四个业务域共享同一骨架 |
| 查基础设施规范 | `backend/internal/platform/AGENTS.md` | config/logger/errors/response/observability/database |
| 查启动装配规范 | `backend/internal/bootstrap/AGENTS.md` | Gin、middleware、shutdown、worker |

## CONVENTIONS
- 根目录不是 Go module 根；不要在仓库根执行 Go 改动时假设 `go.mod` 在这里。
- 默认工作目录应切到 `backend/` 再执行 `go test ./...`、`make test`、`make lint`。

## ANTI-PATTERNS (THIS PROJECT)
- 不要把仓库根当成业务代码入口。
- 不要在根目录新增与 `backend/` 平行的第二套应用结构，除非先更新分层知识库。

## COMMANDS
```bash
cd backend && make run-api
cd backend && make run-worker
cd backend && make test
cd backend && make lint
```

## NOTES
- 当前仓库没有前端/Node 工程痕迹；不要为 `package.json`、TS 配置或前端构建脚本浪费时间。
