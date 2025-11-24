# ToughRADIUS v9 架构重构任务清单

**创建日期**: 2025-11-20  
**基于分支**: v9dev  
**架构审查**: Legacy-aware Architecture Reviewer

---

## 📋 任务概览

本文档记录了 ToughRADIUS v9 架构审查中发现的需要改进的任务项。所有任务按优先级和影响范围分类，每个任务都包含详细的实施建议和验收标准。

### 优先级说明

- **P0 (Critical)**: 影响系统稳定性或安全性，需立即处理
- **P1 (High)**: 影响可维护性和扩展性，建议尽快处理
- **P2 (Medium)**: 改善代码质量，可纳入常规迭代
- **P3 (Low)**: 优化性改进，长期规划

---

## 🎯 高优先级重构任务 (P1)

### 任务 1: 厂商代码管理中心化

**优先级**: P1 (High)  
**预计工作量**: 3 个 PR，约 5-8 工作日  
**影响范围**: `internal/radiusd/`, `internal/radiusd/plugins/`

#### 问题描述

厂商代码（VendorCode）分散在多个位置：

- `internal/radiusd/radius.go` - 常量定义
- `internal/radiusd/plugins/auth/enhancers/vendor_helpers.go` - 重复常量
- `internal/radiusd/plugins/vendorparsers/parsers/*.go` - 各厂商实现

这导致：

- 添加新厂商需修改 4-5 个文件
- 厂商 ID 不一致的风险
- 无法动态注册或移除厂商
- 测试时无法注入 mock 厂商

#### 解决方案

创建统一的 `VendorRegistry` 管理所有厂商元数据：

```go
// internal/radiusd/vendors/registry.go
package vendors

type VendorInfo struct {
    Code        string
    Name        string
    Description string
    Parser      vendorparsers.VendorParser
    Builder     vendorparsers.VendorResponseBuilder
}

type VendorRegistry struct {
    mu      sync.RWMutex
    vendors map[string]*VendorInfo
}

func NewVendorRegistry() *VendorRegistry {
    return &VendorRegistry{
        vendors: make(map[string]*VendorInfo),
    }
}

func (r *VendorRegistry) Register(info *VendorInfo) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    if _, exists := r.vendors[info.Code]; exists {
        return fmt.Errorf("vendor code %s already registered", info.Code)
    }

    r.vendors[info.Code] = info
    return nil
}

func (r *VendorRegistry) Get(code string) (*VendorInfo, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    info, ok := r.vendors[code]
    return info, ok
}

func (r *VendorRegistry) List() []*VendorInfo {
    r.mu.RLock()
    defer r.mu.RUnlock()

    list := make([]*VendorInfo, 0, len(r.vendors))
    for _, info := range r.vendors {
        list = append(list, info)
    }
    return list
}
```

#### 实施步骤

**PR #1: 创建 VendorRegistry 基础设施**

- [x] 创建 `internal/radiusd/vendors/registry.go`
- [x] 定义 `VendorInfo` 和 `VendorRegistry` 结构
- [x] 添加单元测试覆盖所有方法
- [x] 验证并发安全性

**PR #2: 迁移现有厂商到 Registry**

- [x] 将 `radius.go` 中的常量迁移到 Registry
- [x] 更新 `parsers/init.go` 使用 Registry 注册
- [x] 修改 `vendor_parse.go` 从 Registry 获取解析器
- [x] 确保所有测试通过

**PR #3: 清理重复代码**

- [x] 删除 `vendor_helpers.go` 中的重复常量
- [x] 更新所有引用点使用 Registry
- [x] 添加 Registry 使用示例到文档（见 `docs/vendor-registry.md`）
- [x] 运行完整测试套件验证

开发计划：

1. **代码清点与风险评估**

   - 使用 `git grep "Vendor" internal/radiusd/plugins/auth/enhancers` 统计所有仍然存在的厂商常量和辅助函数。
   - 评估 `vendor_helpers.go` 中剩余逻辑是否仍有业务价值（目前包含 `IsHuaweiDevice` 等辅助函数），若仍需保留则规划迁移路径而非直接删除文件。
   - 输出一份简要的“剩余重复项”列表，记录常量名称、文件位置、预期 Registry 替代项。

2. **重构方案设计**

   - 若辅助函数只依赖厂商常量，则将常量替换为 `vendors.CodeHuawei` 等 Registry 暴露的常量，并将通用辅助函数移动到更合适的位置（如 `internal/radiusd/vendors/helpers.go`）。
   - 对复杂函数（例如组合判定逻辑）编写最小可重现的单元测试，以防止删除常量时引入回归。
   - 明确 Registry 使用示例的落脚点：优先在 `docs/` 或 `internal/radiusd/vendors/README.md` 中添加“如何注册/查询厂商”的示例代码。

3. **实施步骤**

   - 先创建新的辅助模块/方法并切换调用点，确保所有引用处改为通过 Registry 访问厂商信息。
   - 所有改动完成后，再删除原有重复常量（必要时整个 `vendor_helpers.go` 文件），并更新 `go test ./internal/radiusd/...` 以验证。
   - 在 PR 描述和文档中附上 Registry 示例片段，满足验收标准。

4. **验证与交付**
   - 运行 `go test ./internal/radiusd/vendors/... ./internal/radiusd/plugins/...`，确保增强器与解析器仍然通过测试。
   - 由代码审查者确认不再存在重复常量，同时 Registry 示例文档可指导添加新厂商的流程。

