---
name: add-react-admin-resource
description: 在 React Admin 管理后台新增资源或页面 (TR-F013)。后端新增 Admin API 后需在前端暴露对应管理界面时使用。
---

# 技能：新增前端管理资源 / 页面

> 关联功能编号：`TR-F013`　适用里程碑：M2 等

## 何时使用
后端新增 Admin API 后，需要在 React Admin 后台暴露对应资源或页面时。

## 前置检索
```text
view web/src/App.tsx                       # 资源 / 路由注册
file_search "web/src/resources/*.tsx"      # 资源范本
view web/src/providers/dataProvider.ts     # REST 映射
view web/src/resources/nodes.tsx           # 标准资源范本
```

## 实现步骤
1. **资源文件**：在 `web/src/resources/<feature>.tsx` 定义 List/Edit/Create（模仿 `nodes.tsx`）。
2. **注册资源**：在 `web/src/App.tsx` 加入 `<Resource name="<api-path>" .../>`，name 对齐后端 API 路径。
3. **数据映射**：确认 `dataProvider.ts` 的分页 / 过滤 / 排序参数与后端一致。
4. **页面（如非标准 CRUD）**：放在 `web/src/pages/`，不要新建独立管理入口。
5. **国际化**：如项目有 i18n key，补齐对应文案。

## 边界
- 不引入独立管理入口；所有页面挂在统一 Admin 框架下。
- 前端只暴露后端已支持且可验证的安全动作（尤其 CoA 等），禁止前端拼装协议包。

## 验收
- [ ] `cd web && npm run build` 成功
- [ ] 列表 / 过滤 / 增删改与后端联通
- [ ] PR 引用 `TR-F013` 与里程碑编号
