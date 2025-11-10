# 资源文件国际化快速指南

## 已完成的国际化

### 核心基础设施

- ✅ i18nProvider 配置
- ✅ 翻译文件（zh-CN.ts, en-US.ts）
- ✅ 语言切换组件

### 页面

- ✅ LoginPage - 登录页
- ✅ Dashboard - 仪表盘

### 资源

- ✅ App.tsx - 移除资源 label（使用自动翻译）
- ✅ onlineSessions.tsx - 在线会话（部分完成）
- ✅ operators.tsx - 操作员（部分完成，状态和级别组件已翻译）

## React Admin 字段自动翻译原理

React Admin 会自动从翻译文件获取字段标签，规则如下：

```typescript
// 翻译文件定义
resources: {
  'radius/users': {
    name: 'RADIUS用户 |||| RADIUS用户',
    fields: {
      username: '用户名',
      email: '邮箱',
    },
  },
}

// 组件中使用（无需 label）
<TextField source="username" />
// 自动显示：用户名

<EmailField source="email" />
// 自动显示：邮箱
```

## 快速国际化步骤

### 1. 删除硬编码 label

**错误示例：**

```tsx
<TextField source="username" label="用户名" />
<DateField source="created_at" label="创建时间" showTime />
```

**正确示例：**

```tsx
<TextField source="username" />
<DateField source="created_at" showTime />
```

### 2. 状态值使用 useTranslate

**错误示例：**

```tsx
const StatusField = () => {
  const record = useRecordContext();
  return <Chip label={record.status === "enabled" ? "启用" : "禁用"} />;
};
```

**正确示例：**

```tsx
const StatusField = () => {
  const record = useRecordContext();
  const translate = useTranslate();
  return <Chip label={translate(`common.${record.status}`)} />;
};
```

### 3. 选择框 choices 使用翻译

**方法一：直接在翻译文件中定义 choices**

```typescript
// 不推荐在组件中硬编码
<SelectInput
  source="status"
  choices={[
    { id: 'enabled', name: '启用' },
    { id: 'disabled', name: '禁用' },
  ]}
/>

// 推荐：使用 translateChoice
<SelectInput
  source="status"
  choices={[
    { id: 'enabled', name: 'enabled' },
    { id: 'disabled', name: 'disabled' },
  ]}
  translateChoice
/>
```

### 4. 表单验证消息

**错误示例：**

```tsx
const validateUsername = [
  required("用户名不能为空"),
  minLength(3, "用户名长度至少3个字符"),
];
```

**正确示例：**

```tsx
const validateUsername = [
  required(), // 使用翻译文件中的 ra.validation.required
  minLength(3), // 使用 ra.validation.minLength，自动填充参数
];
```

## 待处理文件清单

### 资源模块（移除 label）

#### radiusUsers.tsx

- [ ] 列表字段移除 label
- [ ] 表单字段移除 label
- [ ] 筛选器移除 label
- [ ] StatusField 使用 translate

#### nas.tsx

- [ ] 列表字段移除 label
- [ ] 表单字段移除 label
- [ ] 筛选器移除 label

#### radiusProfiles.tsx

- [ ] 列表字段移除 label
- [ ] 表单字段移除 label

#### accounting.tsx

- [ ] 列表字段移除 label
- [ ] 详情字段移除 label

#### nodes.tsx

- [ ] 列表字段移除 label
- [ ] 表单字段移除 label
- [ ] 筛选器移除 label

### 页面组件

#### AccountSettings.tsx

- [ ] 页面标题使用 translate
- [ ] 表单标签使用 translate
- [ ] 按钮文本使用 translate
- [ ] 状态显示使用 translate

#### SystemConfigPage.tsx

- [ ] 页面标题使用 translate
- [ ] 配置项使用 translate

## 批量处理技巧

### 1. 使用正则表达式批量查找

在 VSCode 中：

- 查找：`label="([^"]+)"`
- 这会找到所有硬编码的 label

### 2. 验证翻译键是否存在

检查翻译文件中是否有对应的键：

```
resources['radius/users'].fields.username
```

### 3. 测试国际化

1. 启动开发服务器
2. 使用语言切换器切换语言
3. 检查所有文本是否正确翻译
4. 查看浏览器控制台是否有缺失翻译的警告

## 常见问题

### Q: 删除 label 后字段不显示文本？

A: 确保翻译文件中有对应的 `fields.{fieldName}` 定义

### Q: 如何处理动态文本？

A: 使用 `useTranslate` hook 和参数化翻译

```tsx
const translate = useTranslate();
<span>
  {translate("dashboard.online_count", { count: stats.online_users })}
</span>;
```

### Q: 如何处理表单 section 标题？

A: 将标题添加到翻译文件，然后使用 `translate`

```tsx
const translate = useTranslate();
<FormSection title={translate("resources.radius/users.sections.basic")} />;
```

## 下一步行动

1. 完成所有资源模块的 label 移除
2. 更新 AccountSettings 和 SystemConfigPage
3. 测试所有页面的中英文切换
4. 补充缺失的翻译键
5. 更新文档

## 自动化脚本

可以编写脚本自动移除 label 属性：

```bash
# 示例：移除简单的 label 属性
sed -i '' 's/ label="[^"]*"//g' file.tsx

# 注意：需要手动检查结果，某些 label 可能需要保留
```
