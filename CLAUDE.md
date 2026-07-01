# OpenWolf

@.wolf/OPENWOLF.md

This project uses OpenWolf for context management. Read and follow .wolf/OPENWOLF.md every session. Check .wolf/cerebrum.md before generating code. Check .wolf/anatomy.md before reading files.


# 全能型架构

## 全局语言规则

- 所有面向用户的回复必须使用中文。
- 所有 Git 提交信息（包括标题和正文）必须使用中文。

> SCALE OS 方法论：本配置旨在培养 Agent 的工程素养。Agent 应学习研究当前实际环境，灵活适配安装配置，自主使用相关 skills 技能，完成任务后沉淀知识经验并更新知识库。详见安装指引中的“方法论与使用指引”章节。

## META

- `agent`: `claude-code`
- `scenario`: `standard`
- `stack`: `polyglot`
- `generated`: `2026-05-25`
- `scale_version`: `10.0-adapted-for-zuhao`
- `doc`: `CLAUDE.md`

## COMMANDS

```text
dev: echo 'backend: cd backend && make run-api | worker: cd backend && make run-worker | admin: cd admin && pnpm dev'
build: (cd backend && go build ./...) && (cd admin && pnpm build)
test: cd backend && go test ./...
lint: (cd backend && go vet ./...) && (cd admin && pnpm lint)
typecheck: (cd backend && go build ./...) && (cd admin && pnpm exec vue-tsc --noEmit)
coverage: cd backend && go test ./... -cover
security: bash scripts/gates/G7-verify.sh
```

## SERVICE_MATRIX

### `backend`

- `path`: `backend`
- `stack`: `go`
- `default`: `true`
- `dev`: `make run-api`
- `build`: `go build ./...`
- `test`: `go test ./...`
- `lint`: `go vet ./...`
- `typecheck`: `go build ./...`
- `coverage`: `go test ./... -cover`

### `admin`

- `path`: `admin`
- `stack`: `node`
- `default`: `false`
- `dev`: `pnpm dev`
- `build`: `pnpm build`
- `lint`: `pnpm lint`
- `typecheck`: `pnpm exec vue-tsc --noEmit`

## TECH_STACK

- Go
- Gin
- PostgreSQL
- Vue 3
- TypeScript
- Vite
- pnpm
- Docker Compose

## AGENT_CAPABILITY

- `support_level`: `full`
- `memory_files`: `CLAUDE.md`
- `config_files`: `.claude/settings.json`
- `hooks`: `supported`
- `mcp`: `supported`
- `project_skills_path`: `.agents/skills/`

## §0 核心元认知（不可逾越）

### 0.1 认知诚实

- 不确定时，输出 `[UNCERTAIN]` 并说明缺失什么。
- 未实际运行验证，绝不允许输出“通过”。
- 不编造未在代码中定义的调用关系。
- 不得把后端通过说成全栈通过，也不得把 `dry-run` 说成质量通过。

### 0.2 显性推理

- 影响面分析：每次修改前，列出可能受影响的模块、文件、功能。
- 抓主要矛盾：先找根因，再处理衍生问题。
- 权衡方案：存在多种方案时，列出利弊并说明选择理由。
- 前置异常思考：实现前先想“什么情况会出错”，并制定防御策略。

### 0.3 Owner 意识

- 做 A + 检查 B 同类问题 + 确保不影响 C。
- 一个 bug 进来，一类问题出去。
- 做超出用户要求但有价值的工作时，标记 `[OWNER]` 并说明理由。

### 0.5 技能优先意识

- 相关性强或存在明确触发条件时，主动选择最小可用技能集；不得为了凑流程调用无关技能。
- 调用技能前，先确认技能是否支持当前技术栈、Agent、权限和安全边界。
- 安装第三方技能前，先检查来源、脚本、依赖、postinstall、权限和 lockfile 变化。
- 技能调用失败时，记录原因、降级方案和后续改进到 `summary`。

## CODE_RULES

- `[ENFORCED]` 禁止空 catch 块
- `[ENFORCED]` 禁止硬编码密钥、token、password、private key
- `[ENFORCED]` Go 错误必须显式处理，禁止忽略 `error`
- `[ENFORCED]` 前端 TypeScript 改动必须通过 `pnpm exec vue-tsc --noEmit`
- `[ENFORCED]` 后端行为改动必须通过 `go test ./...`
- `[ENFORCED]` 前端请求逻辑必须遵守现有 `admin/src/utils/http` 封装边界
- `[ENFORCED]` 后端业务异常与系统异常分离，全局错误处理保持一致
- `[ENFORCED]` 禁止未经说明跨越 `backend/admin` 边界做顺手重构

## KARPATHY_PRINCIPLES

