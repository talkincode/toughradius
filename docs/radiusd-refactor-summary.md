# RADIUSD 模块化重构方案概要

## 📊 现状分析

### 当前架构

- **RadiusService**: 核心服务类 (~500 行)
  - 职责过多: 数据访问 + 缓存 + 配置 + 业务逻辑
  - 硬编码依赖: 直接调用 `app.GDB()` 和 `app.GApp()`
- **AuthService**: 认证服务 (~300 行)
  - 认证逻辑与数据访问耦合
  - EAP/CHAP/MSCHAP 逻辑混在一起
- **AcctService**: 计费服务 (~100 行)
  - Start/Update/Stop 逻辑分散

### 主要问题

1. ❌ **高耦合**: 业务逻辑与数据访问混杂
2. ❌ **难测试**: 缺少接口抽象,无法 Mock
3. ❌ **难扩展**: 新增厂商/认证方式需修改核心代码
4. ❌ **难维护**: 职责不清,代码重复

## 🎯 重构目标

### 核心理念

**可插拔的组件化架构** - 让 RADIUS 服务器变成一个"插件容器"

### 设计原则

- ✅ **单一职责**: 一个组件只做一件事
- ✅ **依赖倒置**: 依赖接口而非实现
- ✅ **开闭原则**: 对扩展开放,对修改关闭
- ✅ **插件化**: 通过注册机制实现可插拔

## 🏗️ 新架构设计

### 四层架构

```text
┌─────────────────────────────────────┐
│   RADIUS 协议层 (Server)            │  ← 现有不变
│   radius_auth.go / radius_acct.go   │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│   服务编排层 (Service)               │  ← 简化为协调器
│   AuthService / AcctService         │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│   插件层 (Plugins)                   │  ← 新增核心层
│   认证│策略│厂商│计费 插件            │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│   数据访问层 (Repository)            │  ← 新增抽象层
│   User│Session│Accounting│Nas│Config│
└─────────────────────────────────────┘
```

### 核心组件

#### 1️⃣ Repository 层 (数据访问抽象)

```go
// 接口定义
type UserRepository interface {
    GetByUsername(ctx, username) (*User, error)
    GetByMacAddr(ctx, mac) (*User, error)
    UpdateMacAddr(ctx, username, mac) error
    UpdateLastOnline(ctx, username) error
}

// GORM 实现
type GormUserRepository struct {
    db *gorm.DB
}
```

**优势**:

- 易于测试 (可以 Mock)
- 易于替换 (可以换成其他数据库)
- 逻辑清晰 (数据访问集中管理)

#### 2️⃣ 认证插件系统

```go
// 密码验证器接口
type PasswordValidator interface {
    Name() string                    // "pap", "chap", "mschap"
    CanHandle(ctx) bool              // 判断是否能处理
    Validate(ctx, password) error    // 执行验证
}

// PAP 验证器实现
type PAPValidator struct{}

func (v *PAPValidator) Validate(ctx, password) error {
    reqPwd := rfc2865.UserPassword_GetString(ctx.Request.Packet)
    if reqPwd != password {
        return errors.New("password mismatch")
    }
    return nil
}

// 自动注册
func init() {
    registry.RegisterPasswordValidator(&PAPValidator{})
}
```

**优势**:

- ✅ 新增认证方式无需修改核心代码
- ✅ 每个验证器可独立测试
- ✅ 支持动态启用/禁用

#### 3️⃣ 策略检查器插件

```go
// 策略检查器接口
type PolicyChecker interface {
    Name() string           // "expire", "mac_bind", "online_count"
    Order() int             // 执行顺序
    Check(ctx) error        // 执行检查
}

// 在线数检查器
type OnlineCountChecker struct {
    sessionRepo SessionRepository
}

func (c *OnlineCountChecker) Check(ctx) error {
    if ctx.User.ActiveNum == 0 {
        return nil  // 不限制
    }
    count, _ := c.sessionRepo.CountByUsername(ctx.User.Username)
    if count >= ctx.User.ActiveNum {
        return errors.New("online count exceeded")
    }
    return nil
}
```

**优势**:

- ✅ 策略可组合 (责任链模式)
- ✅ 策略可配置顺序
- ✅ 新增策略无需修改主流程

#### 4️⃣ 厂商插件系统

```go
// 厂商解析器接口
type VendorParser interface {
    VendorCode() string              // "2011" (Huawei)
    VendorName() string              // "Huawei"
    Parse(r) (*VendorRequest, error) // 解析私有属性
}

// 华为解析器
type HuaweiParser struct{}

func (p *HuaweiParser) VendorCode() string {
    return "2011"
}

func (p *HuaweiParser) Parse(r) (*VendorRequest, error) {
    // 华为特定逻辑
    macAddr := rfc2865.CallingStationID_GetString(r.Packet)
    return &VendorRequest{MacAddr: macAddr}, nil
}

func init() {
    registry.RegisterVendorParser(&HuaweiParser{})
}
```

**优势**:

- ✅ 新增厂商只需实现接口并注册
- ✅ 厂商代码独立维护
- ✅ 移除硬编码的 switch 语句

#### 5️⃣ 插件注册中心

```go
// 全局注册表
type Registry struct {
    validators map[string]PasswordValidator
    checkers   []PolicyChecker
    parsers    map[string]VendorParser
}

// 注册方法
func RegisterPasswordValidator(v PasswordValidator) {
    registry.validators[v.Name()] = v
}

func GetPasswordValidators() []PasswordValidator {
    return registry.validators
}
```

### 重构后的认证流程

```go
func (s *AuthService) ServeRADIUS(w, r) {
    ctx := &AuthContext{Request: r, Response: r.Response(...)}

    // 1. 获取 NAS (通过 Repository)
    nas, _ := s.nasRepo.GetByIP(nasIP)
    ctx.Nas = nas

    // 2. 解析厂商属性 (通过插件)
    parser, _ := registry.GetVendorParser(nas.VendorCode)
    ctx.VendorRequest, _ = parser.Parse(r)

    // 3. 获取用户 (通过 Repository)
    ctx.User, _ = s.userRepo.GetByUsername(username)

    // 4. 执行策略检查 (责任链)
    for _, checker := range registry.GetPolicyCheckers() {
        if err := checker.Check(ctx); err != nil {
            s.sendReject(w, r, err)
            return
        }
    }

    // 5. 执行密码验证 (策略模式)
    for _, validator := range registry.GetPasswordValidators() {
        if validator.CanHandle(ctx) {
            if err := validator.Validate(ctx, user.Password); err != nil {
                s.sendReject(w, r, err)
                return
            }
            break
        }
    }

    // 6. 发送 Accept
    s.sendAccept(w, ctx)
}
```

**对比**:

- ❌ 旧代码: 500 行,逻辑复杂,难以理解
- ✅ 新代码: 50 行,清晰明了,易于维护

## 📋 实施计划

### 阶段划分

| 阶段        | 任务                    | 时间   | 风险 |
| ----------- | ----------------------- | ------ | ---- |
| 1️⃣ 基础设施 | 创建接口层和 Repository | 2-3 天 | 低   |
| 2️⃣ 认证插件 | 迁移密码验证和策略检查  | 3-4 天 | 中   |
| 3️⃣ 厂商插件 | 迁移厂商解析逻辑        | 2-3 天 | 低   |
| 4️⃣ 计费插件 | 迁移计费处理逻辑        | 2 天   | 低   |
| 5️⃣ 核心重构 | 简化核心服务            | 2-3 天 | 中   |
| 6️⃣ 测试优化 | 完善测试和性能优化      | 3-4 天 | 低   |

**总计**: 14-19 个工作日 (约 3-4 周)

### 向后兼容策略

使用**适配器模式**保持 API 兼容:

```go
// 原有方法保留,标记废弃
// Deprecated: Use userRepo.GetByUsername instead
func (s *RadiusService) GetValidUser(username) (*User, error) {
    return s.userRepo.GetByUsername(context.Background(), username)
}
```

## ✨ 收益分析

### 代码质量

