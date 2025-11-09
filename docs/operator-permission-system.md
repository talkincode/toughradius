# 操作员权限系统实现文档

## 概述

本文档记录了 ToughRADIUS v9 操作员权限系统的完整实现，包括后端权限控制和前端用户界面优化。

## 一、系统架构

### 权限级别

系统定义了三个权限级别：

1. **super（超级管理员）**

   - 拥有所有系统权限
   - 可以查看和管理所有操作员
   - 可以修改其他操作员的所有信息（包括权限和状态）
   - **不能**修改自己的权限和状态
   - **不能**删除自己的账号

2. **admin（管理员）**

   - 可以查看和管理操作员
   - 可以修改普通操作员的信息
   - **不能**修改超级管理员的权限
   - **不能**删除超级管理员
   - **不能**修改自己的权限和状态

3. **operator（普通操作员）**
   - **不能**访问操作员管理功能
   - 只能通过"账号设置"编辑自己的信息
   - **不能**查看或修改自己的权限和状态

## 二、后端实现

### 文件：`internal/adminapi/operators.go`

#### 1. 路由注册

```go
func registerOperatorsRoutes() {
    // 个人账号设置路由 - 必须在 :id 路由之前注册
    webserver.ApiGET("/system/operators/me", getCurrentOperator)
    webserver.ApiPUT("/system/operators/me", updateCurrentOperator)

    // 操作员管理路由
    webserver.ApiGET("/system/operators", listOperators)
    webserver.ApiGET("/system/operators/:id", getOperator)
    webserver.ApiPOST("/system/operators", createOperator)
    webserver.ApiPUT("/system/operators/:id", updateOperator)
    webserver.ApiDELETE("/system/operators/:id", deleteOperator)
}
```

**注意**：`/me` 路由必须在 `/:id` 之前注册，否则会被 `/:id` 捕获。

#### 2. 权限检查实现

##### listOperators - 列表查询权限

```go
// 只有超级管理员和管理员可以查看操作员列表
if currentOpr.Level != "super" && currentOpr.Level != "admin" {
    return fail(c, http.StatusForbidden, "PERMISSION_DENIED", "只有超级管理员和管理员可以查看操作员列表", nil)
}
```

##### getOperator - 详情查询权限

```go
// 只有超级管理员和管理员可以查看操作员详情
if currentOpr.Level != "super" && currentOpr.Level != "admin" {
    return fail(c, http.StatusForbidden, "PERMISSION_DENIED", "只有超级管理员和管理员可以查看操作员详情", nil)
}
```

##### createOperator - 创建操作员权限

```go
// 只有超级管理员可以创建操作员
if currentOpr.Level != "super" {
    return fail(c, http.StatusForbidden, "PERMISSION_DENIED", "只有超级管理员可以创建操作员", nil)
}
```

##### updateOperator - 更新操作员权限

```go
// 只有超级管理员和管理员可以更新操作员
if currentOpr.Level != "super" && currentOpr.Level != "admin" {
    return fail(c, http.StatusForbidden, "PERMISSION_DENIED", "只有超级管理员和管理员可以更新操作员", nil)
}

// 如果是修改自己，不允许修改权限和状态
isEditingSelf := currentOpr.ID == id
if isEditingSelf && (payload.Level != "" || payload.Status != "") {
    return fail(c, http.StatusForbidden, "CANNOT_MODIFY_SELF_PERMISSION", "不能修改自己的权限和状态", nil)
}
```

##### deleteOperator - 删除操作员权限

```go
// 只有超级管理员和管理员可以删除操作员
if currentOpr.Level != "super" && currentOpr.Level != "admin" {
    return fail(c, http.StatusForbidden, "PERMISSION_DENIED", "只有超级管理员和管理员可以删除操作员", nil)
}

// 不能删除自己
if currentOpr.ID == id {
    return fail(c, http.StatusForbidden, "CANNOT_DELETE_SELF", "不能删除自己的账号", nil)
}

// 只有超级管理员才能删除超级管理员账号
if targetOpr.Level == "super" && currentOpr.Level != "super" {
    return fail(c, http.StatusForbidden, "PERMISSION_DENIED", "只有超级管理员才能删除超级管理员账号", nil)
}
```

#### 3. 个人账号设置 API

##### getCurrentOperator - 获取当前用户信息

