#!/bin/bash

###############################################################################
# ToughRADIUS database initialization script
#
# Purpose: quickly initialize the ToughRADIUS database (PostgreSQL or SQLite supported)
#
# Usage:
#   1. PostgreSQL: ./scripts/init-db.sh postgres
#   2. SQLite:     ./scripts/init-db.sh sqlite
#   3. Use a config file: ./scripts/init-db.sh
#
# Examples:
#   ./scripts/init-db.sh sqlite
#   ./scripts/init-db.sh postgres
###############################################################################

set -e

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Determine the script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}ToughRADIUS 数据库初始化脚本${NC}"
echo -e "${GREEN}================================${NC}"
echo ""

# Check arguments
DB_TYPE="${1:-}"

if [ -z "$DB_TYPE" ]; then
    echo -e "${YELLOW}未指定数据库类型，将使用配置文件中的设置${NC}"
    echo ""
    cd "$PROJECT_ROOT"
    
    # Check if the tool has been compiled
    if [ ! -f "./toughradius" ]; then
        echo -e "${YELLOW}未找到编译文件，开始编译...${NC}"
        go build -o toughradius
        echo -e "${GREEN}✓ 编译完成${NC}"
    fi
    
    echo -e "${YELLOW}开始初始化数据库...${NC}"
    ./toughradius -initdb
    
    echo ""
    echo -e "${GREEN}================================${NC}"
    echo -e "${GREEN}数据库初始化完成！${NC}"
    echo -e "${GREEN}================================${NC}"
    echo ""
    echo -e "${GREEN}默认管理员账号：${NC}"
    echo -e "  用户名: ${YELLOW}admin${NC}"
    echo -e "  密码:   ${YELLOW}toughradius${NC}"
    echo ""
    echo -e "${GREEN}API 用户账号：${NC}"
    echo -e "  用户名: ${YELLOW}apiuser${NC}"
    echo -e "  密码:   ${YELLOW}Api_189${NC}"
    echo ""
    
    exit 0
fi

# Parameterized initialization
case "$DB_TYPE" in
    sqlite)
        echo -e "${GREEN}使用 SQLite 数据库${NC}"
        echo ""
        
        # Create a temporary configuration file
        TEMP_CONFIG="/tmp/toughradius-sqlite.yml"
        cat > "$TEMP_CONFIG" <<EOF
system:
  appid: ToughRADIUS
  location: Asia/Shanghai
  workdir: /var/toughradius
  debug: true

web:
  host: 0.0.0.0
  port: 1816
  tls_port: 1817
  secret: 9b6de5cc-0731-1203-xxtt-0f568ac9da37

database:
  type: sqlite
  name: toughradius.db
  max_conn: 100
  idle_conn: 10
  debug: false

freeradius:
  enabled: true
  host: 0.0.0.0
  port: 1818
  debug: true

radiusd:
  enabled: true
  host: 0.0.0.0
  auth_port: 1812
  acct_port: 1813
  radsec_port: 2083
  radsec_worker: 100
  debug: true

logger:
  mode: development
  file_enable: true
  filename: /var/toughradius/toughradius.log
EOF
        
        echo -e "${YELLOW}SQLite 数据库配置：${NC}"
        echo -e "  数据库文件: ${YELLOW}/var/toughradius/data/toughradius.db${NC}"
        echo ""
        
        cd "$PROJECT_ROOT"
        
        # Check if the tool has been compiled
        if [ ! -f "./toughradius" ]; then
            echo -e "${YELLOW}未找到编译文件，开始编译...${NC}"
            go build -o toughradius
            echo -e "${GREEN}✓ 编译完成${NC}"
        fi
        
        # Create the data directory
        sudo mkdir -p /var/toughradius/data
        sudo chown -R $(whoami) /var/toughradius
        
        echo -e "${YELLOW}开始初始化 SQLite 数据库...${NC}"
        ./toughradius -c "$TEMP_CONFIG" -initdb
        
        # Clean up temporary configuration
        rm -f "$TEMP_CONFIG"
        
        echo ""
        echo -e "${GREEN}================================${NC}"
        echo -e "${GREEN}SQLite 数据库初始化完成！${NC}"
        echo -e "${GREEN}================================${NC}"
        echo ""
        echo -e "${GREEN}数据库位置：${NC}"
        echo -e "  ${YELLOW}/var/toughradius/data/toughradius.db${NC}"
        echo ""
        echo -e "${GREEN}启动命令（使用 SQLite）：${NC}"
        echo -e "  ${YELLOW}./toughradius -c $TEMP_CONFIG${NC}"
        echo ""
        echo -e "${GREEN}或者修改 toughradius.yml 配置文件：${NC}"
        cat <<EOF
  ${YELLOW}database:
    type: sqlite
    name: toughradius.db${NC}
