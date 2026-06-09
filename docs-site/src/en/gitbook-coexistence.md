# mdbook & GitBook Coexistence

> 中文版本：[mdbook 与 GitBook 并存](../zh/gitbook-coexistence.md)

ToughRADIUS also publishes documentation through **GitBook**
(`docs.toughradius.net`). This section records the
decision (roadmap item **M13.0**) for how the new mdBook handbook relates to that
existing pipeline.

## Decision: coexistence, not replacement

The mdBook handbook **coexists** with GitBook. It does not replace or disable the
GitBook site. The two pipelines are kept separate and non-conflicting:

- **mdBook handbook** — lives in the repository under `docs-site/`, is written in
  bilingual chapters, and is built and link-checked by CI on every pull request.
  It is the canonical, version-controlled, review-gated home for documentation
  that must evolve together with the code.
- **GitBook site** — synchronizes from the repository through GitBook's own GitHub
  integration (an external service). It is configured by a committed
  **`.gitbook.yaml`** at the repository root, which points GitBook at the same
  `docs-site/src/` sources, uses `introduction.md` as the landing page, and reads
  the **shared `SUMMARY.md`** as its table of contents. GitBook therefore renders
  the curated bilingual handbook instead of inferring a tree from the whole repo.

Both pipelines build from the **same** `docs-site/src/` sources, but each keeps its
own independent configuration (`book.toml` for mdBook, `.gitbook.yaml` for GitBook).
They do not share a build step and cannot break each other, yet they never drift
because they read the same chapters and the same `SUMMARY.md`.

## Where each site is served

The two pipelines publish to **separate** domains, so they never shadow each other:

- **mdBook handbook** — deployed to **GitHub Pages** by `.github/workflows/pages.yml`
  (roadmap item **M13.5**) and served at the custom domain
  **<https://www.toughradius.net/>**. The deploy workflow writes a `CNAME` file into
  the published artifact, so GitHub Pages keeps that custom domain on every release.
  For the domain to resolve, `www.toughradius.net` must point at GitHub Pages — an
  `A` record set to `185.199.108.153`–`185.199.111.153`, or a DNS-only `CNAME` to
  `talkincode.github.io`. The default project URL
  <https://talkincode.github.io/toughradius/> stays available and redirects to the
  custom domain.
- **GitBook** — serves **`docs.toughradius.net`**, which resolves to GitBook's own
  hosting (Cloudflare / Fastly), not GitHub Pages. Because `www.toughradius.net`
  moves to GitHub Pages, GitBook should be configured to keep only
  `docs.toughradius.net` and release `www`, so the two services never claim the same
  host.

## Single source of truth

To avoid content drift between the two pipelines, every document has exactly one
canonical home. Because mdBook and GitBook now read the **same** `docs-site/src/`
chapters and the same `SUMMARY.md`, the handbook sources are the single source for
both rendered sites. As scattered documents are migrated into the handbook
(roadmap items M13.2 / M13.3), the original file keeps a short pointer back to the
corresponding chapter instead of duplicating its content.

## Editing the shared table of contents

`docs-site/src/SUMMARY.md` is read by **both** tools, so keep it to the subset of
Markdown that they parse the same way: a top `# Summary` title, a non-bulleted
`[Introduction / 引言](./introduction.md)` landing link, and **nested bullet lists**
for grouping. The two language sections are expressed as top-level entries
(`- [English](./en/overview.md)` and `- [中文](./zh/overview.md)`) with their pages
nested beneath them. Avoid `#` / `##` part headers for grouping: mdBook only groups
on `#` while GitBook only groups on `##`, so a nested list is the one form that both
render identically.

## Build and validation

- Local: `mdbook build docs-site` produces the static site in `docs-site/book/`,
  and `mdbook serve docs-site` previews it with live reload.
- CI: a dedicated job builds the handbook and runs an **offline link check** over
  the generated HTML, so a build failure or a broken internal link fails the
  pipeline (roadmap item M13.4). The build output (`docs-site/book/`) is a build
  artifact and is not committed.
