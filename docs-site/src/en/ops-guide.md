# Operations Guide

> 中文版本：[运维指南](../zh/ops-guide.md)

Everything you need to run ToughRADIUS in production: configuration reference,
environment variables, TLS/EAP certificates, storage, monitoring, backup, and
the bundled command-line tools.

## Process model

One static binary runs several services concurrently (web/admin API, RADIUS
auth, RADIUS accounting, RadSec). **If any service fails, the whole process
exits** so a supervisor can restart it — run it under systemd, Docker, or an
equivalent.

```ini
# /etc/systemd/system/toughradius.service (reference)
[Unit]
Description=ToughRADIUS server
After=network-online.target

[Service]
ExecStart=/usr/local/bin/toughradius -c /etc/toughradius.yml
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
```

## Ports

| Port | Protocol | Service | Config key |
| ---- | -------- | ------- | ---------- |
| 1816 | TCP HTTP | Admin UI + REST API | `web.port` |
| 1817 | TCP HTTPS | Admin UI over TLS (optional; start failure is non-fatal) | `web.tls_enabled` / `web.tls_port` |
| 1812 | UDP | RADIUS authentication | `radiusd.auth_port` |
| 1813 | UDP | RADIUS accounting | `radiusd.acct_port` |
| 2083 | TCP TLS | RadSec (RFC 6614) | `radiusd.radsec_port` |
| 3799 | UDP (outbound) | CoA/Disconnect **to** the NAS | per-NAS *CoA port* field |

## Configuration

Lookup order: `-c <file>` → `./toughradius.yml` → `/etc/toughradius.yml` →
embedded defaults. Inspect the merged result with `toughradius -printcfg`.

```yaml
system:
  appid: ToughRADIUS
  location: Asia/Shanghai        # cron/timestamp timezone
  workdir: /var/toughradius      # default in production builds
  debug: false
web:
  host: 0.0.0.0
  port: 1816
  tls_enabled: true              # Enabled by default for compatibility; set false to disable the built-in HTTPS listener
  tls_port: 1817
  secret: <random-string>        # JWT signing secret — change it
database:
  type: sqlite                   # sqlite | postgres
  host: 127.0.0.1                # postgres only
  port: 5432
  name: toughradius.db           # sqlite filename (under {workdir}/data/) or pg database
  user: postgres
  passwd: <password>
  max_conn: 100
  idle_conn: 10
  debug: false
radiusd:
  enabled: true
  host: 0.0.0.0
  auth_port: 1812
  acct_port: 1813
  radsec_port: 2083
  radsec_worker: 100
  radsec_ca_cert: private/ca.crt        # relative paths resolve against workdir
  radsec_cert: private/radsec.tls.crt
  radsec_key: private/radsec.tls.key
  debug: false                   # true = full packet dumps
logger:
  mode: production               # development | production
  file_enable: true
  filename: /var/toughradius/toughradius.log
```

### Working directory layout

On startup ToughRADIUS creates under `system.workdir`:

```text
/var/toughradius/
├── data/        # SQLite database, metrics data
├── logs/
├── private/     # TLS material (mode 0700)
├── public/
└── backup/      # server-side copies of configuration backups
```

### Environment variables

Environment variables override the YAML file:

| Variable | Overrides |
| -------- | --------- |
| `TOUGHRADIUS_SYSTEM_WORKER_DIR` | `system.workdir` |
| `TOUGHRADIUS_SYSTEM_DEBUG` | `system.debug` |
| `TOUGHRADIUS_WEB_HOST` / `_WEB_PORT` / `_WEB_TLS_ENABLED` / `_WEB_TLS_PORT` / `_WEB_SECRET` | `web.*` |
| `TOUGHRADIUS_DB_TYPE` / `_DB_HOST` / `_DB_PORT` / `_DB_NAME` / `_DB_USER` / `_DB_PWD` / `_DB_DEBUG` | `database.*` |
| `TOUGHRADIUS_RADIUS_ENABLED` / `_RADIUS_HOST` / `_RADIUS_AUTHPORT` / `_RADIUS_ACCTPORT` / `_RADIUS_DEBUG` | `radiusd.*` |
| `TOUGHRADIUS_RADIUS_RADSEC_PORT` / `_RADIUS_RADSEC_WORKER` / `_RADIUS_RADSEC_CA_CERT` / `_RADIUS_RADSEC_CERT` / `_RADIUS_RADSEC_KEY` | RadSec settings |
| `TOUGHRADIUS_LOGGER_MODE` / `_LOGGER_FILE_ENABLE` | `logger.*` |
| `TOUGHRADIUS_RADIUS_POOL` | RADIUS worker pool size (default 1024) |

### CLI flags

| Flag | Effect |
| ---- | ------ |
| `-c <file>` | Configuration file path |
| `-initdb` | **Drop and recreate all tables**, then exit |
| `-printcfg` | Print merged configuration as JSON, exit |
| `-v` | Print version / build time / commit, exit |
| `-h` | Usage |

