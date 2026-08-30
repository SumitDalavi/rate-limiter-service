# Changelog

## [2026-08-29] — Phase 2 Evidence
### Added
- Added `tests/e2e/test_redis_outage.sh` script to verify behavior during Redis failure.
- Added reproducible benchmark evidence `benchmarks/results/rate_limit_fairness.json` demonstrating fairness and burst behavior.
- Standardized documentation (`runbook.md`, `decisions.md`, `ARCHITECTURE.md`).
- Added maturity badge and mock boundaries to `README.md`.

### Post-Release Hotfixes
- Resolved lingering CI/CD failures introduced during portfolio elevation.
- Fixed Docker build and permission errors across client/server components.
- Corrected Kubernetes controller GroupVersionKind mismatches and E2E Vault addressing.
- Repaired broken property-based test configurations.
