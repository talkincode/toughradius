# ToughRADIUS 开发路线图 (Roadmap)

> 本路线图是 ToughRADIUS 的长期开发计划，与功能清单 [`docs/feature-checklist.md`](feature-checklist.md) 强绑定。
> 所有里程碑任务必须引用至少一个功能编号（如 `TR-F004`），不在清单中的方向先更新清单再排期。
> 本路线图同时是 **Agent 驱动开发** 的任务来源：agent 由你用自己的 agent/CLI 在**自己的主机**上运行，从这里自上而下取第一个未完成子任务实现并提 PR；不在 CI 中执行。运行方式与护栏见 [`.agents/README.md`](../.agents/README.md)。

## 维护规则

1. 里程碑按 `M<序号>` 编号，每个里程碑映射到一个或多个 `TR-F` 功能编号。
2. 每个里程碑必须拆分为可独立交付、可回滚、可验证的 MVP 子任务。
3. 状态流转：`计划中 → 进行中 → 已交付`；交付以合并到 `main` 且 CI 通过为准。
4. 严禁触碰非目标方向（见功能清单 `TR-N001`~`TR-N005`）：支付/订单、CRM/工单、通用监控平台、多租户 SaaS、重写协议栈或管理框架。
5. Agent 产出一律走 Pull Request + CI + 人工 Review，禁止直接推送 `main`。
6. **自我迭代**：每交付一个子任务后，按 [`.agents/skills/groom-roadmap/SKILL.md`](../.agents/skills/groom-roadmap/SKILL.md) 勾选、更新里程碑状态、回填/拆分/重排子任务，并把新发现需求经功能清单纳入，保持本路线图与清单自洽。

## 状态说明

| 状态 | 含义 |
| --- | --- |
| 计划中 | 已排期未开始 |
| 进行中 | 有进行中的 PR / Issue |
| 已交付 | 已合并且 CI 通过 |
| 阻塞 | 有外部依赖或待决策 |

## 里程碑总览

| 里程碑 | 主题 | 关联编号 | 优先级 | 状态 |
| --- | --- | --- | --- | --- |
| M1 | EAP-TLS 认证支持 | TR-F004 | P1 | 进行中 |
| M2 | CoA 动态授权支持 | TR-F010 / TR-F012 / TR-F013 | P1 | 计划中 |
| M3 | IPv6 能力增强闭环 | TR-F007 / TR-F011 / TR-F015 | P1 | 计划中 |
| M4 | Agent 开发体系与质量门禁 | TR-F022 | P2 | 进行中 |
| M5 | 厂商 VSA 覆盖扩展 | TR-F005 | P2 | 计划中 |
| M6 | 可观测性与运维增强 | TR-F015 | P3 | 计划中 |
| M7 | 上游 RADIUS 库跟踪与协议合规 | TR-F021 / TR-F022 | P2 | 进行中 |

---

## 依赖与协议合规基线（横切约束）

以下为所有里程碑共享的横切要求，agent 在任何相关改动中都必须遵守。

### 上游 RADIUS 库

- 核心协议库：`layeh.com/radius`（原始仓库 <https://github.com/layeh/radius>）。
- 组织 fork：`github.com/talkincode/radius` <https://github.com/talkincode/radius>，通过 `go.mod` 的 `replace layeh.com/radius => github.com/talkincode/radius` 接入。
- **跟踪要求**：原始仓库若有重要修复（安全、协议正确性、属性编解码），必须评估是否同步到 fork 并更新 `replace` 版本。流程见 `.agents/skills/sync-upstream-radius/SKILL.md`（人工定期核对，无自动巡检 workflow）。

### 国际标准协议规范

- 优先检索仓库内 `docs/rfcs/`（已收录 50+ RFC，含 EAP / EAP-TLS / CoA / IPv6 / RadSec 等）。
- 任何协议行为改动必须引用对应 RFC 条款；本地缺失的规范按 `.agents/skills/reference-rfc/SKILL.md` 补充并登记。

### CI 自动化验收测试

- 每个里程碑子任务的验收口径必须由 **可在 CI 自动执行的测试** 背书，不接受仅人工验证。
- 协议级 / 端到端验收用例放在 `test/integration/`（`//go:build integration` 构建标签），由 CI `integration` job（PostgreSQL 服务，`INTEGRATION_REQUIRED=1`）自动执行；写法见 `.agents/skills/add-acceptance-test/SKILL.md`。
- 纯逻辑验收用 `*_test.go` 单元测试，CI `test` job 执行。

---

## M1 — EAP-TLS 认证支持

