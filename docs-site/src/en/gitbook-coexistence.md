# mdbook & GitBook Coexistence

> 中文版本：[mdbook 与 GitBook 并存](../zh/gitbook-coexistence.md)

ToughRADIUS already publishes documentation through **GitBook**
(`docs.toughradius.net` and `www.toughradius.net`). This section records the
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
  integration (an external service). There is **no GitBook configuration committed
  to this repository** (`.gitbook.yaml`, `book.json`, and a GitBook `SUMMARY.md`
  are all absent), so adding `docs-site/` does not change how GitBook builds.

Because the mdBook sources are confined to `docs-site/` and GitBook is configured
externally, the two systems do not share a build step and cannot break each other.

## Where each site is served

The two pipelines publish to **separate** domains, so they never shadow each other:

- **mdBook handbook** — deployed to **GitHub Pages** by `.github/workflows/pages.yml`
  (roadmap item **M13.5**) and served at the repository's default project URL,
  **<https://talkincode.github.io/toughradius/>**. The repository's Pages site has
  **no custom domain** configured, precisely so it does not collide with the
  GitBook domains below.
- **GitBook** — continues to serve the public domains **`www.toughradius.net`** and
  **`docs.toughradius.net`**, which resolve to GitBook's own hosting (Cloudflare /
  Fastly), not GitHub Pages.

## Single source of truth

To avoid content drift between the two pipelines, every document has exactly one
canonical home. As scattered documents are migrated into the handbook
(roadmap items M13.2 / M13.3), the original file keeps a short pointer back to the
corresponding chapter instead of duplicating its content.

## Build and validation

- Local: `mdbook build docs-site` produces the static site in `docs-site/book/`,
  and `mdbook serve docs-site` previews it with live reload.
- CI: a dedicated job builds the handbook and runs an **offline link check** over
  the generated HTML, so a build failure or a broken internal link fails the
  pipeline (roadmap item M13.4). The build output (`docs-site/book/`) is a build
  artifact and is not committed.
