# ToughRADIUS AI Agent Development Guide

## 🤖 AI Agent Working Guidelines

### 🔍 Mandatory Requirement: Understand Existing Code Before Editing

**Before touching any code, thoroughly inspect the relevant implementation and surrounding tests.** Treat code search and context gathering as the very first step of every task.

#### Why This Matters

- ✅ **Precise Targeting** – Quickly locate existing implementations you can extend or reuse
- ✅ **Architectural Awareness** – Learn how modules collaborate before making changes
- ✅ **Consistency** – Mirror naming, data flow, and error-handling patterns already in place
- ✅ **Risk Reduction** – Avoid regressions caused by overlooking hidden dependencies

#### Recommended Search Workflow

1. **Pinpoint the module**: Identify packages, directories, or features involved (e.g., `internal/radiusd/vendors`).
2. **Use repository search tools**: Combine `semantic_search`, `grep_search`, and `file_search` with precise keywords (function names, struct names, RFC identifiers, etc.).
3. **Read surrounding tests**: Open the corresponding `*_test.go` / Playwright specs to understand expected behavior and edge cases.
4. **Record findings**: Jot down key structs, helper functions, or patterns you must follow before implementing anything new.

#### Example Search Prompts

```text
# Before adding new RADIUS vendor support
semantic_search "vendor attribute parsing" in internal/radiusd/vendors
grep_search "VendorCode" --include internal/radiusd/**

# Before adding new API endpoints
file_search "*/internal/adminapi/*routes*.go"
semantic_search "Echo middleware" in internal/webserver

# Before fixing authentication bugs
semantic_search "AuthError" in internal/radiusd
grep_search "app.GDB" --include internal/app/**

# Before large refactors
file_search "*radius_auth*.go"
list_code_usages AuthenticateUser
```

#### Iterative Exploration Pattern

- **Round 1: Macro View** – Scan architecture docs (`docs/v9-architecture.md`), service entry points, and top-level packages.
- **Round 2: Detail Dive** – Read concrete handler/service implementations plus related tests.
- **Round 3: Edge Cases** – Inspect integration tests, benchmarks, or vendor-specific helpers to catch non-obvious behavior.

Always document the insights gained in your task notes or PR description so reviewers know which prior art influenced the change.

### 🧪 Mandatory Requirement: Continuous Verification

**You MUST actively run tests throughout the development lifecycle.**

- **Before changes**: Run existing tests to establish a baseline.
- **During development**: Run relevant tests frequently to catch regressions early.
- **After completion**: Run the full test suite to ensure system integrity.
- **Never assume**: "It should work" is not acceptable. Verify with `go test`.

### 📝 Code is the Best Documentation Principle

#### Core Philosophy: Code IS Documentation, Documentation IS Code

We follow the **"Standard Library Style"** documentation approach. Just as the Go standard library is self-documenting, our codebase should be readable, understandable, and maintainable through comprehensive in-code comments, not separate Markdown files.

**Correct Understanding of "Code Is Documentation":**

- ✅ Documentation exists in the source code as comments
- ✅ Update comments in sync with every code change
- ✅ Full documentation is accessible through `go doc` and IDEs
- ❌ Do not create separate Markdown documents for each module
- ❌ Do not generate documentation files after every development task

#### 1. Write Documentation Directly in Code (Mandatory)

Every exported symbol (function, struct, interface, constant) must have a comprehensive comment **in the source file** that explains **what** it does, **how** to use it, and **why** it behaves that way.

**Standard Library Style Checklist:**

- **Package Comment**: Every package must have a package-level comment explaining its purpose
- **Summary Sentence**: The first sentence should be a concise summary that appears in `go doc`
- **Detailed Description**: Explain the behavior, side effects, and algorithm if necessary
- **Parameter Documentation**: Use bulleted list format with types and constraints
- **Return Value Documentation**: Clearly define what outputs are returned and when
- **Error Handling**: Explicitly state what errors can be returned and under what conditions
- **Usage Examples**: Provide code snippets in comments for complex APIs
- **Side Effects**: Document any state changes, I/O operations, or concurrency implications

#### 2. Documentation Examples

##### Example 1: Package-Level Documentation

```go
// Package radiusd implements the core RADIUS protocol server supporting
// authentication (RFC 2865), accounting (RFC 2866), and RadSec (RFC 6614).
//
// This package provides concurrent request handling using goroutine pools,
// vendor-specific attribute parsing (Huawei, Cisco, Mikrotik), and integration
// with the ToughRADIUS database backend.
//
// Key components:
//   - AuthServer: Handles RADIUS authentication on UDP port 1812
//   - AcctServer: Handles RADIUS accounting on UDP port 1813
//   - RadSecServer: Handles RADIUS over TLS on TCP port 2083
//
// Usage:
//
//	authServer := NewAuthServer(config)
//	go authServer.Start()  // Runs in background
//
// For vendor attribute parsing, see internal/radiusd/vendors/.
package radiusd
```

##### Example 2: Function Documentation

