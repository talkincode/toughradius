---
name: write-go-tests
description: 编写 Go 单元/集成测试的统一约定 (TR-F022)。任何 Go 代码改动配套测试时使用。
---

# 技能：编写 Go 单元 / 集成测试

> 关联功能编号：`TR-F022`

## 何时使用
任何 Go 代码改动都必须配套测试。本技能统一测试约定。

## 前置检索
```text
file_search "internal/**/*_test.go"            # 找最接近的范本
view internal/adminapi/nodes_test.go           # API 测试范本
view internal/radiusd/*_test.go                # 协议测试范本
view internal/app/*_test.go                    # app 层测试范本
```

## 约定
1. **就近测试**：测试文件与被测文件同包同目录，命名 `<file>_test.go`。
2. **短测试标记**：耗时 / 需外部依赖的用 `if testing.Short() { t.Skip(...) }`，CI 跑 `go test -short`。
3. **数据库**：开发 / 测试用 SQLite（纯 Go，`CGO_ENABLED=0`）；schema 改动须双数据库兼容。
4. **覆盖重点**：成功路径 + 失败路径（拒绝原因、超时、校验失败、鉴权失败）。
5. **指标 / 错误**：认证拒绝类改动必须断言对应 `AuthError` / metrics 标签。
6. **表驱动**：多场景用 table-driven test。

## 运行
```bash
go test ./...                       # 全量
go test -short ./...                # CI 等价（快速）
go test -run TestXxx ./internal/... # 单测
go test -bench=. ./internal/radiusd/ # 基准
golangci-lint run                   # lint (v2.12.2)
```

## 验收
- [ ] 新增 / 改动逻辑均有测试覆盖
- [ ] `go test ./...` 与 `golangci-lint run` 通过
- [ ] 关键失败路径有断言
