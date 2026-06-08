---
name: add-acceptance-test
description: 为协议或端到端改动编写 CI 可自动执行的验收/集成测试 (TR-F022)。任何里程碑子任务需要 CI 背书验收时使用，覆盖 test/integration 集成测试与单元测试。
---

# 技能：编写 CI 自动化验收测试

> 关联功能编号：`TR-F022`　适用：所有协议 / 端到端验收

## 目标
每个里程碑子任务的验收口径都要由 **CI 可自动执行的测试** 背书。本技能说明两类测试落点与写法。

## 两类测试

### 1. 单元 / 逻辑验收 — `*_test.go`
- 与被测代码同包；CI `test` job 执行 `go test -short ./...`。
- 适用：纯逻辑、属性编解码、校验器、配置解析等。
- 写法见 `../write-go-tests/SKILL.md`。

### 2. 端到端 / 协议验收 — `test/integration/`
- 文件首行 `//go:build integration`，包 `integration`。
- 由真实 PostgreSQL 支撑（生产默认数据库），驱动真实运行的 RADIUS / Admin 服务发送真实报文。
- CI `integration` job 自动执行：
  ```
  go test -tags=integration -count=1 -v ./test/integration/...
  INTEGRATION_REQUIRED=1   # 环境缺失时 CI 硬失败，杜绝假绿
  ```
- 现有范本：
  - `test/integration/main_test.go` — 共享 harness（建库、起服务）
  - `test/integration/radius_test.go` — 真实 PAP Access-Request 端到端
  - `test/integration/adminapi_test.go` — Admin API 端到端
  - `test/integration/migration_test.go` — 迁移
  - `test/integration/client_test.go`

## 前置检索
```text
view test/integration/main_test.go          # harness 用法
view test/integration/radius_test.go         # RADIUS 端到端范本
grep_search "INTEGRATION_REQUIRED" --include test/**
view Makefile                                # test / test-integration-pg 目标
```

## 本地运行
```bash
# 单元
go test ./...
# 集成（自动拉起 docker-compose.test.yml 的 Postgres，跑完清理）
make test-integration-pg
# 或手动提供 TEST_DATABASE_* 后：
go test -tags=integration -count=1 -v ./test/integration/...
```

## 新增端到端验收用例步骤
1. 在 `test/integration/` 新建或扩展 `<feature>_test.go`，首行加 `//go:build integration`。
2. 复用 `main_test.go` 的 harness 创建数据、起服务；新功能（如 EAP-TLS / CoA）按现有 `radius_test.go` 模式驱动真实报文。
3. 断言成功 **与** 失败路径（拒绝原因、超时、NAS 拒绝等）。
4. 串行依赖全局状态的用例不要 `t.Parallel()`（参考 `radius_test.go` 注释）。
5. 确认 CI `integration` job 覆盖该路径（默认 `./test/integration/...` 全量执行，无需改 CI）。

## 边界
- 集成用例不得依赖外部公网服务；所需依赖通过 service container / compose 提供。
- 不得为了过测降低断言强度或跳过失败路径。

## 验收
- [ ] 新功能有对应的 CI 自动化验收用例（单元或集成）
- [ ] 失败路径有断言
- [ ] `make test-integration-pg` 本地通过
- [ ] CI `test` 与 `integration` job 通过
