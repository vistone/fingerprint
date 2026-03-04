#!/bin/bash
# Fingerprint 项目 - Markdown 自动修复脚本
# 用途: 修复所有 Markdown lint 错误

set -e

echo "🔧 Markdown 自动修复工具"
echo "========================"
echo ""

# 检查是否安装了必要的工具
check_tools() {
    echo "检查必要工具..."

    # 检查 sed
    if ! command -v sed &> /dev/null; then
        echo "❌ sed 未安装"
        exit 1
    fi

    # 检查 find
    if ! command -v find &> /dev/null; then
        echo "❌ find 未安装"
        exit 1
    fi

    echo "✅ 工具检查通过"
    echo ""
}

# 修复常见的 Markdown 问题
fix_markdown_issues() {
    echo "修复 Markdown 问题..."

    # 1. 修复空代码块 (添加 plaintext 语言标识)
    echo "  - 修复空代码块..."
    find . -name "*.md" -not -path "./.git/*" -not -path "./vendor/*" -type f -exec sed -i 's/^```$/```plaintext/g' {} \;

    # 2. 在 heading 周围添加空行
    echo "  - 修复 heading 格式..."
    # 这个需要更复杂的处理，暂时跳过

    # 3. 在列表周围添加空行
    echo "  - 修复列表格式..."
    # 这个也需要更复杂的处理，暂时跳过

    # 4. 修复表格格式
    echo "  - 修复表格格式..."
    # 在表格分隔符的 | 两边添加空格
    find . -name "*.md" -not -path "./.git/*" -not -path "./vendor/*" -type f -exec sed -i 's/|---/| ---/g; s/---|/--- |/g' {} \;

    echo "✅ 基础修复完成"
    echo ""
}

# 手动处理需要注意的问题
print_manual_fixes() {
    echo "⚠️  以下问题需要手动修复:"
    echo ""
    echo "1. Heading 周围缺少空行 (MD022)"
    echo "   - 在每个 heading (###) 前后添加空行"
    echo ""
    echo "2. 列表周围缺少空行 (MD032)"
    echo "   - 在列表 (-) 前后添加空行"
    echo ""
    echo "3. 代码块语言标识 (MD040)"
    echo "   - 已自动添加 plaintext，但建议改为实际语言 (go, bash, yaml 等)"
    echo ""
}

# 验证修复结果
verify_fixes() {
    echo "验证修复结果..."

    # 统计仍存在的问题
    empty_code_blocks=$(grep -r "^\`\`\`$" --include="*.md" . 2>/dev/null | wc -l || echo "0")

    echo "  - 空代码块数量: $empty_code_blocks"

    if [ "$empty_code_blocks" -eq 0 ]; then
        echo "✅ 空代码块已全部修复"
    else
        echo "⚠️  仍有 $empty_code_blocks 个空代码块需要处理"
    fi

    echo ""
}

# 生成修复报告
generate_report() {
    echo "生成修复报告..."

    report_file="markdown_fix_report.txt"

    cat > "$report_file" << EOF
Markdown 修复报告
==================
日期: $(date)

已修复问题:
-----------
- ✅ 空代码块: 已添加 plaintext 标识
- ✅ 表格格式: 已在分隔符两边添加空格

需要手动修复:
-----------
- ⚠️  Heading 周围空行 (MD022)
- ⚠️  列表周围空行 (MD032)
- ⚠️  改进代码块语言标识 (建议使用具体语言而非 plaintext)

下一步操作:
-----------
1. 检查所有 .md 文件
2. 手动处理上述需要注意的问题
3. 运行 markdownlint 验证:
   npm install -g markdownlint-cli2
   markdownlint-cli2 "**/*.md"

EOF

    echo "✅ 报告已保存到: $report_file"
    echo ""
}

# 主函数
main() {
    check_tools
    fix_markdown_issues
    verify_fixes
    generate_report
    print_manual_fixes

    echo "✅ Markdown 自动修复完成!"
    echo ""
    echo "📝 请查看 markdown_fix_report.txt 了解详情"
    echo "📖 完整修复指南: docs/OPTIMIZATION_ROADMAP.md"
}

# 运行
main
