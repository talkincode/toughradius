# RADIUSD 模块化重构 - 第三阶段进展报告

## 已完成工作

### 1. ✅ 解决循环导入问题

**问题**: `internal/radiusd/plugins` 包导入 `internal/radiusd` 造成循环依赖

**解决方案**: 创建独立的 `internal/radiusd/errors` 包

**新建文件**:

- `internal/radiusd/errors/errors.go` - 定义 AuthError 和便捷构造函数

**修改文件** (8 个):

- `plugins/auth/checkers/*.go` (5 个) - 使用 `errors` 包
- `plugins/auth/validators/*.go` (3 个) - 使用 `errors` 包

**验证**: ✅ `go build ./internal/radiusd/...` 成功

---

### 2. ✅ 创建插件集成层

**新建文件**: `internal/radiusd/auth_plugin_integration.go`

**核心方法**:

#### `AuthenticateUserWithPlugins()`

使用插件系统的新认证流程入口：

```go
func (s *AuthService) AuthenticateUserWithPlugins(
    ctx context.Context,
    r *radius.Request,
    response *radius.Packet,
    user *domain.RadiusUser,
    vendorReq *vendorparsers.VendorRequest,
    isMacAuth bool,
) error
```

**流程**:

1. 创建 `AuthContext` 认证上下文
2. 调用 `validatePasswordWithPlugins()` - 密码验证
3. 调用 `checkPoliciesWithPlugins()` - 策略检查

#### `validatePasswordWithPlugins()`

密码验证器插件调度：

- 获取所有已注册的 `PasswordValidator`
- 遍历找到能处理当前请求的验证器
- 执行验证，失败则返回错误
- 如果没有插件，回退到原有 `CheckPassword()` 方法

#### `checkPoliciesWithPlugins()`

策略检查器插件调度：

- 获取所有已注册的 `PolicyChecker`（已按 Order 排序）
- 按顺序执行所有检查器
- 任一检查失败立即返回错误
- 如果没有插件，回退到 `checkPoliciesLegacy()`

#### `checkPoliciesLegacy()`

原有策略检查的后备逻辑：

- 在线数限制检查
- MAC 绑定检查
- VLAN 绑定检查

---

### 3. ✅ 更新认证上下文

**修改文件**: `internal/radiusd/plugins/auth/interfaces.go`

**新增字段**:

```go
type AuthContext struct {
    Request       *radius.Request
    Response      *radius.Packet
    User          *domain.RadiusUser
    Nas           *domain.NetNas
    VendorRequest interface{}
    IsMacAuth     bool                   // ✨ 新增：是否为MAC认证
    Metadata      map[string]interface{} // 额外的元数据
}
```

---

## 架构优势

### 1. 向后兼容

- 保留所有原有方法（标记为 `Deprecated`）
- 新方法与旧方法并存，可以逐步迁移
- 如果没有注册插件，自动回退到原有逻辑

### 2. 可插拔设计

```
认证请求
   ↓
AuthenticateUserWithPlugins()
   ↓
   ├─ validatePasswordWithPlugins()
   │     ├─ PAPValidator (如果能处理)
   │     ├─ CHAPValidator (如果能处理)
   │     └─ MSCHAPValidator (如果能处理)
   │
   └─ checkPoliciesWithPlugins()
         ├─ StatusChecker (Order: 5)
         ├─ ExpireChecker (Order: 10)
         ├─ MacBindChecker (Order: 20)
         ├─ VlanBindChecker (Order: 21)
         └─ OnlineCountChecker (Order: 30)
```

### 3. 责任清晰

- **AuthContext**: 携带所有认证所需数据
- **PasswordValidator**: 只负责密码验证
- **PolicyChecker**: 只负责策略检查
- **Registry**: 管理所有插件注册和查询

---

## 下一步工作

### 优先级 1: 在 radius_auth.go 中使用新方法

需要修改 `ServeRADIUS()` 方法：

```go
// 当前代码 (约 line 150)
user, err := s.GetValidUser(username, isMacAuth)
s.CheckRadAuthError(username, ip, err)

if !isMacAuth {
    s.CheckRadAuthError(username, ip, s.CheckOnlineCount(username, user.ActiveNum))
    s.CheckRadAuthError(username, ip, s.CheckMacBind(user, vendorReq))
    s.CheckRadAuthError(username, ip, s.CheckVlanBind(user, vendorReq))
}

// 获取密码并验证
localpwd, err := s.GetLocalPassword(user, isMacAuth)
s.CheckRadAuthError(username, ip, err)
err = s.CheckPassword(r, username, localpwd, response, isMacAuth)
s.CheckRadAuthError(username, ip, err)

// 建议替换为
user, err := s.GetValidUser(username, isMacAuth)
s.CheckRadAuthError(username, ip, err)

// 使用插件系统进行认证
err = s.AuthenticateUserWithPlugins(ctx, r, response, user, vendorReq, isMacAuth)
s.CheckRadAuthError(username, ip, err)
```

### 优先级 2: 集成厂商解析器

在 `vendor_parse.go` 中使用 `VendorParser` 插件

### 优先级 3: 重构计费流程

修改 `radius_acct.go` 使用 `AccountingHandler` 插件

### 优先级 4: 测试和优化

- 单元测试
- 集成测试
- 性能基准测试

---

## 编译状态

✅ 所有代码编译通过:

```bash
go build ./internal/radiusd/...
```

## 代码统计

**累计新增代码**:

- Phase 1 (基础架构): ~750 行
- Phase 2 (插件实现): ~680 行
- Phase 3 (集成层): ~150 行
- **总计**: ~1580 行

**修改文件数**:

- 新建: 28 个文件
- 修改: 10+ 个文件

---

## 关键决策记录

1. **循环导入解决**: 创建独立 `errors` 包而非在 `radiusd` 中保留
2. **渐进式迁移**: 新旧方法并存，保证平滑过渡
3. **后备机制**: 如果没有插件注册，自动回退到原有逻辑
4. **类型适配**: 在集成层处理 `vendorparsers.VendorRequest` 和 `radiusd.VendorRequest` 的转换
