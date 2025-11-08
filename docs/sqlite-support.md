# ToughRADIUS SQLite 支持文档

## 概述

ToughRADIUS v9 现在支持两种数据库后端：

- **PostgreSQL**: 适合生产环境和大规模部署
- **SQLite**: 适合开发、测试和小规模部署

## 快速开始

### 1. 使用初始化脚本（推荐）

我们提供了一个便捷的初始化脚本来快速设置数据库：

```bash
# SQLite（开发/测试）
./scripts/init-db.sh sqlite

# PostgreSQL（生产）
./scripts/init-db.sh postgres

# 使用现有配置文件
./scripts/init-db.sh
```

### 2. 手动配置

#### SQLite 配置

1. **创建配置文件** `toughradius-sqlite.yml`:

```yaml
database:
  type: sqlite
  name: toughradius.db # 相对路径（存储在 workdir/data/）
  # 或使用绝对路径: /path/to/toughradius.db
  max_conn: 100
  idle_conn: 10
  debug: false
```

2. **初始化数据库**:

```bash
# 编译（需要 CGO）
CGO_ENABLED=1 go build -o toughradius

# 初始化
./toughradius -c toughradius-sqlite.yml -initdb
```

3. **启动服务**:

```bash
./toughradius -c toughradius-sqlite.yml
```

#### PostgreSQL 配置

1. **创建配置文件** `toughradius.yml`:

```yaml
database:
  type: postgres
  host: 127.0.0.1
  port: 5432
  name: toughradius
  user: postgres
  passwd: yourpassword
  max_conn: 100
  idle_conn: 10
  debug: false
```

2. **初始化和启动** 同上

## 编译说明

### 本地开发（支持 SQLite）

```bash
# 使用 Makefile
make build-local

# 或直接使用 go
CGO_ENABLED=1 go build -o toughradius
```

### 生产构建

```bash
# PostgreSQL 版本（静态编译，无 CGO 依赖）
make build

# SQLite 版本（需要 CGO）
make build-sqlite
```

## 注意事项

### SQLite

**优点**:

- 零配置，无需单独的数据库服务
- 轻量级，适合开发和测试
- 数据文件便于备份和迁移

**限制**:

- 需要 CGO 编译（编译时需要设置 `CGO_ENABLED=1`）
- 不适合高并发场景
- 连接数限制（建议设置为 1）

**编译要求**:

- macOS: 需要 Xcode Command Line Tools
- Linux: 需要 gcc
- Windows: 需要 MinGW-w64

### PostgreSQL

**优点**:

- 高性能，适合生产环境
- 支持高并发
- 可以静态编译（无 CGO 依赖）

**要求**:

- 需要单独安装 PostgreSQL 服务

## 默认账号

初始化后的默认管理员账号：

```
管理员账号:
  用户名: admin
  密码:   toughradius

API 用户账号:
  用户名: apiuser
  密码:   Api_189
```

## 数据库文件位置

使用 SQLite 时，默认数据库文件位置：

```
/var/toughradius/data/toughradius.db
```

可以在配置文件中指定其他路径：

```yaml
database:
  type: sqlite
  name: /custom/path/toughradius.db  # 绝对路径
  # 或
  name: mydb.db  # 相对路径，实际位置: workdir/data/mydb.db
```

## 常见问题

### Q: 编译时出现 "Binary was compiled with 'CGO_ENABLED=0'" 错误？

**A**: 使用 SQLite 时必须启用 CGO：

```bash
CGO_ENABLED=1 go build -o toughradius
```

### Q: 如何在 Docker 中使用 SQLite？

**A**: 在 Dockerfile 中需要安装编译工具：

```dockerfile
FROM golang:1.21-alpine AS builder
RUN apk add --no-cache gcc musl-dev
ENV CGO_ENABLED=1
...
```

### Q: 可以在运行时切换数据库吗？

**A**: 可以，只需修改配置文件并重启服务。但需要重新初始化新数据库。

### Q: 如何迁移数据？

**A**:

1. SQLite → PostgreSQL: 使用 `pgloader` 工具
2. PostgreSQL → SQLite: 使用导出/导入脚本
3. 建议使用相同类型的数据库避免迁移问题

## Web 管理界面

启动后访问：`http://localhost:1816/admin`

## 技术支持

- 文档: https://github.com/talkincode/toughradius
- Issues: https://github.com/talkincode/toughradius/issues
