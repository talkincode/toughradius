# 概述

> English version: [Overview](../en/overview.md)

ToughRADIUS 是一款使用 Go 语言编写、功能强大的开源 RADIUS 服务器，面向 ISP、
企业网络与运营商场景。它实现了标准 RADIUS 协议，并支持 RadSec（RADIUS over TLS），
同时附带基于 React Admin 的现代化 Web 管理界面。

## 核心能力

- **标准 RADIUS** —— 完整支持 RFC 2865（认证）与 RFC 2866（计费）。
- **RadSec** —— 基于 TCP 的 TLS 加密 RADIUS（RFC 6614）。
- **动态授权** —— CoA 与 Disconnect 报文（RFC 5176）。
- **EAP / 802.1X 套件** —— 挑战类方法（EAP-MD5、EAP-MSCHAPv2）与隧道类方法：
  EAP-TLS（RFC 5216，基于证书）、PEAPv0/EAP-MSCHAPv2（Windows / AD 兼容）、
  EAP-TTLS（RFC 5281，内层 PAP / MS-CHAPv2）。MS-CHAPv2 类方法以兼容为先，存在类似
  NTLMv1 的攻击面，可控证书环境优先选用 EAP-TLS。TLS 1.3（RFC 9190）、TEAP 与
  EAP-PWD 在路线图中持续推进。
- **多厂商支持** —— 通过厂商私有属性（VSA）兼容 Cisco、MikroTik、华为等主流网络设备。
- **现代化管理界面** —— 基于 React Admin 的控制台，管理用户、套餐、在线会话、计费与审计日志。
- **多数据库** —— PostgreSQL（默认）与 SQLite（纯 Go，无需 CGO）。

## 服务模型

服务器以并发方式运行多个相互独立的服务；任意一个服务失败都会让进程退出，便于由
守护进程干净地重启。

| 服务             | 协议 / 端口          | 用途                                 |
| ---------------- | -------------------- | ------------------------------------ |
| Web / Admin API  | HTTP，TCP `1816`     | 管理界面与 REST API                  |
| RADIUS 认证      | UDP `1812`           | 认证                                 |
| RADIUS 计费      | UDP `1813`           | 计费                                 |
| RadSec           | TLS over TCP `2083`  | 加密的 RADIUS 传输                    |

## 下一步

- [核心术语与概念](./concepts.md) —— AAA 术语、请求流转，以及各概念在产品中的落点。
- [快速开始](./quickstart.md) —— 约十分钟完成安装、初始化、建用户并用 `radtest` 验证。
- [厂商对接指南](./vendor-guide.md) —— MikroTik、华为、Cisco、H3C、中兴、爱快及
  标准设备的对接案例。
- [管理系统用户手册](./admin-manual.md) —— 管理控制台每个页面的说明。
- [运维指南](./ops-guide.md) —— 生产配置、证书、监控、备份与命令行工具。
- [常见问题解答](./faq.md) —— 大家都会问到的问题。
- [文档地图](./documentation-map.md) —— 找到现有的 README、Agent 指南、安全策略、
  功能清单、路线图与 RFC 索引。
- [mdbook 与 GitBook 并存](./gitbook-coexistence.md) —— 本手册与 GitBook 站点的关系，
  以及单一事实来源策略。