EOF
        echo ""
    ;;
    
    postgres|postgresql)
        echo -e "${GREEN}使用 PostgreSQL 数据库${NC}"
        echo ""
        
        # Load the configuration
        read -p "PostgreSQL 主机 [127.0.0.1]: " PG_HOST
        PG_HOST=${PG_HOST:-127.0.0.1}
        
        read -p "PostgreSQL 端口 [5432]: " PG_PORT
        PG_PORT=${PG_PORT:-5432}
        
        read -p "数据库名称 [toughradius]: " PG_DB
        PG_DB=${PG_DB:-toughradius}
        
        read -p "数据库用户 [postgres]: " PG_USER
        PG_USER=${PG_USER:-postgres}
        
        read -sp "数据库密码: " PG_PASS
        echo ""
        
        # Create a temporary configuration file
        TEMP_CONFIG="/tmp/toughradius-postgres.yml"
        cat > "$TEMP_CONFIG" <<EOF
system:
  appid: ToughRADIUS
  location: Asia/Shanghai
  workdir: /var/toughradius
  debug: true

web:
  host: 0.0.0.0
  port: 1816
  tls_port: 1817
  secret: 9b6de5cc-0731-1203-xxtt-0f568ac9da37

database:
  type: postgres
  host: $PG_HOST
  port: $PG_PORT
  name: $PG_DB
  user: $PG_USER
  passwd: $PG_PASS
  max_conn: 100
  idle_conn: 10
  debug: false

freeradius:
  enabled: true
  host: 0.0.0.0
  port: 1818
  debug: true

radiusd:
  enabled: true
  host: 0.0.0.0
  auth_port: 1812
  acct_port: 1813
  radsec_port: 2083
  radsec_worker: 100
  debug: true

logger:
  mode: development
  file_enable: true
  filename: /var/toughradius/toughradius.log
EOF
        
        echo ""
        echo -e "${YELLOW}PostgreSQL 数据库配置：${NC}"
        echo -e "  主机: ${YELLOW}$PG_HOST:$PG_PORT${NC}"
        echo -e "  数据库: ${YELLOW}$PG_DB${NC}"
        echo -e "  用户: ${YELLOW}$PG_USER${NC}"
        echo ""
        
        cd "$PROJECT_ROOT"
        
        # Check if the tool has been compiled
        if [ ! -f "./toughradius" ]; then
            echo -e "${YELLOW}未找到编译文件，开始编译...${NC}"
            go build -o toughradius
            echo -e "${GREEN}✓ 编译完成${NC}"
        fi
        
        echo -e "${YELLOW}开始初始化 PostgreSQL 数据库...${NC}"
        ./toughradius -c "$TEMP_CONFIG" -initdb
        
        # Clean up temporary configuration
        rm -f "$TEMP_CONFIG"
        
        echo ""
        echo -e "${GREEN}================================${NC}"
        echo -e "${GREEN}PostgreSQL 数据库初始化完成！${NC}"
        echo -e "${GREEN}================================${NC}"
        echo ""
    ;;
    
    *)
        echo -e "${RED}错误：不支持的数据库类型 '$DB_TYPE'${NC}"
        echo ""
        echo -e "${YELLOW}支持的数据库类型：${NC}"
        echo "  - sqlite"
        echo "  - postgres (或 postgresql)"
        echo ""
        echo -e "${YELLOW}使用示例：${NC}"
        echo "  ./scripts/init-db.sh sqlite"
        echo "  ./scripts/init-db.sh postgres"
        echo ""
        exit 1
    ;;
esac

echo -e "${GREEN}默认管理员账号：${NC}"
echo -e "  用户名: ${YELLOW}admin${NC}"
echo -e "  密码:   ${YELLOW}toughradius${NC}"
echo ""
echo -e "${GREEN}API 用户账号：${NC}"
echo -e "  用户名: ${YELLOW}apiuser${NC}"
echo -e "  密码:   ${YELLOW}Api_189${NC}"
echo ""
echo -e "${GREEN}Web 管理界面：${NC}"
echo -e "  ${YELLOW}http://localhost:1816/admin${NC}"
echo ""
