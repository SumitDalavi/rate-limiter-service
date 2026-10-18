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

## Release Runbook

### Triggering a New Release
Releases are automated via GitHub Actions (`.github/workflows/release.yml`) and trigger whenever a semantic version tag is pushed:

```bash
# 1. Ensure master is clean and up-to-date
git checkout master
git pull origin master

# 2. Tag the release
git tag -a v0.1.0 -m "Release v0.1.0"

# 3. Push the tag to trigger the pipeline
git push origin v0.1.0
```

### Release Pipeline Stages
1. **Test Verification:** Runs project test suite gates prior to container build.
2. **Container Image Build:** Builds the production container image (`ghcr.io/sumitdalavi/rate-limiter-service:<tag>`).
3. **Trivy Vulnerability Scan:** Scans the built container for CRITICAL and HIGH CVEs via Aqua Security Trivy.
4. **Anchore Syft SBOM Generation:** Creates an SPDX JSON Software Bill of Materials (`sbom.spdx.json`).
5. **GitHub Release Publishing:** Creates the release with auto-generated release notes and attaches `sbom.spdx.json`.

### Verifying the Release
```bash
# Check workflow run status
gh run list --workflow=release.yml --repo SumitDalavi/rate-limiter-service

# Inspect the published release and attached assets
gh release view v0.1.0 --repo SumitDalavi/rate-limiter-service

# Download and inspect the SBOM
gh release download v0.1.0 --repo SumitDalavi/rate-limiter-service -p "sbom.spdx.json"
cat sbom.spdx.json | jq .spdxVersion
```

### Release Failure Modes
| Symptom | Cause | Resolution |
|---|---|---|
| `Unable to resolve action ... unable to find version` | Pinned action tag missing or upstream tag modified | Pin action to a verified release tag (`@v0.36.0`) or commit SHA |
| `invalid reference format: repository name must be lowercase` | Container repository owner contains uppercase characters | Lowercase the repository owner before tagging (`tr '[:upper:]' '[:lower:]'`) |
| `File not found: sbom.spdx.json` during release step | SBOM action configured with `artifact-name` instead of `output-file` | Ensure `output-file: sbom.spdx.json` is set in Anchore action |
