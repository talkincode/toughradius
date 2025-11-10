# ToughRADIUS 前端国际化进度

## 已完成的工作

### ✅ 完全国际化的资源

#### 1. **radiusUsers** (RADIUS 用户管理)

- **文件**: `web/src/resources/radiusUsers.tsx`
- **状态**: 已完成
- **包含内容**:
  - 列表页面 (List)
  - 创建表单 (Create)
  - 编辑表单 (Edit)
  - 详情页面 (Show)
  - 所有字段标签、提示文本、验证消息
  - 状态显示组件

#### 2. **nodes** (网络节点管理)

- **文件**: `web/src/resources/nodes.tsx`
- **状态**: 已完成
- **包含内容**:
  - 列表页面
  - 创建/编辑表单
  - 详情页面
  - 标签显示组件

#### 3. **i18n 翻译文件**

- **文件**:
  - `web/src/i18n/zh-CN.ts` (中文)
  - `web/src/i18n/en-US.ts` (英文)
- **已添加翻译**:
  - `resources.radius/users.*` - RADIUS 用户
  - `resources.network/nodes.*` - 网络节点
  - `resources.system/operators.*` - 操作员管理
  - `resources.network/nas.*` - NAS 设备
  - `resources.radius/profiles.*` - RADIUS 配置

### ✅ 翻译键已准备（待应用到代码）

以下资源的翻译键已添加到 i18n 文件，需要更新 TSX 文件：

1. **operators** (操作员管理)
2. **nas** (NAS 设备)
3. **radiusProfiles** (RADIUS 配置/计费策略)

---

## 国际化模式与最佳实践

### 翻译键命名规范

```typescript
resources.{resource_name}.fields.{field_name}       // 字段标签
resources.{resource_name}.sections.{section_name}   // 表单分段标题
resources.{resource_name}.helpers.{field_name}      // 帮助文本
resources.{resource_name}.status.{status_value}     // 状态值
resources.{resource_name}.validation.{rule_name}    // 验证消息
```

### 代码实现模式

#### 1. 导入 useTranslate Hook

```tsx
import { ..., useTranslate } from 'react-admin';
```

#### 2. 在组件中使用

```tsx
export const UserList = () => {
  const translate = useTranslate();

  return (
    <List>
      <Datagrid>
        <TextField
          source="username"
          label={translate("resources.radius/users.fields.username")}
        />
      </Datagrid>
    </List>
  );
};
```

#### 3. 筛选器使用 Hook

由于筛选器是数组，需要包装在函数中：

```tsx
const useUserFilters = () => {
  const translate = useTranslate();
  return [
    <TextInput
      key="username"
      label={translate("resources.radius/users.fields.username")}
      source="username"
      alwaysOn
    />,
  ];
};

export const UserList = () => {
  const userFilters = useUserFilters();
  return <List filters={userFilters}>...</List>;
};
```

#### 4. 动态状态翻译

```tsx
const StatusField = () => {
  const record = useRecordContext();
  const translate = useTranslate();

  return (
    <Chip
      label={translate(`resources.radius/users.status.${record.status}`)}
      color={record.status === "enabled" ? "success" : "default"}
    />
  );
};
```

---

## 待完成的资源文件

以下资源文件需要应用已准备好的翻译键：

### 🔄 优先级：高

#### 1. operators.tsx (操作员管理)

- **复杂度**: 高 - 包含权限控制逻辑
- **翻译键状态**: ✅ 已完成
- **需要替换的内容**:
  - 验证规则消息 (validateUsername, validatePassword 等)
  - 表单字段标签和帮助文本
  - 权限级别选择项
  - 状态显示

**关键点**:

```tsx
// 将硬编码的验证规则改为使用 translate
const validateUsername = () => {
  const translate = useTranslate();
  return [
    required(
      translate("resources.system/operators.validation.username_required")
    ),
    minLength(
      3,
      translate("resources.system/operators.validation.username_min")
    ),
    // ...
  ];
};
```

#### 2. nas.tsx (NAS 设备管理)

- **复杂度**: 中
- **翻译键状态**: ✅ 已完成
- **参考**: radiusUsers.tsx 的实现模式

#### 3. radiusProfiles.tsx (RADIUS 配置)

- **复杂度**: 中
- **翻译键状态**: ✅ 已完成
- **参考**: radiusUsers.tsx 的实现模式

### 🔄 优先级：中

#### 4. accounting.tsx (计费日志)

