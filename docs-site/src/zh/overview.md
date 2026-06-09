# 概述

> English version: [Overview](../en/overview.md)

ToughRADIUS 是一款使用 Go 语言编写、功能强大的开源 RADIUS 服务器，面向 ISP、
企业网络与运营商场景。它实现了标准 RADIUS 协议，并支持 RadSec（RADIUS over TLS），
同时附带基于 React Admin 的现代化 Web 管理界面。

## 核心能力

- **标准 RADIUS** —— 完整支持 RFC 2865（认证）与 RFC 2866（计费）。
- **RadSec** —— 基于 TCP 的 TLS 加密 RADIUS（RFC 6614）。
- **动态授权** —— CoA 与 Disconnect 报文（RFC 5176）。
- **EAP 套件** —— EAP-TLS、PEAPv0/EAP-MSCHAPv2 以及 EAP-TTLS 内层方法，更多方法
  在路线图中持续推进。
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

- [文档地图](./documentation-map.md) —— 找到现有的 README、Agent 指南、安全策略、
  功能清单、路线图与 RFC 索引。
- [mdbook 与 GitBook 并存](./gitbook-coexistence.md) —— 本手册与 GitBook 站点的关系，
  以及单一事实来源策略。
