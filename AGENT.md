# ToughRADIUS AI Agent 开发指南

## 核心开发原则

本项目**严格遵循**以下三大开发原则，所有代码贡献必须符合这些标准：

### 🧪 测试驱动开发（TDD）

**强制要求：先写测试，后写代码**

#### TDD 工作流程

1. **红灯阶段** - 编写失败的测试

   ```bash
   # 创建测试文件
   touch internal/radius/new_feature_test.go

   # 运行测试（应该失败）
   go test ./internal/radius/new_feature_test.go -v
   ```

2. **绿灯阶段** - 编写最小实现使测试通过

   ```bash
   # 实现功能代码
   vim internal/radius/new_feature.go

   # 再次运行测试（应该通过）
   go test ./internal/radius/new_feature_test.go -v
   ```

3. **重构阶段** - 优化代码同时保持测试通过
   ```bash
   # 持续运行测试确保重构安全
   go test ./... -v
   ```

#### 测试覆盖率要求

- **新增功能代码覆盖率必须 ≥ 80%**
- **核心 RADIUS 协议模块覆盖率必须 ≥ 90%**
- **关键业务逻辑必须有集成测试**

```bash
# 检查测试覆盖率
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# 查看覆盖率统计
go test ./internal/radius -coverprofile=coverage.out
go tool cover -func=coverage.out
```

#### 测试文件组织

```
internal/radius/
├── auth_passwd_check.go          # 实现文件
├── auth_passwd_check_test.go     # 单元测试（同包）
├── radius_auth.go
├── radius_test.go                # 集成测试
└── vendor_parse_test.go          # 特性测试
```

#### 测试用例命名规范

```go
// ✅ 正确：清晰描述测试意图
func TestAuthPasswordCheck_ValidUser_ShouldReturnSuccess(t *testing.T) {}
func TestAuthPasswordCheck_ExpiredUser_ShouldReturnError(t *testing.T) {}
func TestGetNas_UnauthorizedIP_ShouldReturnAuthError(t *testing.T) {}

// ❌ 错误：模糊不清
func TestAuth(t *testing.T) {}
func TestFunc1(t *testing.T) {}
```

#### 表驱动测试（Table-Driven Tests）

对于多场景测试，使用表驱动方式：

```go
func TestVendorParse(t *testing.T) {
    tests := []struct {
        name       string
        vendorCode string
        input      string
        wantMac    string
        wantVlan1  int64
    }{
        {"Huawei VLAN", VendorHuawei, "vlan=100", "", 100},
        {"Mikrotik MAC", VendorMikrotik, "mac=aa:bb:cc:dd:ee:ff", "aa:bb:cc:dd:ee:ff", 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 测试逻辑
        })
    }
}
```

### 🔄 GitHub 工作流

**强制要求：遵循 Git Flow 分支模型和标准 PR 流程**

#### 分支策略

```
main (生产分支)
  ├── v9dev (开发分支)
  │    ├── feature/user-management     # 功能分支
  │    ├── feature/radius-vendor-cisco # 功能分支
  │    ├── bugfix/session-timeout      # 缺陷修复
  │    └── hotfix/security-patch       # 紧急修复
  └── release/v9.1.0                   # 发布分支
```

#### 标准开发流程

**1. 创建功能分支**

```bash
# 从 v9dev 创建功能分支
git checkout v9dev
git pull origin v9dev
git checkout -b feature/add-cisco-vendor

# 分支命名规范
# feature/  - 新功能
# bugfix/   - 缺陷修复
# hotfix/   - 紧急修复
# refactor/ - 代码重构
# docs/     - 文档更新
```

**2. TDD 循环开发**

```bash
# 1️⃣ 先写测试
vim internal/radius/vendors/cisco/cisco_test.go

# 2️⃣ 运行测试（红灯）
go test ./internal/radius/vendors/cisco -v

# 3️⃣ 实现功能
vim internal/radius/vendors/cisco/cisco.go

# 4️⃣ 运行测试（绿灯）
go test ./internal/radius/vendors/cisco -v

# 5️⃣ 提交原子化的变更
git add internal/radius/vendors/cisco/
git commit -m "test: add Cisco vendor attribute parsing tests"
git commit -m "feat: implement Cisco vendor attribute parsing"
```

**3. 提交规范（Conventional Commits）**

```bash
# 格式：<type>(<scope>): <subject>
git commit -m "feat(radius): add Cisco vendor support"
git commit -m "test(radius): add unit tests for Cisco attributes"
git commit -m "fix(auth): correct password validation logic"
git commit -m "docs(api): update RADIUS authentication API docs"
git commit -m "refactor(database): optimize user query performance"
git commit -m "perf(radius): reduce authentication latency by 20%"

# Type 类型
# feat:     新功能
# fix:      缺陷修复
# test:     测试相关
# docs:     文档更新
# refactor: 代码重构
# perf:     性能优化
# style:    代码格式
# chore:    构建/工具变更
```

**4. 创建 Pull Request**

PR 必须包含：

