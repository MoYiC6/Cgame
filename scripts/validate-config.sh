#!/usr/bin/env bash
set -euo pipefail

failures=0
warnings=0

echo "=== SCALE OS Configuration Validator ==="
echo ""

echo "[CHECK] File existence..."
required=(
  ".agent/project.json"
  ".agent/report.json"
  ".scale/workflow.json"
  ".scale/quality-contract.json"
  ".scale/skills-registry.json"
)

for file in "${required[@]}"; do
  if [ ! -f "$file" ]; then
    echo "[FAIL] missing $file"
    failures=$((failures + 1))
  else
    echo "[OK] $file exists"
  fi
done

optional_files=(
  "SCALE-REPORT.md"
  "docs/workflow/QUALITY_CONTRACT.md"
  "scripts/gates/all.sh"
  "scripts/workflow/verify.sh"
  "scripts/tests/run.sh"
  ".claude/hooks/session-start-reminder.sh"
  ".claude/hooks/session-end-gate.sh"
)

for file in "${optional_files[@]}"; do
  if [ ! -f "$file" ]; then
    echo "[WARN] missing optional $file"
    warnings=$((warnings + 1))
  else
    echo "[OK] optional $file exists"
  fi
done

echo ""
echo "[CHECK] JSON validity..."
for file in .agent/project.json .agent/report.json .scale/workflow.json .scale/quality-contract.json .scale/skills-registry.json; do
  if node -e "JSON.parse(require('fs').readFileSync(process.argv[1], 'utf8'))" "$file" 2>/dev/null; then
    echo "[OK] $file is valid JSON"
  else
    echo "[FAIL] $file has invalid JSON syntax"
    failures=$((failures + 1))
  fi
done

echo ""
echo "[CHECK] Project/service consistency..."
if [ -f ".agent/project.json" ]; then
  if [ ! -d "backend" ]; then
    echo "[FAIL] backend directory missing"
    failures=$((failures + 1))
  else
    echo "[OK] backend directory exists"
  fi

  if [ ! -d "admin" ]; then
    echo "[FAIL] admin directory missing"
    failures=$((failures + 1))
  else
    echo "[OK] admin directory exists"
  fi

  STACK=$(node -e "const d=JSON.parse(require('fs').readFileSync('.agent/project.json','utf8')); console.log(d.stack||'unknown')" 2>/dev/null || echo "unknown")
  if [ "$STACK" != "polyglot" ]; then
    echo "[WARN] expected stack 'polyglot', got '$STACK'"
    warnings=$((warnings + 1))
  else
    echo "[OK] stack is polyglot"
  fi
fi

echo ""
echo "[CHECK] Executable scripts..."
for file in scripts/validate-config.sh scripts/tests/run.sh scripts/gates/all.sh scripts/workflow/verify.sh; do
  if [ ! -x "$file" ]; then
    echo "[FAIL] script is not executable: $file"
    failures=$((failures + 1))
  else
    echo "[OK] executable: $file"
  fi
done

echo ""
echo "=== Validation Summary ==="
if [ "$failures" -gt 0 ]; then
  echo "[FAIL] $failures failure(s), $warnings warning(s)"
  exit 1
fi

echo "[OK] Configuration valid ($warnings warning(s))"

