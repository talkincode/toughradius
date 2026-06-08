---
name: add-config-schema
description: Add a dynamic config item editable on the system config page (TR-F014). Use when adding a RADIUS runtime parameter that must support query/edit/reload.
---

# Skill: Add a Dynamic Config Item

> Feature ID: `TR-F014`

## When to use
When adding a RADIUS runtime parameter that should be editable on the system config page.

## Pre-research
```text
view internal/app/config_schemas.json      # config schema (edit here first)
view internal/app/config_manager.go        # read / reload logic
view internal/adminapi/settings.go         # config endpoints
grep_search "GetSettingsStringValue\|GetSettingsInt64Value" --include internal/**
```

## Implementation steps
1. **Schema**: add the config item to `config_schemas.json` with `default / type / range / i18n key`.
2. **Read**: read via `app.GApp().GetSettings*Value("<group>", "<Key>")`; do not hardcode.
3. **Default initialization**: confirm `checkSettings()` / init logic writes the default value.
4. **Frontend**: the system config page `web/src/pages/SystemConfigPage.tsx` renders from the schema automatically; customize only if a special control is needed.

## Conventions
- A new config item must enter the schema before any code reads it.
- Type, range, default, and i18n key are all required.

## Acceptance
- [ ] The config can be queried / edited / reloaded in the backend
- [ ] The new config read path has a test
- [ ] `go test ./internal/app/...` passes
- [ ] PR references `TR-F014`
