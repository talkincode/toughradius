#!/bin/bash

set -e  # 只要遇到一个错误就退出
set -o pipefail

# 彩色输出函数
info() { echo -e "\033[1;34m[INFO]\033[0m $1"; }
error() { echo -e "\033[1;31m[ERROR]\033[0m $1"; }

# 确保当前是 develop 分支
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [[ "$CURRENT_BRANCH" != "develop" ]]; then
    error "请在 develop 分支运行此脚本（当前分支是 $CURRENT_BRANCH）"
    exit 1
fi

# 检查是否有未提交的内容
if ! git diff --quiet || ! git diff --cached --quiet; then
    error "检测到未提交的更改，请先提交或 stash"
    exit 1
fi

# 生成构建信息
info "生成构建信息..."
make buildpre

# 获取本次同步说明（不能为空）
while true; do
    read -p "📝 请输入本次同步说明（develop）: " COMMIT_MSG
    if [[ -n "$COMMIT_MSG" && "$COMMIT_MSG" != " "* && "$COMMIT_MSG" != *" " ]]; then
        break
    else
        error "同步说明不能为空或只包含空格，请重新输入"
    fi
done

# 设置为全局变量供后续使用
export COMMIT_MSG

# 提交当前进度
git commit --allow-empty -am "$(date '+%F %T') : $COMMIT_MSG"

# 切换并拉取 main 最新
info "切换到 main 并拉取远程最新代码..."
git checkout main
git pull origin main

# 回到 develop 进行变基
info "切换回 develop 执行 rebase 操作..."
git checkout develop
set +e
git rebase main
REBASE_STATUS=$?
set -e

# 检查是否发生冲突
if [[ $REBASE_STATUS -ne 0 ]]; then
    info "检测到 rebase 冲突，尝试自动处理 assets/buildinfo.txt"
    
    if git diff --name-only --diff-filter=U | grep -q 'assets/buildinfo.txt'; then
        info "保留 develop 版本（ours）: assets/buildinfo.txt"
        git checkout --ours assets/buildinfo.txt
        git add assets/buildinfo.txt
    fi
    
    # 检查是否还有其他冲突
    UNRESOLVED=$(git diff --name-only --diff-filter=U)
    if [[ -n "$UNRESOLVED" ]]; then
        error "还有其他未解决的冲突:\n$UNRESOLVED\n请手动处理后运行: git rebase --continue"
        exit 1
    fi
    
    info "继续 rebase..."
    git rebase --continue
fi

# rebase 完成后重新生成构建信息
info "重新生成最新的构建信息..."
make buildpre
if ! git diff --quiet assets/buildinfo.txt; then
    info "提交更新后的构建信息..."
    git add assets/buildinfo.txt
    git commit -m "$COMMIT_MSG: 📦 更新构建信息"
fi

# 合并到 main
info "切回 main 合并 develop..."
git checkout main
git merge --no-ff develop -m "$COMMIT_MSG: 🔀 合并 develop ($(date '+%F %T'))"

# 在 main 分支也生成最新的构建信息
info "在 main 分支生成最新构建信息..."
make buildpre
if ! git diff --quiet assets/buildinfo.txt; then
    info "提交 main 分支的构建信息..."
    git add assets/buildinfo.txt
    git commit -m "$COMMIT_MSG: 📦 更新 main 分支构建信息"
fi

# 推送
info "推送 main 分支到远程..."
git push origin main

# 回到 develop
git checkout develop
info "✅ 同步完成！"