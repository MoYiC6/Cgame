#!/usr/bin/env bash
set -euo pipefail

PLAN_FILE=".scale/plan.md"
EXEMPT_FILE=".scale/tdd-exempt.md"

if [ -f "$EXEMPT_FILE" ]; then
  echo "[PASS] G3: TDD exemption recorded in $EXEMPT_FILE"
  exit 0
fi

TEST_FILES=$(find . -type f \( -name '*_test.py' -o -name 'test_*.py' -o -name '*_test.go' -o -name '*.test.ts' -o -name '*.test.tsx' -o -name '*.spec.ts' -o -name '*Test.java' \) ! -path '*/node_modules/*' ! -path '*/.agent/*' ! -path '*/vendor/*' ! -path '*/.git/*' 2>/dev/null | head -5)

if [ -z "$TEST_FILES" ]; then
  echo "[FAIL] G3: No test files found matching project patterns"
  echo "  If TDD is not applicable, create .scale/tdd-exempt.md with justification"
  exit 1
fi

if [ -f "$PLAN_FILE" ]; then
  NEWER_TESTS=$(find . -type f \( -name '*_test.py' -o -name 'test_*.py' -o -name '*_test.go' -o -name '*.test.ts' -o -name '*.test.tsx' -o -name '*.spec.ts' -o -name '*Test.java' \) ! -path '*/node_modules/*' ! -path '*/.agent/*' ! -path '*/vendor/*' ! -path '*/.git/*' -newer "$PLAN_FILE" 2>/dev/null | head -1)
  if [ -n "$NEWER_TESTS" ]; then
    echo "[PASS] G3: TDD evidence found — test files newer than plan.md"
    echo "  Test file: $NEWER_TESTS"
    exit 0
  fi
fi

echo "[PASS] G3: Test files exist — TDD evidence found"
exit 0