#### 验收标准

- [ ] 所有厂商代码集中在 `VendorRegistry` 中
- [ ] 单元测试覆盖率 ≥ 90%
- [ ] 无代码重复（厂商代码常量）
- [ ] 支持运行时注册/注销厂商
- [ ] 文档更新说明如何添加新厂商

#### 后续优化

- 支持从配置文件加载厂商信息
- 提供厂商管理 API 接口
- 实现厂商插件热加载机制

---

### 任务 2: 应用上下文接口隔离

**优先级**: P1 (High)  
**预计工作量**: 3 个 PR，约 6-10 工作日  
**影响范围**: `internal/app/`, `internal/radiusd/`, `internal/adminapi/`, `internal/webserver/`

#### 问题描述

`AppContext` 接口捆绑了 5 个 provider 接口：

```go
type AppContext interface {
    DBProvider
    ConfigProvider
    SettingsProvider
    SchedulerProvider
    ConfigManagerProvider
    // ...
}
```

这违反了接口隔离原则（ISP），导致：

- 服务依赖过多不需要的功能
- 测试必须 mock 整个 `AppContext`
- 添加新功能会污染所有消费者
- 难以理解服务的真实依赖

#### 解决方案

让服务只依赖它们实际需要的接口：

```go
// Before (当前)
func NewRadiusService(appCtx app.AppContext) *RadiusService

// After (建议)
func NewRadiusService(db DBProvider, cfg ConfigProvider) *RadiusService
```

#### 实施步骤

**PR #1: 审计服务依赖**

- [ ] 分析 `RadiusService` 实际使用的 provider
- [ ] 分析 `AdminServer` 实际使用的 provider
- [ ] 分析所有 API handlers 的依赖
- [ ] 创建依赖矩阵文档

依赖矩阵示例：

```
Service            | DBProvider | ConfigProvider | SettingsProvider | SchedulerProvider
-------------------|------------|----------------|------------------|------------------
RadiusService      | ✓          | ✓              | ✗                | ✗
AdminServer        | ✗          | ✓              | ✓                | ✗
AccountingHandler  | ✓          | ✗              | ✗                | ✗
```

**PR #2: 重构服务构造函数**

- [ ] 更新 `RadiusService` 构造函数接受具体 providers
- [ ] 更新 `AdminServer` 构造函数
- [ ] 更新所有 handler/controller 构造函数
- [ ] 调整 `main.go` 中的依赖注入

```go
// internal/radiusd/radius.go
func NewRadiusService(
    db app.DBProvider,
    cfg app.ConfigProvider,
) *RadiusService {
    // ...
}

// main.go
radiusService := radiusd.NewRadiusService(
    application, // 仍实现 DBProvider
    application, // 仍实现 ConfigProvider
)
```

**PR #3: 更新测试和文档**

- [ ] 重构所有单元测试使用具体 provider mocks
- [ ] 更新 `test_helpers.go` 提供轻量级 mocks
- [ ] 更新架构文档说明新的依赖模式
- [ ] 添加依赖注入最佳实践文档

#### 验收标准

- [ ] 每个服务只依赖必需的 provider 接口
- [ ] 测试可以只 mock 实际使用的 providers
- [ ] `AppContext` 保留为便利接口（可选使用）
- [ ] 所有单元测试通过
- [ ] 文档清晰说明依赖注入模式

#### 后续优化

- 引入依赖注入框架（如 wire 或 fx）
- 实现 provider 的生命周期管理
- 支持 provider 的热替换（用于 A/B 测试）

---

### 任务 3: 数据库迁移追踪系统

**优先级**: P1 (High)  
**预计工作量**: 3 个 PR，约 5-7 工作日  
**影响范围**: `internal/app/`, `scripts/`, `docs/`

#### 问题描述

当前 `MigrateDB(track bool)` 使用 GORM 的 `AutoMigrate`：

```go
func (a *Application) MigrateDB(track bool) error {
    return a.gormDB.Migrator().AutoMigrate(domain.Tables...)
}
```

存在的问题：

- 无法追踪 schema 版本
- 无法回滚失败的迁移
- 多人开发时 schema 冲突难以发现
- 生产环境迁移风险高
- 无法查看迁移历史

#### 解决方案

引入版本化迁移系统，使用 `golang-migrate/migrate`：

```
internal/app/migrations/
├── 000001_initial_schema.up.sql
├── 000001_initial_schema.down.sql
├── 000002_add_vendor_tables.up.sql
├── 000002_add_vendor_tables.down.sql
├── 000003_add_plugin_config.up.sql
├── 000003_add_plugin_config.down.sql
└── README.md
```

#### 实施步骤

**PR #1: 添加迁移框架**

- [ ] 添加 `github.com/golang-migrate/migrate/v4` 依赖
- [ ] 创建 `internal/app/migrations/` 目录结构
- [ ] 实现 `MigrationManager` 封装 migrate 库
- [ ] 添加迁移版本追踪表 `schema_migrations`

