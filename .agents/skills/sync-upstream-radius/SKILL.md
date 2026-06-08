---
name: sync-upstream-radius
description: 跟踪并同步上游 layeh.com/radius 库到组织 fork talkincode/radius (TR-F021/TR-F022)。巡检发现上游新提交、或怀疑协议库缺陷、或例行依赖维护时使用。
---

# 技能：跟踪与同步上游 RADIUS 库

> 关联功能编号：`TR-F021` / `TR-F022`　适用里程碑：M7

## 背景
- 核心协议库：`layeh.com/radius`（原始仓库 <https://github.com/layeh/radius>）。
- 组织 fork：`github.com/talkincode/radius`（<https://github.com/talkincode/radius>）。
- 接入方式：`go.mod` 中
  ```
  require layeh.com/radius v0.0.0-<pseudo>
  replace layeh.com/radius => github.com/talkincode/radius v<tag>
  ```
  即实际编译使用的是 fork，`layeh.com/radius` 的 import 路径不变。

## 何时使用
- 例行（建议每周）人工核对上游 `layeh/radius` 是否有新提交时。
- 发现/怀疑协议编解码、属性处理、安全相关缺陷时。
- 例行依赖维护时。

## 流程
1. **确认当前固定版本**
   ```bash
   grep -E "layeh.com/radius|talkincode/radius" go.mod
   ```
   记录 `replace` 指向的 fork tag，以及 `require` 伪版本中的上游提交短 sha。
2. **对比上游差异**
   - 原始仓库 `layeh/radius`：查看自固定提交以来的新提交，例如用 GitHub compare：
     `https://github.com/layeh/radius/compare/<pinned_sha>...master`，或 `git log <pinned_sha>..` 于上游克隆。
   - 重点筛选：安全修复、属性编解码正确性、EAP / VSA / 报文解析相关变更。
3. **评估同步**
   - 若上游修复重要且 fork 尚未包含：在 fork（`talkincode/radius`）合并/cherry-pick，并打新 tag。
   - 在本仓库更新 `go.mod` 的 `replace` 到新 tag，`go mod tidy`。
   - 若修复不影响本项目用法，记录"已评估、暂不同步"的理由（在巡检 Issue 中回复）。
4. **验证**
   ```bash
   go build ./...
   go test ./...
   go test -tags=integration -count=1 ./test/integration/...   # 需 Postgres，见 ../add-acceptance-test/SKILL.md
   golangci-lint run
   ```
5. **回归保护**：若同步是为修某协议缺陷，补一个能复现该缺陷的测试，防止未来回退。

## 边界
- 不在本仓库直接 vendoring 或 patch 第三方库源码；改动走 fork 仓库。
- `replace` 只指向受信任的组织 fork，不引入未知第三方分支。
- 升级必须在 PR 中说明：上游提交范围、风险评估、是否影响协议行为。

## 验收
- [ ] `go.mod` / `go.sum` 一致，`go mod tidy` 无残留
- [ ] 全量 + 集成测试通过
- [ ] 巡检 Issue 有同步决策记录（同步 / 暂不同步 + 理由）
- [ ] PR 引用 `TR-F021` / `TR-F022`
