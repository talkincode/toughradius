---
name: add-eap-method
description: 在现有 EAP handler 体系下新增一种 EAP 认证方法（如 EAP-TLS）(TR-F004)。涉及 EAP 握手、分片、状态管理与失败语义时使用。
---

# 技能：新增 EAP 认证方法

> 关联功能编号：`TR-F004`　适用里程碑：M1（EAP-TLS）

## 何时使用
需要在现有 EAP 体系下新增一种 EAP 方法（如 EAP-TLS）时。

## 前置检索
```text
view internal/radiusd/plugins/eap/coordinator.go         # 协调器，禁止重写
view internal/radiusd/plugins/eap/interfaces.go          # handler 接口
file_search "internal/radiusd/plugins/eap/handlers/*_handler.go"
view internal/radiusd/plugins/eap/statemanager/          # EAP 状态管理
grep_search "EapMethod" --include internal/app/**         # 启用列表配置
```
参考已有：`md5_handler.go`、`mschapv2_handler.go`、`otp_handler.go`。

## 协议规范与参考实现

**国际标准规范（优先读仓库内 `docs/rfcs/`）：**
- `rfc3748-eap.txt` — EAP 框架
- `rfc5216-eap-tls.txt` — EAP-TLS（握手、分片、身份）
- `rfc3579-radius-eap-support.txt` — RADIUS 承载 EAP（EAP-Message / Message-Authenticator）
- `rfc5247-eap-key-management.txt` — EAP 密钥管理
- `rfc7499-packet-fragmentation.txt` — RADIUS 分片
- 其他相关：`rfc5281-eap-ttls.txt`、`rfc7170-teap.txt`（如扩展隧道类方法）

缺失的规范按 `../reference-rfc/SKILL.md` 补录。

**参考实现（仅思路参考，注意许可与协议兼容，禁止直接拷贝不兼容代码）：**
- BeryJu `radius-eap`：<https://github.com/BeryJu/radius-eap>
- 实现笔记：<https://beryju.io/blog/2025-05-implementing-eap/>

## 实现步骤
1. **Handler 骨架**：在 `internal/radiusd/plugins/eap/handlers/<method>_handler.go` 实现 handler 接口（模仿 mschapv2）。
2. **状态管理**：复用 `statemanager`，多轮握手 / 分片（TLS）通过 EAP State 关联，不要在协调器内写分支。
3. **注册**：按现有 handler 注册方式接入协调器与启用列表（`EapMethod` 配置项）。
4. **失败语义**：失败返回明确原因，转换为 `AuthError` 并打 metrics（参考 `internal/radiusd/errors` 与 `radius_metrics.go`）。
5. **配置 schema**：如需新增配置（如证书路径），见 `../add-config-schema/SKILL.md`。

## 边界
- 不重写 EAP 协调器（`coordinator.go`）。
- EAP-TLS 先交付最小可用认证链路，证书吊销 / 策略后续里程碑再扩展。
- `eap-otp` 当前为示例固定 OTP，扩展真实方法时不要照搬其固定值。

## 验收
- [ ] handler 单元测试 + 端到端认证测试通过
- [ ] 失败场景有明确拒绝原因与指标
- [ ] `go test ./internal/radiusd/...` 与 `golangci-lint run` 通过
- [ ] 在 `test/integration/` 增加端到端验收用例（见 `../add-acceptance-test/SKILL.md`），CI 自动执行
- [ ] PR 引用 `TR-F004` 与 M1 子任务编号，并引用所依据的 RFC 条款
