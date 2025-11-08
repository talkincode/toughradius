# RADIUSD 模块化重构设计方案

## 1. 当前架构分析

### 1.1 核心组件结构

```text
internal/radiusd/
├── radius.go              # RadiusService 核心服务类（约500行）
├── server.go              # 服务器启动函数
├── radius_auth.go         # AuthService 认证服务（约300行）
├── radius_acct.go         # AcctService 计费服务（约100行）
├── radsec_server.go       # RadSec TLS服务器
├── radsec_service.go      # RadsecService RadSec服务
│
├── auth_passwd_check.go   # 密码验证逻辑
├── auth_bind_check.go     # MAC/VLAN绑定检查
├── auth_check_online.go   # 在线数检查
├── auth_accept_config.go  # Accept响应配置
│
├── acct_start.go          # 计费Start处理
├── acct_update.go         # 计费Update处理
├── acct_stop.go           # 计费Stop处理
├── acct_ops.go            # 计费通用操作
│
├── radius_eap.go          # EAP通用逻辑
├── radius_eap_mschapv2.go # EAP-MSCHAPv2实现
├── vendor_parse.go        # 厂商属性解析
│
└── vendors/               # 厂商特定实现
    ├── huawei/
    ├── h3c/
    ├── mikrotik/
    └── ...
```

### 1.2 当前问题识别

#### 问题 1: 职责混乱

`RadiusService` 包含过多职责：

- 数据库操作 (GetValidUser, AddRadiusOnline, UpdateRadiusOnlineData...)
- 缓存管理 (AuthRateCache, EapStateCache, RejectCache)
- 配置访问 (GetIntConfig, GetStringConfig)
- 业务逻辑 (CheckAuthRateLimit, ParseVendor)

#### 问题 2: 硬编码依赖

- 直接使用 `app.GDB()` 访问数据库
- 直接使用 `app.GApp()` 访问配置
- 厂商代码硬编码在 switch 语句中

#### 问题 3: 难以测试

- 业务逻辑与数据访问耦合
- 缺少接口抽象
- Mock 困难

#### 问题 4: 扩展性差

- 新增厂商需修改核心代码
- 新增认证方式需修改主流程
- 新增策略检查需修改认证逻辑

#### 问题 5: 代码重复

- 多处错误处理模式重复
- 日志记录模式重复
- 配置读取模式重复

## 2. 模块化重构设计

### 2.1 架构分层

```text
┌─────────────────────────────────────────────────┐
│         RADIUS Protocol Layer (Server)          │
│  (radius_auth.go, radius_acct.go, radsec)      │
└─────────────────┬───────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────┐
│            Service Orchestration                │
│  (AuthService, AcctService - 协调器模式)        │
└─────────────────┬───────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────┐
│          Plugin System (Handler Chain)          │
│  认证插件 │ 策略插件 │ 计费插件 │ 厂商插件      │
└─────────────────┬───────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────┐
│         Repository Layer (Data Access)          │
│    User │ Session │ Accounting │ NAS │ Config   │
└─────────────────────────────────────────────────┘
```

### 2.2 核心接口设计

#### 2.2.1 数据访问层 (Repository)

