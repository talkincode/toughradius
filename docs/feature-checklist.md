# ToughRADIUS 功能清单

英文版本：[docs/feature-checklist.en.md](feature-checklist.en.md)

开发路线图与里程碑：[docs/roadmap.md](roadmap.md)（中文详版：[docs/roadmap.zh.md](roadmap.zh.md)）

本文档是 ToughRADIUS 的功能范围基线。后续需求、Issue、PR 和代码改动必须先对齐本清单中的功能编号；无法映射到现有编号的需求，必须先更新本清单并说明范围变化，再进入实现。

## 维护规则

1. 每个新需求必须引用至少一个功能编号，例如 `TR-F001`。
2. 新增、拆分、下线或改变功能方向时，先更新本清单的状态、范围和验收口径。
3. 功能实现应保持 MVP 粒度：每次只交付可验证、可回滚、可独立使用的最小闭环。
4. 不在本清单中的产品方向，不作为默认开发方向；确需推进时先补充清单并获得明确共识。
5. 清单只描述产品能力边界；具体接口、结构体和算法仍以代码、测试和内联注释为准。

## 状态说明

| 状态 | 含义 |
| --- | --- |
| 核心基线 | 项目主干能力，后续改动必须优先保持兼容和稳定。 |
| 已实现 | 代码中已有主要闭环，可继续补齐测试、体验或边界处理。 |
| 部分实现 | 代码中已有入口或雏形，但不能默认视为完整生产能力。 |
| 可扩展 | 属于允许扩展的方向，但扩展必须走现有架构和测试边界。 |
| 非目标 | 当前不作为项目默认开发方向。 |

## 功能清单表

