# Runbook — rate-limiter-service
> Last updated: 2026-08-29

## Prerequisites
| Tool | Required Version | How to check |
|---|---|---|
| Go | >= 1.21 | `go version` |
| Docker & Compose | Latest | `docker-compose version` |

## Quick Start
```bash
# Start Redis and the Go API
docker-compose up -d --build

# Verify
curl http://localhost:8080/health
```

## Run Tests
```bash
# Unit tests
go test ./... -v

# E2E Redis Outage Test
bash tests/e2e/test_redis_outage.sh
```

Expected output:
```
?       github.com/SumitDalavi/rate-limiter-service     [no test files]
ok      github.com/SumitDalavi/rate-limiter-service/middleware  0.015s
```

## Environment Variables
| Variable | Default | Purpose |
|---|---|---|
| SERVER_PORT | `8080` | HTTP port |
| REDIS_URL | `redis:6379` | Redis connection string |
| LIMIT_RATE | `10` | Tokens per second |
| LIMIT_BURST | `20` | Max burst capacity |

## Common Failure Modes
| Symptom | Cause | Fix |
|---|---|---|
| Every request gets 429 | Redis connection failed | Ensure `REDIS_URL` is reachable |
| High latency | Redis is overloaded | Check Redis CPU and memory usage |
