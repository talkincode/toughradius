# ToughRADIUS AI Coding Agent Instructions

## 🔍 Mandatory Requirements Before Development

**Never modify code blindly. Always gather context from the existing implementation, related tests, and documentation first.**

### When to Perform Context Retrieval

1. **Before Feature Development** – Locate similar features to mirror naming, error handling, and data flow
2. **Before Bug Fixes** – Trace the full execution path (handlers, services, models, DB access) before patching
3. **Before Refactoring** – Map dependencies and side effects to avoid regressions
4. **When Learning Conventions** – Study logging patterns, configuration helpers, and TDD expectations

### Recommended Search Techniques

```text
# Feature exploration
semantic_search "vendor attribute parsing" in internal/radiusd/vendors
file_search "*/internal/adminapi/*routes*.go"

# Bug fixing
list_code_usages AuthenticateUser
grep_search "AuthError" --include internal/radiusd/**

# Refactoring
semantic_search "errgroup" in main.go
grep_search "app.GDB" --include internal/app/**
```

**Core Principle: Understand existing code → Follow project conventions → Maintain consistency**

---

## Project Overview

ToughRADIUS is an enterprise-grade RADIUS server developed in Go, supporting standard RADIUS protocols (RFC 2865/2866) and RadSec (RADIUS over TLS). The frontend uses React Admin framework for the management interface.

## Architecture Highlights

### Core Service Concurrency Model

`main.go` uses `errgroup` to start multiple independent services concurrently. If any service crashes, the entire application exits:

- **Web/Admin API** - Echo framework, port 1816 (`internal/webserver` + `internal/adminapi`)
- **RADIUS Auth** - Authentication service, UDP 1812 (`internal/radiusd`)
- **RADIUS Acct** - Accounting service, UDP 1813 (`internal/radiusd`)
- **RadSec** - TLS-encrypted RADIUS over TCP, port 2083 (`internal/radiusd`)

### Project Structure Pattern

Follows golang-standards/project-layout:

- `internal/` - Private code, cannot be imported externally
  - `domain/` - **Unified data models** (all GORM models listed in `domain/tables.go`)
  - `adminapi/` - New management API routes (v9 refactor)
  - `radiusd/` - RADIUS protocol core implementation
  - `app/` - Global application instance (database, config, scheduled tasks)
- `pkg/` - Reusable public libraries (utilities, encryption, Excel, etc.)
- `web/` - React Admin frontend (TypeScript + Vite)

### Database Access Pattern

**Always** obtain GORM instance through `app.GDB()`, do not inject DB connection directly:

```go
// Correct
user := &domain.RadiusUser{}
app.GDB().Where("username = ?", name).First(user)

// Wrong - Don't do this
type Service struct { DB *gorm.DB }
```

Supports PostgreSQL (default) and SQLite (pure Go, no CGO). Database migration is automatically handled by `app.MigrateDB()`.

### Vendor Extension Handling

RADIUS protocol supports multi-vendor features, distinguished by `VendorCode` field:

- Huawei (2011) - `internal/radiusd/vendors/huawei/`
- Mikrotik (14988) - See `auth_accept_config.go`
- Cisco (9) / Ikuai (10055) / ZTE (3902) / H3C (25506)

When adding new vendor support, define constants in `radius.go`, then add switch cases in `auth_accept_config.go` and related processing functions.

## Key Development Workflows

### Build and Run

**Local Development** (SQLite supported):

```bash
CGO_ENABLED=0 go run main.go -c toughradius.yml
```

**Production Build** (PostgreSQL only, static compilation):

```bash
make build  # Output to ./release/toughradius
```

**Frontend Development**:

```bash
cd web
npm install
npm run dev      # Development server with hot reload
npm run build    # Production build, output to dist/
```

### Database Initialization

```bash
./toughradius -initdb -c toughradius.yml  # Drop and recreate all tables
```

Production environment uses `MigrateDB(false)` for automatic migration (configured in main.go).

### Testing Standards

- RADIUS protocol tests: `internal/radiusd/*_test.go`
- Benchmark tests: `cmd/benchmark/bmtest.go` (standalone tool)
- Frontend tests: Playwright in `web/`

Run tests:

```bash
go test ./...                    # All unit tests
go test -bench=. ./internal/radiusd/  # Benchmark tests
```

## Common Patterns and Conventions

### Code Documentation Standards

**All exported APIs MUST have comprehensive comments:**

```go
// AuthenticateUser validates user credentials against the RADIUS database.
// It checks username/password, account expiration, and session limits.
//
// Parameters:
//   - username: User's login name (case-sensitive)
//   - password: Plain text password (will be hashed internally)
//   - nasIP: Network Access Server IP address for session tracking
//
// Returns:
//   - *domain.RadiusUser: User object if authentication succeeds
//   - error: AuthError with metrics tag if validation fails
//
// Common errors:
//   - MetricsRadiusRejectUserNotFound: Username doesn't exist
//   - MetricsRadiusRejectPasswordError: Password mismatch
func AuthenticateUser(username, password, nasIP string) (*domain.RadiusUser, error) {
    // Implementation
}
```

**Complex logic requires inline comments explaining the "why":**

```go
// Huawei devices expect bandwidth in Kbps, but our plan stores it in Mbps
// Convert using 1024 (binary) not 1000 (decimal) for compatibility
return baseBandwidth * 1024
```

**Vendor-specific code must reference protocol specifications:**

```go
// ParseHuaweiInputAverageRate extracts bandwidth limit from Huawei VSA attribute.
// Format: Type=11, Length=variable, Value=bandwidth_in_kbps
// See: internal/radiusd/vendors/huawei/README.md for full VSA specification
```

### Error Handling

RADIUS authentication errors use custom `AuthError` type with metrics tags:

```go
return NewAuthError(app.MetricsRadiusRejectExpire, "user expire")
```

These errors are automatically recorded to Prometheus metrics (`internal/app/metrics.go`).

### Configuration Reading

Access global config and settings via `app.GApp()`:

```go
// Read RADIUS config items
eapMethod := app.GApp().GetSettingsStringValue("radius", "EapMethod")
maxSessions := app.GApp().GetSettingsInt64Value("radius", "MaxSessions")
```

System configuration is stored in `sys_config` table, initialized with default values through `checkSettings()`.

### Concurrency Handling

RADIUS request processing uses ants goroutine pool to limit concurrency:

```go
radiusService.TaskPool.Submit(func() { /* Handle request */ })
```

Pool size is configured via environment variable `TOUGHRADIUS_RADIUS_POOL` (default 1024).

### Logging Standards

Use zap structured logging, **always** add namespace:

```go
zap.L().Error("update user failed",
    zap.Error(err),
    zap.String("namespace", "radius"))
```

### Admin API Route Registration

When adding new management APIs, create file in `internal/adminapi/` and register in `adminapi.go`'s `Init()`:

```go
// users.go
func registerUserRoutes() {
    // Route definitions
}

// adminapi.go
func Init() {
    registerUserRoutes()  // Add this line
}
```

## Key Dependencies and Integrations

- **Echo v4** - Web framework, middleware configured in `internal/webserver/server.go`
- **GORM** - ORM, automatic migration controlled by `domain.Tables` list
- **layeh.com/radius** - RADIUS protocol library, don't mix with other RADIUS packages
- **React Admin 5.0** - Frontend framework, REST data provider in `web/src/dataProvider.ts`
