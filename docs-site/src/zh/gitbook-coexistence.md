# mdbook 与 GitBook 并存

> English version: [mdbook & GitBook Coexistence](../en/gitbook-coexistence.md)

ToughRADIUS 目前已经通过 **GitBook** 发布文档（`docs.toughradius.net` 与
`www.toughradius.net`）。本节记录路线图条目 **M13.0** 的决策：新的 mdBook 手册
与既有 GitBook 管线如何相处。

## 决策：并存而非替代

mdBook 手册与 GitBook **并存**，不替代、也不停用 GitBook 站点。两套管线相互独立、
互不冲突：

- **mdBook 手册** —— 位于仓库的 `docs-site/` 目录，以双语章节编写，并在每个 Pull
  Request 上由 CI 构建与坏链校验。它是需要与代码同步演进的文档的权威来源：受版本
  控制、走评审门禁。
- **GitBook 站点** —— 通过 GitBook 自有的 GitHub 集成（外部服务）从仓库同步。它由仓库
  根目录提交的 **`.gitbook.yaml`** 配置：将 GitBook 指向同一份 `docs-site/src/` 源文件，
  以 `introduction.md` 作为首页，并读取**共享的 `SUMMARY.md`** 作为目录。因此 GitBook
  渲染的是经过整理的双语手册，而不是从整个仓库推断出的目录树。

两套管线都从**同一份** `docs-site/src/` 源文件构建，但各自保留独立配置（mdBook 用
`book.toml`，GitBook 用 `.gitbook.yaml`）。它们不共享构建步骤、不会相互破坏；同时由于
读取相同的章节与同一个 `SUMMARY.md`，也不会发生内容漂移。

## 各站点的对外地址

两套管线发布到**不同**域名，因此不会相互遮蔽：

- **mdBook 手册** —— 由 `.github/workflows/pages.yml`（路线图条目 **M13.5**）部署到
  **GitHub Pages**，使用仓库默认的项目地址对外服务：
  **<https://talkincode.github.io/toughradius/>**。仓库 Pages 站点**未配置自定义
  域名**，正是为了避免与下方 GitBook 域名冲突。
- **GitBook** —— 继续服务公共域名 **`www.toughradius.net`** 与
  **`docs.toughradius.net`**，这些域名解析到 GitBook 自有托管（Cloudflare / Fastly），
  而非 GitHub Pages。

## 单一事实来源

为避免两套管线之间出现内容漂移，每份文档只有唯一的权威位置。由于 mdBook 与 GitBook
现在读取**同一份** `docs-site/src/` 章节与同一个 `SUMMARY.md`，手册源文件即是两个渲染
站点的单一事实来源。随着散落文档逐步迁入手册（路线图条目 M13.2 / M13.3），原始文件会
保留一个指向对应章节的简短入口，而不是复制其内容。

## 编辑共享目录

`docs-site/src/SUMMARY.md` 会被**两个**工具同时读取，因此只能使用两者解析方式一致的
Markdown 子集：顶部的 `# Summary` 标题、一个非列表项的首页链接
`[Introduction / 引言](./introduction.md)`，以及用于分组的**嵌套列表**。两个语言分区
表示为顶层条目（`- [English](./en/overview.md)` 与 `- [中文](./zh/overview.md)`），其余
页面嵌套其下。请避免使用 `#` / `##` 标题来分组：mdBook 只按 `#` 分组，而 GitBook 只按
`##` 分组，因此嵌套列表是两者渲染方式完全一致的唯一形式。

## 构建与校验

- 本地：`mdbook build docs-site` 会在 `docs-site/book/` 生成静态站点；
  `mdbook serve docs-site` 可开启热重载预览。
- CI：一个独立的任务负责构建手册，并对生成的 HTML 执行**离线坏链检查**，因此构建
  失败或内部坏链都会让流水线变红（路线图条目 M13.4）。构建产物（`docs-site/book/`）
  属于构建工件，不纳入版本控制。
