# 环境变量配置说明

ToughRADIUS 支持通过环境变量进行配置，优先级为：**环境变量 > 配置文件 > 默认值**

## 快速开始

### 方法一：使用 .env 文件（推荐）

1. **复制示例文件**：

```bash
cp .env.example .env
```

2. **编辑配置**（可选，默认已配置好 SQLite）：

```bash
vi .env
```

3. **使用启动脚本**：

```bash
# 初始化数据库
./run.sh -initdb

# 启动服务
./run.sh
```

### 方法二：直接设置环境变量

```bash
# 设置环境变量
export TOUGHRADIUS_DB_TYPE=sqlite
export TOUGHRADIUS_DB_NAME=toughradius.db

# 初始化
./toughradius -initdb

# 启动
./toughradius
```

### 方法三：使用配置文件

```bash
# 使用自定义配置文件
./toughradius -c toughradius.yml
```

## 环境变量列表

### 系统配置

| 环境变量                        | 说明     | 默认值             |
| ------------------------------- | -------- | ------------------ |
| `TOUGHRADIUS_SYSTEM_WORKER_DIR` | 工作目录 | `/var/toughradius` |
| `TOUGHRADIUS_SYSTEM_DEBUG`      | 调试模式 | `true`             |

### Web 服务

| 环境变量                   | 说明       | 默认值     |
| -------------------------- | ---------- | ---------- |
| `TOUGHRADIUS_WEB_HOST`     | 监听地址   | `0.0.0.0`  |
| `TOUGHRADIUS_WEB_PORT`     | HTTP 端口  | `1816`     |
| `TOUGHRADIUS_WEB_TLS_PORT` | HTTPS 端口 | `1817`     |
| `TOUGHRADIUS_WEB_SECRET`   | JWT 密钥   | (随机生成) |

### 数据库配置

| 环境变量               | 说明                             | 默认值           |
| ---------------------- | -------------------------------- | ---------------- |
| `TOUGHRADIUS_DB_TYPE`  | 数据库类型 (`sqlite`/`postgres`) | `sqlite`         |
| `TOUGHRADIUS_DB_NAME`  | 数据库名称/文件                  | `toughradius.db` |
| `TOUGHRADIUS_DB_HOST`  | PostgreSQL 主机                  | `127.0.0.1`      |
| `TOUGHRADIUS_DB_PORT`  | PostgreSQL 端口                  | `5432`           |
| `TOUGHRADIUS_DB_USER`  | PostgreSQL 用户                  | `postgres`       |
| `TOUGHRADIUS_DB_PWD`   | PostgreSQL 密码                  | -                |
| `TOUGHRADIUS_DB_DEBUG` | 数据库调试                       | `false`          |

### RADIUS 服务

| 环境变量                           | 说明            | 默认值    |
| ---------------------------------- | --------------- | --------- |
| `TOUGHRADIUS_RADIUS_ENABLED`       | 启用 RADIUS     | `true`    |
| `TOUGHRADIUS_RADIUS_HOST`          | 监听地址        | `0.0.0.0` |
| `TOUGHRADIUS_RADIUS_AUTHPORT`      | 认证端口        | `1812`    |
| `TOUGHRADIUS_RADIUS_ACCTPORT`      | 计费端口        | `1813`    |
| `TOUGHRADIUS_RADIUS_RADSEC_PORT`   | RadSec 端口     | `2083`    |
| `TOUGHRADIUS_RADIUS_RADSEC_WORKER` | RadSec 工作线程 | `100`     |
| `TOUGHRADIUS_RADIUS_DEBUG`         | RADIUS 调试     | `true`    |

### 日志配置

| 环境变量                         | 说明         | 默认值        |
| -------------------------------- | ------------ | ------------- |
| `TOUGHRADIUS_LOGGER_MODE`        | 日志模式     | `development` |
| `TOUGHRADIUS_LOGGER_FILE_ENABLE` | 启用文件日志 | `true`        |

## 配置示例

### SQLite 配置（默认）

```bash
# .env
TOUGHRADIUS_DB_TYPE=sqlite
TOUGHRADIUS_DB_NAME=toughradius.db
```

### PostgreSQL 配置

```bash
# .env
TOUGHRADIUS_DB_TYPE=postgres
TOUGHRADIUS_DB_HOST=127.0.0.1
TOUGHRADIUS_DB_PORT=5432
TOUGHRADIUS_DB_NAME=toughradius
TOUGHRADIUS_DB_USER=postgres
TOUGHRADIUS_DB_PWD=mypassword
```

### 生产环境配置

```bash
# .env
TOUGHRADIUS_SYSTEM_DEBUG=false
TOUGHRADIUS_DB_TYPE=postgres
TOUGHRADIUS_DB_HOST=db.example.com
TOUGHRADIUS_DB_NAME=toughradius_prod
TOUGHRADIUS_LOGGER_MODE=production
```

## 配置优先级

1. **环境变量**（最高优先级）
2. **配置文件** (`-c` 参数指定的文件)
3. **默认配置**（最低优先级）

## 注意事项

1. `.env` 文件不会被提交到 Git（已在 .gitignore 中）
2. 修改环境变量后需要重启服务
3. 布尔值使用 `true`/`false` 或 `1`/`0` 或 `on`/`off`
4. 使用 SQLite 无需 CGO: `CGO_ENABLED=0 go build`

## Docker 使用

在 `docker-compose.yml` 中使用环境变量：

```yaml
version: "3"
services:
  toughradius:
    image: toughradius:latest
    environment:
      - TOUGHRADIUS_DB_TYPE=postgres
      - TOUGHRADIUS_DB_HOST=postgres
      - TOUGHRADIUS_DB_NAME=toughradius
      - TOUGHRADIUS_DB_USER=postgres
      - TOUGHRADIUS_DB_PWD=password
    ports:
      - "1816:1816"
      - "1812:1812/udp"
      - "1813:1813/udp"
```

或使用 `.env` 文件：

```yaml
version: "3"
services:
  toughradius:
    image: toughradius:latest
    env_file:
      - .env
    ports:
      - "1816:1816"
```
