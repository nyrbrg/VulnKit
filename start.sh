#!/usr/bin/env bash
set -euo pipefail
set -m
cd "$(dirname "$0")"

echo "==> Starting postgres + labs containers..."
docker compose up -d --wait postgres vulnkit-labs

cleanup() {
  echo "==> Shutting down backend/frontend..."
  [ -n "${BACKEND_PID:-}" ] && kill -TERM -"$BACKEND_PID" 2>/dev/null || true
  [ -n "${FRONTEND_PID:-}" ] && kill -TERM -"$FRONTEND_PID" 2>/dev/null || true
  wait "${BACKEND_PID:-}" "${FRONTEND_PID:-}" 2>/dev/null || true
}
trap cleanup EXIT INT TERM

echo "==> Starting backend on :8080..."
(cd backend && go run ./cmd/server) &
BACKEND_PID=$!

echo "==> Starting frontend on :5173..."
(cd frontend && npm run dev) &
FRONTEND_PID=$!

wait
