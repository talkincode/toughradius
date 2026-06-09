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
| M1 | EAP-TLS 认证支持 | TR-F004 | P1 | 已交付 |
| M2 | CoA 动态授权支持 | TR-F010 / TR-F012 / TR-F013 | P1 | 已交付 |
| M3 | IPv6 能力增强闭环 | TR-F007 / TR-F011 / TR-F015 | P1 | 进行中 |
| M4 | Agent 开发体系与质量门禁 | TR-F022 | P2 | 进行中 |
| M5 | 厂商 VSA 覆盖扩展 | TR-F005 | P2 | 计划中 |
| M6 | 可观测性与运维增强 | TR-F015 | P3 | 计划中 |
| M7 | 上游 RADIUS 库跟踪与协议合规 | TR-F021 / TR-F022 | P2 | 进行中 |
| M8 | PEAPv0 / EAP-MSCHAPv2 认证支持 | TR-F004 | P1 | 计划中 |
| M9 | EAP-TTLS 隧道认证支持 | TR-F004 | P1 | 计划中 |
| M10 | EAP-TLS 1.3 / RFC 9190 升级 | TR-F004 | P2 | 计划中 |
| M11 | TEAP 隧道认证（中长期） | TR-F004 | P3 | 计划中 |
| M12 | EAP-PWD 口令认证（按需） | TR-F004 | P3 | 计划中 |
| M13 | 双语文档站点（mdbook） | TR-F023 | P2 | 计划中 |

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
- [x] M1.3 证书校验（CA 链）与用户身份映射
- [x] M1.4 明确失败原因 + AuthError 指标 + 单元/集成测试
- [x] M1.5 在 `config_schemas.json` 增加 EAP-TLS 配置项与默认值
- [x] M1.6 在 `test/integration/` 增加 EAP-TLS 端到端验收用例（CI 自动执行）

验收口径：EAP-TLS 客户端可完成认证；失败场景有明确拒绝原因和指标；**验收由 `test/integration/` 的 CI 用例背书**，新增逻辑全部有测试覆盖。

## M2 — CoA 动态授权支持

- **关联编号**：`TR-F010` / `TR-F012` / `TR-F013`
- **目标**：抽象 CoA 发送服务，支持对在线用户发起 CoA / Disconnect 动态授权并记录结果。
- **开发边界**：后端先抽象发送服务和审计；前端只暴露可验证的安全动作，禁止前端拼装任意 RADIUS 包。
- **技能**：`.agents/skills/add-adminapi-endpoint/SKILL.md`、`.agents/skills/add-react-admin-resource/SKILL.md`、`.agents/skills/add-acceptance-test/SKILL.md`
- **协议规范**：`docs/rfcs/rfc5176-coa-disconnect.txt`（CoA / Disconnect）、`rfc3576-dynamic-authorization.txt`。

子任务：
- [x] M2.1 后端抽象 `CoAService`（发送、超时、重试、结果审计）
- [x] M2.2 Admin API：对在线会话发起 CoA / Disconnect 端点
- [x] M2.3 审计记录：触发动作、目标会话、结果落库
- [x] M2.4 前端在线会话页暴露安全动作按钮 + 结果反馈
- [x] M2.5 单元/集成测试覆盖成功、超时、NAS 拒绝场景
- [x] M2.6 在 `test/integration/` 增加 CoA/Disconnect 端到端验收用例（CI 自动执行）

验收口径：可对在线用户安全发起动态授权；每次触发有可查询的结果记录；**验收由 `test/integration/` 的 CI 用例背书**。

## M3 — IPv6 能力增强闭环

- **关联编号**：`TR-F007` / `TR-F011` / `TR-F015`
- **目标**：完善 IPv6 地址、IPv6 前缀、Delegated-IPv6-Prefix 在用户、在线会话、计费记录、审计、Dashboard 中的查询与展示。
- **开发边界**：不只做字段展示；协议解析、数据库字段、过滤条件、前端列表、审计口径必须一起闭环。
- **协议规范**：`docs/rfcs/rfc3162-radius-ipv6.txt`、`rfc4818-ipv6-prefix-delegation.txt`（Delegated-IPv6-Prefix）、`rfc6911-ipv6-access-networks.txt`。