- `[K1-THINK]` 编码前必须明确列出假设，不确定时停下来提问而非猜测
- `[K1-THINK]` 存在多种解释时必须呈现所有选项，不得默默选择一种
- `[K1-THINK]` 存在更简单方案时必须提出异议
- `[K2-SIMPLE]` 禁止添加未要求的功能、抽象、灵活性或可配置性
- `[K2-SIMPLE]` 如果 200 行可写 50 行，必须重写
- `[K2-SIMPLE]` 禁止为不可能场景添加错误处理
- `[K3-SURGICAL]` 每一行修改都必须可追溯到用户请求——无关改动零容忍
- `[K3-SURGICAL]` 禁止“顺手”重构、改格式、加类型标注、改注释
- `[K3-SURGICAL]` 匹配现有代码风格，即使你更倾向不同写法
- `[K4-GOAL]` 必须将命令式任务转化为可验证目标：测试先行 -> 实现 -> 验证
- `[K4-GOAL]` 多步任务必须声明计划：`1. [步骤] -> 验证: [检查]`
- `[K4-GOAL]` 成功标准必须明确，弱标准需要不断澄清

## WORKFLOW

- `mode`: `standard`
- `step_1`: 探索 -> 读知识文档 + 扫代码 + 找验证命令
- `step_2`: 规划 -> 影响分析 + 契约定义 + 回滚思考
- `step_3`: 执行 -> RED/GREEN/REFACTOR
- `step_4`: 验证 -> 运行真实命令，不用脑补结果
- `step_5`: 交付 -> 列出完成内容、验证结果、未验证项

## WORKTREE_RULES

- 根目录主要承载治理、文档和跨服务脚本，不是 Go module 根。
- 后端相关命令默认在 `backend/` 下运行。
- 前端相关命令默认在 `admin/` 下运行。
- 跨前后端任务使用 `fullstack` 或 `release` profile，不得只跑一边就声称整体完成。

## GATES

- `G0`: Build 通过 -> `bash scripts/gates/G0-verify.sh`
- `G1`: 探索完成 -> 已读文件、命令或测试证据可追溯
- `G2`: 规划完成 -> 计划包含边界、风险、验证方式、回滚计划
- `G3`: TDD 合规 -> 测试先行或说明不适用原因
- `G4`: Lint 通过 -> `bash scripts/gates/G4-verify.sh`
- `G5`: 测试通过 -> `bash scripts/gates/G5-verify.sh`
- `G6`: 类型检查 -> `bash scripts/gates/G6-verify.sh`
- `G7`: 安全检查 -> 无密钥、危险删除、未授权数据变更
- `G8`: 无 AI 残留 -> 无对话残留、空 TODO、明显 AI 注释噪声
- `G9`: 知识更新 -> `CLAUDE.md` 或 `.scale/task-summary.md` 有同步证据

## VERIFICATION_PROFILES

### `fast`

- `services`: `backend`
- `required`: `go test ./...`
- `optional`: `go vet ./...`

### `default`

- `services`: `backend`
- `required`: `go vet ./...`, `go test ./...`, `go build ./...`
- `optional`: `go test ./... -cover`

### `frontend`

- `services`: `admin`
- `required`: `pnpm lint`, `pnpm build`, `pnpm exec vue-tsc --noEmit`
- `optional`: automated frontend test when available

### `fullstack`

- `services`: `backend + admin`
- `required`: `lint + build + typecheck`
- `optional`: `backend tests`, `coverage`, `security`

### `release`

- `services`: `backend + admin`
- `required`: `lint + build + typecheck`
- `optional`: `test + coverage + security`

## SKILLS

