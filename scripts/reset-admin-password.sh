#!/bin/bash

# ToughRADIUS admin password reset script
# Usage: ./reset-admin-password.sh [new password]

set -e

CONFIG_FILE="toughradius.yml"
NEW_PASSWORD="${1:-toughradius}"

echo "========================================"
echo "ToughRADIUS 管理员密码重置工具"
echo "========================================"
echo ""

# Check the configuration file
if [ ! -f "$CONFIG_FILE" ]; then
    echo "错误: 找不到配置文件 $CONFIG_FILE"
    echo "请在 ToughRADIUS 根目录下运行此脚本"
    exit 1
fi

# Detect the database type
DB_TYPE=$(grep -A 5 "^database:" "$CONFIG_FILE" | grep "type:" | awk '{print $2}' | tr -d '"' || echo "postgres")

# Decide whether to enable CGO based on the database type
if [ "$DB_TYPE" = "sqlite" ]; then
    echo "检测到 SQLite 数据库，启用 CGO..."
    export CGO_ENABLED=1
else
    echo "检测到 $DB_TYPE 数据库，使用静态编译..."
    export CGO_ENABLED=0
fi

# Build the password reset tool
echo "正在构建密码重置工具..."
cd cmd/reset-password
go build -o ../../reset-password .
cd ../..

# Run the password reset tool
echo "正在重置管理员密码..."
./reset-password -c "$CONFIG_FILE" -u admin -p "$NEW_PASSWORD"

# Clean up
rm -f reset-password

echo ""
echo "========================================"
echo "密码重置完成!"
echo "========================================"
echo ""
echo "登录信息:"
echo "  用户名: admin"
echo "  密码: $NEW_PASSWORD"
echo "  访问地址: http://localhost:1816"
echo ""