- ✅ **所有测试通过**（`go test ./...`）
- ✅ **代码覆盖率达标**
- ✅ **清晰的描述和变更说明**
- ✅ **关联的 Issue 编号**
- ✅ **至少一个代码审查者批准**

PR 模板：

```markdown
## 变更描述

简要说明本次 PR 的目的和主要变更

## 变更类型

- [ ] 新功能
- [ ] 缺陷修复
- [ ] 性能优化
- [ ] 代码重构
- [ ] 文档更新

## 测试覆盖

- [ ] 添加了单元测试
- [ ] 添加了集成测试
- [ ] 测试覆盖率 ≥ 80%
- [ ] 所有测试通过

## 检查清单

- [ ] 代码遵循项目规范
- [ ] 提交信息符合 Conventional Commits
- [ ] 更新了相关文档
- [ ] 无破坏性变更（或已在 CHANGELOG 中说明）

## 关联 Issue

Closes #123
```

**5. 持续集成检查**

每个 PR 自动触发：

- ✅ `go test ./...` - 运行所有测试
- ✅ `go build` - 确保代码可编译
- ✅ Docker 镜像构建
- ✅ 代码风格检查

#### 发布流程

```bash
# 1. 创建发布分支
git checkout -b release/v9.1.0 v9dev

# 2. 更新版本号和 CHANGELOG
vim VERSION
vim CHANGELOG.md

# 3. 合并到 main 并打标签
git checkout main
git merge --no-ff release/v9.1.0
git tag -a v9.1.0 -m "Release version 9.1.0"
git push origin main --tags

# 4. 触发 GitHub Actions 自动发布
# - 构建 AMD64/ARM64 二进制
# - 发布 Docker 镜像到 DockerHub 和 GHCR
# - 创建 GitHub Release
```

### 📦 最小可用产品（MVP）原则

**强制要求：每个功能必须以最小可用单元交付**

#### MVP 设计方法

1. **确定核心价值**

   - ❓ 这个功能解决什么问题？
   - ❓ 最简化的实现是什么？
   - ❓ 哪些是必需的，哪些是锦上添花？

2. **功能拆分示例**

   ```
   ❌ 错误做法：一次性实现完整功能
   Issue #123: 添加 Cisco 厂商支持
   └── 包含认证、计费、VSA 属性、配置管理、Web 界面...

   ✅ 正确做法：MVP 拆分
   Issue #123: 添加 Cisco 厂商基础认证支持 (MVP-1)
   ├── PR #124: Cisco VSA 属性解析
   ├── PR #125: Cisco 认证流程集成
   └── PR #126: 基础测试用例

   Issue #130: 扩展 Cisco 计费功能 (MVP-2)
   └── 基于 MVP-1 构建

   Issue #135: 添加 Cisco Web 管理界面 (MVP-3)
   └── 基于 MVP-1 + MVP-2 构建
   ```

3. **MVP 交付标准**

   每个 MVP 必须：

   - ✅ **独立可用** - 不依赖未完成的功能
   - ✅ **完整测试** - 覆盖率达标
   - ✅ **文档完善** - API 文档 + 使用说明
   - ✅ **可演示** - 能够运行并展示价值
   - ✅ **可回滚** - 不破坏现有功能

#### MVP 实践案例

**案例 1：新增 RADIUS 厂商支持**

```
MVP-1（第1周）：基础属性解析 ✅
├── vendor_cisco.go          # 厂商常量定义
├── vendor_cisco_test.go     # 解析测试
└── 支持读取基础 VSA 属性

MVP-2（第2周）：认证集成 ✅
├── auth_accept_config.go    # 添加 Cisco case
├── auth_cisco_test.go       # 认证集成测试
└── 支持 Cisco 设备认证流程

MVP-3（第3周）：计费支持 ✅
└── 扩展计费记录 Cisco 特定字段

MVP-4（第4周）：Web 管理 ✅
└── Admin API 添加 Cisco 配置界面
```

**案例 2：性能优化**

```
MVP-1：识别瓶颈 ✅
├── 添加性能测试基准
├── 识别热点函数
└── 建立性能基线

MVP-2：优化数据库查询 ✅
├── 添加索引
├── 优化 N+1 查询
└── 验证性能提升 20%

MVP-3：缓存层 ✅ (可选)
└── 如果性能仍不达标则继续
```

## 开发工作流完整示例

### 场景：添加新的 RADIUS 厂商支持（Cisco）

**第 1 步：创建 Issue（需求分析）**

```markdown
Title: [Feature] 添加 Cisco RADIUS 厂商支持

## MVP-1 范围

- [ ] 解析 Cisco VSA 属性
- [ ] 单元测试覆盖率 ≥ 90%
- [ ] 文档更新

## MVP-2 范围（后续）

- [ ] 认证流程集成
- [ ] 计费功能支持

## 不包含

- Web 管理界面（MVP-3）
- 高级配置管理（MVP-4）
```

**第 2 步：TDD 开发**