```go
func getCurrentOperator(c echo.Context) error {
    currentOpr, err := resolveOperatorFromContext(c)
    if err != nil {
        return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "无法获取当前用户信息", nil)
    }
    return ok(c, currentOpr)
}
```

##### updateCurrentOperator - 更新当前用户信息

```go
func updateCurrentOperator(c echo.Context) error {
    currentOpr, err := resolveOperatorFromContext(c)
    if err != nil {
        return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "无法获取当前用户信息", nil)
    }

    var payload operatorPayload
    if err := c.Bind(&payload); err != nil {
        return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "无法解析操作员参数", nil)
    }

    // 不允许修改权限和状态
    if payload.Level != "" || payload.Status != "" {
        return fail(c, http.StatusForbidden, "CANNOT_MODIFY_PERMISSION", "不能修改权限和状态", nil)
    }

    // ... 其他字段更新逻辑 ...
}
```

**限制**：此接口不允许修改 `level` 和 `status` 字段。

## 三、前端实现

### 1. 菜单权限控制

**文件**：`web/src/components/CustomMenu.tsx`

```tsx
const menuItems = [
  { to: "/", label: "控制台", icon: <DashboardOutlinedIcon /> },
  // ... 其他菜单项 ...
  {
    to: "/system/operators",
    label: "操作员管理",
    icon: <AdminPanelSettingsOutlinedIcon />,
    permissions: ["super", "admin"], // 仅超级管理员和管理员可见
  },
];

// 根据用户权限过滤菜单项
const filteredMenuItems = menuItems.filter((item) => {
  if (!item.permissions) return true;
  if (!identity?.level) return false;
  return item.permissions.includes(identity.level);
});
```

### 2. AppBar 权限控制

**文件**：`web/src/components/CustomAppBar.tsx`

```tsx
{
  /* 系统设置 - 仅超级管理员和管理员可见 */
}
{
  identity?.level === "super" || identity?.level === "admin" ? (
    <Tooltip title="系统设置">
      <IconButton onClick={() => redirect("/system/settings")}>
        <SettingsOutlinedIcon />
      </IconButton>
    </Tooltip>
  ) : null;
}

{
  /* 账号设置 - 所有用户可见 */
}
<Tooltip title="账号设置">
  <IconButton onClick={() => redirect("/account/settings")}>
    <AccountCircleOutlinedIcon />
  </IconButton>
</Tooltip>;
```

### 3. 操作员编辑表单权限控制

**文件**：`web/src/resources/operators.tsx`

```tsx
export const OperatorEdit = () => {
  const { identity } = useGetIdentity();
  const record = useRecordContext();

  const isEditingSelf =
    identity && record && String(identity.id) === String(record.id);
  const canManagePermissions =
    identity?.level === "super" || identity?.level === "admin";

  return (
    <Edit>
      <SimpleForm>
        {/* 基本信息和个人信息 - 所有人可见 */}

        {/* 权限设置 - 仅超级管理员和管理员可见 */}
        {canManagePermissions && (
          <FormSection title="权限设置">
            <SelectInput
              source="level"
              disabled={isEditingSelf} // 不能修改自己的权限
            />
            <SelectInput
              source="status"
              disabled={isEditingSelf} // 不能修改自己的状态
            />
          </FormSection>
        )}
      </SimpleForm>
    </Edit>
  );
};
```

### 4. 账号设置页面

**文件**：`web/src/pages/AccountSettings.tsx`

独立的账号设置页面，所有用户可访问，用于编辑个人信息：

- **可编辑**：用户名、密码、真实姓名、邮箱、手机号、备注
- **只读显示**：权限级别、账号状态
- **路由**：`/account/settings`

### 5. 路由配置

**文件**：`web/src/App.tsx`

```tsx
<Admin>
  {/* ... 其他 Resource ... */}

  {/* 自定义路由 - 账号设置 */}
  <CustomRoutes>
    <Route path="/account/settings" element={<AccountSettings />} />
  </CustomRoutes>
</Admin>
```

## 四、用户体验流程

### 超级管理员

1. 登录后可看到完整菜单，包括"操作员管理"
2. 可通过菜单进入操作员管理页面
3. 可创建、查看、编辑、删除操作员
4. 编辑其他操作员时，可修改所有字段
5. 编辑自己时，权限和状态字段被禁用
6. 无法删除自己的账号
7. 可通过右上角"账号设置"按钮编辑个人信息

### 管理员

