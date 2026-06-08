---
name: reference-rfc
description: Look up and cite international protocol standards (RFC/IEEE) (TR-F021). Any change to RADIUS/EAP/accounting/CoA/IPv6 protocol behavior should first check docs/rfcs and cite the clause.
---

# Skill: Look Up & Cite International Protocol Standards

> Feature ID: `TR-F021` (protocol material) | Scope: all protocol-related changes

## Principle
Any change to RADIUS / EAP / accounting / CoA / IPv6 protocol behavior **must cite the corresponding international standard clause** (RFC / IEEE, etc.), preferring material already in the repo.

## Check the local library first
The repo's `docs/rfcs/` already holds 50+ RFCs (see `docs/rfcs/README.md`); common ones:

| Topic | File |
| --- | --- |
| RADIUS auth / accounting | `rfc2865-radius-authentication.txt` / `rfc2866-radius-accounting.txt` |
| RADIUS extensions / protocol extensions | `rfc2869-radius-extensions.txt` / `rfc6929-protocol-extensions.txt` |
| RADIUS carrying EAP | `rfc3579-radius-eap-support.txt` |
| EAP framework / EAP-TLS / TTLS / TEAP | `rfc3748-eap.txt` / `rfc5216-eap-tls.txt` / `rfc5281-eap-ttls.txt` / `rfc7170-teap.txt` |
| Fragmentation | `rfc7499-packet-fragmentation.txt` |
| Dynamic authorization / CoA | `rfc3576-dynamic-authorization.txt` / `rfc5176-coa-disconnect.txt` |
| IPv6 / prefix delegation | `rfc3162-radius-ipv6.txt` / `rfc4818-ipv6-prefix-delegation.txt` / `rfc6911-ipv6-access-networks.txt` |
| RadSec / RADIUS over TCP | `rfc6614-radsec.txt` / `rfc6613-radius-over-tcp.txt` |
| Status-Server / implementation issues | `rfc5997-status-server.txt` / `rfc5080-implementation-issues.txt` |

Search:
```text
grep -rni "Message-Authenticator" docs/rfcs/
view docs/rfcs/rfc5216-eap-tls.txt
```

## Backfill when missing locally
1. Confirm it is truly missing (`ls docs/rfcs/ | grep <number>`).
2. Fetch the spec text from an authoritative source (IETF datatracker / rfc-editor).
3. Save it per `docs/rfcs/FILE-NAMING.md` (e.g. `rfcXXXX-<slug>.txt`).
4. Update the `docs/rfcs/README.md` index, noting the implemented or to-be-implemented feature ID.

## How to cite
- Cite the clause in inline code comments explaining "why this implementation" (e.g. `// per RFC 5216 section 2.1.5, fragment via EAP-TLS Length+More-Fragments`).
- List the RFC numbers and sections relied upon in the PR description.

## Boundaries
- Spec material does not replace tests; new material must state the implemented or to-be-implemented feature ID (see the `TR-F021` boundary).
- Do not upload copyright-restricted, non-redistributable, non-standard documents.

## Acceptance
- [ ] Protocol changes cite specific RFC clauses
- [ ] New specs are saved per the naming convention and registered in the README
- [ ] Behavioral differences carry inline comments stating the basis
