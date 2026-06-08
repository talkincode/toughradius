---
name: add-adminapi-endpoint
description: Add a group of Admin REST endpoints in the management backend (TR-F012). Use when a feature needs management APIs such as create/read/update/delete.
---

# Skill: Add Admin REST Endpoints

> Feature ID: `TR-F012` | Milestone: M2 and others

## When to use
When adding a group of REST endpoints to the management backend.

## Pre-research
```text
view internal/adminapi/adminapi.go        # Init() registration entry
view internal/adminapi/nodes.go           # standard CRUD reference
view internal/adminapi/helpers.go         # parsePagination / filter helpers
view internal/adminapi/responses.go       # unified responses
view internal/adminapi/authz.go           # requireAdmin() and other authz
```

## Implementation steps
1. **New file** `internal/adminapi/<feature>.go`, package `adminapi`.
2. **Request structs**: define the payload with `validate` tags (reference `nodePayload`).
3. **register function**:
   ```go
   func register<Feature>Routes() {
       webserver.ApiGET("/<path>", list<Feature>)
       webserver.ApiPOST("/<path>", create<Feature>, requireAdmin())
       webserver.ApiPUT("/<path>/:id", update<Feature>, requireAdmin())
       webserver.ApiDELETE("/<path>/:id", delete<Feature>, requireAdmin())
   }
   ```
4. **Register in Init**: add `register<Feature>Routes()` to `Init()` in `adminapi.go`.
5. **Data access**: always use `app.GDB()`; never inject `*gorm.DB`.
6. **Shared conventions**: reuse pagination, filtering, validation, authz, unified responses, and error handling.

## Acceptance
- [ ] Add `<feature>_test.go` covering CRUD and authz failure
- [ ] `go test ./internal/adminapi/...` passes
- [ ] `golangci-lint run` passes
- [ ] If the frontend is involved, pair with `../add-react-admin-resource/SKILL.md`
- [ ] PR references `TR-F012` and the milestone ID
