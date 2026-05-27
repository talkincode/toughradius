#!/bin/bash

# release-text.sh - Generate release notes
# This script generates a changelog starting from the latest tag

# Check if running inside a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "❌ Error: Current directory is not a git repository"
    exit 1
fi

# Get the latest tag
latest_tag=$(git tag --sort=-version:refname | head -n 1)
current_commit=$(git rev-parse HEAD)

# Set the commit range
if [ -z "$latest_tag" ]; then
    echo "⚠️  No tags found, showing all commits"
    commit_range="HEAD"
    version_info="Initial commit → HEAD"
else
    # Check if HEAD is already at the latest tag
    latest_tag_commit=$(git rev-parse "$latest_tag" 2>/dev/null || echo "")
    
    if [ "$current_commit" = "$latest_tag_commit" ]; then
        # If so, use the previous tag as the starting point
        prev_tag=$(git tag --sort=-version:refname | sed -n '2p')
        if [ -n "$prev_tag" ]; then
            echo "📋  $prev_tag —— $latest_tag "
            commit_range="$prev_tag..$latest_tag"
            version_info="$prev_tag → $latest_tag"
        else
            echo "📋 —— $latest_tag"
            commit_range="$latest_tag"
            version_info="Initial commit → $latest_tag"
        fi
    else
        echo "📋 $latest_tag —— "
        commit_range="$latest_tag..HEAD"
        version_info="$latest_tag → HEAD"
    fi
fi


echo ""

# Statistics
total_commits=$(git rev-list --count $commit_range 2>/dev/null || echo "0")
files_changed=$(git diff --name-only $commit_range 2>/dev/null | wc -l | tr -d ' ')
authors=$(git log $commit_range --format='%an' 2>/dev/null | sort -u | wc -l | tr -d ' ')

echo "📊 Statistics:"
echo "   • Commits: $total_commits"
echo "   • Files changed: $files_changed"
echo "   • Authors: $authors"
echo ""

# Display the changelog
echo "📝 Changelog:"
echo ""

# Categorize commits
git log $commit_range --format='%h|%s|%an|%ad' --date=short 2>/dev/null | {
    feat_count=0
    fix_count=0
    refactor_count=0
    other_count=0
    
    # Build temporary arrays
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
    
    # Display categorized results
    if [ $feat_count -gt 0 ]; then
        echo "🚀 New Features ($feat_count):"
        echo -e "$feat_commits"
    fi
    
    if [ $fix_count -gt 0 ]; then
        echo "🐛 Bug Fixes ($fix_count):"
        echo -e "$fix_commits"
    fi
    
    if [ $refactor_count -gt 0 ]; then
        echo "♻️  Refactoring/Optimization ($refactor_count):"
        echo -e "$refactor_commits"
    fi
    
    if [ $other_count -gt 0 ]; then
        echo "📦 Other Changes ($other_count):"
        echo -e "$other_commits"
    fi
}

# Display file change stats (top 10 lines)
if [ "$files_changed" -gt 0 ]; then
    echo "📄 Main File Changes:"
    git diff --stat $commit_range 2>/dev/null | head -10 | sed 's/^/   /'
    echo ""
fi

# Display contributors
if [ "$authors" -gt 0 ]; then
    echo "👥 Contributors:"
    git log $commit_range --format='%an <%ae>' 2>/dev/null | sort -u | sed 's/^/   • /'
fi
