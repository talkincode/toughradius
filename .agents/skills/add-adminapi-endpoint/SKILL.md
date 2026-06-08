---
name: add-adminapi-endpoint
description: 在管理后台新增一组 Admin REST 接口 (TR-F012)。需要为某功能提供增删改查等管理 API 时使用。
---

# 技能：新增 Admin REST 接口

> 关联功能编号：`TR-F012`　适用里程碑：M2 等

## 何时使用
需要在管理后台新增一组 REST 接口时。

## 前置检索
```text
view internal/adminapi/adminapi.go        # Init() 注册入口
view internal/adminapi/nodes.go           # 标准 CRUD 范本
view internal/adminapi/helpers.go         # parsePagination / 过滤辅助
view internal/adminapi/responses.go       # 统一响应
view internal/adminapi/authz.go           # requireAdmin() 等鉴权
```

## 实现步骤
1. **新建文件** `internal/adminapi/<feature>.go`，包 `adminapi`。
2. **请求结构体**：定义 payload，使用 `validate` tag（参考 `nodePayload`）。
3. **register 函数**：
   ```go
   func register<Feature>Routes() {
       webserver.ApiGET("/<path>", list<Feature>)
       webserver.ApiPOST("/<path>", create<Feature>, requireAdmin())
       webserver.ApiPUT("/<path>/:id", update<Feature>, requireAdmin())
       webserver.ApiDELETE("/<path>/:id", delete<Feature>, requireAdmin())
   }
   ```
4. **注册到 Init**：在 `adminapi.go` 的 `Init()` 增加 `register<Feature>Routes()`。
5. **数据访问**：统一用 `app.GDB()`，禁止注入 `*gorm.DB`。
6. **统一约定**：复用分页、过滤、验证、鉴权、统一响应与错误处理。

## 验收
- [ ] 新增 `<feature>_test.go`，覆盖增删改查与鉴权失败
- [ ] `go test ./internal/adminapi/...` 通过
- [ ] `golangci-lint run` 通过
- [ ] 若涉及前端，配套 `../add-react-admin-resource/SKILL.md`
- [ ] PR 引用 `TR-F012` 与里程碑编号
