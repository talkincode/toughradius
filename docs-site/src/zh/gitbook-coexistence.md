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
- **GitBook 站点** —— 通过 GitBook 自有的 GitHub 集成（外部服务）从仓库同步。
  仓库中**没有提交任何 GitBook 配置**（`.gitbook.yaml`、`book.json` 以及 GitBook
  的 `SUMMARY.md` 均不存在），因此新增 `docs-site/` 不会改变 GitBook 的构建方式。

由于 mdBook 源文件被限制在 `docs-site/` 内，而 GitBook 在外部配置，两套系统不共享
构建步骤，也不会相互破坏。

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

为避免两套管线之间出现内容漂移，每份文档只有唯一的权威位置。随着散落文档逐步迁入
手册（路线图条目 M13.2 / M13.3），原始文件会保留一个指向对应章节的简短入口，而不是
复制其内容。

## 构建与校验

- 本地：`mdbook build docs-site` 会在 `docs-site/book/` 生成静态站点；
  `mdbook serve docs-site` 可开启热重载预览。
- CI：一个独立的任务负责构建手册，并对生成的 HTML 执行**离线坏链检查**，因此构建
  失败或内部坏链都会让流水线变红（路线图条目 M13.4）。构建产物（`docs-site/book/`）
  属于构建工件，不纳入版本控制。