```go
// internal/radiusd/repository/interfaces.go

package repository

import (
    "context"
    "github.com/talkincode/toughradius/v9/internal/domain"
)

// UserRepository 用户数据访问接口
type UserRepository interface {
    // GetByUsername 根据用户名查询用户
    GetByUsername(ctx context.Context, username string) (*domain.RadiusUser, error)

    // GetByMacAddr 根据MAC地址查询用户
    GetByMacAddr(ctx context.Context, macAddr string) (*domain.RadiusUser, error)

    // UpdateMacAddr 更新用户MAC地址
    UpdateMacAddr(ctx context.Context, username, macAddr string) error

    // UpdateVlanId 更新用户VLAN ID
    UpdateVlanId(ctx context.Context, username string, vlanId1, vlanId2 int) error

    // UpdateLastOnline 更新最后在线时间
    UpdateLastOnline(ctx context.Context, username string) error
}

// SessionRepository 在线会话管理接口
type SessionRepository interface {
    // Create 创建在线会话
    Create(ctx context.Context, session *domain.RadiusOnline) error

    // Update 更新会话数据
    Update(ctx context.Context, session *domain.RadiusOnline) error

    // Delete 删除会话
    Delete(ctx context.Context, sessionId string) error

    // GetBySessionId 根据会话ID查询
    GetBySessionId(ctx context.Context, sessionId string) (*domain.RadiusOnline, error)

    // CountByUsername 统计用户在线数
    CountByUsername(ctx context.Context, username string) (int, error)

    // BatchDelete 批量删除
    BatchDelete(ctx context.Context, ids []string) error
}

// AccountingRepository 计费记录接口
type AccountingRepository interface {
    // Create 创建计费记录
    Create(ctx context.Context, accounting *domain.RadiusAccounting) error

    // UpdateStop 更新停止时间和流量
    UpdateStop(ctx context.Context, sessionId string, accounting *domain.RadiusAccounting) error
}

// NasRepository NAS设备管理接口
type NasRepository interface {
    // GetByIP 根据IP查询NAS
    GetByIP(ctx context.Context, ip string) (*domain.NetNas, error)

    // GetByIdentifier 根据标识符查询NAS
    GetByIdentifier(ctx context.Context, identifier string) (*domain.NetNas, error)
}

// ConfigRepository 配置访问接口
type ConfigRepository interface {
    // GetString 获取字符串配置
    GetString(ctx context.Context, category, key string) string

    // GetInt 获取整数配置
    GetInt(ctx context.Context, category, key string, defaultVal int64) int64

    // GetBool 获取布尔配置
    GetBool(ctx context.Context, category, key string) bool
}
```

#### 2.2.2 认证处理插件

```go
// internal/radiusd/plugins/auth/interfaces.go

package auth

import (
    "context"
    "github.com/talkincode/toughradius/v9/internal/domain"
    "layeh.com/radius"
)

// AuthContext 认证上下文
type AuthContext struct {
    Request       *radius.Request
    Response      *radius.Packet
    User          *domain.RadiusUser
    Nas           *domain.NetNas
    VendorRequest *VendorRequest
    Metadata      map[string]interface{}
}

// PasswordValidator 密码验证器接口
type PasswordValidator interface {
    // Name 返回验证器名称 (pap, chap, mschap, eap-md5, etc.)
    Name() string

    // CanHandle 判断是否可以处理该请求
    CanHandle(ctx *AuthContext) bool

    // Validate 执行密码验证
    Validate(ctx *AuthContext, password string) error
}

// PolicyChecker 策略检查器接口
type PolicyChecker interface {
    // Name 返回检查器名称
    Name() string

    // Check 执行策略检查
    Check(ctx *AuthContext) error

    // Order 返回执行顺序（数字越小越先执行）
    Order() int
}

// ResponseEnhancer 响应增强器接口
type ResponseEnhancer interface {
    // Name 返回增强器名称
    Name() string

    // Enhance 增强响应内容（添加厂商属性等）
    Enhance(ctx *AuthContext) error
}
```

#### 2.2.3 厂商插件接口

```go
// internal/radiusd/plugins/vendor/interfaces.go

package vendor

import (
    "layeh.com/radius"
)

// VendorRequest 厂商请求数据
type VendorRequest struct {
    MacAddr string
    Vlanid1 int64
    Vlanid2 int64
}

// VendorParser 厂商属性解析器接口
type VendorParser interface {
    // VendorCode 返回厂商代码
    VendorCode() string

    // VendorName 返回厂商名称
    VendorName() string

    // Parse 解析厂商私有属性
    Parse(r *radius.Request) (*VendorRequest, error)
}

// VendorResponseBuilder 厂商响应构建器接口
type VendorResponseBuilder interface {
    // VendorCode 返回厂商代码
    VendorCode() string

    // Build 构建厂商特定的响应属性
    Build(resp *radius.Packet, user interface{}, vlanId1, vlanId2 int) error
}
```