```go
// AuthenticateUser validates user credentials against the RADIUS database.
// It checks username/password, account expiration, and session limits.
//
// The function performs the following validations in order:
//  1. User existence check
//  2. Password verification (supports PAP, CHAP, MS-CHAPv2)
//  3. Account status check (disabled/expired)
//  4. Concurrent session limit check
//
// Parameters:
//   - username: User's login name (case-sensitive, max 255 chars)
//   - password: Plain text password (will be hashed internally using configured algorithm)
//   - nasIP: Network Access Server IP address for session tracking and NAS authorization
//
// Returns:
//   - *domain.RadiusUser: User object if authentication succeeds, contains billing plan info
//   - error: AuthError with metrics tag if validation fails, nil on success
//
// Common errors:
//   - MetricsRadiusRejectUserNotFound: Username doesn't exist in database
//   - MetricsRadiusRejectPasswordError: Password mismatch
//   - MetricsRadiusRejectExpire: Account expired (ExpireTime < now)
//   - MetricsRadiusRejectDisable: Account is disabled (Status != "enabled")
//   - MetricsRadiusRejectConcurrent: Concurrent session limit exceeded
//
// Side effects:
//   - Increments Prometheus metrics counter for auth attempts
//   - Logs authentication result to zap logger with namespace "radius"
//
// Concurrency: Safe for concurrent use. Database queries use GORM connection pool.
//
// Example:
//
//	user, err := AuthenticateUser("john@example.com", "secret123", "192.168.1.1")
//	if err != nil {
//	    if errors.Is(err, ErrUserNotFound) {
//	        log.Error("user not found", zap.Error(err))
//	    }
//	    return err
//	}
//	log.Info("auth success", zap.String("username", user.Username))
func AuthenticateUser(username, password, nasIP string) (*domain.RadiusUser, error) {
    // Implementation
}
```

##### Example 3: Struct Documentation

```go
// RadiusUser represents a user account in the RADIUS authentication system.
// It maps to the "radius_user" table in the database and is managed through
// the Admin API and web interface.
//
// This struct holds all user-specific configuration including authentication
// credentials, billing plan reference, account status, and session limits.
//
// Database table: radius_user
// GORM features: Auto-migration, soft delete (DeletedAt), timestamps
//
// Lifecycle:
//  1. Created via Admin API POST /api/v1/users
//  2. Authenticated during RADIUS Access-Request
//  3. Disabled/Expired based on Status and ExpireTime fields
//  4. Soft-deleted when removed (can be recovered)
//
// Concurrency: GORM handles database-level locking. Application code should
// use transactions when updating user and session data together.
type RadiusUser struct {
    // ID is the auto-incrementing primary key.
    // Generated by database on INSERT, immutable after creation.
    ID int64 `json:"id" gorm:"primaryKey"`

    // Username is the login name used for RADIUS authentication.
    // Must be unique across the system (enforced by unique index).
    // Format: Usually email-like (user@realm) or simple username.
    // Constraints: Non-null, max 255 characters, case-sensitive.
    Username string `json:"username" gorm:"uniqueIndex;not null;size:255"`

    // Password is the hashed password for authentication.
    // Stored in plaintext (legacy) or bcrypt hash depending on config.
    // When using PAP, compared directly. For CHAP/MS-CHAPv2, must be plaintext.
    // Security: Consider migrating to bcrypt for PAP-only deployments.
    Password string `json:"password" gorm:"not null"`

    // Status indicates the user's account status.
    // Possible values: "enabled", "disabled", "expired"
    // Default: "enabled"
    // Authentication is rejected if Status != "enabled".
    Status string `json:"status" gorm:"default:'enabled';size:20"`

    // ExpireTime is the account expiration timestamp.
    // If current time > ExpireTime, authentication is rejected with MetricsRadiusRejectExpire.
    // Zero value means no expiration.
    ExpireTime time.Time `json:"expire_time"`

    // ProductId references the billing plan (radius_product table).
    // Foreign key relationship (not enforced by DB, handled in application).
    // Used to determine bandwidth limits, session timeout, and other policies.
    ProductId int64 `json:"product_id" gorm:"index"`

    // ... other fields

    // CreatedAt is automatically set by GORM on INSERT.
    CreatedAt time.Time `json:"created_at"`

    // UpdatedAt is automatically updated by GORM on UPDATE.
    UpdatedAt time.Time `json:"updated_at"`

    // DeletedAt enables GORM soft delete. Non-null means record is deleted.
    // Soft-deleted records are excluded from queries unless Unscoped() is used.
    DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
```

##### Example 4: Interface Documentation