```go
// internal/app/migration_manager.go
package app

import (
    "database/sql"
    "fmt"

    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/postgres"
    "github.com/golang-migrate/migrate/v4/database/sqlite3"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

type MigrationManager struct {
    db      *sql.DB
    migrate *migrate.Migrate
}

func NewMigrationManager(db *sql.DB, dbType, migrationsPath string) (*MigrationManager, error) {
    var driver migrate.DatabaseDriver
    var err error

    switch dbType {
    case "postgres":
        driver, err = postgres.WithInstance(db, &postgres.Config{})
    case "sqlite":
        driver, err = sqlite3.WithInstance(db, &sqlite3.Config{})
    default:
        return nil, fmt.Errorf("unsupported database type: %s", dbType)
    }

    if err != nil {
        return nil, err
    }

    m, err := migrate.NewWithDatabaseInstance(
        fmt.Sprintf("file://%s", migrationsPath),
        dbType,
        driver,
    )

    if err != nil {
        return nil, err
    }

    return &MigrationManager{
        db:      db,
        migrate: m,
    }, nil
}

func (m *MigrationManager) Up() error {
    return m.migrate.Up()
}

func (m *MigrationManager) Down() error {
    return m.migrate.Down()
}

func (m *MigrationManager) Version() (uint, bool, error) {
    return m.migrate.Version()
}

func (m *MigrationManager) Steps(n int) error {
    return m.migrate.Steps(n)
}
```

**PR #2: 生成初始迁移**

- [ ] 从当前 `domain.Tables` 生成 `000001_initial_schema.up.sql`
- [ ] 编写对应的 `.down.sql` 回滚脚本
- [ ] 测试 up/down 迁移的幂等性
- [ ] 添加迁移创建工具脚本

```bash
# scripts/create-migration.sh
#!/bin/bash
NAME=$1
TIMESTAMP=$(date +%Y%m%d%H%M%S)
UP_FILE="internal/app/migrations/${TIMESTAMP}_${NAME}.up.sql"
DOWN_FILE="internal/app/migrations/${TIMESTAMP}_${NAME}.down.sql"

echo "-- ${NAME} up migration" > $UP_FILE
echo "-- ${NAME} down migration" > $DOWN_FILE

echo "Created:"
echo "  $UP_FILE"
echo "  $DOWN_FILE"
```

**PR #3: 集成到应用生命周期**

- [ ] 替换 `MigrateDB()` 使用 `MigrationManager`
- [ ] 保留 `AutoMigrate` 仅用于开发模式
- [ ] 添加 `-migrate-up/-migrate-down` 命令行参数
- [ ] 更新 `scripts/init-db.sh` 使用新迁移系统
- [ ] 更新文档说明迁移流程

```go
// internal/app/app.go
func (a *Application) MigrateDB(track bool) error {
    sqlDB, err := a.gormDB.DB()
    if err != nil {
        return err
    }

    migrationsPath := path.Join(a.appConfig.System.Workdir, "migrations")

    mgr, err := NewMigrationManager(sqlDB, a.appConfig.Database.Type, migrationsPath)
    if err != nil {
        return err
    }

    if err := mgr.Up(); err != nil && err != migrate.ErrNoChange {
        return err
    }

    version, dirty, err := mgr.Version()
    if err != nil {
        return err
    }

    zap.S().Infof("Database migrated to version %d (dirty: %v)", version, dirty)
    return nil
}
```

#### 验收标准

- [ ] 所有 schema 变更使用版本化迁移
- [ ] 支持 up/down 迁移
- [ ] `schema_migrations` 表正确追踪版本
- [ ] 提供迁移创建工具
- [ ] 文档说明迁移工作流程
- [ ] 测试覆盖迁移失败场景

#### 后续优化

- 添加迁移版本 API 端点 `/api/v1/database/version`
- 实现迁移状态监控
- 支持分布式环境的迁移锁
- 添加迁移预检查（dry-run）

---

### 任务 4: RADIUS 协议库抽象层

**优先级**: P1 (High)  
**预计工作量**: 5 个 PR，约 10-15 工作日  
**影响范围**: `internal/radiusd/`, `internal/radiusd/plugins/`

#### 问题描述

当前代码直接依赖 `layeh.com/radius` 库，分散在 30+ 文件中：

```go
import "layeh.com/radius"
import "layeh.com/radius/rfc2865"

// 直接使用 radius.Packet
packet := radius.New(radius.CodeAccessAccept, secret)
rfc2865.UserName_SetString(packet, username)
```

存在的问题：

- 如果库被废弃或需要替换，需重写 50+ 函数
- 无法 A/B 测试不同的 RADIUS 实现
- 测试时必须导入真实 RADIUS 库
- 与 RADIUS 协议细节耦合过紧

#### 解决方案

创建抽象层隔离 RADIUS 协议实现：

```go
// internal/radiusd/protocol/interfaces.go
package protocol

type PacketReader interface {
    GetUsername() (string, error)
    GetPassword() (string, error)
    GetAttribute(attrType byte) ([]byte, bool)
    GetVendorAttribute(vendorID uint32, attrType byte) ([]byte, bool)
    GetCallingStationID() (string, error)
    GetNASIdentifier() (string, error)
}

type PacketWriter interface {
    SetAttribute(attrType byte, value []byte) error
    AddAttribute(attrType byte, value []byte) error
    AddVendorAttribute(vendorID uint32, attrType byte, value []byte) error
    SetReplyMessage(message string) error
}

type Packet interface {
    PacketReader
    PacketWriter
    Code() byte
    Identifier() byte
}

type PacketFactory interface {
    NewAccessAccept(secret []byte) Packet
    NewAccessReject(secret []byte) Packet
    NewAccessChallenge(secret []byte) Packet
    NewAccountingResponse(secret []byte) Packet
}
```

