---
name: add-acceptance-test
description: Write CI-executable acceptance/integration tests for protocol or end-to-end changes (TR-F022). Use whenever a milestone subtask needs CI-backed acceptance, covering test/integration integration tests and unit tests.
---

# Skill: Write CI Automated Acceptance Tests

> Feature ID: `TR-F022` | Scope: all protocol / end-to-end acceptance

## Goal
Every milestone subtask's acceptance criteria must be backed by a **CI-executable test**. This skill explains where the two test types live and how to write them.

## Two test types

### 1. Unit / logic acceptance - `*_test.go`
- Same package as the code under test; the CI `test` job runs `go test -short ./...`.
- For: pure logic, attribute encode/decode, validators, config parsing, etc.
- How-to: see `../write-go-tests/SKILL.md`.

### 2. End-to-end / protocol acceptance - `test/integration/`
- First line `//go:build integration`, package `integration`.
- Backed by a real PostgreSQL (the production default database), driving live RADIUS / Admin services with real packets.
- The CI `integration` job runs it automatically:
  ```
  go test -tags=integration -count=1 -v ./test/integration/...
  INTEGRATION_REQUIRED=1   # CI hard-fails when the environment is missing, preventing false green
  ```
- Existing templates:
  - `test/integration/main_test.go` - shared harness (DB setup, service startup)
  - `test/integration/radius_test.go` - real PAP Access-Request end-to-end
  - `test/integration/adminapi_test.go` - Admin API end-to-end
  - `test/integration/migration_test.go` - migration
  - `test/integration/client_test.go`

## Pre-research
```text
view test/integration/main_test.go          # harness usage
view test/integration/radius_test.go         # RADIUS end-to-end template
grep_search "INTEGRATION_REQUIRED" --include test/**
view Makefile                                # test / test-integration-pg targets
```

## Local run
```bash
# unit
go test ./...
# integration (auto-starts the Postgres in docker-compose.test.yml, cleans up afterwards)
make test-integration-pg
# or, after manually providing TEST_DATABASE_*:
go test -tags=integration -count=1 -v ./test/integration/...
```

## Steps to add an end-to-end acceptance case
1. Create or extend `<feature>_test.go` under `test/integration/`, with `//go:build integration` as the first line.
2. Reuse the `main_test.go` harness to create data and start services; drive real packets for the new feature (e.g. EAP-TLS / CoA) following the existing `radius_test.go` pattern.
3. Assert both success **and** failure paths (reject reasons, timeouts, NAS rejects, etc.).
4. Do not `t.Parallel()` for cases that serially depend on global state (see the comments in `radius_test.go`).
5. Confirm the CI `integration` job covers the path (it runs all of `./test/integration/...` by default; no CI change needed).

## Boundaries
- Integration cases must not depend on external public services; provide dependencies via a service container / compose.
- Do not weaken assertions or skip failure paths just to pass.

## Acceptance
- [ ] The new feature has a corresponding CI automated acceptance case (unit or integration)
- [ ] Failure paths are asserted
- [ ] `make test-integration-pg` passes locally
- [ ] The CI `test` and `integration` jobs pass
