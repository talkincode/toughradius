---
name: document-go-apis
description: Standard-library-style Go API documentation and comment conventions (TR-F024). Use when adding or changing exported Go identifiers, writing package comments, or backfilling godoc on a package so it reads like the Go standard library.
---

# Skill: Document Go APIs (standard-library style)

> Feature ID: `TR-F024` | Milestone: M4 (M4.11 / M4.12)

## When to use
- Adding or changing any **exported** Go identifier (func, type, method, const, var).
- Writing or updating a package comment (`doc.go`).
- Backfilling godoc on an existing package so it reads like the standard library.

The goal: a reader using `go doc ./...` or pkg.go.dev should understand the API
**without reading the implementation**. Optimize for information value, not coverage theater.

## Pre-research
```text
go doc ./internal/radiusd                 # see how a package currently documents
view internal/radiusd/coa_service.go      # a well-documented reference in this repo
grep -rn "^// Package " internal pkg       # find existing package comments
```
Mirror the nearest well-documented neighbor; do not invent a new comment style.

## Core conventions (godoc / stdlib)
1. **Every exported identifier has a doc comment**, and the comment **starts with the identifier name**:
   ```go
   // CoAService sends RFC 5176 Dynamic Authorization (CoA / Disconnect) requests
   // to a NAS and reports a structured result.
   type CoAService struct { ... }

   // Disconnect sends a Disconnect-Request for the given session identity and
   // blocks until an ACK/NAK is received or the per-attempt timeout elapses.
   func (s *CoAService) Disconnect(ctx context.Context, ...) (*CoAResult, error) { ... }
   ```
2. **Package comment**: each non-`main` package has exactly one package comment, in a dedicated
   `doc.go` when it is more than a line. Start with `// Package <name> ...`:
   ```go
   // Package radiusd implements the ToughRADIUS authentication, accounting,
   // and dynamic-authorization (CoA) protocol services.
   package radiusd
   ```
3. **Full sentences, present tense**, ending with a period. First sentence is a self-contained
   summary (it is what pkg.go.dev shows in lists). Keep the summary on the first line.
4. **Document the contract, not the mechanics**: parameters' meaning, return values, the zero value
   if it is usable, units (e.g. Kbps vs Mbps), ownership, and what counts as success vs failure.
5. **Errors**: state what error types / sentinel values callers should branch on
   (e.g. `AuthError` with a metrics tag). Wrap with `%w` and say so when relevant.
6. **Concurrency**: explicitly state whether a type/func is safe for concurrent use
   ("safe for concurrent use by multiple goroutines" or "not safe for concurrent use").
7. **Context & blocking**: note when a function blocks, honors `ctx` cancellation, performs I/O,
   or has timeout/retry semantics.
8. **Runnable examples** for non-trivial APIs: add `Example<Name>` functions in `*_test.go`
   (package `<pkg>_test`); they are compiled and run by `go test` and rendered on pkg.go.dev.
9. **Deprecation**: use a `// Deprecated: use X instead.` paragraph; do not delete the symbol abruptly.
10. **Unexported code**: comment only where the *why* is non-obvious (protocol quirks, vendor unit
    conversions, security-sensitive branches). Do not narrate obvious code.

## Anti-patterns (reject these)
- `// GetUser gets the user.` — restates the name, zero information. Say *what* it loads, from where,
  and the failure mode.
- Comment that does not start with the identifier name (breaks godoc association).
- Per-line narration of self-evident statements.
- Documenting unexported helpers exhaustively while exported API stays bare.

## How to check
```bash
go doc ./internal/<pkg>                     # read the rendered API; gaps are obvious
gofmt -l .                                  # comment formatting is part of gofmt
go vet ./...                                # catches some doc/format issues
go test ./...                              # compiles and runs Example functions
golangci-lint run                           # v2.12.2
```
**Lint enforcement (incremental, M4.12):** `.golangci.yml` currently disables the godoc checks
(`staticcheck` `ST1000` package-comment, `ST1020`/`ST1021` func/type comment format). Do **not**
flip them on globally in one shot — that would flag the whole tree. Backfill a package's docs first,
then enable the relevant check (or `revive`'s `exported` rule) scoped to that package, so the gate
ratchets forward without a giant noisy diff.

## Acceptance
- [ ] Every new/changed exported identifier has a doc comment that starts with its name and states the contract
- [ ] The package has a package comment (`doc.go` when multi-line)
- [ ] Units, error types to branch on, and concurrency safety are stated where they apply
- [ ] Non-trivial new APIs ship a runnable `Example`
- [ ] `gofmt -l` clean, `go vet ./...`, `go test ./...`, and `golangci-lint run` pass
- [ ] PR references `TR-F024` (and the milestone subtask, e.g. M4.12, when backfilling)