#### 实施步骤

**PR #1: 创建协议抽象接口**

- [ ] 创建 `internal/radiusd/protocol/` 包
- [ ] 定义 `PacketReader/Writer/Packet/Factory` 接口
- [ ] 添加接口文档和使用示例
- [ ] 创建接口契约测试

**PR #2: 实现 layeh 适配器**

- [ ] 创建 `protocol/layeh/adapter.go`
- [ ] 实现 `LayehPacket` 包装 `*radius.Packet`
- [ ] 实现 `LayehPacketFactory`
- [ ] 运行契约测试验证实现

```go
// internal/radiusd/protocol/layeh/adapter.go
package layeh

import (
    "github.com/talkincode/toughradius/v9/internal/radiusd/protocol"
    "layeh.com/radius"
    "layeh.com/radius/rfc2865"
)

type LayehPacket struct {
    packet *radius.Packet
}

func (p *LayehPacket) GetUsername() (string, error) {
    username := rfc2865.UserName_GetString(p.packet)
    if username == "" {
        return "", protocol.ErrAttributeNotFound
    }
    return username, nil
}

func (p *LayehPacket) GetPassword() (string, error) {
    password := rfc2865.UserPassword_GetString(p.packet)
    if password == "" {
        return "", protocol.ErrAttributeNotFound
    }
    return password, nil
}

// ... 实现其他方法

type LayehPacketFactory struct {
    secret []byte
}

func NewLayehPacketFactory(secret []byte) *LayehPacketFactory {
    return &LayehPacketFactory{secret: secret}
}

func (f *LayehPacketFactory) NewAccessAccept(secret []byte) protocol.Packet {
    return &LayehPacket{
        packet: radius.New(radius.CodeAccessAccept, secret),
    }
}

// ... 实现其他工厂方法
```

**PR #3-4: 渐进式迁移核心模块**

- [ ] PR #3: 重构 `radius_auth.go` 使用抽象接口
- [ ] PR #4: 重构 `vendor_parse.go` 使用抽象接口
- [ ] 更新相关测试使用 mock Packet
- [ ] 确保所有测试通过

**PR #5: 迁移插件系统**

- [ ] 重构 `plugins/vendorparsers` 使用抽象接口
- [ ] 重构 `plugins/auth` 使用抽象接口
- [ ] 更新所有插件测试
- [ ] 清理所有 `layeh.com/radius` 直接引用

#### 验收标准

- [ ] 核心代码不直接导入 `layeh.com/radius`
- [ ] 所有 RADIUS 操作通过抽象接口
- [ ] 测试可以使用 mock Packet
- [ ] 适配器层有 100% 测试覆盖
- [ ] 文档说明如何切换 RADIUS 实现

#### 后续优化

- 实现基于 `goradius.org` 的替代适配器
- 添加性能基准测试对比不同实现
- 支持运行时切换 RADIUS 实现

---

## 🔧 中优先级改进任务 (P2)

### 任务 5: 厂商解析器配置外部化

**优先级**: P2 (Medium)  
**预计工作量**: 3 个 PR，约 4-6 工作日  
**影响范围**: `internal/radiusd/plugins/vendorparsers/`, `config/`

#### 问题描述

当前厂商解析器通过 `init()` 硬编码注册：

```go
// parsers/init.go
func init() {
    registry.RegisterVendorParser(&DefaultParser{})
    registry.RegisterVendorParser(&HuaweiParser{})
    registry.RegisterVendorParser(&H3CParser{})
    registry.RegisterVendorParser(&ZTEParser{})
}
```

限制：

- 添加新厂商必须重新编译
- 无法热修复厂商特定 bug
- 不同部署无法使用不同厂商配置
- 测试时无法注入测试厂商

#### 解决方案

支持从外部配置文件加载厂商信息：

```yaml
# config/vendors/huawei.yaml
vendor_code: "2011"
vendor_name: "Huawei"
description: "Huawei RADIUS vendor support"
parser_type: "builtin" # or "plugin", "script"
enabled: true

# Attribute mapping configuration
attributes:
  mac_address:
    source: "Calling-Station-Id"
    format: "colon-separated" # aa:bb:cc:dd:ee:ff
  vlan_id:
    source: "NAS-Port-Id"
    regex: "vlanid=(\\d+)"

# Response building configuration
response:
  bandwidth_unit: "kbps" # Huawei expects Kbps
  bandwidth_multiplier: 1024
```

#### 实施步骤

**PR #1: 定义配置 Schema**

- [ ] 创建 `config/vendors/` 目录
- [ ] 定义 `VendorConfig` 结构体
- [ ] 实现配置文件加载器
- [ ] 添加配置验证逻辑

