#!/usr/bin/env bash
set -euo pipefail

EVIDENCE_DIR=".scale/evidence"
EXPLORE_FILE=".scale/explore.md"
found_evidence=0

if [ -f "$EXPLORE_FILE" ]; then
  content=$(grep -v '^#' "$EXPLORE_FILE" | grep -v '^$' | wc -l | tr -d ' ')
  if [ "$content" -gt 3 ]; then
    echo "[PASS] G1: Exploration evidence found in $EXPLORE_FILE"
    found_evidence=1
  fi
fi

if [ -d "$EVIDENCE_DIR" ]; then
  file_count=$(find "$EVIDENCE_DIR" -type f -name '*.md' 2>/dev/null | wc -l | tr -d ' ')
  if [ "$file_count" -gt 0 ]; then
    echo "[PASS] G1: Exploration evidence found in $EVIDENCE_DIR ($file_count files)"
    found_evidence=1
  fi
fi

if [ -f ".scale/g1-verified" ]; then
  echo "[PASS] G1: Exploration manually verified"
  found_evidence=1
fi

if [ "$found_evidence" -eq 0 ]; then
  echo "[FAIL] G1: No exploration evidence found. Create .scale/explore.md or .scale/g1-verified"
  exit 1
fi

exit 0

