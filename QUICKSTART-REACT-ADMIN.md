# ToughRADIUS 精简版重构计划

## 1. 核心目标
- 构建一个「仅保留 RADIUS 核心能力」的后端，删除 TR069、CPE 以及所有遗留兼容层。
- 交付一个经静态编译并嵌入 Go 可执行文件的 React Admin 前端，运行在 `/admin`。
- 精简仓库结构与脚本，只保留编译、测试、发布链路所需内容。
- 后端 Admin API 与新前端一一对应，返回结构统一、可测试、可扩展。

## 2. 重构范围与准则
- **删减优先**：任何与 RADIUS 管理无直接关系的代码、配置、脚本全部移除。
- **API 完全重写**：旧版 Admin API 不再兼容；以 REST + JWT 重新设计。
- **单一前端**：只保留 `web`（React Admin）；`web-old` 与其它 UI 脚本全部删除。
- **可嵌入交付**：前端以 `npm run build` 生成静态资产并由 Go `embed` 打包。
- **自动化验证**：构建、单测、端到端校验脚本最少化但保持可复现。

## 3. 阶段划分
1. **清理阶段**
   - 移除 TR069/CPE 模块、依赖、配置及脚本条目。
   - 删除旧前端、TR069 相关文档、脚本（含 build-all 子流程）。
   - 精简 `scripts/`、`Makefile`、CI 任务，仅保留 `build`, `test`, `release`.
2. **后端重构阶段**
   - 重新梳理 `cmd/toughradius` 启动逻辑，确保仅加载 RADIUS 所需组件。
   - 划分 `internal/adminapi`（新 API）、`internal/radiusd`, `internal/storage`.
   - 统一配置（`config/` + `toughradius.yml`）仅暴露 RADIUS 与 Admin API 相关字段。
   - 引入版本化的 migration 与种子数据，确保 Admin UI 初始可登陆。
3. **前端重构阶段**
   - `web/` 目录保留 React Admin + Vite；配置服务端 API 根路径 `/api/v2`.
   - 编写资源：认证、用户、在线会话、计费记录、配置模板等。
   - 定义与后端契合的数据模型、字段映射、筛选器。
   - 在 `web/static.go` 重新生成 embed 代码，Go 构建产物即内置前端。
4. **集成与交付阶段**
   - 在 `internal/webserver` 内提供 SPA fallback。
   - 新增 e2e smoke 测试：登陆、列表、详情。
   - 更新 `README.md` 与 `docs/`，仅保留与精简版相关的说明。

## 4. 后端 Admin API 设计纲要
- **Auth**：`POST /api/v2/auth/login`，返回 `token` 与权限；支持后续 `refresh/logout`。
- **RADIUS Users**：`/api/v2/users`（CRUD + 批量导入）。
- **Online Sessions**：`/api/v2/sessions`（分页、下线操作）。
- **Accounting Records**：`/api/v2/accounting`（过滤、导出）。
- **Profiles & Policies**：`/api/v2/profiles`，承载计费模板、限速等配置。
- **系统配置**：`/api/v2/system/settings`（如 NAS、日志级别）。
- 统一返回：
  ```json
  {
    "data": {},
    "meta": {
      "total": 0,
      "page": 1,
      "pageSize": 20
    }
  }
  ```
- 错误格式：`{ "error": "INVALID_CREDENTIALS", "message": "..." }`.

## 5. 前端精简策略
- 仅依赖 React Admin + 必需 UI/表单库，移除多余组件、Charts、TR069 页面。
- `src/providers`：自定义 dataProvider/authProvider，封装 token 续期与错误处理。
- `src/resources`：为 Users、Sessions、Accounting、Profiles、Settings 创建模块化目录。
- 使用 `vite.config.ts` 的 `base: "/admin"`，确保嵌入式路径正确。
- 脚本统一：
  - `npm run dev`、`npm run build`、`npm run lint`。
  - `npm run check` 组合 lint + typecheck。

## 6. 仓库结构（目标状态）
```
cmd/
config/
docs/
internal/
  adminapi/
  radius/
  storage/
  webserver/
migrations/
pkg/
web/
  src/
  vite.config.ts
  static.go
Makefile
README.md
scripts/
  build-frontend.sh
  build-backend.sh
```

## 7. 构建与交付流程
1. `npm install && npm run build`（web）。
2. 运行 `scripts/build-frontend.sh` 生成 `web/dist` 并触发 `go generate ./web`.
3. `go test ./...`；必要时增加 `internal/adminapi` 的接口测试。
4. `go build -o toughradius ./cmd/toughradius` 打包嵌入式前端。
5. `./toughradius -c toughradius.yml` 启动，自带 `/api/v2` 与 `/admin`.

## 8. 立即行动清单
- [ ] 删除 TR069/CPE/旧前端相关目录、脚本、依赖。
- [ ] 重新整理 `Makefile` 与 `scripts/`，确保只包含当前流程。
- [ ] 设计并实现 `/api/v2` 的路由、数据模型、存储层。
- [ ] 用 React Admin 搭建新的最小界面，验证 CRUD + 登录。
- [ ] 更新文档（README、部署说明）与版本号，标记为精简版。

此计划将 ToughRADIUS 聚焦于核心 RADIUS 能力，并确保前后端协同、交付链路精简，为后续功能添加留出清晰扩展点。