```go
// internal/radiusd/vendors/config.go
package vendors

type VendorConfig struct {
    VendorCode  string                 `yaml:"vendor_code"`
    VendorName  string                 `yaml:"vendor_name"`
    Description string                 `yaml:"description"`
    ParserType  string                 `yaml:"parser_type"`  // builtin, plugin, script
    Enabled     bool                   `yaml:"enabled"`
    Attributes  map[string]AttrConfig  `yaml:"attributes"`
    Response    ResponseConfig         `yaml:"response"`
}

type AttrConfig struct {
    Source   string `yaml:"source"`
    Format   string `yaml:"format"`
    Regex    string `yaml:"regex"`
}

type ResponseConfig struct {
    BandwidthUnit       string `yaml:"bandwidth_unit"`
    BandwidthMultiplier int    `yaml:"bandwidth_multiplier"`
}

type VendorConfigLoader struct {
    configDir string
}

func (l *VendorConfigLoader) Load() ([]*VendorConfig, error) {
    // 扫描 config/vendors/ 目录
    // 解析所有 .yaml 文件
    // 验证配置
    // 返回配置列表
}
```

**PR #2: 实现配置驱动的解析器**

- [ ] 创建 `ConfigurableParser` 实现
- [ ] 支持基于配置的属性提取
- [ ] 添加配置重载机制
- [ ] 测试配置变更热加载

```go
// internal/radiusd/plugins/vendorparsers/configurable_parser.go
type ConfigurableParser struct {
    config *vendors.VendorConfig
}

func NewConfigurableParser(config *vendors.VendorConfig) *ConfigurableParser {
    return &ConfigurableParser{config: config}
}

func (p *ConfigurableParser) Parse(r *radius.Request) (*VendorRequest, error) {
    vr := &VendorRequest{}

    // 根据配置提取 MAC 地址
    if macCfg, ok := p.config.Attributes["mac_address"]; ok {
        // 从配置的 source 属性提取
        // 按配置的 format 格式化
    }

    // 根据配置提取 VLAN
    if vlanCfg, ok := p.config.Attributes["vlan_id"]; ok {
        // 使用配置的 regex 提取
    }

    return vr, nil
}
```

**PR #3: 集成并向后兼容**

- [ ] 保留现有 `init()` 注册作为 fallback
- [ ] 优先使用配置文件中的厂商
- [ ] 添加 `-reload-vendors` 命令行选项
- [ ] 更新文档说明配置方式

#### 验收标准

- [ ] 支持从 YAML 配置加载厂商
- [ ] 配置变更可热加载
- [ ] 向后兼容现有硬编码厂商
- [ ] 配置验证防止无效配置
- [ ] 文档说明配置格式和选项

#### 后续优化

- 支持厂商配置的版本管理
- 提供配置校验 CLI 工具
- 实现配置变更审计日志

---

### 任务 6: Repository 契约测试

**优先级**: P2 (Medium)  
**预计工作量**: 3 个 PR，约 5-7 工作日  
**影响范围**: `internal/radiusd/repository/`

#### 问题描述

当前 Repository 接口缺少契约测试：

- 如果有人实现新的 Repository（Redis, MongoDB），无测试验证
- Repository 行为规范只在类型签名中，未形成可执行规范
- Bug 只能在集成测试中发现

#### 解决方案

创建共享的契约测试套件：

```go
// internal/radiusd/repository/contract/user_repository_test.go
package contract

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/talkincode/toughradius/v9/internal/domain"
    "github.com/talkincode/toughradius/v9/internal/radiusd/repository"
)

// TestUserRepositoryContract 定义 UserRepository 必须满足的契约
func TestUserRepositoryContract(t *testing.T, factory RepositoryFactory) {
    t.Run("GetByUsername_ExistingUser_ShouldReturnUser", func(t *testing.T) {
        repo := factory.NewUserRepository()
        ctx := context.Background()

        // Setup: 创建测试用户
        user := &domain.RadiusUser{
            Username: "test_user",
            Password: "hashed_password",
            Status:   "enabled",
        }
        err := repo.Create(ctx, user)
        assert.NoError(t, err)

        // Act: 查询用户
        found, err := repo.GetByUsername(ctx, "test_user")

        // Assert
        assert.NoError(t, err)
        assert.NotNil(t, found)
        assert.Equal(t, "test_user", found.Username)
    })

    t.Run("GetByUsername_NonExistent_ShouldReturnError", func(t *testing.T) {
        repo := factory.NewUserRepository()
        ctx := context.Background()

        // Act
        found, err := repo.GetByUsername(ctx, "non_existent")

        // Assert
        assert.Error(t, err)
        assert.Nil(t, found)
        assert.Equal(t, gorm.ErrRecordNotFound, err)
    })

    // ... 更多契约测试
}
```

#### 实施步骤

**PR #1: 创建契约测试框架**

- [ ] 创建 `repository/contract/` 包
- [ ] 定义 `RepositoryFactory` 接口
- [ ] 实现 `UserRepositoryContract` 测试套件
- [ ] 实现 `SessionRepositoryContract` 测试套件

**PR #2: GORM 实现运行契约测试**

- [ ] 在 `gorm/*_repository_test.go` 中运行契约测试
- [ ] 修复发现的契约违规
- [ ] 确保 100% 契约通过

```go
// internal/radiusd/repository/gorm/user_repository_test.go
package gorm

import (
    "testing"

    "github.com/talkincode/toughradius/v9/internal/radiusd/repository/contract"
)

type gormRepositoryFactory struct {
    db *gorm.DB
}

func (f *gormRepositoryFactory) NewUserRepository() repository.UserRepository {
    return NewGormUserRepository(f.db)
}

func TestGormUserRepository_Contract(t *testing.T) {
    db := setupTestDB(t)
    factory := &gormRepositoryFactory{db: db}

    contract.TestUserRepositoryContract(t, factory)
}
```