| 编号 | 功能域 | 功能项 | 标准范围 / 验收口径 | 现有入口 / 模块 | 状态 | 开发边界 |
| --- | --- | --- | --- | --- | --- | --- |
| TR-F001 | RADIUS 协议 | RADIUS 认证服务 | 支持 UDP Access-Request，完成 NAS 识别、共享密钥、用户认证、策略校验、Access-Accept / Access-Reject 响应、认证日志和指标。 | `main.go`, `internal/radiusd/radius_auth.go`, `internal/radiusd/auth_pipeline.go`, `internal/radiusd/plugins/auth` | 核心基线 | 不绕过认证流水线；新增认证逻辑必须注册为 validator、checker、guard 或 enhancer，并补测试。 |
| TR-F002 | RADIUS 协议 | RADIUS 计费服务 | 支持 UDP Accounting-Request，处理 start、interim/update、stop、accounting on/off，维护在线会话和计费记录。 | `internal/radiusd/radius_acct.go`, `internal/radiusd/plugins/accounting`, `internal/domain/radius.go` | 核心基线 | 不直接在协议入口写业务分支；新增计费行为优先扩展 accounting handler。 |
| TR-F003 | RADIUS 协议 | RadSec 服务 | 支持 RADIUS over TLS 入口，将认证和计费请求分发到既有服务，使用配置中的证书路径和端口。 | `main.go`, `internal/radiusd/radsec_server.go`, `internal/radiusd/radsec_service.go`, `pkg/certgen` | 核心基线 | 必须保持与普通 RADIUS 认证/计费逻辑复用；证书和 TLS 行为变更需覆盖配置与集成测试。 |
| TR-F004 | RADIUS 协议 | EAP 认证 | 支持 EAP handler 注册、启用列表配置、EAP 状态管理与 TLS 隧道分片；已交付方法：EAP-MD5、EAP-MSCHAPv2、EAP-TLS、PEAPv0/EAP-MSCHAPv2、EAP-TTLS（内层 PAP 与 MS-CHAP-V2）。 | `internal/radiusd/eap_helper.go`, `internal/radiusd/plugins/eap`, `internal/radiusd/plugins/eap/handlers`, `internal/radiusd/plugins/eap/tlsfragment`, `internal/app/config_schemas.json` | 已实现 | EAP 方法枚举为 `eap-md5/eap-mschapv2/eap-tls/eap-peap/eap-ttls`（默认 `eap-md5`）；PEAP/TTLS 复用 EAP-TLS 证书建立 TLS 1.2 隧道，文档须保留 MS-CHAPv2 类 NTLMv1 攻击面提示，默认不削弱外层 TLS。`eap-otp` 仅存于代码（未纳入枚举）且含示例固定 OTP，启用前须接入真实校验服务并更新测试与安全说明。 |
| TR-F005 | 厂商兼容 | VSA 解析与响应增强 | 支持标准属性、厂商字典、请求解析和响应增强；重点维护 Huawei、H3C、ZTE、Mikrotik、iKuai 等现有路径。 | `share/dictionary*`, `internal/radiusd/vendors`, `internal/radiusd/plugins/vendorparsers`, `internal/radiusd/plugins/auth/enhancers` | 可扩展 | 新厂商必须按 parser / enhancer / registry 模式接入，并用厂商样例包覆盖解析和响应属性。 |
| TR-F006 | 认证策略 | 用户状态、过期、在线数、MAC/VLAN 绑定 | 认证时校验用户启停、过期时间、在线数量限制、MAC 绑定、VLAN 绑定，并输出明确拒绝原因和指标。 | `internal/radiusd/plugins/auth/checkers`, `internal/radiusd/errors`, `internal/app/radius_metrics.go` | 核心基线 | 策略变化不得降低默认安全性；新增拒绝场景必须定义错误类型、指标和测试。 |
| TR-F007 | 用户与资费 | RADIUS 用户管理 | 管理用户基础信息、账号密码、Profile 关联、地址池、IPv4/IPv6、速率、VLAN、MAC、状态和到期时间。 | `internal/adminapi/users.go`, `internal/domain/radius.go`, `web/src/resources/radiusUsers.tsx` | 已实现 | 用户字段新增必须同步后端请求结构、领域模型、前端资源、验证和列表过滤。 |
| TR-F008 | 用户与资费 | RADIUS Profile 管理 | 管理资费/Profile 的速率、地址池、并发数、域、IPv6 前缀池、MAC/VLAN 绑定和启停状态。 | `internal/adminapi/profiles.go`, `internal/app/profile_cache.go`, `web/src/resources/radiusProfiles.tsx` | 已实现 | 动态 Profile 与静态快照行为必须保持可预测；变更需覆盖缓存和用户继承逻辑。 |
| TR-F009 | 网络资源 | 节点与 NAS 设备管理 | 管理网络节点、NAS 名称、IP、标识、Secret、CoA 端口、厂商编码、状态、标签和备注。 | `internal/adminapi/nodes.go`, `internal/adminapi/nas.go`, `internal/domain/network.go`, `web/src/resources/nodes.tsx`, `web/src/resources/nas.tsx` | 已实现 | NAS Secret、厂商编码和 CoA 行为变更必须保持与协议入口兼容。 |
| TR-F010 | 会话控制 | 在线会话查询与强制下线 | 查询、过滤、查看在线会话；删除会话时尝试向 NAS 发送 Disconnect-Request。 | `internal/adminapi/sessions.go`, `internal/domain/radius.go`, `web/src/resources/onlineSessions.tsx` | 已实现 | 强制下线是数据库删除加异步 CoA；若改成强一致流程，必须定义超时、重试和失败反馈。 |
| TR-F011 | 计费审计 | 计费记录查询 | 查询、过滤、排序和查看认证计费历史，支持时间范围、用户、NAS、会话、IP、MAC 等常用条件。 | `internal/adminapi/accounting.go`, `web/src/resources/accounting.tsx` | 已实现 | 计费记录是审计数据，默认不提供任意修改；清理策略必须可配置且可测试。 |
| TR-F012 | 管理 API | Admin REST API | 提供登录、当前用户、Dashboard、用户、Profile、NAS、节点、会话、计费、系统设置、操作员等 REST 接口。 | `internal/adminapi`, `internal/webserver/server.go`, `web/src/providers/dataProvider.ts` | 核心基线 | 新接口必须复用统一响应、分页、过滤、验证、鉴权和测试模式。 |
| TR-F013 | 管理前端 | React Admin 管理后台 | 提供登录、Dashboard、用户、Profile、NAS、节点、会话、计费、操作员、账号设置和系统配置页面。 | `web/src/App.tsx`, `web/src/resources`, `web/src/pages`, `web/src/providers` | 已实现 | 前端新增页面必须对齐已有资源路由和 API 映射，不引入独立管理入口。 |
| TR-F014 | 系统配置 | 动态配置与配置 Schema | 通过 `sys_config` 和内嵌 schema 管理 RADIUS 运行参数，支持配置查询、编辑、schema 输出和 reload。 | `internal/app/config_manager.go`, `internal/app/config_schemas.json`, `internal/adminapi/settings.go`, `web/src/pages/SystemConfigPage.tsx` | 已实现 | 新配置项必须先加入 schema，提供默认值、类型、范围、国际化 key 和测试。 |
| TR-F015 | 运维监控 | Dashboard、指标和运行监控 | 展示用户数、在线数、认证/计费趋势、流量、Profile 分布；采集系统和进程 CPU/内存以及 RADIUS 指标。 | `internal/adminapi/dashboard.go`, `internal/app/jobs.go`, `internal/app/radius_metrics.go`, `pkg/metrics` | 已实现 | 指标名和 Dashboard 数据结构变更必须兼容前端和历史解释口径。 |
| TR-F016 | 系统管理 | 操作员、登录与账号设置 | 支持默认超级管理员初始化、JWT 登录、当前账号信息、操作员 CRUD、账号资料和密码变更。 | `internal/app/initdb.go`, `internal/adminapi/auth.go`, `internal/adminapi/operators.go`, `web/src/pages/AccountSettings.tsx`, `web/src/resources/operators.tsx` | 已实现 | 权限模型当前较轻；引入 RBAC 前必须先拆分权限范围和迁移路径。 |
| TR-F017 | 数据存储 | PostgreSQL / SQLite 数据库支持 | 支持 GORM 自动迁移、PostgreSQL 生产部署和 SQLite 开发/轻量部署。 | `config/config.go`, `internal/app/database.go`, `internal/domain`, `internal/app/app.go` | 核心基线 | Schema 变更必须兼容两类数据库，并补迁移/查询测试。 |
| TR-F018 | Web 服务 | 静态前端、API、健康检查和中间件 | 提供 `/admin` SPA、API base path、ready/realip、CORS、JWT、请求上下文和统一错误处理。 | `internal/webserver/server.go`, `web/static.go`, `pkg/web` | 已实现 | Web 层只做协议和中间件承载，业务规则应留在 adminapi、app 或 radiusd。 |
| TR-F019 | CLI 工具 | 运维与开发辅助命令 | 提供主程序参数、RADIUS 测试、证书生成、配置 schema 校验、压测、密码重置和演示数据工具。 | `main.go`, `cmd/radtest`, `cmd/certgen`, `cmd/config-tool`, `cmd/benchmark`, `cmd/reset-password`, `cmd/demo-seed` | 已实现 | CLI 必须保持脚本友好，输出和退出码变化需在测试或文档中说明。 |
| TR-F020 | 构建部署 | 构建、Docker 和前端嵌入 | 支持 Go 构建、前端构建、静态资源嵌入、Docker 镜像和 Makefile 工作流。 | `Makefile`, `Dockerfile`, `web/vite.config.ts`, `web/static.go`, `.github` | 已实现 | 构建链路变更必须同时验证后端二进制和前端产物。 |
| TR-F021 | 协议资料 | RFC 与字典资料维护 | 保留 RADIUS、EAP、RadSec、VSA 相关 RFC 和 FreeRADIUS 字典资料，支撑协议实现和厂商扩展。 | `docs/rfcs`, `share`, `internal/radiusd/vendors` | 已实现 | 协议资料更新不能替代代码测试；新增资料需说明对应实现或待实现功能编号。 |
| TR-F022 | 安全与质量 | 测试、验证、输入约束和审计习惯 | 通过单元测试、集成测试、白名单排序、输入校验、密码哈希、JWT 和日志指标降低回归风险。 | `*_test.go`, `.golangci.yml`, `internal/adminapi/helpers.go`, `pkg/validator`, `pkg/common` | 核心基线 | 安全边界变更必须有针对性测试；不得为了快速开发移除验证或鉴权。 |
| TR-F023 | 文档工程 | 双语文档站点（mdbook） | 用 mdbook 构建中英文双语文档站点，收编散落文档（README、AGENT、SECURITY、RFC 索引等），并以链接方式暴露仍由专用技能维护的功能清单与路线图，提供统一导航、本地 `mdbook build` 构建与 CI 产物校验。 | `docs-site/book.toml`, `docs-site/src`（zh / en 双语章节）, `.github/workflows/ci.yml`（docs 任务）, `docs/` | 已实现 | 文档站点只做现有文档的结构化与双语化，不替代以代码与测试为准的口径；中英文章节必须一一对应、同步维护，且遵守 `TR-N003` / `TR-N006`，不扩展为产品门户或托管 Portal；GitHub Pages / GitBook 的发布边界必须保持单一事实来源，避免内容漂移。 |
| TR-F024 | 代码规范 | Go API 文档与注释规范 | 对齐 Go 标准库风格：导出标识符必须有 godoc 注释，包注释（`doc.go`）、可运行示例（`Example`）、错误与并发语义说明齐备，并可由 lint / CI 度量。 | `.agents/skills/document-go-apis`, 各包 doc 注释, `.golangci.yml` | 已实现 | 规范已通过增量 ratchet 纳入常态门禁；后续新增导出 API 必须保持标准库 godoc 习惯，禁止无信息量的机械式注释。 |
| TR-F026 | 证书管理 | EAP/RadSec 证书集中管理 | 集中管理 EAP-TLS / PEAP / TTLS 与 RadSec 所需的服务器证书与客户端 CA：支持 PEM 导入（解析主题、签发者、序列号、指纹、有效期等元数据并校验私钥与证书匹配）、本地名称编辑、安全导出（默认仅证书，私钥需显式授权并审计）、删除；EAP 配置以证书名称引用（下拉选择）作为证书材料的唯一来源，不再使用文件路径。 | `internal/adminapi/certificates.go`, `internal/domain/system.go`（SysCert）, `internal/app/certstore.go`, `internal/radiusd/plugins/eap/handlers/tls_config.go`, `internal/app/config_schemas.json`, `web/src/resources/certificates.tsx`, `web/src/pages/SystemConfigPage.tsx` | 已实现 | 私钥永不随列表/详情接口返回（`json:"-"`）；导出私钥须显式 `include_key` 且记录操作员审计日志；EAP-TLS/PEAP/TTLS 证书材料仅来自托管证书库（集成测试通过向数据库写入 `sys_cert` 记录验证），未选择证书时基于证书的 EAP 安全拒绝；不在前端拼装协议包，仅暴露后端可验证的安全操作。 |

