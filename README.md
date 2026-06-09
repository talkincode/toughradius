Welcome to the TOUGHRADIUS project!

     _____   _____   _   _   _____   _   _   _____        ___   _____   _   _   _   _____
    |_   _| /  _  \ | | | | /  ___| | | | | |  _  \      /   | |  _  \ | | | | | | /  ___/
      | |   | | | | | | | | | |     | |_| | | |_| |     / /| | | | | | | | | | | | | |___
      | |   | | | | | | | | | |  _  |  _  | |  _  /    / / | | | | | | | | | | | | \___  \
      | |   | |_| | | |_| | | |_| | | | | | | | \ \   / /  | | | |_| | | | | |_| |  ___| |
      |_|   \_____/ \_____/ \_____/ |_| |_| |_|  \_\ /_/   |_| |_____/ |_| \_____/ /_____/

# TOUGHRADIUS

[![License](https://img.shields.io/github/license/talkincode/toughradius)](https://github.com/talkincode/toughradius/blob/main/LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/talkincode/toughradius)](go.mod)
[![Release](https://img.shields.io/github/v/release/talkincode/toughradius)](https://github.com/talkincode/toughradius/releases)
[![Build Status](https://github.com/talkincode/toughradius/actions/workflows/ci.yml/badge.svg)](https://github.com/talkincode/toughradius/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/talkincode/toughradius/graph/badge.svg)](https://codecov.io/gh/talkincode/toughradius)
[![Docker Pulls](https://img.shields.io/docker/pulls/talkincode/toughradius)](https://hub.docker.com/r/talkincode/toughradius)

A powerful, open-source RADIUS server designed for ISPs, enterprise networks, and carriers. Supports standard RADIUS protocols, a full EAP / 802.1X authentication suite (EAP-TLS, PEAPv0/EAP-MSCHAPv2, EAP-TTLS), RadSec (RADIUS over TLS), and a modern Web management interface.

## ✨ Core Features

### RADIUS Protocol Support

- 🔐 **Standard RADIUS** - Full support for RFC 2865/2866 authentication and accounting protocols
- 🔒 **RadSec** - TLS encrypted RADIUS over TCP (RFC 6614)
- 🌐 **Multi-Vendor Support** - Compatible with major network devices like Cisco, Mikrotik, Huawei, etc.
- ⚡ **High Performance** - Built with Go, supporting high concurrency processing

### EAP / 802.1X Authentication

A pluggable EAP handler registry covers both challenge and tunneled methods:

- 🪪 **EAP-MD5 / EAP-MSCHAPv2** - Challenge-based password methods (RFC 3748)
- 🔐 **EAP-TLS** - Certificate-based mutual authentication with TLS handshake, fragmentation reassembly, and certificate-to-identity mapping (RFC 5216)
- 🪟 **PEAPv0 / EAP-MSCHAPv2** - Server-certificate TLS tunnel carrying inner EAP-MSCHAPv2 with MPPE key derivation, for Windows / AD / legacy enterprise networks
- 🧩 **EAP-TTLS** - TLS tunnel carrying inner PAP / MS-CHAPv2, onboarding LDAP and legacy credential stores without a per-client certificate rollout (RFC 5281)

> ⚠️ **Compatibility note:** PEAP and EAP-MSCHAPv2 are *compatibility-first* methods. MS-CHAPv2-style exchanges carry an NTLMv1-like attack surface (see Microsoft guidance). Use them to serve legacy devices and AD users; prefer **EAP-TLS** for new deployments where you control the client certificate estate.

### Management Features

- 📊 **React Admin Interface** - Modern Web management dashboard
- 👥 **User Management** - Complete user account and profile management
- 📈 **Real-time Monitoring** - Online session monitoring and accounting record queries
- 🔍 **Log Auditing** - Detailed authentication and accounting logs

### Integration Capabilities

- **Multi-Database Support** - PostgreSQL, SQLite
- 🔌 **Flexible Extension** - Supports custom authentication and accounting logic
- 📡 **Multi-Vendor VSA** - Huawei, Mikrotik, Cisco, H3C, etc.

## 🚀 Quick Start

### Prerequisites

- Go 1.25+ (for building from source)
- PostgreSQL or SQLite
- Node.js 18+ (for frontend development)

### Installation

#### 1. Build from Source

```bash
# Clone repository
git clone https://github.com/talkincode/toughradius.git
cd toughradius

# Build frontend
cd web
npm install
npm run build
cd ..

# Build backend
go build -o toughradius main.go
```

#### 2. Use Pre-compiled Version

Download the latest version from the [Releases](https://github.com/talkincode/toughradius/releases) page.

### Configuration

1. Copy the configuration template:

```bash
cp toughradius.yml toughradius.prod.yml
```

2. Edit `toughradius.prod.yml` configuration file:

```yaml
system:
  appid: ToughRADIUS
  location: Asia/Shanghai
  workdir: ./rundata

database:
  type: sqlite # or postgres
  name: toughradius.db
  # PostgreSQL configuration
  # host: localhost
  # port: 5432
  # user: toughradius
  # passwd: your_password

radiusd:
  enabled: true
  host: 0.0.0.0
  auth_port: 1812 # RADIUS authentication port
  acct_port: 1813 # RADIUS accounting port
  radsec_port: 2083 # RadSec port

web:
  host: 0.0.0.0
  port: 1816 # Web management interface port
```

### EAP Configuration

ToughRADIUS registers the following EAP handlers out of the box:

| Method | Kind | Notes |
| --- | --- | --- |
| `eap-md5` | Challenge | Default; password challenge (RFC 3748) |
| `eap-mschapv2` | Challenge | MS-CHAPv2 password challenge |
| `eap-tls` | Tunneled (certificate) | Certificate-based mutual authentication (RFC 5216) |
| `eap-peap` | Tunneled | PEAPv0 with inner EAP-MSCHAPv2 (Windows / AD) |
| `eap-ttls` | Tunneled | Inner PAP / MS-CHAPv2 (RFC 5281) |

Fine-tune authentication behavior via system configuration (`sys_config`):

- `radius.EapMethod`: Preferred EAP method offered on EAP-Identity (default `eap-md5`).
- `radius.EapEnabledHandlers`: Allow-list of enabled handlers, comma-separated, e.g. `eap-md5,eap-mschapv2,eap-tls`. Use `*` to enable all registered handlers.

This lets you disable unauthorized EAP methods without interrupting the service.

> ⚠️ MS-CHAPv2-based methods (`eap-mschapv2`, `eap-peap`, and TTLS inner MS-CHAPv2) are compatibility-oriented and carry an NTLMv1-like attack surface. Prefer `eap-tls` for new deployments where you control client certificates.

### Running

```bash
# Initialize database
./toughradius -initdb -c toughradius.prod.yml

# Start service
./toughradius -c toughradius.prod.yml
```

Access Web Management Interface: <http://localhost:1816>

Default Admin Account:

- Username: admin
- Password: Please check the initialization log output

## 📖 Documentation

- 📚 **[Bilingual Handbook (mdbook)](https://www.toughradius.net/)** - CN/EN documentation site (source in [`docs-site/`](docs-site/)) consolidating the overview, security policy, RFC reference, and more; built, link-checked, and deployed to GitHub Pages by CI
- [Roadmap](docs/roadmap.md) - Milestones and the EAP suite delivery plan (EAP-TLS / PEAP / TTLS, with TLS 1.3, TEAP, EAP-PWD tracked)
- [Feature Checklist](docs/feature-checklist.md) / [English](docs/feature-checklist.en.md) - Product scope baseline for aligning future development with feature IDs and avoiding uncontrolled direction changes
- [Architecture](docs/v9-architecture.md) - v9 version architecture design
- [React Admin Refactor](docs/react-admin-refactor.md) - Frontend management interface explanation
- [SQLite Support](docs/sqlite-support.md) - SQLite database configuration
- [Environment Variables](docs/environment-variables.md) - Environment variable configuration guide

## 🏗️ Project Structure

```text
toughradius/
├── cmd/             # Application entry points
├── internal/        # Private application code
│   ├── adminapi/   # Admin API (New version)
│   ├── radiusd/    # RADIUS service core
│   ├── domain/     # Data models
│   └── webserver/  # Web server
├── pkg/            # Public libraries
├── web/            # React Admin frontend
└── docs/           # Documentation
```

## 🔧 Development

### Backend Development

```bash
# Run tests
go test ./...

# Run benchmark tests
go test -bench=. ./internal/radiusd/

# Start development mode
go run main.go -c toughradius.yml
```

### Frontend Development

```bash
cd web
npm install
npm run dev       # Development server
npm run build     # Production build
npm run lint      # Code linting
```

## 🤝 Contribution

We welcome contributions in various forms, including but not limited to:

- 🐛 Submitting Bug reports and feature requests
- 📝 Improving documentation
- 💻 Submitting code patches and new features
- 🌍 Helping with translation

## 📜 License

This project is licensed under the [MIT License](LICENSE).

### Third-Party Resources

The RADIUS dictionary files in the `share/` directory are derived from the [FreeRADIUS](https://freeradius.org/) project and are licensed under the [Creative Commons Attribution 4.0 International License (CC BY 4.0)](share/LICENSE).

## 🔗 Related Links

- [Official Website](https://www.toughradius.net/)
- [Online Documentation](https://github.com/talkincode/toughradius/wiki)
- [RadSec RFC 6614](https://tools.ietf.org/html/rfc6614)
- [RADIUS RFC 2865](https://tools.ietf.org/html/rfc2865)
- [EAP RFC 3748](https://tools.ietf.org/html/rfc3748)
- [EAP-TLS RFC 5216](https://tools.ietf.org/html/rfc5216)
- [EAP-TTLS RFC 5281](https://tools.ietf.org/html/rfc5281)
- [Mikrotik RADIUS Configuration](https://wiki.mikrotik.com/wiki/Manual:RADIUS_Client)

## 💎 Sponsors

Thanks to [JetBrains](https://jb.gg/OpenSourceSupport) for supporting this project!

![JetBrains Logo](https://resources.jetbrains.com/storage/products/company/brand/logos/jb_beam.svg)
