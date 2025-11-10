#!/bin/bash

# ToughRADIUS 管理员密码重置脚本
# 使用方法: ./reset-admin-password.sh [新密码]

set -e

CONFIG_FILE="toughradius.yml"
NEW_PASSWORD="${1:-toughradius}"

echo "========================================"
echo "ToughRADIUS 管理员密码重置工具"
echo "========================================"
echo ""

# 检查配置文件
if [ ! -f "$CONFIG_FILE" ]; then
    echo "错误: 找不到配置文件 $CONFIG_FILE"
    echo "请在 ToughRADIUS 根目录下运行此脚本"
    exit 1
fi

# 检测数据库类型
DB_TYPE=$(grep -A 5 "^database:" "$CONFIG_FILE" | grep "type:" | awk '{print $2}' | tr -d '"' || echo "postgres")

# 根据数据库类型决定是否启用 CGO
if [ "$DB_TYPE" = "sqlite" ]; then
    echo "检测到 SQLite 数据库，启用 CGO..."
    export CGO_ENABLED=1
else
    echo "检测到 $DB_TYPE 数据库，使用静态编译..."
    export CGO_ENABLED=0
fi

# 构建重置密码工具
echo "正在构建密码重置工具..."
cd cmd/reset-password
go build -o ../../reset-password .
cd ../..

# 运行重置密码工具
echo "正在重置管理员密码..."
./reset-password -c "$CONFIG_FILE" -u admin -p "$NEW_PASSWORD"

# 清理
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
