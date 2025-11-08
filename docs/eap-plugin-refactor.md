# EAP 插件系统重构文档

## 概述

本文档记录了 ToughRADIUS v9 中 EAP (Extensible Authentication Protocol) 认证模块的插件化重构过程。

## 重构目标

1. **解耦 EAP 认证逻辑**: 将硬编码在 `radius_auth.go` 中的 EAP 处理逻辑提取为独立的插件系统
2. **提高可扩展性**: 支持通过插件方式轻松添加新的 EAP 认证方法
3. **改善代码结构**: 遵循单一职责原则,每个 EAP 方法独立实现
4. **便于测试**: 每个 EAP 处理器可以独立测试

## 架构设计

### 核心组件

```
internal/radiusd/plugins/eap/
├── interfaces.go           # EAP 插件接口定义
├── errors.go              # EAP 错误定义
├── utils.go               # EAP 工具函数
├── coordinator.go         # EAP 协调器
├── statemanager/          # 状态管理器
│   └── memory_state_manager.go
└── handlers/              # EAP 处理器实现
    ├── md5_handler.go     # EAP-MD5
    ├── otp_handler.go     # EAP-OTP
    └── mschapv2_handler.go (待实现)
```

### 关键接口

#### 1. EAPHandler 接口

```go
type EAPHandler interface {
    Name() string
    EAPType() uint8
    CanHandle(ctx *EAPContext) bool
    HandleIdentity(ctx *EAPContext) (bool, error)
    HandleResponse(ctx *EAPContext) (bool, error)
}
```

每个 EAP 认证方法实现此接口:

- `HandleIdentity`: 处理 EAP-Response/Identity,发送 Challenge
- `HandleResponse`: 处理 Challenge Response,验证密码

#### 2. EAPStateManager 接口

```go
type EAPStateManager interface {
    GetState(stateID string) (*EAPState, error)
    SetState(stateID string, state *EAPState) error
    DeleteState(stateID string) error
}
```

管理 EAP 认证会话状态。当前实现:

- `MemoryStateManager`: 基于内存的状态管理器

#### 3. EAPCoordinator 协调器

```go
type Coordinator struct {
    stateManager    EAPStateManager
    pwdProvider     PasswordProvider
    handlerRegistry HandlerRegistry
}
```

职责:

- 解析 EAP 消息
- 根据 EAP Type 分发到相应的处理器
- 处理 Identity 和 Nak 消息
- 发送 Success/Failure 响应

## 已实现的 EAP 方法

### 1. EAP-MD5 (RFC 2284)

- **处理器**: `MD5Handler`
- **文件**: `handlers/md5_handler.go`
- **认证流程**:
  1. 收到 Identity -> 生成随机 Challenge (16 字节)
  2. 收到 Response -> 验证 MD5(identifier + password + challenge)

**使用示例**:

```yaml
# toughradius.yml
radius:
  eap_method: "eap-md5"
```

### 2. EAP-OTP (RFC 2284)

- **处理器**: `OTPHandler`
- **文件**: `handlers/otp_handler.go`
- **认证流程**:
  1. 收到 Identity -> 发送 OTP 提示消息
  2. 收到 Response -> 验证 OTP (TODO: 集成 TOTP/HOTP)

**使用示例**:

```yaml
# toughradius.yml
radius:
  eap_method: "eap-otp"
```

### 3. EAP-MSCHAPv2 (RFC 2759)

- **处理器**: `MSCHAPv2Handler`
- **文件**: `handlers/mschapv2_handler.go`
- **认证流程**:
  1. 收到 Identity -> 生成随机 Authenticator Challenge (16 字节)
  2. 收到 Response -> 解析 Peer-Challenge 和 NT-Response
  3. 使用 RFC 2759 算法验证 NT-Response
  4. 验证成功后生成 MPPE 加密密钥 (RFC 3079)
  5. 返回 Authenticator-Response 和加密密钥

**特性**:

- 完整支持 MS-CHAPv2 协议 (RFC 2759)
- 自动生成 MPPE 加密密钥 (Send-Key 和 Recv-Key)
- 支持 Microsoft 厂商特定属性
- 包含完整的单元测试覆盖

**使用示例**:

```yaml
# toughradius.yml
radius:
  eap_method: "eap-mschapv2"
```

**测试**:

```bash
go test -v ./internal/radiusd/plugins/eap/handlers/... -run TestMSCHAPv2Handler
```

## 插件注册

所有 EAP 处理器在 `internal/radiusd/plugins/init.go` 中注册:

```go
func InitPlugins(sessionRepo repository.SessionRepository, accountingRepo repository.AccountingRepository) {
    // ... 其他插件注册

    // 注册 EAP 处理器
    registry.RegisterEAPHandler(eaphandlers.NewMD5Handler())
    registry.RegisterEAPHandler(eaphandlers.NewOTPHandler())
    registry.RegisterEAPHandler(eaphandlers.NewMSCHAPv2Handler())
}
```

## 集成方式

### 在 RadiusService 中使用

```go
import (
    "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
    "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/statemanager"
    "github.com/talkincode/toughradius/v9/internal/radiusd/registry"
)

// 创建 EAP 协调器
stateManager := statemanager.NewMemoryStateManager()
pwdProvider := &YourPasswordProvider{} // 实现 PasswordProvider 接口
handlerRegistry := registry.GetGlobalRegistry()

eapCoordinator := eap.NewCoordinator(stateManager, pwdProvider, handlerRegistry)

// 处理 EAP 请求
configuredMethod := app.GApp().GetSettingsStringValue("radius", "EapMethod")
response := r.Response(radius.CodeAccessAccept)
handled, success, err := eapCoordinator.HandleEAPRequest(
    w, r, user, nas, response, secret, isMacAuth, configuredMethod,
)

if handled {
    if success {
        eapCoordinator.SendEAPSuccess(w, r, response, secret)
    } else {
        // 发送 EAP-Failure + RADIUS Access-Reject
        eapCoordinator.SendEAPFailure(w, r, secret, err)
    }
    // 清理状态
    eapCoordinator.CleanupState(r)
}
```

