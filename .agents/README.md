# ToughRADIUS Agent 技能库 (.agents/skills)

本目录是 **agent 驱动开发** 的可复用技能（SOP）库，遵循 Agent Skills 约定：
每个技能是 `.agents/skills/<name>/SKILL.md`，含 `name` / `description` frontmatter，
描述一类标准开发任务的"上下文检索 → 实现 → 测试 → 验收"流程，
确保不同 agent / 会话产出一致、可审查、不偏离项目约定。

## 与其他规范的关系

| 文档 | 作用 |
| --- | --- |
| `AGENT.md` | 项目级 AI 工作总纲（先读现有代码、对齐功能清单） |
| `docs/feature-checklist.md` | 功能范围基线，所有任务锚定 `TR-F` 编号 |
| `docs/roadmap.md` | 长期路线图与里程碑，agent 任务来源 |
| `.github/copilot-instructions.md` | 仓库级 Copilot 指令 |
| `.agents/skills/<name>/SKILL.md` | **本目录**：具体任务类型的执行 SOP |

## 通用前置约束（所有技能共享）

1. **先检索后动手**：用 grep/glob/view 定位现有实现与测试，模仿其命名、错误处理、数据流。
2. **锚定功能编号**：任务必须映射到 `TR-F` 编号；无法映射先更新功能清单。
3. **最小闭环**：每次只交付可独立验证、可回滚的 MVP。
4. **质量门禁**：`go build ./...`、`go test ./...`、`golangci-lint run`（v2.12.2）必须通过；前端改动跑 `cd web && npm run build`；改动 `.github/workflows/**` 或 `.github/actionlint.y*ml` 时必须查看 `Workflow Lint`，该门禁运行 `actionlint -shellcheck=`，只验证 GitHub Actions YAML / expression / action 输入，不混入既有 shell 风格告警。
5. **数据库双兼容**：schema 变更必须同时兼容 PostgreSQL 与 SQLite。
6. **走 PR**：禁止直接推 `main`，PR 描述引用里程碑与功能编号。
7. **协议合规**：协议行为改动必须引用 `docs/rfcs/` 对应 RFC 条款（见 `skills/reference-rfc/SKILL.md`）。
8. **CI 验收测试**：协议 / 端到端改动必须带 CI 可自动执行的验收用例（`test/integration/`，见 `skills/add-acceptance-test/SKILL.md`）。
9. **上游依赖**：核心库 `layeh.com/radius` 经 `go.mod` `replace` 指向组织 fork `github.com/talkincode/radius`；上游重要修复按 `skills/sync-upstream-radius/SKILL.md` 评估同步。

## 技能索引

| 技能 | 适用场景 | 关联编号 |
| --- | --- | --- |
| [release-version](skills/release-version/SKILL.md) | **发布审查**：审查上次 tag 后已合并 PR，判断 no-release / patch / minor / major，必要时给 `origin/main` 创建 annotated tag | TR-F022 / release operations |
| [add-radius-vendor](skills/add-radius-vendor/SKILL.md) | 新增厂商 VSA 解析 / 响应增强 | TR-F005 |
| [add-eap-method](skills/add-eap-method/SKILL.md) | 新增 EAP 认证方法 | TR-F004 |
| [add-adminapi-endpoint](skills/add-adminapi-endpoint/SKILL.md) | 新增 Admin REST 接口 | TR-F012 |
| [add-react-admin-resource](skills/add-react-admin-resource/SKILL.md) | 新增前端管理资源 / 页面 | TR-F013 |
| [add-config-schema](skills/add-config-schema/SKILL.md) | 新增动态配置项 | TR-F014 |
| [add-acceptance-test](skills/add-acceptance-test/SKILL.md) | 编写 CI 自动化验收 / 集成测试 | TR-F022 |
| [sync-upstream-radius](skills/sync-upstream-radius/SKILL.md) | 跟踪 / 同步上游 radius 库 | TR-F021 / TR-F022 |
| [reference-rfc](skills/reference-rfc/SKILL.md) | 检索 / 引用国际标准协议规范 | TR-F021 |
| [align-feature-checklist](skills/align-feature-checklist/SKILL.md) | 需求对齐 / 更新功能清单 | 全部 |
| [write-go-tests](skills/write-go-tests/SKILL.md) | 编写 Go 单元 / 集成测试 | TR-F022 |
| [document-go-apis](skills/document-go-apis/SKILL.md) | 编写 Go 标准库风格 API 文档与注释 | TR-F024 |

## 发布审查入口

准备发版、审查未发布变更或决定是否打 tag 时，使用 [`release-version`](skills/release-version/SKILL.md)。该 SOP 先同步 `origin/main` 与 tags，审查上次 tag 后的已合并 PR / commit，再给出 no-release、patch、minor 或 major 的书面判断；只有确认需要发布且门禁满足时，才在 `origin/main` 的目标 SHA 上创建 annotated tag。