Runtime RADIUS settings (EAP method, certificates, intervals, reject-delay…)
live in the database and are edited in **System Config** — no restart needed.
See the [Admin UI Manual](./admin-manual.md#system-config).

## Database

- **SQLite** (default) — pure-Go driver, zero CGO, file at
  `{workdir}/data/<name>`. Fine for small/medium deployments; back up the file.
- **PostgreSQL** — set `database.type: postgres` plus host/user/password.
  Recommended for production scale and concurrent accounting load.

Schema migration (GORM `AutoMigrate`) runs automatically at every startup, so
upgrades are: stop, replace the binary, start. `-initdb` is for first
installation only — it **destroys all data**.

Large tables to watch: `radius_accounting` (grows with every session) and
`radius_online`. The `radius.AccountingHistoryDays` setting (default 90, set to
`0` to disable) defines the accounting retention window: a `@daily` job deletes
**terminated** `radius_accounting` rows older than that many days (active
sessions are untouched) and clears dangling `radius_online` rows that have
missed several interim updates. The operator action log (`sys_opr_log`) is
purged automatically after one year. For very high volumes, still consider
database-level archiving as part of your own ops.

## TLS and certificates

Three independent certificate consumers:

| Consumer | Files | Notes |
| -------- | ----- | ----- |
| **RadSec** | `radiusd.radsec_ca_cert` / `radsec_cert` / `radsec_key` | TLS 1.2+; client certificates are verified **if presented** (`VerifyClientCertIfGiven`) |
| **Web HTTPS** | `{workdir}/private/toughradius.tls.crt` + `.key` (fixed paths) | Listens on `web.tls_port` when `web.tls_enabled` is true; failure to load is logged, HTTP keeps running |
| **EAP (TLS/PEAP/TTLS)** | System Config → `EapTlsServerCert`, `EapTlsClientCa`, `EapTlsMinVersion` (import certificates on the Certificates page and select them by name) | No server certificate selected disables certificate-based EAP methods |

Generate a complete CA/server/client set with the bundled tool:

```bash
go run ./cmd/certgen -type all -output /var/toughradius/private \
  -server-cn radius.example.com -server-dns radius.example.com \
  -days 3650
# then point radsec_cert/radsec_key (or the EAP settings) at the files
```

## Logging

zap structured logging. `logger.mode: development` = human-readable console;
`production` = JSON. File output controlled by `logger.file_enable` +
`logger.filename`. RADIUS verbosity is additionally tunable at runtime via
**System Config → LogLevel**; `radiusd.debug: true` dumps full packets (keep
off in production).

## Metrics

Counters are kept in memory and surfaced through the admin dashboard (there is
no Prometheus `/metrics` HTTP endpoint). RADIUS counters include: `radus_accept`,
`radus_online`/`radus_offline`, `radus_accounting`, `radus_auth_drop` /
`radus_acct_drop`, `radus_radsec_saturated`, and per-cause reject counters —
`radus_reject_passwd_error`, `radus_reject_not_exists`, `radus_reject_expire`,
`radus_reject_disabled`, `radus_reject_limit`, `radus_reject_bind_error`,
`radus_reject_ldap_error`, `radus_reject_unauthorized`, `radus_reject_other`.
`radus_reject_ldap_error` means the LDAP/AD backend could not give an
authentication answer — for example the directory is unreachable, TLS/StartTLS
failed, the service account bind failed, or the LDAP configuration is wrong;
wrong passwords are still counted under `radus_reject_passwd_error`.
Accounting-Requests dropped
at ingress are classified by reason — `radus_acct_drop_nas` (unknown or
unauthorized NAS), `radus_acct_drop_username` (missing username), and
`radus_acct_drop_secret` (bad Request Authenticator) — while `radus_acct_drop`
remains the catch-all for back-pressure and response-write drops. System gauges
(CPU/memory, process CPU/memory) are sampled every 30 s.

For external monitoring, probe the service ports and watch the log file; treat
process exit as the failure signal (the process model is fail-fast).

## Backup and restore

**System Config → Backup** downloads a JSON snapshot (schema version 9.0) of:
nodes, NAS devices, profiles, users, system settings, operators, and managed
certificates (including their private keys, so certificate-based EAP keeps
working after a restore). A copy is also written to `{workdir}/backup/`.
**Restore** re-imports such a file.

> The snapshot does **not** include accounting history or online sessions. For
> a full disaster-recovery story, also back up the database itself (copy the
> SQLite file or use `pg_dump`).

> **Security**: the snapshot contains secrets — RADIUS user passwords,
> operator password hashes, and the PEM private keys of managed EAP
> certificates. Store and transfer backup files as you would any credential.

## Command-line tools

All live under `cmd/` and run with `go run ./cmd/<tool>`:

| Tool | Purpose |
| ---- | ------- |
| `radtest` | Mini RADIUS client: `auth`, `acct`, `flow` (auth + start + stop). Flags: `-server`, `-secret`, `-username`, `-password`, `-calling-station`, `-framed-ip`, `-session-id` |
| `certgen` | Generate CA / server / client certificates (see above) |
| `benchmark` | Load tester: total requests `-n`, concurrency `-c`, auth/acct modes, CSV stats output |
| `reset-password` | Reset a console operator password: `go run ./cmd/reset-password -c <cfg> -u admin -p <new>` |
| `demo-seed` | Populate demo nodes/NAS/profiles/users/sessions for evaluation |
| `config-tool` | Validate / summarize the settings schema JSON |

## Production hardening checklist

- [ ] Change `web.secret` and the default `admin` password.
- [ ] `radiusd.debug: false`, `logger.mode: production`.
- [ ] Restrict UDP 1812/1813 and TCP 1816 to trusted networks (firewall).
- [ ] Use RadSec (2083) or a trusted L2/VPN path for RADIUS across untrusted networks.
- [ ] Unique, strong shared secret per NAS.
- [ ] EAP: prefer `eap-tls`/`eap-peap`/`eap-ttls` with real certificates; mind
      the MS-CHAPv2 caveats in the [Security Policy](./security-policy.md).
- [ ] Database backups scheduled (config snapshot + DB dump).
- [ ] Supervisor with restart policy; alert on process exit.