- **关联编号**：`TR-F004`
- **目标**：在现有 EAP handler 体系下新增 EAP-TLS，先交付最小可用认证链路，再扩展证书策略。
- **开发边界**：不重写 EAP 协调器；复用 `internal/radiusd/plugins/eap` 注册机制。
- **技能**：`.agents/skills/add-eap-method/SKILL.md`、`.agents/skills/add-acceptance-test/SKILL.md`、`.agents/skills/write-go-tests/SKILL.md`
- **协议规范**：`docs/rfcs/rfc5216-eap-tls.txt`（EAP-TLS）、`rfc3748-eap.txt`（EAP）、`rfc3579-radius-eap-support.txt`（RADIUS 承载 EAP）、`rfc5247-eap-key-management.txt`、`rfc7499-packet-fragmentation.txt`（分片）。
- **参考实现**：BeryJu `radius-eap` <https://github.com/BeryJu/radius-eap>、实现笔记 <https://beryju.io/blog/2025-05-implementing-eap/>（仅作思路参考；注意许可与协议兼容，禁止直接拷贝不兼容代码）。

子任务：
- [x] M1.1 在 `internal/radiusd/plugins/eap` 注册 EAP-TLS handler 骨架与启用列表配置
- [x] M1.2 实现 TLS 握手状态管理与分片重组（参照 RFC 5216 §3 / RFC 7499）
- [ ] M1.3 证书校验（CA 链）与用户身份映射
- [ ] M1.4 明确失败原因 + AuthError 指标 + 单元/集成测试
- [ ] M1.5 在 `config_schemas.json` 增加 EAP-TLS 配置项与默认值
- [ ] M1.6 在 `test/integration/` 增加 EAP-TLS 端到端验收用例（CI 自动执行）

验收口径：EAP-TLS 客户端可完成认证；失败场景有明确拒绝原因和指标；**验收由 `test/integration/` 的 CI 用例背书**，新增逻辑全部有测试覆盖。

## M2 — CoA 动态授权支持

- **关联编号**：`TR-F010` / `TR-F012` / `TR-F013`
- **目标**：抽象 CoA 发送服务，支持对在线用户发起 CoA / Disconnect 动态授权并记录结果。
- **开发边界**：后端先抽象发送服务和审计；前端只暴露可验证的安全动作，禁止前端拼装任意 RADIUS 包。
- **技能**：`.agents/skills/add-adminapi-endpoint/SKILL.md`、`.agents/skills/add-react-admin-resource/SKILL.md`、`.agents/skills/add-acceptance-test/SKILL.md`
- **协议规范**：`docs/rfcs/rfc5176-coa-disconnect.txt`（CoA / Disconnect）、`rfc3576-dynamic-authorization.txt`。

子任务：
- [ ] M2.1 后端抽象 `CoAService`（发送、超时、重试、结果审计）
- [ ] M2.2 Admin API：对在线会话发起 CoA / Disconnect 端点
- [ ] M2.3 审计记录：触发动作、目标会话、结果落库
- [ ] M2.4 前端在线会话页暴露安全动作按钮 + 结果反馈
- [ ] M2.5 单元/集成测试覆盖成功、超时、NAS 拒绝场景
- [ ] M2.6 在 `test/integration/` 增加 CoA/Disconnect 端到端验收用例（CI 自动执行）

验收口径：可对在线用户安全发起动态授权；每次触发有可查询的结果记录；**验收由 `test/integration/` 的 CI 用例背书**。

## M3 — IPv6 能力增强闭环

- **关联编号**：`TR-F007` / `TR-F011` / `TR-F015`
- **目标**：完善 IPv6 地址、IPv6 前缀、Delegated-IPv6-Prefix 在用户、在线会话、计费记录、审计、Dashboard 中的查询与展示。
- **开发边界**：不只做字段展示；协议解析、数据库字段、过滤条件、前端列表、审计口径必须一起闭环。
- **协议规范**：`docs/rfcs/rfc3162-radius-ipv6.txt`、`rfc4818-ipv6-prefix-delegation.txt`（Delegated-IPv6-Prefix）、`rfc6911-ipv6-access-networks.txt`。

子任务：
- [ ] M3.1 协议层解析 / 下发 Delegated-IPv6-Prefix 等属性
- [ ] M3.2 数据库字段与迁移（PostgreSQL + SQLite 双兼容）
- [ ] M3.3 用户 / 会话 / 计费的 IPv6 过滤与展示
- [ ] M3.4 Dashboard IPv6 维度统计
- [ ] M3.5 端到端测试与字段一致性校验（`test/integration/`，CI 自动执行）

验收口径：IPv6 全链路可查询、可过滤、可审计，双数据库一致；**验收由 `test/integration/` 的 CI 用例背书**。

