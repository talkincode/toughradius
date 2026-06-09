# ToughRADIUS Handbook

Welcome to the **ToughRADIUS Handbook** — the canonical, in-repository, bilingual
documentation site for the [ToughRADIUS](https://github.com/talkincode/toughradius)
RADIUS server. It is built with [mdBook](https://rust-lang.github.io/mdBook/) and
lives in the repository under `docs-site/`, so the documentation is versioned,
reviewed, and validated by CI together with the code it describes.

The handbook is organized into two mirrored language sections. Pick a language to
begin:

- **English** — start with the [Overview](./en/overview.md).
- **中文** — 从[概述](./zh/overview.md)开始。

Each chapter exists in both languages with a matching structure, and the pages
cross-link to their counterparts so you can switch language at any point.

> **Relationship to the existing GitBook site.** ToughRADIUS also publishes
> documentation through GitBook (`docs.toughradius.net` / `www.toughradius.net`).
> This mdBook handbook **coexists** with GitBook rather than replacing it; see
> [mdbook & GitBook Coexistence](./en/gitbook-coexistence.md) /
> [mdbook 与 GitBook 并存](./zh/gitbook-coexistence.md) for the single-source-of-truth
> policy and the boundary between the two pipelines.

## Build it locally

```bash
# Install mdBook (https://rust-lang.github.io/mdBook/guide/installation.html)
cargo install mdbook            # or: brew install mdbook

# Build the static site into docs-site/book/
mdbook build docs-site

# Or serve with live reload at http://localhost:3000
mdbook serve docs-site
```

---

# ToughRADIUS 使用手册

欢迎来到 **ToughRADIUS 使用手册** —— 这是
[ToughRADIUS](https://github.com/talkincode/toughradius) RADIUS 服务器
随仓库维护的中英文双语文档站点。它基于
[mdBook](https://rust-lang.github.io/mdBook/) 构建，位于仓库的 `docs-site/`
目录下，因此文档与代码一起被版本化、评审，并由 CI 校验。

手册分为结构一一对应的两个语言区。请选择语言开始阅读：

- **English** — 从 [Overview](./en/overview.md) 开始。
- **中文** — 从[概述](./zh/overview.md)开始。

每个章节都提供中英文两个版本且结构对应，页面之间相互交叉链接，方便随时切换语言。

> **与现有 GitBook 站点的关系。** ToughRADIUS 目前还通过 GitBook 发布文档
> （`docs.toughradius.net` / `www.toughradius.net`）。本 mdBook 手册与 GitBook
> **并存**而非替代；单一事实来源策略与两套管线的边界详见
> [mdbook 与 GitBook 并存](./zh/gitbook-coexistence.md) /
> [mdbook & GitBook Coexistence](./en/gitbook-coexistence.md)。

## 本地构建

```bash
# 安装 mdBook（https://rust-lang.github.io/mdBook/guide/installation.html）
cargo install mdbook            # 或：brew install mdbook

# 将静态站点构建到 docs-site/book/
mdbook build docs-site

# 或开启热重载预览，访问 http://localhost:3000
mdbook serve docs-site
```
