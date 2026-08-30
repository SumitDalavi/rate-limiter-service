#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

# Ensure Redis is running
if ! nc -z localhost 6379 2>/dev/null; then
    echo "Starting Redis via docker-compose..."
    docker compose up -d redis
    sleep 3
fi

echo "Running Rate Limiter Benchmark..."
go run benchmarks/run.go