#### 2.2.4 计费处理插件

```go
// internal/radiusd/plugins/accounting/interfaces.go

package accounting

import (
    "context"
    "github.com/talkincode/toughradius/v9/internal/domain"
    "layeh.com/radius"
)

// AcctContext 计费上下文
type AcctContext struct {
    Request       *radius.Request
    User          *domain.RadiusUser
    Nas           *domain.NetNas
    VendorRequest interface{}
    NasRealIP     string
}

// AccountingHandler 计费处理器接口
type AccountingHandler interface {
    // StatusType 返回处理的状态类型 (Start, Stop, Update, etc.)
    StatusType() uint32

    // Handle 处理计费请求
    Handle(ctx context.Context, acctCtx *AcctContext) error
}
```

### 2.3 插件注册机制

```go
// internal/radiusd/registry/registry.go

package registry

import (
    "fmt"
    "sync"
)

// Registry 插件注册中心
type Registry struct {
    passwordValidators map[string]PasswordValidator
    policyCheckers     []PolicyChecker
    vendorParsers      map[string]VendorParser
    acctHandlers       map[uint32]AccountingHandler
    mu                 sync.RWMutex
}

var globalRegistry = &Registry{
    passwordValidators: make(map[string]PasswordValidator),
    vendorParsers:      make(map[string]VendorParser),
    acctHandlers:       make(map[uint32]AccountingHandler),
    policyCheckers:     make([]PolicyChecker, 0),
}

// RegisterPasswordValidator 注册密码验证器
func RegisterPasswordValidator(validator PasswordValidator) {
    globalRegistry.mu.Lock()
    defer globalRegistry.mu.Unlock()
    globalRegistry.passwordValidators[validator.Name()] = validator
}

// RegisterVendorParser 注册厂商解析器
func RegisterVendorParser(parser VendorParser) {
    globalRegistry.mu.Lock()
    defer globalRegistry.mu.Unlock()
    globalRegistry.vendorParsers[parser.VendorCode()] = parser
}

// GetVendorParser 获取厂商解析器
func GetVendorParser(vendorCode string) (VendorParser, error) {
    globalRegistry.mu.RLock()
    defer globalRegistry.mu.RUnlock()

    parser, ok := globalRegistry.vendorParsers[vendorCode]
    if !ok {
        // 返回默认解析器
        return globalRegistry.vendorParsers["default"], nil
    }
    return parser, nil
}

// 其他注册和获取方法...
```

### 2.4 实现示例

#### 2.4.1 PAP 密码验证器实现

```go
// internal/radiusd/plugins/auth/validators/pap_validator.go

package validators

import (
    "github.com/talkincode/toughradius/v9/internal/app"
    "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
    "layeh.com/radius/rfc2865"
)

type PAPValidator struct{}

func (v *PAPValidator) Name() string {
    return "pap"
}

func (v *PAPValidator) CanHandle(ctx *auth.AuthContext) bool {
    password := rfc2865.UserPassword_GetString(ctx.Request.Packet)
    return password != ""
}

func (v *PAPValidator) Validate(ctx *auth.AuthContext, password string) error {
    requestPassword := rfc2865.UserPassword_GetString(ctx.Request.Packet)

    if requestPassword != password {
        return NewAuthError(app.MetricsRadiusRejectPasswdError,
            "PAP password mismatch")
    }

    return nil
}

// 在 init 函数中自动注册
func init() {
    registry.RegisterPasswordValidator(&PAPValidator{})
}
```

#### 2.4.2 在线数检查器实现

