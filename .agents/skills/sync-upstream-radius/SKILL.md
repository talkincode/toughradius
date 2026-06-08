---
name: sync-upstream-radius
description: Track and sync the upstream layeh.com/radius library to the org fork talkincode/radius (TR-F021/TR-F022). Use when a check finds new upstream commits, a protocol-library defect is suspected, or for routine dependency maintenance.
---

# Skill: Track & Sync the Upstream RADIUS Library

> Feature IDs: `TR-F021` / `TR-F022` | Milestone: M7

## Background
- Core protocol library: `layeh.com/radius` (original repo <https://github.com/layeh/radius>).
- Org fork: `github.com/talkincode/radius` (<https://github.com/talkincode/radius>).
- Wiring: in `go.mod`
  ```
  require layeh.com/radius v0.0.0-<pseudo>
  replace layeh.com/radius => github.com/talkincode/radius v<tag>
  ```
  i.e. the build actually uses the fork, while the `layeh.com/radius` import path stays unchanged.

## When to use
- Routine (weekly recommended) manual check of whether upstream `layeh/radius` has new commits.
- When a protocol codec / attribute-handling / security defect is found or suspected.
- During routine dependency maintenance.

## Steps
1. **Confirm the currently pinned version**
   ```bash
   grep -E "layeh.com/radius|talkincode/radius" go.mod
   ```
   Record the fork tag the `replace` points to, and the upstream commit short SHA in the `require` pseudo-version.
2. **Diff against upstream**
   - Original repo `layeh/radius`: review new commits since the pinned commit, e.g. via GitHub compare:
     `https://github.com/layeh/radius/compare/<pinned_sha>...master`, or `git log <pinned_sha>..` in an upstream clone.
   - Focus on: security fixes, attribute codec correctness, EAP / VSA / packet-parsing changes.
3. **Assess the sync**
   - If an upstream fix is important and the fork lacks it: merge/cherry-pick it into the fork (`talkincode/radius`) and cut a new tag.
   - Update this repo's `go.mod` `replace` to the new tag and run `go mod tidy`.
   - If the fix does not affect this project's usage, record the "assessed, sync deferred" rationale (reply in the tracking issue).
4. **Verify**
   ```bash
   go build ./...
   go test ./...
   go test -tags=integration -count=1 ./test/integration/...   # needs Postgres, see ../add-acceptance-test/SKILL.md
   golangci-lint run
   ```
5. **Regression protection**: if the sync fixes a protocol defect, add a test that reproduces it to prevent future regressions.

## Boundaries
- Do not vendor or patch third-party library source directly in this repo; changes go through the fork repo.
- `replace` points only to the trusted org fork; do not introduce unknown third-party branches.
- Upgrades must be explained in the PR: upstream commit range, risk assessment, and whether protocol behavior is affected.

## Acceptance
- [ ] `go.mod` / `go.sum` are consistent, `go mod tidy` leaves no residue
- [ ] Full + integration tests pass
- [ ] The tracking issue records the sync decision (sync / defer + rationale)
- [ ] PR references `TR-F021` / `TR-F022`
