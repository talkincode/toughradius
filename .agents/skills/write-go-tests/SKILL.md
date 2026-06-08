---
name: write-go-tests
description: Unified conventions for writing Go unit/integration tests (TR-F022). Use when adding tests for any Go code change.
---

# Skill: Write Go Unit / Integration Tests

> Feature ID: `TR-F022`

## When to use
Any Go code change must ship with tests. This skill unifies the test conventions.

## Pre-research
```text
file_search "internal/**/*_test.go"            # find the closest template
view internal/adminapi/nodes_test.go           # API test template
view internal/radiusd/*_test.go                # protocol test template
view internal/app/*_test.go                    # app-layer test template
```

## Conventions
1. **Co-located tests**: the test file is in the same package and directory as the code under test, named `<file>_test.go`.
2. **Short-test marking**: for slow / externally dependent tests use `if testing.Short() { t.Skip(...) }`; CI runs `go test -short`.
3. **Database**: dev / test use SQLite (pure Go, `CGO_ENABLED=0`); schema changes must be compatible with both databases.
4. **Coverage focus**: success path + failure path (reject reasons, timeouts, validation failures, authz failures).
5. **Metrics / errors**: auth-reject changes must assert the corresponding `AuthError` / metrics tag.
6. **Table-driven**: use table-driven tests for multiple scenarios.

## Run
```bash
go test ./...                       # all
go test -short ./...                # CI-equivalent (fast)
go test -run TestXxx ./internal/... # single test
go test -bench=. ./internal/radiusd/ # benchmark
golangci-lint run                   # lint (v2.12.2)
```

## Acceptance
- [ ] New / changed logic has test coverage
- [ ] `go test ./...` and `golangci-lint run` pass
- [ ] Critical failure paths are asserted
