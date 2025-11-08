# ToughRadius v9 架构说明

## 项目结构

```
toughradius/
├── cmd/                      # 应用程序入口
│   └── toughradius/         # 主程序
│       └── main.go
│
├── internal/                 # 私有应用代码
│   ├── api/                 # API 文档 (Swagger)
│   ├── app/                 # 应用初始化和全局管理
│   │   ├── app.go          # 应用程序实例
│   │   ├── database.go     # 数据库初始化
│   │   ├── jobs.go         # 定时任务
│   │   └── initdb.go       # 数据库初始化脚本
│   ├── domain/              # 领域模型 (统一的数据模型)
│   │   ├── radius.go       # RADIUS 相关模型
│   │   ├── network.go      # 网络设备模型
│   │   └── system.go       # 系统管理模型
│   ├── freeradius/          # FreeRADIUS 集成服务
│   │   ├── freeradius.go   # 业务逻辑
│   │   ├── handlers.go     # REST API 处理器
│   │   └── server.go       # HTTP 服务器
│   ├── handler/             # Web API 处理器 (统一的 Controllers)
│   │   ├── radius.go       # RADIUS 管理 API
│   │   ├── user.go         # 用户管理 API
│   │   ├── node.go         # 网络节点管理
│   │   ├── vpe.go          # VPE 管理
│   │   ├── opr.go          # 操作员管理
│   │   ├── settings.go     # 系统设置
│   │   ├── dashboard.go    # 仪表盘
│   │   └── metrics.go      # 监控指标
│   ├── middleware/          # HTTP 中间件
│   └── webserver/           # Web 服务器
│       └── server.go       # Echo 服务器配置
│
├── pkg/                      # 公共库 (可被外部引用)
│   ├── common/              # 通用工具函数
│   ├── radius/              # RADIUS 协议实现
│   │   ├── server.go       # RADIUS 服务器
│   │   ├── auth_*.go       # 认证逻辑
│   │   ├── acct_*.go       # 计费逻辑
│   │   └── vendors/        # 厂商扩展
│   ├── logger/              # 日志库 (Zap 封装)
│   ├── database/            # 数据库工具
│   ├── utils/               # 工具函数
│   └── [其他工具库]/        # AES, RSA, MFA, SNMP 等
│
├── web/                      # 静态资源
│   ├── static/              # 静态文件 (CSS, JS, 图片)
│   └── templates/           # HTML 模板
│
└── scripts/                  # 构建和部署脚本
    └── build.sh             # 构建脚本
```

## 架构设计原则

### 1. 清晰的职责划分

- **cmd/**: 应用程序入口,负责启动各个服务
- **internal/**: 私有业务逻辑,外部无法引用
- **pkg/**: 可复用的公共库,可被外部项目引用

### 2. 统一的数据模型

所有数据模型统一放在 `internal/domain/`:

- `radius.go` - RADIUS 相关的所有模型
- `network.go` - 网络设备相关模型
- `system.go` - 系统管理相关模型

这样避免了模型分散,便于维护和理解。

### 3. 统一的 Web Controller

所有 HTTP API 处理器统一放在 `internal/handler/`:

- 按功能模块组织文件(radius, user, node, vpe, settings 等)
- 避免了按 domain 分散导致的代码查找困难
- 便于统一管理路由和中间件

### 4. 服务模块化

- `internal/freeradius/` - FreeRADIUS 集成服务
- `internal/webserver/` - Web 管理服务
- `pkg/radius/` - RADIUS 协议核心实现

每个服务都是独立的模块,可以单独测试和维护。

## 依赖关系

```
cmd/toughradius
    ├─> internal/app          (应用初始化)
    ├─> internal/webserver    (Web 服务)
    ├─> internal/freeradius   (FreeRADIUS 服务)
    └─> pkg/radius           (RADIUS 服务)

internal/webserver
    ├─> internal/handler      (API 处理器)
    ├─> internal/domain       (数据模型)
    └─> pkg/*                (公共库)

internal/handler
    ├─> internal/app          (访问数据库等)
    ├─> internal/domain       (使用模型)
    └─> pkg/*                (使用工具库)

pkg/radius
    ├─> internal/domain       (使用模型)
    ├─> internal/app          (访问数据库)
    └─> pkg/logger           (日志)
```

## 优势

1. **结构清晰**: 符合 Go 标准项目布局,易于理解
2. **模型统一**: 所有 domain 模型集中管理
3. **控制器统一**: 所有 HTTP handlers 集中管理
4. **职责明确**: internal vs pkg 界限清晰
5. **易于维护**: 代码位置可预测,便于查找和修改
6. **可测试性**: 模块化设计便于单元测试

## 构建和运行

```bash
# 构建前端
./scripts/build-frontend.sh

# 构建后端（输出 release/toughradius）
./scripts/build-backend.sh

# 运行
./release/toughradius -c toughradius.yml

# 初始化数据库
./release/toughradius -initdb -c toughradius.yml
```

## 与 v8 的主要变化

| 方面       | v8              | v9                              |
| ---------- | --------------- | ------------------------------- |
| Go 版本    | 1.21            | 1.24                            |
| 项目结构   | 扁平化          | golang-standards/project-layout |
| 模型位置   | models/         | internal/domain/                |
| 控制器位置 | controllers/_/_ | internal/handler/               |
| 公共工具   | common/         | pkg/                            |
| API 文档   | docs/           | internal/api/                   |
| Web 服务   | webserver/      | internal/webserver/             |
| FreeRADIUS | freeradius/     | internal/freeradius/            |
| 静态资源   | assets/         | web/                            |

## 下一步优化建议

1. 考虑引入 DDD (领域驱动设计) 更细致地组织 domain 层
2. 增加 service 层处理复杂业务逻辑
3. 引入依赖注入框架 (如 wire) 管理依赖
4. 完善单元测试和集成测试
5. 添加 API 版本管理 (v1, v2)