```go
// Authenticator defines the interface for RADIUS authentication protocol implementations.
// Different implementations support various authentication methods defined in RFC 2865
// and related RFCs (PAP, CHAP, MS-CHAPv2, EAP).
//
// Implementations must be stateless and safe for concurrent use, as the same
// instance may be called from multiple goroutines handling different RADIUS requests.
//
// Standard implementations:
//   - PAPAuthenticator: Plain-text password (RFC 2865 Section 5.2)
//   - CHAPAuthenticator: Challenge-Response (RFC 2865 Section 5.3)
//   - MSCHAPv2Authenticator: Microsoft CHAP version 2 (RFC 2759)
//   - EAPAuthenticator: Extensible Authentication Protocol (RFC 3748)
//
// Example:
//
//	var auth Authenticator = NewPAPAuthenticator()
//	if valid, err := auth.CheckPassword(user, "secret123"); err != nil {
//	    return err
//	} else if !valid {
//	    return ErrPasswordMismatch
//	}
type Authenticator interface {
    // CheckPassword verifies the provided password against the stored user credentials.
    //
    // The verification method depends on the authentication protocol:
    //   - PAP: Direct comparison or bcrypt hash verification
    //   - CHAP: Challenge-response validation using MD5
    //   - MS-CHAPv2: NT hash verification with challenge/response
    //
    // Parameters:
    //   - user: The user object retrieved from database, must not be nil
    //   - password: The password or response from RADIUS Access-Request packet
    //
    // Returns:
    //   - bool: True if password matches, false if password is incorrect
    //   - error: Non-nil if verification process fails (e.g., invalid hash format,
    //            missing challenge attribute). Returns nil for simple mismatch.
    //
    // Concurrency: Must be safe for concurrent calls.
    //
    // Example:
    //
    //	valid, err := auth.CheckPassword(user, "plaintext_password")
    //	if err != nil {
    //	    return fmt.Errorf("password check failed: %w", err)
    //	}
    //	if !valid {
    //	    return ErrAuthenticationFailed
    //	}
    CheckPassword(user *domain.RadiusUser, password string) (bool, error)
}
```

#### 3. Inline Comments for "Why" Not "What"

Use inline comments to explain **why** a specific implementation choice was made, especially for complex logic, workarounds, or optimizations. Do not explain **what** the code is doing if it's obvious from reading the code itself.

**Good inline comments explain:**

- Non-obvious business logic
- Performance optimizations
- Workarounds for vendor quirks
- Protocol-specific requirements
- Edge cases and their handling

```go
// ✅ Correct: Explain the "why" and context
// Huawei devices expect bandwidth in Kbps, but our billing plan stores it in Mbps.
// We multiply by 1024 (binary) instead of 1000 (decimal) to match Huawei's VSA spec.
// See: RFC 2548 Section 6.2 and internal/radiusd/vendors/huawei/README.md
return baseBandwidth * 1024

// ✅ Correct: Explain non-obvious behavior
// PostgreSQL GORM driver has a bug with empty IN clauses.
// Explicitly return empty result to avoid SQL syntax error.
if len(ids) == 0 {
    return []domain.RadiusUser{}, nil
}

// ✅ Correct: Explain performance consideration
// Using goroutine pool instead of unlimited goroutines to prevent
// memory exhaustion under high request rate (>10k req/s observed in prod).
s.taskPool.Submit(func() { handleRequest(req) })

// ❌ Wrong: Explain the "what" (redundant)
// Multiply by 1024
return baseBandwidth * 1024

// ❌ Wrong: Obvious from code
// Set user status to enabled
user.Status = "enabled"

// ❌ Wrong: Commented-out code (use git history instead)
// return baseBandwidth * 1000  // Old calculation
return baseBandwidth * 1024
```

#### 4. Vendor-Specific Code Must Document Protocol Details

Vendor-specific implementations must reference the authoritative specification
and explain the data format, encoding rules, and any quirks or workarounds.

```go
// ParseHuaweiInputAverageRate extracts downstream bandwidth limit from Huawei VSA attribute.
//
// Huawei VSA attribute format (based on RFC 2865 Vendor-Specific attribute):
//   Vendor-Type: 11 (Huawei-Input-Average-Rate)
//   Vendor-Length: Variable (typically 4-8 bytes)
//   Vendor-Data: Unsigned integer representing bandwidth in Kbps
//
// Examples:
//   - Value 1024 = 1 Mbps downstream limit
//   - Value 10240 = 10 Mbps downstream limit
//   - Value 102400 = 100 Mbps downstream limit
//
// Note: Huawei devices use binary Kbps (1024-based) not decimal (1000-based).
// This differs from Cisco which uses decimal Kbps.
//
// Parameters:
//   - attr: RADIUS attribute with Vendor-ID = 2011 (Huawei)
//
// Returns:
//   - int64: Bandwidth limit in Kbps, 0 if attribute is malformed
//
// References:
//   - Huawei VSA Dictionary: internal/radiusd/vendors/huawei/README.md
//   - RFC 2865 Section 5.26: Vendor-Specific Attribute Format
func ParseHuaweiInputAverageRate(attr *radius.Attribute) int64 {
    if attr == nil || len(attr.Value) < 4 {
        return 0
    }

    // Huawei stores bandwidth as 32-bit big-endian unsigned integer
    return int64(binary.BigEndian.Uint32(attr.Value))
}
```

#### Documentation Strategy: Code-First Approach

**When to Write In-Code Comments (Always Required):**

Every code file should be self-documenting. Comments are NOT optional.

- ✅ **Package comment** - First file in each package must have package doc
- ✅ **All exported symbols** - Functions, methods, types, constants, variables
- ✅ **Complex algorithms** - Step-by-step explanation of non-trivial logic
- ✅ **Non-obvious design decisions** - Why this approach was chosen
- ✅ **Protocol implementations** - Reference RFC numbers, vendor specs
- ✅ **Performance-critical code** - Explain optimizations and trade-offs
- ✅ **Error conditions** - What errors can occur and why
- ✅ **Concurrency guarantees** - Thread-safety, locking, goroutine usage
- ✅ **Side effects** - Database writes, I/O, metric updates, logging

