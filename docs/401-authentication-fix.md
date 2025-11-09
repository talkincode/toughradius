# 401 认证问题修复验证指南

## 问题症状

```
登录成功（200）→ 后续请求返回 401 未授权
POST /api/v1/auth/login HTTP/1.1 200 ✓
GET /api/v1/system/operators/xxx HTTP/1.1 401 ✗
```

## 已实施的修复

### 1. 改进 authProvider.login 时序控制

**文件**: `web/src/providers/authProvider.ts`

添加了：

- ✅ 确保 token 正确保存到 localStorage
- ✅ 添加微小延迟确保数据写入完成
- ✅ 添加调试日志验证 token 保存状态

### 2. httpClient 调试增强

**文件**: `web/src/providers/dataProvider.ts`

添加了：

- ✅ 请求发送时记录是否找到 token
- ✅ Console 日志显示每个请求的 token 状态

## 验证步骤

### 1. 清除旧数据

打开浏览器开发者工具（F12）→ Application → Local Storage → 清除所有项

### 2. 重新登录

1. 访问 `http://localhost:1816/admin`
2. 输入用户名和密码
3. 点击登录

### 3. 检查 Console 日志

**应该看到**：

```
登录成功，token 已保存: ✓
✓ Token 已添加到请求: /api/v1/...
✓ Token 已添加到请求: /api/v1/...
```

**不应该看到**：

```
✗ 未找到 token，请求: /api/v1/...
```

### 4. 检查 Network 面板

在 Network 标签中，查看任何 API 请求：

**Request Headers 中应该包含**：

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

### 5. 检查 Application → Local Storage

应该看到以下项：

- `token`: `eyJhbGciOiJIUzI1NiIs...`
- `username`: `admin`
- `user`: `{"id":123,"username":"admin",...}`
- `permissions`: `[]`

## 如果仍然出现 401

### 场景 1: Console 显示 "未找到 token"

**原因**: localStorage 中没有 token
**解决**:

1. 检查登录响应是否包含 token
2. 检查 Console 是否显示 "登录成功，token 已保存: ✓"
3. 手动检查 localStorage

### 场景 2: Console 显示 "Token 已添加" 但仍然 401

**原因**: Token 无效或过期
**检查**:

1. 复制 token 值
2. 在 https://jwt.io 解码查看内容和过期时间
3. 检查后端日志是否有 token 验证错误

**后端验证**:

```bash
# 使用 curl 测试 token
TOKEN="从 localStorage 复制的 token"
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:1816/api/v1/auth/me
```

### 场景 3: 随机出现 401

**原因**: 竞态条件，某些请求在 token 保存前发出
**解决**: 已在 authProvider.login 中添加延迟

### 场景 4: 后端日志显示 "missing token context"

**原因**: Authorization header 格式不正确
**检查**:

- 应该是 `Bearer <token>`
- 不是 `bearer <token>` (小写)
- token 前有空格

## 后端调试

### 检查 JWT 中间件配置

**文件**: `internal/webserver/server.go`

确认跳过列表包含登录接口：

```go
var JwtSkipPrefix = []string{
    "/ready",
    "/realip",
    "/api/v1/auth/login",    // ✓ 登录接口应该跳过 JWT 验证
    "/api/v1/auth/refresh",
}
```

### 检查后端日志

```bash
tail -f rundata/logs/toughradius.log | grep -E "(auth|401|token)"
```

常见错误：

- `invalid token claims` - token 格式错误
- `invalid token subject` - token 中缺少用户 ID
- `missing token context` - 请求头中没有 token
- `record not found` - token 中的用户 ID 不存在

## 调试命令

### 测试完整登录流程

```bash
# 1. 登录获取 token
curl -X POST http://localhost:1816/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"yourpassword"}' \
  -v 2>&1 | tee login-response.txt

# 2. 提取 token
TOKEN=$(grep -o '"token":"[^"]*' login-response.txt | cut -d'"' -f4)
echo "Token: $TOKEN"

# 3. 测试认证请求
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:1816/api/v1/auth/me \
  -v

# 4. 测试操作员查询
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:1816/api/v1/system/operators/1 \
  -v
```

## 预期结果

登录后所有请求都应该：

1. **200 OK** - 请求成功
2. **Console**: 显示 "✓ Token 已添加到请求"
3. **Network**: Request Headers 包含 Authorization
4. **后端日志**: 无 401 或认证错误

## 生产环境检查清单

- [ ] 清除浏览器缓存和 localStorage
- [ ] 确认前端已重新构建（`npm run build`）
- [ ] 确认后端已重启
- [ ] 确认后端 JWT secret 配置正确
- [ ] 确认 token 有效期配置（默认 12 小时）
- [ ] 检查 CORS 配置（如果前后端分离部署）

## 移除调试日志

在生产环境部署前，可以移除调试日志：

**authProvider.ts**:

```typescript
// 移除这行
console.log(
  "登录成功，token 已保存:",
  localStorage.getItem("token") ? "✓" : "✗"
);
```

**dataProvider.ts**:

```typescript
// 移除这些行
console.log("✓ Token 已添加到请求:", url);
console.warn("✗ 未找到 token，请求:", url);
```

或者使用环境变量控制：

```typescript
if (process.env.NODE_ENV === "development") {
  console.log("调试信息...");
}
```
