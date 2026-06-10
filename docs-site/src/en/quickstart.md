# Quick Start

> 中文版本：[快速开始](../zh/quickstart.md)

This chapter takes you from nothing to a working RADIUS server with one test
user, then shows how to debug what the server is doing. Default ports: web UI
`1816`, RADIUS authentication UDP `1812`, accounting UDP `1813`, RadSec TCP
`2083`.

## 1. Install

Pick one of three paths.

### Option A — pre-built binary

Download the binary for your platform from the
[GitHub Releases](https://github.com/talkincode/toughradius/releases) page
(`toughradius_linux_amd64`, `toughradius_linux_arm64`, `toughradius_darwin_arm64`,
`toughradius_windows_amd64.exe`, …), then:

```bash
chmod +x toughradius_linux_amd64
sudo mv toughradius_linux_amd64 /usr/local/bin/toughradius
```

### Option B — Docker

```bash
docker pull talkincode/toughradius:latest

docker run -d --name toughradius \
  -p 1816:1816 -p 1812:1812/udp -p 1813:1813/udp -p 2083:2083 \
  -v toughradius-data:/var/toughradius \
  talkincode/toughradius:latest -c /etc/toughradius.yml
```

The image exposes `1816/tcp`, `1812/udp`, `1813/udp`, and `2083/tcp`. Mount a
volume at `/var/toughradius` (the default working directory) so the SQLite
database, logs, and certificates survive container restarts.

### Option C — build from source

Requires Go 1.25+ and Node.js 20+ (the React Admin frontend is embedded into
the binary):

```bash
git clone https://github.com/talkincode/toughradius.git
cd toughradius
make build          # builds web/ then the Go binary → release/toughradius
```

For backend-only development with hot SQLite defaults:

```bash
make runs           # CGO_ENABLED=0 go run main.go -c toughradius.yml
make runf           # frontend dev server at http://localhost:3000/admin
```

## 2. Configure

ToughRADIUS looks for its configuration in this order: the `-c <file>` flag,
`./toughradius.yml`, `/etc/toughradius.yml`, and finally built-in defaults.
Environment variables override the file (see the
[Operations Guide](./ops-guide.md#environment-variables)).

A minimal production-style configuration:

```yaml
system:
  appid: ToughRADIUS
  location: Asia/Shanghai
  workdir: /var/toughradius     # data/logs/certs live here
  debug: false

web:
  host: 0.0.0.0
  port: 1816
  secret: change-me-to-a-long-random-string   # JWT signing secret

database:
  type: sqlite                  # or: postgres (+ host/port/user/passwd)
  name: toughradius.db          # stored under {workdir}/data/

radiusd:
  enabled: true
  host: 0.0.0.0
  auth_port: 1812
  acct_port: 1813
  radsec_port: 2083
  debug: true                   # log full packet dumps; disable in production

logger:
  mode: production
  file_enable: true
  filename: /var/toughradius/toughradius.log
```

> **Change `web.secret`.** It signs admin login tokens. Likewise change the
> default admin password immediately after the first login.

## 3. Initialize the database and run

```bash
# FIRST TIME ONLY — drops and recreates all tables
toughradius -initdb -c /etc/toughradius.yml

# start the server
toughradius -c /etc/toughradius.yml
```

`-initdb` is destructive; on later upgrades just start the server — schema
migration runs automatically at startup. Other flags: `-v` prints the version,
`-printcfg` prints the merged configuration as JSON.

## 4. Log in to the admin UI

Open `http://<server>:1816`. The default administrator is:

- Username: `admin`
- Password: `toughradius`

Change the password right away under **Account Settings**, or reset a lost one
with `cmd/reset-password` (see the [FAQ](./faq.md)).

## 5. Register a NAS and create a user

1. **Network Nodes → Create** — make a node (a logical group), e.g. `default`.
2. **NAS Devices → Create** — register your network device:
   - *IP address*: the address the device sends RADIUS from.
   - *Secret*: the shared secret, e.g. `testing123`.
   - *Vendor code*: pick the device vendor (Standard / Cisco / Huawei /
     MikroTik / H3C / ZTE / iKuai) — this controls vendor-specific attributes;
     see the [Vendor Integration Guide](./vendor-guide.md).
   - *CoA port*: leave `3799` unless your device uses another port.
3. **Billing Profiles → Create** — e.g. `100M`: concurrency `1`, up rate
   `51200` Kbps, down rate `102400` Kbps.
4. **RADIUS Users → Create** — username `test1`, password `111111`, pick the
   profile and an expiration date.

## 6. Verify with radtest

The repository ships a small RADIUS client (defaults shown):

```bash
go run ./cmd/radtest auth \
  -server 127.0.0.1 -secret testing123 \
  -username test1 -password 111111

go run ./cmd/radtest flow ...   # auth + acct-start + acct-stop in one run
```

A successful run prints `Access-Accept` with the returned attributes. The
session then appears under **Online Sessions** (for `flow`) and the dashboard
counters increase.

## 7. Debugging

| What you need | How |
| ------------- | --- |
| Full RADIUS packet dumps | `radiusd.debug: true` in the YAML (or env `TOUGHRADIUS_RADIUS_DEBUG=true`), or set **System Config → RADIUS → Log Level** to `debug` at runtime |
| Log file location | `logger.filename`, default `{workdir}/toughradius.log`; `logger.mode: development` for human-readable console output |
| Why a user is rejected | Reject reasons are counted per cause (wrong password, expired, bound MAC mismatch, …) and shown on the dashboard; the log carries the detail |
| Inspect effective config | `toughradius -printcfg -c <file>` |
| Load testing | `go run ./cmd/benchmark` — see the [Operations Guide](./ops-guide.md#command-line-tools) |

## Next steps

- [Vendor Integration Guide](./vendor-guide.md) — configure Cisco, Huawei,
  MikroTik, H3C, ZTE, iKuai and standard devices.
- [Admin UI Manual](./admin-manual.md) — every page of the management UI.
- [Operations Guide](./ops-guide.md) — production deployment, TLS, EAP
  certificates, backups, and monitoring.