**When to Create Separate Markdown Documentation (Rare):**

Only create separate docs when in-code comments are insufficient for the target audience.

- ✅ **Architecture Overview** (`docs/architecture.md`) - High-level system design for newcomers
- ✅ **User Guides** (`README.md`, `docs/deployment.md`) - For end users, not developers
- ✅ **API Contracts** (`docs/api-spec.md`) - For external integrators (REST API, gRPC)
- ✅ **Protocol References** (`docs/rfcs/`) - Standard specifications (read-only)
- ✅ **Migration Guides** (`docs/migration-v8-to-v9.md`) - Breaking changes between versions
- ❌ **NOT for module documentation** - Put it in the code as package/function comments
- ❌ **NOT for feature descriptions** - Describe in code comments + Git commits
- ❌ **NOT for work summaries** - Use Git history and PR descriptions

**When to Update Existing Docs (Only When Necessary):**

- ✅ **Public API changes** → Update code comments first, then external API docs if needed
- ✅ **Configuration options** → Update example config files and inline comments
- ✅ **Breaking changes** → Update CHANGELOG.md + migration guide
- ✅ **Deployment process** → Update deployment docs
- ❌ **Internal refactoring** → Only update code comments, no separate doc needed
- ❌ **Bug fixes** → Git commit message is enough, no doc update
- ❌ **Performance improvements** → Update inline comments if algo changed

**Documentation Verification:**

Use `go doc` to verify your documentation is accessible:

```bash
# View package documentation
go doc internal/radiusd

# View specific function documentation
go doc internal/radiusd.AuthenticateUser

# Generate HTML documentation for entire project
godoc -http=:6060
# Visit http://localhost:6060/pkg/github.com/talkincode/toughradius/
```

#### Auto-Generated Documentation Rule (Strictly Prohibited)

**AI Agent must NOT create separate documentation files unless explicitly requested.**

The correct workflow is:

1. Write comprehensive **in-code comments** for all changes
2. Write clear **Git commit messages** explaining what and why
3. Update **existing docs** only if public API/config/deployment changed
4. Give brief completion confirmation in chat (1-2 sentences max)

**Prohibited Actions:**

- ❌ **Auto-creating** `SUMMARY.md`, `WORK_REPORT.md`, `DOCUMENTATION.md` after task completion
- ❌ **Auto-creating** module-specific doc files like `module-name-guide.md`
- ❌ **Adding lengthy summaries** in chat responses (multi-paragraph reports)
- ❌ **Duplicating information** that already exists in code comments or Git history
- ❌ **Creating "documentation"** as a separate deliverable for internal features

**Allowed Actions:**

- ✅ **In-code comments** - Always required, this IS the documentation
- ✅ **Git commit messages** - Explain what changed and why (Conventional Commits format)
- ✅ **Brief completion note** - 1-2 sentences confirming what was done
- ✅ **Update existing docs** - If public API, config, or deployment process changed
- ✅ **Create docs when requested** - User explicitly asks for a guide or tutorial

**Correct Completion Response:**

```text
✅ Added Cisco vendor support with comprehensive in-code documentation and tests (98% coverage).
```

**Incorrect Completion Response:**

```text
## Cisco Vendor Implementation Summary ❌

### Overview
This document describes the implementation of Cisco vendor support...

### Architecture
The Cisco vendor module is located in...

### API Reference
...

(This is all redundant - should be in code comments and Git commits!)
```

**What Reviewers Should See:**

```bash
# 1. Clear code comments (primary documentation)
$ go doc internal/radiusd/vendors/cisco
package cisco // import "github.com/talkincode/toughradius/v9/internal/radiusd/vendors/cisco"

Package cisco implements Cisco RADIUS Vendor-Specific Attribute (VSA) parsing...

func ParseCiscoAVPair(attr *radius.Attribute) map[string]string
    ParseCiscoAVPair extracts key-value pairs from Cisco VSA attribute...

# 2. Git history (change log)
$ git log --oneline
a1b2c3d feat(radius): add Cisco vendor VSA parsing support
d4e5f6g test(radius): add Cisco vendor parsing tests (98% coverage)
g7h8i9j docs(radius): update vendor support list in README

# 3. No separate markdown files polluting the repo
$ ls docs/
architecture.md  deployment.md  api-integration.md  rfcs/
# NOT: cisco-vendor-guide.md, work-summary.md, etc.
```

#### Documentation Quality Standards

**In-Code Documentation Must Be:**

1. **Clear** - Use simple, plain language; avoid jargon unless explaining technical protocols
2. **Complete** - Document all parameters, returns, errors, side effects, concurrency
3. **Practical** - Include real-world usage examples for complex APIs
4. **Accurate** - Keep code and comments in sync (update comments when code changes)
5. **Concise** - Avoid redundant explanations of obvious code
6. **Discoverable** - Structured so `go doc` and IDEs can parse and display properly

**Documentation Review Checklist:**

Before submitting a PR, verify each exported symbol has:

