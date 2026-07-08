#!/usr/bin/env bash
set -euo pipefail

source "scripts/lib/project-config.sh"
cmd="$(command_for_gate "typecheck")"

if [ -z "$cmd" ] || [ "$cmd" = "N/A" ]; then
  echo "[SKIP] G6 Type verification: no configured command"
  exit 0
fi

if [[ "$cmd" == echo*"["*"]"* ]]; then
  echo "[SKIP] G6 Type verification: placeholder command"
  exit 0
fi

echo "[RUN] G6 Type verification: $cmd"
bash -lc "$cmd"

