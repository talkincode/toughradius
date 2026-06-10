# 文档地图

> English version: [Documentation Map](../en/documentation-map.md)

本手册采用分批收编的方式构建。下表先列出手册自身的章节，再列出仍位于仓库中的
文档位置，使所有文档都能从同一入口访问。

## 手册章节

| 章节 | 内容 |
| ---- | ---- |
| [概述](./overview.md) | 项目简介、核心能力与服务模型 |
| [核心术语与概念](./concepts.md) | AAA 术语、认证请求流转与密码协议 |
| [快速开始](./quickstart.md) | 安装、初始化、首个用户与调试 |
| [厂商对接指南](./vendor-guide.md) | MikroTik、华为、Cisco、H3C、中兴、爱快等对接案例 |
| [管理系统用户手册](./admin-manual.md) | 管理控制台逐页说明 |
| [运维指南](./ops-guide.md) | 配置参考、证书、监控、备份与命令行工具 |
| [常见问题解答](./faq.md) | 按主题分组的常见问题 |
| [协议与 RFC 索引](./rfc-index.md) | 协议标准与代码的对应关系 |
| [安全策略](./security-policy.md) | 安全公告与升级指引 |

## 仓库内文档

| 文档              | 说明                                       | 当前位置 |
| ----------------- | ------------------------------------------ | -------- |
| README            | 项目介绍、特性与快速上手                    | [README.md](https://github.com/talkincode/toughradius/blob/main/README.md) |
| Agent 指南        | AI Agent 开发指南与协作规则                 | [AGENT.md](https://github.com/talkincode/toughradius/blob/main/AGENT.md) |
| 安全策略          | 安全公告与升级指引                          | [安全策略](./security-policy.md)（权威） · [SECURITY.md](https://github.com/talkincode/toughradius/blob/main/SECURITY.md)（指针） |
| 功能清单          | 功能范围基线（`TR-F` 编号）                 | [docs/feature-checklist.md](https://github.com/talkincode/toughradius/blob/main/docs/feature-checklist.md) · [English](https://github.com/talkincode/toughradius/blob/main/docs/feature-checklist.en.md) |
| 路线图            | 长期路线图与里程碑                          | [docs/roadmap.md](https://github.com/talkincode/toughradius/blob/main/docs/roadmap.md) |
| RFC 索引          | 项目使用的协议标准索引                       | [协议与 RFC 索引](./rfc-index.md)（权威） · [docs/rfcs/README.md](https://github.com/talkincode/toughradius/blob/main/docs/rfcs/README.md)（原始目录） |

> **迁移计划。** 手册现已覆盖 README 中面向用户的内容（概述、快速开始、厂商
> 对接、管理手册、运维、FAQ）；README 继续作为 GitHub 首页并链接到这里。
> 功能清单与路线图是由专门流程维护的**活文档**，保持在 `docs/` 中并以链接
> 方式纳入；Agent 指南留在仓库根目录。
