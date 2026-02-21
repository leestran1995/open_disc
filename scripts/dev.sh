#!/usr/bin/env bash
#
# Start the open_disc dev environment (backend + frontend)
# Usage: ./scripts/dev.sh
#
# Ctrl+C to stop both servers.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

cleanup() {
  echo ""
  echo "Shutting down..."
  kill $BACKEND_PID 2>/dev/null || true
  kill $FRONTEND_PID 2>/dev/null || true
  wait $BACKEND_PID 2>/dev/null || true
  wait $FRONTEND_PID 2>/dev/null || true
  echo "Done."
}
trap cleanup EXIT

# Kill anything already on our ports
for port in 8080 8081 4000; do
  pid=$(lsof -ti:$port 2>/dev/null || true)
  if [ -n "$pid" ]; then
    echo "Killing existing process on port $port (pid $pid)"
    kill $pid 2>/dev/null || true
    sleep 0.5
  fi
done

# Start backend
echo "Starting backend (ports 8080 + 8081)..."
cd "$ROOT_DIR"
go run ./internal/main/main.go &
BACKEND_PID=$!

# Start frontend
echo "Starting frontend (port 4000)..."
cd "$ROOT_DIR/frontend"
bun run dev &
FRONTEND_PID=$!

echo ""
echo "Open http://localhost:4000"
echo "Press Ctrl+C to stop."
echo ""

wait
