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
4. **质量门禁**：`go build ./...`、`go test ./...`、`golangci-lint run`（v2.12.2）必须通过；前端改动跑 `cd web && npm run build`。
5. **数据库双兼容**：schema 变更必须同时兼容 PostgreSQL 与 SQLite。
6. **走 PR**：禁止直接推 `main`，PR 描述引用里程碑与功能编号。
7. **协议合规**：协议行为改动必须引用 `docs/rfcs/` 对应 RFC 条款（见 `skills/reference-rfc/SKILL.md`）。
8. **CI 验收测试**：协议 / 端到端改动必须带 CI 可自动执行的验收用例（`test/integration/`，见 `skills/add-acceptance-test/SKILL.md`）。
9. **上游依赖**：核心库 `layeh.com/radius` 经 `go.mod` `replace` 指向组织 fork `github.com/talkincode/radius`；上游重要修复按 `skills/sync-upstream-radius/SKILL.md` 评估同步。

## 技能索引

> **总调度层**：`orchestrate-roadmap` 是入口角色——收到"自动委托开发"类指令时由它统筹全程，开 PR 后交 `review-pr` 独立审查门禁（审过且 CI 绿才自动合并），并在交付后调用 `groom-roadmap` 迭代计划；其余技能是被它编排的执行 SOP。

| 技能 | 适用场景 | 关联编号 |
| --- | --- | --- |
| [orchestrate-roadmap](skills/orchestrate-roadmap/SKILL.md) | **总调度**：接自动委托指令，统筹 选任务→选 SOP→派工→门禁→PR→审查→合并→迭代 | 全部 / TR-F022 |
| [review-pr](skills/review-pr/SKILL.md) | **审查门禁**：对委托 PR 做对抗式、以 CI 为锚的独立审查；打回用 `needs-rework`，审过且 CI 绿则自动合并 | TR-F022 |
| [groom-roadmap](skills/groom-roadmap/SKILL.md) | 交付后自我迭代路线图与计划（勾选 / 补子任务 / 重排 / 对齐清单） | TR-F022 |
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

## 工具链版本（与 CI 对齐）

- Go `1.25`，`CGO_ENABLED=0`
- Node `20`，前端在 `web/`，`npm ci && npm run build`
- golangci-lint `v2.12.2`


## 在自己的主机上运行 Agent（不在 CI 里执行）

> 本仓库**不内置任何自动执行 agent 的 GitHub workflow**。路线图、技能库与约束是供 agent 使用的"知识与护栏"；具体执行由你用**自己的 agent**在**自己的主机**上手动或定时运行。这样密钥不进 CI，执行环境完全自控。

### 委托一轮开发

用任意支持工具调用的编码 agent 即可，本仓库不绑定具体 agent 或 CLI。委托一轮开发时，让 agent 严格按 [`orchestrate-roadmap`](skills/orchestrate-roadmap/SKILL.md) 统筹：**先清在途 PR**（按 [`review-pr`](skills/review-pr/SKILL.md) 处理 `needs-rework` 的评论、合并已审过且 CI 绿的）；再读 `AGENT.md`、本文件、`docs/roadmap.md`、`docs/feature-checklist.md`；自上而下取第一个未勾选的 `- [ ] M*.*` 子任务；选用匹配的执行 SKILL；只做最小闭环；协议改动引用 `docs/rfcs/`；补 CI 可执行测试；通过质量门禁；改动走 PR（打 `agent-roadmap` 标签），交 `review-pr` 审查——审过且 CI 绿则自动合并，有问题打 `needs-rework` 留待下一轮处理；PR 合并后按 [`groom-roadmap`](skills/groom-roadmap/SKILL.md) 勾选并迭代路线图。

> **闭环要点**：路线图的 `- [x]` 只在**合并进 `main`** 后才前进。每轮先清在途 PR、审过即自动合并，下一轮才不会重选同一任务、避免重复冲突的 PR。审查与合并默认在**单一身份**下用 `label + COMMENT review + CI 门禁`实现（GitHub 不允许作者审批自己的 PR）；若提供独立的审查身份/Bot token，可改用 GitHub 正式的 approve/request_changes 状态。

### 给 agent 的护栏（无论在哪运行都适用）

- **锚定功能编号**：任务必须映射到 `TR-F` 编号；严禁触碰非目标 `TR-N001`~`TR-N005`（支付/CRM/通用监控/多租户/重写）。
- **选任务口径**：`docs/roadmap.md` 自上而下第一个未勾选的 `- [ ] M*.*`；选之前先按 `skills/review-pr/SKILL.md` 清在途 `agent-roadmap` PR。
- **遵循 SOP**：按任务类型选用对应 `.agents/skills/<name>/SKILL.md`。
- **质量门禁**：`go build ./...`、`go test ./...`、`golangci-lint run`（v2.12.2）通过；前端改动跑 `cd web && npm run build`。
- **协议合规**：协议行为改动引用 `docs/rfcs/` 对应 RFC（见 `skills/reference-rfc/SKILL.md`）。
- **CI 验收测试**：协议 / 端到端改动带 CI 可自动执行的验收用例（`test/integration/`，见 `skills/add-acceptance-test/SKILL.md`）。
- **走 PR + 审查门禁**：产出一律 Pull Request，禁止直接推 `main`；PR 打 `agent-roadmap` 标签并交 `skills/review-pr/SKILL.md` 审查，**仅在 `agent-approved` 且 CI 全绿时自动合并**，有阻断问题打 `needs-rework`、超轮次打 `needs-human` 交人。
- **PR 模板**：agent 任务 PR 默认使用 [`.github/pull_request_template.md`](../.github/pull_request_template.md)，完整填写里程碑子任务、`TR-F`、验收与门禁结果。
- **审查清单**：review gate 按 [`.github/review-checklists/agent-roadmap.md`](../.github/review-checklists/agent-roadmap.md) 逐项判定，避免凭主观印象放行。

### Agent 质量度量（M4.8 / TR-F022）

为收敛 agent 产出质量，仓库提供脚本 `scripts/agent-roadmap-quality-metrics.sh`，统一统计：

- **CI 通过率**：统计窗口内、已合并 `agent-roadmap` PR 关联的已完成 CI 工作流运行（`success / total`）。
- **回滚率**：统计窗口内、已合并 `agent-roadmap` PR 中，其合并提交随后被 `main` 上 `This reverts commit <sha>` 回滚的比例。

示例：

```bash
scripts/agent-roadmap-quality-metrics.sh \
  --days 30 \
  --json-output .agents/reports/agent-roadmap-quality.json \
  --markdown-output .agents/reports/agent-roadmap-quality.md
```

建议每轮自动委托开发结束后执行一次，并在回滚率或失败 CI 增高时优先做根因修复，不扩展新范围。

### 上游 radius 库跟踪（手动）

核心库 `layeh.com/radius` 经 `go.mod` `replace` 指向组织 fork `github.com/talkincode/radius`。无自动巡检 workflow，按 `skills/sync-upstream-radius/SKILL.md` 的步骤定期人工核对上游是否有安全 / 协议修复并决定是否同步。
