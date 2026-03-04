#!/bin/bash
# Fingerprint 项目 - 优化快速启动脚本
# 用途: 一键开始优化工作

set -e

# 颜色定义
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 打印横幅
print_banner() {
    echo -e "${BLUE}"
    cat << 'EOF'
╔═══════════════════════════════════════════════╗
║   Fingerprint 项目优化 - 快速启动工具        ║
║   Version: 1.0.0                              ║
║   Date: 2026-03-04                            ║
╚═══════════════════════════════════════════════╝
EOF
    echo -e "${NC}"
}

# 打印菜单
print_menu() {
    echo -e "${CYAN}请选择操作:${NC}"
    echo ""
    echo "  ${GREEN}1)${NC} 运行测试覆盖率分析"
    echo "  ${GREEN}2)${NC} 修复 Markdown 格式问题"
    echo "  ${GREEN}3)${NC} 运行安全检查"
    echo "  ${GREEN}4)${NC} 运行完整测试套件"
    echo "  ${GREEN}5)${NC} 生成性能基准报告"
    echo "  ${GREEN}6)${NC} 查看项目状态概览"
    echo "  ${GREEN}7)${NC} 开始 Week 1 安全修复（创建分支）"
    echo "  ${GREEN}8)${NC} 全部执行（推荐）"
    echo "  ${RED}0)${NC} 退出"
    echo ""
    echo -n "请输入选项 [0-8]: "
}

# 1. 运行测试覆盖率分析
run_coverage_analysis() {
    echo -e "\n${CYAN}=== 运行测试覆盖率分析 ===${NC}\n"

    if [ ! -f "scripts/coverage_analysis.sh" ]; then
        echo -e "${RED}❌ 脚本不存在: scripts/coverage_analysis.sh${NC}"
        return 1
    fi

    bash scripts/coverage_analysis.sh

    echo -e "\n${GREEN}✅ 覆盖率分析完成${NC}"
    echo -e "${YELLOW}💡 查看报告: cat coverage_analysis.txt${NC}"
    echo -e "${YELLOW}💡 打开可视化: open coverage.html${NC}"
}

# 2. 修复 Markdown
fix_markdown() {
    echo -e "\n${CYAN}=== 修复 Markdown 格式 ===${NC}\n"

    if [ ! -f "scripts/fix_markdown.sh" ]; then
        echo -e "${RED}❌ 脚本不存在: scripts/fix_markdown.sh${NC}"
        return 1
    fi

    bash scripts/fix_markdown.sh

    echo -e "\n${GREEN}✅ Markdown 修复完成${NC}"
    echo -e "${YELLOW}💡 查看报告: cat markdown_fix_report.txt${NC}"
}

# 3. 安全检查
run_security_check() {
    echo -e "\n${CYAN}=== 运行安全检查 ===${NC}\n"

    echo "正在检查..."

    # 检查 gosec
    if ! command -v gosec &> /dev/null; then
        echo -e "${YELLOW}⚠️  gosec 未安装，正在安装...${NC}"
        go install github.com/securego/gosec/v2/cmd/gosec@latest
    fi

    # 运行 gosec
    echo "运行 gosec 安全扫描..."
    gosec -fmt=text -out=security_report.txt ./... || true

    echo -e "\n${GREEN}✅ 安全检查完成${NC}"
    echo -e "${YELLOW}💡 查看报告: cat security_report.txt${NC}"
}

# 4. 运行完整测试
run_full_tests() {
    echo -e "\n${CYAN}=== 运行完整测试套件 ===${NC}\n"

    echo "运行测试（带竞态检测）..."
    go test -v -race -timeout=5m ./... 2>&1 | tee test_results.txt

    echo -e "\n${GREEN}✅ 测试完成${NC}"
    echo -e "${YELLOW}💡 查看结果: cat test_results.txt${NC}"
}

# 5. 性能基准
run_benchmarks() {
    echo -e "\n${CYAN}=== 运行性能基准测试 ===${NC}\n"

    echo "运行基准测试..."
    go test -bench=. -benchmem -run=^$ -timeout=30m ./test/ 2>&1 | tee benchmark_new.txt

    echo -e "\n${GREEN}✅ 基准测试完成${NC}"
    echo -e "${YELLOW}💡 查看结果: cat benchmark_new.txt${NC}"

    # 如果存在旧的基准，进行对比
    if [ -f "test/benchmark_baseline.txt" ]; then
        echo -e "\n${CYAN}对比基准差异...${NC}"
        echo "（需要安装 benchstat: go install golang.org/x/perf/cmd/benchstat@latest）"

        if command -v benchstat &> /dev/null; then
            benchstat test/benchmark_baseline.txt benchmark_new.txt | tee benchmark_comparison.txt
            echo -e "${YELLOW}💡 查看对比: cat benchmark_comparison.txt${NC}"
        fi
    fi
}

