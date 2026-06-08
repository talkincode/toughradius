---
name: add-eap-method
description: Add an EAP authentication method (e.g. EAP-TLS) within the existing EAP handler system (TR-F004). Use when the change involves EAP handshake, fragmentation, state management, or failure semantics.
---

# Skill: Add an EAP Authentication Method

> Feature ID: `TR-F004` | Milestone: M1 (EAP-TLS)

## When to use
When adding a new EAP method (e.g. EAP-TLS) within the existing EAP system.

## Pre-research
```text
view internal/radiusd/plugins/eap/coordinator.go         # coordinator, do NOT rewrite
view internal/radiusd/plugins/eap/interfaces.go          # handler interface
file_search "internal/radiusd/plugins/eap/handlers/*_handler.go"
view internal/radiusd/plugins/eap/statemanager/          # EAP state management
grep_search "EapMethod" --include internal/app/**         # enabled-list config
```
Existing references: `md5_handler.go`, `mschapv2_handler.go`, `otp_handler.go`.

## Protocol specs & reference implementations

**International standards (read repo-local `docs/rfcs/` first):**
- `rfc3748-eap.txt` - EAP framework
- `rfc5216-eap-tls.txt` - EAP-TLS (handshake, fragmentation, identity)
- `rfc3579-radius-eap-support.txt` - RADIUS carrying EAP (EAP-Message / Message-Authenticator)
- `rfc5247-eap-key-management.txt` - EAP key management
- `rfc7499-packet-fragmentation.txt` - RADIUS fragmentation
- Others: `rfc5281-eap-ttls.txt`, `rfc7170-teap.txt` (if extending tunneled methods)

Backfill missing specs per `../reference-rfc/SKILL.md`.

**Reference implementations (for ideas only; mind licensing and protocol compatibility, do not copy incompatible code):**
- BeryJu `radius-eap`: <https://github.com/BeryJu/radius-eap>
- Implementation notes: <https://beryju.io/blog/2025-05-implementing-eap/>

## Implementation steps
1. **Handler skeleton**: implement the handler interface at `internal/radiusd/plugins/eap/handlers/<method>_handler.go` (mirror mschapv2).
2. **State management**: reuse `statemanager`; correlate multi-round handshakes / fragmentation (TLS) via the EAP State, and do not add branches inside the coordinator.
3. **Registration**: wire into the coordinator and the enabled list (`EapMethod` config) the same way existing handlers do.
4. **Failure semantics**: return an explicit reason on failure, convert it to `AuthError`, and emit metrics (reference `internal/radiusd/errors` and `radius_metrics.go`).
5. **Config schema**: if new config is needed (e.g. certificate paths), see `../add-config-schema/SKILL.md`.

## Boundaries
- Do not rewrite the EAP coordinator (`coordinator.go`).
- For EAP-TLS, deliver a minimal working auth path first; certificate revocation / policy come in later milestones.
- `eap-otp` currently uses a fixed sample OTP; do not copy its fixed value when implementing a real method.

## Acceptance
- [ ] Handler unit tests + end-to-end auth tests pass
- [ ] Failure cases have explicit reject reasons and metrics
- [ ] `go test ./internal/radiusd/...` and `golangci-lint run` pass
- [ ] An end-to-end acceptance test is added under `test/integration/` (see `../add-acceptance-test/SKILL.md`), executed by CI
- [ ] PR references `TR-F004` and the M1 subtask ID, and cites the RFC clauses relied upon
