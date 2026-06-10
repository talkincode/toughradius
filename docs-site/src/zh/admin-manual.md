# 管理系统用户手册

> English version: [Admin UI Manual](../en/admin-manual.md)

管理控制台运行在 `1816` 端口（HTTP），基于 React Admin 构建，提供中英双语
（默认中文，可在顶栏语言菜单切换）及明暗两套主题。本章逐页介绍。

## 登录与账户

访问 `http://<服务器>:1816`，使用操作员账号登录。初始管理员为
`admin` / `toughradius` —— 请立即在 **账户设置**（右上角头像）中修改，该页
同时可编辑个人资料。密码至少 6 位。

操作员角色（在 **操作员** 页面设置）：

| 级别 | 含义 |
| ---- | ---- |
| `super` | 完整权限，包括系统配置与操作员管理 |
| `admin` | 菜单可见性与 super 相同 |
| `operator` | 仅日常页面——系统配置与操作员菜单不可见 |

## 仪表盘

首页汇总部署状态：

- **统计卡片** —— 用户总数（含停用/过期数）、在线用户数、今日认证与计费
  请求数、今日上行/下行流量（GB）。
- **IPv6 指标条** —— 携带 IPv6 的在线会话、IPv6 地址/前缀/委派前缀计数、
  配置了静态 IPv6 的用户数。
- **图表** —— 认证趋势、策略分布饼图、24 小时上下行流量图。

计数来自内存指标，见[运维指南](./ops-guide.md#指标)。

## 网络节点

用于组织 NAS 设备的逻辑分组（名称、标签、备注）。请先建节点——每台 NAS
都归属一个节点。

## NAS 设备

允许与本服务器通信的网络设备清单。未知源地址的请求会被丢弃。

| 字段 | 说明 |
| ---- | ---- |
| 名称 / Identifier | 自由命名；`identifier` 与 RADIUS `NAS-Identifier` 匹配 |
| IP 地址 / 主机名 | RADIUS 报文的源地址 |
| 厂商 | **决定厂商属性行为** —— 标准、Cisco、华为、MikroTik、H3C、中兴、爱快（见[厂商对接指南](./vendor-guide.md)） |
| 密钥 | RADIUS 共享密钥 |
| CoA 端口 | CoA/Disconnect 的目标端口，默认 `3799` |
| 节点 / 状态 / 标签 / 备注 | 归属与生命周期管理 |

## RADIUS 用户

拨入用户账号。列表支持按用户名、姓名、邮箱、手机、IP 过滤，支持 CSV 导出与
列排序。

主要表单字段：用户名/密码、状态（启用/停用）、**计费策略**（必选——速率、
并发、地址池由其继承）、**过期时间**（决定 `Session-Timeout`）、静态
`ip_addr` / `ipv6_addr`、IPv6 前缀池与委派前缀（编辑视图）、联系方式、备注。

**批量导入**：列表工具栏的导入按钮接受 `.xlsx`、`.csv`、`.json` 文件，并
报告成功/失败数量。可先导出现有列表作为模板参考。

详情视图分组展示全部信息，包括 IPv6 细节、MAC/VLAN 绑定值与时间戳。

## 计费策略

可复用的授权模板：

| 字段 | 含义 |
| ---- | ---- |
| `active_num` | 单用户最大并发会话数（0 = 不限） |
| `up_rate` / `down_rate` | 带宽，单位 **Kbps**；按厂商换算（列表对 ≥1024 的值以 Mbps 展示） |
| `addr_pool` | NAS 用于分配地址的 `Framed-Pool` 池名 |
| `ipv6_prefix` / 域 | IPv6 与华为域授权 |
| `bind_mac` / `bind_vlan` | 将用户锁定到首次出现的 MAC / VLAN |

## 在线会话

实时会话（`radius_online`）。列包括会话 ID、Framed IP、NAS 地址/端口、开始
时间、时长、超时与流量计数。过滤器覆盖用户名、会话 ID、IPv4/IPv6 地址与
前缀、NAS 地址、MAC 及开始时间范围。

行级操作（RFC 5176 动态授权的入口）：

- **修改授权（CoA）** —— 发送 `CoA-Request`，携带新的会话超时和/或
  `Filter-Id`。
- **强制下线（Disconnect）** —— 发送 `Disconnect-Request` 终止会话（需确认）。

两者都发往 NAS 的 CoA 端口，结果以通知形式反馈。

## 计费记录

历史计费数据（`radius_accounting`），只读：会话 ID、地址、起止时间、上下行
总流量。过滤条件与在线会话一致并增加结束时间；支持 CSV 导出。

## 系统配置

仅 `super`/`admin` 可见。配置项由 schema 驱动、按组折叠展示（RADIUS 组默认
展开），存储在 `sys_config` 表——修改即时生效，无需重启。13 项 RADIUS 配置：

| 键 | 默认值 | 用途 |
| -- | ------ | ---- |
| `EapMethod` | `eap-md5` | 当前 EAP 方法：`eap-md5`、`eap-mschapv2`、`eap-tls`、`eap-peap`、`eap-ttls` |
| `EapEnabledHandlers` | `*` | 允许的 EAP 处理器逗号白名单 |
| `EapTlsCertFile` / `EapTlsKeyFile` | 空 | EAP-TLS/PEAP/TTLS 的服务器证书/私钥；**留空即禁用基于 TLS 的 EAP** |
| `EapTlsCaFile` | 空 | 校验客户端证书的 CA 链 |
| `EapTlsMinVersion` | `1.2` | 最低 TLS 版本（`1.2`/`1.3`） |
| `IgnorePassword` | `false` | 跳过密码校验（仅测试用） |
| `AccountingHistoryDays` | `90` | 数据清理使用的计费保留窗口 |
| `AcctInterimInterval` | `300` | NAS 中间计费更新间隔（秒） |
| `SessionTimeout` | `3600` | 默认会话超时（秒） |
| `LogLevel` | `info` | RADIUS 日志级别（`debug`/`info`/`warn`/`error`） |
| `RejectDelayMaxRejects` | `7` | 触发延迟响应的连续拒绝次数 |
| `RejectDelayWindowSeconds` | `10` | 拒绝计数窗口（秒） |

工具栏：**保存**、**刷新**、**重置**（恢复默认值，需确认）以及
**备份 / 恢复** —— 导出或回导一份包含节点、NAS、策略、用户、配置与操作员的
JSON 快照（不包含的内容见[运维指南](./ops-guide.md#备份与恢复)）。

## 操作员

控制台账号管理（仅 `super`/`admin`）：用户名、密码、联系方式、级别、状态。
操作日志保存在数据库（`sys_opr_log`），一年后自动清理。

## 界面便利功能

- **语言切换**（顶栏）—— 简体中文 / English，按浏览器记忆。
- **主题切换** —— 明/暗。
- 用户、会话、计费、NAS、节点、操作员列表均支持 **CSV 导出**。
- 所有列表页支持服务端分页与已激活过滤条件标签。
