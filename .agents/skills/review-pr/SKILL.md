---
name: review-pr
description: Independent review gate for auto-delegation PRs (TR-F022). Use after an orchestration round opens a PR, and as the first step of every new round to clear in-flight PRs - run an adversarial, CI-anchored review, request changes via labels + comments, and auto-merge only when the review passes and CI is green.
---

# Skill: PR Review Gate & Auto-Merge (Reviewer)

> Scope: all milestones / `TR-F022` | Role: **review gate** between "open PR" and "merge to main". Invoked by `../orchestrate-roadmap/SKILL.md`: once right after a PR is opened, and again at the start of every new round to clear in-flight PRs.

## Why this exists
Auto-delegation only forms a closed loop when PRs actually merge. The roadmap checkbox advances **only on merge to `main`**, so an unmerged PR makes the next round re-pick the same task and open a conflicting duplicate. This skill is the gate that turns "open PR" into either "merged" or "explicitly needs rework", so each round starts from a clean state.

## Identity constraint (read first)
GitHub **forbids a PR author from approving or requesting changes on their own PR** - the author may only post `COMMENT` reviews. Auto-delegation runs under a single identity (it both opens the PR and runs this skill), so the formal `APPROVE` / `REQUEST_CHANGES` states are unavailable.

Therefore this skill encodes its verdict with **labels + a COMMENT review**, not GitHub's formal review state:

| Intent | Mechanism |
| --- | --- |
| request changes | label `needs-rework` + a COMMENT review listing the blocking issues (file:line) |
| approve | label `agent-approved` + a COMMENT review stating the gate passed |
| give up to a human | label `needs-human` + a COMMENT review stating why |
| mark the PR as a roadmap round | label `agent-roadmap` (added at PR creation) |

> Upgrade path: if a **separate reviewer identity** (a second account / GitHub App token, distinct from the PR author) is provided, this skill MAY use real `gh pr review --approve` / `--request-changes` and gate auto-merge on GitHub's review state plus required checks. Default below assumes a single identity.

The gate has teeth only because it is **anchored to mechanical signals (CI) and run as an independent pass**, not because of the label name. Never approve on "looks fine" alone.

## Reviewing one PR
1. **Precondition - CI must be green.** `gh pr checks <n>`:
   - Any job failing for a **real** reason (compile/test/lint/integration) -> this is a blocking issue; go to step 3 (rework) and cite the failing job.
   - Failing only from transient infra (e.g. registry/network timeout in "Initialize containers") -> `gh run rerun <run-id> --failed` once; re-check.
   - Still pending -> do not approve this round; leave the PR untouched. A pending-CI PR still counts as **unreviewed**, so per the in-flight gate below the round does **not** select a new task while it is open - move on to the next PR in the drain list, then wait for CI / stop the round rather than starting new work.
2. **Independent adversarial read.** Review as a *separate pass* with fresh context - in a multi-agent environment dispatch a `code-review` sub-agent; otherwise consciously switch roles. Feed it only: `gh pr diff <n>`, the PR description, the linked milestone + `TR-F` ID, and the acceptance criteria. Apply a high signal bar - flag only issues that genuinely matter:
   - correctness / logic bugs, nil-deref, race, resource leaks;
   - security: leaked secrets, auth bypass, injection, plaintext where hashing is required;
   - scope: any change touching a non-goal `TR-N001`-`TR-N005`, or beyond the claimed `TR-F`;
   - protocol changes without a `docs/rfcs/` citation, or that contradict the cited clause;
   - missing/insufficient tests for the changed behavior; for protocol / E2E changes, a missing `test/integration/` acceptance test;
   - pushed to `main` directly, generated artifacts or secrets committed.
   - Do **not** comment on style, formatting, or naming taste.
3. **Record the verdict.**
   - **Blocking issues found** -> `gh pr comment`/`gh pr review --comment` enumerating each issue with `file:line` and a concrete fix; `gh pr edit <n> --add-label needs-rework --remove-label agent-approved`. Do not merge.
   - **No blocking issues AND CI green** -> post a COMMENT review summarizing what was checked; `gh pr edit <n> --add-label agent-approved --remove-label needs-rework`.
4. **Merge gate.** Only when the PR carries `agent-approved`, has **no** `needs-rework`/`needs-human`, and `gh pr checks <n>` re-confirms all green:
   `gh pr merge <n> --squash --delete-branch`.
   After merge, the round's `groom-roadmap` checks the subtask off on `main`.

## Clearing in-flight PRs (run first, every new round)
Before `orchestrate-roadmap` selects any new task, drain the queue - oldest first:
```
gh pr list --state open --label agent-roadmap --json number,labels,headRefName
```
- `needs-rework` -> read the review thread, **address every blocking comment** by re-running the original execution SOP on that branch, push, then re-run "Reviewing one PR" from step 1. This is mandatory work, not optional.
- `agent-approved` + green -> merge (gate above).
- `needs-human` -> skip; leave it for a human.
- pending CI / no verdict yet -> run "Reviewing one PR".

Only when no `agent-roadmap` PR is left in `needs-rework`/unreviewed state may the round proceed to pick a new roadmap subtask. This back-pressure is intentional: a stuck PR pauses new work instead of spawning duplicates.

## Bounded rounds (no infinite ping-pong)
Count prior `needs-rework` cycles on the PR (e.g. review comments by this skill). If a PR would enter its **4th** rework cycle, stop auto-handling it: `--add-label needs-human`, post a COMMENT explaining the unresolved issue, and leave it for a human. Never loop review->rework indefinitely.

## Bootstrap (idempotent)
Ensure the labels exist before first use:
```
gh label create agent-approved -c 0E8A16 -d "Review skill passed; eligible for auto-merge once CI is green" 2>/dev/null || true
gh label create needs-rework   -c D93F0B -d "Review skill found blocking issues; must be addressed before merge" 2>/dev/null || true
gh label create needs-human    -c B60205 -d "Exceeded auto-review rounds or blocked; needs a human decision" 2>/dev/null || true
gh label create agent-roadmap  -c 1D76DB -d "PR produced by an auto-delegation roadmap round" 2>/dev/null || true
```

## Boundaries
- Auto-merge **only** through this gate: never merge a PR that lacks `agent-approved` or whose CI is not green.
- The reviewer reads and judges; when it requests rework, the **execution SOP** does the fixing - this skill does not silently rewrite the feature to make its own review pass.
- Never relax the guardrails to approve: TR-N non-goals, PR-only, quality gates, and RFC citations remain hard requirements.
- Do not approve your own reasoning blindly - if CI is red or the diff touches a non-goal, the verdict is `needs-rework`/`needs-human`, regardless of how the code reads.

## Acceptance
- [ ] Every reviewed PR has a recorded COMMENT review and exactly one state label (`agent-approved` / `needs-rework` / `needs-human`)
- [ ] No PR is merged unless `agent-approved` **and** `gh pr checks` is fully green
- [ ] Each new round drains in-flight `agent-roadmap` PRs (rework addressed, approved-and-green merged) before selecting a new task
- [ ] Review rounds are bounded; a PR exceeding the cap is handed to a human via `needs-human`
- [ ] No `TR-N` was approved; protocol changes carry an RFC citation and an acceptance test
