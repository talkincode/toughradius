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

### 📝 Code is the Best Documentation Principle

#### Core Philosophy: Treat Code as the Primary Source of Truth

We follow the **"Standard Library Style"** documentation approach. Just as the Go standard library is self-documenting, our codebase should be readable, understandable, and maintainable without relying on external documents.

#### 1. Write Documentation Like the Standard Library

Every exported symbol (function, struct, interface, constant) must have a comment that explains **what** it does, **how** to use it, and **why** it behaves that way.

**Standard Library Style Checklist:**

- **Summary Sentence**: The first sentence should be a concise summary of the function's purpose.
- **Detailed Description**: Explain the behavior, side effects, and algorithm if necessary.
- **Parameter & Return Value Documentation**: Clearly define what inputs are expected and what outputs are returned.
- **Error Handling**: Explicitly state what errors can be returned and under what conditions.
- **Usage Examples**: Provide code snippets for complex APIs.

#### 2. Documentation Examples

##### Example 1: Function Documentation

```go
// AuthenticateUser validates user credentials against the RADIUS database.
// It checks username/password, account expiration, and session limits.
//
// If authentication succeeds, it returns the user object.
// If authentication fails, it returns an error with a specific metrics tag.
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
//   - MetricsRadiusRejectExpire: Account expired
//
// Example:
//   user, err := AuthenticateUser("john", "secret123", "192.168.1.1")
//   if err != nil {
//       log.Error("auth failed", zap.Error(err))
//       return err
//   }
func AuthenticateUser(username, password, nasIP string) (*domain.RadiusUser, error) {
    // Implementation
}
```

##### Example 2: Struct Documentation

```go
// RadiusUser represents a user in the RADIUS system.
// It maps to the "radius_user" table in the database.
//
// This struct holds all user-specific configuration, including
// authentication credentials, billing information, and network policies.
type RadiusUser struct {
    // ID is the unique identifier for the user.
    ID int64 `json:"id" gorm:"primaryKey"`

    // Username is the login name used for RADIUS authentication.
    // It must be unique across the system.
    Username string `json:"username" gorm:"uniqueIndex;not null"`

    // Status indicates the user's account status.
    // Possible values: "enabled", "disabled", "expire".
    Status string `json:"status" gorm:"default:'enabled'"`

    // ... other fields
}
```

##### Example 3: Interface Documentation

```go
// Authenticator defines the interface for user authentication strategies.
// Different implementations can support PAP, CHAP, MS-CHAPv2, or EAP.
type Authenticator interface {
    // CheckPassword verifies the provided password against the stored credentials.
    //
    // Parameters:
    //   - user: The user object retrieved from the database.
    //   - password: The password provided in the RADIUS request.
    //
    // Returns:
    //   - bool: True if the password matches, false otherwise.
    //   - error: Any error that occurred during verification (e.g., hashing failure).
    CheckPassword(user *domain.RadiusUser, password string) (bool, error)
}
```

#### 3. Inline Comments for "Why"

Use inline comments to explain **why** a specific implementation choice was made, especially for complex logic, workarounds, or optimizations. Do not explain **what** the code is doing if it's obvious.

```go
// ✅ Correct: Explain the "why"
// Huawei devices expect bandwidth in Kbps, but our plan stores it in Mbps.
// We multiply by 1024 (binary) instead of 1000 (decimal) to match vendor specs.
return baseBandwidth * 1024

// ❌ Wrong: Explain the "what"
// Multiply by 1024
return baseBandwidth * 1024
```

#### 4. Vendor-Specific Code Must Document Protocol Details

```go
// ParseHuaweiInputAverageRate extracts bandwidth limit from Huawei VSA attribute.
//
// Huawei-Input-Average-Rate format (RFC 2865):
//   Type=11, Length=variable, Value=bandwidth_in_kbps
//   Example: "10240" means 10 Mbps downstream limit
//
// See: internal/radiusd/vendors/huawei/README.md for full VSA specification
func ParseHuaweiInputAverageRate(attr *radius.Attribute) int64 {
    // Implementation
}
```

#### Documentation Strategy

**When to Write Code Comments (Always Required):**

- ✅ All exported functions, methods, types, constants
- ✅ Complex algorithms or business logic
- ✅ Non-obvious design decisions
- ✅ Protocol implementations (RADIUS RFCs, vendor specs)
- ✅ Performance-critical code with optimization notes
- ✅ Error handling with expected error scenarios

**When to Create Separate Documentation (Minimal):**

- ✅ **API Integration Guide** - Only for external API consumers (e.g., `docs/api-integration.md`)
- ✅ **Architecture Overview** - High-level system design (e.g., `docs/v9-architecture.md`)
- ✅ **Protocol RFCs** - Already in `docs/rfcs/` (no duplication needed)
- ❌ **NOT for** - Individual features, bug fixes, or code changes
- ❌ **NOT for** - Work summaries, completion reports, or change logs (use Git history)

**When to Update Existing Docs (Only When Necessary):**

- ✅ Public API behavior changes → Update API docs
- ✅ New configuration options → Update config guide
- ✅ Breaking changes → Update `CHANGELOG.md` and migration guide
- ❌ Internal refactoring → No doc update needed (Git commit is enough)

#### Auto-Generated Documentation Rule

**After completing work, unless explicitly requested by the user, AI Agent should NOT create summary documents or work reports.**

- ❌ **Prohibited**: Auto-creating `SUMMARY.md`, `WORK_REPORT.md` after each task
- ❌ **Prohibited**: Adding lengthy "work summaries" or "completion reports" in chat responses
- ❌ **Prohibited**: Creating separate doc files for simple feature additions
- ✅ **Allowed**: Briefly confirming task completion status (1-2 sentences)
- ✅ **Allowed**: Creating documents when explicitly requested by user
- ✅ **Required**: Comprehensive code comments in the implementation itself

**Correct Completion Response:**

```text
✅ Completed config package testing with 98.5% coverage, all tests passing.
```

**Incorrect Completion Response:**

```text
## Work Summary Report ❌

### Completed Items
1. xxxxx
2. xxxxx
...(lengthy summary - this belongs in Git commit messages, not separate docs)
```

#### Documentation Quality Standards

**API-Level Documentation Must Be:**

1. **Clear** - Use simple, plain language; avoid jargon
2. **Complete** - Document all parameters, returns, errors
3. **Practical** - Include real-world usage examples
4. **Accurate** - Keep code and comments in sync
5. **Concise** - No redundant or obvious information

##### Example: High-Quality API Comment

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

**Remember:**

- 📝 **Code comments are mandatory** - Think of them as inline documentation
- 🎯 **Optimize for readers** - Your future self and team members
- 🚫 **Minimize separate docs** - Only create when code comments aren't enough
- ✅ **Git is your changelog** - Commit messages record what/why/when

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
