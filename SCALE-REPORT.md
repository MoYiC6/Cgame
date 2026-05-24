# SCALE Generation Report

- Project: zuhao
- Agent: claude-code
- Support Level: full
- Stack: polyglot (Go backend + Vue admin)

## Must Run
- `bash scripts/validate-config.sh`
- `bash scripts/tests/run.sh`
- `bash scripts/gates/all.sh --dry-run`
- `bash scripts/workflow/verify.sh default`

## Degraded Or Pending
- `frontend-tests`: `admin/` 目前没有独立自动化测试脚本，`frontend` profile 暂时只强制 lint/build/typecheck。

## Quality Governance
- Runtime config: `.agent/project.json`
- Workflow state: `.scale/workflow.json`
- Quality contract: `.scale/quality-contract.json`
- Skill registry: `.scale/skills-registry.json`

## Honest Delivery
- Do not claim tests passed unless the command was actually run and exited 0.
- `--dry-run` only proves scripts are schedulable; it does not prove quality gates passed.
- Skipped or missing checks must be reported explicitly.
