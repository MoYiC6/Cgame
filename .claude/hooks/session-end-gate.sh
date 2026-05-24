#!/usr/bin/env bash
set -euo pipefail

STATE_FILE=".claude/session/.flow-state"
VERIFY_MARKER=".claude/session/.verified"

if [[ -f "$STATE_FILE" ]]; then
  PHASES="SKILL_SCAN EXPLORE PLAN EXECUTE VERIFY SETTLE"
  MISSING=""
  for phase in $PHASES; do
    if ! grep -q "${phase}=✓" "$STATE_FILE" 2>/dev/null; then
      MISSING="$MISSING $phase"
    fi
  done
  if [[ -n "$MISSING" ]]; then
    echo "[GATE BLOCK] 认知工作流未完成 — 缺失:$MISSING"
    exit 2
  fi
fi

if [[ -f "$VERIFY_MARKER" ]]; then
  echo "[GATE PASS] 认知工作流验证通过"
else
  echo "[WARN] 未发现显式验证标记 .claude/session/.verified，继续依赖 scripts/gates/all.sh 与 scale before-stop"
fi

exit 0