- **复杂度**: 低 - 主要是列表展示
- **翻译键状态**: ✅ 已存在 (需检查完整性)

#### 5. onlineSessions.tsx (在线会话)

- **复杂度**: 低 - 主要是列表展示
- **翻译键状态**: ✅ 已存在 (需检查完整性)

---

## 剩余工作清单

### 立即可做

1. **operators.tsx**:

   - [ ] 更新验证规则使用 translate
   - [ ] 更新所有字段标签
   - [ ] 更新帮助文本
   - [ ] 更新权限级别选项
   - [ ] 更新状态显示组件

2. **nas.tsx**:

   - [ ] 添加 useTranslate 导入
   - [ ] 更新列表字段标签
   - [ ] 更新创建/编辑表单
   - [ ] 更新详情页面

3. **radiusProfiles.tsx**:
   - [ ] 添加 useTranslate 导入
   - [ ] 更新列表字段标签
   - [ ] 更新创建/编辑表单
   - [ ] 更新详情页面

### 需要检查

4. **accounting.tsx 和 onlineSessions.tsx**:
   - [ ] 检查 i18n 文件中的翻译键是否完整
   - [ ] 如果缺少，添加翻译键
   - [ ] 应用到组件

---

## 快速参考：完整示例

### 列表页面

```tsx
export const NasList = () => {
  const translate = useTranslate();
  const nasFilters = useNasFilters(); // 自定义 hook

  return (
    <List actions={<NasListActions />} filters={nasFilters}>
      <Datagrid rowClick="show">
        <TextField
          source="name"
          label={translate("resources.network/nas.fields.name")}
        />
        <TextField
          source="ip_addr"
          label={translate("resources.network/nas.fields.ip_addr")}
        />
        <DateField
          source="created_at"
          label={translate("resources.network/nas.fields.created_at")}
          showTime
        />
      </Datagrid>
    </List>
  );
};
```

### 创建/编辑表单

```tsx
export const NasEdit = () => {
  const translate = useTranslate();

  return (
    <Edit>
      <SimpleForm>
        <FormSection
          title={translate("resources.network/nas.sections.basic")}
          description={translate("resources.network/nas.sections.basic_desc")}
        >
          <TextInput
            source="name"
            label={translate("resources.network/nas.fields.name")}
            helperText={translate("resources.network/nas.helpers.name")}
            validate={[required()]}
          />
        </FormSection>
      </SimpleForm>
    </Edit>
  );
};
```

### 详情页面

```tsx
export const NasShow = () => {
  const translate = useTranslate();

  return (
    <Show>
      <Card>
        <CardContent>
          <Typography variant="h6">
            {translate("resources.network/nas.sections.basic")}
          </Typography>
          <TableContainer>
            <Table>
              <TableBody>
                <DetailRow
                  label={translate("resources.network/nas.fields.name")}
                  value={<TextField source="name" />}
                />
              </TableBody>
            </Table>
          </TableContainer>
        </CardContent>
      </Card>
    </Show>
  );
};
```

---

## 测试建议

完成每个资源的国际化后：

1. **功能测试**:

   - 切换语言 (中文 <-> English)
   - 检查所有页面是否正确显示
   - 检查创建/编辑表单是否正常工作

2. **视觉检查**:

   - 长文本是否溢出
   - 布局是否正常
   - 翻译是否准确

3. **控制台检查**:
   - 无翻译键缺失警告
   - 无 TypeScript 错误

---

## 进度追踪

| 资源           | TSX 文件 | 中文翻译 | 英文翻译 | 状态      |
| -------------- | -------- | -------- | -------- | --------- |
| radiusUsers    | ✅       | ✅       | ✅       | ✅ 完成   |
| nodes          | ✅       | ✅       | ✅       | ✅ 完成   |
| operators      | ⏳       | ✅       | ✅       | 🔄 进行中 |
| nas            | ⏳       | ✅       | ✅       | ⏳ 待开始 |
| radiusProfiles | ⏳       | ✅       | ✅       | ⏳ 待开始 |
| accounting     | ⏳       | ✅       | ✅       | ⏳ 待开始 |
| onlineSessions | ⏳       | ✅       | ✅       | ⏳ 待开始 |

---

## 总结

**已完成**: 2/7 资源 (29%)
**翻译键准备完成**: 7/7 (100%)
**预计剩余时间**: 2-3 小时

所有翻译键已经准备完毕，剩余工作主要是将这些翻译键应用到 TSX 文件中。
按照本文档提供的模式，可以系统地完成剩余资源的国际化工作。