```go
// internal/radiusd/plugins/auth/checkers/online_count_checker.go

package checkers

import (
    "context"
    "github.com/talkincode/toughradius/v9/internal/app"
    "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
    "github.com/talkincode/toughradius/v9/internal/radiusd/repository"
)

type OnlineCountChecker struct {
    sessionRepo repository.SessionRepository
}

func NewOnlineCountChecker(sessionRepo repository.SessionRepository) *OnlineCountChecker {
    return &OnlineCountChecker{sessionRepo: sessionRepo}
}

func (c *OnlineCountChecker) Name() string {
    return "online_count"
}

func (c *OnlineCountChecker) Order() int {
    return 30 // 在MAC绑定之后执行
}

func (c *OnlineCountChecker) Check(ctx *auth.AuthContext) error {
    user := ctx.User

    // activeNum为0表示不限制
    if user.ActiveNum == 0 {
        return nil
    }

    count, err := c.sessionRepo.CountByUsername(
        context.Background(),
        user.Username,
    )
    if err != nil {
        return err
    }

    if count >= user.ActiveNum {
        return NewAuthError(app.MetricsRadiusRejectLimit,
            "user online count exceeded")
    }

    return nil
}

func init() {
    // 需要在应用启动时注册，因为需要依赖注入
}
```

#### 2.4.3 华为厂商解析器实现

```go
// internal/radiusd/plugins/vendor/parsers/huawei_parser.go

package parsers

import (
    "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendor"
    "github.com/talkincode/toughradius/v9/internal/radiusd/vendors/huawei"
    "layeh.com/radius"
    "layeh.com/radius/rfc2865"
)

type HuaweiParser struct{}

func (p *HuaweiParser) VendorCode() string {
    return "2011"
}

func (p *HuaweiParser) VendorName() string {
    return "Huawei"
}

func (p *HuaweiParser) Parse(r *radius.Request) (*vendor.VendorRequest, error) {
    vr := &vendor.VendorRequest{}

    // 解析MAC地址
    macAddr := rfc2865.CallingStationID_GetString(r.Packet)
    if macAddr != "" {
        vr.MacAddr = macAddr
    }

    // 解析VLAN ID（华为特定逻辑）
    // ... 华为特定实现

    return vr, nil
}

func init() {
    registry.RegisterVendorParser(&HuaweiParser{})
}
```

### 2.5 重构后的认证服务

```go
// internal/radiusd/services/auth_service.go

package services

import (
    "context"
    "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
    "github.com/talkincode/toughradius/v9/internal/radiusd/registry"
    "github.com/talkincode/toughradius/v9/internal/radiusd/repository"
    "layeh.com/radius"
)

type AuthService struct {
    userRepo    repository.UserRepository
    sessionRepo repository.SessionRepository
    nasRepo     repository.NasRepository
    configRepo  repository.ConfigRepository
}

func NewAuthService(
    userRepo repository.UserRepository,
    sessionRepo repository.SessionRepository,
    nasRepo repository.NasRepository,
    configRepo repository.ConfigRepository,
) *AuthService {
    return &AuthService{
        userRepo:    userRepo,
        sessionRepo: sessionRepo,
        nasRepo:     nasRepo,
        configRepo:  configRepo,
    }
}

func (s *AuthService) ServeRADIUS(w radius.ResponseWriter, r *radius.Request) {
    ctx := context.Background()

    // 1. 创建认证上下文
    authCtx := &auth.AuthContext{
        Request:  r,
        Response: r.Response(radius.CodeAccessAccept),
        Metadata: make(map[string]interface{}),
    }

    // 2. 解析NAS信息
    nas, err := s.parseNas(ctx, r)
    if err != nil {
        s.sendReject(w, r, err)
        return
    }
    authCtx.Nas = nas

    // 3. 解析厂商属性
    vendorParser, _ := registry.GetVendorParser(nas.VendorCode)
    vendorReq, _ := vendorParser.Parse(r)
    authCtx.VendorRequest = vendorReq

    // 4. 获取用户信息
    username := rfc2865.UserName_GetString(r.Packet)
    user, err := s.userRepo.GetByUsername(ctx, username)
    if err != nil {
        s.sendReject(w, r, err)
        return
    }
    authCtx.User = user

    // 5. 执行策略检查（责任链模式）
    checkers := registry.GetPolicyCheckers()
    for _, checker := range checkers {
        if err := checker.Check(authCtx); err != nil {
            s.sendReject(w, r, err)
            return
        }
    }

    // 6. 执行密码验证
    validators := registry.GetPasswordValidators()
    validated := false
    for _, validator := range validators {
        if validator.CanHandle(authCtx) {
            if err := validator.Validate(authCtx, user.Password); err != nil {
                s.sendReject(w, r, err)
                return
            }
            validated = true
            break
        }
    }

    if !validated {
        s.sendReject(w, r, errors.New("no validator found"))
        return
    }

    // 7. 增强响应（添加厂商属性）
    enhancers := registry.GetResponseEnhancers()
    for _, enhancer := range enhancers {
        enhancer.Enhance(authCtx)
    }

    // 8. 发送Accept
    s.sendAccept(w, authCtx)
}
```

