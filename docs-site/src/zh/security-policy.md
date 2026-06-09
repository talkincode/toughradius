# 安全策略

> English version: [Security Policy](../en/security-policy.md)

本章是 ToughRADIUS 安全公告及其配套指引的权威归处。仓库根目录的
[`SECURITY.md`](https://github.com/talkincode/toughradius/blob/main/SECURITY.md)
保留一个指向本章的简短指针，以保证单一事实来源。

## 安全公告

### XSS 漏洞修复（v8.0.8）

**v8.0.8** 版本修复了一个严重的跨站脚本（XSS）漏洞。该问题位于登录接口对
`errmsg` 参数的处理。

| 项目     | 详情                          |
| -------- | ----------------------------- |
| 漏洞类型 | 跨站脚本（XSS）               |
| 严重程度 | 严重                          |
| 影响版本 | v8.0.1 – v8.0.7              |
| 修复版本 | v8.0.8                        |
| 受影响组件 | 登录接口（`errmsg` 参数）    |

#### 建议措施

强烈建议所有用户立即升级到最新版本。可参考[文档地图](./documentation-map.md)
中的 README 与构建说明完成部署升级。
