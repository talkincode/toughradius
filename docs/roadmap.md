# ToughRADIUS Development Roadmap

Chinese version: [docs/roadmap.zh.md](roadmap.zh.md)

This roadmap is the long-term development plan for ToughRADIUS. It is bound to the feature checklist in [`docs/feature-checklist.en.md`](feature-checklist.en.md): every milestone must reference at least one `TR-F` feature ID, and scope that cannot be mapped to the checklist must update the checklist before it is scheduled here.

The Chinese roadmap keeps the detailed agent delivery log. This English roadmap is the default planning surface and task source: milestone status, guardrails, current execution order, and the next deliverable work. When roadmap scope or status changes, update both files in the same PR.

## Maintenance Rules

1. Milestones use `M<number>` IDs and map to one or more `TR-F` feature IDs.
2. Each milestone is split into MVP subtasks that are independently deliverable, reversible, and verifiable.
3. Status flows as `Planned -> In progress -> Delivered`; delivery means merged to `main` with passing CI.
4. Do not schedule non-goals from `TR-N001` through `TR-N006`: billing/orders, CRM/tickets, generic monitoring, multi-tenant SaaS, protocol-stack/framework rewrites, or hosted captive portal products.
5. Agent output must go through pull request, CI, and human review. Direct pushes to `main` are forbidden.
6. After each delivered subtask, use `.agents/skills/groom-roadmap/SKILL.md` to update status, split or reorder work, and keep the roadmap consistent with the checklist.
7. Delivered subtasks should keep only outcome, evidence, residual risk, and traceable entry points. Long implementation narratives belong in PRs, commits, or changelogs.
8. Automatic delegation picks the highest-priority unchecked subtask that is not marked `Blocked` or `waiting for evidence`.

## Status Definitions

| Status | Meaning |
| --- | --- |
| Planned | Scheduled but not started |
| In progress | Has active PR, issue, or partially delivered loop |
| Delivered | Merged to `main` with passing CI |
| Blocked | Waiting for external dependency, evidence, or decision |

## Milestone Overview

| Milestone | Theme | Related IDs | Priority | Status |
| --- | --- | --- | --- | --- |
| M1 | EAP-TLS authentication | TR-F004 | P1 | Delivered |
| M2 | CoA dynamic authorization | TR-F010 / TR-F012 / TR-F013 | P1 | Delivered |
| M3 | IPv6 capability closure | TR-F007 / TR-F011 / TR-F015 | P1 | Delivered |
| M4 | Agent development system and quality gates | TR-F022 / TR-F024 | P2 | Delivered |
| M5 | Vendor VSA coverage expansion | TR-F005 | P2 | In progress |
| M6 | Observability and operations improvements | TR-F015 | P3 | Planned |
| M7 | Upstream RADIUS library tracking and protocol compliance | TR-F021 / TR-F022 | P2 | In progress |
| M8 | PEAPv0 / EAP-MSCHAPv2 authentication | TR-F004 | P1 | Delivered |
| M9 | EAP-TTLS tunneled authentication | TR-F004 | P1 | Delivered |
| M10 | EAP-TLS 1.3 / RFC 9190 upgrade | TR-F004 | P2 | Planned |
| M11 | TEAP tunneled authentication | TR-F004 | P3 | Planned |
| M12 | EAP-PWD password authentication | TR-F004 | P3 | Planned |
| M13 | Bilingual documentation site with mdbook | TR-F023 | P2 | Delivered |
| M14 | LDAP / AD bind authentication backend for PAP-family methods | TR-F025 | P2 | In progress |

## Cross-Cutting Baseline

- Upstream RADIUS library tracking: ToughRADIUS uses `layeh.com/radius` through the `github.com/talkincode/radius` fork via `go.mod` `replace`. Important upstream fixes for security, protocol correctness, or attribute encoding must be evaluated for fork sync.
- Protocol references: protocol behavior changes must cite the relevant RFC. Check `docs/rfcs/` first; missing standards are added through `.agents/skills/reference-rfc/SKILL.md`.
- CI-backed acceptance: milestone acceptance must be backed by tests that run in CI. Protocol and end-to-end cases live under `test/integration/` with the `integration` build tag; pure logic belongs in `*_test.go`.

## Current Execution Queue

