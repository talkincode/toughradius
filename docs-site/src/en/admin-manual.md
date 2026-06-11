# Admin UI Manual

> 中文版本：[管理系统用户手册](../zh/admin-manual.md)

The management console runs on port `1816` (HTTP) and is built with React
Admin. It is bilingual (中文 default, English via the app-bar language menu) and
supports light/dark themes. This chapter walks through every page.

## Logging in and accounts

Sign in at `http://<server>:1816` with an operator account. The initial
administrator is `admin` / `toughradius` — change it immediately under
**Account Settings** (top-right avatar), which also edits your profile info.
Passwords must be at least 6 characters.

Operator roles (set under **Operators**):

| Level | Meaning |
| ----- | ------- |
| `super` | Full access, including System Config and Operators |
| `admin` | Same menu visibility as super |
| `operator` | Day-to-day pages only — System Config and Operators are hidden |

## Dashboard

The landing page summarizes the deployment:

- **Stat cards** — total users (with disabled/expired counts), online users,
  today's authentication and accounting request counts, today's upload/download
  traffic in GB.
- **IPv6 strip** — online sessions with IPv6, IPv6 address/prefix/delegated
  prefix counts, users with static IPv6 configuration.
- **Charts** — authentication trend, profile distribution pie, and a 24-hour
  upload/download traffic chart.

Counters are driven by in-memory metrics; see the
[Operations Guide](./ops-guide.md#metrics).

## Network Nodes

Logical groups (name, tags, remark) used to organize NAS devices. Create nodes
first — each NAS belongs to one.

## NAS Devices

Registry of network devices allowed to talk RADIUS to this server. Requests
from unknown source addresses are dropped.

| Field | Notes |
| ----- | ----- |
| Name / Identifier | Free-form name; `identifier` matches the RADIUS `NAS-Identifier` |
| IP address / Hostname | The source address of RADIUS packets |
| Vendor | **Decides vendor attribute behavior** — Standard, Cisco, Huawei, MikroTik, H3C, ZTE, iKuai (see the [Vendor Integration Guide](./vendor-guide.md)) |
| Secret | RADIUS shared secret |
| CoA port | Where CoA/Disconnect are sent; default `3799` |
| Node / Status / Tags / Remark | Organization and lifecycle |

## RADIUS Users

Subscriber accounts. List supports filtering by username, real name, email,
mobile, and IP, plus CSV export and column sorting.

Key form fields: username/password, status (enabled/disabled), **billing
profile** (required — rates, concurrency, pools inherit from it), **expire
time** (drives `Session-Timeout`), static `ip_addr` / `ipv6_addr`, IPv6 prefix
pools and delegated prefixes (edit view), contact info, remark.

**Batch import**: the list toolbar's import button accepts `.xlsx`, `.csv`, or
`.json` files and reports created/failed counts. Export the current list as a
template reference.

The show view groups everything including IPv6 details, MAC/VLAN binding
values, and timestamps.

## Billing Profiles

Reusable authorization templates:

| Field | Meaning |
| ----- | ------- |
| `active_num` | Max concurrent sessions per user (0 = unlimited) |
| `up_rate` / `down_rate` | Bandwidth in **Kbps**; converted per vendor (list shows Mbps for values ≥ 1024) |
| `addr_pool` | `Framed-Pool` name the NAS allocates from |
| `ipv6_prefix` / domain | IPv6 and Huawei domain authorization |
| `bind_mac` / `bind_vlan` | Lock users to first-seen MAC / VLANs |

## Online Sessions

Live sessions (`radius_online`). Columns include session ID, framed IP, NAS
address/port, start time, duration, timeout, and traffic counters. Filters
cover username, session ID, IPv4/IPv6 addresses and prefixes, NAS address, MAC,
and start-time ranges.

Per-row actions (this is where RFC 5176 dynamic authorization lives):

- **修改授权 / CoA** — send a `CoA-Request` with a new session timeout and/or
  `Filter-Id`.
- **强制下线 / Disconnect** — send a `Disconnect-Request` to terminate the
  session (with confirmation).

Both target the NAS CoA port and report success/failure as a notification.

## Accounting

Historical accounting records (`radius_accounting`), read-only: session ID,
addresses, start/stop time, total input/output traffic. Same filter set as
Online Sessions plus stop-time data; CSV export supported.

## System Config

Visible to `super`/`admin` only. Settings are schema-driven, grouped into
accordions (RADIUS group expanded by default), and live in the `sys_config`
table — changes apply without restart. The 13 RADIUS settings:

| Key | Default | Purpose |
| --- | ------- | ------- |
| `EapMethod` | `eap-md5` | Active EAP method: `eap-md5`, `eap-mschapv2`, `eap-tls`, `eap-peap`, `eap-ttls` |
| `EapEnabledHandlers` | `*` | Comma allow-list of permitted EAP handlers |
| `EapTlsCertFile` / `EapTlsKeyFile` | empty | Server certificate/key for EAP-TLS/PEAP/TTLS; **empty disables TLS-based EAP** |
| `EapTlsCaFile` | empty | CA bundle for client-certificate validation |
| `EapTlsMinVersion` | `1.2` | Minimum TLS version (`1.2`/`1.3`) |
| `IgnorePassword` | `false` | Skip password verification (testing only) |
| `AccountingHistoryDays` | `90` | Accounting retention days (`@daily` cleanup; `0` disables) |
| `AcctInterimInterval` | `300` | Seconds between NAS interim updates |
| `SessionTimeout` | `3600` | Default session timeout seconds |
| `LogLevel` | `info` | RADIUS log verbosity (`debug`/`info`/`warn`/`error`) |
| `RejectDelayMaxRejects` | `7` | Consecutive rejects before delaying responses |
| `RejectDelayWindowSeconds` | `10` | Window for the reject counter |

Toolbar: **Save**, **Refresh**, **Reset** (restore defaults, confirmed), and
**Backup / Restore** — export or re-import a JSON snapshot of nodes, NAS,
profiles, users, settings, and operators (see the
[Operations Guide](./ops-guide.md#backup-and-restore) for what it does *not*
include).

## Operators

Management of console accounts (`super`/`admin` only): username, password,
contact info, level, status. The operator action log is kept in the database
(`sys_opr_log`) and purged automatically after one year.

## UI conveniences

- **Language switcher** (app bar) — 简体中文 / English, persisted per browser.
- **Theme toggle** — light/dark.
- **CSV export** on users, sessions, accounting, NAS, nodes, operators.
- Server-side pagination and active-filter chips on all list pages.
