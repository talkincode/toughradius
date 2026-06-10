# 快速开始

> English version: [Quick Start](../en/quickstart.md)

本章带你从零搭建一个可用的 RADIUS 服务器并创建一个测试用户，然后介绍如何调试
服务器行为。默认端口：管理界面 `1816`、RADIUS 认证 UDP `1812`、计费 UDP
`1813`、RadSec TCP `2083`。

## 1. 安装

三种方式任选其一。

### 方式 A —— 预编译二进制

从 [GitHub Releases](https://github.com/talkincode/toughradius/releases)
页面下载对应平台的二进制（`toughradius_linux_amd64`、`toughradius_linux_arm64`、
`toughradius_darwin_arm64`、`toughradius_windows_amd64.exe` 等），然后：

```bash
chmod +x toughradius_linux_amd64
sudo mv toughradius_linux_amd64 /usr/local/bin/toughradius
```

### 方式 B —— Docker

```bash
docker pull talkincode/toughradius:latest

docker run -d --name toughradius \
  -p 1816:1816 -p 1812:1812/udp -p 1813:1813/udp -p 2083:2083 \
  -v toughradius-data:/var/toughradius \
  talkincode/toughradius:latest -c /etc/toughradius.yml
```

镜像暴露 `1816/tcp`、`1812/udp`、`1813/udp`、`2083/tcp`。请将卷挂载到
`/var/toughradius`（默认工作目录），使 SQLite 数据库、日志与证书在容器重启后
得以保留。

### 方式 C —— 源码构建

需要 Go 1.25+ 与 Node.js 20+（React Admin 前端会被嵌入二进制）：

```bash
git clone https://github.com/talkincode/toughradius.git
cd toughradius
make build          # 先构建 web/ 再编译 Go 二进制 → release/toughradius
```

仅后端开发（SQLite 默认配置）：

```bash
make runs           # CGO_ENABLED=0 go run main.go -c toughradius.yml
make runf           # 前端开发服务器 http://localhost:3000/admin
```

## 2. 配置

ToughRADIUS 按以下顺序查找配置：`-c <文件>` 参数、`./toughradius.yml`、
`/etc/toughradius.yml`、内置默认值。环境变量优先于配置文件
（见[运维指南](./ops-guide.md#环境变量)）。

一份精简的生产风格配置：

```yaml
system:
  appid: ToughRADIUS
  location: Asia/Shanghai
  workdir: /var/toughradius     # 数据/日志/证书都在这里
  debug: false

web:
  host: 0.0.0.0
  port: 1816
  secret: change-me-to-a-long-random-string   # JWT 签名密钥

database:
  type: sqlite                  # 或 postgres（需配 host/port/user/passwd）
  name: toughradius.db          # 存放于 {workdir}/data/ 下

radiusd:
  enabled: true
  host: 0.0.0.0
  auth_port: 1812
  acct_port: 1813
  radsec_port: 2083
  debug: true                   # 输出完整报文转储；生产环境建议关闭

logger:
  mode: production
  file_enable: true
  filename: /var/toughradius/toughradius.log
```

> **务必修改 `web.secret`**，它用于签发管理端登录令牌。首次登录后请立即修改
> 默认管理员密码。

## 3. 初始化数据库并启动

```bash
# 仅第一次执行 —— 会删除并重建全部数据表
toughradius -initdb -c /etc/toughradius.yml

# 启动服务
toughradius -c /etc/toughradius.yml
```

`-initdb` 是破坏性操作；后续升级直接启动即可——结构迁移在启动时自动完成。
其他参数：`-v` 打印版本，`-printcfg` 以 JSON 打印合并后的配置。

## 4. 登录管理界面

打开 `http://<服务器>:1816`。默认管理员：

- 用户名：`admin`
- 密码：`toughradius`

请立即在 **账户设置** 中修改密码；忘记密码可用 `cmd/reset-password` 重置
（见[常见问题](./faq.md)）。

## 5. 登记 NAS 并创建用户

1. **网络节点 → 新建** —— 创建一个节点（逻辑分组），如 `default`。
2. **NAS 设备 → 新建** —— 登记你的网络设备：
   - *IP 地址*：设备发出 RADIUS 报文的源地址。
   - *密钥*：共享密钥，如 `testing123`。
   - *厂商代码*：选择设备厂商（标准 / Cisco / 华为 / MikroTik / H3C / 中兴 /
     爱快）——它决定下发哪些厂商私有属性，见[厂商对接指南](./vendor-guide.md)。
   - *CoA 端口*：除非设备使用其他端口，保持 `3799`。
3. **计费策略 → 新建** —— 例如 `100M`：并发数 `1`、上行 `51200` Kbps、
   下行 `102400` Kbps。
4. **RADIUS 用户 → 新建** —— 用户名 `test1`、密码 `111111`，选择策略并设置
   过期时间。

## 6. 用 radtest 验证

仓库自带一个小型 RADIUS 客户端（示例即默认值）：

```bash
go run ./cmd/radtest auth \
  -server 127.0.0.1 -secret testing123 \
  -username test1 -password 111111

go run ./cmd/radtest flow ...   # 一次跑完 认证 + 计费开始 + 计费结束
```

成功时会打印 `Access-Accept` 及返回的属性。`flow` 模式的会话会出现在
**在线会话** 中，仪表盘计数随之增长。

## 7. 调试

| 需求 | 方法 |
| ---- | ---- |
| 完整 RADIUS 报文转储 | YAML 中 `radiusd.debug: true`（或环境变量 `TOUGHRADIUS_RADIUS_DEBUG=true`），也可在运行时将 **系统配置 → RADIUS → 日志级别** 设为 `debug` |
| 日志文件位置 | `logger.filename`，默认 `{workdir}/toughradius.log`；`logger.mode: development` 输出适合人读的控制台格式 |
| 用户为何被拒绝 | 拒绝原因按类别计数（密码错误、已过期、MAC 绑定不符等）并显示在仪表盘；细节见日志 |
| 查看生效配置 | `toughradius -printcfg -c <文件>` |
| 压力测试 | `go run ./cmd/benchmark` —— 见[运维指南](./ops-guide.md#命令行工具) |

## 下一步

- [厂商对接指南](./vendor-guide.md) —— 配置 Cisco、华为、MikroTik、H3C、
  中兴、爱快及标准设备。
- [管理系统用户手册](./admin-manual.md) —— 管理界面每个页面的说明。
- [运维指南](./ops-guide.md) —— 生产部署、TLS、EAP 证书、备份与监控。
