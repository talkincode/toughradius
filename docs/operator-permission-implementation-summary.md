# 操作员权限系统实现总结

## 完成的工作

### 后端（Go）

1. **权限检查**（`internal/adminapi/operators.go`）

   - ✅ `listOperators` - 只有 super/admin 可以查看操作员列表
   - ✅ `getOperator` - 只有 super/admin 可以查看操作员详情
   - ✅ `createOperator` - 只有 super 可以创建操作员
   - ✅ `updateOperator` - 只有 super/admin 可以更新，且不能修改自己的权限/状态
   - ✅ `deleteOperator` - 只有 super/admin 可以删除，且不能删除自己

2. **个人账号设置 API**
   - ✅ `GET /system/operators/me` - 获取当前登录用户信息
   - ✅ `PUT /system/operators/me` - 更新当前用户信息（不允许修改权限/状态）

### 前端（React + TypeScript）

1. **菜单权限控制**（`web/src/components/CustomMenu.tsx`）

   - ✅ 根据用户权限过滤菜单项
   - ✅ 普通操作员看不到"操作员管理"菜单

2. **AppBar 优化**（`web/src/components/CustomAppBar.tsx`）

   - ✅ "系统设置"按钮仅对 super/admin 显示
   - ✅ 新增"账号设置"按钮，所有用户可见

3. **操作员编辑表单优化**（`web/src/resources/operators.tsx`）

   - ✅ 根据用户权限隐藏/显示"权限设置"区域
   - ✅ 普通操作员看不到权限和状态字段
   - ✅ 管理员编辑自己时，权限/状态字段被禁用

4. **账号设置页面**（`web/src/pages/AccountSettings.tsx`）
   - ✅ 独立的账号设置页面（`/account/settings`）
   - ✅ 所有用户可编辑个人信息
   - ✅ 只读显示权限级别和状态
   - ✅ 完整的表单验证

## 核心特性

### 权限层级

- **超级管理员**：完全控制，可管理所有操作员（不能修改自己权限/删除自己）
- **管理员**：可管理普通操作员（不能修改自己权限/删除自己）
- **普通操作员**：只能通过"账号设置"编辑个人信息

### 安全机制

1. **双重保护**：前端隐藏 + 后端权限检查
2. **自我保护**：不能修改自己权限/状态，不能删除自己
3. **层级保护**：管理员无法删除超级管理员
4. **最小权限原则**：每个角色只能访问必要的功能

## 用户体验

### 超级管理员/管理员

```text
登录 → 看到完整菜单 →
  - 可访问"操作员管理" → 完整 CRUD 功能
  - 可访问"系统设置"
  - 可访问"账号设置" → 编辑个人信息
```

### 普通操作员

```text
登录 → 看到精简菜单 →
  - 没有"操作员管理"菜单
  - 没有"系统设置"按钮
  - 只能通过右上角"账号设置"按钮 → 编辑个人信息（不显示权限/状态）
```

## API 端点

| 端点                    | 方法   | 权限         | 说明                       |
| ----------------------- | ------ | ------------ | -------------------------- |
| `/system/operators`     | GET    | super, admin | 列表                       |
| `/system/operators/:id` | GET    | super, admin | 详情                       |
| `/system/operators`     | POST   | super        | 创建                       |
| `/system/operators/:id` | PUT    | super, admin | 更新（不能改自己权限）     |
| `/system/operators/:id` | DELETE | super, admin | 删除（不能删自己）         |
| `/system/operators/me`  | GET    | 所有用户     | 获取自己信息               |
| `/system/operators/me`  | PUT    | 所有用户     | 更新自己信息（不能改权限） |

## 文件清单

### 后端

- `internal/adminapi/operators.go` - 主要逻辑和权限控制
- `internal/adminapi/auth.go` - 用户上下文解析

### 前端

- `web/src/components/CustomMenu.tsx` - 菜单权限
- `web/src/components/CustomAppBar.tsx` - 顶栏按钮
- `web/src/resources/operators.tsx` - 操作员管理界面
- `web/src/pages/AccountSettings.tsx` - 账号设置页面
- `web/src/App.tsx` - 路由配置

## 测试状态

- ✅ 前端编译成功（`npm run build`）
- ⚠️ 后端测试需要更新（添加认证上下文模拟）

## 注意事项

1. **路由顺序**：`/system/operators/me` 必须在 `/:id` 之前注册
2. **测试更新**：现有测试需要添加 JWT 认证上下文才能通过
3. **权限验证**：所有权限检查在后端实现，前端仅用于 UI 优化

## 详细文档

完整的实现细节、安全特性、用户流程等详见：
`docs/operator-permission-system.md`
