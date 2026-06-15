# Documentation Map

> 中文版本：[文档地图](../zh/documentation-map.md)

This handbook is being assembled incrementally. The first table lists the
handbook's own chapters; the second points to documents that still live
elsewhere in the repository, so everything is reachable from a single place.

## Handbook chapters

| Chapter | Contents |
| ------- | -------- |
| [Overview](./overview.md) | Project introduction, core capabilities, service model |
| [Concepts & Terminology](./concepts.md) | AAA vocabulary, the authentication flow, password protocols |
| [Quick Start](./quickstart.md) | Install, initialize, first user, debugging |
| [Vendor Integration Guide](./vendor-guide.md) | Case studies: MikroTik, Huawei, Cisco, H3C, ZTE, iKuai, … |
| [Portal / Hotspot Integration Boundary](./portal-hotspot-boundary.md) | Hard boundary: ToughRADIUS is the RADIUS AAA backend, not a hosted captive portal product |
| [Scenario Cookbook](./cookbook.md) | End-to-end ops scenarios (five-part): MikroTik PPPoE / Hotspot / CoA; Huawei BRAS speed tiers / line binding / CoA; H3C / ZTE / iKuai / Cisco per-vendor diffs |
| [Admin UI Manual](./admin-manual.md) | The management console, page by page |
| [Operations Guide](./ops-guide.md) | Configuration reference, certificates, monitoring, backup, CLI tools |
| [LDAP / AD Authentication](./auth-ldap.md) | Verify passwords against an LDAP/AD directory (PAP-family only): bind modes, TLS, security, metrics |
| [FAQ](./faq.md) | Frequently asked questions by theme |
| [Protocol & RFC Reference](./rfc-index.md) | Protocol standards mapped to the code |
| [Security Policy](./security-policy.md) | Security advisories and update guidance |
| [Agent Development Guide](./agent-guide.md) | Contributor digest of the AI-agent workflow, quality gates, and auto-delegation loop |

## Repository documents

| Document            | Description                                            | Current location |
| ------------------- | ----------------------------------------------------- | ---------------- |
| README              | Project introduction, features, and quick start       | [README.md](https://github.com/talkincode/toughradius/blob/main/README.md) |
| Agent guide         | AI-agent development guide and working rules           | [Agent Development Guide](./agent-guide.md) (handbook digest) · [AGENT.md](https://github.com/talkincode/toughradius/blob/main/AGENT.md) (canonical) |
| Security policy     | Security advisories and update guidance                | [Security Policy](./security-policy.md) (canonical) · [SECURITY.md](https://github.com/talkincode/toughradius/blob/main/SECURITY.md) (pointer) |
| Feature checklist   | Feature scope baseline (`TR-F` IDs)                    | [docs/feature-checklist.md](https://github.com/talkincode/toughradius/blob/main/docs/feature-checklist.md) · [English](https://github.com/talkincode/toughradius/blob/main/docs/feature-checklist.en.md) |
| Roadmap             | Long-term roadmap and milestones                       | [docs/roadmap.md](https://github.com/talkincode/toughradius/blob/main/docs/roadmap.md) |
| RFC index           | Protocol standards index used by the project           | [Protocol & RFC Reference](./rfc-index.md) (canonical) · [docs/rfcs/README.md](https://github.com/talkincode/toughradius/blob/main/docs/rfcs/README.md) (raw catalog) |

> **Migration plan.** The handbook now covers the README's user-facing content
> (overview, quick start, vendor integration, admin manual, operations, FAQ);
> the README remains the GitHub landing page and links here. The agent guide now
> has a handbook [digest chapter](./agent-guide.md) that points to the canonical
> [`AGENT.md`](https://github.com/talkincode/toughradius/blob/main/AGENT.md),
> which stays in the repository root because the agent tooling references it
> directly. The feature checklist and roadmap are **living documents** maintained
> in `docs/` by dedicated workflows and are linked rather than migrated.
