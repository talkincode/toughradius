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
- **EAP / 802.1X suite** — challenge methods (EAP-MD5, EAP-MSCHAPv2) and tunneled
  methods: EAP-TLS (RFC 5216, certificate-based), PEAPv0/EAP-MSCHAPv2 (Windows / AD
  compatibility), and EAP-TTLS (RFC 5281, inner PAP / MS-CHAPv2). MS-CHAPv2-based
  methods are compatibility-oriented and carry an NTLMv1-like attack surface; prefer
  EAP-TLS where you control client certificates. TLS 1.3 (RFC 9190), TEAP, and
  EAP-PWD are tracked on the roadmap.
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

- [Concepts & Terminology](./concepts.md) — AAA vocabulary, the request flow,
  and how each concept maps to the product.
- [Quick Start](./quickstart.md) — install, initialize, create a user, and
  verify with `radtest` in about ten minutes.
- [Vendor Integration Guide](./vendor-guide.md) — case studies for MikroTik,
  Huawei, Cisco, H3C, ZTE, iKuai, and standard devices.
- [Admin UI Manual](./admin-manual.md) — every page of the management console.
- [Operations Guide](./ops-guide.md) — production configuration, certificates,
  monitoring, backup, and CLI tools.
- [FAQ](./faq.md) — answers to the questions everyone asks.
- [Documentation Map](./documentation-map.md) — find the existing README, agent
  guide, security policy, feature checklist, roadmap, and RFC index.
- [mdbook & GitBook Coexistence](./gitbook-coexistence.md) — how this handbook
  relates to the GitBook site and the single-source-of-truth policy.
