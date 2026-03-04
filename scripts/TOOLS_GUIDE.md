# 优化工具使用指南

本目录包含用于项目优化的实用脚本和工具。

## 🚀 快速开始

### 一键启动

```bash
# 运行交互式优化工具
./scripts/quick_start.sh
```plaintext

这会启动一个交互式菜单，提供以下功能：

1. **测试覆盖率分析** - 分析所有包的测试覆盖率并生成改进建议
2. **Markdown 修复** - 自动修复 Markdown 格式问题
3. **安全检查** - 运行 gosec 安全扫描
4. **完整测试套件** - 运行所有测试（带竞态检测）
5. **性能基准测试** - 运行并对比性能基准
6. **项目状态概览** - 显示项目当前状态
7. **开始 Week 1 安全修复** - 创建分支并开始第一个安全修复任务
8. **全部执行** - 依次运行所有检查（推荐首次使用）

## 📋 独立脚本

### 1. 测试覆盖率分析

分析测试覆盖率并生成详细报告和改进建议。

```bash
./scripts/coverage_analysis.sh
```plaintext

**输出文件**:
- `coverage_analysis.txt` - 文本格式的详细分析报告
- `coverage.html` - 可视化覆盖率报告
- `coverage.out` - 原始覆盖率数据

**报告内容**:
- 各包覆盖率统计（带颜色标识）
- 统计摘要（总包数、无覆盖数、低于目标数、达标数）
- 按优先级分级的改进建议
- 测试编写指南和代码示例
- 快速修复命令

### 2. Markdown 修复

自动修复常见的 Markdown lint 错误。

```bash
./scripts/fix_markdown.sh
```plaintext

**修复内容**:
- ✅ 空代码块（添加 `plaintext` 标识）
- ✅ 表格格式（在 `|` 两边添加空格）
- ⚠️ 手动处理项（heading 空行、列表空行）

**输出文件**:
- `markdown_fix_report.txt` - 修复报告

### 3. 性能基准对比

```bash
# 运行新的基准测试
go test -bench=. -benchmem -run=^$ ./test/ > benchmark_new.txt

# 与基线对比（需要安装 benchstat）
go install golang.org/x/perf/cmd/benchstat@latest
benchstat test/benchmark_baseline.txt benchmark_new.txt
```plaintext

## 📊 使用场景

### 场景 1: 首次优化检查

```bash
# 运行完整检查
./scripts/quick_start.sh

# 选择选项 8 (全部执行)
# 这会生成所有报告，了解项目现状
```plaintext

### 场景 2: 提升测试覆盖率

```bash
# 1. 运行覆盖率分析
./scripts/coverage_analysis.sh

# 2. 查看报告，找出 0% 覆盖率的包
cat coverage_analysis.txt | grep "❌"

# 3. 为目标包创建测试文件
# 例如: internal/utils/
touch internal/utils/utils_test.go

# 4. 编写测试（参考报告中的示例）

# 5. 验证
go test -cover ./internal/utils/...

# 6. 重新运行分析查看进展
./scripts/coverage_analysis.sh
```plaintext

### 场景 3: 开始安全修复

```bash
# 1. 确保工作区干净
git status

# 2. 运行快速启动工具
./scripts/quick_start.sh

# 3. 选择选项 7 (开始 Week 1 安全修复)
# 这会创建一个新分支并显示详细的操作指南

# 4. 按照指南修复 JA3 输入验证问题
```plaintext

### 场景 4: 修复 Markdown 问题

```bash
# 1. 运行自动修复
./scripts/fix_markdown.sh

# 2. 查看报告
cat markdown_fix_report.txt

# 3. 手动处理需要注意的问题
# 参考报告中的指导

# 4. 提交更改
git add .
git commit -m "docs: fix markdown lint issues"
```plaintext

### 场景 5: 性能优化验证

```bash
# 1. 保存当前基准
go test -bench=. -benchmem -run=^$ ./test/ > benchmark_before.txt

# 2. 进行优化（例如: 减少内存分配）

# 3. 运行新的基准测试
go test -bench=. -benchmem -run=^$ ./test/ > benchmark_after.txt

# 4. 对比结果
benchstat benchmark_before.txt benchmark_after.txt

# 5. 如果性能提升，更新基线
cp benchmark_after.txt test/benchmark_baseline.txt
```plaintext

