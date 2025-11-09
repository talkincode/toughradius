# 登录功能测试指南

## 问题描述

"无法认证跳转管理界面" - 登录后无法正常进入管理界面

## 已实施的修复

### 1. 改进 authProvider 错误处理

- ✅ 增强登录错误消息显示
- ✅ 验证 token 存在性
- ✅ 添加详细的错误日志

### 2. 创建自定义登录页面

- ✅ 美观的登录界面
- ✅ 密码可见性切换
- ✅ 加载状态显示
- ✅ 错误消息提示

### 3. 修复 App.tsx 配置

- ✅ 修正标题为 "ToughRADIUS v9"
- ✅ 添加 `requireAuth` 属性
- ✅ 配置自定义登录页面

## 测试步骤

### 1. 启动后端服务

```bash
cd /Volumes/ExtDISK/github/toughradius
go run main.go -c toughradius.yml
```

### 2. 构建前端（如果需要）

```bash
cd /Volumes/ExtDISK/github/toughradius/web
npm run build
```

### 3. 访问管理界面

打开浏览器访问：`http://localhost:1816/admin`

### 4. 测试登录

#### 正常登录流程

1. 输入正确的用户名和密码
2. 点击"登录"按钮
3. 应该看到加载指示器
4. 成功后自动跳转到控制台页面

#### 错误处理测试

1. **空用户名/密码**

   - 应显示警告："请输入用户名和密码"

2. **错误的用户名/密码**

   - 应显示错误消息："用户名或密码错误"

3. **禁用账号**

   - 应显示："账号已被禁用"

4. **网络错误**
   - 应显示具体的错误信息

## 检查后端日志

如果登录仍然失败，检查后端日志：

```bash
# 查看日志文件
tail -f rundata/logs/toughradius.log
```

常见错误：

- `INVALID_CREDENTIALS` - 用户名或密码错误
- `ACCOUNT_DISABLED` - 账号被禁用
- `DATABASE_ERROR` - 数据库查询失败
- `TOKEN_ERROR` - JWT token 生成失败

## 检查浏览器控制台

打开浏览器开发者工具（F12），检查：

1. **Network 标签**

   - 查看 `/api/v1/auth/login` 请求
   - 检查请求体和响应
   - 确认状态码（应为 200）

2. **Console 标签**

   - 查看是否有 JavaScript 错误
   - 检查 "登录错误:" 日志消息

3. **Application 标签**
   - 检查 localStorage 是否存储了：
     - `token`
     - `username`
     - `user`
     - `permissions`

## 默认管理员账号

如果没有账号，可以通过以下方式创建：

```bash
# 初始化数据库（会创建默认管理员账号）
./toughradius -initdb -c toughradius.yml
```

默认账号通常为：

- 用户名: `admin`
- 密码: `toughradius` 或查看配置文件

## API 端点测试

可以使用 curl 测试登录 API：

```bash
curl -X POST http://localhost:1816/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"yourpassword"}'
```

成功响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": 1,
      "username": "admin",
      "realname": "系统管理员",
      "level": "super",
      "status": "enabled"
    },
    "permissions": [],
    "tokenExpires": 1699999999
  }
}
```

## 常见问题排查

### 1. 登录后立即退出

- 检查 token 是否正确保存到 localStorage
- 检查 authProvider 的 checkAuth 方法
- 确认后端 JWT 配置正确

### 2. 401 错误

- 检查 token 是否过期（默认 12 小时）
- 确认 Authorization header 格式：`Bearer <token>`
- 检查后端 JWT secret 配置

### 3. CORS 错误

- 前端和后端应该在同一域名/端口
- 或者后端需要配置 CORS

### 4. 页面空白

- 检查浏览器控制台是否有 JavaScript 错误
- 确认前端构建成功
- 检查静态文件是否正确提供

## 修改的文件清单

### 前端

- `web/src/providers/authProvider.ts` - 改进错误处理
- `web/src/pages/LoginPage.tsx` - 新增自定义登录页面
- `web/src/App.tsx` - 配置登录页面和认证要求

### 文档

- `docs/login-troubleshooting.md` - 本文档

## 下一步

如果问题仍然存在，请提供：

1. 浏览器控制台的完整错误信息
2. 后端日志中的错误信息
3. 登录请求的网络响应详情