```bash
# 1️⃣ 创建分支
git checkout -b feature/cisco-vendor-mvp1 v9dev

# 2️⃣ 先写测试（红灯）
cat > internal/radius/vendors/cisco/cisco_test.go << 'EOF'
package cisco

import "testing"

func TestParseCiscoAVPair(t *testing.T) {
    tests := []struct{
        name  string
        input string
        want  map[string]string
    }{
        {"basic", "key=value", map[string]string{"key": "value"}},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := ParseAVPair(tt.input)
            // 断言逻辑
        })
    }
}
EOF

# 3️⃣ 运行测试（应该失败）
go test ./internal/radius/vendors/cisco -v

# 4️⃣ 实现最小可用代码（绿灯）
cat > internal/radius/vendors/cisco/cisco.go << 'EOF'
package cisco

func ParseAVPair(input string) map[string]string {
    // 最简实现
    return map[string]string{}
}
EOF

# 5️⃣ 运行测试（应该通过）
go test ./internal/radius/vendors/cisco -v

# 6️⃣ 重构优化
# 改进实现，持续确保测试通过

# 7️⃣ 检查覆盖率
go test ./internal/radius/vendors/cisco -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total
```

**第 3 步：提交代码**

```bash
# 原子化提交
git add internal/radius/vendors/cisco/cisco_test.go
git commit -m "test(radius): add Cisco AVPair parsing tests"

git add internal/radius/vendors/cisco/cisco.go
git commit -m "feat(radius): implement Cisco AVPair parsing (MVP-1)"

git add docs/radius/cisco-vendor.md
git commit -m "docs(radius): add Cisco vendor documentation"
```

**第 4 步：创建 Pull Request**

```bash
git push origin feature/cisco-vendor-mvp1
# 在 GitHub 上创建 PR，填写 PR 模板
```

**第 5 步：代码审查和合并**

- 等待 CI 通过
- 代码审查反馈
- 修复问题，推送更新
- 获得批准后合并到 v9dev

**第 6 步：计划 MVP-2**

- 创建新的 Issue 用于下一个 MVP
- 重复上述流程

## 质量门禁（Quality Gates）

所有代码合并前必须通过：

### ✅ 自动化检查

- [ ] 所有单元测试通过（`go test ./...`）
- [ ] 代码覆盖率 ≥ 80%
- [ ] 编译无错误（`go build`）
- [ ] Docker 镜像构建成功
- [ ] 前端测试通过（`npm run test`）

### ✅ 代码审查

- [ ] 至少一个维护者批准
- [ ] 无未解决的审查意见
- [ ] 符合代码规范

### ✅ 文档要求

- [ ] 代码注释充分（特别是导出函数）
- [ ] API 变更更新了文档
- [ ] CHANGELOG.md 已更新（如果面向用户）

### ✅ MVP 验收

- [ ] 功能独立可用
- [ ] 满足最小需求
- [ ] 不引入技术债务

## 常见反模式（禁止）

### ❌ 反模式 1：无测试提交

```bash
# 错误示例
git commit -m "feat: add new feature"  # 无对应测试文件

# 正确做法
git commit -m "test: add tests for new feature"
git commit -m "feat: add new feature"
```

### ❌ 反模式 2：巨型 PR

```
❌ PR #100: 完整实现用户管理系统
   +2000 -500 lines across 50 files

✅ 拆分为：
   PR #101: User model and database migration (MVP-1)
   PR #102: User CRUD API endpoints (MVP-2)
   PR #103: User management UI (MVP-3)
```

### ❌ 反模式 3：先实现后测试

```go
// ❌ 错误流程
1. 实现完整功能
2. 功能已经很复杂
3. 补充测试困难
4. 测试覆盖率不足

// ✅ TDD 流程
1. 写测试（定义行为）
2. 最小实现
3. 重构优化
4. 测试覆盖率自然达标
```

### ❌ 反模式 4：跳过 Code Review

```bash
# ❌ 直接推送到主分支
git push origin main  # 被保护分支拒绝

# ✅ 通过 PR 流程
git push origin feature/my-feature
# 创建 PR → CI 检查 → 代码审查 → 合并
```

## 工具配置

### 本地开发环境设置

```bash
# 安装 Git hooks（自动化测试）
cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
echo "运行测试..."
go test ./...
if [ $? -ne 0 ]; then
    echo "❌ 测试失败，提交已阻止"
    exit 1
fi
echo "✅ 测试通过"
EOF
chmod +x .git/hooks/pre-commit

# 配置 commit 模板
git config commit.template .gitmessage.txt
```

### 推荐 VS Code 扩展

- **Go** - Go 语言支持
- **Go Test Explorer** - 测试可视化
- **Coverage Gutters** - 覆盖率展示
- **Conventional Commits** - 提交规范辅助
- **GitLens** - Git 增强

## 参考资料

- [TDD 实践指南](https://martinfowler.com/bliki/TestDrivenDevelopment.html)
- [Git Flow 工作流](https://nvie.com/posts/a-successful-git-branching-model/)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [MVP 方法论](https://www.agilealliance.org/glossary/mvp/)
- [Go 测试最佳实践](https://go.dev/doc/tutorial/add-a-test)

---

**记住：质量优于速度，可用优于完美，测试先于代码！**
