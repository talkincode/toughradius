---
agent: agent
description: Test, fix, commit, and push changes to the repository
---

# Test, Fix, Commit and Push Workflow

Execute a complete development workflow that ensures code quality before pushing changes to the remote repository.

## Workflow Steps

### 1. Run All Tests

Execute comprehensive testing across the entire project:

**Backend Tests:**

- Run all Go unit tests: `go test ./...`
- Run RADIUS protocol tests: `go test ./internal/radiusd/...`
- Run benchmark tests if applicable: `go test -bench=. ./internal/radiusd/`
- Check for race conditions: `go test -race ./...`

**Frontend Tests:**

- Navigate to `web/` directory
- Run frontend tests: `npm test` or `npm run test`
- Run build validation: `npm run build` to ensure production build succeeds

**Integration Tests:**

- Verify database migrations work: Check `app.MigrateDB()` functionality
- Test RADIUS authentication flow end-to-end if changes affect core services
- Validate API endpoints if admin API routes were modified

### 2. Fix Issues

If any tests fail:

- Analyze the failure output carefully
- Trace the root cause through the codebase
- Apply fixes following project conventions:
  - Maintain consistent error handling patterns
  - Follow existing logging standards (zap with namespace)
  - Preserve architectural patterns (errgroup for services, app.GDB() for database access)
  - Update vendor-specific code according to specifications
- Re-run ALL tests after each fix
- Iterate until all tests pass

**Do not proceed to commit until all tests pass successfully.**

### 3. Commit Changes

Once all tests pass, commit the changes with proper organization:

**Commit Strategy:**

- **Small changes (1-5 files):** Single commit
- **Large changes:** Categorize into logical batches:
  - Group by feature/component (e.g., "auth service", "admin API", "frontend UI")
  - Separate backend and frontend changes
  - Isolate refactoring from new features
  - Keep test updates with their corresponding code changes

**Commit Message Format:**

Follow conventional commits specification:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**

- `feat`: New feature
- `fix`: Bug fix
- `refactor`: Code restructuring without behavior change
- `perf`: Performance improvement
- `test`: Adding or updating tests
- `docs`: Documentation changes
- `chore`: Build process, dependencies, configs
- `style`: Code formatting (not UI styles)

**Scope examples:**

- `radiusd`: RADIUS service core
- `adminapi`: Admin API routes
- `webserver`: Web server/middleware
- `domain`: Data models
- `vendors`: Vendor-specific implementations
- `frontend`: React Admin UI
- `config`: Configuration handling

**Examples:**

```
feat(radiusd): add support for Huawei bandwidth VSA attributes

Implement parsing and handling of Huawei-specific VSA attributes
for dynamic bandwidth allocation. Supports input/output average
rate configuration per RFC and Huawei specifications.

- Add ParseHuaweiInputAverageRate function
- Update auth_accept_config.go with bandwidth conversion
- Add unit tests for VSA attribute parsing

Closes #123
```

```
fix(adminapi): resolve user session limit validation error

Correct the max session check logic that was incorrectly rejecting
valid concurrent sessions. The previous implementation didn't account
for NAS IP filtering.

- Update session count query to include NAS IP
- Add integration test for multi-NAS scenarios
- Update error message for clarity
```

```
refactor(domain): migrate user model to use GORM hooks

Replace manual timestamp updates with GORM's built-in hooks for
better consistency and reduced boilerplate.

- Add BeforeCreate/BeforeUpdate hooks
- Remove manual timestamp setting in services
- Update related tests
```

**Commit Command Pattern:**

```bash
git add <files-for-this-batch>
git commit -m "<type>(<scope>): <subject>" -m "<body>" -m "<footer>"
```

### 4. Push to Remote

After all commits are created:

```bash
# Review all commits before pushing
git log --oneline -n <number-of-commits>

# Push to remote
git push origin <branch-name>
```

**Pre-push Checklist:**

- ✅ All tests passing
- ✅ Commits follow conventional format
- ✅ Related changes grouped logically
- ✅ Commit messages are clear and descriptive
- ✅ No sensitive data in commits
- ✅ Branch is up to date with remote (pull/rebase if needed)

## Error Handling

If push fails:

- **Conflict:** Pull latest changes, resolve conflicts, re-test, then push
- **Rejected:** Check if branch protection requires PR or review
- **Authentication:** Verify Git credentials and remote URL

## Success Criteria

The workflow is complete when:

1. All backend and frontend tests pass
2. All changes are committed with proper messages
3. Commits are successfully pushed to remote repository
4. No errors or warnings in the process

## Notes

- Never skip tests even for "small" changes
- Always verify production build succeeds for frontend changes
- Keep commits atomic and focused
- Write commit messages for future maintainers, not just yourself
- Follow ToughRADIUS coding standards from `.github/copilot-instructions.md`
