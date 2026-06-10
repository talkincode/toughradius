# Agent Development Guide

> 中文版本：[Agent 开发指南](../zh/agent-guide.md)

This chapter is a **contributor-oriented** digest of how ToughRADIUS is built
with AI coding agents. It summarizes the working rules, quality gates, and the
auto-delegation loop so the workflow is discoverable from the handbook.

The **canonical** rules live in the repository's
[`AGENT.md`](https://github.com/talkincode/toughradius/blob/main/AGENT.md);
that file stays authoritative and is referenced directly by the agent tooling.
This chapter does not replace it — when in doubt, follow `AGENT.md`.

## Product scope baseline

Development stays anchored to the feature checklist; it never drifts into
unrelated product directions.

- The canonical scope baseline is
  [`docs/feature-checklist.md`](https://github.com/talkincode/toughradius/blob/main/docs/feature-checklist.md)
  (with an English copy at
  [`docs/feature-checklist.en.md`](https://github.com/talkincode/toughradius/blob/main/docs/feature-checklist.en.md)).
- Every task, issue, PR, test, and review note maps to a feature ID such as
  `TR-F004`.
- If a request does not map to an existing ID, the checklist is updated first
  (scope, status, acceptance boundary, rationale) before any code changes.
- Non-goals `TR-N001`–`TR-N005` (payment, CRM, generic monitoring stack,
  multi-tenancy, full rewrite) are out of scope unless the checklist is
  explicitly revised first.

## Roadmap and skill library

Agent-driven development is organized around three artifacts:

- [`docs/roadmap.md`](https://github.com/talkincode/toughradius/blob/main/docs/roadmap.md)
  — long-term roadmap and milestones, each mapped to `TR-F` IDs. This is the
  task source for agent work.
- [`.agents/skills/`](https://github.com/talkincode/toughradius/tree/main/.agents/skills)
  — reusable skill SOPs, one folder per skill (`.agents/skills/<name>/SKILL.md`).
- [`.agents/README.md`](https://github.com/talkincode/toughradius/blob/main/.agents/README.md)
  — the delegation reference and shared guardrails.

A **coordinator layer** drives the loop, while the execution SOPs do the
domain-specific work:

| Role | Skill | Purpose |
| --- | --- | --- |
| Coordinator | `orchestrate-roadmap` | Entry role for "auto-delegate development": selects the next unchecked subtask, picks the matching SOP, enforces gates, opens a PR |
| Gate | `review-pr` | Independent, CI-anchored review; requests changes via labels/comments and auto-merges only when approved **and** CI is green |
| Self-iteration | `groom-roadmap` | After each merge, checks off the delivered subtask and re-grooms the roadmap |

Execution SOPs include: add vendor VSA (`add-radius-vendor`), add EAP method
(`add-eap-method`), add Admin API (`add-adminapi-endpoint`), add React Admin
resource (`add-react-admin-resource`), add config schema (`add-config-schema`),
add acceptance test (`add-acceptance-test`), sync upstream radius
(`sync-upstream-radius`), reference RFC (`reference-rfc`), align checklist
(`align-feature-checklist`), write Go tests (`write-go-tests`), and document Go
APIs (`document-go-apis`). Pick the matching skill before starting a task type.

Agents run **on your own host** with your own agent/CLI, not via a CI workflow,
so credentials never enter CI and the execution environment stays under your
control.

## Working guidelines

### Understand existing code before editing

Never change code blindly. Locate the existing implementation, related tests,
and documentation first, then mirror the project's naming, error handling, and
data flow. Trace the full execution path before fixing a bug, and map
dependencies and side effects before refactoring.

### Continuous verification

Do not wait until the end to run tests. Every logical change is followed by a
test run so regressions surface immediately rather than at the end of a large
batch.

### Code is the best documentation

- **All exported APIs carry comprehensive godoc comments** — purpose,
  parameters with constraints, return values, and error conditions (with usage
  examples for complex APIs). See the `document-go-apis` skill for the
  standard-library-style conventions.
- **Complex logic carries inline comments** explaining the *why*, not the *what*.
- **Vendor-specific code references the protocol spec** (RFC numbers, VSA docs).
- **No redundant standalone summary documents** — information belongs in code
  comments and Git history.

## Core development principles

### Test-driven development (TDD)

Write the test first, then the code: define the expected behavior in a failing
test (red), write the minimum code to pass it (green), then refactor while the
tests stay green. If `internal/radiusd/auth.go` changes, the test lives in
`internal/radiusd/auth_test.go`. See the `write-go-tests` skill for the unified
conventions.

### GitHub workflow

- **Pull requests only** — never push to `main`; the protected branch rejects
  direct pushes.
- **Conventional commits** — `<type>(<scope>): <subject>` with types such as
  `feat`, `fix`, `test`, `docs`, `refactor`, `perf`, `chore`.
- **Small, atomic changes** over giant PRs.

### Minimum viable product (MVP)

Each change is delivered as a minimal, independently usable, rollback-safe unit
that does not break existing behavior. Large efforts are broken into MVP
increments (for example, vendor attribute parsing → auth integration →
accounting → management UI) rather than landing in one oversized PR.

## Quality gates

Every agent change must pass these gates before merging:

- `go build ./...` — no compilation errors.
- `go test ./...` — all unit tests pass.
- `golangci-lint run` — clean (pinned to **v2.12.2**, matching CI).
- `cd web && npm run build` — for any frontend change.
- **Protocol / end-to-end changes** ship a CI-executable acceptance test under
  [`test/integration/`](https://github.com/talkincode/toughradius/tree/main/test/integration)
  and cite the relevant spec under
  [`docs/rfcs/`](https://github.com/talkincode/toughradius/tree/main/docs/rfcs).
- Output goes through a PR labeled `agent-roadmap`, gated by `review-pr`, and is
  merged only when `agent-approved` with green CI.

## Technical constraints

- **No CGO** — the project builds with `CGO_ENABLED=0` for easy cross-platform
  deployment. Use pure Go drivers only (for example `github.com/glebarez/sqlite`
  instead of `github.com/mattn/go-sqlite3`).
- **Database dual-compatibility** — every schema change must work on both
  PostgreSQL (default) and SQLite.
- **Upstream dependency** — the core `layeh.com/radius` library is `replace`d in
  `go.mod` to the organization fork `github.com/talkincode/radius`; important
  upstream fixes are evaluated via the `sync-upstream-radius` skill.

## Common anti-patterns (prohibited)

- Exporting APIs without documentation.
- Committing implementation without tests.
- Giant PRs that mix many concerns.
- Implementing before writing tests.
- Pushing directly to `main` or skipping review.
- Creating redundant standalone summary/report documents.
- Introducing CGO dependencies.

## Where to go next

- [`AGENT.md`](https://github.com/talkincode/toughradius/blob/main/AGENT.md) —
  the full, canonical agent development guide.
- [Documentation Map](./documentation-map.md) — find the README, security
  policy, feature checklist, roadmap, and RFC index.
- [Protocol & RFC Reference](./rfc-index.md) — protocol standards mapped to the
  code and milestones.
