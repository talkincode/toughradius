---
name: groom-roadmap
description: Roadmap and development-plan self-iteration (grooming). After each subtask delivery, or during periodic maintenance, check off completed items, update milestone status, backfill / split / re-prioritize subtasks, and fold newly discovered needs in via the feature checklist to keep the plan self-consistent and aligned with the checklist.
---

# Skill: Roadmap / Plan Self-Iteration (Grooming)

> Scope: `TR-F022` | Usually invoked by `../orchestrate-roadmap/SKILL.md` after delivery; can also run standalone.

## When to use
- After a milestone subtask is delivered (PR merged or ready).
- During periodic roadmap maintenance, when a subtask is too large / outdated / already implicitly done.
- When implementation surfaces a new gap, new dependency, or a change in requirement boundary.

## Steps
1. **Check off delivery**: change the delivered subtask `- [ ] M*.*` to `- [x]`; sync the milestone overview table and the milestone section status (`planned -> in progress -> delivered`; delivery means merged to `main` + CI passing).
2. **Backfill learnings**: based on the round's PR result / blockers,
   - split follow-up work discovered during implementation into new `- [ ] M*.*` subtasks (keep each independently deliverable, revertible, verifiable);
   - split oversized subtasks; annotate removals with the reason;
   - mark items with an external dependency / pending decision as `blocked` with the reason.
3. **Re-prioritize**: keep `M1 -> M2 -> M3` by default; adjust only when dependencies / blockers / new delivery evidence change, and explain the reason in the PR.
4. **Align scope**: route newly discovered needs through `../align-feature-checklist/SKILL.md` - schedule only what maps to a `TR-F`; reject anything hitting `TR-N001`-`TR-N005`; never silently expand the roadmap.
5. **Consistency check**: ensure the milestone overview status <-> subtask checkboxes <-> `docs/feature-checklist.md` status are consistent; keep the CN/EN checklists in sync; no dangling references to deleted files / skills.
6. **Commit**: plan-doc changes go through a PR - either folded into the delivery PR or as a small follow-up PR; reference the milestone ID.

## Boundaries
- Only edit **plan / scope docs** (`docs/roadmap.md`, `docs/feature-checklist.md` and its English version, the in-session plan); do not change product code in this skill.
- Do not invent tasks: every new subtask must anchor to a `TR-F` and be a minimal closed loop.
- Do not relax guardrails: non-goals, PR-only, and quality-gate requirements stay unchanged.

## Acceptance
- [ ] Delivered subtasks are checked off; the milestone status table is in sync
- [ ] Added / split / blocked subtasks all anchor to a `TR-F` and are independently verifiable
- [ ] New needs are aligned via the feature checklist (CN/EN in sync); no `TR-N` touched
- [ ] Roadmap <-> feature checklist status are consistent; no dangling references
