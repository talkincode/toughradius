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
| M3 | IPv6 能力增强闭环 | TR-F007 / TR-F011 / TR-F015 | P1 | 已完成 |
| M4 | Agent 开发体系与质量门禁 | TR-F022 | P2 | 进行中 |
| M5 | 厂商 VSA 覆盖扩展 | TR-F005 | P2 | 计划中 |
| M6 | 可观测性与运维增强 | TR-F015 | P3 | 计划中 |
| M7 | 上游 RADIUS 库跟踪与协议合规 | TR-F021 / TR-F022 | P2 | 进行中 |
| M8 | PEAPv0 / EAP-MSCHAPv2 认证支持 | TR-F004 | P1 | 已完成 |
| M9 | EAP-TTLS 隧道认证支持 | TR-F004 | P1 | 已完成 |
| M10 | EAP-TLS 1.3 / RFC 9190 升级 | TR-F004 | P2 | 计划中 |
| M11 | TEAP 隧道认证（中长期） | TR-F004 | P3 | 计划中 |
| M12 | EAP-PWD 口令认证（按需） | TR-F004 | P3 | 计划中 |
| M13 | 双语文档站点（mdbook） | TR-F023 | P2（优先） | 进行中 |
| M14 | 认证后端扩展（LDAP / AD bind，PAP 族） | TR-F025 | P2 | 计划中 |

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
- [x] M3.5 端到端测试与字段一致性校验（`test/integration/ipv6_test.go`，CI `integration` job 自动执行）：真实 PostgreSQL + 在线 RADIUS auth/acct 服务驱动 IPv6 全链路——为用户配置静态 IPv6 主机地址、SLAAC 前缀池、静态 Delegated-IPv6-Prefix 与 DHCPv6-PD 委派池后，PAP 认证断言 Access-Accept 同时携带 Framed-IPv6-Prefix(RFC 3162)/Framed-IPv6-Address(RFC 6911 §2.1)/Framed-IPv6-Pool(RFC 3162)/Delegated-IPv6-Prefix(RFC 4818 #123)/Delegated-IPv6-Prefix-Pool(RFC 6911 #171)，再以 NAS 回显的属性发起 Accounting-Start 并断言 RadiusOnline 与 RadiusAccounting 落库的 IPv6 字段与下发值逐一一致（auth→acct→DB 字段一致性）；另含动态 link 模式下从 Profile 继承 Framed/Delegated IPv6 池的端到端用例。**该用例发现并修复了真实缺陷**：`ApplyAcceptEnhancers` 构造的 `AuthContext` 缺失 `Metadata`，导致全部 Accept 增强器拿到 nil `profile_cache`/`config_mgr`，动态 link 模式下 Profile 继承的地址/IPv6 池、速率、domain 等属性被静默丢弃、且计费 interim 间隔退化为硬编码缺省值——现已注入 `profile_cache` 与 `config_mgr`，并补充 `TestApplyAcceptEnhancersWiresProfileCacheMetadata` 单元回归
- [x] M3.6 下发 Delegated-IPv6-Prefix / Delegated-IPv6-Prefix-Pool（先于 M3.5 实现：M3.5 的端到端一致性校验需覆盖下发环节，故按依赖关系前置）：default Access-Accept enhancer 从 `RadiusUser.DelegatedIpv6Prefix` 下发 Delegated-IPv6-Prefix（RFC 4818 #123，裸地址归一为 /128，IPv4/非法值跳过避免畸形 4 字节前缀），并经 `GetDelegatedIPv6PrefixPool` 下发 Delegated-IPv6-Prefix-Pool（RFC 6911 #171，支持 Profile 继承，按 §2.4 与 Framed-IPv6-Pool 区分）；含单元测试覆盖网络前缀/裸地址归一/继承/跳过

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
- [x] M4.7 为 agent 任务建立 PR 模板与 review checklist
- [x] M4.8 收敛 agent 产出质量度量（CI 通过率、回滚率）<br/>已交付：新增 `scripts/agent-roadmap-quality-metrics.sh`，按时间窗口统计已合并 `agent-roadmap` PR 的 CI 通过率（关联 CI 工作流 completed runs 的 `success/total`、失败 run 数、`attempt>1` 重跑数）与回滚率（合并提交是否被 `main` 上 `This reverts commit <sha>` 回滚），支持输出 JSON/Markdown 报告；`.agents/README.md` 增补统一口径与执行示例（`--days/--json-output/--markdown-output`），用于每轮自动委托后的质量回看与回归优先级决策。
- [ ] M4.12 按模块增量补齐导出标识符 godoc 注释与包注释（顺序建议：`internal/adminapi` → `internal/radiusd` → `pkg`），并探索 lint 度量（`TR-F024`）<br/>进行中（增量交付）：批次 1 已补齐 `internal/adminapi` 框架层——新增包注释（`adminapi.go` 顶部，标准库风格，说明 /api/v1 管理 API 定位、handler 无状态从 echo context 取依赖、统一 `Response`/`ErrorResponse` 信封与 `RequireLevel` 鉴权模型）、`Init`/`GetAppContext`/`GetDB`/`GetConfig` 契约注释（读取来源、中间件契约、缺失即 panic 的程序性错误语义、按 key 回落），及 `Meta`/`Response`/`ErrorResponse` 信封类型契约；并以 `nas.go`（`ListNAS`/`GetNAS`/`CreateNAS`/`UpdateNAS`/`DeleteNAS`）为资源处理器范本：首段散文化契约（HTTP 方法/路径、鉴权级别、参数与默认值/钳制、错误码与响应体形状、SQL 注入护栏），保留既有 Swagger `@` 注解并以空注释行与散文分段（`go doc` 首段即干净摘要）。门禁：`gofmt`、`go build ./...`、`go vet`、`go test ./internal/adminapi/...`、golangci-lint v2.12.2 对该包 0 问题。余 `internal/adminapi` 其余资源处理器（users/profiles/sessions/accounting/dashboard/operators/nodes/settings/system_backup/system_logs/session_actions/auth/users_import）及 `internal/radiusd`、`pkg` 待后续批次续接；lint 度量（按包 ratchet 开启 ST1000/ST1020/ST1021 或 revive `exported`）待该包补齐后再探索。<br/>批次 2：已补齐 `internal/adminapi/profiles.go` 资源——为 5 个导出 handler（`ListProfiles`/`GetProfile`/`CreateProfile`/`UpdateProfile`/`DeleteProfile`）与 2 个导出请求 DTO（`ProfileRequest`/`ProfileUpdateRequest`）补齐标准库风格契约注释（HTTP 方法/路径、鉴权级别——读端开放/写端 `requireAdmin`、参数默认值与钳制、错误码与 `Response`/`Meta` 信封形状、`ORDER BY` 白名单防注入、缓存失效语义），保留 Swagger `@` 注解并以空注释行分段（`go doc` 首段即干净摘要）；顺带修正 `toRadiusProfile`/`registerProfileRoutes` 的破碎英文注释。门禁：`gofmt`、`go build`、`go vet`、`go test ./internal/adminapi/...`、golangci-lint v2.12.2 对该包 0 问题。余 `internal/adminapi` 其余资源（users/users_import/sessions/session_actions/accounting/dashboard/operators/nodes/settings/system_backup/system_logs）及 `internal/radiusd`、`pkg` 续接。<br/>批次 3：已补齐 `internal/adminapi/sessions.go` 在线会话资源——为 3 个导出 handler（`ListOnlineSessions`/`GetOnlineSession`/`DeleteOnlineSession`）与 `registerSessionRoutes` 补齐标准库风格契约注释：列表的分页/钳制、过滤字段（精确 vs 转义 LIKE 子串）、多格式时间窗、`ORDER BY` 白名单防注入与分页信封；`DeleteOnlineSession` 强制下线语义（删除在线记录 + 异步尽力而为发送 RFC 5176 Disconnect-Request、5s 超时、ACK/NAK 仅记日志不影响响应、NAS 缺失则仅删库告警）与错误码（`INVALID_ID`/`NOT_FOUND`/`DELETE_FAILED`）；并据实记录鉴权级别（读 + 强制下线对任意已认证操作员开放，而 `POST /sessions/:id/disconnect`、`/coa` 由 `requireAdmin` 守护）。门禁：`gofmt`、`go build`、`go vet`、`go test ./internal/adminapi/...`、golangci-lint v2.12.2 对该包 0 问题。<br/>批次 4：已补齐 `internal/adminapi/dashboard.go` ——为 `GetDashboardStats` 写明聚合契约（一次请求返回实时计数 + 今日量值 + 7 天认证趋势 / 24 小时流量 / 在线用户按 profile 分布 + IPv6 采用率三类时序；今日以本地零点起算、趋势/流量按桶零填充、流量按 1024³ 折算 GB、`TodayAuthCount` 为按当日新会话估算而非独立认证计数器、单段查询失败仅记日志回落零值、对任意已认证操作员开放），并将 `DashboardStats`/`DashboardAuthTrendPoint`/`DashboardTrafficPoint`/`DashboardProfileSlice` 类型头注释升级为标准库风格契约、补齐 `DashboardProfileSlice` 字段注释（`DashboardIPv6Stats` 此前已具完整 godoc，未改）。门禁：`gofmt`、`go build`、`go vet`、`go test ./internal/adminapi/...`、golangci-lint v2.12.2 对该包 0 问题。
  <br/>批次 5：已补齐 `internal/adminapi/accounting.go` 资源——为 2 个导出 handler（`ListAccounting`/`GetAccounting`）补齐标准库风格契约注释：列表接口明确分页默认值与钳制、`sort/order` 白名单防注入、用户名/NAS/会话 ID/IPv4/IPv6/MAC/委派前缀模糊过滤语义，以及 `acct_start_time_gte/lte` 支持 RFC3339、datetime-local、date-only 三种格式且解析失败按现状忽略；详情接口明确 `INVALID_ID`(400)/`NOT_FOUND`(404) 错误码与鉴权边界（任意已认证操作员可读）。同步整理 `parseFlexibleTime`/`escapeLikePattern`/`registerAccountingRoutes` 注释以提升 `go doc` 可读性。门禁：`gofmt`、`go build ./...`、`go test ./...`、golangci-lint v2.12.2 通过，且 PR #373 已合并。<br/>批次 6：已补齐 `internal/adminapi/users.go` 与 `internal/adminapi/users_import.go` 的导出类型注释——`UserRequest`/`UserUpdateRequest` 明确旧版字段与 mixed-type 兼容契约，`ImportUserError`/`ImportUserResult` 明确 1-based 行号与失败行反馈语义。门禁：`go build ./...`、`go test ./...`、golangci-lint v2.12.2 通过，且 PR #383 已合并。<br/>批次 7：完成 lint 度量探索并落地增量 ratchet——在 `.golangci.yml` 启用 staticcheck 导出注释检查（保持 `ST1000` 包注释检查暂缓），并修正 `internal/radiusd`、`pkg`、`web` 等模块的导出标识符注释以通过门禁；对应 PR #385 已合并。<br/>批次 8：补齐 `internal/radiusd` 与 `pkg/web` 包级 `doc.go` 注释（标准库风格），并用 `staticcheck -checks ST1000 ./internal/radiusd ./pkg/web` 做定向度量验证；全量门禁 `go build ./...`、`go test ./...`、golangci-lint v2.12.2 通过；对应 PR #388 已合并。<br/>批次 9：补齐 `internal/adminapi/system_backup.go` 导出类型 `SystemBackup`/`SystemRestoreResult` 的标准库风格契约注释，明确备份 schema 版本语义、敏感数据处理边界与恢复计数零值语义，并补齐字段级注释以提升 `go doc` 可读性；全量门禁 `go build ./...`、`go test ./...`、golangci-lint v2.12.2 通过；对应 PR #390 已合并。<br/>
  批次 10：补齐 `internal/radiusd/cache`、`pkg/excel`、`pkg/timeutil`、`pkg/validator`、`pkg/validutil` 的包级 `doc.go` 标准库风格注释，并以 `staticcheck -checks ST1000` 对上述包做定向度量验证；全量门禁 `go build ./...`、`go test ./...`、`golangci-lint v2.12.2` 通过；对应 PR #392 已合并。<br/>
  批次 11：补齐 `internal/radiusd/plugins/accounting/handlers`、`internal/radiusd/plugins/auth/{checkers,enhancers,guards,validators}` 的包级 `doc.go` 标准库风格注释，并以 `go run honnef.co/go/tools/cmd/staticcheck@latest -checks ST1000` 对上述包做定向度量验证；全量门禁 `go build ./...`、`go test ./...`、`go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run` 通过；对应 PR #394 已合并。<br/>
  批次 12：补齐 `internal/radiusd/plugins/eap`、`internal/radiusd/plugins/eap/handlers`、`internal/radiusd/plugins/eap/statemanager`、`internal/radiusd/plugins/vendorparsers/parsers` 的包级 `doc.go` 标准库风格注释，并以 `go run honnef.co/go/tools/cmd/staticcheck@latest -checks ST1000 ./internal/radiusd/plugins/eap/... ./internal/radiusd/plugins/vendorparsers/parsers/...` 做定向度量验证；全量门禁 `go build ./...`、`go test ./...`、`go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run` 通过；对应 PR #396 已合并。<br/>
  批次 13：补齐 `internal/radiusd/vendors/*` 全部 15 个生成的厂商 VSA 字典包（alcatel/aruba/cisco/f5/h3c/hillstone/huawei/ikuai/juniper/microsoft/mikrotik/pfSense/radback/unix/zte）的包级 `doc.go` 标准库风格注释——每个注释以独立、非生成的 `doc.go` 承载（避免 `radius-dict-gen` 重新生成 `generated.go` 时被覆盖），以 RFC 2865 §5.26 的精确措辞「SMI Network Management Private Enterprise Code」+ 数字厂商码（与各包 `generated.go` 内 `_<Vendor>_VendorID` 常量逐一核对一致）标识包身份而不臆断厂商法人名称，并描述生成的 `_Add/_Get/_Gets/_Lookup/_Set/_Del` 访问器在 `radius.Packet` 上编解码该厂商 VSA 的契约；至此 `internal/radiusd/vendors` 子树的 ST1000（包注释）缺口已全部补齐。以 `staticcheck -checks ST1000 ./internal/radiusd/vendors/...` 做定向度量验证（clean）；全量门禁 `gofmt -l`、`go build ./...`、`go vet ./...`、`go test ./...`、golangci-lint v2.12.2（0 问题）通过；对应 PR #398 已合并。全局 ST1000 仍暂缓，后续按批次继续 ratchet。
  <br/>批次 14：补齐 `internal/radiusd/plugins` 包级 `doc.go` 标准库风格注释，并以 `go run honnef.co/go/tools/cmd/staticcheck@latest -checks ST1000 ./internal/radiusd/plugins` 做定向度量验证；全量门禁 `go build ./...`、`go test ./...`、`go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run` 通过；对应 PR #400 已合并。
  <br/>批次 15：补齐 `internal/app`、`internal/domain`、`internal/webserver` 包级 `doc.go` 标准库风格注释，并以 `go run honnef.co/go/tools/cmd/staticcheck@latest -checks ST1000 ./internal/app/... ./internal/domain/... ./internal/webserver/...` 做定向度量验证；全量门禁 `go build ./...`、`go test ./...`、`go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run` 通过（使用 fresh `GOLANGCI_LINT_CACHE` 规避历史失效路径缓存）；对应 PR #403 已合并。
  <br/>批次 16：补齐 `internal/radiusd/plugins/accounting/handlers` 导出 API（`StartHandler`/`StopHandler`/`UpdateHandler`/`NasStateHandler`）的标准库风格契约注释，明确 `Accounting-Start` 幂等与回滚语义、`Stop` 终止更新与在线会话清理行为、`Interim-Update` 计数刷新语义，以及 NAS `Accounting-On/Off` 批量清理边界（comments-only，无运行逻辑变更）；全量门禁 `go build ./...`、`go test ./...`、`go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run` 通过（使用 fresh `GOLANGCI_LINT_CACHE` 规避历史失效路径缓存）；对应 PR #405 已合并。
  <br/>批次 17：继续补齐 `internal/radiusd/plugins/accounting/handlers` 契约注释细节——强化 `StopHandler`/`UpdateHandler` 的返回语义与边界说明，补充 `buildRadiusOnline`/`buildRadiusAccounting`/`buildOnlineFromRequest` 的映射契约注释，并清理 `start_handler.go` 中导入段尾部失真注释（comments-only，无运行逻辑变更）；全量门禁 `go build ./...`、`go test ./...`、`go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run` 通过（使用 fresh `GOLANGCI_LINT_CACHE` 规避历史失效路径缓存）；对应 PR #410 已合并。
- [x] M4.13 外部/Fork PR 投毒防护门禁（区分内部 agent PR 与外部贡献者 PR 的信任域，`TR-F022`）<br/>动机：自动委托循环对自身 `copilot/*` 分支产出无人值守地评审合并；但来自 fork 的外部 PR 属于不可信代码（可夹带恶意逻辑或 CI 变更），必须由维护者本人亲自审批，绝不能进入自动合并通道。交付内容：<br/>① 新增 `.github/workflows/external-pr-gate.yml`——以 `pull_request_target`（不检出 PR 代码，规避 pull_request_target 提权投毒）+ `pull_request_review` + `merge_group` 触发，`permissions` 仅 `contents:read`/`pull-requests:read`；job 名即必需检查上下文 `external-pr-gate`。判定逻辑：`merge_group` 直接放行；head 与 base 同仓（内部分支）直接放行；fork PR 仅当存在来自具 write/admin 权限维护者、且 `commit_id` 等于当前 `pr.head.sha` 的 APPROVED 评审时放行，否则 `core.setFailed` 阻断；head.repo 为空（fork 被删）按 fork 处理（fail-safe 阻断）。审批与 head 提交绑定：`synchronize` 会在 fork 每次推送后重跑门禁，故旧提交上的过期 approval 不得放行新推送的未审代码（`commit_id` 为空亦按未覆盖处理）。<br/>② 新增 `.github/CODEOWNERS`——`@jamiesun` 拥有 `/internal/radiusd/`、`/internal/adminapi/`、`/internal/domain/`、`/go.mod`、`/go.sum`、`/.github/`、`/.agents/` 等敏感路径，提供自动 review 请求与就绪态；`require_code_owner_review` 刻意保持关闭（agent 以 jamiesun 身份运行，无法批准自己的 PR，开启会死锁内部 PR），fork 门禁才是「外部 PR 需我亲批」的精准机制。<br/>③ SOP 护栏：`review-pr`、`orchestrate-roadmap` 技能与 `.github/review-checklists/agent-roadmap.md` 增补「外部/Fork PR 属不同信任域」条目——agent 绝不自动评审/合并 cross-repository PR（`gh pr view --json isCrossRepository` 判定），一律 `needs-human` 留给维护者；自动域仅限本仓 `copilot/*` 内部 PR。<br/>④ Actions 设置加固：`default_workflow_permissions=read`、`can_approve_pull_request_reviews=false`（5 个工作流均自声明 `permissions`，收紧默认值安全）；并配合仓库已启用的 `sha_pinning_required` 策略，将全部工作流（`ci.yml`/`docker-publish.yml`/`pages.yml`/`release-publish.yml`/`external-pr-gate.yml`）中 40 处第三方/官方 action 由可变 tag 全部钉死为 40 位 commit SHA（保留 `# vX` 注释），既修复策略开启后 CI 在 *Set up job* 阶段整体失败，也消除供应链投毒面。⑤ 将 `external-pr-gate` 接入主 ruleset 作为必需状态检查（内部 PR 自动通过、fork PR 阻断至维护者批准），并经合并队列回归验证不影响内部 PR 合并。门禁：actionlint 对 `external-pr-gate.yml` exit 0；gate JS 的 node stub 单测 12/12（内部放行 / fork 无评审阻断 / fork 维护者审批当前 head 放行 / fork 维护者审批旧提交即过期 approval 阻断 / approval `commit_id` 为空 fail-safe 阻断 / approve A 后推送 B 阻断 / fork 非维护者阻断 / 仅 COMMENTED 阻断 / fork write 权限审批放行 / fork 被删 fail-safe 阻断 / merge_group 放行 / APPROVED 后 CHANGES_REQUESTED 阻断）。<br/>交付记录：初始门禁随 PR #407 合入 `main`；随后独立安全评审发现两处阻断项——(a) 仓库 `sha_pinning_required` 已开启致全部工作流（含本门禁）action 未钉死而在 *Set up job* 失败、CI 全仓阻塞且门禁不可运行；(b) 审批未与 head 提交绑定的过期审批旁路——遂以跟进 PR 修复（全量 SHA 钉死 + `commit_id===head.sha` 绑定）。跟进修复 PR #409 已合入 `main`：全部 40 处 action 已钉死为 40 位 SHA（仓库已无可变 `@vN` 引用）、fork 审批与 `commit_id===head.sha` 绑定生效，gate node stub 单测 12/12、actionlint exit 0、真实 CI 全绿。收尾：`external-pr-gate` 已接入主分支 ruleset（id 17085844）作为必需状态检查（`required_status_checks`，strict=false）——fork PR 未获维护者绑定 head 的 APPROVED 即被阻断、不得进入合并队列，而内部 `copilot/*` PR 经 `pull_request_target`（head==base 放行）与 `merge_group`（短路放行）双双通过，自动委托合并链路不受影响；Actions 设置已加固（`default_workflow_permissions=read`、`can_approve_pull_request_reviews=false`）。本路线图 groom PR 本身即经合并队列回归验证：在 `external-pr-gate` 必需检查下，内部 PR 仍自动通过队列合并。

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
- [ ] M6.4 修复休眠的数据清理任务：`Application.SchedClearExpireData`（清理 `radius_online` 残留行与按 `AccountingHistoryDays` 清理 `radius_accounting`）已定义但**从未注册到 cron**——当前 `@daily` 任务仅删除一年前的操作日志（`SysOprLog`），计费历史与残留在线记录实际不会自动清理。需将其按 `@daily`（或可配置周期）注册并补单元测试；发现来源：M13.9 FAQ 校准。

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
- [x] M8.1 注册 PEAP handler 骨架与启用列表配置（`EapMethod`）<br/>已交付：新增 `eap.TypePEAP=25` 常量与 `ErrPEAPNotImplemented` 安全护栏；`peap_handler.go` 骨架（`Name/EAPType/CanHandle`，`HandleIdentity` 下发 PEAPv0 Start（S 位 + version 0，复用 EAP-TLS Flags/分片帧格式），`HandleResponse` 在外层隧道（M8.2）落地前一律以 `ErrPEAPNotImplemented` 拒绝，永不放行）；coordinator 选择分支与 `plugins/init.go` 注册接入；`config_schemas.json` 与硬编码兜底 `EapMethod` 枚举加入 `eap-peap`。单测覆盖 handler、coordinator 方法选择与枚举。引用 RFC 3748/3579、RFC 5216 §3.1 帧格式与 PEAPv0 [MS-PEAP]/draft-kamath-pppext-peapv0。
- [x] M8.2 PEAP 外层 TLS 隧道建立与分片重组（复用 EAP-TLS 状态机）<br/>已交付：`tlsengine` 新增 `Config.ServerOnly`（仅服务器证书认证，不要求 `ClientCAs`，`ClientAuth=NoClientCert`；EAP-TLS 路径 `ServerOnly=false` 行为字节级不变）；抽取共享隧道驱动 `tls_tunnel.go`（`tlsTunnel` 三阶段状态机，参数化 `eapType`/`configProvider`/`onHandshakeComplete`，EAP-TLS 与 PEAP 各 `newTunnel()` 复用，握手态仍全部经 StateManager 持久化）；`tls_handler.go` 行为保持式委派到隧道（`eapType=TypeTLS`，回调 `finalizeWithEngine` 凭证身份放行）；`peap_handler.go` 新增 `NewPEAPHandlerWithConfig`/`newTunnel()`，`onHandshakeComplete` 关闭引擎并 `Success=false` 返回 `ErrPEAPInnerNotImplemented`——**隧道建成但 M8.2 永不放行**（内层 EAP-MSCHAPv2 留待 M8.3）；`tls_config.go` 新增 `NewSettingsPEAPConfigProvider`（复用 `EapTlsCertFile/KeyFile`，`ServerOnly:true`，未配置返回 nil 安全拒绝）；`init.go` 在 ConfigMgr 可用时注入 PEAP 配置 provider。测试：server-only 引擎握手 + `Identity()` 返回 `ErrNoPeerCertificate`；PEAP 完整握手（单帧 + `maxFragment=64` 分片）均断言隧道建成后以 `ErrPEAPInnerNotImplemented` 拒绝且 `state.Success=false`；未配置证书时以 `ErrTLSNotConfigured` 拒绝；既有 EAP-TLS 握手测试不变通过。引用 RFC 5216 §2.1.5/§3.1（分片与帧格式）、PEAPv0 [MS-PEAP]。
- [x] M8.3a `tlsengine` 隧道建立后应用数据收发（`WriteApplication`/`ReadApplication`）与 TLS 密钥材料导出（RFC 5705 `ExportKeyingMaterial`，供 MPPE/MSK 派生）；握手完成后保持引擎存活<br/>已交付（PR #352）：`Engine` 新增 `WriteApplication`（内层 EAP 请求加密为 TLS 记录）/`ReadApplication`（喂入重组入站 flight 返回解密内层响应，解密读取在有界超时下进行，超时关闭 transport 返回 `ErrAppReadTimeout`，截断/恶意 flight 不卡死 worker）/`ExportKey`（RFC 5705 `ExportKeyingMaterial`，label `"client EAP encryption"`、64 字节导出 MSK→MS-MPPE-Recv/Send-Key）/`HandshakeComplete`（就绪门控，三方法握手完成前一律 `ErrHandshakeNotComplete`）；握手成功后停止握手超时定时器避免慢速内层交换被误杀。测试覆盖 TLS1.2/1.3 双向应用数据往返、server/client 导出密钥字节级一致、预握手拒绝、超时路径，`-race` 干净。引用 RFC 5705、RFC 5216 §2.3、RFC 2548、RFC 7627。
- [x] M8.3b 隧道内 EAP-MSCHAPv2 交换（复用 mschapv2 校验）与 MPPE 密钥导出（基于 M8.3a 能力，隧道内跑内层 EAP，认证成功后从 TLS 导出 MSK 下发 MS-MPPE-Send/Recv-Key）<br/>已交付：`tls_tunnel.go` 在共享隧道驱动上叠加内层应用数据阶段——新增 `onApplicationData` 回调字段与 `startInner`/`handleInnerRound`/`driveInner`/`startAppFlight`/`sendNextAppFragment`，外层握手完成后 PEAP 的 `onHandshakeComplete` 改为调用 `startInner`（置 `stateKeyInnerActive`、驱动开场内层请求）而非拒绝；内层每轮以一条外层 EAP-Response 承载 TLS 应用数据，复用既有重组器/分片器（`maxFragment` 可强制内层分片），引擎跨轮存活于 `state.Data[stateKeyEngine]`，**仅 `driveInner` 在内层回调 `success==true` 时放行**，任何错误路径均 `closeEngine` 并拒绝；EAP-TLS 路径 `onApplicationData=nil` 且永不置 `stateKeyInnerActive`，行为字节级不变。`peap_inner.go` 新增内层 EAP-MSCHAPv2 状态机（`handleInnerEAP` 按 `stateKeyInnerPhase` 分发 identity→challenge→success-ack 三阶段）：复用 `MSCHAPv2Handler.buildChallengeRequest`/`parseResponse` 与 rfc2759 `GenerateNTResponse`/`GenerateAuthenticatorResponse`，NT-Response 校验失败返回 `ErrPasswordMismatch`（拒绝）；**MPPE 密钥来自 TLS 导出器**（`engine.ExportKey("client EAP encryption", nil, 64)`，RFC 5705/5216 §2.3）而非内层 MSCHAPv2 rfc3079 密钥，MSK[0:32]→MS-MPPE-Recv-Key、MSK[32:64]→MS-MPPE-Send-Key 并附加加密策略（RFC 2548）写入外层 Access-Accept。`tls_handler.go` 新增内层状态键与 `getString/setString/getUint8/setUint8` 类型化 state 助手；`errors.go` 新增 `ErrPEAPInnerProtocol`。测试：`peapPeer` 端到端 supplicant 跑通完整内层认证（单帧 + `maxFragment=48` 分片）并断言 Access-Accept 携带 MS-MPPE-Recv/Send-Key；错误口令断言以 `ErrPasswordMismatch` 拒绝；既有 EAP-TLS 握手测试不变通过，`-race` 干净。已知局限（留待 M8.4）：匿名外层身份→内层身份的用户记录映射尚未实现，当前 password 查找仍走外层 `ctx.User`（测试用非匿名身份，外层=内层完全正确）。引用 RFC 2759（MS-CHAP-V2）、RFC 2548（MS-MPPE VSA）、RFC 5705（TLS 导出器）、RFC 5216 §2.3、PEAPv0 [MS-PEAP]。
- [x] M8.4 明确失败原因 + AuthError 指标 + 单元测试；配置项与默认值（含安全风险说明）<br/>已交付：`auth_stages.go` 的 `mapEAPDispatchError` 新增 PEAP 专属拒绝原因映射——`ErrPEAPInnerProtocol`→`MetricsRadiusRejectOther`「peap inner eap-mschapv2 protocol violation」、`ErrPEAPInnerNotImplemented`→`MetricsRadiusRejectOther`「peap inner eap method unavailable」（均经 `errors.Is` 支持包装链），内层口令失败继续走 `ErrPasswordMismatch`→`MetricsRadiusRejectPasswdError`，告别此前一律落入 default「eap authentication failed」的笼统原因；`auth_stages_test.go` 表驱动用例补充三条 PEAP 断言（协议违规/未实现/口令失败的 metricsType + stage + reason）。配置与安全风险说明：`config_schemas.json` 与 `config_manager.go` 兜底 schema 的 `radius.EapMethod` 描述明示 eap-peap 复用 EAP-TLS 服务器证书/私钥建立外层隧道、内层 EAP-MSCHAPv2 存在类似 NTLMv1 的攻击面（按 Microsoft），须保持外层 TLS 强度、客户端支持证书时优先 eap-tls；`web/src/i18n/{en-US,zh-CN}.ts` 同步双语 UI 描述；`tls_config.go` `NewSettingsPEAPConfigProvider` 补充安全说明 doc（ServerOnly + 运维选定最小 TLS 版本，永不削弱）。`go test ./...` 与 `golangci-lint v2.12.2` 全绿，`cd web && npm run build` 通过。引用 RFC 3748/3579（EAP/RADIUS-EAP）、RFC 2759（MS-CHAP-V2 攻击面）、PEAPv0 [MS-PEAP]。
- [x] M8.5 在 `test/integration/` 增加 PEAP-MSCHAPv2 端到端验收用例（CI 自动执行）<br/>已交付：新增 `test/integration/peap_mschapv2_test.go`（`//go:build integration`，package `integration`，CI `integration` job 自动执行），实现真实「线上」PEAPv0 / EAP-MSCHAPv2 supplicant——经真实 UDP RADIUS 报文驱动运行中的认证服务，复用 `eap_tls_test.go` 的内存双工流（`eapStream`/`eapConn`）、测试 CA 助手与 `assertEAPCode`，并对齐单测 `peapPeer` 的内层状态机但跑在线上：`peapSupplicant` 完成外层 EAP-TLS 握手后**保持 `tls.Conn` 存活**，将内层 EAP-MSCHAPv2（Identity→Challenge→Success-ACK，NT-Response 由 rfc2759 `GenerateNTResponse` 计算）作为 TLS 应用数据承载于外层 EAP-PEAP 帧（复用 `tlsfragment` 分片/重组、State 续传、identifier 回显），客户端 pin TLS 1.2 保证外层帧确定性。`configurePEAP`（仅写服务器证书/私钥 + `EapMethod=eap-peap`，PEAP 为 ServerOnly 无需客户端 CA）与 `seedPEAPUser`（明文口令，provider 原样返回）经动态配置热切换 EAP 方法并以 `restoreEapMethod` 归位。两条子用例断言：① 正确口令→Access-Accept + EAP-Success + MS-MPPE-Send/Recv-Key（以**请求 Authenticator** 重绑后解密均为 32 字节会话密钥，RFC 2548 §2.4）；② 错误口令→Access-Reject + EAP-Failure 且不泄漏 MPPE 密钥（`MSMPPERecvKey_Lookup` 返回 `ErrNoAttribute`）。串行（无 `t.Parallel`，依赖进程级全局插件注册/动态配置/限速器）。`make test-integration-pg` 本地全绿（真实 PostgreSQL）；`go build ./...`、`go test ./...`、`go vet -tags=integration` 通过。引用 RFC 2759（MS-CHAP-V2）、RFC 2548（MS-MPPE VSA salt 加密）、RFC 5216 §2.1.5/§3.1（EAP-TLS 分片帧格式）、PEAPv0 [MS-PEAP]。

验收口径：PEAP-MSCHAPv2 客户端（Windows/AD 典型配置）可完成认证，MPPE 密钥正确下发；失败有明确拒绝原因与指标；**验收由 `test/integration/` 的 CI 用例背书**。

## M9 — EAP-TTLS 隧道认证支持

- **关联编号**：`TR-F004`
- **目标**：按 RFC 5281 用服务器证书建立 TLS 隧道，隧道内承载 PAP / CHAP / MS-CHAP / MS-CHAP-V2 内层认证，让 LDAP、老账号库、混合客户端无需立即改造证书体系即可接入。
- **定位**：后端适配优先。现实价值是“后端用户库不用立刻改成证书体系”——先用服务器证书保护隧道，再把用户名口令塞进去；很多企业认证体系是历史债务盘出来的，TTLS 是务实的过渡。
- **开发边界**：内层方法逐个交付（先 PAP，再 MS-CHAP-V2）；后端用户库适配走现有认证流水线，不在协议入口写库分支；不重写 EAP 协调器。
- **技能**：`.agents/skills/add-eap-method/SKILL.md`、`.agents/skills/add-acceptance-test/SKILL.md`、`.agents/skills/write-go-tests/SKILL.md`
- **协议规范**：`docs/rfcs/rfc5281-eap-ttls.txt`（EAP-TTLS）、`rfc3748-eap.txt`、`rfc2759-mschapv2.txt`（内层 MS-CHAP-V2）、`rfc3579-radius-eap-support.txt`；TLS 1.3 隧道按 `rfc9427`（本地缺失，按 `reference-rfc` 登记）。

子任务：
- [x] M9.1 注册 EAP-TTLS handler 骨架与启用列表配置 ✅ 实现 `TTLSHandler`（EAP type 21 / name `eap-ttls`），对 EAP-Response/Identity 回送 EAP-TTLSv0 Start（RFC 5281 §9.2，Flags 复用 EAP-TLS 框架 RFC 5216 §3.1，version 位为 0），并在协调器 `handleIdentityResponse` 增加 `eap-ttls` 分支、`init.go` 无条件注册该 handler；`HandleResponse` 为安全桩，对任何挑战响应一律返回 `ErrTTLSNotImplemented`、永不放行（外层隧道见 M9.2）。`EapMethod` 枚举在 `config_schemas.json` + `config_manager.go` + 前端 `i18n/{en-US,zh-CN}.ts` 三处同步加入 `eap-ttls` 并补充双语说明。单测覆盖 Name/EAPType(21)/CanHandle/HandleIdentity(Start 帧+状态持久化)/buildStartRequest/HandleResponse 永不认证，并在 `coordinator_test.go` 增加 `eap-ttls` 选路用例。门禁：`go build ./...`、`go test ./...`、golangci-lint v2.12.2、`cd web && npm run build` 均通过。
- [x] M9.2 外层 TLS 隧道建立与分片重组（复用 EAP-TLS 状态机）✅ `TTLSHandler` 增加 `NewTTLSHandlerWithConfig` 与 `configProvider`/`maxFragment`，`HandleResponse` 复用共享 `tlsTunnel` 状态机（`eapType=TypeTTLS`）驱动 server-only 外层 TLS 握手与 RFC 5216 §2.1.5/§3.1 分片重组；握手完成回调 `onHandshakeComplete` 在内层 AVP 认证（M9.3+）落地前返回 `eap.ErrTTLSInnerNotImplemented`，隧道可建立可分片但**永不放行**。新增 `NewSettingsTTLSConfigProvider`（server-only，复用 `EapTlsCertFile/KeyFile/MinVersion`）；`init.go` 按 `ConfigMgr()` 是否可用分支注册（缺省回落到未配置 handler，按 `ErrTLSNotConfigured` 安全拒绝）。测试：未配置→`ErrTLSNotConfigured`；真实 server-only TLS 握手经隧道跑通后→`ErrTTLSInnerNotImplemented`（含强制 `maxFragment=48` 的分片重组用例）。门禁 `go build/test ./...`、golangci-lint v2.12.2 全过。
- [x] M9.3 隧道内 AVP 封装与 PAP 内层认证（最小可用闭环）✅ 新增 `ttls_avp.go`：按 RFC 5281 §10.1/§10.2 解析隧道内 AVP 序列（Code 4B | Flags 1B[V/M] | Length 3B | [Vendor-ID 4B] | Data，4 字节边界零填充且 Length 不含填充），对截断头部 / Length 小于头部 / Length 越界一律 `eap.ErrTTLSInnerProtocol` 拒绝；`findTTLSAVP` 仅匹配非 Vendor AVP，`stripTTLSPasswordPadding` 去除 User-Password 尾部 NUL（RFC 5281 §11.2.5）。新增 `ttls_inner.go`：`handleInnerAVP` 实现内层 PAP——提取 User-Password（缺失即非 PAP 内层方法→`eap.ErrTTLSInnerNotImplemented`，留待 M9.4）、`subtle.ConstantTimeCompare` 比对 `PwdProvider` 口令（不符→`eap.ErrPasswordMismatch`），成功后用 RFC 5705 导出器以标签 `"ttls keying material"`（RFC 5281 §8）导出 64B MSK 并按 RFC 2548 拆分 MS-MPPE-Recv/Send-Key 加入 Access-Accept。共享 `tlsTunnel` 仅新增 `clientSpeaksFirst` 字段：EAP-TTLS phase 2 为 peer-initiated（RFC 5281 §7.3），在 pendingSuccess 分支对非 ACK 帧转入内层（不影响 EAP-TLS/PEAP 的 server-first 行为）；`tlsengine.Config` 新增 `MaxVersion`，`NewSettingsTTLSConfigProvider` 钉死 TLS 1.2（TLS 1.3/RFC 9427 隧道为后续里程碑）。测试：AVP 编解码/填充/畸形/Vendor 用例，及经真实 server-only TLS 握手的整链 PAP——成功（Access-Accept 含 MS-MPPE 双钥、内层 User-Name 入状态）、强制 `maxFragment=48` 分片、错误口令→`ErrPasswordMismatch`、无 User-Password→`ErrTTLSInnerNotImplemented`。门禁 `go build/test ./...`、golangci-lint v2.12.2 全过。
- [x] M9.4 隧道内 MS-CHAP-V2 认证与密钥导出 ✅ 新增 `ttls_mschapv2.go`：实现 RFC 5281 §11.2.4 的非 EAP（RADIUS VSA）形态 MS-CHAP-V2——客户端隧道内发送 User-Name + MS-CHAP-Challenge(311/11) + MS-CHAP2-Response(311/25) 三个 Microsoft Vendor AVP，服务端校验通过后隧道回送 MS-CHAP2-Success(311/26)，客户端以**空 EAP-TTLS 帧**确认后才放行。防 MITM：服务端与客户端均以 RFC 5281 §11.1 隐式挑战 `engine.ExportKey("ttls challenge", nil, 17)` 派生 17 字节（16B 认证挑战 + 1B Ident，TLS 1.2 + EMS/RFC 7627 下复用 M9.3 钉死的 1.2 隧道），`subtle.ConstantTimeCompare` 校验客户端 MS-CHAP-Challenge AVP 与派生挑战一致、MS-CHAP2-Response 的 Ident 与派生 Ident 一致（任一不符→`eap.ErrTTLSInnerProtocol`，客户端无法自选挑战），再以 RFC 2759 `GenerateNTResponse` 常量时间比对 NT-Response（不符→`eap.ErrPasswordMismatch`）；MS-CHAP2-Success 取 RFC 2759 `GenerateAuthenticatorResponse` 的 `S=` 串按 RFC 2548 §2.3.3 组帧。`ttls_inner.go` 的 `handleInnerAVP` 重构为按 `stateKeyInnerPhase` 分发的相位机（`handleInnerStart`→PAP/MS-CHAP-V2 选路 + `handleInnerPAP` + `innerUsername`），第二轮空帧由共享 `tlsTunnel.handleInnerRound` 仅在 `clientSpeaksFirst && frag.IsACK() && !reassemblyInProgress(state)` 时以 `inner==nil` 路由到 `handleMSCHAPv2Ack`（非空帧一律拒绝，绝不在未校验 NT-Response 前放行）；该分支受 `clientSpeaksFirst` 闸门保护，EAP-TLS/PEAP 字节级行为不变，`coordinator.go` 未改。成功后密钥仍由 TLS 导出器（`deriveMPPEKeys`，标签 `"ttls keying material"`，RFC 5281 §8 / RFC 2548）产生，非取自 MS-CHAP-V2 机密。测试：经真实 server-only TLS 握手的整链 MS-CHAP-V2——成功（Access-Accept 含 MS-MPPE 双钥、内层 User-Name 入状态）、强制 `maxFragment=48` 分片、错误口令→`ErrPasswordMismatch`、篡改 MS-CHAP-Challenge→`ErrTTLSInnerProtocol`。门禁 `go build/test ./...`、golangci-lint v2.12.2 全过（RFC 5281 §11.1/§11.2.4、RFC 2548 §2.3.2/§2.3.3、RFC 2759）。
- [x] M9.5 明确失败原因 + 指标 + 单元测试；配置项与默认值 ✅ `auth_stages.go` 的 `mapEAPDispatchError` 新增 EAP-TTLS 专属拒绝原因映射——`ErrTTLSInnerProtocol`→`MetricsRadiusRejectOther`「eap-ttls inner avp protocol violation」、`ErrTTLSInnerNotImplemented`→`MetricsRadiusRejectOther`「eap-ttls inner method unavailable」（均经 `errors.Is` 支持包装链），内层 PAP/MS-CHAP-V2 口令失败继续走 `ErrPasswordMismatch`→`MetricsRadiusRejectPasswdError`、外层隧道错误复用既有 `ErrTLSNotConfigured`/`ErrTLSHandshakeFailed`/`ErrTLSUnexpectedFragment` 映射，告别一律落入 default「eap authentication failed」的笼统原因；`auth_stages_test.go` 表驱动用例补充两条 TTLS 断言（协议违规/未实现的 metricsType + stage + reason）。配置与安全风险说明：`config_schemas.json` 与 `config_manager.go` 兜底 schema 的 `radius.EapMethod` 描述更新——明示 eap-ttls 现已支持内层 PAP（明文口令仅由 TLS 隧道保护）与内层 MS-CHAP-V2（与 eap-peap 同样存在类似 NTLMv1 的攻击面），两者均需 `EapTlsCertFile/EapTlsKeyFile` 且隧道钉死 TLS 1.2，全员可用证书时优先 eap-tls（取代此前「目前仅协商隧道 Start」的过时描述）；`web/src/i18n/{en-US,zh-CN}.ts` 同步双语 UI 描述；`tls_config.go` `NewSettingsTTLSConfigProvider` 安全说明 doc 补充 MS-CHAP-V2 攻击面。`go build/test ./...`、golangci-lint v2.12.2、`cd web && npm run build` 全过。引用 RFC 3748/3579（EAP/RADIUS-EAP）、RFC 5281（EAP-TTLS）、RFC 2759（MS-CHAP-V2 攻击面）。
- [x] M9.6 在 `test/integration/` 增加 TTLS-PAP / TTLS-MSCHAPv2 端到端验收用例（CI 自动执行）✅ 已交付：新增 `test/integration/ttls_test.go`（`//go:build integration`，package `integration`，CI `integration` job 自动执行），实现真实「线上」EAP-TTLSv0 supplicant——经真实 UDP RADIUS 报文驱动运行中的认证服务，复用 `eap_tls_test.go` 的内存双工流（`eapStream`/`eapConn`）、测试 CA 助手与 `assertEAPCode`。关键差异：EAP-TTLS phase 2 为 **peer-initiated**（RFC 5281 §7.3，client-speaks-first），`ttlsSupplicant.authenticate` 在外层 EAP-TLS 握手（`tls.Conn` 客户端 pin TLS 1.2 保证外层帧确定性）完成（`finished` 置位）后**主动先发**内层 AVP flight，而非如 PEAP 般先发空 ACK 取服务端内层请求。两条内层路径各两子用例：① 内层 PAP（RFC 5281 §11.2.5，单轮 User-Name + NUL 填充 User-Password AVP）正确口令→Access-Accept + 32B MS-MPPE-Send/Recv-Key（以**请求 Authenticator** 重绑后解密，RFC 2548 §2.4），错误口令→Access-Reject 且不泄漏 MPPE 密钥；② 内层 MS-CHAP-V2（RFC 5281 §11.2.4 非 EAP/RADIUS VSA 形态）：客户端用 `tls.ConnectionState().ExportKeyingMaterial("ttls challenge", nil, 17)` 派生 17B 隐式挑战（16B 认证挑战 + 1B Ident，RFC 5281 §11.1），以 rfc2759 `GenerateNTResponse` 计算真实 NT-Response，隧道内发 User-Name + MS-CHAP-Challenge(311/11) + MS-CHAP2-Response(311/25) AVP，服务端校验后隧道回送 MS-CHAP2-Success(311/26)（断言 Vendor=311 与 `S=` 串），客户端以**空 EAP-TTLS 帧**确认（触发 `clientSpeaksFirst && IsACK()` 放行分支）→Access-Accept + MPPE 双钥，错误口令→Access-Reject 且不泄漏密钥。AVP 编码器 `ttlsEncodeAVP`/`ttlsEncodePAPFlight`/`ttlsEncodeMSCHAPv2Flight` 复刻生产侧 `encodeTTLSAVP` 线格式（Code 4B|Flags 1B[V/M]|Length 3B|[Vendor 4B]|Data，4 字节边界零填充，gosec 安全字节写）。`configureTTLS`（仅写服务器证书/私钥 + `EapMethod=eap-ttls`，TTLS 为 server-only 无需客户端 CA）+ `seedTTLSUser`（明文口令）经动态配置热切换并以 `restoreEapMethod` 归位；串行（无 `t.Parallel`，依赖进程级全局插件注册/动态配置/限速器）。`make test-integration-pg` 本地全绿（真实 PostgreSQL，4 子用例 + 完整 integration 套件）；`go build ./...`、`go vet -tags=integration`、golangci-lint v2.12.2（`--build-tags=integration` 对本文件 0 问题）通过。引用 RFC 5281 §7.3/§11.1/§11.2.4/§11.2.5（EAP-TTLS phase 2 / 隐式挑战 / 内层 MS-CHAP-V2 / 内层 PAP）、RFC 2759（MS-CHAP-V2 NT-Response）、RFC 2548 §2.4（MS-MPPE VSA salt 加密）、RFC 5216 §2.1.5/§3.1（EAP-TLS 分片帧格式）。

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
- **优先级调整（用户指令）**：mdbook 文档站点列为 P2 **置顶优先实现**——先收编散落文档并对外呈现既有能力。本轮已先行将 `README` 与手册 `overview` 章节补全 **EAP / 802.1X 套件**能力（EAP-MD5 / EAP-MSCHAPv2 / EAP-TLS / PEAP / EAP-TTLS，含 MS-CHAPv2 兼容性优先与类 NTLMv1 攻击面提示，引用 RFC 3748 / 5216 / 5281）；M13.2 中 README「迁入手册 + 留指针」的收敛动作在后续批次进行。

子任务：
- [x] M13.0 评估与现有 GitBook 发布的关系（替代 / 并存），确定单一事实来源与发布管线边界 —— 决策：**并存**。mdbook 为随仓库维护、CI 校验的双语权威手册（置于独立目录 `docs-site/`）；GitBook 经其 GitHub 集成在外部同步（仓库无 `.gitbook.yaml`/`book.json`/GitBook `SUMMARY.md`），两套管线不共享构建步骤、互不冲突；迁移采用「迁入手册 + 原文件留指针」保证单一事实来源。决策已写入站点章节 `gitbook-coexistence`。
- [x] M13.1 mdbook 骨架：`book.toml` + `src/`（zh / en 双语目录）+ 本地构建（`mdbook build`）—— `docs-site/book.toml`（仅 `[output.html]`，扁平输出 `docs-site/book/`）+ `src/SUMMARY.md`（English 与 中文 两个 part，章节一一对应）+ `src/{en,zh}/{overview,documentation-map,gitbook-coexistence}.md` + 双语 `introduction.md`。本地 `mdbook build docs-site` 通过。**注意**：mdbook 0.5.3 不支持原生 `[language.*]` 多语言（"unknown field language"），故采用「单 book + 两 part」结构实现双语。
- [x] M13.2 迁移核心文档（README / AGENT / SECURITY）为双语章节，原文件保留指向站点的入口<br/>进行中（分批迁移）：批次 1 已迁移 `SECURITY.md` —— 新增双语章节 `docs-site/src/{en,zh}/security-policy.md`（忠实呈现 v8.0.8 XSS 公告与升级建议、互为中英文交叉链接），登记进 `SUMMARY.md`（置于 Overview/概述 之后），并更新 `documentation-map` 的安全策略条目指向手册章节（标注「权威/指针」）；按 M13.0「迁入手册 + 原文件留指针」决策，将根目录 `SECURITY.md` 收敛为指向双语章节的简短指针。门禁：`mdbook build docs-site` 通过、lychee `--offline` 离线坏链校验 0 错误。余 `README`（与既有 `overview` 章节有重叠）、`AGENT`（1271 行且被 agent 护栏引用，需谨慎，避免指针化破坏工具链引用）待后续批次。<br/>批次 2（收尾）已收编 `AGENT`：新增双语章节 `docs-site/src/{en,zh}/agent-guide.md` —— 忠实呈现 AGENT.md 的结构化要点（产品范围基线与 `TR-F` 锚定、非目标 `TR-N001`~`TR-N005`；路线图/技能库与 orchestrate→review→groom 调度循环；动手前先理解代码 / 持续验证 / 代码即文档；TDD；仅走 PR + 约定式提交；MVP 最小闭环；质量门禁 go build/test + golangci-lint v2.12.2 + web build + `test/integration` 验收 + `docs/rfcs` 引用；禁用 CGO 与数据库双兼容；反模式清单），明确标注**面向贡献者**且 `AGENT.md` 为权威、本章只摘要不替代；登记进 `SUMMARY.md`（en/zh 对称，置于 security-policy 之后、documentation-map 之前），并更新双语 `documentation-map`（新增手册章节行 + 将 Agent 指南仓库文档行改为「手册摘要 · AGENT.md 权威」+ 刷新迁移计划说明）。为不破坏工具链引用，**未对 AGENT.md 做指针化**，仅在 H1 后加一段保留权威性的简短双语指针（不删改/重排既有内容）。`README` 处置：其面向用户内容已分布于既有 overview/quickstart/vendor/admin/ops/faq 章节，README 继续作为 GitHub 首页并已链接手册（doc-map 已载明），无需指针化。至此 README/AGENT/SECURITY 三份核心文档均已在手册有双语呈现并保留指向站点的入口。门禁：`mdbook build docs-site` 通过、lychee `--offline` 0 错误（1259 OK / 0 Errors）、`go build ./...` 通过；en/zh 文件名对称（语言切换可用）。
- [ ] M13.3 收编功能清单 / 路线图 / RFC 索引为站点章节，建立中英文交叉链接<br/>进行中（分批迁移）：批次 1 已收编 **RFC 索引** —— 新增双语章节 `docs-site/src/{en,zh}/rfc-index.md`（精选、面向实现的协议/RFC 索引，将 RADIUS 核心 / RADIUS-EAP 集成 / EAP 框架与方法 / 动态授权 / 安全传输 / 厂商属性 / IPv6 各 RFC 映射到代码与里程碑，并列出路线图标准 RFC 9190+9427/7170+9930/5931 对应 M10/M11/M12，互为中英文交叉链接），登记进 `SUMMARY.md`（置于 文档地图 之后），更新 `documentation-map` 的 RFC 索引条目指向手册章节（标注「权威/原始目录」）。引用经校正：EAP-MD5 归 RFC 3748 §5.4（非 RFC 3851），EAP-PWD 归 RFC 5931（非 RFC 7542=NAI）；并在 `docs/rfcs/README.md` 顶部加指向手册章节的指针。门禁：`mdbook build docs-site` 通过、lychee `--offline` 0 错误。余 功能清单、路线图 两份待后续批次。
- [x] M13.4 CI 增加 `mdbook build` 产物校验（构建失败或坏链即红）—— `.github/workflows/ci.yml` 新增独立 `docs` 任务：`peaceiris/actions-mdbook`（钉 0.5.3 与本地一致）→ `mdbook build docs-site` → `lychee --offline`（经 `taiki-e/install-action` 安装）对生成 HTML 做离线坏链校验。构建失败或内部坏链均使流水线变红；`--offline` 跳过 http(s) 链接避免网络抖动。**注意**：未用 `mdbook-linkcheck`（0.7.7 与 mdbook 0.5.3 的 RenderContext 不兼容，"missing field sections"），改用 lychee（与渲染协议解耦）。
- [x] M13.5 GitHub Pages 部署工作流 —— 新增 `.github/workflows/pages.yml`：`peaceiris/actions-mdbook`（钉 0.5.3，与 CI/本地一致）构建 `docs-site` → `actions/upload-pages-artifact@v3` → `actions/deploy-pages@v4`（OIDC，`permissions: pages:write/id-token:write`，`concurrency: pages`）。触发：push `main` 命中 `docs-site/**` 或本工作流 + 手动 `workflow_dispatch`。仓库 Pages 源为「GitHub Actions」（REST `build_type: workflow`）。**修复「Pages 无法访问」（两处根因）**：(1) 已配 `build_type=workflow` 却无部署工作流，站点自 2025-11 起未再发布；(2) Pages 自定义域名为 `www.toughradius.net`，但该域名 DNS 指向 Cloudflare/Fastly（GitBook 栈）而非 GitHub Pages（185.199.x），导致 `talkincode.github.io/toughradius` 301 跳转到无法服务 Pages 内容的 `www` → 404。处置：① 本工作流补齐内容发布；② 在仓库设置移除失效的 `www` 自定义域名（`cname=null`），手册改由项目默认地址 <https://talkincode.github.io/toughradius/> 对外服务（已验证 200，内含 RFC 索引等章节）。**后续决策**：手册正式落到自定义域名 `www.toughradius.net` —— 部署工作流在产物写入 `CNAME`，Pages 以该域名对外服务（默认 `talkincode.github.io/toughradius` 自动重定向）；需将 `www` DNS 指向 GitHub Pages（`A` 记录 `185.199.108.153-185.199.111.153` 或仅 DNS 的 `CNAME` → `talkincode.github.io`），GitBook 仅保留 `docs.toughradius.net`（见 `gitbook-coexistence`）。actionlint 通过。
- [x] M13.6 手册内容扩充（用户指令：「文档站点内容过于单薄」）—— 新增 6 个双语章节（`docs-site/src/{en,zh}/`，中英一一对应、互为交叉链接）：`concepts`（核心术语与概念 + 认证请求流转 + 密码协议，链接 RFC 索引）、`quickstart`（二进制 / Docker / 源码三种安装、初始化、默认账号、首个 NAS/用户、`radtest` 验证、调试）、`vendor-guide`（厂商对接案例 ≥7：MikroTik 14988 / 华为 2011 / Cisco 9 / H3C 25506 / 中兴 3902 / 爱快 10055 / 标准 0，含各厂商限速属性与换算公式、VLAN 解析边界、CoA 说明与设备侧参考配置）、`admin-manual`（管理界面逐页手册：仪表盘 / 节点 / NAS / 用户含批量导入 / 计费策略 / 在线会话含 CoA 与强制下线 / 计费记录 / 系统配置 13 项 / 操作员角色）、`ops-guide`（进程模型与 systemd、端口、配置全量参考、环境变量、CLI 参数、数据库、三类 TLS 证书、日志、内存指标口径、备份范围、`cmd/` 工具、加固清单）、`faq`（按主题分组）。全部事实以代码为准（增强器 / 解析器 / config_schemas.json / 前端资源盘点）；`SUMMARY.md`（嵌套列表，双解析器兼容）+ 概述「下一步」+ 文档地图（新增手册章节表）同步更新。门禁：`mdbook build docs-site` 通过、lychee `--offline` 0 错误。
- [x] M13.7 导航栏中英文切换（用户指令：「中英文切换应该在导航栏提供切换链接」）—— 通过 `book.toml` 的 `additional-js` / `additional-css` 注入 `docs-site/assets/lang-toggle.{js,css}`（不 fork 主题模板，升级安全）：英文页菜单栏显示「中文」、中文页显示「English」，改写路径最后一段 `/en/` ↔ `/zh/`（依赖两语言目录文件名一一对应，适配自定义域名 / github.io 项目路径 / 本地预览）；根级中性页（introduction / print）同时给出两个语言入口；样式复用 mdBook 主题变量（`--icons` / `--icons-hover`），带 `aria-label` / `hreflang`。该切换仅作用于 mdBook/Pages 管线，GitBook 继续使用侧边栏分区与每章交叉链接（已在 gitbook-coexistence 中英章节注明）。门禁：`mdbook build` 通过、lychee `--offline` 0 错误、node 对 7 种路径形态的映射断言全部通过。
- [ ] M13.8 厂商场景实战手册（Scenario Cookbook，用户指令：「尽量提供靠近实际应用运维场景的范例指导」）—— 在 `vendor-guide`（属性参考卡）之上新增「场景实战」层：每个场景采用统一的「需求 → ToughRADIUS 侧配置 → 设备侧配置 → 验证 → 排障」五段式，ToughRADIUS 侧每条声明锚定真实代码（enhancer 实际下发的属性、checker 实际执行的拒绝逻辑），设备侧配置标注「示例，以实际固件为准」。分批交付：**批次 1 = MikroTik RouterOS 旗舰场景**（PPPoE 宽带 ISP 分级套餐 + 地址池 + 到期断网 + 并发限制；Hotspot + MAC 认证；CoA/强制下线与 FUP——按代码实际能力，CoA 仅 Session-Timeout/Filter-Id，变速走 Disconnect 重连）；后续批次 = 华为 / H3C / 中兴 / 爱快（对应已有 enhancer）+ Cisco/标准属性场景。门禁：`mdbook build` + lychee `--offline` 0 错误，中英双语对称。
- [x] M13.9 FAQ 实战化扩充（用户指令：「从网上收集相关需求，作为常见问答方式来提供建议」）—— 将 `faq` 从概念问答升级为「症状 → 定位 → 解决」排障式条目，覆盖网络收集的真实痛点（限速不生效 / 单位与方向、地址池不下发、CoA 不通与防火墙放行 UDP 3799、计费在线数不同步、共享密钥与时钟漂移、多 NAS、MAC 认证格式、连续认证失败后响应变慢=reject-delay 守卫）；每条锚定代码事实（属性名、checker、配置项默认值），破除网络误传（CoA 端口非 1700 而是 3799）。门禁同上。**已交付**：新增「地址池不下发 / Framed-Pool 名匹配」「多 NAS 各自密钥」「时钟漂移」「在线 FUP 变速（CoA 仅 Session-Timeout/Filter-Id，变速走 Disconnect 重连）」「CoA 找不到会话」5 条 Q&A；并据代码校准既有错误——计费/在线自动清理实为**休眠**（`SchedClearExpireData` 未注册 cron，仅操作日志按 `@daily` 清理 >1 年，已开 M6.4 跟踪代码修复）。中英双语，`mdbook build` + lychee `--offline`（含 `--include-fragments`）0 错误。

验收口径：`mdbook build` 在 CI 通过且无坏链，中英文章节一一对应，核心散落文档可从站点统一访问，场景实战手册每条 ToughRADIUS 侧声明可追溯到代码；**验收由 CI 构建用例背书**。

---

## M14 — 认证后端扩展：LDAP / AD（bind 校验，PAP 族）

- **关联编号**：`TR-F025`
- **目标**：按 RFC 4511 / RFC 4513 以 LDAP / Active Directory 的 **bind 操作**校验目录账号口令，作为认证流水线中一个**可插拔的 PAP 族校验后端**，让统一身份（LDAP/AD）账号经裸 PAP 与 `EAP-TTLS/PAP`（M9 已交付）接入——补完 EAP-TTLS（M9）「让 LDAP、老账号库无需立即改造证书体系即可接入」所缺的真正后端归宿。来源：issue #199。
- **关键约束（务必先读）**：LDAP/AD bind 只能校验**服务器能拿到明文口令**的方法，即 PAP 族（裸 PAP、`EAP-TTLS/PAP`；将来若新增 `PEAP-GTC` 内层亦可复用同一后端）。`CHAP / MS-CHAP / MS-CHAPv2 / EAP-MD5 / PEAP-MSCHAPv2` **物理上不可行**——这些方法服务器需用明文口令（或 NT-hash）计算挑战响应，而 LDAP 永不交出口令（AD 暴露可读 NT-hash 属性属特权且非通用，不在本里程碑范围）。此约束必须在文档、配置说明与拒绝日志中明示，**不得伪装成「全方法支持 LDAP」**。
- **定位**：服务统一身份 / AD / LDAP 老账号库场景；P2，排在 EAP-TTLS（M9）之后作为其后端补全。
- **开发边界**：作为可插拔校验后端挂在现有 `internal/radiusd/plugins/auth` 流水线之后（围绕 `GetLocalPassword` / `AuthenticateUserWithPlugins` 的口令解析抽象点），**不在协议入口写库分支、不重写认证流水线、不动 EAP 协调器**；默认关闭，凭配置启用；连接 / 超时 / 目录不可达必须有明确拒绝语义与指标（复用 `AuthError` + Prometheus 指标）。
- **技能**：`.agents/skills/reference-rfc/SKILL.md`、`.agents/skills/add-config-schema/SKILL.md`、`.agents/skills/write-go-tests/SKILL.md`、`.agents/skills/add-acceptance-test/SKILL.md`、`.agents/skills/align-feature-checklist/SKILL.md`
- **协议规范**：`rfc4511`（LDAPv3 协议，bind 操作）、`rfc4513`（LDAP 认证方法与安全机制，simple / SASL bind 与 StartTLS）——本地缺失，按 `reference-rfc` 登记；并与 `rfc5281`（EAP-TTLS）、`rfc3748`（EAP 内层）衔接。

子任务：
- [ ] M14.1 LDAP/AD bind 校验后端骨架 + 配置 schema（连接 URL、Base DN、bind DN 模板 / 搜索后 bind、StartTLS/LDAPS、超时、用户过滤模板），可开关、默认关闭（`add-config-schema`）
- [ ] M14.2 接入认证流水线作为 PAP 族校验后端：裸 PAP、`EAP-TTLS/PAP` 复用同一入口；非 PAP 族（CHAP 系 / EAP-MD5 / PEAP-MSCHAPv2）明确拒绝并记录可诊断原因，不在协议入口分叉
- [ ] M14.3 连接健壮性与可观测：连接池 / 重连、超时、bind 失败与目录不可达的拒绝语义与指标（复用 `AuthError` + metrics）
- [ ] M14.4 单元测试（mock / 内嵌 LDAP，覆盖 bind 成功 / 失败 / 不可达 / 仅 PAP 族放行）+ `test/integration/` 验收用例（CI 自动执行）+ 双语文档章节（明示「LDAP 仅 PAP 族」）

验收口径：配置 LDAP / AD 后，PAP 族（含 `EAP-TTLS/PAP`）可用目录账号经 bind 完成认证，且对 `CHAP/MS-CHAP/MS-CHAPv2/EAP-MD5/PEAP-MSCHAPv2` 明确拒绝并给出可诊断原因；**验收由 `test/integration/` 的 CI 用例背书**。

---

## Agent 排期约定

- **入口（自动委托）**：收到"自动委托开发 / 继续推进路线图"类指令时，由 [`.agents/skills/orchestrate-roadmap/SKILL.md`](../.agents/skills/orchestrate-roadmap/SKILL.md) 作为总调度统筹一轮：选题 → 选执行 SOP → 派工 → 质量门禁 → PR → 迭代路线图。
- 调度优先级：先 P1（`M2 → M3 → M8 → M9`），再 P2（**`M13` 双语文档站点置顶优先**，其后 `M4 / M5 / M7 / M10`），最后 P3（`M6 / M11 / M12`）；同优先级里程碑除 M13 置顶外按序号取，P2/P3 仅在更高优先级里程碑无可执行子任务时填充。**M13 置顶依据**：用户指令将 mdbook 文档站点列为优先实现，先收编散落文档并对外呈现既有能力（含 EAP 套件）。EAP 套件优先续接 M1（EAP-TLS）：先 PEAP-MSCHAPv2（兼容）、再 EAP-TTLS（后端适配），TLS 1.3 / TEAP / EAP-PWD 列为中长期 / 按需。
- 单次 agent 任务只认领一个未勾选子任务（最小闭环），完成后在本文件勾选并在 PR 引用里程碑编号。
- **自我迭代**：每轮交付后由 [`.agents/skills/groom-roadmap/SKILL.md`](../.agents/skills/groom-roadmap/SKILL.md) 勾选已交付项、更新里程碑状态、回填/拆分/重排子任务，并保持本文件与功能清单状态一致。
- 任何超出 `TR-F` 清单的需求，必须先提交清单更新 PR，再排入本路线图。
- 每个涉及协议或数据流的子任务，交付时必须带 **CI 可自动执行的验收测试**（单元或 `test/integration/`）。
- 选任务口径：优先取**优先级最高里程碑**中自上而下第一个未勾选的 `- [ ] M*.*`（优先级见里程碑总览的优先级列与上面的调度优先级排期）；同优先级里程碑按序号顺序。agent 在**本机**用你自己的 agent 运行，不在 CI 执行；运行参考与护栏见 [`.agents/README.md`](../.agents/README.md)。