子任务：
- [x] M3.1 协议层解析 / 下发 IPv6 访问属性（计费侧解析 Framed-IPv6-Address(RFC 6911)、修正 Framed/Delegated-IPv6-Prefix 缺省值落库；Access-Accept 下发 Framed-IPv6-Address）
- [x] M3.2 数据库字段与迁移（PostgreSQL + SQLite 双兼容）：新增 `RadiusUser.DelegatedIpv6Prefix`（静态 RFC 4818 #123，按用户）、`RadiusUser.DelegatedIpv6PrefixPool` 与 `RadiusProfile.DelegatedIpv6PrefixPool`（RFC 6911 #171 DHCPv6-PD 池，按 §2.4 与 Framed-IPv6-Pool 区分）；GORM AutoMigrate 双库建列，Admin API 用户/套餐增改可持久化并支持 Profile 继承
- [x] M3.3 用户 / 会话 / 计费的 IPv6 过滤与展示（会话/计费侧 IPv6 过滤与展示此前已闭环；本轮补齐用户侧：Admin API 用户列表新增 `ipv6_addr`(RFC 6911)/`delegated_ipv6_prefix`(RFC 4818) 过滤，前端用户编辑表单与详情页展示 `ipv6_prefix_pool`/`delegated_ipv6_prefix`/`delegated_ipv6_prefix_pool`（zh/en 双语），并修复用户静态 IPv6 地址更新写入幽灵列 `ipv6_addr` 导致从不落库的历史缺陷（正确列名 `ip_v6_addr`））
- [x] M3.4 Dashboard IPv6 维度统计（在线会话 IPv6 占比/地址/Framed 前缀/委派前缀此前已闭环但缺测试；本轮新增用户库静态 IPv6 配置维度（已配置静态 IPv6 地址 / 委派前缀用户数），补齐 `TestGetDashboardIPv6Stats` 回归测试覆盖在线与用户两个维度，前端 IPv6 覆盖面板新增用户维度卡片（zh/en 双语））
- [ ] M3.5 端到端测试与字段一致性校验（`test/integration/`，CI 自动执行）
- [ ] M3.6 下发 Delegated-IPv6-Prefix / Delegated-IPv6-Prefix-Pool（依赖 M3.2 新增用户/Profile 配置字段；RFC 4818 / RFC 6911，按 RFC 6911 §2.4 与 Framed-IPv6-Pool 区分用途，禁止混用同一字段）

验收口径：IPv6 全链路可查询、可过滤、可审计，双数据库一致；**验收由 `test/integration/` 的 CI 用例背书**。

## M4 — Agent 开发体系与质量门禁

- **关联编号**：`TR-F022` / `TR-F024`
- **目标**：建立可持续的 agent 驱动开发流程、技能库与质量门禁，并对齐标准库风格的 Go API 文档规范。
- **状态**：进行中

子任务：
- [x] M4.1 建立路线图与里程碑（本文件）
- [x] M4.2 建立 `.agents/skills` 技能库
- [x] M4.3 制定 agent 通用护栏与质量门禁（`AGENT.md` / `.agents/README.md` / `.github/copilot-instructions.md`）
- [x] M4.6 协议规范检索技能与 CI 验收测试技能
- [x] M4.9 约定本机无头运行 agent 的方式与护栏（不在 CI 执行；见 `.agents/README.md`）
- [x] M4.10 建立总调度与自我迭代技能（`orchestrate-roadmap` 统筹委托循环 + `groom-roadmap` 路线图自我迭代）
- [x] M4.11 制定 Go API 文档/注释规范并建立 `document-go-apis` 技能（标准库 godoc 风格，关联 `TR-F024`）
- [ ] M4.7 为 agent 任务建立 PR 模板与 review checklist
- [ ] M4.8 收敛 agent 产出质量度量（CI 通过率、回滚率）
- [ ] M4.12 按模块增量补齐导出标识符 godoc 注释与包注释（顺序建议：`internal/adminapi` → `internal/radiusd` → `pkg`），并探索 lint 度量（`TR-F024`）

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

## M8 — PEAPv0 / EAP-MSCHAPv2 认证支持

