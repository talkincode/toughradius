---
model: GPT-5
tools:
  [
    "search",
    "azure/search",
    "usages",
    "problems",
    "changes",
    "githubRepo",
    "todos",
  ]
description: "TeamsACS Project Code Quality Automated Detection and Analysis"
---

# Code Quality Detection Prompt (review.prompt)

This prompt guides the intelligent assistant to conduct a systematic and comprehensive code quality detection and analysis for the TeamsACS project, identifying potential issues, security risks, performance bottlenecks, and maintainability risks. Follows AGENTS.md and project best practices.

---

## 1. Detection Goals

- **Code Health Assessment**: Complexity, duplicate code, function length, file size
- **Security Risk Identification**: Hardcoded keys, missing input validation, sensitive information leakage
- **Concurrency Safety Check**: Data races, lock contention, goroutine leak risks
- **Performance Issue Discovery**: Unnecessary allocations, inefficient algorithms, frequent I/O operations
- **Test Coverage Analysis**: Missing tests, brittle tests, tests dependent on external environments
- **Architecture Consistency**: Cross-layer calls, circular dependencies, unclear responsibilities
- **Maintainability Assessment**: Missing comments, confusing naming, magic numbers

---

## 2. Detection Dimensions and Standards

### 2.1 Code Complexity

| Metric                | Threshold   | Description                           |
| --------------------- | ----------- | ------------------------------------- |
| Function Length       | >100 lines  | Should be split into smaller units    |
| Cyclomatic Complexity | >15         | Too many logic branches, hard to test |
| Nesting Level         | >4 levels   | Poor readability, error-prone         |
| File Length           | >1000 lines | Likely too many responsibilities      |
| Parameter Count       | >5          | Consider encapsulating in a struct    |

### 2.2 Code Quality

- **Duplicate Code**: Similar logic fragments >= 3 places
- **Unused Code**: Exported but unreferenced functions/types
- **Error Handling**: Ignoring errors, swallowing errors, empty panic/recover
- **Naming Conventions**: Unclear abbreviations, single-letter variables (non-loop indices)
- **Comment Completeness**: Missing documentation comments for exported symbols

### 2.3 Security Detection

- Hardcoded keys/passwords/tokens (strings containing "password", "secret", "token", "key", etc.)
- Sensitive information written to logs (passwords, keys, certificate content)
- SQL concatenation (potential injection risk)
- Command concatenation (os/exec using unvalidated parameters)
- Missing input validation (HTTP handlers using parameters directly)
- Insecure random number generation (math/rand used for security scenarios)
- Improper TLS configuration (InsecureSkipVerify = true)

### 2.4 Concurrency Safety

- Shared variables without lock protection (multiple goroutines writing to the same variable)
- Lock scope too large/nested locks (deadlock risk)
- Goroutine leaks (no context cancellation mechanism)
- Channel unclosed or double closed
- sync type copying (Mutex, WaitGroup)
- Map concurrent read/write (not using sync.Map or locks)

### 2.5 Performance Issues

- String concatenation in loops (not using strings.Builder/bytes.Buffer)
- Unnecessary type conversion or serialization
- defer in tight loops (may affect performance)
- Frequent small object allocations (consider sync.Pool)
- Repeated I/O operations (database/file/network)

### 2.6 Test Quality

- Critical modules without test files
- Test coverage <50% (controllers, service/app, common core packages)
- Tests dependent on external resources (no mock/isolation)
- Tests without assertions (running only without verification)
- Unclear test names (Test1, Test2)

---

## 3. Detection Process (Standard Loop)

### Phase 1: Scope Determination

1. **User Specified** or **Full Project Scan**
   - If user specifies file/directory, detect only that scope
   - For full project, priority: controllers > service > common > jobs/events
2. **Generate Todo List** (todos)
   - Use todos tool to manage detection tasks
   - Split into: Complexity -> Security -> Concurrency -> Performance -> Testing -> Architecture

### Phase 2: Static Analysis

1. **Search Key Patterns** (grep_search / semantic_search)
   - Hardcoded keys: `password.*=|secret.*=|token.*=|apikey.*=`
   - Ignored errors: `err\s*:?=.*\n\s*$` (unchecked)
   - SQL concatenation: `fmt.Sprintf.*SELECT|"SELECT.*\+`
   - Shared variables: Identify package-level variables + goroutine usage
2. **List Problematic Functions** (list_code_usages)
   - Find callers of high-complexity functions
   - Identify cross-package coupling links
3. **Statistical Metrics** (search + regex)
   - Function line count distribution
   - File size sorting
   - Exported symbol reference count

### Phase 3: Deep Inspection

1. **Concurrency Safety Analysis**
   - Use `grep_search` to find `var.*sync.Mutex|map\[.*\]`
   - Check if goroutine start points have context control
   - Look for `go func()` without defer recover protection