`release-version` 不自动创建 GitHub Release，不修改源码、路线图或版本文件；如仓库后续引入 changelog、release notes 或版本文件更新约定，需先通过单独 PR 补齐这些源文件变更，再执行打 tag。

推送 `v*` tag 会同时触发 `.github/workflows/release-publish.yml` 和
`.github/workflows/docker-publish.yml`。发版前确认 Docker Hub 的
`DOCKERHUB_USERNAME` / `DOCKERHUB_TOKEN` 可写；GHCR 需要 package repository
access / inherited access 允许本仓库 `GITHUB_TOKEN` 写入，或配置具备
`write:packages` 的 `PKG_GITHUB_TOKEN`（若 token 所属账号不同于 tag 触发者，可选配
`PKG_GITHUB_USERNAME`）。Docker Hub 发布是必选门禁；GHCR 发布在 workflow 中独立
执行，且会先做写权限预检，凭据不可写时直接跳过 GHCR push；按 run summary 修复 package
access 或 token 后重跑 tag workflow，避免重复创建同一源码的错误版本 tag。

## 工具链版本（与 CI 对齐）

- Go `1.25`，`CGO_ENABLED=0`
- Node `20`，前端在 `web/`，`npm ci && npm run build`
- golangci-lint `v2.12.2`


## 在自己的主机上运行 Agent（不在 CI 里执行）

> 本仓库**不内置任何自动执行 agent 的 GitHub workflow**。路线图、技能库与约束是供 agent 使用的"知识与护栏"；具体执行由你用**自己的 agent**在**自己的主机**上手动或定时运行。这样密钥不进 CI，执行环境完全自控。

### 开发执行流程

用任意支持工具调用的编码 agent 即可，本仓库不绑定具体 agent 或 CLI。开展开发时，agent 应先读 `AGENT.md`、本文件、`docs/roadmap.md`、`docs/feature-checklist.md`；选用匹配的执行 SKILL；只做最小闭环；协议改动引用 `docs/rfcs/`；补 CI 可执行测试；通过本地与 CI 质量门禁；改动走 PR 并经审查合并。

### 给 agent 的护栏（无论在哪运行都适用）

- **锚定功能编号**：任务必须映射到 `TR-F` 编号；严禁触碰非目标 `TR-N001`~`TR-N005`（支付/CRM/通用监控/多租户/重写）。
- **选任务口径**：`docs/roadmap.md` 自上而下第一个未勾选的 `- [ ] M*.*`。
- **遵循 SOP**：按任务类型选用对应 `.agents/skills/<name>/SKILL.md`。
- **质量门禁**：`go build ./...`、`go test ./...`、`golangci-lint run`（v2.12.2）通过；前端改动跑 `cd web && npm run build`；workflow / actionlint 配置改动必须确认 `Workflow Lint` 通过（`actionlint -shellcheck=`，范围限于 GitHub Actions YAML、表达式与 action 输入；配置路径为 `.github/actionlint.yml` / `.github/actionlint.yaml`）。
- **协议合规**：协议行为改动引用 `docs/rfcs/` 对应 RFC（见 `skills/reference-rfc/SKILL.md`）。
- **CI 验收测试**：协议 / 端到端改动带 CI 可自动执行的验收用例（`test/integration/`，见 `skills/add-acceptance-test/SKILL.md`）。
- **走 PR**：产出一律 Pull Request，禁止直接推 `main`；严格通过质量门禁与代码审查后合并。
- **PR 模板**：agent 任务 PR 默认使用 [`.github/pull_request_template.md`](../.github/pull_request_template.md)，完整填写里程碑子任务、`TR-F`、验收与门禁结果。
- **审查清单**：按 [`.github/review-checklists/pull-request.md`](../.github/review-checklists/pull-request.md) 逐项判定，避免凭主观印象放行。

### 本地集成测试说明（替代开发容器）

运行涉及真实数据库或 LDAP 的端到端集成测试（`test/integration/`）时，可通过本地 Docker Compose 一键启动依赖并执行测试：

```bash
make test-integration-pg
```

该命令基于 `docker-compose.test.yml` 启动 PostgreSQL 与 OpenLDAP 服务容器，运行标记为 `//go:build integration` 的测试用例后自动清理。

### 上游 radius 库跟踪（手动）

核心库 `layeh.com/radius` 经 `go.mod` `replace` 指向组织 fork `github.com/talkincode/radius`。无自动巡检 workflow，按 `skills/sync-upstream-radius/SKILL.md` 的步骤定期人工核对上游是否有安全 / 协议修复并决定是否同步。