- **关联编号**：`TR-F004`
- **目标**：在现有 EAP handler 体系下，用服务器证书建立 PEAP TLS 隧道，隧道内运行 EAP-MSCHAPv2，为 Windows / AD / 传统企业网络提供兼容认证，并正确导出 MPPE 会话密钥。
- **定位**：兼容性优先，**不是安全先进性卖点**。MS-CHAPv2 类连接存在类似 NTLMv1 的攻击面（见 Microsoft 文档），适合“必须服务一堆旧设备与 AD 用户”的场景；文档与配置必须明示该风险，默认不削弱外层 TLS 强度。
- **开发边界**：不重写 EAP 协调器；复用 `internal/radiusd/plugins/eap` 注册与状态管理，PEAP 隧道分片与 EAP-TLS 机制对齐；内层 MS-CHAPv2 复用现有 `mschapv2_handler` 校验逻辑。
- **技能**：`.agents/skills/add-eap-method/SKILL.md`、`.agents/skills/add-config-schema/SKILL.md`、`.agents/skills/add-acceptance-test/SKILL.md`、`.agents/skills/write-go-tests/SKILL.md`
- **协议规范**：`docs/rfcs/rfc3748-eap.txt`（EAP）、`rfc2759-mschapv2.txt`（MS-CHAP-V2）、`rfc2548-microsoft-vsa.txt`（MS-MPPE 密钥）、`rfc3579-radius-eap-support.txt`；PEAPv0 无正式 RFC，参考 Microsoft `[MS-PEAP]` 与 IETF `draft-kamath-pppext-peapv0`（本地缺失，按 `reference-rfc` 登记）。

子任务：
- [ ] M8.1 注册 PEAP handler 骨架与启用列表配置（`EapMethod`）
- [ ] M8.2 PEAP 外层 TLS 隧道建立与分片重组（复用 EAP-TLS 状态机）
- [ ] M8.3 隧道内 EAP-MSCHAPv2 交换（复用 mschapv2 校验）与 MPPE 密钥导出
- [ ] M8.4 明确失败原因 + AuthError 指标 + 单元测试；配置项与默认值（含安全风险说明）
- [ ] M8.5 在 `test/integration/` 增加 PEAP-MSCHAPv2 端到端验收用例（CI 自动执行）

验收口径：PEAP-MSCHAPv2 客户端（Windows/AD 典型配置）可完成认证，MPPE 密钥正确下发；失败有明确拒绝原因与指标；**验收由 `test/integration/` 的 CI 用例背书**。

## M9 — EAP-TTLS 隧道认证支持

- **关联编号**：`TR-F004`
- **目标**：按 RFC 5281 用服务器证书建立 TLS 隧道，隧道内承载 PAP / CHAP / MS-CHAP / MS-CHAP-V2 内层认证，让 LDAP、老账号库、混合客户端无需立即改造证书体系即可接入。
- **定位**：后端适配优先。现实价值是“后端用户库不用立刻改成证书体系”——先用服务器证书保护隧道，再把用户名口令塞进去；很多企业认证体系是历史债务盘出来的，TTLS 是务实的过渡。
- **开发边界**：内层方法逐个交付（先 PAP，再 MS-CHAP-V2）；后端用户库适配走现有认证流水线，不在协议入口写库分支；不重写 EAP 协调器。
- **技能**：`.agents/skills/add-eap-method/SKILL.md`、`.agents/skills/add-acceptance-test/SKILL.md`、`.agents/skills/write-go-tests/SKILL.md`
- **协议规范**：`docs/rfcs/rfc5281-eap-ttls.txt`（EAP-TTLS）、`rfc3748-eap.txt`、`rfc2759-mschapv2.txt`（内层 MS-CHAP-V2）、`rfc3579-radius-eap-support.txt`；TLS 1.3 隧道按 `rfc9427`（本地缺失，按 `reference-rfc` 登记）。

子任务：
- [ ] M9.1 注册 EAP-TTLS handler 骨架与启用列表配置
- [ ] M9.2 外层 TLS 隧道建立与分片重组（复用 EAP-TLS 状态机）
- [ ] M9.3 隧道内 AVP 封装与 PAP 内层认证（最小可用闭环）
- [ ] M9.4 增加 MS-CHAP-V2 内层认证与密钥导出
- [ ] M9.5 明确失败原因 + 指标 + 单元测试；配置项与默认值
- [ ] M9.6 在 `test/integration/` 增加 TTLS-PAP / TTLS-MSCHAPv2 端到端验收用例（CI 自动执行）

验收口径：TTLS 客户端可经隧道用 PAP / MS-CHAP-V2 完成认证，内层 AVP 解析正确；失败有明确原因与指标；**验收由 `test/integration/` 的 CI 用例背书**。

## M10 — EAP-TLS 1.3 / RFC 9190 升级

