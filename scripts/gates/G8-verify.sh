#!/usr/bin/env bash
set -euo pipefail

SLOP_PATTERNS=0
echo "[CHECK] G8: scanning for AI-generated code quality issues..."

AI_REMNANTS=$(grep -RIn --include='*.ts' --include='*.tsx' --include='*.js' --include='*.py' --include='*.go' --include='*.java' \
  -iE "// (i'll|let me|now let's|i will|i can|here's)|# (i'll|let me|now let's|i will|i can|here's)" . 2>/dev/null | grep -v 'node_modules' | grep -v '.agent/' | head -5 || true)

if [ -n "$AI_REMNANTS" ]; then
  echo "[WARN] G8: Found AI conversation remnants in comments:"
  echo "$AI_REMNANTS"
  SLOP_PATTERNS=$((SLOP_PATTERNS + 1))
fi

EMPTY_TODOS=$(grep -RIn --include='*.ts' --include='*.tsx' --include='*.js' --include='*.py' --include='*.go' --include='*.java' \
  -E '// TODO:?\s*$|# TODO:?\s*$' . 2>/dev/null | grep -v 'node_modules' | grep -v '.agent/' | head -5 || true)

if [ -n "$EMPTY_TODOS" ]; then
  echo "[WARN] G8: Found empty TODO comments without descriptions:"
  echo "$EMPTY_TODOS"
  SLOP_PATTERNS=$((SLOP_PATTERNS + 1))
fi

if [ "$SLOP_PATTERNS" -gt 2 ]; then
  echo "[FAIL] G8: Too many AI-slop patterns detected"
  exit 1
fi

echo "[PASS] G8: No-AI-slop check passed (warnings: $SLOP_PATTERNS)"
exit 0

