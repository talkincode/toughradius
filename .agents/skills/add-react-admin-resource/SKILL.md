---
name: add-react-admin-resource
description: Add a resource or page in the React Admin management backend (TR-F013). Use when a new Admin API needs a corresponding management UI in the frontend.
---

# Skill: Add a Frontend Management Resource / Page

> Feature ID: `TR-F013` | Milestone: M2 and others

## When to use
After adding an Admin API on the backend, when the React Admin backend needs to expose the corresponding resource or page.

## Pre-research
```text
view web/src/App.tsx                       # resource / route registration
file_search "web/src/resources/*.tsx"      # resource references
view web/src/providers/dataProvider.ts     # REST mapping
view web/src/resources/nodes.tsx           # standard resource reference
```

## Implementation steps
1. **Resource file**: define List/Edit/Create in `web/src/resources/<feature>.tsx` (mirror `nodes.tsx`).
2. **Register the resource**: add `<Resource name="<api-path>" .../>` in `web/src/App.tsx`; the name must align with the backend API path.
3. **Data mapping**: confirm `dataProvider.ts` pagination / filter / sort params match the backend.
4. **Pages (if not standard CRUD)**: put them under `web/src/pages/`; do not create a separate management entry.
5. **i18n**: if the project has i18n keys, add the corresponding strings.

## Boundaries
- Do not introduce a separate management entry; all pages hang under the unified Admin framework.
- The frontend only exposes backend-supported, verifiable, safe actions (especially CoA); never assemble protocol packets in the frontend.

## Acceptance
- [ ] `cd web && npm run build` succeeds
- [ ] List / filter / CRUD are connected to the backend
- [ ] PR references `TR-F013` and the milestone ID