- **关联编号**：`TR-F004`
- **目标**：在 M1 已交付的 TLS 1.2 EAP-TLS 基线上，按 RFC 9190 支持 TLS 1.3 握手与会话密钥派生。
- **开发边界**：保持与 TLS 1.2 客户端向后兼容；先协商再切换，不破坏既有 CA 链校验与身份映射；遵循 RFC 9427 的 TLS 1.3 派生规则。
- **技能**：`.agents/skills/add-eap-method/SKILL.md`、`.agents/skills/reference-rfc/SKILL.md`、`.agents/skills/add-acceptance-test/SKILL.md`、`.agents/skills/write-go-tests/SKILL.md`
- **协议规范**：`rfc9190`（EAP-TLS 1.3，本地缺失，按 `reference-rfc` 登记）、`rfc9427`（TLS-Based EAP Types and TLS 1.3，本地缺失）、`docs/rfcs/rfc5216-eap-tls.txt`（1.2 基线）、`rfc3748-eap.txt`。

子任务：
- [ ] M10.1 TLS 1.3 握手协商与版本回退（兼容 1.2 客户端）
- [ ] M10.2 按 RFC 9190 / RFC 9427 实现 TLS 1.3 密钥派生（MSK/EMSK）
- [ ] M10.3 `close_notify` / 身份保护等 TLS 1.3 语义差异处理
- [ ] M10.4 单元测试 + `test/integration/` TLS 1.3 端到端验收用例（CI 自动执行）

验收口径：TLS 1.3 与 1.2 客户端均可完成 EAP-TLS 认证，密钥派生符合 RFC 9190；**验收由 `test/integration/` 的 CI 用例背书**。

## M11 — TEAP 隧道认证（中长期）

- **关联编号**：`TR-F004`
- **目标**：按 RFC 7170 / RFC 9930（TEAPv1）实现现代隧道 EAP，支持 machine + user chaining、证书 + 密码组合认证，作为 PEAP / FAST 的长期替代。
- **定位**：中长期路线。客户端生态弱于 PEAP，不适合第一版当主菜；仅在客户端环境完全可控时优先。
- **开发边界**：不与 PEAP / TTLS 抢第一版资源；复用既有隧道与状态机抽象；TLS 1.3 下采用 RFC 9427 派生规则。
- **技能**：`.agents/skills/add-eap-method/SKILL.md`、`.agents/skills/reference-rfc/SKILL.md`、`.agents/skills/add-acceptance-test/SKILL.md`、`.agents/skills/write-go-tests/SKILL.md`
- **协议规范**：`docs/rfcs/rfc7170-teap.txt`（TEAP）、`rfc9930`（TEAPv1，本地缺失，按 `reference-rfc` 登记）、`rfc9427`（TLS 1.3 派生，本地缺失）、`rfc3748-eap.txt`。

子任务：
- [ ] M11.1 TEAP 外层隧道与 TLV 框架（Crypto-Binding、Result TLV）
- [ ] M11.2 隧道内 EAP 方法链（machine + user chaining）
- [ ] M11.3 证书 + 密码组合认证与 Crypto-Binding 校验
- [ ] M11.4 单元测试 + `test/integration/` TEAP 端到端验收用例（CI 自动执行）

验收口径：TEAP 客户端可完成至少一种 chaining 组合认证，Crypto-Binding 校验正确；**验收由 `test/integration/` 的 CI 用例背书**。

## M12 — EAP-PWD 口令认证（按需）

- **关联编号**：`TR-F004`
- **目标**：按 RFC 5931 以共享口令完成认证，不为每客户端签发证书，适合 IoT、嵌入式、受控小规模设备。
- **定位**：按需推进，非通用企业 Wi-Fi 首选；不为“协议完整性”把自己拖进维护沼泽。
- **开发边界**：仅在有明确 IoT / 受控设备需求时排期；复用现有 EAP 注册与状态管理；口令交换的抗字典 / 主动 / 被动攻击特性必须有测试覆盖。
- **技能**：`.agents/skills/add-eap-method/SKILL.md`、`.agents/skills/reference-rfc/SKILL.md`、`.agents/skills/write-go-tests/SKILL.md`、`.agents/skills/add-acceptance-test/SKILL.md`
- **协议规范**：`rfc5931`（EAP-PWD，本地缺失，按 `reference-rfc` 登记；注意 `docs/rfcs/rfc7542-eap-pwd.txt` 命名有误，RFC 7542 实为 NAI，应补录真正的 RFC 5931）、`rfc3748-eap.txt`。

