# FAQ

> 中文版本：[常见问题解答](../zh/faq.md)

Frequently asked questions, grouped by theme. If your question is not covered,
search the [GitHub issues](https://github.com/talkincode/toughradius/issues) or
open a new one.

## Installation & access

### I forgot the admin password — how do I reset it?

Use the bundled tool against the same configuration file the server runs with:

```bash
go run ./cmd/reset-password -c /etc/toughradius.yml -u admin -p <new-password>
```

The default account after `-initdb` is `admin` / `toughradius`.

### Which database should I choose, SQLite or PostgreSQL?

SQLite (the default) requires nothing extra — pure-Go driver, single file under
`{workdir}/data/` — and suits labs and small deployments. Choose PostgreSQL for
production scale, high accounting volume, or when you need external backup
tooling (`pg_dump`, replication).

### Can I run it on a port other than 1812/1813/1816?

Yes — every port is configurable (`radiusd.auth_port`, `radiusd.acct_port`,
`web.port`, …) via YAML or environment variables. See the
[Operations Guide](./ops-guide.md#ports).

### The HTTPS admin port (1817) doesn't work

The web TLS listener needs `{workdir}/private/toughradius.tls.crt` and `.key`.
If they are missing or invalid the failure is logged and **only** the HTTPS
listener stops — plain HTTP on 1816 keeps serving. Generate certificates with
`cmd/certgen` or provide your own.

### Is `-initdb` safe to run again?

**No.** It drops and recreates every table. Run it only at first installation.
Regular upgrades need no manual schema step — migration runs automatically at
startup.

## Authentication

### All requests are ignored / time out

The two most common causes:

1. The request's **source IP is not registered** as a NAS device — add it under
   **NAS Devices** (or fix NAT so the expected address is seen).
2. **Shared secret mismatch** — RADIUS silently discards packets that fail
   authentication of the packet itself.

Turn on `radiusd.debug: true` or set LogLevel to `debug` to see what arrives.

### Users authenticate but get no bandwidth limit

Rate attributes are sent only when the NAS record's **vendor** is one that has
a rate VSA (Huawei, MikroTik, H3C, ZTE, iKuai). A NAS registered as `Standard`
or `Cisco` receives no proprietary rate attribute — see the
[Vendor Integration Guide](./vendor-guide.md#rate-limit-units). Also check that
the user's profile actually sets `up_rate`/`down_rate` (Kbps).

### Why was a specific user rejected?

Rejects are categorized (wrong password, user not found, expired, disabled,
session limit, MAC/VLAN binding mismatch, unauthorized NAS…) — the dashboard
shows per-cause counters and the log records the detail. The most surprising
one is the **binding mismatch**: with `bind_mac`/`bind_vlan` enabled, the first
seen MAC/VLAN is stored and later requests must match; clear the stored value
on the user after replacing hardware.

### Repeated wrong passwords respond slowly — why?

That is the reject-delay brute-force guard: after
`RejectDelayMaxRejects` (default 7) consecutive rejects within
`RejectDelayWindowSeconds` (default 10 s), responses are delayed. Tune both in
**System Config**.

### Does ToughRADIUS support 802.1X / Wi-Fi Enterprise?

Yes. Supported EAP methods: EAP-MD5, EAP-MSCHAPv2, EAP-TLS, PEAPv0/EAP-MSCHAPv2
and EAP-TTLS (inner PAP / MS-CHAP-V2). Select the method in **System Config →
EapMethod**.

### EAP-TLS / PEAP / EAP-TTLS doesn't start

TLS-based EAP requires `EapTlsCertFile` + `EapTlsKeyFile` in System Config —
when they are empty the methods are disabled by design. Generate a server
certificate with `cmd/certgen`, set the paths, and retry. `EapTlsCaFile` is
needed only to validate client certificates (EAP-TLS).

### Which EAP method should I pick?

- **EAP-TLS** — strongest (mutual certificates), needs client-cert rollout.
- **PEAPv0/EAP-MSCHAPv2** — Windows/AD compatibility; mind the MS-CHAPv2
  NTLMv1-like attack surface (see [Security Policy](./security-policy.md)).
- **EAP-TTLS** — legacy/LDAP backends via inner PAP, keeping passwords inside
  the TLS tunnel.

## Sessions, CoA & accounting

### Disconnect / CoA from Online Sessions has no effect

Check, in order: the device has dynamic authorization enabled (e.g.
`radius incoming` on RouterOS, `aaa server radius dynamic-author` on IOS); the
**CoA port** on the NAS record matches the device (default 3799); and the
device accepts requests from the server address. ToughRADIUS waits 5 s and
retries twice before reporting failure.

### Online sessions show users that already disconnected

Online entries are created/refreshed by NAS accounting packets. If the NAS
stops sending (reboot, link loss) the row can linger. Ensure accounting and
interim updates are enabled on the device (`Acct-Interim-Interval` is sent in
every Access-Accept, default 300 s). Stale entries can also be cleared manually
from the UI with Disconnect/delete.

### Accounting table grows forever — what is cleaned automatically?

The operator action log is purged after one year. For `radius_accounting`,
the `AccountingHistoryDays` setting (default 90) defines the retention window
used for accounting cleanup; in high-volume deployments monitor growth and
include database-level archiving in your ops routine. Configuration backups do
**not** include accounting history — see
[Backup and restore](./ops-guide.md#backup-and-restore).

### Concurrent sessions are not limited

`active_num` in the billing profile is the per-user concurrency cap (0 means
unlimited). The check counts rows in `radius_online`, which requires working
accounting from the NAS — without accounting Start packets the server cannot
know who is online.

## Operations

### How do I monitor it with Prometheus?

There is currently no `/metrics` HTTP endpoint; counters are in-memory and
shown on the dashboard. For external monitoring probe the ports, watch the log
file, and alert on process exit (the process is fail-fast by design).

### How do I upgrade safely?

Stop the service, replace the binary (or pull the new Docker tag), start.
Schema migration is automatic. Take a configuration backup (System Config →
Backup) and a database backup first.

### Where are logs / data / certificates stored?

Everything lives under `system.workdir` (default `/var/toughradius`):
`data/` (SQLite DB), `logs/`, `private/` (TLS material), `backup/`
(config snapshots). See the
[Operations Guide](./ops-guide.md#working-directory-layout).