- [ ] Package comment explaining purpose and key components
- [ ] Summary sentence (first line, appears in `go doc` listings)
- [ ] Parameter documentation with types and constraints
- [ ] Return value documentation with possible values
- [ ] Error conditions explicitly listed
- [ ] Side effects documented (DB writes, metrics, logs, etc.)
- [ ] Concurrency guarantees stated (thread-safe or not)
- [ ] Usage example for non-trivial functions
- [ ] References to RFCs or specs for protocol implementations

##### Example: High-Quality In-Code Documentation

```go
// GetUserOnlineSessions retrieves all active RADIUS sessions for a user.
//
// This function queries the accounting table for sessions with null stop time,
// indicating they are still active. It's used to enforce MaxSessions limits.
//
// Parameters:
//   - username: User's login name (exact match, case-sensitive)
//
// Returns:
//   - []*domain.RadiusAccounting: Slice of active session records (empty if none)
//   - error: Database error if query fails (nil on success)
//
// Performance: Uses index on (username, acct_stop_time) for fast lookup.
// For users with >1000 sessions, consider pagination.
//
// Example:
//   sessions, err := GetUserOnlineSessions("john@example.com")
//   if err != nil {
//       return fmt.Errorf("failed to check sessions: %w", err)
//   }
//   if len(sessions) >= maxSessions {
//       return ErrMaxSessionsExceeded
//   }
func GetUserOnlineSessions(username string) ([]*domain.RadiusAccounting, error) {
    var sessions []*domain.RadiusAccounting
    err := app.GDB().Where("username = ? AND acct_stop_time IS NULL", username).
        Find(&sessions).Error
    return sessions, err
}
```

**Golden Rules for "Code IS Documentation":**

- 📝 **In-code comments are THE documentation** - Not supplementary, not optional
- 🎯 **Write for readers** - Future self, teammates, open-source contributors
- 🚫 **No redundant Markdown files** - Don't duplicate what's in code comments
- ✅ **Git history is the changelog** - Commit messages record what/when/why
- 🔍 **Verify with `go doc`** - If it doesn't show up in `go doc`, add more comments
- ♻️ **Update comments when code changes** - Comments are part of the code, not separate
- 📚 **Separate docs only for end users** - Deployment guides, user manuals, API contracts

**Anti-Pattern Recognition:**

```text
❌ Wrong Workflow:
1. Write code
2. Code works
3. Create DOCUMENTATION.md to explain the code
4. PR includes both code files and doc files

✅ Correct Workflow:
1. Write code WITH comprehensive comments
2. Code AND comments work together
3. Verify with `go doc package.Function`
4. PR includes only code files with excellent comments
5. Git commit message explains what/why
```

## Core Development Principles

This project **strictly follows** these three core development principles. All code contributions must comply with these standards:

### 🧪 Test-Driven Development (TDD)

#### Mandatory Requirement: Write Tests First, Then Code

Before implementing any feature or fixing any bug, you **MUST** write a test case that reproduces the issue or defines the expected behavior.

1. **Create Test File**: If `internal/radiusd/auth.go` is being modified, create/edit `internal/radiusd/auth_test.go`.
2. **Define Test Case**: Write a test function `TestAuth_UserNotFound` that asserts the expected failure.
3. **Run Test (Fail)**: Execute the test to confirm it fails (Red).
4. **Implement Code**: Write the minimum code necessary to pass the test.
5. **Run Test (Pass)**: Execute the test to confirm it passes (Green).
6. **Refactor**: Clean up the code while keeping tests passing.

#### TDD Workflow

1. **Red Phase** - Write failing tests

   ```bash
   # Create test file
   touch internal/radiusd/new_feature_test.go

   # Run tests (should fail)
   go test ./internal/radiusd/new_feature_test.go -v
   ```

2. **Green Phase** - Write minimal implementation to pass tests

   ```bash
   # Implement feature code
   vim internal/radiusd/new_feature.go

   # Run tests again (should pass)
   go test ./internal/radiusd/new_feature_test.go -v
   ```

3. **Refactor Phase** - Optimize code while keeping tests passing

   ```bash
   # Continuously run tests to ensure safe refactoring
   go test ./... -v
   ```

#### 🔄 Continuous Verification (Mandatory)

**Do not wait until the end to run tests.**

- **Every logical change** should be followed by a test run.
- **If a test fails**, stop and fix it immediately. Do not pile up changes on top of broken tests.
- **Use `go test ./...`** frequently to check for side effects in other packages.
- **Verify compilation** with `go build ./...` to ensure no syntax errors or type mismatches were introduced.

#### Test Coverage Requirements

- **New feature code coverage must be ≥ 80%**
- **Core RADIUS protocol module coverage must be ≥ 90%**
- **Critical business logic must have integration tests**

```bash
# Check test coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# View coverage statistics
go test ./internal/radiusd -coverprofile=coverage.out
go tool cover -func=coverage.out
```

#### Test File Organization

```text
internal/radiusd/
├── auth_passwd_check.go          # Implementation file
├── auth_passwd_check_test.go     # Unit tests (same package)
├── radius_auth.go
├── radius_test.go                # Integration tests
└── vendor_parse_test.go          # Feature tests
```

#### Test Case Naming Convention

