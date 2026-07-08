#!/usr/bin/env bash
set -euo pipefail

KNOWLEDGE_DOC="CLAUDE.md"
SUMMARY_FILE=".scale/task-summary.md"
found_update=0

if [ -f "$KNOWLEDGE_DOC" ]; then
  content=$(grep -v '^#' "$KNOWLEDGE_DOC" | grep -v '^$' | grep -v '^>' | wc -l | tr -d ' ')
  if [ "$content" -gt 5 ]; then
    echo "[PASS] G9: Knowledge document $KNOWLEDGE_DOC has substantive content"
    found_update=1
  fi
fi

if [ -f "$SUMMARY_FILE" ]; then
  echo "[PASS] G9: Task summary exists at $SUMMARY_FILE"
  found_update=1
fi

if [ -f ".scale/g9-verified" ]; then
  echo "[PASS] G9: Knowledge update manually verified"
  found_update=1
fi

if [ "$found_update" -eq 0 ]; then
  echo "[WARN] G9: No knowledge update evidence found"
  echo "  Update $KNOWLEDGE_DOC or create .scale/task-summary.md after completing tasks"
fi

exit 0

