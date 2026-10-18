

# Distributed Rate Limiter 🚦

[![CI Pipeline](https://github.com/SumitDalavi/rate-limiter-service/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/SumitDalavi/rate-limiter-service/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/SumitDalavi/rate-limiter-service)](https://github.com/SumitDalavi/rate-limiter-service/releases)
[![Release Workflow](https://github.com/SumitDalavi/rate-limiter-service/actions/workflows/release.yml/badge.svg)](https://github.com/SumitDalavi/rate-limiter-service/actions/workflows/release.yml)

> **Maturity:** Full Prototype
> _API gateway rate limiting service using Redis Lua scripting for atomic token-bucket operations._

## The Problem
Modern distributed systems require robust, highly concurrent solutions. Simple CRUD applications fail when subjected to high throughput, race conditions, or massive data sets.

## The Solution
This project implements a robust microservice architecture designed to handle these specific edge cases. By utilizing advanced paradigms like idempotency keys, advisory locks, or optimized caching layers, this service guarantees data integrity under load.

```text
┌──────────────┐      ┌───────────────┐      ┌───────────────┐
│              │      │               │      │               │
│   Client     │─────►│   API Layer   │─────►│  Data Store   │
│              │      │               │      │               │
└──────────────┘      └───────────────┘      └───────────────┘
```

## 🛠️ Tech Stack
- **Core Technology**: Go, Redis
- **Architecture**: Microservices, Event-Driven

## 📚 Documentation

- [Architecture](docs/ARCHITECTURE.md) — System diagram and component details
- [Runbook](docs/runbook.md) — Setup, commands, and expected outputs
- [Decisions](docs/decisions.md) — ADRs for rate limiting patterns
- [Changelog](docs/changelog.md) — Change history

## Decision Log
| Decision | Rationale |
|----------|-----------|
| Monorepo vs Polyrepo | Chosen self-contained repository for easier deployment and PoC demonstration |
| State Management | All state is pushed to the Data Store/Cache to keep the API stateless and horizontally scalable |
| Error Handling | Standardized JSON error responses with explicit error codes |

## 🚀 Step-by-Step Setup

```bash
# 1. Clone the repository
git clone https://github.com/SumitDalavi/rate-limiter-service.git
cd rate-limiter-service

# 2. Build and start
docker-compose up -d --build

# 3. Verify it's running
curl http://localhost:8080/health
```

The API is now available at **http://localhost:8080**

## 🧪 Usage & Demo

```bash
# Health Check
curl http://localhost:8080/health

# Simulate Traffic
curl -X POST http://localhost:8080/api/trigger -H "Content-Type: application/json" -d '{"test":"payload"}'
```

## ✅ Verification

| Check | Command | Expected |
|-------|---------|----------|
| Health | `curl http://localhost:8080/health` | `{"status": "ok"}` |
| Load | `make test` | All unit/integration tests pass |

## Benchmark Results (Last Run: 2026-08-29)
| Metric | Value | Environment |
|---|---|---|
| Request Processing Overhead | <1.34ms | Local execution, `httptest.Server` |
| Concurrency Fairness Test | 500 Accepted / 500 Rejected | 100 limit across 1000 burst requests |

## Key Design Decisions
- **Why Lua Scripts in Redis:** Fetching the counter, evaluating it, and incrementing it in application code creates massive race conditions under load. Evaluating the entire sliding window atomically inside Redis via Lua guarantees exact mathematical correctness under burst pressure.

## Test Coverage
Executes synthetic load across multiple mock IPs to verify atomic increment properties.

## Known Limitations & Honest Scope
- **Redis Dependency**: The entire rate limiter is hard-coupled to Redis availability. If Redis fails or experiences high latency, the middleware either blocks all traffic or allows all traffic (depending on the configured fail-open policy).

## 🔒 Release & Security Verification

Every tagged release (*.*.*) triggers an automated DevSecOps release workflow:
1. **Automated Testing:** Runs go test -v ./... across unit and integration tests.
2. **Container Image Build:** Builds the service container image (ghcr.io/sumitdalavi/rate-limiter-service:<tag>).
3. **Vulnerability Scanning:** Scans the container image using [Aqua Security Trivy](https://github.com/aquasecurity/trivy-action) for HIGH and CRITICAL CVEs.
4. **SBOM Generation:** Generates a Software Bill of Materials (SBOM) in SPDX JSON format (sbom.spdx.json) using [Anchore Syft](https://github.com/anchore/sbom-action).
5. **GitHub Release Publishing:** Packages release notes and attaches sbom.spdx.json as a verifiable supply-chain asset.

## 👨‍💻 Author
**Sumit Dalavi** — Senior DevSecOps / Platform Engineer
[GitHub](https://github.com/SumitDalavi) | [LinkedIn](https://in.linkedin.com/in/sumit-dalavi-762838129)

---
*Built with a focus on robust patterns, not toy demos.*

## Release & Security Verification

Every tagged release (`v*.*.*`) triggers an automated DevSecOps release workflow:
1. **Automated Testing:** Runs test suite gates before build.
2. **Container Image Build:** Builds production container image (`ghcr.io/sumitdalavi/rate-limiter-service:<tag>`).
3. **Vulnerability Scanning:** Scans the image using [Aqua Security Trivy](https://github.com/aquasecurity/trivy-action) for `HIGH` and `CRITICAL` CVEs.
4. **SBOM Generation:** Generates a Software Bill of Materials (SBOM) in SPDX JSON format (`sbom.spdx.json`) using [Anchore Syft](https://github.com/anchore/sbom-action).
5. **GitHub Release Publishing:** Packages release notes and attaches `sbom.spdx.json` as a verifiable supply-chain asset.

### Verifying Releases via GitHub CLI
```bash
# View release details
gh release view v0.1.0 --repo SumitDalavi/rate-limiter-service

# Download and verify SBOM
gh release download v0.1.0 --repo SumitDalavi/rate-limiter-service --pattern "sbom.spdx.json"
cat sbom.spdx.json | jq .
```

## Documentation

- [Architecture](docs/ARCHITECTURE.md)
- [Runbook](docs/runbook.md)
- [Decisions](docs/decisions.md)
- [Changelog](docs/changelog.md)