---
name: add-config-schema
description: 新增可在系统配置页编辑的动态配置项 (TR-F014)。需要新增 RADIUS 运行参数并支持查询/编辑/reload 时使用。
---

# 技能：新增动态配置项

> 关联功能编号：`TR-F014`

## 何时使用
需要新增可在系统配置页编辑的 RADIUS 运行参数时。

## 前置检索
```text
view internal/app/config_schemas.json      # 配置 schema（先改这里）
view internal/app/config_manager.go        # 读取 / reload 逻辑
view internal/adminapi/settings.go         # 配置接口
grep_search "GetSettingsStringValue\|GetSettingsInt64Value" --include internal/**
```

## 实现步骤
1. **Schema**：在 `config_schemas.json` 新增配置项，提供 `默认值 / 类型 / 范围 / 国际化 key`。
2. **读取**：通过 `app.GApp().GetSettings*Value("<group>", "<Key>")` 读取，不要硬编码。
3. **默认值初始化**：确认 `checkSettings()` / 初始化逻辑会写入默认值。
4. **前端**：系统配置页 `web/src/pages/SystemConfigPage.tsx` 自动按 schema 渲染；如需特殊控件再定制。

## 约定
- 新配置项必须先进 schema 再被代码读取。
- 类型、范围、默认值、国际化 key 缺一不可。

## 验收
- [ ] 配置可在后台查询 / 编辑 / reload
- [ ] 新增配置读取路径有测试
- [ ] `go test ./internal/app/...` 通过
- [ ] PR 引用 `TR-F014`