## 3. 重构实施计划

### 3.1 第一阶段：建立基础设施（2-3 天）

**目标**: 创建接口层，不影响现有功能

1. 创建目录结构

   ```text
   internal/radiusd/
   ├── repository/
   │   ├── interfaces.go
   │   └── gorm/              # GORM实现
   ├── plugins/
   │   ├── auth/
   │   │   ├── interfaces.go
   │   │   ├── validators/
   │   │   └── checkers/
   │   ├── vendor/
   │   │   ├── interfaces.go
   │   │   └── parsers/
   │   └── accounting/
   │       ├── interfaces.go
   │       └── handlers/
   └── registry/
       └── registry.go
   ```

2. 实现 Repository 接口

   - UserRepository (GORM 实现)
   - SessionRepository (GORM 实现)
   - AccountingRepository (GORM 实现)
   - NasRepository (GORM 实现)
   - ConfigRepository (GORM 实现)

3. 建立插件注册机制
   - 创建全局 Registry
   - 实现注册和查询方法

### 3.2 第二阶段：迁移认证插件（3-4 天）

**目标**: 迁移密码验证和策略检查逻辑

1. 实现密码验证器

   - PAPValidator
   - CHAPValidator
   - MSCHAPValidator
   - EAPMD5Validator
   - EAPMSCHAPv2Validator
   - EAPOTPValidator

2. 实现策略检查器

   - ExpireChecker
   - StatusChecker
   - MacBindChecker
   - VlanBindChecker
   - OnlineCountChecker
   - RateLimitChecker

3. 适配器模式兼容现有代码

   ```go
   // 保留原有方法,内部调用新架构
   func (s *AuthService) CheckPassword(...) error {
       // 调用新的validator
       return s.passwordValidator.Validate(...)
   }
   ```

### 3.3 第三阶段：迁移厂商插件（2-3 天）

**目标**: 厂商解析和响应构建插件化

1. 实现 VendorParser

   - StandardParser (默认)
   - HuaweiParser
   - H3CParser
   - MikrotikParser
   - ZTEParser
   - CiscoParser
   - 等等...

2. 实现 VendorResponseBuilder

   - 每个厂商对应的响应构建器

3. 移除硬编码的 switch 语句

### 3.4 第四阶段：迁移计费插件（2 天）

**目标**: 计费处理插件化

1. 实现 AccountingHandler

   - StartHandler
   - UpdateHandler
   - StopHandler
   - NasOnHandler
   - NasOffHandler

2. 重构 AcctService 使用新架构

### 3.5 第五阶段：重构核心服务（2-3 天）

**目标**: 简化核心服务，使用依赖注入