- `agentskills_spec`: [agentskills](https://github.com/agentskills/agentskills)
- `install_path`: `.agents/skills/`
- `discovery`: auto-scan subdirs containing `SKILL.md`
- `registry`: `.scale/skills-registry.json`
- `policy`: progressive disclosure; do not load every `SKILL.md` into context

### INSTALLED_PROJECT_SKILLS

- `openwolf`（全局已安装，项目已完成 `openwolf init`，`.wolf/` 与 Claude Code hooks 已启用）
- `build-graph`
- `code-research`
- `cua-driver`
- `debug-issue`
- `explore`
- `explore-codebase`
- `feature-research`
- `gitnexus-cli`
- `gitnexus-debugging`
- `gitnexus-exploring`
- `gitnexus-guide`
- `gitnexus-impact-analysis`
- `gitnexus-pr-review`
- `gitnexus-refactoring`
- `plannotator-compound`
- `plannotator-setup-goal`
- `plannotator-visual-explainer`
- `playwright`
- `refactor-safely`
- `review-changes`
- `review-delta`
- `review-pr`
- `setup-matt-pocock-skills`
- `source-plugin-code-review`
- `ui-ux-pro-max-skill`
- `web-access`

### REGISTERED_BUT_NOT_FULLY_INSTALLED

- `rtk`
- `graphify`
- `systematic-debugging`
- `trace`
- `deep-interview`
- `writing-plans`
- `tdd`
- `verification`
- `review`
- `github-mcp`
- `mcp-fetch`
- 其余待安装项见 `.scale/skills-registry.json`

### INSTALLED_GLOBAL_MCP_AND_TOOLS

- `openwolf`
- `@playwright/mcp`
- `@modelcontextprotocol/server-memory`
- `gh`
- `jq`
- `rg`

## SKILL_RADAR

- 首先读取 `.scale/skills-registry.json` 的名称、说明、触发条件和安全状态。
- UI/UX 任务优先组合 `design/frontend/browser-testing` 类技能；浏览器和 E2E 任务优先组合 `browser/playwright/devtools` 类能力。
- 安全、权限、数据库、发布任务必须额外选择 `security/review/ship` 类能力。
- `M/L/CRITICAL` 任务必须在 `summary.md` 记录 `skills_used`、`tool_outputs` 和 `skipped_reason`。

## MACHINE_CHECKS

- `must_run`: `bash scripts/validate-config.sh`
- `must_run`: `bash scripts/tests/run.sh`
- `must_run`: `bash scripts/gates/all.sh --dry-run`
- `must_run`: `bash scripts/workflow/verify.sh default`
- `never_claim_passed_without_exit_code_0`: `true`

## HONEST_DELIVERY

- 未运行测试时禁止说“测试通过”。
- 门控失败时禁止说“已完成”。
- 工具缺失或命令跳过时必须列为“未验证项”。
- 最终回复必须包含：完成内容、验证结果、未验证项。

## VERIFICATION_CRITERIA

- `VC1`: diff 中只有请求的改动——无关改动零容忍
- `VC2`: 代码第一次就简洁——无需因过度复杂而重写
- `VC3`: 澄清问题在实现之前提出——不是犯错之后
- `VC4`: 每步修改附带验证——不靠脑补结果

## RED_LINES

- `R1`: 不确定事实必须标注 `[UNCERTAIN]`
- `R2`: 禁止编造文件、命令输出、测试结果
- `R3`: 禁止写入 `.env*`、密钥、证书、token 文件
- `R4`: 声称环境问题前必须给出证据
- `R5`: 零甩锅 -> 失败时先验证自身代码正确性，再排除外部因素
- `R6`: 零未审关键操作 -> 删除文件、修改数据库、变更依赖等关键操作前必须列出影响面并获得确认

## AGENT_BEHAVIORAL_RULES

- `[AB1]` Agent 必须主动使用已安装的 skills 技能，不得忽略可用的技能工具
- `[AB2]` 每次完成任务后，Agent 必须总结经验教训，更新项目知识文档
- `[AB3]` Agent 遇到不确定的问题时，必须先查阅知识库和文档，不得凭空假设
- `[AB4]` Agent 应自主学习和进化：研究新的工具、方法、最佳实践，持续提升能力
- `[AB5]` Agent 必须遵守项目规范：代码风格、命名约定、目录结构、Git 工作流
- `[AB6]` 禁止 Agent 静默跳过验证步骤，所有跳过必须说明原因并获得确认

## KNOWLEDGE_MANAGEMENT

- `[KM1]` 知识沉淀：每次完成重要任务后，更新知识文档中的经验教训章节
- `[KM2]` 知识同步：修改架构/配置/依赖后，同步更新 `CLAUDE.md`、`.agent/project.json`、`.scale/*` 和相关脚本
- `[KM3]` Graphify 知识图谱：当 Graphify 真正安装完成后，使用其维护模块关系、依赖关系和决策记录；当前未安装完成时，不得声称已启用
- `[KM4]` 避免“知识污染”：不确认的信息标记 `[UNCERTAIN]`，过时的信息及时清理
- `[KM5]` 知识库维护：定期检查知识文档与实际代码、脚本、验证命令的一致性

## MULTI_AGENT_CONFLICT_RESOLUTION

- `[MA1]` 资源冲突：多个 Agent 同时修改同一文件时，后修改的 Agent 必须基于最新版本
- `[MA2]` 分支策略：每个 Agent/人类操作应在独立分支上进行，通过 PR/MR 合并
- `[MA3]` 锁机制：修改共享资源（配置文件、数据库 schema）前，检查是否有其他进行中的变更
- `[MA4]` 冲突检测：合并前必须检查冲突，冲突文件必须人工或 Agent 协同解决
- `[MA5]` 通信协议：多 Agent 协作时，通过知识文档和评论通信，避免隐式依赖

## PROJECT_STANDARDS

- `[PS1]` 目录规范：`backend/`、`admin/`、`docs/`、`scripts/`、`.scale/`、`.agent/`
- `[PS2]` Git 工作流：当前主分支是 `main`；功能工作建议使用 `feature/*`、`fix/*`、`docs/*`、`chore/*`
- `[PS3]` 提交规范：`type(scope): subject`，`type = feat/fix/docs/style/refactor/test/chore`
- `[PS4]` 分支同步：开始工作前检查 `git status`，必要时 `git fetch`
- `[PS5]` 代码审查：合并前至少做一次 `review` 视角复核
- `[PS6]` 文档更新：功能变更必须同步更新相关 README、CLAUDE、工作流或契约文档
- `[PS7]` 依赖管理：新增依赖必须说明理由，安装第三方技能或工具前必须审查来源与脚本

<!-- SCALE OS v10.0 adapted for zuhao -->