子任务：
- [ ] M12.1 注册 EAP-PWD handler 骨架与启用列表配置
- [ ] M12.2 PWD 口令元素与 PWE 推导（按 RFC 5931，含群组协商）
- [ ] M12.3 Commit / Confirm 交换与密钥导出
- [ ] M12.4 单元测试（含抗字典攻击向量）+ `test/integration/` 验收用例（CI 自动执行）

验收口径：EAP-PWD 客户端可用共享口令完成认证，交换符合 RFC 5931；**验收由 `test/integration/` 的 CI 用例背书**。

## M13 — 双语文档站点（mdbook）

- **关联编号**：`TR-F023`
- **目标**：用 mdbook 搭建中英文双语文档站点，收编散落文档（README / AGENT / SECURITY / 功能清单 / 路线图 / RFC 索引），提供统一导航与可发布产物。
- **开发边界**：先骨架与导航，再分批迁移；中英文目录结构对应、同步维护；文档站点不替代以代码与测试为准的口径（遵守 `TR-N003`，不扩展为产品门户）。**注意**：仓库当前已通过 GitBook 集成发布（`docs.toughradius.net` / `www.toughradius.net`），M13 必须先明确 mdbook 与 GitBook 是替代还是并存，避免两套发布管线产生冲突或内容漂移。
- **技能**：文档工程为主；复用 `.agents/skills/reference-rfc/SKILL.md` 维护协议资料索引、`.agents/skills/align-feature-checklist/SKILL.md` 保持中英文同步。

子任务：
- [ ] M13.0 评估与现有 GitBook 发布的关系（替代 / 并存），确定单一事实来源与发布管线边界
- [ ] M13.1 mdbook 骨架：`book.toml` + `src/`（zh / en 双语目录）+ 本地构建（`mdbook build`）
- [ ] M13.2 迁移核心文档（README / AGENT / SECURITY）为双语章节，原文件保留指向站点的入口
- [ ] M13.3 收编功能清单 / 路线图 / RFC 索引为站点章节，建立中英文交叉链接
- [ ] M13.4 CI 增加 `mdbook build` 产物校验（构建失败或坏链即红）
- [ ] M13.5 （可选）GitHub Pages 部署工作流

验收口径：`mdbook build` 在 CI 通过且无坏链，中英文章节一一对应，核心散落文档可从站点统一访问；**验收由 CI 构建用例背书**。

---

## Agent 排期约定

- **入口（自动委托）**：收到"自动委托开发 / 继续推进路线图"类指令时，由 [`.agents/skills/orchestrate-roadmap/SKILL.md`](../.agents/skills/orchestrate-roadmap/SKILL.md) 作为总调度统筹一轮：选题 → 选执行 SOP → 派工 → 质量门禁 → PR → 迭代路线图。
- 调度优先级：先 P1（`M2 → M3 → M8 → M9`），再 P2（`M4 / M5 / M7 / M10 / M13`），最后 P3（`M6 / M11 / M12`）；同优先级里程碑按序号取，P2/P3 仅在更高优先级里程碑无可执行子任务时填充。EAP 套件优先续接 M1（EAP-TLS）：先 PEAP-MSCHAPv2（兼容）、再 EAP-TTLS（后端适配），TLS 1.3 / TEAP / EAP-PWD 列为中长期 / 按需。
- 单次 agent 任务只认领一个未勾选子任务（最小闭环），完成后在本文件勾选并在 PR 引用里程碑编号。
- **自我迭代**：每轮交付后由 [`.agents/skills/groom-roadmap/SKILL.md`](../.agents/skills/groom-roadmap/SKILL.md) 勾选已交付项、更新里程碑状态、回填/拆分/重排子任务，并保持本文件与功能清单状态一致。
- 任何超出 `TR-F` 清单的需求，必须先提交清单更新 PR，再排入本路线图。
- 每个涉及协议或数据流的子任务，交付时必须带 **CI 可自动执行的验收测试**（单元或 `test/integration/`）。
- 选任务口径：优先取**优先级最高里程碑**中自上而下第一个未勾选的 `- [ ] M*.*`（优先级见里程碑总览的优先级列与上面的调度优先级排期）；同优先级里程碑按序号顺序。agent 在**本机**用你自己的 agent 运行，不在 CI 执行；运行参考与护栏见 [`.agents/README.md`](../.agents/README.md)。