1. 重构 AuthService

   - 使用 Repository 而非直接访问数据库
   - 使用 Plugin 而非硬编码逻辑
   - 实现责任链模式

2. 重构 AcctService

   - 使用 Repository
   - 使用 Handler 插件

3. 重构 RadiusService
   - 移除数据访问代码
   - 保留必要的协调逻辑

### 3.6 第六阶段：测试和优化（3-4 天）

1. 单元测试

   - 每个插件的单元测试
   - Repository 的单元测试
   - Mock 测试

2. 集成测试

   - 完整认证流程测试
   - 完整计费流程测试
   - 各厂商设备测试

3. 性能测试

   - 基准测试对比
   - 并发压力测试
   - 优化性能瓶颈

4. 文档更新
   - API 文档
   - 插件开发指南
   - 迁移指南

## 4. 向后兼容策略

### 4.1 适配器模式

为保持 API 兼容，保留原有公共方法，内部调用新架构：

```go
// 原有方法保留，标记为废弃
// Deprecated: Use repository.UserRepository instead
func (s *RadiusService) GetValidUser(username string) (*domain.RadiusUser, error) {
    return s.userRepo.GetByUsername(context.Background(), username)
}
```

### 4.2 渐进式迁移

- 新功能使用新架构
- 旧功能逐步迁移
- 保持测试通过

### 4.3 配置兼容

- 配置项保持不变
- 新增插件配置项（可选）

## 5. 收益评估

### 5.1 可维护性提升

- **代码行数减少**: 预计核心文件减少 30-40%
- **圈复杂度降低**: 单个方法复杂度从 15-20 降至 5-8
- **职责清晰**: 每个模块职责单一

### 5.2 可测试性提升

- **单元测试覆盖率**: 从 30%提升至 80%+
- **Mock 容易**: 基于接口的 Mock
- **集成测试简化**: 可以独立测试每个插件

### 5.3 可扩展性提升

- **新增厂商**: 只需实现 Parser 接口，无需修改核心代码
- **新增认证方式**: 只需实现 Validator 接口
- **新增策略**: 只需实现 Checker 接口

### 5.4 性能影响

- **接口调用开销**: 约 2-5%性能损失（可接受）
- **插件查找开销**: 使用 map 查找，O(1)复杂度
- **优化空间**: 可以通过缓存、预编译等方式优化

## 6. 风险和应对

### 6.1 风险识别

1. **重构引入 Bug**: 大量代码修改可能引入新问题
2. **性能下降**: 接口抽象可能影响性能
3. **学习曲线**: 团队需要适应新架构
4. **测试覆盖不足**: 可能遗漏边界情况

### 6.2 应对措施

1. **完善测试**: 先建立完善的集成测试
2. **渐进式重构**: 增量迁移，每步都保持测试通过
3. **性能基准**: 建立性能基准，持续监控
4. **代码审查**: 严格的代码审查流程
5. **文档完善**: 详细的文档和示例

## 7. 总结

### 7.1 可行性结论

**完全可行**，建议采用渐进式重构策略。

### 7.2 关键成功因素

1. **完善的测试**: 重构前建立完整的测试覆盖
2. **渐进式迁移**: 不要一次性重构所有代码
3. **向后兼容**: 保持 API 兼容性
4. **性能监控**: 持续监控性能指标
5. **团队协作**: 团队成员充分理解新架构

### 7.3 预期时间

- **第一阶段**: 2-3 天
- **第二阶段**: 3-4 天
- **第三阶段**: 2-3 天
- **第四阶段**: 2 天
- **第五阶段**: 2-3 天
- **第六阶段**: 3-4 天

**总计**: 约 14-19 个工作日（3-4 周）

### 7.4 下一步行动

1. ✅ 完成架构设计文档（本文档）
2. ⬜ 建立完整的集成测试套件
3. ⬜ 创建第一阶段的 PR（接口层）
4. ⬜ 逐步实施后续阶段
5. ⬜ 持续优化和文档更新
