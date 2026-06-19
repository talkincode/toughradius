---
name: release-version
description: Review merged PRs since the last Git tag and decide whether a ToughRADIUS release is warranted. Use when Codex is asked to prepare a version release, audit unreleased changes, decide whether to publish, create a new release tag, or tag origin/main after PR review.
---

# Skill: Release Version

Use this skill to turn merged PR history into a release decision and, when warranted, an annotated Git tag. It is intentionally conservative: tag only after reviewing the PRs since the previous tag and confirming the target revision is current and releasable.

## Workflow

1. **Synchronize release refs first.**
   ```
   git fetch origin main --tags --prune
   ```
   Use `origin/main` as the release target unless the user explicitly names another ref. The current worktree may be dirty; do not tag the local `HEAD` unless it is the requested release target.

2. **Collect the release context.**
   ```
   .agents/skills/release-version/scripts/release_context.py --fetch --ref origin/main
   ```
   The script reports the latest reachable tag, merged PRs/commits since that tag, a deterministic impact classification, and a suggested next tag. Treat this as evidence collection, not as the final decision.

3. **Review every PR since the last tag.**
   For each PR reported by the script, inspect at least the title, body, file list, and release-relevant diff:
   ```
   gh pr view <n> --json number,title,body,labels,mergedAt,files,url
   gh pr diff <n>
   ```
   Focus on user-visible behavior, protocol behavior, security, configuration, database/schema changes, dependency/runtime changes, docs-site/public documentation changes, and operational risk. Ignore purely internal agent metadata unless it changes shipped behavior.

4. **Decide release need.**
   - **No tag**: no PRs since the last tag, or only tests/CI/internal docs/agent-only cleanup with no shipped behavior.
   - **Patch**: bug fixes, security hardening without incompatible behavior, dependency updates, metrics/logging correctness, docs corrections for a published release, or low-risk operational fixes.
   - **Minor**: new user-visible features, new protocol/vendor support, new config/API surface, new automation that affects operators, or meaningful docs/product capability expansion.
   - **Major**: breaking API/config/protocol behavior, incompatible data migration, removed feature, changed deployment contract, or explicit `BREAKING CHANGE`.
   When in doubt between no tag and patch, prefer no tag unless a real consumer/operator would benefit from a release artifact.

5. **Check release gates before tagging.**
   - `git rev-parse origin/main` is the intended target SHA.
   - The proposed tag does not already exist locally or remotely.
   - Open release-blocking PRs/issues are understood. At minimum, list open PRs:
     ```
     gh pr list --state open --json number,title,headRefName,labels,url
     ```
   - CI for the target SHA is green or explicitly accepted by the user:
     ```
     gh run list --branch main --limit 10 --json databaseId,headSha,status,conclusion,workflowName,event
     ```
   - `v*` tags also trigger `.github/workflows/docker-publish.yml`. Confirm
     Docker Hub credentials (`DOCKERHUB_USERNAME` / `DOCKERHUB_TOKEN`) are
     configured and writable. For GHCR, confirm the `talkincode/toughradius`
     package inherits this repository's Actions access, or that
     `PKG_GITHUB_TOKEN` has `write:packages` (`PKG_GITHUB_USERNAME` is optional
     when the token owner differs from the tag actor). The Docker workflow treats
     Docker Hub as required and reports GHCR permission failures in the run
     summary; fix package access and rerun the tag workflow rather than creating
     a duplicate tag for the same source.
   - If the repository has release notes, changelog, packaging, or version-file conventions, update them in a PR first. This skill only creates a tag directly when no source-file change is required.

6. **Create the tag only when release is warranted.**
   Use an annotated tag at the exact target SHA:
   ```
   target=$(git rev-parse origin/main)
   git tag -a <new-tag> "$target" -m "$(cat /tmp/toughradius-release-notes.txt)"
   git push origin <new-tag>
   ```
   The tag message should include: previous tag, target SHA, release impact, reviewed PR numbers, and a concise release summary. Do not create a GitHub Release unless the user asks for one.

7. **Report the outcome.**
   If tagged, report the tag, target SHA, previous tag, release type, and PR summary. If not tagged, report the last tag, reviewed PRs/commits, and the concrete reason no release is needed.

## Guardrails

- Never tag stale local `main`; tag `origin/main` or the explicitly requested ref.
- Never tag unreviewed PRs just because the script recommends a version bump.
- Never tag if the proposed version is ambiguous; stop and report the ambiguity.
- Never skip release notes in the tag message.
- Never push commits or modify roadmap/checklist files as part of this skill unless the release convention requires a preparatory PR and the user approves that work.
- Never assume GHCR success from `packages: write` alone; package access can be
  denied independently of workflow permissions (see issue #503).
- If GitHub metadata is unavailable, do not create a tag. Report a blocked release review instead.

## Script

`scripts/release_context.py` provides repeatable context collection:

```
.agents/skills/release-version/scripts/release_context.py --ref origin/main
.agents/skills/release-version/scripts/release_context.py --ref origin/main --format json
```

Use `--fetch` for live release work. Use `--format json` when another automation needs to consume the result.

## Acceptance

- [ ] Fetched `origin/main` and tags before review
- [ ] Identified the previous reachable tag and exact target SHA
- [ ] Reviewed every PR/commit since the previous tag
- [ ] Chose no-release/patch/minor/major with a written rationale
- [ ] Verified tag uniqueness and target CI state before tagging
- [ ] Checked Docker Hub and GHCR publish prerequisites for the tag workflow
- [ ] Created and pushed an annotated tag only when release is warranted