## 添加新的 EAP 方法

### 步骤 1: 实现 EAPHandler 接口

```go
// handlers/new_eap_handler.go
package handlers

import "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"

type NewEAPHandler struct{}

func NewNewEAPHandler() *NewEAPHandler {
    return &NewEAPHandler{}
}

func (h *NewEAPHandler) Name() string {
    return "eap-new"
}

func (h *NewEAPHandler) EAPType() uint8 {
    return eap.TypeXXX // 定义新的类型
}

func (h *NewEAPHandler) CanHandle(ctx *eap.EAPContext) bool {
    return ctx.EAPMessage.Type == eap.TypeXXX
}

func (h *NewEAPHandler) HandleIdentity(ctx *eap.EAPContext) (bool, error) {
    // 实现 Identity 处理逻辑
}

func (h *NewEAPHandler) HandleResponse(ctx *eap.EAPContext) (bool, error) {
    // 实现 Response 验证逻辑
}
```

### 步骤 2: 注册处理器

```go
// plugins/init.go
registry.RegisterEAPHandler(handlers.NewNewEAPHandler())
```

### 步骤 3: 配置使用

```yaml
# toughradius.yml
radius:
  eap_method: "eap-new"
```

## 待完成工作

### 1. ~~EAP-MSCHAPv2 插件化~~ (✅ 已完成)

- ✅ 文件: `handlers/mschapv2_handler.go`
- ✅ 已实现完整的 MS-CHAPv2 协议支持 (RFC 2759)
- ✅ 已实现 MPPE 密钥生成 (RFC 3079)
- ✅ 已添加完整的单元测试

### 2. ~~集成到 AuthService~~ (✅ 已完成)

- ✅ `radius_auth.go` 中的 EAP 处理逻辑完全交由 `EAPAuthHelper` + `Coordinator`
- ✅ `AuthenticateUserWithPlugins()` 支持 `SkipPasswordValidation()`，EAP 成功后只运行策略插件
- ✅ `SendAccept/SendReject` 统一使用新的状态管理器清理会话

### 3. ~~添加单元测试~~ (✅ 部分完成)

- ✅ EAP-MD5 处理器测试
- ✅ EAP-MSCHAPv2 处理器测试
- TODO: EAP-OTP 处理器测试
- TODO: Coordinator 的消息分发逻辑测试
- TODO: StateManager 的并发安全性测试

### 4. 支持更多 EAP 方法 (优先级: 低)

- EAP-TLS
- EAP-PEAP
- EAP-TTLS
- EAP-FAST

### 5. 状态持久化 (优先级: 低)

- 实现 RedisStateManager
- 支持集群部署时的状态共享

## 性能考虑

1. **内存占用**: MemoryStateManager 使用 map 存储状态,需要定期清理过期状态
2. **并发安全**: 使用 sync.RWMutex 保护状态访问
3. **查找性能**: Handler 查找使用 map,O(1) 复杂度

## 向后兼容

- 保留 `radius_eap.go` 中的原有函数作为适配器
- 新功能使用插件系统
- 旧代码逐步迁移

## 相关文件

### 新增文件

- `internal/radiusd/plugins/eap/interfaces.go`
- `internal/radiusd/plugins/eap/errors.go`
- `internal/radiusd/plugins/eap/utils.go`
- `internal/radiusd/plugins/eap/coordinator.go`
- `internal/radiusd/plugins/eap/password_provider.go`
- `internal/radiusd/plugins/eap/statemanager/memory_state_manager.go`
- `internal/radiusd/plugins/eap/handlers/md5_handler.go`
- `internal/radiusd/plugins/eap/handlers/md5_handler_test.go`
- `internal/radiusd/plugins/eap/handlers/otp_handler.go`
- `internal/radiusd/plugins/eap/handlers/mschapv2_handler.go` ✅
- `internal/radiusd/plugins/eap/handlers/mschapv2_handler_test.go` ✅
- `internal/radiusd/eap_helper.go`

### 修改文件

- `internal/radiusd/plugins/init.go` (添加 EAP 插件注册) ✅
- `internal/radiusd/registry/registry.go` (添加 GetHandler 方法)

### 新增修改

- `internal/radiusd/radius_auth.go`（接入 EAP 协调器、移除旧的 switch-case）
- `internal/radiusd/auth_plugin_integration.go`（新增认证选项机制）
- `internal/radiusd/eap_helper.go`（向协调器传递 RadiusUser/NAS/Response）
- `internal/radiusd/plugins/eap/coordinator.go`（上下文补全 user/nas/response）

## 参考资料

- RFC 2284: PPP Extensible Authentication Protocol (EAP)
- RFC 2869: RADIUS Extensions
- RFC 2759: Microsoft PPP CHAP Extensions, Version 2
- RFC 3079: Deriving Keys for use with Microsoft Point-to-Point Encryption (MPPE)
- RFC 3748: Extensible Authentication Protocol (EAP)
- [ToughRADIUS 重构设计文档](./radiusd-refactor-design.md)

## 维护者

- 初始实现: @github-copilot
- MSCHAPv2 实现: @github-copilot (2025-01-08)
- 日期: 2025-11-08
- 版本: v9dev
