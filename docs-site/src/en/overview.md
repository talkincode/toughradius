# Overview

> 中文版本：[概述](../zh/overview.md)

ToughRADIUS is a powerful, open-source RADIUS server written in Go, designed for
ISPs, enterprise networks, and carriers. It implements the standard RADIUS
protocols together with RadSec (RADIUS over TLS) and ships with a modern React
Admin web management interface.

## Core capabilities

- **Standard RADIUS** — full RFC 2865 (authentication) and RFC 2866 (accounting)
  support.
- **RadSec** — TLS-encrypted RADIUS over TCP (RFC 6614).
- **Dynamic authorization** — CoA and Disconnect messages (RFC 5176).
- **EAP suite** — EAP-TLS, PEAPv0/EAP-MSCHAPv2, and EAP-TTLS inner methods, with
  more methods tracked on the roadmap.
- **Multi-vendor support** — compatible with Cisco, MikroTik, Huawei, and other
  major network devices through vendor-specific attributes (VSAs).
- **Modern management UI** — a React Admin dashboard for users, profiles, online
  sessions, accounting, and audit logs.
- **Multi-database** — PostgreSQL (default) and SQLite (pure Go, no CGO).

## Service model

The server runs several independent services concurrently; if any one fails, the
process exits so a supervisor can restart it cleanly.

| Service          | Protocol / Port      | Purpose                              |
| ---------------- | -------------------- | ------------------------------------ |
| Web / Admin API  | HTTP, TCP `1816`     | Management UI and REST API           |
| RADIUS Auth      | UDP `1812`           | Authentication                       |
| RADIUS Acct      | UDP `1813`           | Accounting                           |
| RadSec           | TLS over TCP `2083`  | Encrypted RADIUS transport           |

## Where to go next

- [Documentation Map](./documentation-map.md) — find the existing README, agent
  guide, security policy, feature checklist, roadmap, and RFC index.
- [mdbook & GitBook Coexistence](./gitbook-coexistence.md) — how this handbook
  relates to the GitBook site and the single-source-of-truth policy.
