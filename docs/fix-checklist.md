# ToughRADIUS 修复清单

> 来源：[docs/code-review-2026-06.md](code-review-2026-06.md)（2026-06-06 评审）
> 用法：每项修复对应一个 `FIX-xxx` 编号，PR 标题/提交信息引用编号；完成后勾选 `[x]` 并补充验收说明。

## 优先级说明

| 优先级 | 含义 | 期望处理时间 |
| --- | --- | --- |
| P0 | 安全漏洞 / 数据正确性，需立即处理 | 本周内 |
| P1 | 安全加固 / 重要可靠性 | 2 周内 |
| P2 | 技术债 / 一致性 / 测试补齐 | 排期处理 |

---

## P0 — 立即修复

- [ ] **FIX-001｜SQL 注入：`sort` 参数加白名单**
  - 位置：`internal/adminapi/profiles.go:34,83`、`internal/adminapi/nas.go:68,104`
  - 做法：复用 `users.go:20` 的 `allowedXxxSortFields` 白名单模式，非法字段回退默认列；与 `order` 一样做枚举校验。
  - 验收：构造 `?sort=` 注入 payload 返回 400/被忽略；新增负向测试。

- [ ] **FIX-002｜越权：写操作与 backup/restore 增加角色鉴权**
  - 位置：`internal/adminapi/system_backup.go`、其余无 Level 校验的写端点；参考 `operators.go:139`
  - 做法：抽取统一的 Level 鉴权中间件（或 helper），对 backup/restore/settings/nas/profiles 等写操作要求 `admin`/`super`。
  - 验收：`operator` 级 token 调用 backup/restore 返回 403；新增授权负向测试。

- [ ] **FIX-003｜移除/收敛 `TOUGHRADIUS_DEVMODE` 全局认证旁路**
  - 位置：`internal/webserver/server.go:303-306`
  - 做法：删除该旁路，或限定仅在非 production 日志模式 + 显式构建标签下生效，并在启动时打印醒目告警。
  - 验收：生产配置下设置该环境变量不再跳过 JWT。

- [ ] **FIX-004｜EAP-OTP 硬编码口令**
  - 位置：`internal/radiusd/plugins/eap/handlers/otp_handler.go:92-99`
  - 做法：接入真实 OTP 校验；在未实现前**不注册**该 handler（避免可用的固定口令旁路）。
  - 验收：固定值 `123456` 不再可认证；默认配置不暴露该方法。

- [ ] **FIX-005｜`UpdateBind` VLAN 写错字段 Bug**
  - 位置：`internal/radiusd/radius.go:471-475`
  - 做法：`user.Vlanid1 != reqvid1` 分支改为调用 `UpdateUserVlanid1(user.Username, reqvid1)`。
  - 验收：新增单测断言 vlanid1 变更被正确持久化到 vlanid1。

---

## P1 — 安全与可靠性加固

- [ ] **FIX-006｜恢复记账请求共享密钥校验**
  - 位置：`internal/radiusd/radius_acct.go:85`、`internal/radiusd/radius.go:115-117`
  - 做法：启用 `CheckRequestSecret` 校验 Accounting-Request Authenticator；`RADIUSSecret` 返回真实 per-NAS 密钥而非硬编码 `"mysecret"`。
  - 验收：错误密钥的记账请求被拒绝并计入指标。

- [ ] **FIX-007｜EAP 状态加 TTL 清理 + 移除死代码 `EapStateCache`**
  - 位置：`memory_state_manager.go`、`internal/radiusd/radius.go:59-65,428-456`
  - 做法：为 EAP 状态加 TTL/后台过期；删除未使用的 `EapStateCache`。
  - 验收：未完成握手的状态在超时后被回收；内存不再无界增长。

- [ ] **FIX-008｜记账过载不再无界 goroutine**
  - 位置：`internal/radiusd/radius_acct.go:116-123`
  - 做法：池满时丢弃/排队并计指标，而非裸 `go task()`；保留背压。
  - 验收：压测下 goroutine 数有界。

- [ ] **FIX-009｜登录限流 + 请求体大小限制**
  - 位置：`internal/adminapi/auth.go:32`、`internal/webserver/server.go`、`system_backup.go:107`、`adminapi.go:318`
  - 做法：登录端点加速率限制；全局 `middleware.BodyLimit`；上传/恢复设上限。
  - 验收：暴力登录被限流；超大请求体被拒绝。

- [ ] **FIX-010｜补齐安全中间件（CORS / 安全头）**
  - 位置：`internal/webserver/server.go`
  - 做法：按需加 `middleware.CORS` 与 `middleware.Secure`（X-Frame-Options / X-Content-Type-Options / HSTS 等）。
  - 验收：响应包含安全头；CORS 行为可配置。

- [ ] **FIX-011｜停止向客户端泄露 DB 原始错误**
  - 位置：`auth.go:49`、`settings.go:48` 等多处
  - 做法：对外返回通用错误码/文案，详细错误仅记录到日志。
  - 验收：API 错误响应不含表/列名等内部信息。

- [ ] **FIX-012｜禁用账号的已签发 Token 失效**
  - 位置：`internal/adminapi/auth.go:103`
  - 做法：`resolveOperatorFromContext` 复检账号 `Status`；或引入短 TTL + 刷新/吊销机制。
  - 验收：禁用账号后其 token 立即失效。

- [ ] **FIX-013｜Restore 输入强校验**
  - 位置：`internal/adminapi/system_backup.go:129-176`
  - 做法：校验 schema/版本兼容性、字段白名单；限制可被 restore 覆盖的敏感表/字段。
  - 验收：恶意/越权的 restore payload 被拒绝。

---

## P2 — 技术债与一致性

- [ ] **FIX-014｜抽取分页/排序/过滤公共 helper**（根治 SQLi 类不一致）
  - 位置：`users.go`/`nas.go`/`profiles.go`/`operators.go`/`settings.go`
- [ ] **FIX-015｜完成仓储层去 `Deprecated`**，使非 Deprecated 路径成为主路径
  - 位置：`internal/radiusd/radius.go`
- [ ] **FIX-016｜合并/厘清 `registry/` 与 `vendors/` 两套注册表职责**
- [ ] **FIX-017｜补全厂商解析**：Huawei VLAN 解析去 stub；明确「字典支持 ≠ 解析支持」
  - 位置：`internal/radiusd/plugins/vendorparsers/parsers/`
- [ ] **FIX-018｜移除重复 DevTools well-known 路由注册**
  - 位置：`internal/webserver/server.go:107,110`
- [ ] **FIX-019｜统一 JWT 库版本**（v4 / v5 二选一）
  - 位置：`internal/webserver/server.go`、`internal/adminapi/auth.go`
- [ ] **FIX-020｜修正拼写**：`EapMethad`、`Sendresponse` 等
- [ ] **FIX-021｜实现 `SaveSettings`（移除空 placeholder）**
  - 位置：`internal/app/app.go:211`
- [ ] **FIX-022｜移除生产代码中的测试后门分支**
  - 位置：`internal/adminapi/auth.go:105`（`current_operator` 上下文读取）
- [ ] **FIX-023｜补测试**：`settings.go` CRUD、授权负向用例、`UpdateBind`、记账错误分支、缓存淘汰路径

---

## 跟踪汇总

| 优先级 | 编号范围 | 项数 |
| --- | --- | --- |
| P0 | FIX-001 ~ FIX-005 | 5 |
| P1 | FIX-006 ~ FIX-013 | 8 |
| P2 | FIX-014 ~ FIX-023 | 10 |