**PR #3: 文档化契约要求**

- [ ] 在接口注释中添加契约说明
- [ ] 创建 `REPOSITORY_CONTRACT.md` 文档
- [ ] 说明如何为新实现运行契约测试
- [ ] 添加契约测试示例

#### 验收标准

- [ ] 所有 Repository 接口有契约测试
- [ ] GORM 实现通过所有契约测试
- [ ] 契约测试覆盖正常和异常场景
- [ ] 文档说明契约要求
- [ ] 新 Repository 实现必须通过契约测试

#### 后续优化

- 添加性能契约（如查询延迟上限）
- 实现并发安全性契约测试
- 支持契约测试的自动生成

---

### 任务 7: 插件系统依赖注入重构

**优先级**: P2 (Medium)  
**预计工作量**: 3 个 PR，约 4-6 工作日  
**影响范围**: `internal/radiusd/plugins/`, `main.go`

#### 问题描述

当前插件初始化是两阶段模式：

```go
// main.go
radiusService := radiusd.NewRadiusService(application)
defer radiusService.Release()

// 之后才初始化插件
plugins.InitPlugins(radiusService.SessionRepo, radiusService.AccountingRepo)
```

问题：

- 插件无法在测试中独立创建
- 生命周期管理不清晰
- 全局初始化违反依赖注入原则
- 初始化顺序依赖隐式约定

#### 解决方案

将插件转为可注入的服务：

```go
// Before
plugins.InitPlugins(sessionRepo, accountingRepo)

// After
pluginManager := plugins.NewPluginManager(
    plugins.WithSessionRepo(sessionRepo),
    plugins.WithAccountingRepo(accountingRepo),
)
radiusService.SetPluginManager(pluginManager)
```

#### 实施步骤

**PR #1: 创建 PluginManager**

- [ ] 创建 `PluginManager` 结构体
- [ ] 封装所有插件依赖
- [ ] 实现 Option 模式配置
- [ ] 添加单元测试

```go
// internal/radiusd/plugins/manager.go
package plugins

type PluginManager struct {
    sessionRepo    repository.SessionRepository
    accountingRepo repository.AccountingRepository

    authCheckers   []auth.Checker
    authGuards     []auth.Guard
    authEnhancers  []auth.Enhancer

    acctHandlers   []accounting.AccountingHandler
}

type Option func(*PluginManager)

func WithSessionRepo(repo repository.SessionRepository) Option {
    return func(m *PluginManager) {
        m.sessionRepo = repo
    }
}

func WithAccountingRepo(repo repository.AccountingRepository) Option {
    return func(m *PluginManager) {
        m.accountingRepo = repo
    }
}

func NewPluginManager(opts ...Option) *PluginManager {
    m := &PluginManager{
        authCheckers:  make([]auth.Checker, 0),
        authGuards:    make([]auth.Guard, 0),
        authEnhancers: make([]auth.Enhancer, 0),
        acctHandlers:  make([]accounting.AccountingHandler, 0),
    }

    for _, opt := range opts {
        opt(m)
    }

    // 注册所有插件
    m.registerPlugins()

    return m
}

func (m *PluginManager) registerPlugins() {
    // 注册认证检查器
    m.authCheckers = append(m.authCheckers,
        checkers.NewStatusChecker(),
        checkers.NewExpireChecker(),
        checkers.NewMacBindChecker(),
        checkers.NewVlanBindChecker(),
        checkers.NewSessionLimitChecker(m.sessionRepo),
    )

    // 注册其他插件...
}

func (m *PluginManager) GetAuthCheckers() []auth.Checker {
    return m.authCheckers
}

// ... 其他 getter 方法
```

**PR #2: 重构 RadiusService 集成**

- [ ] 在 `RadiusService` 中添加 `PluginManager` 字段
- [ ] 更新 `main.go` 使用新的初始化方式
- [ ] 移除 `plugins.InitPlugins()` 全局函数
- [ ] 确保所有功能正常

```go
// internal/radiusd/radius.go
type RadiusService struct {
    appCtx         app.AppContext
    pluginManager  *plugins.PluginManager
    // ... 其他字段
}

func NewRadiusService(appCtx app.AppContext) *RadiusService {
    db := appCtx.DB()
    s := &RadiusService{
        appCtx:        appCtx,
        // 初始化 repositories
        UserRepo:      repogorm.NewGormUserRepository(db),
        SessionRepo:   repogorm.NewGormSessionRepository(db),
        AccountingRepo: repogorm.NewGormAccountingRepository(db),
        NasRepo:       repogorm.NewGormNasRepository(db),
        // ... 其他初始化
    }

    // 创建并注入 PluginManager
    s.pluginManager = plugins.NewPluginManager(
        plugins.WithSessionRepo(s.SessionRepo),
        plugins.WithAccountingRepo(s.AccountingRepo),
    )

    return s
}

// main.go
radiusService := radiusd.NewRadiusService(application)
// 不再需要单独的 InitPlugins 调用
```

**PR #3: 更新测试**

- [ ] 重构所有插件测试使用 `PluginManager`
- [ ] 创建测试辅助函数
- [ ] 验证插件隔离性
- [ ] 更新文档