## M4 — Agent 开发体系与质量门禁

- **关联编号**：`TR-F022`
- **目标**：建立可持续的 agent 驱动开发流程、技能库与质量门禁。
- **状态**：进行中

子任务：
- [x] M4.1 建立路线图与里程碑（本文件）
- [x] M4.2 建立 `.agents/skills` 技能库
- [x] M4.3 制定 agent 通用护栏与质量门禁（`AGENT.md` / `.agents/README.md` / `.github/copilot-instructions.md`）
- [x] M4.6 协议规范检索技能与 CI 验收测试技能
- [x] M4.9 约定本机无头运行 agent 的方式与护栏（不在 CI 执行；见 `.agents/README.md`）
- [x] M4.10 建立总调度与自我迭代技能（`orchestrate-roadmap` 统筹委托循环 + `groom-roadmap` 路线图自我迭代）
- [ ] M4.7 为 agent 任务建立 PR 模板与 review checklist
- [ ] M4.8 收敛 agent 产出质量度量（CI 通过率、回滚率）

## M5 — 厂商 VSA 覆盖扩展

- **关联编号**：`TR-F005`
- **目标**：按 parser / enhancer / registry 模式扩展更多厂商 VSA 覆盖，补齐样例包测试。
- **技能**：`.agents/skills/add-radius-vendor/SKILL.md`

子任务：
- [ ] M5.1 梳理待补厂商清单与字典差异
- [ ] M5.2 逐厂商按现有模式接入 parser / enhancer
- [ ] M5.3 厂商样例包覆盖解析与响应属性

## M6 — 可观测性与运维增强

- **关联编号**：`TR-F015`
- **目标**：在不替代 Prometheus/Grafana 的前提下，增强 RADIUS 运维视图指标维度。
- **边界**：遵守 `TR-N003`，不扩展为通用监控平台。

子任务：
- [ ] M6.1 补充认证/计费失败分类指标
- [ ] M6.2 Dashboard 趋势维度扩展
- [ ] M6.3 指标命名兼容性回归

## M7 — 上游 RADIUS 库跟踪与协议合规

- **关联编号**：`TR-F021` / `TR-F022`
- **目标**：持续跟踪 `layeh.com/radius` 原始仓库与 `talkincode/radius` fork，及时评估并同步重要修复；保持协议实现与国际标准规范一致。
- **技能**：`.agents/skills/sync-upstream-radius/SKILL.md`、`.agents/skills/reference-rfc/SKILL.md`、`.agents/skills/add-acceptance-test/SKILL.md`

子任务：
- [ ] M7.1 人工核对上游发现新提交后，评估安全/协议修复并决定是否同步 fork（更新 `go.mod` replace 版本）
- [ ] M7.2 同步后跑全量 + 集成测试，必要时补充回归用例
- [ ] M7.3 定期核对协议行为与 `docs/rfcs/` 规范，缺失规范按技能补录
- [ ] M7.4 关键协议路径（认证、计费、CoA、EAP）补齐 CI 自动化验收用例

验收口径：上游重要修复有评估记录与同步决策；协议改动均引用 RFC 并有 CI 验收用例背书。

---

## Agent 排期约定

- **入口（自动委托）**：收到"自动委托开发 / 继续推进路线图"类指令时，由 [`.agents/skills/orchestrate-roadmap/SKILL.md`](../.agents/skills/orchestrate-roadmap/SKILL.md) 作为总调度统筹一轮：选题 → 选执行 SOP → 派工 → 质量门禁 → PR → 迭代路线图。
- 调度优先级：`M1 → M2 → M3`，P2/P3 在 P1 里程碑无可执行子任务时填充。
- 单次 agent 任务只认领一个未勾选子任务（最小闭环），完成后在本文件勾选并在 PR 引用里程碑编号。
- **自我迭代**：每轮交付后由 [`.agents/skills/groom-roadmap/SKILL.md`](../.agents/skills/groom-roadmap/SKILL.md) 勾选已交付项、更新里程碑状态、回填/拆分/重排子任务，并保持本文件与功能清单状态一致。
- 任何超出 `TR-F` 清单的需求，必须先提交清单更新 PR，再排入本路线图。
- 每个涉及协议或数据流的子任务，交付时必须带 **CI 可自动执行的验收测试**（单元或 `test/integration/`）。
- 选任务口径：`docs/roadmap.md` 自上而下第一个未勾选的 `- [ ] M*.*`。agent 在**本机**用你自己的 agent 运行，不在 CI 执行；运行参考与护栏见 [`.agents/README.md`](../.agents/README.md)。

