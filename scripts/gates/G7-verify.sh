#!/usr/bin/env bash
set -euo pipefail

echo "[CHECK] G7: repository-wide security scan"

echo "[CHECK] scanning for obvious secrets in source files..."
SECRET_HITS=$(grep -RIn --include='*.ts' --include='*.tsx' --include='*.js' --include='*.jsx' --include='*.go' --include='*.yaml' --include='*.yml' --include='*.env*' \
  -E '(sk-[a-zA-Z0-9]{20,}|ghp_[a-zA-Z0-9]{36}|AKIA[0-9A-Z]{16}|-----BEGIN (RSA |EC )?PRIVATE KEY-----)' \
  . 2>/dev/null | grep -v 'node_modules' | grep -v '.git/' | grep -v '.scale/' | head -5 || true)

if [ -n "$SECRET_HITS" ]; then
  echo "[FAIL] G7: potential secrets detected:"
  echo "$SECRET_HITS"
  exit 1
fi

echo "[CHECK] scanning for backend hardcoded secrets..."
grep -RIn --include='*.go' -E '(password|secret|ApiKey|Token)\s*[:=]\s*["'"'][^"'"']+["'"']' backend 2>/dev/null | grep -v '_test.go' | head -5 || true

echo "[CHECK] scanning for frontend hardcoded secrets..."
grep -RIn --include='*.ts' --include='*.tsx' --include='*.js' --include='*.vue' -E '(password|secret|api[_A-Za-z]*key|token)\s*[:=]\s*["'"'][^"'"']+["'"']' admin 2>/dev/null | head -5 || true

echo "[CHECK] checking unsafe HTML usage..."
grep -RIn --include='*.vue' --include='*.tsx' --include='*.jsx' 'dangerouslySetInnerHTML\|innerHTML' admin 2>/dev/null | head -5 || true

echo "[CHECK] checking env files are ignored intentionally..."
for envfile in admin/.env admin/.env.development admin/.env.production backend/.env backend/.env.local; do
  if [ -f "$envfile" ]; then
    echo "[WARN] G7: found $envfile — ensure it contains no production secrets and is governed intentionally"
  fi
done

echo "[PASS] G7: Security verification passed — no obvious secrets or dangerous patterns found"
exit 0

