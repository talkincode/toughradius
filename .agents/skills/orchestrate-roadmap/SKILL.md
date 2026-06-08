---
name: orchestrate-roadmap
description: 全局统筹 / 总调度角色。当用户发出"自动委托开发 / 继续推进路线图 / auto-delegate"等委托指令（不指定具体子任务）时启动，自动完成 选任务 → 选 SOP → 派工执行 → 质量门禁 → 提 PR → 自我迭代路线图 的闭环。
---

# 技能：路线图统筹与自动委托（Orchestrator）

> 关联：全部里程碑 / `TR-F022`　角色：**总调度**（其余技能为执行 SOP，本技能负责编排它们）

## 何时使用
当用户发出**自动委托开发**类指令、且不指定具体子任务时，本技能作为入口启动：
- "开始自动开发 / 自动委托开发 / 帮我推进路线图 / 继续推进开发 / 把 M1 做了 / auto-delegate / orchestrate"。
- 周期性"推进一轮开发"。

本技能不亲自堆砌实现细节，而是按路线图选题、选用对应执行技能、把关质量门禁，并在交付后驱动计划自我迭代。

## 统筹闭环（每轮）
1. **上下文同步**：先读 `AGENT.md`、`.agents/README.md`（通用前置约束 + 护栏）、`docs/roadmap.md`、`docs/feature-checklist.md`，确认工具链版本与禁区（`TR-N001`~`TR-N005`）。
2. **选任务**：
   - 口径：`grep -nE '^- \[ \] M[0-9]+\.[0-9]+' docs/roadmap.md | head -1` 取自上而下第一个未勾选子任务。
   - 优先级：`M1 → M2 → M3`；P1 里程碑无可执行子任务时再用 P2/P3 填充（见"里程碑总览"优先级列）。
   - 锚定该子任务的 `TR-F` 编号；命中非目标 `TR-N*` 立即停止并回报，不得自行扩张。
3. **选 SOP**：按任务类型匹配 `.agents/skills/<name>/SKILL.md`：
   - 厂商 VSA → `add-radius-vendor`；EAP 方法 → `add-eap-method`；Admin 接口 → `add-adminapi-endpoint`；前端资源 → `add-react-admin-resource`；配置项 → `add-config-schema`；上游库 → `sync-upstream-radius`；协议规范 → `reference-rfc`；测试 → `add-acceptance-test` / `write-go-tests`；需求对齐 → `align-feature-checklist`。
   - 找不到匹配 SOP 时，先用 `align-feature-checklist` 补齐范围与编号再继续。
4. **派工执行**：
   - 多 agent 环境：为该子任务派发一个**带完整上下文**的子 agent（注入选定 SKILL.md + 关联 RFC + 验收口径），一次只交付一个最小闭环。
   - 单 agent 环境：自己按选定 SOP 顺序执行。
   - 默认每轮只认领**一个**未勾选子任务（最小闭环、可回滚）。
5. **质量门禁**：`go build ./...`、`go test ./...`、`golangci-lint run`（v2.12.2）必须通过；涉及前端跑 `cd web && npm run build`；协议 / 端到端改动必须带 CI 可执行验收用例（`test/integration/`）。
6. **出 PR**：禁止直接推 `main`；PR 描述引用里程碑编号 + `TR-F` + 关联 RFC + 验收用例。
7. **自我迭代路线图**：按 `../groom-roadmap/SKILL.md` 勾选已交付子任务、更新里程碑状态、回填后续子任务 / 拆分 / 重排，并把实施中新发现的需求经 `align-feature-checklist` 纳入。
8. **循环或停**：
   - 默认完成一个闭环即停，回报本轮结果与下一个待办子任务。
   - 用户要求"持续推进"时回到第 2 步循环；遇到阻塞（缺规范 / 外部依赖 / 需决策）在安全检查点停下并在路线图标注 `阻塞` 原因。

## 边界
- 统筹者不绕过任何护栏：选任务口径、TR-N 禁区、PR-only、质量门禁全部强制。
- 一轮只推进一个最小闭环；不得为"多做"把多个子任务塞进一个 PR。
- 不自行扩张范围：任何超出 `TR-F` 清单的方向先走 `align-feature-checklist`。
- 只协调与编排，不替代执行 SOP 的具体约定（厂商单位换算、EAP 分片等仍以对应 SKILL.md 为准）。

## 验收
- [ ] 本轮选题来自路线图自上而下第一个未勾选子任务，且锚定 `TR-F`
- [ ] 使用了匹配的执行 SOP，未触碰 `TR-N`
- [ ] 质量门禁与（必要时）CI 验收用例通过
- [ ] 产出走 PR 并引用里程碑 / 编号 / RFC
- [ ] 交付后已按 `groom-roadmap` 迭代路线图与计划
