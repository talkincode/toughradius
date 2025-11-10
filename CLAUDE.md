# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ToughRADIUS is an enterprise-grade RADIUS server built in Go with React Admin frontend, supporting standard RADIUS protocols (RFC 2865/2866) and RadSec (RADIUS over TLS).

## Architecture

The application runs multiple concurrent services using errgroup - if any service crashes, the entire application exits:

- **Web/Admin API** - Echo framework, port 1816 (`internal/webserver` + `internal/adminapi`)
- **RADIUS Auth** - Authentication service, UDP 1812 (`internal/radiusd`)
- **RADIUS Acct** - Accounting service, UDP 1813 (`internal/radiusd`)
- **RadSec** - TLS-encrypted RADIUS over TCP, port 2083 (`internal/radiusd`)

## Development Commands

### Backend Development

```bash
# Start backend with SQLite support (development)
make runs
# or
CGO_ENABLED=1 go run main.go -c toughradius.yml

# Build production version (PostgreSQL only, static compilation)
make build

# Initialize database (destructive - deletes all data)
make initdb

# Run tests
go test ./...

# Run benchmarks
go test -bench=. ./internal/radiusd/
```

### Frontend Development

```bash
cd web

# Install dependencies
npm install

# Start development server (hot reload)
npm run dev
# Access: http://localhost:3000/admin

# Build for production
npm run build

# Run tests
npm test

# Type checking
npm run type-check

# Linting
npm run lint
```

### Full Development Workflow

```bash
# Start both frontend and backend (in separate terminals)
make runs  # Terminal 1 - backend
make runf  # Terminal 2 - frontend

# Or use convenience commands
make dev          # Shows instructions for starting both
make quick-start  # Starts both in background
make killfs       # Stop all services
```

## Project Structure

```
toughradius/
├── cmd/              # Application entry points
├── internal/         # Private application code
│   ├── adminapi/    # New admin API (v9 refactor)
│   ├── app/         # Global application instance (DB, config, tasks)
│   ├── domain/      # Unified data models (all GORM models)
│   ├── radiusd/     # RADIUS service core implementation
│   └── webserver/   # Web server and API handlers
├── pkg/             # Public packages (utilities, crypto, Excel)
├── web/             # React Admin frontend (TypeScript + Vite)
└── docs/           # Documentation
```

## Development Principles

### 1. TDD-Driven Development (测试先行)

**Testing First**: Always write tests before implementation code. This ensures:

- Clear requirements and interface design
- Immediate feedback on implementation correctness
- Built-in regression protection
- Self-documenting code through test cases

**Test Coverage Requirements**:

- Unit tests for all business logic in `internal/`
- Integration tests for API endpoints
- Performance benchmarks for RADIUS protocol handling
- E2E tests for critical user workflows in the frontend

### 2. Minimal Viable Design (最小可用设计原则)

**Simplest Working Solution**: Implement the minimum functionality that satisfies requirements:

- Start with basic implementation, add complexity only when needed
- Prefer standard library solutions over external dependencies
- Build incrementally - each addition should be justified by real requirements
- YAGNI (You Aren't Gonna Need It) - avoid speculative features

**Refactoring Mindset**: Code should be continuously improved:

- Start simple, refactor as requirements evolve
- Maintain clean abstractions but avoid over-engineering
- Design for current needs, not hypothetical future scenarios

### 3. GitHub Workflow Compliance

**Branch Strategy**:

- `main` - always production-ready
- `develop` - integration branch for features
- `feature/*` - isolated feature development
- `hotfix/*` - emergency fixes to production

**Commit Standards**:

- Conventional commit messages: `type(scope): description`
- Types: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`
- Atomic commits - one logical change per commit
- Include tests with feature implementations

**Pull Request Process**:

- All changes must go through PR review
- Automated tests must pass
- Code coverage must not decrease
- PR descriptions must include testing strategy
- No direct pushes to `main` branch

## Key Development Patterns

### Database Access

**Always** use `app.GDB()` to get GORM instance - never inject DB connections directly:

```go
// Correct
user := &domain.RadiusUser{}
app.GDB().Where("username = ?", name).First(user)

// Wrong - don't do this
type Service struct { DB *gorm.DB }
```

Database support: PostgreSQL (production) and SQLite (development with `CGO_ENABLED=1`).

### Vendor-Specific Extensions

RADIUS protocol supports multiple vendors through `VendorCode` field:

- Huawei (2011) - `internal/radiusd/vendors/huawei/`
- Mikrotik (14988) - see `auth_accept_config.go`
- Cisco (9) / Ikuai (10055) / ZTE (3902) / H3C (25506)

Add new vendor support by defining constants in `radius.go` and adding switch cases in `auth_accept_config.go`.

### Configuration Access

Use `app.GApp()` for global configuration and settings:

```go
eapMethod := app.GApp().GetSettingsStringValue("radius", "EapMethod")
maxSessions := app.GApp().GetSettingsInt64Value("radius", "MaxSessions")
```

### Error Handling

RADIUS authentication errors use custom `AuthError` type with metrics tags:

```go
return NewAuthError(app.MetricsRadiusRejectExpire, "user expire")
```

### Logging

Always use zap structured logging with namespace:

```go
zap.L().Error("update user failed",
    zap.Error(err),
    zap.String("namespace", "radius"))
```

### API Route Registration

Add new admin API routes in `internal/adminapi/` and register in `adminapi.go`:

```go
// In adminapi.go Init() function
func Init() {
    registerUserRoutes()  // Add this line
}
```

## Configuration

Main config file: `toughradius.yml`

Key ports:

- RADIUS Auth: 1812 (UDP)
- RADIUS Acct: 1813 (UDP)
- RadSec: 2083 (TCP)
- Web/Admin API: 1816

Environment variables:

- `TOUGHRADIUS_RADIUS_POOL` - RADIUS request pool size (default: 1024)

## Frontend Architecture

- **Framework**: React Admin 5.0 with TypeScript
- **Build Tool**: Vite
- **Testing**: Playwright E2E tests
- **HTTP Client**: ra-data-simple-rest for API communication
- **Charts**: ECharts for dashboard visualizations

API proxy configuration in `vite.config.ts` forwards `/api` requests to `http://localhost:1816`.

## Key Dependencies

**Backend:**

- Echo v4 (web framework)
- GORM v2 (ORM)
- zap (logging)
- layeh.com/radius (RADIUS protocol)
- ants (goroutine pool)

**Frontend:**

- React 18.3.1
- React Admin 5.0
- TypeScript 5.5.3
- Vite 5.3.1
- ECharts 5.5.0

## Testing

- Backend unit tests: `go test ./...`
- Backend benchmarks: `go test -bench=. ./internal/radiusd/`
- Frontend E2E tests: Playwright in `web/` directory

## Database Models

All GORM models are defined in `internal/domain/tables.go`. The database supports automatic migrations through `app.MigrateDB()`.

## Default Credentials

Default admin credentials are shown in initialization logs when running `-initdb`.
