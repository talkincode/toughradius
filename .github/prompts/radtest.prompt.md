---
model: Claude Sonnet 4.5
tools:
  [
    "search",
    "azure/search",
    "usages",
    "problems",
    "changes",
    "githubRepo",
    "todos",
  ]
description: "自动测试 toughradius 服务器的 RADIUS 协议功能"
---

# RADIUS 协议自动测试提示 (radtest.prompt)

本提示引导智能助手自动测试 toughradius 服务器的 RADIUS 协议功能，包括认证、计费和会话管理等方面。测试涵盖常见场景和边界情况，确保服务器在各种条件下的正确性和稳定性。遵循 AGENTS.md 和项目最佳实践。

---

## 测试工具

- 测试数据管理： 通过 bin/testdata 自动创建测试数据， 测试完成后清理数据。
- RADIUS 客户端模拟： 使用 cmd/radtest 工具模拟 RADIUS 客户端发送各种请求。
- 基准测试： 使用 bin/benchmark 进行性能基准测试。
- 使用 help 参数 或者查询 cmd下的代码文档，了解各工具的使用方法和参数。


## 测试目标

- 验证 RADIUS 认证请求的正确处理，包括成功和失败的认证。
- 测试 RADIUS 计费请求的处理，确保正确记录和响应。
- 检查 RADIUS 会话管理功能，包括启动、停止和中间更新请求
- 评估服务器在高并发和异常条件下的稳定性和性能。

## 随机取样测试

- 通过 sqlite3 查询 toughradius.db 数据库，随机选择用户进行测试。
- 确保测试覆盖不同用户类型和配置。

## 测试结果分析

- 分析测试结果，识别潜在问题和改进点。
- 生成 Markdown 格式测试报告，