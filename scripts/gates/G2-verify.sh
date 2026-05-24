#!/usr/bin/env bash
set -euo pipefail

PLAN_FILE=".scale/plan.md"

if [ ! -f "$PLAN_FILE" ]; then
  echo "[FAIL] G2: Plan document missing at $PLAN_FILE"
  echo "  Create plan.md with Impact Analysis, Verification Strategy, and Rollback Plan"
  exit 1
fi

missing_sections=()
if ! grep -qi 'impact\|影响' "$PLAN_FILE"; then
  missing_sections+=("Impact Analysis / 影响分析")
fi
if ! grep -qi 'verif\|验证\|test' "$PLAN_FILE"; then
  missing_sections+=("Verification Strategy / 验证策略")
fi
if ! grep -qi 'rollback\|回滚' "$PLAN_FILE"; then
  missing_sections+=("Rollback Plan / 回滚计划")
fi

if [ ${#missing_sections[@]} -gt 0 ]; then
  echo "[FAIL] G2: Plan exists but missing required sections:"
  for section in "${missing_sections[@]}"; do
    echo "  - $section"
  done
  exit 1
fi

echo "[PASS] G2: Plan document verified with required sections"
exit 0

