---
name: align-feature-checklist
description: Align requirements with and update the feature checklist docs/feature-checklist.md. Use before developing any new requirement/issue, or when a requirement cannot be mapped to an existing TR-F ID.
---

# Skill: Align Requirements / Update the Feature Checklist

> Scope: all feature IDs | Baseline doc: `docs/feature-checklist.md` (English version `feature-checklist.en.md`)

## When to use
- Before developing any new requirement / issue.
- When a requirement cannot be mapped to an existing `TR-F` ID.

## Steps
1. **Mapping check**: find the matching `TR-F0xx` in `docs/feature-checklist.md`. If it maps, anchor it and proceed to implementation.
2. **Non-goal check**: confirm the requirement is not in `TR-N001`-`TR-N005` (payment/orders, CRM/tickets, generic monitoring, multi-tenant SaaS, rewriting the protocol stack/framework). If it hits a non-goal, reject it or raise a scope decision first.
3. **Cannot map**: first submit a PR that **only modifies the feature checklist**:
   - Add the feature ID, category, scope description, status, related code paths, and acceptance/dev boundary.
   - Sync the English version `feature-checklist.en.md`.
4. **Scheduling**: if needed, add a milestone or subtask in `docs/roadmap.md`.
5. **Implementation**: only enter code implementation after the checklist is merged; the PR references the corresponding ID.

## Conventions
- The checklist describes capability boundaries only; interfaces / structs / algorithms are governed by code, tests, and inline comments.
- Deliver only one minimal closed loop at a time.

## Acceptance
- [ ] The requirement is anchored to / added as a `TR-F` ID
- [ ] No non-goal direction was touched
- [ ] CN and EN checklists are in sync
