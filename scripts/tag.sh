#!/bin/bash

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🏷️  开始获取最新标签...${NC}"

# 获取最新标签
git fetch --tags

# 如果没有标签，返回 v0.0.0 作为兜底
latest_tag=$(git describe --tags `git rev-list --tags --max-count=1` 2>/dev/null || echo "v0.0.0")
echo -e "${YELLOW}📋 Latest tag: ${latest_tag}${NC}"

# 解析版本号
version=${latest_tag#v}
IFS='.' read -r -a parts <<<"$version"
last_idx=$((${#parts[@]} - 1))
parts[$last_idx]=$((${parts[$last_idx]} + 1))
new_version=$(IFS='.'; echo "${parts[*]}")
new_tag="v$new_version"

echo -e "${GREEN}🎯 New tag: ${new_tag}${NC}"

# 确认创建标签
echo -e -n "${YELLOW}确认创建标签 ${new_tag}? (y/n): ${NC}"
read confirm

if [ "$confirm" = "y" ] || [ "$confirm" = "Y" ]; then
    echo -e "${BLUE}🚀 创建标签 ${new_tag}...${NC}"
    git tag $new_tag

    echo -e "${BLUE}📤 推送标签到远程仓库...${NC}"
    git push origin $new_tag

    echo -e "${GREEN}✅ 标签 ${new_tag} 创建并推送成功！${NC}"
else
    echo -e "${RED}❌ 标签创建已取消${NC}"
fi