# 6. 项目状态概览
show_project_status() {
    echo -e "\n${CYAN}=== 项目状态概览 ===${NC}\n"

    echo -e "${BLUE}📊 代码统计${NC}"
    echo "  Go 文件数: $(find . -name '*.go' -not -path './.git/*' -not -path './vendor/*' | wc -l)"
    echo "  代码行数: $(find . -name '*.go' -not -path './.git/*' -not -path './vendor/*' -exec wc -l {} + 2>/dev/null | tail -1 | awk '{print $1}')"
    echo "  测试文件数: $(find . -name '*_test.go' -not -path './.git/*' -not -path './vendor/*' | wc -l)"
    echo ""

    echo -e "${BLUE}📦 依赖统计${NC}"
    echo "  直接依赖: $(go list -m all | wc -l)"
    echo ""

    echo -e "${BLUE}🔍 已知问题${NC}"
    echo "  Markdown lint 错误: 342"
    echo "  高危安全问题: 2"
    echo "  中危安全问题: 4"
    echo ""

    echo -e "${BLUE}✅ 当前覆盖率${NC}"
    go test -cover ./... 2>&1 | grep "coverage:" | tail -5
    echo ""

    echo -e "${BLUE}📚 相关文档${NC}"
    echo "  - README.md                          (项目介绍)"
    echo "  - OPTIMIZATION_SUMMARY.md            (优化摘要)"
    echo "  - docs/OPTIMIZATION_ROADMAP.md       (详细方案)"
    echo "  - docs/ARCHITECTURE.md               (架构文档)"
    echo "  - docs/SECURITY_AUDIT.md             (安全审计)"
    echo ""
}

# 7. 开始 Week 1 安全修复
start_week1_security_fixes() {
    echo -e "\n${CYAN}=== 开始 Week 1 安全修复 ===${NC}\n"

    echo -e "${YELLOW}这将创建一个新的 git 分支并开始第一个安全修复任务${NC}"
    echo -n "继续? (y/N): "
    read -r confirm

    if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
        echo "取消操作"
        return
    fi

    # 检查 git 状态
    if [ -n "$(git status --porcelain)" ]; then
        echo -e "${RED}❌ 工作区有未提交的更改，请先提交或暂存${NC}"
        git status --short
        return 1
    fi

    # 创建分支
    branch_name="security/ja3-validation-$(date +%Y%m%d)"
    echo "创建分支: $branch_name"
    git checkout -b "$branch_name"

    echo -e "\n${GREEN}✅ 分支创建成功${NC}"
    echo ""
    echo -e "${CYAN}下一步操作:${NC}"
    echo "  1. 编辑文件: tls/ja3/parse.go"
    echo "     - 添加输入长度验证"
    echo "     - 添加格式验证"
    echo "     - 使用 ParseUint 替代 Atoi"
    echo ""
    echo "  2. 创建测试: tls/ja3/parse_test.go"
    echo "     - 添加单元测试"
    echo "     - 添加边界测试"
    echo ""
    echo "  3. 创建模糊测试: tls/ja3/fuzz_test.go"
    echo "     - 添加 FuzzJA3Parse 函数"
    echo ""
    echo "  4. 运行测试:"
    echo "     go test -v -race ./tls/ja3/..."
    echo "     go test -fuzz=FuzzJA3Parse -fuzztime=30s ./tls/ja3/..."
    echo ""
    echo "  5. 提交更改:"
    echo "     git add ."
    echo "     git commit -m 'fix(security): add JA3 input validation (HIGH-1)'"
    echo ""
    echo -e "${YELLOW}💡 详细说明: docs/OPTIMIZATION_ROADMAP.md (搜索 HIGH-1)${NC}"
}

# 8. 全部执行
run_all() {
    echo -e "\n${CYAN}=== 执行所有检查 ===${NC}\n"

    run_coverage_analysis
    echo -e "\n${CYAN}------------------------${NC}"

    fix_markdown
    echo -e "\n${CYAN}------------------------${NC}"

    run_security_check
    echo -e "\n${CYAN}------------------------${NC}"

    run_full_tests
    echo -e "\n${CYAN}------------------------${NC}"

    show_project_status

    echo -e "\n${GREEN}✅ 所有检查完成！${NC}"
    echo -e "\n${YELLOW}📊 生成的报告文件:${NC}"
    echo "  - coverage_analysis.txt"
    echo "  - coverage.html"
    echo "  - markdown_fix_report.txt"
    echo "  - security_report.txt"
    echo "  - test_results.txt"
    echo ""
    echo -e "${CYAN}📖 下一步: 查看 OPTIMIZATION_SUMMARY.md 开始优化工作${NC}"
}

# 主函数
main() {
    print_banner

    while true; do
        print_menu
        read -r choice

        case $choice in
            1)
                run_coverage_analysis
                ;;
            2)
                fix_markdown
                ;;
            3)
                run_security_check
                ;;
            4)
                run_full_tests
                ;;
            5)
                run_benchmarks
                ;;
            6)
                show_project_status
                ;;
            7)
                start_week1_security_fixes
                ;;
            8)
                run_all
                ;;
            0)
                echo -e "\n${GREEN}👋 再见！${NC}\n"
                exit 0
                ;;
            *)
                echo -e "\n${RED}❌ 无效选项，请重新选择${NC}\n"
                ;;
        esac

        echo ""
        echo -e "${YELLOW}按 Enter 继续...${NC}"
        read -r
        clear
        print_banner
    done
}

# 运行
main