```go
// 测试辅助函数
func createTestPluginManager(t *testing.T) *plugins.PluginManager {
    mockSessionRepo := &mocks.SessionRepository{}
    mockAccountingRepo := &mocks.AccountingRepository{}

    return plugins.NewPluginManager(
        plugins.WithSessionRepo(mockSessionRepo),
        plugins.WithAccountingRepo(mockAccountingRepo),
    )
}
```

#### 验收标准

- [ ] 插件通过 `PluginManager` 管理
- [ ] 无全局插件初始化函数
- [ ] 插件可在测试中独立创建
- [ ] 依赖通过构造函数明确注入
- [ ] 所有测试通过
- [ ] 文档更新说明新模式

#### 后续优化

- 支持插件的动态加载/卸载
- 实现插件的热重载
- 添加插件依赖关系管理

---

## ⚡ 快速收益任务 (P3)

以下任务工作量小但能立即改善代码质量：

### 任务 8: 厂商注册验证

**优先级**: P3 (Low)  
**预计工作量**: 1-2 小时

在 `RegisterVendorParser()` 中添加重复注册检测：

```go
func RegisterVendorParser(parser vendorparserspkg.VendorParser) {
    globalRegistry.mu.Lock()
    defer globalRegistry.mu.Unlock()

    // 添加重复检测
    if _, exists := globalRegistry.vendorParsers[parser.VendorCode()]; exists {
        panic(fmt.Sprintf("vendor parser %s already registered", parser.VendorCode()))
    }

    globalRegistry.vendorParsers[parser.VendorCode()] = parser
}
```

---

### 任务 9: CGO 要求文档化

**优先级**: P3 (Low)  
**预计工作量**: 1 小时

在 `internal/app/database.go` 中添加注释：

```go
// getSqliteDatabase returns a SQLite database connection
//
// IMPORTANT: This project requires CGO_ENABLED=0 for cross-platform static compilation.
// We use github.com/glebarez/sqlite (pure Go) instead of github.com/mattn/go-sqlite3 (requires CGO).
//
// Trade-offs:
// - ✓ Cross-platform static binaries
// - ✓ No C compiler dependency
// - ✗ Slightly slower than CGO version
// - ✗ Some advanced SQLite features unavailable
//
// See: docs/sqlite-support.md for details
func getSqliteDatabase(config config.DBConfig, workdir string) *gorm.DB {
    // ...
}
```

---

### 任务 10: 提取魔术数字

**优先级**: P3 (Low)  
**预计工作量**: 2 小时

将 `radius.go` 中的硬编码值移到配置：

```go
// Before
const (
    RadiusRejectDelayTimes = 7  // 为什么是 7？
    RadiusAuthRateInterval = 1  // 为什么是 1 秒？
)

// After
// config/config.go
type RadiusConfig struct {
    // RejectDelaySeconds specifies how long to delay before sending reject response
    // to slow down brute-force attacks (default: 7 seconds)
    RejectDelaySeconds int `yaml:"reject_delay_seconds"`

    // AuthRateIntervalSeconds specifies the rate limiting window
    // for authentication attempts (default: 1 second)
    AuthRateIntervalSeconds int `yaml:"auth_rate_interval_seconds"`
}
```

---

### 任务 11-20: 其他快速收益

11. **健康检查端点** - 为每个 repository 添加 `/health/database`, `/health/cache`
12. **标准化错误包装** - Repository 层统一使用 `fmt.Errorf("repo: %w", err)`
13. **厂商常量合并** - 创建 `internal/radiusd/vendors/codes.go` 统一管理
14. **Context 超时** - 所有 repository 方法添加 context 超时处理
15. **AppContext 文档** - 在接口注释中说明生命周期（creation → init → release）
16. **Repository 日志** - 添加 debug 级别的调用日志
17. **集成测试辅助** - 创建 `CreateTestAppContext()` 避免重复
18. **配置验证** - 在 `LoadConfig()` 中添加配置完整性检查
19. **错误码标准化** - 定义统一的错误码常量
20. **性能监控埋点** - 在关键路径添加 metrics 采集点

---

## 🚀 未来演进建议 (长期规划)

### 1-2 年技术路线图

#### 1. 厂商解析器市场化 (6-12 个月)

**目标**: 支持第三方厂商插件

**步骤**:

1. Q1: 设计 Go 插件接口规范
2. Q2: 实现插件加载框架（.so 文件）
3. Q3: 创建插件开发 SDK 和文档
4. Q4: 建立插件市场/注册中心

**收益**:

- 社区可贡献厂商支持
- 厂商可自行开发插件
- 减轻核心维护负担

---

#### 2. 多租户数据库支持 (6-9 个月)

**目标**: 支持 SaaS 多租户部署

**步骤**:

1. Q1: 扩展 `DBProvider` 接口添加 `DBForTenant(id)`
2. Q2: 实现租户隔离策略（schema/database）
3. Q3: 添加租户管理 API
4. Q4: 实现租户数据迁移工具

**架构变更**:

```go
type MultiTenantDBProvider interface {
    DB() *gorm.DB  // 保留默认行为
    DBForTenant(tenantID string) (*gorm.DB, error)
    ListTenants() ([]string, error)
}
```

---

#### 3. Repository 缓存装饰器 (3-6 个月)

**目标**: 分离缓存逻辑，提高可维护性

**步骤**:

