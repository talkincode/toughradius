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

# 生成标签消息
echo -e "${BLUE}📝 生成标签消息...${NC}"

# 创建临时文件存储标签消息
tag_message_file=$(mktemp)

# 确定提交范围
if [ "$latest_tag" = "v0.0.0" ]; then
    commit_range="HEAD"
else
    commit_range="$latest_tag..HEAD"
fi

# 生成提交日志清单
git log $commit_range --format='- %h %s (%an, %ad)' --date=short 2>/dev/null > "$tag_message_file"

# 显示标签消息预览
echo -e "${YELLOW}📋 标签消息预览:${NC}"
echo "----------------------------------------"
head -20 "$tag_message_file"
if [ $(wc -l < "$tag_message_file") -gt 20 ]; then
    echo "..."
fi
echo "----------------------------------------"

# 确认创建标签
echo -e -n "${YELLOW}确认创建带消息的标签 ${new_tag}? (y/n): ${NC}"
read confirm

if [ "$confirm" = "y" ] || [ "$confirm" = "Y" ]; then
    echo -e "${BLUE}🚀 创建带消息的标签 ${new_tag}...${NC}"

    # 使用标签消息文件创建注释标签
    git tag -a $new_tag -F "$tag_message_file"

    echo -e "${BLUE}📤 推送标签到远程仓库...${NC}"
    git push origin $new_tag

    echo -e "${GREEN}✅ 标签 ${new_tag} 创建并推送成功！${NC}"
    echo -e "${GREEN}📋 标签包含完整的变更清单${NC}"

    # 清理临时文件
    rm -f "$tag_message_file"
else
    echo -e "${RED}❌ 标签创建已取消${NC}"
    # 清理临时文件
    rm -f "$tag_message_file"
fi