1. 登录后可看到完整菜单，包括"操作员管理"
2. 可查看、编辑、删除普通操作员
3. 不能删除超级管理员
4. 编辑自己时，权限和状态字段被禁用
5. 无法删除自己的账号
6. 可通过右上角"账号设置"按钮编辑个人信息

### 普通操作员

1. 登录后菜单中**没有**"操作员管理"选项
2. 顶部 AppBar 中**没有**"系统设置"按钮
3. 只能通过右上角"账号设置"按钮访问个人信息编辑
4. 在账号设置页面中：
   - 可以修改：用户名、密码、真实姓名、邮箱、手机号、备注
   - 只能查看：权限级别、账号状态

## 五、安全特性

1. **后端强制验证**：所有权限检查在后端实现，无法通过前端绕过
2. **双重保护**：前端隐藏+后端拒绝，防止直接 API 调用
3. **自我保护**：管理员无法修改自己的权限或删除自己
4. **层级保护**：管理员无法删除超级管理员
5. **最小权限原则**：普通操作员只能访问自己的账号信息

## 六、API 端点总结

| 端点                    | 方法   | 权限要求       | 说明                                  |
| ----------------------- | ------ | -------------- | ------------------------------------- |
| `/system/operators`     | GET    | super, admin   | 获取操作员列表                        |
| `/system/operators/:id` | GET    | super, admin   | 获取操作员详情                        |
| `/system/operators`     | POST   | super          | 创建操作员                            |
| `/system/operators/:id` | PUT    | super, admin   | 更新操作员（不能修改自己的权限/状态） |
| `/system/operators/:id` | DELETE | super, admin   | 删除操作员（不能删除自己）            |
| `/system/operators/me`  | GET    | 所有已认证用户 | 获取当前用户信息                      |
| `/system/operators/me`  | PUT    | 所有已认证用户 | 更新当前用户信息（不能修改权限/状态） |

## 七、前端页面总结

| 路由                         | 权限要求       | 说明           |
| ---------------------------- | -------------- | -------------- |
| `/system/operators`          | super, admin   | 操作员管理列表 |
| `/system/operators/:id`      | super, admin   | 操作员详情     |
| `/system/operators/:id/edit` | super, admin   | 编辑操作员     |
| `/system/operators/create`   | super          | 创建操作员     |
| `/account/settings`          | 所有已认证用户 | 账号设置       |

## 八、错误码

| 错误码                          | HTTP 状态 | 说明                          |
| ------------------------------- | --------- | ----------------------------- |
| `PERMISSION_DENIED`             | 403       | 权限不足                      |
| `UNAUTHORIZED`                  | 401       | 未认证                        |
| `CANNOT_DELETE_SELF`            | 403       | 不能删除自己                  |
| `CANNOT_MODIFY_SELF_PERMISSION` | 403       | 不能修改自己的权限/状态       |
| `CANNOT_MODIFY_PERMISSION`      | 403       | 不能通过账号设置修改权限/状态 |

## 九、测试建议

由于引入了权限检查，现有测试需要更新：

1. **添加认证上下文**：测试中需要模拟 JWT token 和用户上下文
2. **分权限测试**：为不同权限级别编写独立测试用例
3. **边界测试**：测试权限边界（如普通操作员访问管理接口应返回 403）
4. **自我保护测试**：测试管理员修改自己权限应被拒绝

## 十、未来优化建议

1. **细粒度权限**：引入基于资源的权限控制（RBAC）
2. **审计日志**：记录所有权限相关操作
3. **权限缓存**：减少数据库查询
4. **动态权限**：支持运行时权限配置
5. **批量操作**：批量修改操作员时的权限检查

## 十一、相关文件清单

### 后端文件

- `internal/adminapi/operators.go` - 操作员 CRUD 和权限控制
- `internal/adminapi/auth.go` - 认证和用户上下文解析
- `internal/adminapi/adminapi.go` - 路由注册

### 前端文件

- `web/src/components/CustomMenu.tsx` - 菜单权限控制
- `web/src/components/CustomAppBar.tsx` - AppBar 按钮权限控制
- `web/src/resources/operators.tsx` - 操作员管理界面
- `web/src/pages/AccountSettings.tsx` - 账号设置页面
- `web/src/providers/authProvider.ts` - 认证提供者
- `web/src/App.tsx` - 路由配置

---

**文档版本**：1.0  
**最后更新**：2025 年 1 月  
**维护者**：ToughRADIUS 开发团队
