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

### Repository issue/PR automation

Several GitHub Actions add lightweight triage around issues and pull requests.
They help maintainers scan the queue, but they do not replace reading the
original issue, PR diff, and CI output.

| Workflow | Trigger and result | Maintainer notes |
| --- | --- | --- |
| Stale | `.github/workflows/stale.yml` runs daily at `04:24 UTC` and by manual dispatch. After 60 days without activity it adds `stale`; after 14 more inactive days it closes the item. | Comment, push a commit, or remove `stale` to keep work open. Issues with `pinned`, `security`, `help wanted`, `agent-roadmap`, or `needs-human` are exempt; PRs with `pinned`, `security`, `agent-roadmap`, or `needs-human` are exempt; all milestones are exempt. |
| Labeler | `.github/workflows/labeler.yml` runs on `pull_request_target` and applies labels from `.github/labeler.yml` based on changed paths. | It labels `go`, `javascript`, `github_actions`, `dependencies`, and `doc`. The action reads the changed-file list and base-branch config; it does not check out or execute PR code. |
| Workflow Lint | `.github/workflows/workflow-lint.yml` runs on PRs that change `.github/workflows/**` or `.github/actionlint.yml` / `.github/actionlint.yaml`, on pushes to `main` touching those paths, and by manual dispatch. | It runs `actionlint -shellcheck=` so the gate covers GitHub Actions YAML, expressions, and action inputs without turning existing shell-style warnings into this gate's scope. It is static validation only; it does not exercise release tags, Docker publishing, Pages deployment, or secrets. |
| Greetings | `.github/workflows/greetings.yml` runs when a contributor opens their first issue or PR and posts onboarding guidance. | The comment is informational. It does not change review requirements or issue priority. |

<a id="release-docker-publish"></a>

### Release tag and Docker publish automation

Pushing a `v*` tag triggers two release workflows. Release operators need to
understand both outputs and their credential boundaries before tagging.

| Workflow | Trigger and output | Release prerequisites |
| --- | --- | --- |
| Release publish | `.github/workflows/release-publish.yml` runs on `v*` tag pushes, builds multi-platform binaries, generates checksums, and creates the GitHub Release. | The tag must point at the verified `origin/main` target SHA; release notes are generated by the workflow from the tag and closed issues. |
| Docker publish | `.github/workflows/docker-publish.yml` runs on `v*` tag pushes, publishes `talkincode/toughradius:latest` and the version tag to Docker Hub, then independently attempts `ghcr.io/<owner>/toughradius:latest` and the version tag. | Docker Hub requires `DOCKERHUB_USERNAME` / `DOCKERHUB_TOKEN`. GHCR requires package repository access / inherited access that allows this repository's `GITHUB_TOKEN` to write, or `PKG_GITHUB_TOKEN` with `write:packages`; `PKG_GITHUB_USERNAME` is optional when the token owner differs from the tag actor. |

Docker Hub publishing is the required gate. GHCR publishing depends on external
GitHub Packages access settings, so the workflow isolates the GHCR push in its
own step and runs a non-destructive write-access preflight probe first: when
the credentials cannot write the package (the `permission_denied:
write_package` signature), the GHCR build is skipped instead of failing
mid-push, and the run summary reports a warning and recovery guidance without
making maintainers guess whether Docker Hub published.

Before tagging:

- Confirm the target `origin/main` CI is green and the tag version is unused.
- Confirm the Docker Hub secrets exist and still have write access.
- Confirm the GHCR package `talkincode/toughradius` inherits this repository's
  access, or configure `PKG_GITHUB_TOKEN` (optionally `PKG_GITHUB_USERNAME`);
  docs record only secret names and permission requirements, never token values.
- After tagging, check the GitHub Release, Docker Hub `latest` / version tag,
  GHCR `latest` / version tag, and both workflow conclusions.

Failure recovery:

- If the GitHub Release and Docker Hub succeed but GHCR fails, fix GHCR package
  access or token permissions, then rerun the tag workflow; do not create a
  duplicate bad version tag for the same source.
- If Docker Hub publishing fails, treat the release as incomplete and rerun the
  workflow after fixing the secret or registry problem.
- If a new patch tag is needed for validation, rerun the `release-version` SOP
  first so the version decision is not skipped.

<a id="report-pr-automation"></a>

### Report PR automation

Weekly report workflows publish generated reports through signed commits on
short-lived automation branches. The workflows always upload their generated
artifact before the report PR step, so maintainers can inspect the report even
when PR creation is skipped or blocked.

| Workflow | Trigger and published files | PR requirements and fallback |
| --- | --- | --- |
| EAP acceptance weekly | `.github/workflows/eap-acceptance-weekly.yml` runs every Monday at `09:17 UTC` and by manual dispatch. It publishes `docs/reports/eap/<date>.md`, `docs/reports/eap/latest.md`, and the EAP report index pages. | When the external EAP acceptance step succeeds and report files changed, the PR step requires the repository secret `EAP_REPORT_SIGNING_KEY` and repository variables `EAP_REPORT_SIGNING_EMAIL` and optional `EAP_REPORT_SIGNING_NAME`. Missing key or email configuration blocks the PR step because protected `main` requires verified signed commits; use the uploaded `eap-acceptance-<run_id>` artifact for the generated report. |
| Performance benchmark weekly | `.github/workflows/performance-benchmark-weekly.yml` runs every Monday at `10:37 UTC` and by manual dispatch. It publishes `docs/reports/performance/<date>.md`, `docs/reports/performance/latest.md`, and the performance report index pages. | When benchmarks complete and report files changed, the PR step uses `PERFORMANCE_REPORT_SIGNING_KEY`, `PERFORMANCE_REPORT_SIGNING_EMAIL`, and optional `PERFORMANCE_REPORT_SIGNING_NAME`. If these are absent, it reuses `EAP_REPORT_SIGNING_KEY`, `EAP_REPORT_SIGNING_EMAIL`, and optional `EAP_REPORT_SIGNING_NAME`. If no signing key/email pair is configured, the workflow keeps the generated report in the `performance-benchmark-<run_id>` artifact, writes a workflow step summary explaining that PR creation was skipped, and exits the PR step successfully. |

The signing key values are never documented here or in issue/PR comments. Only
record the secret and variable names, their purpose, and whether they are
configured.

When reviewing a generated report run:

- Check the workflow summary first for skip or credential messages.
- Download the run artifact if no report PR was opened.
- If a report PR exists, verify that its head commit is shown as verified by
  GitHub before merging.
- Check the published file list in the PR body against the workflow's expected
  report paths.
- Treat report PR CI job state and `mergeStateStatus=UNSTABLE` behavior as the
  separate implementation boundary tracked by issue #494.

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
- `Workflow Lint` — required when `.github/workflows/**` or
  `.github/actionlint.yml` / `.github/actionlint.yaml` changes; it runs `actionlint -shellcheck=` for
  GitHub Actions YAML, expression, and action input validation.
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
