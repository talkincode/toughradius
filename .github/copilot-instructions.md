# ToughRADIUS AI Coding Agent Instructions

## 项目概述

ToughRADIUS 是一个用 Go 语言开发的企业级 RADIUS 服务器，支持标准 RADIUS 协议（RFC 2865/2866）、RadSec（RADIUS over TLS）以及 FreeRADIUS REST API 集成。前端使用 React Admin 框架构建管理界面。

## 架构要点

### 核心服务并发模型

`main.go` 使用 `errgroup` 并发启动多个独立服务，任一服务崩溃会导致整个应用退出：

- **Web/Admin API** - Echo 框架，端口 1816（`internal/webserver` + `internal/adminapi`）
- **FreeRADIUS API** - REST 集成服务，端口 1818（`internal/freeradius`）
- **RADIUS Auth** - 认证服务，UDP 1812（`internal/radiusd`）
- **RADIUS Acct** - 计费服务，UDP 1813（`internal/radiusd`）
- **RadSec** - TLS 加密的 RADIUS over TCP，端口 2083（`internal/radiusd`）

### 项目结构模式

遵循 golang-standards/project-layout：

- `internal/` - 私有代码，外部不可导入
  - `domain/` - **统一数据模型**（所有 GORM 模型定义在 `domain/tables.go` 列出）
  - `adminapi/` - 新版管理 API 路由（v9 重构）
  - `radius/` - RADIUS 协议核心实现
  - `freeradius/` - FreeRADIUS REST API 适配层
  - `app/` - 全局应用实例（数据库、配置、定时任务）
- `pkg/` - 可复用公共库（工具函数、加密、Excel 等）
- `web/` - React Admin 前端（TypeScript + Vite）

### 数据库访问模式

**始终**通过 `app.GDB()` 获取 GORM 实例，不要直接注入 DB 连接：

```go
// 正确
user := &domain.RadiusUser{}
app.GDB().Where("username = ?", name).First(user)

// 错误 - 不要这样做
type Service struct { DB *gorm.DB }
```

支持 PostgreSQL（默认）和 SQLite（需要 `CGO_ENABLED=1` 编译）。数据库迁移通过 `app.MigrateDB()` 自动完成。

### 厂商扩展处理

RADIUS 协议支持多厂商特性，通过 `VendorCode` 字段区分：

- Huawei (2011) - `internal/radiusd/vendors/huawei/`
- Mikrotik (14988) - 见 `auth_accept_config.go`
- Cisco (9) / Ikuai (10055) / ZTE (3902) / H3C (25506)

添加新厂商支持时，在 `radius.go` 中定义常量，然后在 `auth_accept_config.go` 和相关处理函数中添加 switch case。

## 关键开发流程

### 构建与运行

**本地开发**（支持 SQLite）：

```bash
CGO_ENABLED=1 go run main.go -c toughradius.yml
```

**生产构建**（PostgreSQL only，静态编译）：

```bash
make build  # 输出到 ./release/toughradius
```

**前端开发**：

```bash
cd web
npm install
npm run dev      # 开发服务器，热重载
npm run build    # 生产构建，输出到 dist/
```

### 数据库初始化

```bash
./toughradius -initdb -c toughradius.yml  # 删除并重建所有表
```

生产环境使用 `MigrateDB(false)` 自动迁移（main.go 中已配置）。

### 测试规范

- RADIUS 协议测试：`internal/radiusd/*_test.go`
- 基准测试：`cmd/benchmark/bmtest.go`（独立工具）
- 前端测试：`web/` 中使用 Playwright

运行测试：

```bash
go test ./...                    # 全部单元测试
go test -bench=. ./internal/radiusd/  # 基准测试
```

## 常见模式与约定

### 错误处理

RADIUS 认证错误使用自定义类型 `AuthError`，携带 metrics 标签：

```go
return NewAuthError(app.MetricsRadiusRejectExpire, "user expire")
```

这些错误会自动记录到 Prometheus metrics（`internal/app/metrics.go`）。

### 配置读取

通过 `app.GApp()` 访问全局配置和设置：

```go
// 读取 RADIUS 配置项
eapMethod := app.GApp().GetSettingsStringValue("radius", "EapMethod")
maxSessions := app.GApp().GetSettingsInt64Value("radius", "MaxSessions")
```

系统配置存储在 `sys_config` 表中，通过 `checkSettings()` 初始化默认值。

### 并发处理

RADIUS 请求处理使用 ants 协程池限制并发：

```go
radiusService.TaskPool.Submit(func() { /* 处理请求 */ })
```

池大小通过环境变量 `TOUGHRADIUS_RADIUS_POOL` 配置（默认 1024）。

### 日志规范

使用 zap 结构化日志，**始终**添加 namespace：

```go
zap.L().Error("update user failed",
    zap.Error(err),
    zap.String("namespace", "radius"))
```

### Admin API 路由注册

新增管理 API 时，在 `internal/adminapi/` 创建文件并在 `adminapi.go` 的 `Init()` 中注册：

```go
// users.go
func registerUserRoutes() {
    // 路由定义
}

// adminapi.go
func Init() {
    registerUserRoutes()  // 添加这一行
}
```

## 关键依赖与集成

- **Echo v4** - Web 框架，中间件配置在 `internal/webserver/server.go`
- **GORM** - ORM，自动迁移通过 `domain.Tables` 列表控制
- **layeh.com/radius** - RADIUS 协议库，不要与其他 RADIUS 包混用
- **React Admin 5.0** - 前端框架，REST 数据提供者在 `web/src/dataProvider.ts`

FreeRADIUS 集成通过 HTTP REST 接口（端口 1818），FreeRADIUS 配置示例在 `assets/freeradius/`。