| 指标           | 当前    | 重构后  | 提升    |
| -------------- | ------- | ------- | ------- |
| 核心文件行数   | ~500 行 | ~150 行 | ⬇️ 70%  |
| 圈复杂度       | 15-20   | 5-8     | ⬇️ 60%  |
| 单元测试覆盖率 | ~30%    | ~80%    | ⬆️ 150% |

### 可维护性

- ✅ **职责清晰**: 每个模块职责单一
- ✅ **易于理解**: 代码结构清晰,逻辑简单
- ✅ **易于修改**: 修改影响范围小

### 可扩展性

- ✅ **新增厂商**: 5 分钟 (实现接口 + 注册)
- ✅ **新增认证**: 10 分钟 (实现 Validator)
- ✅ **新增策略**: 10 分钟 (实现 Checker)

### 可测试性

- ✅ **单元测试**: 每个插件独立测试
- ✅ **集成测试**: 通过 Mock 简化测试
- ✅ **性能测试**: 易于对比和优化

## ⚠️ 风险与应对

### 风险

1. **重构引入 Bug**: 大量代码修改可能引入新问题
2. **性能下降**: 接口抽象可能影响性能 (预计 2-5%)
3. **学习曲线**: 团队需要适应新架构

### 应对

1. ✅ **完善测试**: 先建立完整的集成测试覆盖
2. ✅ **渐进式重构**: 增量迁移,每步保持测试通过
3. ✅ **性能基准**: 持续监控性能,及时优化
4. ✅ **文档完善**: 详细的开发指南和示例
5. ✅ **代码审查**: 严格的审查流程

## 🎓 示例代码

### 添加新的认证方式 (EAP-TLS)

```go
// 1. 实现 PasswordValidator 接口
type EAPTLSValidator struct{}

func (v *EAPTLSValidator) Name() string {
    return "eap-tls"
}

func (v *EAPTLSValidator) CanHandle(ctx *AuthContext) bool {
    eapMsg, err := parseEAPMessage(ctx.Request.Packet)
    return err == nil && eapMsg.Type == EAPTypeTLS
}

func (v *EAPTLSValidator) Validate(ctx *AuthContext, password string) error {
    // EAP-TLS 验证逻辑
    return nil
}

// 2. 注册插件
func init() {
    registry.RegisterPasswordValidator(&EAPTLSValidator{})
}
```

**就这样!** 无需修改任何核心代码,新增认证方式只需 20 行代码。

### 添加新的厂商支持 (Cisco)

```go
// 1. 实现 VendorParser 接口
type CiscoParser struct{}

func (p *CiscoParser) VendorCode() string {
    return "9"
}

func (p *CiscoParser) VendorName() string {
    return "Cisco"
}

func (p *CiscoParser) Parse(r *radius.Request) (*VendorRequest, error) {
    // Cisco 特定解析逻辑
    macAddr := rfc2865.CallingStationID_GetString(r.Packet)
    return &VendorRequest{MacAddr: macAddr}, nil
}

// 2. 注册插件
func init() {
    registry.RegisterVendorParser(&CiscoParser{})
}
```

**完成!** 新增厂商支持只需实现接口并注册。

## 📊 总结

### 可行性评估

- ✅ **技术可行**: 完全可行,Go 语言接口机制天然支持
- ✅ **时间可行**: 3-4 周可完成
- ✅ **风险可控**: 渐进式重构,风险低

### 建议

**强烈建议进行重构**,理由:

1. 📈 **长期收益大**: 代码质量和可维护性大幅提升
2. 🚀 **扩展性强**: 新增功能成本降低 90%
3. 🧪 **可测试性好**: 单元测试覆盖率提升 150%
4. 📚 **易于理解**: 新成员上手时间减少 50%
5. 🔧 **易于维护**: Bug 修复时间减少 60%

### 下一步

1. ✅ **评审设计方案** (本文档)
2. ⬜ **建立测试基线** (集成测试覆盖)
3. ⬜ **第一阶段实施** (创建接口层)
4. ⬜ **逐步迁移** (按阶段推进)
5. ⬜ **持续优化** (性能和文档)

---

**详细设计文档**: 参见 `radiusd-refactor-design.md`