```go
// ✅ Correct: Clearly describe test intent
func TestAuthPasswordCheck_ValidUser_ShouldReturnSuccess(t *testing.T) {}
func TestAuthPasswordCheck_ExpiredUser_ShouldReturnError(t *testing.T) {}
func TestGetNas_UnauthorizedIP_ShouldReturnAuthError(t *testing.T) {}

// ❌ Wrong: Ambiguous
func TestAuth(t *testing.T) {}
func TestFunc1(t *testing.T) {}
```

#### Table-Driven Tests

For multi-scenario testing, use table-driven approach:

```go
func TestVendorParse(t *testing.T) {
    tests := []struct {
        name       string
        vendorCode string
        input      string
        wantMac    string
        wantVlan1  int64
    }{
        {"Huawei VLAN", VendorHuawei, "vlan=100", "", 100},
        {"Mikrotik MAC", VendorMikrotik, "mac=aa:bb:cc:dd:ee:ff", "aa:bb:cc:dd:ee:ff", 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test logic
        })
    }
}
```

### 🔄 GitHub Workflow

#### Mandatory Requirement: Follow Git Flow branching model and standard PR process

#### Branching Strategy

```text
main (production branch)
  ├── v9dev (development branch)
  │    ├── feature/user-management     # Feature branch
  │    ├── feature/radius-vendor-cisco # Feature branch
  │    ├── bugfix/session-timeout      # Bug fix
  │    └── hotfix/security-patch       # Hotfix
  └── release/v9.1.0                   # Release branch
```

#### Standard Development Process

##### 1. Create Feature Branch

```bash
# Create feature branch from v9dev
git checkout v9dev
git pull origin v9dev
git checkout -b feature/add-cisco-vendor

# Branch naming convention
# feature/  - New features
# bugfix/   - Bug fixes
# hotfix/   - Hotfixes
# refactor/ - Code refactoring
# docs/     - Documentation updates
```

##### 2. TDD Loop Development

```bash
# 1️⃣ Write tests first
vim internal/radiusd/vendors/cisco/cisco_test.go

# 2️⃣ Run tests (red)
go test ./internal/radiusd/vendors/cisco -v

# 3️⃣ Implement feature
vim internal/radiusd/vendors/cisco/cisco.go

# 4️⃣ Run tests (green)
go test ./internal/radiusd/vendors/cisco -v

# 5️⃣ Commit atomic changes
git add internal/radiusd/vendors/cisco/
git commit -m "test: add Cisco vendor attribute parsing tests"
git commit -m "feat: implement Cisco vendor attribute parsing"
```

##### 3. Commit Convention (Conventional Commits)

```bash
# Format: <type>(<scope>): <subject>
git commit -m "feat(radius): add Cisco vendor support"
git commit -m "test(radius): add unit tests for Cisco attributes"
git commit -m "fix(auth): correct password validation logic"
git commit -m "docs(api): update RADIUS authentication API docs"
git commit -m "refactor(database): optimize user query performance"
git commit -m "perf(radius): reduce authentication latency by 20%"

# Type definitions
# feat:     New features
# fix:      Bug fixes
# test:     Testing related
# docs:     Documentation updates
# refactor: Code refactoring
# perf:     Performance optimization
# style:    Code formatting
# chore:    Build/tool changes
```

##### 4. Create Pull Request

PR must include:

- ✅ **All tests passing** (`go test ./...`)
- ✅ **Code coverage meets requirements**
- ✅ **Clear description and change summary**
- ✅ **Associated Issue number**
- ✅ **At least one code reviewer approval**

PR Template:

```markdown
## Change Description

Brief description of the purpose and main changes of this PR

## Change Type

- [ ] New feature
- [ ] Bug fix
- [ ] Performance optimization
- [ ] Code refactoring
- [ ] Documentation update

## Test Coverage

- [ ] Added unit tests
- [ ] Added integration tests
- [ ] Test coverage ≥ 80%
- [ ] All tests passing

## Checklist

- [ ] Code follows project conventions
- [ ] Commit messages follow Conventional Commits
- [ ] Updated relevant documentation
- [ ] No breaking changes (or documented in CHANGELOG)

## Related Issue

Closes #123
```

##### 5. Continuous Integration Checks

Each PR automatically triggers:

- ✅ `go test ./...` - Run all tests
- ✅ `go build` - Ensure code compiles
- ✅ Docker image build
- ✅ Code style checks

#### Release Process

```bash
# 1. Create release branch
git checkout -b release/v9.1.0 v9dev

# 2. Update version and CHANGELOG
vim VERSION
vim CHANGELOG.md

# 3. Merge to main and tag
git checkout main
git merge --no-ff release/v9.1.0
git tag -a v9.1.0 -m "Release version 9.1.0"
git push origin main --tags

# 4. Trigger GitHub Actions auto-release
# - Build AMD64/ARM64 binaries
# - Publish Docker images to DockerHub and GHCR
# - Create GitHub Release
```

### 📦 Minimum Viable Product (MVP) Principle

#### Mandatory Requirement: Each feature must be delivered in minimum viable units

#### MVP Design Method

1. **Identify Core Value**

   - ❓ What problem does this feature solve?
   - ❓ What is the simplest implementation?
   - ❓ What is essential vs. nice-to-have?

