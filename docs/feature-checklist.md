# ToughRADIUS 功能清单

英文版本：[docs/feature-checklist.en.md](feature-checklist.en.md)

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
| TR-F004 | RADIUS 协议 | EAP 认证 | 支持 EAP handler 注册、启用列表配置和 EAP 状态管理；当前生产基线以 EAP-MD5、EAP-MSCHAPv2 为主。 | `internal/radiusd/eap_helper.go`, `internal/radiusd/plugins/eap`, `internal/app/config_schemas.json` | 已实现 | `eap-otp` 当前含示例固定 OTP，扩展前必须接入真实校验服务并更新测试与安全说明。 |
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

## 优先扩展功能方向

以下方向是当前允许优先推进的产品扩展。实现前仍需拆分 MVP、补充测试，并在 Issue / PR 中引用对应功能编号。

| 优先级 | 关联编号 | 扩展方向 | 目标范围 | 开发边界 |
| --- | --- | --- | --- | --- |
| P1 | TR-F004 | EAP-TLS 支持 | 在现有 EAP handler 体系下新增 EAP-TLS，支持证书校验、TLS 握手状态管理、用户身份映射和明确的失败原因。 | 不重写 EAP 协调器；先交付最小可用认证链路，再扩展证书策略、吊销检查和管理端配置。 |
| P1 | TR-F010 / TR-F012 / TR-F013 | CoA 动态授权支持 | 在用户管理面板提供授权策略触发能力，支持对在线用户发起 CoA / Disconnect 等动态授权动作，并记录触发结果。 | 后端先抽象 CoA 发送服务和审计结果；前端只暴露可验证的安全动作，避免直接拼装任意 RADIUS 包。 |
| P1 | TR-F007 / TR-F011 / TR-F015 | IPv6 相关能力增强 | 完善 IPv6 地址、IPv6 前缀、Delegated-IPv6-Prefix 在用户、在线会话、计费记录、审计日志和 Dashboard 中的查询与展示。 | 不只做字段展示；协议解析、数据库字段、过滤条件、前端列表和审计口径必须一起闭环。 |

## 当前非目标方向

| 编号 | 非目标方向 | 说明 |
| --- | --- | --- |
| TR-N001 | 计费支付 / 订单 / 财务系统 | 当前项目只维护 RADIUS 计费记录与审计，不默认扩展为完整收费系统。 |
| TR-N002 | CRM / 工单 / 客户自助门户 | 当前管理后台面向 RADIUS 运维管理，不默认扩展为客户运营平台。 |
| TR-N003 | 通用可视化监控平台 | 当前 Dashboard 只服务 RADIUS 运维视图，不替代 Prometheus、Grafana 等通用监控系统。 |
| TR-N004 | 多租户 SaaS 平台 | 当前模型以单实例管理为基线；多租户需要先完成权限、数据隔离和迁移设计。 |
| TR-N005 | 重写协议栈或替换管理框架 | 除非有明确缺陷和迁移方案，否则不以重写为开发方向。 |