## 🔧 工具依赖

### 必需工具
- Go 1.24+
- git
- bash
- sed, awk, grep (标准 Unix 工具)

### 可选工具

安装以下工具以获得完整功能：

```bash
# gosec - 安全扫描
go install github.com/securego/gosec/v2/cmd/gosec@latest

# benchstat - 基准测试对比
go install golang.org/x/perf/cmd/benchstat@latest

# golangci-lint - 代码检查
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# markdownlint - Markdown 检查
npm install -g markdownlint-cli2
```plaintext

## 📝 报告文件说明

运行脚本后会生成以下报告文件：

| 文件名 | 说明 | 生成脚本 |
| -------- | ------ | --------- |
| `coverage_analysis.txt` | 测试覆盖率详细分析 | coverage_analysis.sh |
| `coverage.html` | 可视化覆盖率报告 | coverage_analysis.sh |
| `coverage.out` | 原始覆盖率数据 | coverage_analysis.sh |
| `markdown_fix_report.txt` | Markdown 修复报告 | fix_markdown.sh |
| `security_report.txt` | 安全扫描报告 | quick_start.sh (选项 3) |
| `test_results.txt` | 完整测试结果 | quick_start.sh (选项 4) |
| `benchmark_new.txt` | 最新基准测试结果 | quick_start.sh (选项 5) |
| `benchmark_comparison.txt` | 基准测试对比 | quick_start.sh (选项 5) |

## 🎯 优化目标

### Week 1-2 目标
- [x] 修复 2 个高危安全问题
- [ ] 提升测试覆盖率到 50%+
- [ ] 清零 Markdown lint 错误
- [ ] 所有测试通过竞态检测

### Week 3-8 目标
- [ ] 完成包结构重构 (Phase 1-3)
- [ ] 性能优化（减少 50% 内存分配）
- [ ] 并发安全加固

### Week 9-24 目标
- [ ] 插件化架构
- [ ] 完整可观测性
- [ ] 依赖优化

## 📚 相关文档

- [优化执行摘要](../OPTIMIZATION_SUMMARY.md) - 快速参考指南
- [完整优化方案](../docs/OPTIMIZATION_ROADMAP.md) - 详细的 6 个月优化计划
- [架构设计](../docs/ARCHITECTURE.md) - 项目架构说明
- [安全审计](../docs/SECURITY_AUDIT.md) - 安全问题详细报告
- [包重构计划](../docs/5-process/package-restructuring-plan.md) - 包结构重构指南

## 🆘 常见问题

### Q: 运行脚本时提示权限不足？

```bash
chmod +x scripts/*.sh
```plaintext

### Q: go test 运行很慢？

```bash
# 只运行特定包
go test -v ./internal/utils/...

# 使用缓存
go test -v ./...

# 并行运行
go test -v -parallel 4 ./...
```plaintext

### Q: 如何跳过长时间运行的测试？

```bash
# 使用 -short 标志
go test -short ./...

# 在测试中添加:
if testing.Short() {
    t.Skip("skipping in short mode")
}
```plaintext

### Q: 覆盖率报告中的数字不准确？

```bash
# 清除缓存
go clean -testcache

# 重新运行
go test -cover ./...
```plaintext

### Q: 如何只运行特定的测试？

```bash
# 运行特定测试函数
go test -run TestFunctionName ./...

# 运行特定基准测试
go test -bench BenchmarkFunctionName ./...
```plaintext

## 💡 最佳实践

1. **定期运行全面检查**: 每周至少运行一次 `quick_start.sh` 选项 8
2. **提交前运行测试**: 每次提交前运行 `go test -race ./...`
3. **关注覆盖率**: 每个 PR 都应该包含测试，且覆盖率不下降
4. **性能基准对比**: 性能优化后务必对比基准
5. **及时修复安全问题**: 高危安全问题应该立即修复

## 🔗 快速链接

- [立即开始](../OPTIMIZATION_SUMMARY.md#-本周行动计划-week-1)
- [测试指南](../docs/OPTIMIZATION_ROADMAP.md#4-测试策略)
- [性能优化](../docs/OPTIMIZATION_ROADMAP.md#22-性能优化---内存分配减少)
- [安全修复](../docs/OPTIMIZATION_ROADMAP.md#11-安全漏洞修复-high)

---

**最后更新**: 2026-03-04  
**维护者**: @vistone