## 优先扩展功能方向

以下为优先级扩展方向及其交付状态（权威里程碑见 [`docs/roadmap.md`](roadmap.md)）。已交付方向的能力边界已并入上方功能清单表，此处保留以追溯范围与开发边界；未交付方向在实现前仍需拆分 MVP、补充测试，并在 Issue / PR 中引用对应功能编号。

| 优先级 | 关联编号 | 扩展方向 | 状态 | 目标范围 | 开发边界 |
| --- | --- | --- | --- | --- | --- |
| P1 | TR-F004 | EAP-TLS 支持 | 已交付（M1） | 在现有 EAP handler 体系下新增 EAP-TLS，支持证书校验、TLS 握手状态管理、用户身份映射和明确的失败原因。 | 不重写 EAP 协调器；先交付最小可用认证链路，再扩展证书策略、吊销检查和管理端配置。 |
| P1 | TR-F010 / TR-F012 / TR-F013 | CoA 动态授权支持 | 已交付（M2） | 在用户管理面板提供授权策略触发能力，支持对在线用户发起 CoA / Disconnect 等动态授权动作，并记录触发结果。 | 后端先抽象 CoA 发送服务和审计结果；前端只暴露可验证的安全动作，避免直接拼装任意 RADIUS 包。 |
| P1 | TR-F007 / TR-F011 / TR-F015 | IPv6 相关能力增强 | 已交付（M3） | 完善 IPv6 地址、IPv6 前缀、Delegated-IPv6-Prefix 在用户、在线会话、计费记录、审计日志和 Dashboard 中的查询与展示。 | 不只做字段展示；协议解析、数据库字段、过滤条件、前端列表和审计口径必须一起闭环。 |
| P1 | TR-F004 | PEAPv0 / EAP-MSCHAPv2 | 已交付（M8） | 用服务器证书建立 PEAP TLS 隧道，隧道内运行 EAP-MSCHAPv2，为 Windows / AD / 传统企业网络提供兼容认证，并正确导出 MPPE 会话密钥。 | 兼容性优先，不作为先进安全卖点；文档与配置必须明示 MS-CHAPv2 存在类似 NTLMv1 的攻击面（见 Microsoft 文档），默认不削弱外层 TLS；不重写 EAP 协调器。 |
| P1 | TR-F004 | EAP-TTLS（隧道 + 内层 PAP/CHAP/MS-CHAP/MS-CHAPv2） | 已交付（M9，内层 PAP + MS-CHAP-V2；CHAP/MS-CHAP 待续） | 按 RFC 5281 用服务器证书建立 TLS 隧道，隧道内承载 PAP / CHAP / MS-CHAP / MS-CHAP-V2（及内层 EAP），让 LDAP、老账号库、混合客户端无需立即改造证书体系即可接入。 | 内层方法逐个交付（先 PAP，再 MS-CHAP-V2）；后端用户库适配走现有认证流水线，不在协议入口写库分支；不重写 EAP 协调器。 |
| P2 | TR-F004 | EAP-TLS 1.3 升级（RFC 9190） | 计划中（M10） | 在 M1 已交付的 TLS 1.2 EAP-TLS 基线上，按 RFC 9190 支持 TLS 1.3 握手与会话密钥派生，遵循 RFC 9427 的 TLS 1.3 派生规则。 | 保持与 TLS 1.2 客户端向后兼容；先协商再切换，不破坏既有 CA 链校验与身份映射。 |
| P3 | TR-F004 | TEAP（隧道，machine + user chaining） | 计划中（M11） | 按 RFC 7170 / RFC 9930（TEAPv1）实现现代隧道 EAP，支持 machine + user chaining、证书 + 密码组合认证；TLS 1.3 下采用 RFC 9427 派生规则。 | 中长期方向，客户端生态弱于 PEAP；仅在客户端环境可控时优先，不与 PEAP / TTLS 抢第一版资源。 |
| P3 | TR-F004 | EAP-PWD（按需） | 计划中（M12） | 按 RFC 5931 以共享口令完成认证，不为每客户端签发证书，适合 IoT、嵌入式、受控小规模设备。 | 非通用企业 Wi-Fi 首选；按需推进，避免为协议完整性拖入维护沼泽。 |
| P2 | TR-F025 | 认证后端扩展：LDAP / AD（bind 校验，PAP 族） | 进行中（M14） | 按 RFC 4511 / RFC 4513 以 LDAP / Active Directory 的 bind 操作校验目录账号口令，作为认证流水线中可插拔的 PAP 族校验后端，让统一身份（LDAP/AD）账号经裸 PAP 与 `EAP-TTLS/PAP`（M9 已交付）接入，补完 EAP-TTLS「让 LDAP、老账号库无需立即改造证书体系即可接入」所缺的真正后端（来源：issue #199）。 | 仅支持 PAP 族（服务器可拿到明文口令）：`CHAP / MS-CHAP / MS-CHAPv2 / EAP-MD5 / PEAP-MSCHAPv2` 因服务器需明文或 NT-hash 计算挑战、而 LDAP bind 永不交出口令而**物理不可行**，必须在文档/配置/拒绝日志明示、不得伪装支持；作为可插拔后端挂在现有 `internal/radiusd/plugins/auth` 之后，不在协议入口写库分支、不重写认证流水线、不动 EAP 协调器；M14.6 集成验收交付前仍不得视为完整生产闭环。 |
| P2 | TR-F023 | 双语文档站点（mdbook） | 已交付（M13） | 用 mdbook 搭建中英文双语文档站点，收编 README / AGENT / SECURITY / RFC 索引等散落文档，并以交叉链接暴露功能清单与路线图，提供统一导航、本地构建与 CI 产物校验。 | 文档不替代以代码与测试为准的口径；中英文目录结构对应、同步维护；功能清单与路线图作为 living docs 保持在 `docs/`，由专用技能维护。 |
| P2 | TR-F024 | Go API 文档与注释规范（标准库风格） | 已交付（M4） | 制定并落地 godoc / 标准库风格注释规范：导出标识符注释齐全、包注释、`Example`、错误与并发语义说明；提供配套技能与可度量门禁。 | 规范已纳入 lint / CI ratchet；新增或变更导出 API 必须继续满足标准库 godoc 风格。 |

## 当前非目标方向

| 编号 | 非目标方向 | 说明 |
| --- | --- | --- |
| TR-N001 | 计费支付 / 订单 / 财务系统 | 当前项目只维护 RADIUS 计费记录与审计，不默认扩展为完整收费系统。 |
| TR-N002 | CRM / 工单 / 客户自助门户 | 当前管理后台面向 RADIUS 运维管理，不默认扩展为客户运营平台。 |
| TR-N003 | 通用可视化监控平台 | 当前 Dashboard 只服务 RADIUS 运维视图，不替代 Prometheus、Grafana 等通用监控系统。 |
| TR-N004 | 多租户 SaaS 平台 | 当前模型以单实例管理为基线；多租户需要先完成权限、数据隔离和迁移设计。 |
| TR-N005 | 重写协议栈或替换管理框架 | 除非有明确缺陷和迁移方案，否则不以重写为开发方向。 |
| TR-N006 | 托管式 Captive Portal / 访客门户产品 | ToughRADIUS 只作为 RADIUS auth/accounting 后端，不提供、不托管、不运营 Portal 登录页、访客开户、券码、短信/微信/支付 onboarding 或厂商 Portal Server 状态机；这些属于其他产品。 |
