# ToughRADIUS 前端国际化实施指南

## 概述

ToughRADIUS v9 前端使用 **React Admin 5.0** 的内置国际化支持，基于 **polyglot.js** 实现多语言功能。

## 架构

### 文件结构

```
web/src/i18n/
├── index.ts          # i18nProvider 配置，导出给 App.tsx 使用
├── zh-CN.ts          # 中文翻译
└── en-US.ts          # 英文翻译
```

### 核心配置

- **i18nProvider**: `web/src/i18n/index.ts`
- **默认语言**: 简体中文 (zh-CN)
- **支持语言**: 简体中文、English
- **语言切换**: 通过顶部工具栏的语言图标

## 使用方法

### 1. 在页面组件中使用翻译

```typescript
import { useTranslate } from "react-admin";

const MyComponent = () => {
  const translate = useTranslate();

  return (
    <div>
      <h1>{translate("dashboard.title")}</h1>
      <p>{translate("dashboard.subtitle")}</p>
    </div>
  );
};
```

### 2. 在资源定义中使用翻译

React Admin 会自动使用 `resources.<resource-name>.name` 作为资源显示名称：

```typescript
// App.tsx
<Resource
  name="radius/users"
  list={RadiusUserList}
  // 不需要手动指定 label，React Admin 会自动从翻译文件获取
/>
```

对应的翻译文件：

```typescript
// zh-CN.ts
resources: {
  'radius/users': {
    name: 'RADIUS用户 |||| RADIUS用户',  // 单数 |||| 复数
    fields: {
      username: '用户名',
      password: '密码',
      // ...
    },
  },
}
```

### 3. 字段标签自动翻译

React Admin 会自动翻译字段标签：

```typescript
// 在 List 或 Edit 组件中
<TextField source="username" />
// 自动使用 resources['radius/users'].fields.username
```

### 4. 验证消息翻译

```typescript
import { required, minLength, email } from "react-admin";

<TextInput source="email" validate={[required(), email()]} />;
// 验证消息自动从 validation.* 键获取
```

## 翻译文件结构

### 顶层结构

```typescript
{
  app: {},           // 应用级文本
  auth: {},          // 认证相关
  menu: {},          // 菜单项
  dashboard: {},     // 仪表盘
  resources: {},     // 资源定义
  validation: {},    // 验证消息
  // React Admin 内置键（从 ra-language-chinese 继承）
  ra: {},
}
```

### Resources 结构示例

```typescript
resources: {
  'radius/users': {
    name: 'RADIUS用户 |||| RADIUS用户',
    fields: {
      id: 'ID',
      username: '用户名',
      realname: '真实姓名',
      // ...
    },
    status_enabled: '启用',
    status_disabled: '禁用',
    sections: {
      basic: '基本信息',
      basic_desc: '用户的基础识别信息',
    },
  },
}
```

## 添加新语言

### 1. 创建翻译文件

```bash
cp web/src/i18n/zh-CN.ts web/src/i18n/ja-JP.ts
# 编辑 ja-JP.ts 添加日文翻译
```

### 2. 更新 i18nProvider

```typescript
// web/src/i18n/index.ts
import jaJP from "./ja-JP";

const translations = {
  "zh-CN": zhCN,
  "en-US": enUS,
  "ja-JP": jaJP, // 添加新语言
};

export const i18nProvider = polyglotI18nProvider(
  (locale) => translations[locale] || translations["zh-CN"],
  "zh-CN",
  [
    { locale: "zh-CN", name: "简体中文" },
    { locale: "en-US", name: "English" },
    { locale: "ja-JP", name: "日本語" }, // 添加语言选项
  ]
);
```

### 3. 更新语言切换器

编辑 `CustomAppBar.tsx` 添加新的 MenuItem。

## 最佳实践

### 1. 翻译键命名规范

- 使用小写字母和下划线
- 按功能模块分组
- 例如: `dashboard.total_users`, `resources.radius/users.fields.username`

### 2. 复数形式

React Admin 支持复数形式，使用 `||||` 分隔：

```typescript
name: '用户 |||| 用户',  // 中文单复数相同
name: 'User |||| Users', // 英文有单复数变化
```

### 3. 参数化翻译

```typescript
// 翻译文件
validation: {
  minLength: '最少需要 %{min} 个字符',
}

// 使用
translate('validation.minLength', { min: 6 })
// 输出: "最少需要 6 个字符"
```

### 4. 避免硬编码文本

❌ 错误示例:

```typescript
<Chip label="启用" />
```

✅ 正确示例:

```typescript
<Chip label={translate("resources.radius/users.status_enabled")} />
```

## 国际化检查清单

- [ ] 所有用户可见文本都已翻译
- [ ] 验证消息已配置
- [ ] 资源名称和字段标签已定义
- [ ] 所有支持语言都有完整翻译
- [ ] 测试语言切换功能
- [ ] 检查文本长度在不同语言下的显示效果

## 故障排查

### 翻译未生效

1. 检查翻译键是否正确
2. 确认 i18nProvider 已正确配置在 App.tsx
3. 检查浏览器控制台是否有警告

### 语言切换不生效

1. 清除浏览器缓存
2. 检查 localStorage 中的 locale 值
3. 确认 useSetLocale 正确调用

### 部分文本显示为键值

说明该翻译键在翻译文件中不存在，需要添加对应的翻译。

## 已完成的国际化模块

✅ 登录页面 (LoginPage)
✅ 仪表盘 (Dashboard)  
✅ 语言切换组件 (CustomAppBar)
✅ 资源定义 (App.tsx)

## 待完成的国际化

- [ ] 账户设置页面
- [ ] 系统配置页面
- [ ] 各资源模块的详细表单
- [ ] 错误提示消息
- [ ] 成功提示消息

## 参考资料

- [React Admin 国际化文档](https://marmelab.com/react-admin/Translation.html)
- [Polyglot.js 文档](https://airbnb.io/polyglot.js/)
- [ra-language-chinese](https://github.com/marmelab/react-admin/tree/master/packages/ra-language-chinese)
