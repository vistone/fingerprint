#!/bin/bash
# Fingerprint 项目 - 测试覆盖率分析工具
# 用途: 分析测试覆盖率并生成改进建议

set -e

COVERAGE_TARGET=75
REPORT_FILE="coverage_analysis.txt"
HTML_FILE="coverage_analysis.html"

echo "📊 测试覆盖率分析工具"
echo "======================"
echo "目标覆盖率: ${COVERAGE_TARGET}%"
echo ""

# 颜色定义
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# 运行测试并生成覆盖率报告
run_coverage() {
    echo "运行测试并生成覆盖率报告..."

    # 运行测试
    go test -cover -coverprofile=coverage.out ./... > /dev/null 2>&1 || {
        echo "❌ 测试失败，请先修复失败的测试"
        exit 1
    }

    # 生成 HTML 报告
    go tool cover -html=coverage.out -o coverage.html

    echo "✅ 覆盖率报告已生成"
    echo "   - coverage.out (原始数据)"
    echo "   - coverage.html (可视化)"
    echo ""
}

# 分析各包的覆盖率
analyze_coverage() {
    echo "分析各包覆盖率..."
    echo ""

    # 创建报告
    {
        echo "测试覆盖率分析报告"
        echo "=================="
        echo "生成时间: $(date)"
        echo "目标覆盖率: ${COVERAGE_TARGET}%"
        echo ""
        echo "包覆盖率统计"
        echo "============"
        echo ""
    } > "$REPORT_FILE"

    # 统计变量
    total_packages=0
    zero_coverage=0
    below_target=0
    above_target=0

    # 解析覆盖率
    while IFS= read -r line; do
        if [[ $line =~ ^ok.*coverage:\ ([0-9.]+)% ]]; then
            coverage="${BASH_REMATCH[1]}"
            package=$(echo "$line" | awk '{print $2}')

            total_packages=$((total_packages + 1))

            # 分类
            if [ "$(echo "$coverage == 0" | bc -l)" -eq 1 ]; then
                zero_coverage=$((zero_coverage + 1))
                echo -e "${RED}❌${NC} $package: ${RED}${coverage}%${NC} (无覆盖)"
                echo "❌ $package: ${coverage}% (无覆盖)" >> "$REPORT_FILE"
            elif [ "$(echo "$coverage < $COVERAGE_TARGET" | bc -l)" -eq 1 ]; then
                below_target=$((below_target + 1))
                echo -e "${YELLOW}⚠️${NC}  $package: ${YELLOW}${coverage}%${NC} (低于目标)"
                echo "⚠️  $package: ${coverage}% (低于目标)" >> "$REPORT_FILE"
            else
                above_target=$((above_target + 1))
                echo -e "${GREEN}✅${NC} $package: ${GREEN}${coverage}%${NC}"
                echo "✅ $package: ${coverage}%" >> "$REPORT_FILE"
            fi
        elif [[ $line =~ coverage:\ ([0-9.]+)% ]]; then
            # 处理没有测试文件的包
            coverage="${BASH_REMATCH[1]}"
            package=$(echo "$line" | awk '{print $1}')

            total_packages=$((total_packages + 1))

            if [ "$(echo "$coverage == 0" | bc -l)" -eq 1 ]; then
                zero_coverage=$((zero_coverage + 1))
                echo -e "${RED}❌${NC} $package: ${RED}${coverage}%${NC} (无覆盖)"
                echo "❌ $package: ${coverage}% (无覆盖)" >> "$REPORT_FILE"
            fi
        fi
    done < <(go test -cover ./... 2>&1)

    echo ""

    # 生成统计摘要
    {
        echo ""
        echo "统计摘要"
        echo "========"
        echo "总包数: $total_packages"
        echo "无覆盖 (0%): $zero_coverage"
        echo "低于目标 (<${COVERAGE_TARGET}%): $below_target"
        echo "达标 (≥${COVERAGE_TARGET}%): $above_target"
        echo ""
        echo "达标率: $(echo "scale=2; $above_target * 100 / $total_packages" | bc)%"
    } | tee -a "$REPORT_FILE"

    echo ""
}