2. **Error Handling Review**
   - Search `_ =` or `err != nil { }` (empty handling)
   - Check if panic usage is reasonable (prohibited outside initialization phase)
3. **Log Security**
   - Search `log.*password|log.*secret|log.*token`
   - Check if fmt.Println is used in production paths

### Phase 4: Test Coverage Assessment

1. **Use runTests mode=coverage**
   - Analyze files with coverage <50%
   - Identify critical modules with absolutely no tests
2. **Test Quality Check**
   - Search `t.Skip` usage (is it reasonable)
   - Check if tests depend on real database/network

### Phase 5: Report Generation

1. **Categorize Issue List**
   - By severity: High (Security/Concurrency) > Medium (Performance/Complexity) > Low (Style)
2. **Quantify Metrics**
   - Total issues, category distribution
   - Top 10 high-risk files
3. **Fix Suggestions**
   - Specific fix direction for each issue
   - Mark priority and dependencies

---

## 4. Detection Rule Library

### Rule 1: Hardcoded Key Detection

```regex
Regex: (password|secret|token|apikey|private_key)\s*[:=]\s*["'](?!{{|$|xxx)[^"']{8,}
Location: .go files (exclude test constants in *_test.go)
Severity: High
```

### Rule 2: Error Ignored

```regex
Regex: err\s*:?=.*\n(?!\s*(if|return|log|panic))
Exclude: defer statements, known error-free functions (Close, etc.)
Severity: Medium
```

### Rule 3: SQL Injection Risk

```regex
Regex: (fmt\.Sprintf|fmt\.Sprint|"\s*\+\s*).*(SELECT|INSERT|UPDATE|DELETE)
Location: Raw SQL not wrapped by ORM
Severity: High
```

### Rule 4: Concurrency Data Race

```go
Pattern:
- Package-level variable var + non-const/sync type
- Assignment inside goroutine
- No Mutex protection or non-atomic operation
Severity: High
```

### Rule 5: Goroutine Leak

```go
Pattern:
- go func() without context.Done() listening
- channel unclosed and no timeout
- Infinite loop without exit condition
Severity: Medium
```

### Rule 6: Sensitive Logs

```regex
Regex: (log\.|zap\.|fmt\.Print).*(password|secret|token|key|cert)
Exclude: Masked (*** placeholder)
Severity: High
```

### Rule 7: Magic Numbers

```go
Pattern:
- Hardcoded numbers (non 0, 1, -1)
- Appearing repeatedly >= 3 times
- Not defined as constant
Severity: Low
```

---

## 5. Output Report Template

````markdown
# TeamsACS Code Quality Detection Report

**Detection Time**: <timestamp>
**Detection Scope**: <Full Project | Specified Path>
**Tool Version**: review.prompt v1.0

---

## 📊 Overall Score

| Dimension               | Score     | Description                                    |
| ----------------------- | --------- | ---------------------------------------------- |
| Code Complexity         | <A-F>     | Average function length, cyclomatic complexity |
| Security                | <A-F>     | Hardcoding, injection, log leakage             |
| Concurrency Safety      | <A-F>     | Data races, lock usage                         |
| Test Coverage           | <A-F>     | Coverage rate and test quality                 |
| Maintainability         | <A-F>     | Naming, comments, structure                    |
| **Comprehensive Score** | **<A-F>** | Weighted average                               |

---

## 🚨 High Priority Issues (Top 10)

### 1. [Security] Hardcoded Key - controllers/api/auth.go:45

**Issue**:

```go
apiKey := "sk-1234567890abcdef"  // Hardcoded
```
````

**Impact**: Key leakage risk
**Fix Suggestion**:

- Use environment variables or configuration files
- Example: `os.Getenv("API_KEY")`

### 2. [Concurrency] Data Race - service/cache/cache.go:78

**Issue**:

```go
var cacheMap = make(map[string]interface{})  // No lock
func Set(k string, v interface{}) { cacheMap[k] = v }  // Concurrent write
```

**Impact**: Runtime panic
**Fix Suggestion**:

- Use sync.Map or add locks
- Reference: `common/M1Cache` implementation

...(Other issues)

---

## 📈 Metric Details

### Code Complexity Distribution

| File                       | Function Count | Avg Length | Max Cyclomatic Complexity | Risk Level |
| -------------------------- | -------------- | ---------- | ------------------------- | ---------- |
| controllers/cpe/handler.go | 25             | 85 lines   | 18                        | High       |
| service/app/app.go         | 18             | 120 lines  | 12                        | Medium     |
| ...                        | ...            | ...        | ...                       | ...        |

### Security Issue Statistics

- Hardcoded keys: <count> places
- SQL injection risks: <count> places
- Sensitive logs: <count> places
- Missing input validation: <count> places

### Concurrency Issue Statistics

- Potential data races: <count> places
- Goroutine leak risks: <count> places
- Improper lock usage: <count> places

### Test Coverage

