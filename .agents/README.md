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

| 技能 | 适用场景 | 关联编号 |
| --- | --- | --- |
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

> 本仓库**不内置任何自动执行 agent 的 GitHub workflow**。路线图、技能库与约束是供 agent 使用的"知识与护栏"；具体执行由你在**自己的主机**上手动或定时运行（例如 Codex CLI 无头模式）。这样密钥不进 CI，执行环境完全自控。

### 本地运行 Codex（无头）参考

1. 安装：`npm install -g @openai/codex`
2. 配置 `~/.codex/config.toml`（Azure 端点示例；API Key 经 `env_key` 从环境变量读取，**不要写进配置文件**）：

```toml
preferred_auth_method = "apikey"
model = "gpt-5.5"
model_provider = "azure"
model_reasoning_effort = "high"
personality = "pragmatic"

[model_providers.azure]
name = "Azure"
base_url = "<你的 Azure 端点>"
env_key = "WJT_AZURE_OPENAI_API_KEY"
wire_api = "responses"
service_tier = "priority"
```

3. 导出密钥、选取路线图下一个未完成子任务后运行：

```bash
export WJT_AZURE_OPENAI_API_KEY=...   # 仅留在你本机环境，不入库、不落配置
task="$(grep -nE '^- \[ \] M[0-9]+\.[0-9]+' docs/roadmap.md | head -1)"
codex exec "实现路线图子任务：$task。严格遵循 AGENT.md、docs/feature-checklist.md 与 .agents/skills/<对应技能>/SKILL.md；只做最小闭环；协议改动引用 docs/rfcs/；补 CI 可执行的测试；通过 go build/test、golangci-lint、（涉及前端）npm run build；改动走 PR，禁止直接推 main。"
```

> `sandbox_mode = "danger-full-access"` 或 `--dangerously-bypass-approvals-and-sandbox` 这类"免审批全权限"参数，请仅在你信任的隔离环境里自行决定是否启用，本仓库不预设。

### 给 agent 的护栏（无论在哪运行都适用）

- **锚定功能编号**：任务必须映射到 `TR-F` 编号；严禁触碰非目标 `TR-N001`~`TR-N005`（支付/CRM/通用监控/多租户/重写）。
- **选任务口径**：`docs/roadmap.md` 自上而下第一个未勾选的 `- [ ] M*.*`。
- **遵循 SOP**：按任务类型选用对应 `.agents/skills/<name>/SKILL.md`。
- **质量门禁**：`go build ./...`、`go test ./...`、`golangci-lint run`（v2.12.2）通过；前端改动跑 `cd web && npm run build`。
- **协议合规**：协议行为改动引用 `docs/rfcs/` 对应 RFC（见 `skills/reference-rfc/SKILL.md`）。
- **CI 验收测试**：协议 / 端到端改动带 CI 可自动执行的验收用例（`test/integration/`，见 `skills/add-acceptance-test/SKILL.md`）。
- **走 PR**：产出一律 Pull Request + Review，禁止直接推 `main`。

### 上游 radius 库跟踪（手动）

核心库 `layeh.com/radius` 经 `go.mod` `replace` 指向组织 fork `github.com/talkincode/radius`。无自动巡检 workflow，按 `skills/sync-upstream-radius/SKILL.md` 的步骤定期人工核对上游是否有安全 / 协议修复并决定是否同步。
