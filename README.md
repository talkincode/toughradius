Welcome to the TOUGHRADIUS project!

     _____   _____   _   _   _____   _   _   _____        ___   _____   _   _   _   _____
    |_   _| /  _  \ | | | | /  ___| | | | | |  _  \      /   | |  _  \ | | | | | | /  ___/
      | |   | | | | | | | | | |     | |_| | | |_| |     / /| | | | | | | | | | | | | |___
      | |   | | | | | | | | | |  _  |  _  | |  _  /    / / | | | | | | | | | | | | \___  \
      | |   | |_| | | |_| | | |_| | | | | | | | \ \   / /  | | | |_| | | | | |_| |  ___| |
      |_|   \_____/ \_____/ \_____/ |_| |_| |_|  \_\ /_/   |_| |_____/ |_| \_____/ /_____/

# TOUGHRADIUS

[![License](https://img.shields.io/github/license/talkincode/toughradius)](https://github.com/talkincode/toughradius/blob/main/LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/talkincode/toughradius)](go.mod)
[![Release](https://img.shields.io/github/v/release/talkincode/toughradius)](https://github.com/talkincode/toughradius/releases)

一个功能强大、开源的 RADIUS 服务器，专为 ISP、企业网络和运营商设计。支持标准 RADIUS 协议、RadSec（RADIUS over TLS）以及现代化的 Web 管理界面。

## ✨ 核心特性

### RADIUS 协议支持

- 🔐 **标准 RADIUS** - 完整支持 RFC 2865/2866 认证和计费协议
- 🔒 **RadSec** - TLS 加密的 RADIUS over TCP（RFC 6614）
- 🌐 **多厂商支持** - 兼容 Cisco、Mikrotik、华为等主流网络设备
- ⚡ **高性能** - 基于 Go 语言构建，支持高并发处理

### 管理功能

- 📊 **React Admin 界面** - 现代化的 Web 管理后台
- 👥 **用户管理** - 完整的用户账号、配置文件（Profile）管理
- 📈 **实时监控** - 在线会话监控、计费记录查询
- 🔍 **日志审计** - 详细的认证和计费日志

### 集成能力

- 🔄 **FreeRADIUS 兼容** - 支持 FreeRADIUS REST API 集成
- 💾 **多数据库支持** - PostgreSQL、SQLite
- 🔌 **灵活扩展** - 支持自定义认证、计费逻辑

## 🚀 快速开始

### 前置要求

- Go 1.24+ (用于从源码构建)
- PostgreSQL 或 SQLite
- Node.js 18+ (用于前端开发)

### 安装

#### 1. 从源码构建

```bash
# 克隆仓库
git clone https://github.com/talkincode/toughradius.git
cd toughradius

# 构建前端
cd web
npm install
npm run build
cd ..

# 构建后端
go build -o toughradius main.go
```

#### 2. 使用预编译版本

从 [Releases](https://github.com/talkincode/toughradius/releases) 页面下载最新版本。

### 配置

1. 复制配置文件模板：

```bash
cp toughradius.yml toughradius.prod.yml
```

2. 编辑 `toughradius.prod.yml` 配置文件：

```yaml
system:
  appid: ToughRADIUS
  location: Asia/Shanghai
  workdir: ./rundata

database:
  type: sqlite # 或 postgres
  name: toughradius.db
  # PostgreSQL 配置
  # host: localhost
  # port: 5432
  # user: toughradius
  # passwd: your_password

radiusd:
  enabled: true
  host: 0.0.0.0
  auth_port: 1812 # RADIUS 认证端口
  acct_port: 1813 # RADIUS 计费端口
  radsec_port: 2083 # RadSec 端口

web:
  host: 0.0.0.0
  port: 1816 # Web 管理界面端口
```

### 运行

```bash
# 初始化数据库
./toughradius -initdb -c toughradius.prod.yml

# 启动服务
./toughradius -c toughradius.prod.yml
```

访问 Web 管理界面：<http://localhost:1816>

默认管理员账号：

- 用户名：admin
- 密码：请查看初始化日志输出

## 📖 文档

- [架构说明](docs/v9-architecture.md) - v9 版本架构设计
- [React Admin 重构](docs/react-admin-refactor.md) - 前端管理界面说明
- [SQLite 支持](docs/sqlite-support.md) - SQLite 数据库配置
- [环境变量](docs/environment-variables.md) - 环境变量配置说明

## 🏗️ 项目结构

```text
toughradius/
├── cmd/             # 应用程序入口
├── internal/        # 私有应用代码
│   ├── adminapi/   # Admin API（新版）
│   ├── radius/     # RADIUS 服务核心
│   ├── freeradius/ # FreeRADIUS 集成
│   └── webserver/  # Web 服务器
├── pkg/            # 公共库
├── web/            # React Admin 前端
├── migrations/     # 数据库迁移
└── docs/           # 文档
```

## 🔧 开发

### 后端开发

```bash
# 运行测试
go test ./...

# 运行基准测试
go test -bench=. ./internal/radius/

# 启动开发模式
go run main.go -c toughradius.yml
```

### 前端开发

```bash
cd web
npm install
npm run dev       # 开发服务器
npm run build     # 生产构建
npm run lint      # 代码检查
```

## 🤝 贡献

我们欢迎各种形式的贡献，包括但不限于：

- 🐛 提交 Bug 报告和功能请求
- 📝 改进文档
- 💻 提交代码补丁和新特性
- 🌍 帮助翻译


## 📜 许可证

本项目采用 [MIT License](LICENSE) 开源协议。

## 🔗 相关链接

- [官方网站](https://www.toughradius.net/)
- [在线文档](https://github.com/talkincode/toughradius/wiki)
- [RadSec RFC 6614](https://tools.ietf.org/html/rfc6614)
- [RADIUS RFC 2865](https://tools.ietf.org/html/rfc2865)
- [Mikrotik RADIUS 配置](https://wiki.mikrotik.com/wiki/Manual:RADIUS_Client)

## 💎 赞助商

感谢 [JetBrains](https://jb.gg/OpenSourceSupport) 对本项目的支持！

![JetBrains Logo](https://resources.jetbrains.com/storage/products/company/brand/logos/jb_beam.svg)
