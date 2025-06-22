#!/bin/bash

# release-text.sh - 生成发布信息脚本
# 该脚本从最后一个标签开始生成提交日志清单

# 检查是否在git仓库中
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "❌ 错误: 当前目录不是git仓库"
    exit 1
fi

# 获取最新标签
latest_tag=$(git tag --sort=-version:refname | head -n 1)
current_commit=$(git rev-parse HEAD)

# 设置提交范围
if [ -z "$latest_tag" ]; then
    echo "⚠️  未找到任何标签，显示所有提交记录"
    commit_range="HEAD"
    version_info="初始提交 → HEAD"
else
    # 检查当前HEAD是否就是最新标签
    latest_tag_commit=$(git rev-parse "$latest_tag" 2>/dev/null || echo "")
    
    if [ "$current_commit" = "$latest_tag_commit" ]; then
        # 如果当前HEAD就是最新标签，则从倒数第二个标签开始
        prev_tag=$(git tag --sort=-version:refname | sed -n '2p')
        if [ -n "$prev_tag" ]; then
            echo "📋  $prev_tag —— $latest_tag "
            commit_range="$prev_tag..$latest_tag"
            version_info="$prev_tag → $latest_tag"
        else
            echo "📋 —— $latest_tag"
            commit_range="$latest_tag"
            version_info="初始提交 → $latest_tag"
        fi
    else
        echo "📋 $latest_tag —— "
        commit_range="$latest_tag..HEAD"
        version_info="$latest_tag → HEAD"
    fi
fi


echo ""

# 统计信息
total_commits=$(git rev-list --count $commit_range 2>/dev/null || echo "0")
files_changed=$(git diff --name-only $commit_range 2>/dev/null | wc -l | tr -d ' ')
authors=$(git log $commit_range --format='%an' 2>/dev/null | sort -u | wc -l | tr -d ' ')

echo "📊 统计信息:"
echo "   • 提交数量: $total_commits"
echo "   • 文件变更: $files_changed"
echo "   • 参与作者: $authors"
echo ""

# 显示变更清单
echo "📝 变更清单:"
echo ""

# 分类显示提交
git log $commit_range --format='%h|%s|%an|%ad' --date=short 2>/dev/null | {
    feat_count=0
    fix_count=0
    refactor_count=0
    other_count=0
    
    # 创建临时数组
    feat_commits=""
    fix_commits=""
    refactor_commits=""
    other_commits=""
    
    while IFS='|' read -r hash subject author date; do
        line="   • $hash $subject ($author, $date)"
        
        case "$subject" in
            feat*|feature*|新增*|添加*|增加*)
                feat_commits="$feat_commits$line\n"
                feat_count=$((feat_count + 1))
            ;;
            fix*|修复*|修正*|bugfix*)
                fix_commits="$fix_commits$line\n"
                fix_count=$((fix_count + 1))
            ;;
            refactor*|重构*|优化*)
                refactor_commits="$refactor_commits$line\n"
                refactor_count=$((refactor_count + 1))
            ;;
            *)
                other_commits="$other_commits$line\n"
                other_count=$((other_count + 1))
            ;;
        esac
    done
    
    # 显示分类结果
    if [ $feat_count -gt 0 ]; then
        echo "🚀 新功能 ($feat_count):"
        echo -e "$feat_commits"
    fi
    
    if [ $fix_count -gt 0 ]; then
        echo "🐛 Bug修复 ($fix_count):"
        echo -e "$fix_commits"
    fi
    
    if [ $refactor_count -gt 0 ]; then
        echo "♻️  重构/优化 ($refactor_count):"
        echo -e "$refactor_commits"
    fi
    
    if [ $other_count -gt 0 ]; then
        echo "📦 其他变更 ($other_count):"
        echo -e "$other_commits"
    fi
}

# 显示文件变更统计（仅前10行）
if [ "$files_changed" -gt 0 ]; then
    echo "📄 主要文件变更:"
    git diff --stat $commit_range 2>/dev/null | head -10 | sed 's/^/   /'
    echo ""
fi

# 显示贡献者
if [ "$authors" -gt 0 ]; then
    echo "👥 贡献者:"
    git log $commit_range --format='%an <%ae>' 2>/dev/null | sort -u | sed 's/^/   • /'
fi
