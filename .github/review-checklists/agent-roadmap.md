# Agent Roadmap PR Review Checklist (TR-F022)

Use this checklist when reviewing PRs labeled `agent-roadmap`.

## 1) Context and Scope

- [ ] PR explicitly maps to one roadmap subtask (`M?.?`) and `TR-F` ID(s)
- [ ] No non-goal scope (`TR-N001` to `TR-N005`) is introduced
- [ ] Claimed acceptance criteria match the actual diff

## 2) Mechanical Gates

- [ ] `gh pr checks <n>` is fully green (or a transient failure has been rerun once)
- [ ] Required local gates are reported in PR (`go build`, `go test`, `golangci-lint`, optional `web build`)

## 3) Code and Safety Review

- [ ] No correctness regressions (logic, nil dereference, race, resource leak)
- [ ] No security regressions (auth bypass, injection, secret/plaintext leakage)
- [ ] Data/schema changes preserve PostgreSQL and SQLite compatibility
- [ ] Protocol behavior changes include RFC references and do not contradict cited clauses

## 4) Test Sufficiency

- [ ] Changed behavior is covered by tests (unit/integration)
- [ ] Protocol/E2E changes include `test/integration/` acceptance coverage

## 5) Review Verdict

- [ ] Blocking findings are posted with `file:line` and concrete fix guidance
- [ ] Verdict label is exactly one of: `agent-approved`, `needs-rework`, `needs-human`
- [ ] Approval comment records the reviewed head SHA
- [ ] Merge happens only when `agent-approved` + green checks + SHA still matches