1. Q1: 设计 Repository 装饰器模式
2. Q2: 实现 `CachedRepository` 装饰器
3. Q3: 提取现有缓存逻辑到装饰器
4. Q4: 支持多级缓存（本地 + Redis）

**示例**:

```go
baseRepo := gorm.NewGormUserRepository(db)
cachedRepo := cache.NewCachedRepository(baseRepo,
    cache.WithTTL(5*time.Minute),
    cache.WithRedis(redisClient),
)
```

---

#### 4. 配置热重载 (3-4 个月)

**目标**: 支持配置动态更新

**步骤**:

1. 设计 `ConfigProvider.Watch(key, callback)` 接口
2. 实现文件监控和配置重载
3. 添加配置变更通知机制
4. 实现渐进式配置更新

**示例**:

```go
app.ConfigProvider().Watch("radius.max_sessions", func(old, new interface{}) {
    zap.L().Info("config changed",
        zap.String("key", "radius.max_sessions"),
        zap.Any("old", old),
        zap.Any("new", new),
    )
    // 自动应用新配置
})
```

---

#### 5. 可观测性增强 (3-6 个月)

**目标**: 完善追踪、指标、日志体系

**步骤**:

1. Q1: Repository 添加 OpenTelemetry 追踪
2. Q2: 实现 Repository 中间件模式
3. Q3: 添加自定义 metrics 导出
4. Q4: 集成分布式追踪系统

**装饰器链示例**:

```go
repo := gorm.NewGormUserRepository(db)
repo = metrics.NewMetricsRepository(repo, metricsCollector)
repo = tracing.NewTracingRepository(repo, tracer)
repo = cache.NewCachedRepository(repo, cacheConfig)
```

---

#### 6. Schema 版本可见性 (1-2 个月)

**目标**: 运维可查询数据库 schema 版本

**步骤**:

1. 添加 `/api/v1/database/version` API 端点
2. 返回当前 schema 版本和迁移历史
3. 添加迁移状态监控
4. 实现迁移健康检查

**API 响应示例**:

```json
{
  "current_version": 5,
  "dirty": false,
  "applied_migrations": [
    {
      "version": 1,
      "name": "initial_schema",
      "applied_at": "2025-01-01T00:00:00Z"
    },
    {
      "version": 5,
      "name": "add_plugin_config",
      "applied_at": "2025-11-20T10:30:00Z"
    }
  ],
  "pending_migrations": []
}
```

---

## 📊 任务优先级矩阵

| 任务                | 优先级 | 影响范围 | 工作量 | ROI | 建议开始时间 |
| ------------------- | ------ | -------- | ------ | --- | ------------ |
| 厂商代码管理中心化  | P1     | 高       | 中     | 高  | 立即         |
| 应用上下文接口隔离  | P1     | 高       | 高     | 高  | 本周         |
| 数据库迁移追踪      | P1     | 高       | 中     | 高  | 本周         |
| RADIUS 协议库抽象   | P1     | 高       | 高     | 中  | 下周         |
| 厂商配置外部化      | P2     | 中       | 中     | 中  | 2 周后       |
| Repository 契约测试 | P2     | 中       | 中     | 高  | 2 周后       |
| 插件依赖注入        | P2     | 中       | 中     | 中  | 3 周后       |
| 快速收益任务        | P3     | 低       | 低     | 高  | 随时         |

---

## 🎯 建议实施顺序

### 第一阶段（当前冲刺，2-3 周）

1. 完成所有 P3 快速收益任务（提升代码质量基线）
2. 启动"厂商代码管理中心化"（任务 1）
3. 并行进行"数据库迁移追踪"（任务 3）

### 第二阶段（下一冲刺，3-4 周）

1. 完成"应用上下文接口隔离"（任务 2）
2. 启动"RADIUS 协议库抽象"（任务 4）
3. 并行进行"Repository 契约测试"（任务 6）

### 第三阶段（后续迭代，4-6 周）

1. 完成"厂商配置外部化"（任务 5）
2. 完成"插件依赖注入"（任务 7）
3. 评估长期演进任务优先级

---

## 📝 附录

### A. 风险评估

| 任务           | 技术风险 | 业务风险 | 缓解措施                       |
| -------------- | -------- | -------- | ------------------------------ |
| 厂商代码中心化 | 低       | 低       | 完善测试覆盖                   |
| 接口隔离       | 中       | 低       | 渐进式迁移，保持向后兼容       |
| 迁移追踪       | 中       | 中       | 充分测试迁移脚本，准备回滚方案 |
| 协议库抽象     | 高       | 中       | 分阶段重构，保留适配层         |

### B. 成功指标

- **代码质量**: 测试覆盖率 > 85%，lint 无警告
- **可维护性**: 新增厂商支持 < 2 小时（vs 当前 1 天）
- **部署灵活性**: 支持配置热更新，零停机迁移
- **团队效率**: PR 平均审查时间 < 2 小时

### C. 相关文档

- [TDD Practice Guide](https://martinfowler.com/bliki/TestDrivenDevelopment.html)
- [Interface Segregation Principle](https://en.wikipedia.org/wiki/Interface_segregation_principle)
- [Database Migration Best Practices](https://www.dbchoices.com/migration-best-practices)
- [Repository Pattern](https://martinfowler.com/eaaCatalog/repository.html)

---

**最后更新**: 2025-11-20  
**维护者**: ToughRADIUS 架构团队  
**审查周期**: 每月更新优先级和进度
