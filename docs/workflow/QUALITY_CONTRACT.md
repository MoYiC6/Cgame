# QUALITY CONTRACT

本仓库的下载包风格治理配置以以下文件为准：

- `.agent/project.json`：运行命令、service matrix、verification profile
- `.scale/workflow.json`：工作流阶段和门控状态机
- `.scale/quality-contract.json`：任务等级、交付证据和红线
- `.scale/skills-registry.json`：技能发现、安装安全策略和推荐动作

## 项目形态

- 仓库根目录：治理层、文档、脚本
- `backend/`：Go 服务，默认验证入口
- `admin/`：Vue/Vite/TypeScript 管理端，前端独立验证入口

## 默认验证命令

```bash
bash scripts/workflow/verify.sh default
```

该命令会对 `backend/` 执行：

- `go vet ./...`
- `go test ./...`
- `go build ./...`

## 前端验证命令

```bash
bash scripts/workflow/verify.sh frontend
```

该命令会对 `admin/` 执行：

- `pnpm lint`
- `pnpm build`
- `pnpm exec vue-tsc --noEmit`

## 真实交付规则

1. 未实际运行验证，不得声称通过。
2. 默认后端 profile 通过，不代表前端 profile 自动通过。
3. `scripts/gates/all.sh --dry-run` 只验证门控脚本存在和可调度。
4. 跨前后端任务使用 `fullstack` 或 `release` profile，并在 summary 中记录实际命令结果。