# 生成改进建议
generate_recommendations() {
    echo "生成改进建议..."
    echo ""

    {
        echo ""
        echo "改进建议"
        echo "========"
        echo ""
        echo "🎯 优先级 1: 无覆盖包 (0%)"
        echo "----------------------------"
    } >> "$REPORT_FILE"

    # 列出所有 0% 覆盖率的包
    go test -cover ./... 2>&1 | grep "coverage: 0.0%" | awk '{print $2}' | while read -r pkg; do
        echo "  - $pkg" >> "$REPORT_FILE"

        # 生成建议
        {
            echo "    建议:"
            echo "      1. 创建测试文件: ${pkg}_test.go"
            echo "      2. 添加基础单元测试"
            echo "      3. 目标: 50%+ 覆盖率"
            echo ""
        } >> "$REPORT_FILE"
    done

    {
        echo ""
        echo "🎯 优先级 2: 低于目标包 (<${COVERAGE_TARGET}%)"
        echo "-------------------------------------------"
    } >> "$REPORT_FILE"

    # 列出低于目标的包
    go test -cover ./... 2>&1 | grep -E "coverage: [0-9.]+" | while read -r line; do
        if [[ $line =~ coverage:\ ([0-9.]+)% ]]; then
            coverage="${BASH_REMATCH[1]}"
            package=$(echo "$line" | awk '{print $2}')

            if [ "$(echo "$coverage > 0 && $coverage < $COVERAGE_TARGET" | bc -l)" -eq 1 ]; then
                echo "  - $package (当前: ${coverage}%)" >> "$REPORT_FILE"
                {
                    echo "    建议:"
                    echo "      1. 增加边界条件测试"
                    echo "      2. 增加错误处理测试"
                    echo "      3. 增加并发安全测试"
                    echo "      4. 目标: ${COVERAGE_TARGET}%+"
                    echo ""
                } >> "$REPORT_FILE"
            fi
        fi
    done

    {
        echo ""
        echo "📝 测试编写指南"
        echo "==============="
        echo ""
        echo "1. 表驱动测试示例:"
        echo "   func TestFunction(t *testing.T) {"
        echo "       tests := []struct {"
        echo "           name    string"
        echo "           input   Type"
        echo "           want    Type"
        echo "           wantErr bool"
        echo "       }{"
        echo "           {\"case1\", input1, want1, false},"
        echo "           {\"case2\", input2, want2, true},"
        echo "       }"
        echo "       for _, tt := range tests {"
        echo "           t.Run(tt.name, func(t *testing.T) {"
        echo "               got, err := Function(tt.input)"
        echo "               if (err != nil) != tt.wantErr {"
        echo "                   t.Errorf(\"error = %v, wantErr %v\", err, tt.wantErr)"
        echo "               }"
        echo "               if got != tt.want {"
        echo "                   t.Errorf(\"got %v, want %v\", got, tt.want)"
        echo "               }"
        echo "           })"
        echo "       }"
        echo "   }"
        echo ""
        echo "2. 基准测试示例:"
        echo "   func BenchmarkFunction(b *testing.B) {"
        echo "       for i := 0; i < b.N; i++ {"
        echo "           Function(input)"
        echo "       }"
        echo "   }"
        echo ""
        echo "3. 模糊测试示例:"
        echo "   func FuzzFunction(f *testing.F) {"
        echo "       f.Add(\"test input\")"
        echo "       f.Fuzz(func(t *testing.T, input string) {"
        echo "           Function(input) // 不应该 panic"
        echo "       })"
        echo "   }"
        echo ""
        echo "4. 并发测试示例:"
        echo "   func TestConcurrent(t *testing.T) {"
        echo "       var wg sync.WaitGroup"
        echo "       for i := 0; i < 100; i++ {"
        echo "           wg.Add(1)"
        echo "           go func() {"
        echo "               defer wg.Done()"
        echo "               Function() // 测试并发安全"
        echo "           }()"
        echo "       }"
        echo "       wg.Wait()"
        echo "   }"
        echo ""
    } >> "$REPORT_FILE"

    echo "✅ 改进建议已生成"
    echo ""
}

# 生成快速修复命令
generate_quick_fixes() {
    echo "生成快速修复命令..."
    echo ""

    {
        echo ""
        echo "🚀 快速修复命令"
        echo "==============="
        echo ""
        echo "# 1. 为所有 0% 覆盖率的包创建测试文件"
        echo ""
    } >> "$REPORT_FILE"

    go test -cover ./... 2>&1 | grep "coverage: 0.0%" | awk '{print $2}' | while read -r pkg; do
        # 转换包路径为文件路径
        pkg_path=$(echo "$pkg" | sed 's|github.com/vistone/fingerprint/||')

        # 找到包中的 Go 文件
        first_file=$(find "$pkg_path" -maxdepth 1 -name "*.go" -not -name "*_test.go" | head -1)

        if [ -n "$first_file" ]; then
            test_file="${first_file%.go}_test.go"
            {
                echo "# 为 $pkg 创建测试"
                echo "cat > $test_file << 'EOF'"
                echo "package $(basename "$pkg_path")"
                echo ""
                echo "import \"testing\""
                echo ""
                echo "func TestBasic(t *testing.T) {"
                echo "    // TODO: 添加测试用例"
                echo "    t.Skip(\"待实现\")"
                echo "}"
                echo "EOF"
                echo ""
            } >> "$REPORT_FILE"
        fi
    done

    {
        echo "# 2. 运行测试验证"
        echo "go test -v -race ./..."
        echo ""
        echo "# 3. 生成新的覆盖率报告"
        echo "bash scripts/coverage_analysis.sh"
        echo ""
    } >> "$REPORT_FILE"

    echo "✅ 快速修复命令已生成"
    echo ""
}

# 主函数
main() {
    run_coverage
    analyze_coverage
    generate_recommendations
    generate_quick_fixes

    echo "========================="
    echo "✅ 分析完成!"
    echo ""
    echo "📄 报告文件:"
    echo "   - $REPORT_FILE (文本报告)"
    echo "   - coverage.html (可视化报告)"
    echo ""
    echo "📖 详细优化方案:"
    echo "   - docs/OPTIMIZATION_ROADMAP.md"
    echo "   - OPTIMIZATION_SUMMARY.md"
    echo ""
    echo "🚀 下一步:"
    echo "   1. 查看报告: cat $REPORT_FILE"
    echo "   2. 打开可视化: open coverage.html"
    echo "   3. 按优先级修复无覆盖包"
}

# 运行
main