2. **Feature Breakdown Example**

   ```text
   ❌ Wrong approach: Implement complete feature at once
   Issue #123: Add Cisco vendor support
   └── Includes auth, accounting, VSA attributes, config management, Web UI...

   ✅ Correct approach: MVP breakdown
   Issue #123: Add Cisco vendor basic auth support (MVP-1)
   ├── PR #124: Cisco VSA attribute parsing
   ├── PR #125: Cisco auth flow integration
   └── PR #126: Basic test cases

   Issue #130: Extend Cisco accounting features (MVP-2)
   └── Built on MVP-1

   Issue #135: Add Cisco Web management UI (MVP-3)
   └── Built on MVP-1 + MVP-2
   ```

3. **MVP Delivery Standards**

   Each MVP must be:

   - ✅ **Independently Usable** - Does not depend on incomplete features
   - ✅ **Fully Tested** - Coverage meets requirements
   - ✅ **Well Documented** - API docs + usage guide
   - ✅ **Demonstrable** - Can run and show value
   - ✅ **Rollback-Safe** - Does not break existing functionality

#### MVP Practice Examples

##### Example 1: Adding RADIUS Vendor Support

```text
MVP-1 (Week 1): Basic attribute parsing ✅
├── vendor_cisco.go          # Vendor constant definitions
├── vendor_cisco_test.go     # Parsing tests
└── Support reading basic VSA attributes

MVP-2 (Week 2): Authentication integration ✅
├── auth_accept_config.go    # Add Cisco case
├── auth_cisco_test.go       # Auth integration tests
└── Support Cisco device auth flow

MVP-3 (Week 3): Accounting support ✅
└── Extend accounting records with Cisco-specific fields

MVP-4 (Week 4): Web management ✅
└── Admin API adds Cisco configuration UI
```

##### Example 2: Performance Optimization

```text
MVP-1: Identify bottlenecks ✅
├── Add performance test benchmarks
├── Identify hotspot functions
└── Establish performance baseline

MVP-2: Optimize database queries ✅
├── Add indexes
├── Optimize N+1 queries
└── Verify 20% performance improvement

MVP-3: Caching layer ✅ (optional)
└── Continue if performance still not meeting targets
```

## Complete Development Workflow Example

### Scenario: Adding New RADIUS Vendor Support (Cisco)

#### Step 1: Create Issue (Requirements Analysis)

```markdown
Title: [Feature] Add Cisco RADIUS Vendor Support

## MVP-1 Scope

- [ ] Parse Cisco VSA attributes
- [ ] Unit test coverage ≥ 90%
- [ ] Documentation update

## MVP-2 Scope (Future)

- [ ] Authentication flow integration
- [ ] Accounting feature support

## Not Included

- Web management UI (MVP-3)
- Advanced configuration management (MVP-4)
```

#### Step 2: TDD Development

```bash
# 1️⃣ Create branch
git checkout -b feature/cisco-vendor-mvp1 v9dev

# 2️⃣ Write tests first (red)
cat > internal/radiusd/vendors/cisco/cisco_test.go << 'EOF'
package cisco

import "testing"

func TestParseCiscoAVPair(t *testing.T) {
    tests := []struct{
        name  string
        input string
        want  map[string]string
    }{
        {"basic", "key=value", map[string]string{"key": "value"}},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := ParseAVPair(tt.input)
            // Assertion logic
        })
    }
}
EOF

# 3️⃣ Run tests (should fail)
go test ./internal/radiusd/vendors/cisco -v

# 4️⃣ Implement minimal code (green)
cat > internal/radiusd/vendors/cisco/cisco.go << 'EOF'
package cisco

func ParseAVPair(input string) map[string]string {
    // Minimal implementation
    return map[string]string{}
}
EOF

# 5️⃣ Run tests (should pass)
go test ./internal/radiusd/vendors/cisco -v

# 6️⃣ Refactor and optimize
# Improve implementation while keeping tests passing

# 7️⃣ Check coverage
go test ./internal/radiusd/vendors/cisco -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total
```

#### Step 3: Commit Code

```bash
# Atomic commits
git add internal/radiusd/vendors/cisco/cisco_test.go
git commit -m "test(radius): add Cisco AVPair parsing tests"

git add internal/radiusd/vendors/cisco/cisco.go
git commit -m "feat(radius): implement Cisco AVPair parsing (MVP-1)"

git add docs/radius/cisco-vendor.md
git commit -m "docs(radius): add Cisco vendor documentation"
```

#### Step 4: Create Pull Request

```bash
git push origin feature/cisco-vendor-mvp1
# Create PR on GitHub, fill in PR template
```

#### Step 5: Code Review and Merge

- Wait for CI to pass
- Code review feedback
- Fix issues, push updates
- Merge to v9dev after approval

#### Step 6: Plan MVP-2

- Create new Issue for next MVP
- Repeat the above process

## Quality Gates

All code must pass before merging:

### ✅ Automated Checks

- [ ] All unit tests pass (`go test ./...`)
- [ ] Code coverage ≥ 80%
- [ ] No compilation errors (`go build`)
- [ ] Docker image builds successfully
- [ ] Frontend tests pass (`npm run test`)

### ✅ Code Review

