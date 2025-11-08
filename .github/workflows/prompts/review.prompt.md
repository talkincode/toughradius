---
mode: 'agent'
model: GPT-5
tools: ['edit', 'runNotebooks', 'search', 'new', 'runCommands', 'runTasks', 'executePrompt', 'usages', 'vscodeAPI', 'think', 'problems', 'changes', 'testFailure', 'openSimpleBrowser', 'fetch', 'githubRepo', 'extensions', 'todos', 'runTests']
description: '项目代码质量自动检测与分析'
---


# 代码质量检测提示 (review.prompt)

本提示用于指导智能助手对项目进行系统化、全面的代码质量检测与分析,识别潜在问题、安全隐患、性能瓶颈与可维护性风险。遵循 AGENTS.md 与项目最佳实践。

---

## 1. 检测目标

- **代码健康度评估**: 复杂度、重复代码、函数长度、文件规模
- **安全隐患识别**: 硬编码密钥、输入验证缺失、敏感信息泄露
- **并发安全检查**: 数据竞争、锁竞争、goroutine 泄露风险
- **性能问题发现**: 不必要分配、低效算法、频繁 I/O 操作
- **测试覆盖分析**: 缺失测试、脆弱测试、依赖外部环境测试
- **架构一致性**: 跨层调用、循环依赖、职责不清
- **可维护性评估**: 注释缺失、命名混乱、魔法数字

---
