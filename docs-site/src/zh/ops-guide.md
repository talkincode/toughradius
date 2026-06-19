# 运维指南

> English version: [Operations Guide](../en/ops-guide.md)

在生产环境运行 ToughRADIUS 所需的一切：配置参考、环境变量、TLS/EAP 证书、
存储、监控、备份以及随附的命令行工具。

## 进程模型

一个静态二进制并发运行多个服务（Web/管理 API、RADIUS 认证、RADIUS 计费、
RadSec）。**任一服务失败，整个进程退出**，交由守护程序重启——请使用
systemd、Docker 或同类工具托管。

```ini
# /etc/systemd/system/toughradius.service（参考）
[Unit]
Description=ToughRADIUS server
After=network-online.target

[Service]
ExecStart=/usr/local/bin/toughradius -c /etc/toughradius.yml
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
```

## 端口

| 端口 | 协议 | 服务 | 配置键 |
| ---- | ---- | ---- | ------ |
| 1816 | TCP HTTP | 管理界面 + REST API | `web.port` |
| 1817 | TCP HTTPS | 管理界面 TLS（可选；启动失败不影响整体） | `web.tls_port` |
| 1812 | UDP | RADIUS 认证 | `radiusd.auth_port` |
| 1813 | UDP | RADIUS 计费 | `radiusd.acct_port` |
| 2083 | TCP TLS | RadSec（RFC 6614） | `radiusd.radsec_port` |
| 3799 | UDP（出向） | 发往 NAS 的 CoA/Disconnect | 按 NAS 的 *CoA 端口* 字段 |

## 配置

查找顺序：`-c <文件>` → `./toughradius.yml` → `/etc/toughradius.yml` →
内置默认值。可用 `toughradius -printcfg` 查看合并结果。

```yaml
system:
  appid: ToughRADIUS
  location: Asia/Shanghai        # 定时任务/时间戳所用时区
  workdir: /var/toughradius      # 生产构建的默认值
  debug: false
web:
  host: 0.0.0.0
  port: 1816
  tls_port: 1817
  secret: <随机字符串>            # JWT 签名密钥——务必修改
database:
  type: sqlite                   # sqlite | postgres
  host: 127.0.0.1                # 仅 postgres
  port: 5432
  name: toughradius.db           # sqlite 文件名（位于 {workdir}/data/）或 pg 库名
  user: postgres
  passwd: <密码>
  max_conn: 100
  idle_conn: 10
  debug: false
radiusd:
  enabled: true
  host: 0.0.0.0
  auth_port: 1812
  acct_port: 1813
  radsec_port: 2083
  radsec_worker: 100
  radsec_ca_cert: private/ca.crt        # 相对路径基于 workdir 解析
  radsec_cert: private/radsec.tls.crt
  radsec_key: private/radsec.tls.key
  debug: false                   # true = 完整报文转储
logger:
  mode: production               # development | production
  file_enable: true
  filename: /var/toughradius/toughradius.log
```

### 工作目录结构

启动时 ToughRADIUS 在 `system.workdir` 下创建：

```text
/var/toughradius/
├── data/        # SQLite 数据库、指标数据
├── logs/
├── private/     # TLS 材料（权限 0700）
├── public/
└── backup/      # 配置备份的服务器侧副本
```

### 环境变量

环境变量优先于 YAML 文件：

| 变量 | 覆盖项 |
| ---- | ------ |
| `TOUGHRADIUS_SYSTEM_WORKER_DIR` | `system.workdir` |
| `TOUGHRADIUS_SYSTEM_DEBUG` | `system.debug` |
| `TOUGHRADIUS_WEB_HOST` / `_WEB_PORT` / `_WEB_TLS_PORT` / `_WEB_SECRET` | `web.*` |
| `TOUGHRADIUS_DB_TYPE` / `_DB_HOST` / `_DB_PORT` / `_DB_NAME` / `_DB_USER` / `_DB_PWD` / `_DB_DEBUG` | `database.*` |
| `TOUGHRADIUS_RADIUS_ENABLED` / `_RADIUS_HOST` / `_RADIUS_AUTHPORT` / `_RADIUS_ACCTPORT` / `_RADIUS_DEBUG` | `radiusd.*` |
| `TOUGHRADIUS_RADIUS_RADSEC_PORT` / `_RADIUS_RADSEC_WORKER` / `_RADIUS_RADSEC_CA_CERT` / `_RADIUS_RADSEC_CERT` / `_RADIUS_RADSEC_KEY` | RadSec 配置 |
| `TOUGHRADIUS_LOGGER_MODE` / `_LOGGER_FILE_ENABLE` | `logger.*` |
| `TOUGHRADIUS_RADIUS_POOL` | RADIUS 工作池大小（默认 1024） |

### 命令行参数

| 参数 | 作用 |
| ---- | ---- |
| `-c <文件>` | 指定配置文件 |
| `-initdb` | **删除并重建全部数据表**后退出 |
| `-printcfg` | 以 JSON 打印合并后的配置并退出 |
| `-v` | 打印版本 / 构建时间 / 提交号并退出 |
| `-h` | 帮助 |