- [ ] At least one maintainer approval
- [ ] No unresolved review comments
- [ ] Follows code conventions

### ✅ Documentation Requirements (Code-First Approach)

- [ ] **All exported APIs have comprehensive comments** (mandatory)
  - Function/method purpose clearly explained
  - All parameters documented with types and constraints
  - Return values and error conditions described
  - Real-world usage examples included (for complex APIs)
- [ ] **Complex logic has inline comments** explaining the "why"
- [ ] **Vendor-specific code references protocol specs** (RFC numbers, VSA docs)
- [ ] **API behavior changes** → Update existing API documentation only if external-facing
- [ ] **Breaking changes** → CHANGELOG.md updated with migration guide
- [ ] **No redundant separate documentation** unless explicitly required

### ✅ MVP Acceptance

- [ ] Feature independently usable
- [ ] Meets minimum requirements
- [ ] Does not introduce technical debt

## Common Anti-Patterns (Prohibited)

### ❌ Anti-Pattern 1: Exporting APIs Without Documentation

```go
// ❌ Wrong: Exported function with no comment
func AuthenticateUser(username, password, nasIP string) (*domain.RadiusUser, error) {
    // Implementation
}

// ✅ Correct: Comprehensive API documentation
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
func AuthenticateUser(username, password, nasIP string) (*domain.RadiusUser, error) {
    // Implementation
}
```

### ❌ Anti-Pattern 2: Committing Without Tests

```bash
# Wrong example
git commit -m "feat: add new feature"  # No corresponding test file

# Correct approach
git commit -m "test: add tests for new feature"
git commit -m "feat: add new feature"
```

### ❌ Anti-Pattern 3: Giant PRs

```text
❌ PR #100: Complete user management system implementation
   +2000 -500 lines across 50 files

✅ Split into:
   PR #101: User model and database migration (MVP-1)
   PR #102: User CRUD API endpoints (MVP-2)
   PR #103: User management UI (MVP-3)
```

### ❌ Anti-Pattern 4: Implementation Before Tests

```go
// ❌ Wrong flow
1. Implement complete feature
2. Feature becomes complex
3. Difficult to add tests
4. Insufficient test coverage

// ✅ TDD flow
1. Write tests (define behavior)
2. Minimal implementation
3. Refactor and optimize
4. Test coverage naturally achieved
```

### ❌ Anti-Pattern 5: Skipping Code Review

```bash
# ❌ Direct push to main branch
git push origin main  # Rejected by protected branch

# ✅ Through PR process
git push origin feature/my-feature
# Create PR → CI checks → Code review → Merge
```

### ❌ Anti-Pattern 6: Creating Redundant Documentation

```bash
# ❌ Wrong: Auto-generating summary docs after each task
feature-implementation.go  # Implementation
feature-implementation-summary.md  # Redundant - info should be in code comments
WORK_REPORT.md  # Redundant - info should be in Git commits

# ✅ Correct: Code + Comments + Git History
feature-implementation.go  # Implementation with comprehensive comments
# Git commit messages record the what/why/when
# No separate summary document needed
```

### ❌ Anti-Pattern 7: Introducing CGO Dependencies

```go
// ❌ Wrong: Importing a library that requires CGO
import "github.com/mattn/go-sqlite3"

// ✅ Correct: Using a pure Go alternative
import "github.com/glebarez/sqlite"
```

## Tool Configuration

### Local Development Environment Setup

```bash
# Install Git hooks (automated testing)
cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
echo "Running tests..."
export CGO_ENABLED=0
go test ./...
if [ $? -ne 0 ]; then
    echo "❌ Tests failed, commit blocked"
    exit 1
fi
echo "✅ Tests passed"
EOF
chmod +x .git/hooks/pre-commit

# Configure commit template
git config commit.template .gitmessage.txt
```

### Recommended VS Code Extensions

- **Go** - Go language support
- **Go Test Explorer** - Test visualization
- **Coverage Gutters** - Coverage display
- **Conventional Commits** - Commit convention helper
- **GitLens** - Git enhancement

## References

- [TDD Practice Guide](https://martinfowler.com/bliki/TestDrivenDevelopment.html)
- [Git Flow Workflow](https://nvie.com/posts/a-successful-git-branching-model/)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [MVP Methodology](https://www.agilealliance.org/glossary/mvp/)
- [Go Testing Best Practices](https://go.dev/doc/tutorial/add-a-test)

---

**Remember: Quality over speed, Usable over perfect, Tests before code, Code is the documentation!**

**Documentation Hierarchy:**

1. 🥇 **Code + Comprehensive Comments** - First-class documentation
2. 🥈 **Git Commit History** - Records what/why/when
3. 🥉 **Minimal Separate Docs** - Only for architecture & external APIs

### 🚫 Technical Constraints

#### Mandatory Requirement: No CGO Dependencies

This project is designed to be cross-platform and easily deployable.

- **Strictly Forbidden**: Introducing any library that requires `CGO_ENABLED=1`.
- **Database Drivers**: Use pure Go drivers only (e.g., `glebarez/sqlite` instead of `mattn/go-sqlite3`).
- **Build Flags**: Always ensure `CGO_ENABLED=0` is set in build scripts and CI configurations.
