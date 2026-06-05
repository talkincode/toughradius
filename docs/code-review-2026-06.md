# ToughRADIUS 整体评审分析报告

> 评审日期：2026-06-06　|　评审范围：后端 Go (~106K 行 / 222 文件) + 前端 React (~17K 行 TS)
> 配套修复清单：[docs/fix-checklist.md](fix-checklist.md)

本报告对 ToughRADIUS v9 做一次整体架构与代码质量评审，所有「严重问题」均已在源码中逐行核实并附 `file:line` 证据。本文为只读分析，未修改任何业务代码。

---

## 1. 项目概览

| 维度 | 现状 |
| --- | --- |
| 语言 / 版本 | Go 1.25，前端 TypeScript + React Admin 5 + Vite |
| 规模 | Go ~105,813 行 / 222 文件（其中 80 个测试文件）；前端 ~17,118 行 |
| 核心依赖 | Echo v4、GORM、`layeh.com/radius`、`panjf2000/ants`、zap、`robfig/cron` |
| 数据库 | PostgreSQL（默认）、SQLite（纯 Go，无 CGO） |
| 服务模型 | `main.go` 用 `errgroup` 并发启动 Web/Admin、RADIUS Auth、RADIUS Acct、RadSec 四服务 |
| 自报测试覆盖率 | ~27.7%（见 `internal/radiusd/TEST_REPORT.md`） |

---

## 2. 架构亮点

| 方面 | 评价 | 关键位置 |
| --- | --- | --- |
| 并发模型 | `errgroup` 并发四服务，任一崩溃即整体退出，边界清晰 | `main.go:89-123` |
| 认证流水线 | stage 流水线（metadata→nas_lookup→rate_limit→vendor_parse→load_user→eap→plugin_auth）干净、可测、可插拔 | `internal/radiusd/auth_pipeline.go`, `auth_stages.go:27-41` |
| 错误体系 | typed error（`AuthError`/`AcctError` 携带 metrics tag，支持 `errors.As`/`Unwrap`），全项目工程质量最高的部分 | `internal/radiusd/errors/errors.go` |
| 插件注册表 | 读取访问器返回切片**副本**，并发读安全 | `internal/radiusd/registry/registry.go:81-115` |
| 依赖注入 | `Application` 通过接口（DBProvider / ConfigProvider / SettingsProvider…）注入，避免全局耦合 | `internal/app/app.go:31-39` |
| 统一响应封装 | `ok/paged/fail` 信封一致 | `internal/adminapi/responses.go` |
| CI/CD | lint + 单测 + PostgreSQL 集成测试 + 多平台交叉编译 + govulncheck + gosec | `.github/workflows/ci.yml` |

---

## 3. 🔴 严重问题（已逐一在代码中验证）

### S1. SQL 注入 — `ORDER BY` 未做白名单
`profiles.go:83` 与 `nas.go:104`：`query.Order(sortField + " " + order)`，其中 `sortField = c.QueryParam("sort")` 直接拼接、**无白名单**。对比 `users.go:20` 有 `allowedUserSortFields` 白名单——说明本应统一防护却被遗漏。
- 证据：`internal/adminapi/profiles.go:34,83`、`internal/adminapi/nas.go:68,104`、对照 `internal/adminapi/users.go:20,276`

### S2. 鉴权严重缺失 — 越权访问
角色/Level 检查**仅存在于 `operators.go`**。`system_backup.go` 等所有其他端点无任何角色校验。任意已登录 operator 即可：
- 调用 `/system/backup`（按代码注释会导出**管理员密码哈希 + RADIUS 明文密码**）
- 调用 `/system/restore`（覆盖全表，可写入任意 `super` 账号 → 账号接管）

JWT 携带 `role` claim 但基本未用于授权。
- 证据：`internal/adminapi/operators.go:139,178,207`（唯一有 Level 检查处）、`internal/adminapi/system_backup.go`（无 `resolveOperator`/`Level`/`c.Get`）

### S3. 全局认证旁路开关
`TOUGHRADIUS_DEVMODE=true` 时 `jwtSkipFunc` 对**所有请求**返回跳过 → 一旦生产误设该环境变量，整个 API 裸奔。
- 证据：`internal/webserver/server.go:303-306`

### S4. EAP-OTP 认证后门
`expectedOTP := "123456"`（带 TODO）。若启用 `eap-otp`，等于固定口令认证旁路。
- 证据：`internal/radiusd/plugins/eap/handlers/otp_handler.go:92-99`

