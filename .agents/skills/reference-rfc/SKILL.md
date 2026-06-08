---
name: reference-rfc
description: 检索并引用国际标准协议规范（RFC/IEEE）(TR-F021)。任何改变 RADIUS/EAP/计费/CoA/IPv6 等协议行为的改动都应先查 docs/rfcs 并引用条款。
---

# 技能：检索与引用国际标准协议规范

> 关联功能编号：`TR-F021`（协议资料）　适用：所有协议相关改动

## 原则
任何改变 RADIUS / EAP / 计费 / CoA / IPv6 等协议行为的改动，**必须引用对应的国际标准规范条款**（RFC / IEEE 等），并优先使用仓库内已收录资料。

## 先查本地资料库
仓库内 `docs/rfcs/` 已收录 50+ RFC（见 `docs/rfcs/README.md`），常用：

| 主题 | 文件 |
| --- | --- |
| RADIUS 认证 / 计费 | `rfc2865-radius-authentication.txt` / `rfc2866-radius-accounting.txt` |
| RADIUS 扩展 / 协议扩展 | `rfc2869-radius-extensions.txt` / `rfc6929-protocol-extensions.txt` |
| RADIUS 承载 EAP | `rfc3579-radius-eap-support.txt` |
| EAP 框架 / EAP-TLS / TTLS / TEAP | `rfc3748-eap.txt` / `rfc5216-eap-tls.txt` / `rfc5281-eap-ttls.txt` / `rfc7170-teap.txt` |
| 分片 | `rfc7499-packet-fragmentation.txt` |
| 动态授权 / CoA | `rfc3576-dynamic-authorization.txt` / `rfc5176-coa-disconnect.txt` |
| IPv6 / 前缀委派 | `rfc3162-radius-ipv6.txt` / `rfc4818-ipv6-prefix-delegation.txt` / `rfc6911-ipv6-access-networks.txt` |
| RadSec / RADIUS over TCP | `rfc6614-radsec.txt` / `rfc6613-radius-over-tcp.txt` |
| Status-Server / 实现问题 | `rfc5997-status-server.txt` / `rfc5080-implementation-issues.txt` |

检索：
```text
grep -rni "Message-Authenticator" docs/rfcs/
view docs/rfcs/rfc5216-eap-tls.txt
```

## 本地缺失时补录
1. 确认确实缺失（`ls docs/rfcs/ | grep <编号>`）。
2. 从权威来源获取规范文本（IETF datatracker / rfc-editor）。
3. 按 `docs/rfcs/FILE-NAMING.md` 的命名规范保存（如 `rfcXXXX-<slug>.txt`）。
4. 更新 `docs/rfcs/README.md` 索引，注明对应实现或待实现的功能编号。

## 引用方式
- 代码内联注释引用条款，解释"为什么这样实现"（如 `// per RFC 5216 §2.1.5, fragment via EAP-TLS Length+More-Fragments`）。
- PR 描述列出依据的 RFC 编号与小节。

## 边界
- 规范资料不替代测试；新增资料必须说明对应实现或待实现功能编号（见 `TR-F021` 边界）。
- 不上传有版权限制、禁止再分发的非标准文档。

## 验收
- [ ] 协议改动引用了具体 RFC 条款
- [ ] 新增规范按命名规范保存并登记 README
- [ ] 行为差异有内联注释说明依据