RADIUS 运行时配置（EAP 方法、证书、间隔、拒绝延迟等）存储在数据库中，通过
**系统配置** 页面修改——无需重启。见
[管理系统用户手册](./admin-manual.md#系统配置)。

## 数据库

- **SQLite**（默认）—— 纯 Go 驱动，无 CGO，文件位于
  `{workdir}/data/<name>`。适合中小规模部署；备份即拷贝文件。
- **PostgreSQL** —— 设置 `database.type: postgres` 及 host/user/password。
  推荐用于生产规模与高并发计费负载。

结构迁移（GORM `AutoMigrate`）在每次启动时自动执行，升级流程即：停止、
替换二进制、启动。`-initdb` 仅用于首次安装——它**销毁全部数据**。

需要关注的大表：`radius_accounting`（随会话持续增长）与 `radius_online`。
`radius.AccountingHistoryDays` 配置（默认 90，设为 `0` 关闭）定义计费历史的保留
窗口：`@daily` 定时任务会删除超过该天数的**已结束** `radius_accounting` 记录（在线
会话不受影响），并清理连续多个计费中间更新周期未刷新的 `radius_online` 残留行。操作
日志（`sys_opr_log`）一年后自动清理。若数据量很大，仍建议把表增长监控与数据库级归档
纳入自己的运维流程。

## TLS 与证书

三处相互独立的证书使用方：

| 使用方 | 文件 | 说明 |
| ------ | ---- | ---- |
| **RadSec** | `radiusd.radsec_ca_cert` / `radsec_cert` / `radsec_key` | TLS 1.2+；客户端证书**提供则校验**（`VerifyClientCertIfGiven`） |
| **Web HTTPS** | `{workdir}/private/toughradius.tls.crt` + `.key`（固定路径） | 监听 `web.tls_port`；加载失败仅记日志，HTTP 继续运行 |
| **EAP（TLS/PEAP/TTLS）** | 系统配置 → `EapTlsCertFile`、`EapTlsKeyFile`、`EapTlsCaFile`、`EapTlsMinVersion` | 证书/私钥留空即禁用基于 TLS 的 EAP 方法 |

用内置工具一次生成 CA/服务器/客户端全套证书：

```bash
go run ./cmd/certgen -type all -output /var/toughradius/private \
  -server-cn radius.example.com -server-dns radius.example.com \
  -days 3650
# 然后将 radsec_cert/radsec_key（或 EAP 配置）指向生成的文件
```

## 日志

zap 结构化日志。`logger.mode: development` 输出适合人读的控制台格式；
`production` 输出 JSON。文件输出由 `logger.file_enable` + `logger.filename`
控制。RADIUS 日志级别还可在运行时通过 **系统配置 → LogLevel** 调整；
`radiusd.debug: true` 会转储完整报文（生产环境请关闭）。

## 指标

计数器保存在内存中，经由管理仪表盘展示（没有 Prometheus `/metrics` HTTP
端点）。RADIUS 计数器包括：`radus_accept`、`radus_online`/`radus_offline`、
`radus_accounting`、`radus_auth_drop` / `radus_acct_drop`、
`radus_radsec_saturated`，以及按原因细分的拒绝计数——
`radus_reject_passwd_error`、`radus_reject_not_exists`、`radus_reject_expire`、
`radus_reject_disabled`、`radus_reject_limit`、`radus_reject_bind_error`、
`radus_reject_ldap_error`、`radus_reject_unauthorized`、`radus_reject_other`。
`radus_reject_ldap_error` 表示 LDAP/AD 后端无法给出认证答案——例如目录不可达、
TLS/StartTLS 失败、服务账号 bind 失败或 LDAP 配置错误；口令错误仍归入
`radus_reject_passwd_error`。计费请求在入口被丢弃时按
原因细分——`radus_acct_drop_nas`（未知/未授权 NAS）、`radus_acct_drop_username`
（缺少用户名）、`radus_acct_drop_secret`（Request Authenticator 校验失败）——
而 `radus_acct_drop` 作为背压与响应写入丢弃的兜底计数。系统仪表（CPU/内存、进程
CPU/内存）每 30 秒采样一次。

外部监控建议探测服务端口并采集日志文件；进程退出即故障信号（fail-fast
进程模型）。

## 备份与恢复

**系统配置 → 备份** 下载一份 JSON 快照（schema 版本 9.0），包含：节点、NAS
设备、计费策略、用户、系统配置、操作员。服务器侧同时在 `{workdir}/backup/`
留存一份。**恢复** 可回导该文件。

> 快照**不包含**计费历史与在线会话。完整的灾备方案还需备份数据库本身
> （拷贝 SQLite 文件或使用 `pg_dump`）。

## 命令行工具

均位于 `cmd/` 下，以 `go run ./cmd/<工具>` 运行：

| 工具 | 用途 |
| ---- | ---- |
| `radtest` | 迷你 RADIUS 客户端：`auth`、`acct`、`flow`（认证 + 开始 + 结束）。参数：`-server`、`-secret`、`-username`、`-password`、`-calling-station`、`-framed-ip`、`-session-id` |
| `certgen` | 生成 CA / 服务器 / 客户端证书（见上文） |
| `benchmark` | 压力测试：总请求数 `-n`、并发 `-c`、认证/计费模式、CSV 统计输出 |
| `reset-password` | 重置控制台操作员密码：`go run ./cmd/reset-password -c <配置> -u admin -p <新密码>` |
| `demo-seed` | 填充演示用节点/NAS/策略/用户/会话数据 |
| `config-tool` | 校验 / 汇总配置 schema JSON |

## 生产加固清单

- [ ] 修改 `web.secret` 与默认 `admin` 密码。
- [ ] `radiusd.debug: false`、`logger.mode: production`。
- [ ] 用防火墙将 UDP 1812/1813 与 TCP 1816 限制在可信网络内。
- [ ] 跨不可信网络传输 RADIUS 时使用 RadSec（2083）或可信二层/VPN 通道。
- [ ] 每台 NAS 使用独立且强度足够的共享密钥。
- [ ] EAP：优先 `eap-tls`/`eap-peap`/`eap-ttls` 并使用正式证书；MS-CHAPv2
      的注意事项见[安全策略](./security-policy.md)。
- [ ] 制定数据库备份计划（配置快照 + 数据库转储）。
- [ ] 守护进程配置自动重启，并对进程退出告警。
