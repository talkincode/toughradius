# 常见问题解答

> English version: [FAQ](../en/faq.md)

按主题分组的常见问题。如果这里没有覆盖你的问题，请检索
[GitHub issues](https://github.com/talkincode/toughradius/issues) 或新开一个。

## 安装与访问

### 忘记管理员密码怎么办？

使用内置工具，指向服务器实际使用的配置文件：

```bash
go run ./cmd/reset-password -c /etc/toughradius.yml -u admin -p <新密码>
```

`-initdb` 之后的默认账号为 `admin` / `toughradius`。

### SQLite 和 PostgreSQL 怎么选？

SQLite（默认）零依赖——纯 Go 驱动、单文件存于 `{workdir}/data/`——适合实验
与小规模部署。生产规模、高计费量，或需要外部备份工具（`pg_dump`、复制）时
选 PostgreSQL。

### 能不能不用 1812/1813/1816 这些端口？

可以——所有端口都可通过 YAML 或环境变量配置（`radiusd.auth_port`、
`radiusd.acct_port`、`web.port` 等）。见[运维指南](./ops-guide.md#端口)。

### 管理界面的 HTTPS 端口（1817）不工作

Web TLS 监听需要 `{workdir}/private/toughradius.tls.crt` 与 `.key`。文件缺失
或无效时只记录日志并**仅**停掉 HTTPS 监听——1816 上的 HTTP 继续服务。可用
`cmd/certgen` 生成证书或提供自己的证书。

### `-initdb` 可以再跑一次吗？

**不可以。** 它会删除并重建所有表。仅在首次安装时执行。常规升级无需手动
处理表结构——迁移在启动时自动完成。

## 认证

### 所有请求都被忽略 / 超时

最常见的两个原因：

1. 请求的**源 IP 未登记**为 NAS 设备——在 **NAS 设备** 中添加（或修正 NAT
   使源地址符合预期）。
2. **共享密钥不一致**——RADIUS 对报文本身校验失败的包会静默丢弃。

打开 `radiusd.debug: true` 或把日志级别调到 `debug` 查看实际到达的报文。

### 用户认证成功但没有限速

只有 NAS 记录的**厂商**具备限速 VSA（华为、MikroTik、H3C、中兴、爱快）时才
会下发限速属性。登记为 `Standard` 或 `Cisco` 的 NAS 不会收到私有限速属性——
见[厂商对接指南](./vendor-guide.md#限速单位)。同时确认用户的策略确实设置了
`up_rate`/`down_rate`（Kbps）。

### 某个用户为什么被拒绝？

拒绝按原因分类（密码错误、用户不存在、已过期、已停用、并发超限、MAC/VLAN
绑定不符、未授权 NAS 等）——仪表盘展示各类计数，日志记录细节。最容易让人
意外的是**绑定不符**：开启 `bind_mac`/`bind_vlan` 后，首次出现的 MAC/VLAN
会被记录，后续请求必须匹配；更换硬件后请清空用户上的已存绑定值。

### 连续输错密码后响应变慢，为什么？

这是拒绝延迟防爆破机制：在 `RejectDelayWindowSeconds`（默认 10 秒）窗口内
连续拒绝达到 `RejectDelayMaxRejects`（默认 7 次）后，响应会被延迟。两个参数
都可在 **系统配置** 中调整。

### ToughRADIUS 支持 802.1X / 企业级 Wi-Fi 吗？

支持。可用的 EAP 方法：EAP-MD5、EAP-MSCHAPv2、EAP-TLS、PEAPv0/EAP-MSCHAPv2
与 EAP-TTLS（内层 PAP / MS-CHAP-V2）。在 **系统配置 → EapMethod** 中选择。

### EAP-TLS / PEAP / EAP-TTLS 启动不了

基于 TLS 的 EAP 需要在系统配置中设置 `EapTlsCertFile` + `EapTlsKeyFile`——
留空时这些方法按设计处于禁用状态。用 `cmd/certgen` 生成服务器证书、填好
路径后重试。`EapTlsCaFile` 仅在需要校验客户端证书（EAP-TLS）时配置。

### EAP 方法怎么选？

- **EAP-TLS** —— 最强（双向证书），需要给客户端发证。
- **PEAPv0/EAP-MSCHAPv2** —— Windows/AD 兼容；注意 MS-CHAPv2 的类 NTLMv1
  攻击面（见[安全策略](./security-policy.md)）。
- **EAP-TTLS** —— 通过内层 PAP 对接遗留/LDAP 后端，密码受 TLS 隧道保护。

## 会话、CoA 与计费

### 在线会话页的强制下线 / CoA 没有效果

依次检查：设备已开启动态授权（如 RouterOS 的 `radius incoming`、IOS 的
`aaa server radius dynamic-author`）；NAS 记录上的 **CoA 端口** 与设备一致
（默认 3799）；设备接受来自服务器地址的请求。ToughRADIUS 等待 5 秒并重试
2 次后才报告失败。

### 在线会话里出现早已下线的用户

在线记录由 NAS 的计费报文创建/刷新。NAS 停发（重启、断链）时记录可能残留。
请确认设备开启了计费与中间更新（每个 Access-Accept 都会下发
`Acct-Interim-Interval`，默认 300 秒）。残留记录也可在界面上手动下线/删除。

### 计费表一直增长，哪些会自动清理？

操作日志一年后自动清理。对于 `radius_accounting`，`AccountingHistoryDays`
配置（默认 90）定义了计费数据清理的保留窗口；大流量部署请监控表增长并将
数据库级归档纳入运维流程。配置备份**不包含**计费历史——见
[备份与恢复](./ops-guide.md#备份与恢复)。

### 并发会话数没有被限制

计费策略中的 `active_num` 是单用户并发上限（0 表示不限）。该检查统计
`radius_online` 表中的行数，依赖 NAS 正常上报计费——没有计费 Start 报文，
服务器无从得知谁在线。

## 运维

### 怎么接 Prometheus 监控？

目前没有 `/metrics` HTTP 端点；计数器在内存中并通过仪表盘展示。外部监控
建议探测端口、采集日志文件，并对进程退出告警（进程模型为 fail-fast）。

### 怎么安全升级？

停止服务、替换二进制（或拉取新的 Docker 标签）、启动。表结构迁移自动完成。
升级前先做配置备份（系统配置 → 备份）和数据库备份。

### 日志 / 数据 / 证书都存在哪里？

全部位于 `system.workdir`（默认 `/var/toughradius`）之下：`data/`（SQLite
数据库）、`logs/`、`private/`（TLS 材料）、`backup/`（配置快照）。见
[运维指南](./ops-guide.md#工作目录结构)。
