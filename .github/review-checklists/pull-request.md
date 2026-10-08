# Pull Request Review Checklist (TR-F022)

Use this checklist when preparing or reviewing pull requests.

## 1) Context and Scope

- [ ] PR explicitly maps to relevant milestone subtask (`M?.?`) and `TR-F` ID(s)
- [ ] No non-goal scope (`TR-N001` to `TR-N006`) is introduced
- [ ] Claimed acceptance criteria match the actual diff

## 2) Mechanical Gates

- [ ] `gh pr checks <n>` is fully green (or a transient failure has been rerun once)
- [ ] Required local gates pass (`go build ./...`, `go test ./...`, `golangci-lint run`, optional `cd web && npm run build`)
- [ ] If `.github/workflows/**` or actionlint configs changed, `Workflow Lint` passes

## 3) Code and Safety Review

- [ ] No correctness regressions (logic bugs, nil dereferences, race conditions, resource leaks)
- [ ] No security regressions (auth bypass, injection, secret/plaintext credential leakage)
- [ ] Data/schema changes preserve PostgreSQL and SQLite compatibility
- [ ] Protocol behavior changes include RFC references under `docs/rfcs/` and do not contradict cited clauses

## 4) Test Sufficiency

- [ ] Changed behavior is covered by tests (unit / integration)
- [ ] Protocol / E2E changes include `test/integration/` acceptance coverage

## 5) Review Verdict

- [ ] Blocking findings are posted with `file:line` and concrete fix guidance
- [ ] Approved by maintainer before merging