The scheduled **M5 vendor VSA expansion** batch (M5.1 inventory + M5.2/M5.3/M5.4 parsers/enhancers) is delivered — M5.4 shipped the Cisco `cisco-avpair` Access-Accept enhancer (#543) — and the remaining vendor enhancers stay demand-driven, so the next pickable subtask is **M7.1** (evaluate upstream `layeh.com/radius` fixes). M14.6 now has CI-backed OpenLDAP acceptance coverage; M14.5 remains blocked until load evidence justifies connection pooling / reconnect work.

| Order | Task | Status | Acceptance focus |
| --- | --- | --- | --- |
| 1 | M5 vendor VSA expansion | In progress | M5.1 inventory + M5.2/M5.3/M5.4 parsers/enhancers delivered (M5.4 Cisco `cisco-avpair` enhancer, #543); remaining vendor enhancers demand-driven |
| 2 | M7 upstream and RFC compliance tracking | In progress | Evaluate upstream fixes and add RFC-backed regression tests when behavior changes |
| 3 | M10 EAP-TLS 1.3 / RFC 9190 | Planned | Keep TLS 1.2 compatibility while adding TLS 1.3 key derivation and close-notify semantics |
| 4 | M14.5 LDAP connection robustness | Blocked: waiting for load evidence | Revisit pooling/reconnect design only when connection cost or cancellation evidence justifies the complexity |

Agent-facing unchecked tasks:

- [x] M14.6 LDAP integration acceptance tests: add CI-executable `test/integration/` coverage with a real OpenLDAP service container and seed LDIF. Delivered via `test/integration/ldap_test.go` plus CI / local compose OpenLDAP service wiring: plain PAP and `EAP-TTLS/PAP` authenticate through real LDAP bind while local `RadiusUser.Password` is deliberately wrong; wrong password rejects as `radus_reject_passwd_error`; directory unavailable and non-PAP CHAP reject as `radus_reject_ldap_error` with diagnostic reply text.
- [ ] M14.5 (Blocked: waiting for load evidence) LDAP connection robustness: revisit pooling, reconnect, and request-context propagation only when load evidence shows connection setup cost or cancellation behavior justifies the added complexity.
- [x] M5.1 Inventory pending vendor VSA gaps and dictionary differences. Delivered: `docs/vendor-vsa-gap-baseline.md` refreshed to HEAD `9882f79e` — registered parsers `default + huawei + h3c + zte + radback + alcatel + aruba + juniper`, response enhancers `default + huawei + h3c + zte + mikrotik + ikuai + aruba`, a corrected gap matrix, a delta-since-#433 section, and the next-batch backlog. (The first baseline #433 was superseded once M5.2/M5.3 landed; `#470` had re-opened this checkbox.)
- [x] M5.2 Add request-side vendor parsers for genuine MAC/VLAN request VSAs. Delivered: `radback` (#449), `alcatel` (#450), `aruba` (#451), and `juniper` (#453) request parsers, plus the `vendors.CodeAlcatel` / `CodeAruba` constants.
- [x] M5.3 Add the first vendor Access-Accept response enhancer. Delivered: `aruba` response enhancer registered in `plugins/init.go` (#456) with sample-based tests.
- [x] M5.4 Add a Cisco `cisco-avpair` Access-Accept response enhancer. Delivered (#543): `CiscoAcceptEnhancer` (`accept-cisco`) emits `Cisco-AVPair="ip:addr-pool=<pool>"` from `GetAddrPool` (appended, since Cisco-AVPair is multi-valued), registered in `plugins/init.go`, with sample-based tests (named / empty / `N/A` / vendor-mismatch / nil-safety). Rate limiting stays device-side (Cisco has no portable numeric-rate VSA); the standard `Framed-Pool` from `default_enhancer` is complemented, not replaced. Remaining enhancers (`alcatel` / `juniper` / `radback` / `microsoft` / `f5` / `hillstone` / `pfSense`) stay demand-driven.
- [ ] M7.1 Manually evaluate important upstream `layeh.com/radius` fixes and decide whether to sync the `talkincode/radius` fork and update the `go.mod` replacement.
- [ ] M10.1 Add TLS 1.3 handshake negotiation and TLS 1.2 fallback for EAP-TLS.

## Non-Goals

- Do not build a billing, order, or finance system.
- Do not turn the admin UI into CRM, ticketing, or a customer self-service portal.
- Do not turn the dashboard into a generic monitoring platform.
- Do not add multi-tenant SaaS semantics without a prior scope decision and migration design.
- Do not rewrite the RADIUS protocol stack or replace the management framework without a specific defect and migration plan.
- Do not build, host, or operate captive portal login pages, voucher/SMS/WeChat/payment onboarding, or vendor portal-server state machines. Portal belongs to another product; ToughRADIUS only integrates as the RADIUS auth/accounting backend.

## Completion Standard

A milestone is complete only when the user-facing or operator-facing capability works end to end, failure modes are diagnosable, relevant protocol behavior cites RFCs, and regressions are guarded by CI-executable tests. Documentation and checklist status must match the implemented capability.
