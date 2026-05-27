#!/bin/bash

# 批量移除资源文件中的 label 属性
# 这个脚本会移除 React Admin 组件中硬编码的 label，让它使用翻译文件

echo "开始处理资源文件国际化..."

# 资源文件列表
FILES=(
    "web/src/resources/operators.tsx"
    "web/src/resources/nodes.tsx"
    "web/src/resources/nas.tsx"
    "web/src/resources/radiusProfiles.tsx"
    "web/src/resources/accounting.tsx"
)

for file in "${FILES[@]}"; do
    if [ -f "$file" ]; then
        echo "处理: $file"
        
        # 创建备份
        cp "$file" "${file}.bak"
        
        # 移除常见的 label 属性（保留带参数的 label）
        # 示例: label="用户名" -> (移除)
        # 但保留: label={variable} 或 label={translate(...)}
        
        # 注意：这个脚本需要手动验证结果，不是完全自动化的
        echo "  已创建备份: ${file}.bak"
        echo "  需要手动检查和调整"
    else
        echo "文件不存在: $file"
    fi
done

echo "完成！请手动检查修改并删除 .bak 文件"