### S5. 真实 Bug — VLAN 绑定写错字段
`UpdateBind` 中 `user.Vlanid1 != reqvid1` 分支错误调用 `UpdateUserVlanid2(...reqvid1)`。两个分支都更新 vlanid2，**vlanid1 永远不会被持久化**。
- 证据：`internal/radiusd/radius.go:471-475`

### S6. 记账请求未校验共享密钥
`CheckRequestSecret` 被注释掉；三个服务 `InsecureSkipVerify:true`；`RADIUSSecret` 返回硬编码 `"mysecret"`。Accounting-Request 的 Authenticator 本可验证却被关闭 → 可伪造记账记录。
- 证据：`internal/radiusd/radius_acct.go:85`、`internal/radiusd/server.go:20,36,56`、`internal/radiusd/radius.go:115-117`
- 说明：Access-Request Authenticator 按 RFC 本身不可验证，auth 侧 `InsecureSkipVerify` 属正常；问题集中在记账侧与硬编码 fallback 密钥。

---

## 4. 🟠 中等问题

| 编号 | 问题 | 证据 |
| --- | --- | --- |
| M1 | EAP 状态内存无限增长：`MemoryStateManager` 无 TTL/后台过期，未完成握手永久驻留 → 慢速内存泄漏/DoS；另有死代码 `EapStateCache` 同样无过期 | `internal/radiusd/plugins/eap/.../memory_state_manager.go`、`radius.go:59-65,428-456` |
| M2 | 记账过载时无界 goroutine：ants 池满时 fallback 到裸 `go task()`，丧失背压 | `internal/radiusd/radius_acct.go:116-123` |
| M3 | 登录端点无限流 → 暴力破解 | `internal/adminapi/auth.go:32` |
| M4 | 无请求体大小限制：`restoreSystem`/`ImportData` 用 `io.ReadAll` 读整文件 → 内存耗尽 DoS | `internal/adminapi/system_backup.go:107`、`adminapi.go:318` |
| M5 | 缺 CORS / 安全响应头 / BodyLimit 中间件 | `internal/webserver/server.go` |
| M6 | DB 原始错误外泄：多处把 `err.Error()` 作为 details 返回客户端，泄露表/列名 | `auth.go:49`、`settings.go:48` 等 |
| M7 | Token 不可吊销：账号被禁用后已签发的 12h token 仍有效 | `internal/adminapi/auth.go:103` |
| M8 | Restore 校验薄弱：仅检查 `Version != ""` 即盲目 upsert 攻击者提供的 operator 行 | `internal/adminapi/system_backup.go:129-176` |

---

## 5. 🟡 代码质量与技术债

- **仓储重构半成品**：`radius.go` 大量 `Deprecated:` 方法（GetNas / GetValidUser / AddRadiusOnline…）仍是线上主路径。
- **两套并行注册表**（`registry/` 与 `vendors/`）职责重叠，增加认知成本。
- **厂商解析名不副实**：~18 个厂商字典但仅 4 个真实 parser（default/huawei/h3c/zte）；Huawei VLAN 解析是 stub（硬编码 `Vlanid1=0,Vlanid2=0`）。
- **分页/排序/过滤代码在 5+ 文件复制粘贴**——正是白名单被漏掉导致 S1 SQLi 的根因；应抽公共 helper。
- **重复路由**：DevTools well-known 路由在 `server.go` 注册两次。
- **JWT 库版本混用**：同层并存 `golang-jwt/jwt/v4`（server.go）与 `jwt/v5`（auth.go）。
- **标识符拼写错误**：`EapMethad`、`Sendresponse` 等，机器翻译痕迹。
- **`SaveSettings` 是空 placeholder**（带 TODO）：`internal/app/app.go:211`。
- **测试后门入生产**：`resolveOperatorFromContext` 含读取 `current_operator` 的测试分支：`auth.go:105`。

---

## 6. 测试与 CI

- 80 个测试文件，覆盖面广，含真实 UDP 集成测试 + PostgreSQL CI 集成。
- **缺口**：`settings.go` 无测试；**无授权负向测试**（无任何断言 operator 被拒绝 backup/restore——因为根本没有该检查）；记账错误分支、`UpdateBind`、缓存淘汰路径未覆盖。
- `web/dist` 缺失时本地直接 `go test ./...` 会因 `go:embed` 失败，需先 `cd web && npm run build`。

---

## 7. 总评

架构骨架优秀（流水线、错误体系、DI、CI 均为加分项），但**安全鉴权层与共享密钥处理是明显短板**，且存在 1 个数据持久化真实 bug（S5）。属于「好底子 + 待加固」的中后期项目，建议优先以安全为主线做一轮加固，再清理技术债。

后续修复请按 [docs/fix-checklist.md](fix-checklist.md) 逐项跟踪。
