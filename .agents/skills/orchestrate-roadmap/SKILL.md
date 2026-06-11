---
name: orchestrate-roadmap
description: Global orchestrator / dispatcher role. Activated when the user issues an "auto-delegate development / continue the roadmap / auto-delegate" command (without naming a specific subtask); runs the full loop - select task, pick SOP, dispatch, quality gates, open PR, self-iterate the roadmap.
---

# Skill: Roadmap Orchestration & Auto-Delegation (Orchestrator)

> Scope: all milestones / `TR-F022` | Role: **orchestrator** (the other skills are execution SOPs; this skill coordinates them).

## When to use
When the user issues an **auto-delegate development** style command without naming a specific subtask, this skill is the entry point:
- "start auto development / auto-delegate development / push the roadmap forward / continue development / do M1 / auto-delegate / orchestrate".
- A periodic "advance one development round".

This skill does not write implementation details itself; it selects work from the roadmap, picks the matching execution skill, enforces quality gates, and drives plan self-iteration after delivery.

## Orchestration loop (each round)
1. **Sync context**: first read `AGENT.md`, `.agents/README.md` (shared pre-conditions + guardrails), `docs/roadmap.md`, `docs/feature-checklist.md`; confirm toolchain versions and the non-goals (`TR-N001`-`TR-N005`).
2. **Clear in-flight PRs first** (per `../review-pr/SKILL.md`, before selecting anything new): drain open **internal** `agent-roadmap` PRs oldest-first - address every `needs-rework` review comment by re-running the matching execution SOP and re-review; merge any PR that is `agent-approved` **and** CI-green; skip `needs-human`. **Skip fork / cross-repository PRs** (`gh pr view <n> --json isCrossRepository` -> `true`): they are a different trust domain, are never auto-handled here, and are left `needs-human` for a maintainer's personal review (see `../review-pr/SKILL.md` "External / fork PRs are out of scope"). Only when no internal `agent-roadmap` PR is left unreviewed or in `needs-rework` may you select a new task. This prevents re-picking a task whose PR is still open (which would create a conflicting duplicate).
3. **Select task**:
   - Rule: `grep -nE '^- \[ \] M[0-9]+\.[0-9]+' docs/roadmap.md | head -1` to take the first unchecked subtask top-down.
   - Priority: `M1 -> M2 -> M3`; only fall back to P2/P3 when P1 milestones have no actionable subtask (see the priority column in the milestone overview).
   - Anchor the subtask's `TR-F` ID; if it hits a non-goal `TR-N*`, stop immediately and report back - never expand scope on your own.
4. **Pick SOP**: match the task type to `.agents/skills/<name>/SKILL.md`:
   - vendor VSA -> `add-radius-vendor`; EAP method -> `add-eap-method`; Admin API -> `add-adminapi-endpoint`; frontend resource -> `add-react-admin-resource`; config item -> `add-config-schema`; upstream library -> `sync-upstream-radius`; protocol spec -> `reference-rfc`; tests -> `add-acceptance-test` / `write-go-tests`; requirement alignment -> `align-feature-checklist`.
   - If no SOP matches, first use `align-feature-checklist` to fix the scope and IDs before continuing.
5. **Dispatch & execute**:
   - Multi-agent environment: dispatch a sub-agent for the subtask with **full context** (inject the selected SKILL.md + related RFCs + acceptance criteria); deliver exactly one minimal closed loop.
   - Single-agent environment: execute the selected SOP yourself, in order.
   - By default each round claims **one** unchecked subtask only (minimal closed loop, revertible).
6. **Quality gates**: `go build ./...`, `go test ./...`, `golangci-lint run` (v2.12.2) must pass; for frontend changes run `cd web && npm run build`; protocol / end-to-end changes must ship a CI-executable acceptance test (`test/integration/`).
7. **Open PR**: never push to `main` directly; the PR description references the milestone ID + `TR-F` + related RFC + acceptance test; tag it with the `agent-roadmap` label so the review gate can track it.
8. **Review gate**: hand the PR to `../review-pr/SKILL.md` for an independent, CI-anchored review. If it lands `agent-approved` with green CI, that skill auto-merges (squash); if it lands `needs-rework`, the PR stays open and is drained at the **next round's step 2** - do not merge past a `needs-rework`/red-CI verdict, and do not check the subtask off until it is merged.
9. **Self-iterate the roadmap**: per `../groom-roadmap/SKILL.md`, **after the PR is merged**, check off the delivered subtask, update milestone status, backfill follow-up subtasks / split / re-prioritize, and fold newly discovered needs in via `align-feature-checklist`.
10. **Loop or stop**:
   - By default stop after one closed loop; report the round's result and the next pending subtask.
   - When the user asks to "keep going", return to step 2 (clear in-flight first, then select) and loop; on a blocker (missing spec / external dependency / decision needed), stop at a safe checkpoint and mark the reason as `blocked` in the roadmap.

## Boundaries
- The orchestrator never bypasses any guardrail: the task-selection rule, TR-N non-goals, PR-only, and quality gates are all mandatory.
- The autonomous domain is **internal** work only: roadmap subtasks delivered on this repo's own `copilot/*` branches. Fork / cross-repository PRs from external contributors are out of scope - never review-to-merge them autonomously; route them to a maintainer (`needs-human`), backed by the `external-pr-gate` required check.
- One round advances only one minimal closed loop; do not pack multiple subtasks into one PR to "do more".
- Do not expand scope: any direction beyond the `TR-F` checklist goes through `align-feature-checklist` first.
- It only coordinates and orchestrates; it does not replace the specifics of each execution SOP (vendor unit conversion, EAP fragmentation, etc. still follow their own SKILL.md).

## Acceptance
- [ ] In-flight `agent-roadmap` PRs were drained (rework addressed / approved-and-green merged) before a new task was selected
- [ ] The round's task is the first unchecked subtask top-down in the roadmap, anchored to a `TR-F`
- [ ] A matching execution SOP was used; no `TR-N` was touched
- [ ] Quality gates and (where required) CI acceptance tests pass
- [ ] Output goes through a PR (labeled `agent-roadmap`) referencing the milestone / ID / RFC, gated by `review-pr` and merged only when `agent-approved` + CI-green
- [ ] After the PR is merged, the roadmap and plan were iterated per `groom-roadmap`
