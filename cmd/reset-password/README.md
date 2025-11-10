# 管理员密码重置工具

当你忘记管理员密码时，可以使用此工具重置密码。

## 快速使用

### 方法 1: 使用脚本 (推荐)

```bash
# 重置为默认密码 'toughradius'
./scripts/reset-admin-password.sh

# 或指定新密码
./scripts/reset-admin-password.sh myNewPassword123
```

### 方法 2: 手动编译运行

**PostgreSQL 数据库:**

```bash
# 1. 编译重置密码工具 (静态编译)
cd cmd/reset-password
CGO_ENABLED=0 go build -o reset-password .

# 2. 运行工具
./reset-password -c ../../toughradius.yml -u admin -p toughradius
```

**SQLite 数据库:**

```bash
# 1. 编译重置密码工具 (需要 CGO)
cd cmd/reset-password
CGO_ENABLED=1 go build -o reset-password .

# 2. 运行工具
./reset-password -c ../../toughradius.yml -u admin -p toughradius
```

**注意**: 脚本会自动检测数据库类型并设置正确的编译参数。

### 方法 3: 直接使用默认密码

如果管理员账号尚未修改过密码，尝试使用默认密码登录：

- **用户名**: `admin`
- **密码**: `toughradius`

## 命令行参数

```bash
./reset-password -c <配置文件> [-u <用户名>] [-p <新密码>]
```

参数说明：

- `-c`: 配置文件路径 (必需)
- `-u`: 要重置密码的用户名 (默认: admin)
- `-p`: 新密码 (默认: toughradius)

## 示例

```bash
# 重置 admin 用户密码为 newpass123
./reset-password -c toughradius.yml -u admin -p newpass123

# 重置其他操作员密码
./reset-password -c toughradius.yml -u operator1 -p password123
```

## 注意事项

1. 运行此工具需要停止 ToughRADIUS 服务
2. 确保配置文件路径正确
3. 密码会被哈希加密后存储到数据库
4. 建议在生产环境使用强密码

## 故障排除

**问题**: "go-sqlite3 requires cgo to work"

- **原因**: SQLite 需要 CGO 支持
- **解决**: 使用 `CGO_ENABLED=1` 编译，或使用脚本自动处理
- **示例**: `CGO_ENABLED=1 go build -o reset-password .`

**问题**: "Failed to find user"

- 检查用户名是否正确
- 确认用户在数据库中存在

**问题**: "Failed to connect to database"

- 检查配置文件中的数据库连接信息
- 确保数据库服务正在运行
- SQLite: 确认数据库文件路径存在
- PostgreSQL: 确认 PostgreSQL 服务运行中

**问题**: 权限错误

- 使用 `chmod +x scripts/reset-admin-password.sh` 添加执行权限
