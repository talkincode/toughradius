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

### Dial-up authenticates but gets no IP / the address pool has no effect

`Framed-Pool` is sent only when a pool is **set on the user or its rate
profile**, and the emitted pool name **must match the pool name on the NAS**
(e.g. RouterOS `/ip pool`). A name mismatch — or no pool at all — leaves the
client without a (correct) address. For a fixed IP, set a static IPv4 on the
user (emits `Framed-IP-Address`, overriding the pool). See the end-to-end
[Scenario Cookbook · MikroTik](./cookbook-mikrotik.md).

### I have multiple NAS devices — must each be configured separately?

Yes. **Every NAS must be registered individually under NAS Devices**, each with
its own source IP (or identifier) and **its own shared secret**. ToughRADIUS
matches the packet's source address (or NAS identifier) to the NAS record and
its secret; requests from an unregistered source are logged and rejected as an
unauthorized NAS. Different NAS devices may use different secrets — they need
not be uniform.

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

### What breaks if the server and device clocks drift apart?

Clock skew causes subtle problems: TLS-based EAP (EAP-TLS / PEAP / TTLS)
validates certificate validity windows (`NotBefore` / `NotAfter`), so a large
time gap can fail the handshake; accounting start/stop times and durations
become wrong; and the brute-force reject-delay window (default 10 s) is measured
on the server clock. Run NTP on both the server and the NAS to keep time in
sync.

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

### How do I change a user's speed live (FUP / over-quota throttling)?

ToughRADIUS's Change of Authorization (CoA) carries **only `Session-Timeout` and
`Filter-Id`** — it does **not** rewrite rate attributes like `Mikrotik-Rate-Limit`
live. The standard way to change speed in real time is to change the rate on the
profile / user first, **then force a disconnect**; the client redials and is
re-authorized at the new speed. Alternatively use `Filter-Id` to apply a
pre-defined rate rule on the device. See
[Scenario Cookbook · MikroTik · Live control](./cookbook-mikrotik.md).

### CoA / disconnect says the session was not found

An operator-initiated CoA / Disconnect first locates the session on the Online
Sessions page, then addresses the device using its NAS record and session
identity (e.g. `Acct-Session-Id`). If the online row is stale (the NAS never
sent an accounting stop) or the session already ended, there is nothing to
match — refresh the online list, confirm accounting works on the device, and
retry.

### Online sessions show users that already disconnected

Online entries are created/refreshed by NAS accounting packets. If the NAS
stops sending (reboot, link loss) the row can linger. Ensure accounting and
interim updates are enabled on the device (`Acct-Interim-Interval` is sent in
every Access-Accept, default 300 s). Stale entries can also be cleared manually
from the UI with Disconnect/delete.

### Accounting table grows forever — what is cleaned automatically?

Right now **only the operator action log is purged automatically**: a `@daily`
job deletes operation logs (`SysOprLog`) older than one year. `radius_accounting`
(accounting history) and stale `radius_online` rows are **not** cleaned
automatically — the cleanup routine `SchedClearExpireData` (which would purge
accounting by `AccountingHistoryDays`, default 90, and remove expired online
rows) is implemented but **not yet registered with the scheduler** (tracked as
roadmap M6.4). So in high-volume deployments add **database-level archiving /
cleanup** to your ops routine, and clear stale online rows manually from the UI.
Configuration backups do **not** include accounting history — see
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
