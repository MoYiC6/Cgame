#!/usr/bin/env bash
set -euo pipefail

failures=0
for file in .agent/project.json .agent/report.json .scale/workflow.json .scale/quality-contract.json .scale/skills-registry.json; do
  node -e "JSON.parse(require('fs').readFileSync(process.argv[1], 'utf8'))" "$file" || failures=$((failures + 1))
done

if [ ! -x scripts/validate-config.sh ] || [ ! -x scripts/gates/all.sh ] || [ ! -x scripts/workflow/verify.sh ]; then
  echo "[FAIL] expected generated scripts to be executable"
  failures=$((failures + 1))
fi

echo "$failures 失败"
[ "$failures" -eq 0 ]