| Package         | Coverage | Missing Test Functions     |
| --------------- | -------- | -------------------------- |
| controllers/api | 35%      | HandleLogin, ValidateToken |
| service/app     | 60%      | Init error branch          |
| common/cziploc  | 80%      | -                          |

---

## 🛠️ Fix Plan Suggestions

### Phase 1 (High Priority - Within 1 Week)

1. Fix all hardcoded keys
2. Supplement input validation (controllers/\*)
3. Fix known data races

### Phase 2 (Medium Priority - Within 2 Weeks)

1. Split large functions (>100 lines)
2. Improve test coverage to 60%
3. Standardize error handling

### Phase 3 (Continuous Optimization)

1. Refactor circular dependencies
2. Optimize performance hotspots
3. Supplement documentation comments

---

## 📝 Appendix

### A. Detection Configuration

- Ignored paths: vendor/, assets/static/, bin/
- Detection rule version: v1.0
- Coverage tool: go test -cover

### B. References

- AGENTS.md
- deprecated.md
- Go Code Review Comments

---

**Note**: This report is generated by an automated tool, manual review is recommended before executing fixes.

```

---

## 6. Usage

### Full Project Detection
```

Use review.prompt for comprehensive code quality detection

```

### Specified Module Detection
```

Detect code quality of controllers/api directory

```

### Specialized Detection
```

Detect security risks in the project
Detect concurrency safety issues
Analyze test coverage

```

---

## 7. Detection Constraints and Best Practices

### Must Follow
1. **Non-destructive**: Analyze only, do not modify (unless user explicitly requests auto-fix)
2. **Verifiability**: Each issue must provide specific location and code snippet
3. **Priority Sorting**: High-risk issues (Security/Concurrency) displayed first
4. **False Positive Control**: Mark uncertain issues as "To be manually confirmed"
5. **Quantified Output**: Provide measurable metrics and trends

### Prohibited Behaviors
- Do not give vague or unlocatable issue descriptions
- Do not judge "errors" without understanding context
- Do not automatically delete or rewrite code (unless authorized by user)
- Do not miss high-priority issues while obsessing over low-priority details

---

## 8. Toolchain Integration

### Use search Tool
- Precisely locate problematic code
- Analyze call chains and dependencies

### Use think Tool
- Output reasoning process before analyzing complex issues
- Identify potential architectural issues

### Use todos Tool
- Manage detection task progress
- Track fix status

### Use runTests Tool
- Get real-time coverage data
- Verify if issues are reproducible

### Use problems Tool
- Get compiler-level errors/warnings
- Assist static analysis

---

## 9. Scoring Algorithm

### Dimension Weights
- Security: 30%
- Concurrency Safety: 25%
- Code Complexity: 20%
- Test Coverage: 15%
- Maintainability: 10%

### Grade Standards
| Grade | Score | Description |
|---|---|---|
| A | 90-100 | Excellent, follows best practices |
| B | 80-89 | Good, minor improvements needed |
| C | 70-79 | Medium, needs significant improvement |
| D | 60-69 | Poor, obvious issues exist |
| F | <60 | Fail, severe issues |

### Deduction Rules
- Each high-risk issue: -5 points
- Each medium-risk issue: -2 points
- Each low-risk issue: -0.5 points
- Test coverage <50%: -10 points
- Existence of hardcoded keys: -15 points

---

## 10. Example Detection Conversation

**User**: Detect code quality of controllers/api

**Assistant**:
1. Use search to scan controllers/api directory structure
2. Use grep_search to find security patterns (keys, SQL concatenation)
3. Analyze function complexity (count lines, nesting)
4. Check test coverage (runTests mode=coverage)
5. Generate report (Issue list + Priority + Fix suggestions)
6. Output todos task list for follow-up

---

## 11. Extended Detection Items

### Architecture Level
- Circular dependency detection (package level)
- Interface abuse (single implementation interface)
- Cross-layer calls (controller directly accessing DB)

### Go Specific
- Improper Context usage (nil context, stored in struct)
- defer misuse (in loop, return value impact)
- slice/map accidental sharing

### Performance Specific
- Unnecessary reflection
- Frequent JSON encoding/decoding
- Large object copying/passing

---

## 12. Continuous Improvement

### Baseline Establishment
- First detection result as baseline
- Periodic (weekly/monthly) re-detection
- Track metric trends

### Rule Optimization
- Adjust rules based on false positives
- Add project-specific detection patterns
- Keep in sync with AGENTS.md

---

## 13. Output Language and Format

- Report in **English**
- Keep code snippets in original English
- Use Markdown format for tables
- Bold or highlight key issues
- Attach specific file names and line numbers

---

## 14. Conclusion

Code quality detection is the starting point of continuous improvement, not the end. Follow the closed loop of "Discover Issues -> Prioritize -> Make Plan -> Small Step Fixes -> Verify Improvements" to gradually improve the overall health of the TeamsACS project.

**Detection is not for criticism, but for better understanding and improvement.**

```
