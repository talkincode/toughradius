#!/bin/bash

set -e

echo "Building ToughRADIUS v9 Web UI..."

# 进入 web 目录
cd "$(dirname "$0")"

# 安装依赖
echo "Installing dependencies..."
npm install

# 构建生产版本
echo "Building production bundle..."
npm run build

echo "Build completed! Output: web/dist/"
echo ""
echo "To embed into Go binary, run:"
echo "  cd .. && go build"
