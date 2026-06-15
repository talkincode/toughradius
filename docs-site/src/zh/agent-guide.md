# Agent 开发指南

> English version: [Agent Development Guide](../en/agent-guide.md)

本章是一份**面向贡献者**的摘要，介绍 ToughRADIUS 如何借助 AI 编码 agent 进行
开发，归纳工作规则、质量门禁与自动委托循环，使该工作流可从手册统一发现。

**权威**规则位于仓库根目录的
[`AGENT.md`](https://github.com/talkincode/toughradius/blob/main/AGENT.md)；该文件
保持权威地位，并被 agent 工具链直接引用。本章不替代它——如有歧义，以
`AGENT.md` 为准。

## 产品范围基线

开发始终锚定功能清单，绝不漂移到无关的产品方向。

- 权威范围基线为
  [`docs/feature-checklist.md`](https://github.com/talkincode/toughradius/blob/main/docs/feature-checklist.md)
  （英文版见
  [`docs/feature-checklist.en.md`](https://github.com/talkincode/toughradius/blob/main/docs/feature-checklist.en.md)）。
- 每个任务、issue、PR、测试与评审记录都映射到形如 `TR-F004` 的功能编号。
- 若某需求无法映射到现有编号，先更新功能清单（范围、状态、验收边界、理由），
  再改动代码。
- 非目标 `TR-N001`–`TR-N005`（支付、CRM、通用监控栈、多租户、整体重写）除非先
  显式修订清单，否则一律不在范围内。

## 路线图与技能库

agent 驱动开发围绕三份产物组织：

- [`docs/roadmap.md`](https://github.com/talkincode/toughradius/blob/main/docs/roadmap.md)
  —— 长期路线图与里程碑，每项映射到 `TR-F` 编号，是 agent 工作的任务来源。
- [`.agents/skills/`](https://github.com/talkincode/toughradius/tree/main/.agents/skills)
  —— 可复用的技能 SOP，每个技能一个目录（`.agents/skills/<name>/SKILL.md`）。
- [`.agents/README.md`](https://github.com/talkincode/toughradius/blob/main/.agents/README.md)
  —— 委托参考与共享护栏。

一个**总调度层**驱动循环，执行 SOP 负责领域内的具体工作：

| 角色 | 技能 | 职责 |
| --- | --- | --- |
| 总调度 | `orchestrate-roadmap` | “自动委托开发”入口：选取下一个未勾选子任务、匹配 SOP、执行门禁、开 PR |
| 门禁 | `review-pr` | 以 CI 为锚的独立审查；通过标签/评论打回，仅在审过**且** CI 绿时自动合并 |
| 自我迭代 | `groom-roadmap` | 每次合并后勾选已交付子任务并重新梳理路线图 |

执行 SOP 包括：新增厂商 VSA（`add-radius-vendor`）、新增 EAP 方法
（`add-eap-method`）、新增 Admin API（`add-adminapi-endpoint`）、新增 React Admin
资源（`add-react-admin-resource`）、新增配置项（`add-config-schema`）、新增验收
测试（`add-acceptance-test`）、同步上游 radius（`sync-upstream-radius`）、引用 RFC
（`reference-rfc`）、对齐清单（`align-feature-checklist`）、编写 Go 测试
（`write-go-tests`）、编写 Go API 文档（`document-go-apis`）。开始某类任务前先选用
匹配的技能。

agent **在你自己的主机**上用你自己的 agent/CLI 运行，而非通过 CI 工作流执行，
因此密钥不会进入 CI，执行环境完全自控。

## 工作准则

### 动手前先理解现有代码

绝不盲目改代码。先用检索定位现有实现、相关测试与文档，再模仿项目的命名、错误
处理与数据流。修 bug 前先追踪完整执行路径，重构前先梳理依赖与副作用。

### 持续验证

不要等到最后才跑测试。每个逻辑改动后都跑一次测试，使回归立即暴露，而不是堆到
大批量改动的末尾。

### 代码是最好的文档

- **所有导出 API 均带完整 godoc 注释**——用途、带约束的参数、返回值与错误条件
  （复杂 API 附使用示例）。标准库风格的约定见 `document-go-apis` 技能。
- **复杂逻辑带行内注释**，解释*为什么*而非*做什么*。
- **厂商相关代码引用协议规范**（RFC 编号、VSA 文档）。
- **不产出冗余的独立总结文档**——信息应留在代码注释与 Git 历史中。

## 核心开发原则

### 测试驱动开发（TDD）

先写测试再写代码：用失败的测试定义预期行为（红），写最少代码使其通过（绿），
随后在测试保持通过的前提下重构。若改动 `internal/radiusd/auth.go`，测试位于
`internal/radiusd/auth_test.go`。统一约定见 `write-go-tests` 技能。

### GitHub 工作流

- **仅走 Pull Request**——禁止直接推 `main`；受保护分支会拒绝直接推送。
- **约定式提交**——`<type>(<scope>): <subject>`，类型如 `feat`、`fix`、`test`、
  `docs`、`refactor`、`perf`、`chore`。
- **小而原子的改动**优于巨型 PR。

### 仓库 issue/PR 自动化

仓库通过若干 GitHub Actions 做轻量分流，帮助维护者扫描队列；它们不替代阅读
原始 issue、PR diff 与 CI 输出。

| Workflow | 触发与结果 | 维护者注意事项 |
| --- | --- | --- |
| AI issue summary | `.github/workflows/ai-issue-summary.yml` 在 issue opened 时运行，并用 GitHub Models 生成简短摘要作为 issue 评论。 | issue 标题和正文是不可信输入。自动摘要只作阅读便利，不能作为权威事实；判断仍以 issue 原文为准。 |
| Stale | `.github/workflows/stale.yml` 每天 `04:24 UTC` 自动运行，也可手动触发。60 天无活动后添加 `stale`，再过 14 天仍无活动则关闭。 | 评论、推送提交或移除 `stale` 可保活。带 `pinned`、`security`、`help wanted`、`agent-roadmap`、`needs-human` 的 issue 豁免；带 `pinned`、`security`、`agent-roadmap`、`needs-human` 的 PR 豁免；所有 milestone 豁免。 |
| Labeler | `.github/workflows/labeler.yml` 在 `pull_request_target` 上运行，并按 `.github/labeler.yml` 的路径规则打标签。 | 自动标签包括 `go`、`javascript`、`github_actions`、`dependencies`、`doc`。该 action 只读取变更文件列表和基础分支配置，不 checkout 或执行 PR 代码。 |
| Greetings | `.github/workflows/greetings.yml` 在贡献者首次打开 issue 或 PR 时运行，并发送入门提示评论。 | 评论仅用于引导，不改变评审要求或 issue 优先级。 |

### 最小可行产品（MVP）

每次改动以最小、可独立使用、可回滚且不破坏既有行为的单元交付。大型工作拆成
MVP 增量（例如：厂商属性解析 → 认证集成 → 计费 → 管理界面），而非一次性塞进
一个超大 PR。

## 质量门禁

每个 agent 改动在合并前必须通过以下门禁：

- `go build ./...` —— 无编译错误。
- `go test ./...` —— 全部单元测试通过。
- `golangci-lint run` —— 干净（钉 **v2.12.2**，与 CI 一致）。
- `cd web && npm run build` —— 任何前端改动。
- **协议 / 端到端改动**在
  [`test/integration/`](https://github.com/talkincode/toughradius/tree/main/test/integration)
  下附 CI 可执行的验收测试，并引用
  [`docs/rfcs/`](https://github.com/talkincode/toughradius/tree/main/docs/rfcs)
  下对应的规范。
- 产出一律走打了 `agent-roadmap` 标签的 PR，由 `review-pr` 门禁把关，仅在
  `agent-approved` 且 CI 全绿时合并。

## 技术约束

- **禁用 CGO**——项目以 `CGO_ENABLED=0` 构建以便跨平台部署。仅使用纯 Go 驱动
  （例如用 `github.com/glebarez/sqlite` 而非 `github.com/mattn/go-sqlite3`）。
- **数据库双兼容**——每次 schema 变更都必须同时兼容 PostgreSQL（默认）与 SQLite。
- **上游依赖**——核心库 `layeh.com/radius` 经 `go.mod` `replace` 指向组织 fork
  `github.com/talkincode/radius`；上游重要修复通过 `sync-upstream-radius` 技能评估。

## 常见反模式（禁止）

- 导出 API 不写文档。
- 不写测试就提交实现。
- 混杂多个关注点的巨型 PR。
- 先实现后补测试。
- 直接推 `main` 或跳过评审。
- 产出冗余的独立总结/报告文档。
- 引入 CGO 依赖。

## 下一步

- [`AGENT.md`](https://github.com/talkincode/toughradius/blob/main/AGENT.md)
  —— 完整、权威的 agent 开发指南。
- [文档地图](./documentation-map.md) —— 找到 README、安全策略、功能清单、路线图
  与 RFC 索引。
- [协议与 RFC 索引](./rfc-index.md) —— 协议标准与代码、里程碑的对应关系。
