# 🚀 立即行动清单 (Week 1-2)

**时间**: 现在 - 2周（10个工作日）  
**目标**: 完成 Phase 1 - 紧急修复  
**状态**: 🔴 尚未开始

---

## 📋 Day 1-2: 环境准备 & 问题评估

### ✅ 清单项

- [ ] **10:00** - 阅读本行动计划 (15分钟)

```plaintext
你应该:
1. 理解Phase1的4个主要任务
2. 确认手上的资源
3. 准备开发环境
```plaintext

- [ ] **10:15** - 运行项目诊断 (10分钟)

```bash
cd /media/stone/data/fingerprint

# 1. 检查Go环境
go version
go env | grep -E "GOPATH|GOROOT|GOVERSION"

# 2. 验证依赖
go mod download
go mod verify

# 3. 运行现有测试 (检查baseline)
go test ./... -v -count=1 2>&1 | tee test_baseline.log

# 4. 检查代码检查工具
go vet ./...
golangci-lint run ./... 2>&1 | tee lint_baseline.log
```plaintext

- [ ] **10:30** - 查看现有工具 (5分钟)

```bash
# 查看可用脚本
ls -lh scripts/*.sh

# 列出当前的分析文档
ls -lh *.md | head -10
```plaintext

**预期输出时间**: 10:45

---

- [ ] **10:45** - 创建工作分支 (5分钟)

```bash
# 创建本周工作分支
git checkout -b phase1/week1-security-foundation

# 或如果是多人协作
git checkout -b phase1/week1-security-foundation-$(date +%Y%m%d)

# 检查分支状态
git status
git log --oneline -5
```plaintext

**预期输出时间**: 10:50

---

- [ ] **11:00** - 深度分析现有问题 (30分钟)

```bash
# 1. 分析JA3模块代码
go list -m all | head -20
find . -path ./vendor -prune -o -name "*.go" -print | xargs wc -l | tail -20

# 2. 重点检查这些文件
cat tls/ja3/parse.go | head -50   # 查看parse函数
cat cmd/profilegen/parser.go | head -50

# 3. 查看现有测试
ls -la tls/ja3/*_test.go
ls -la cmd/profilegen/*_test.go
```plaintext

**预期输出时间**: 11:30

---

### 💡 Day 1-2 的期望产出

```plaintext
完成后应该有:
✅ 项目环境正常工作
✅ 理解了2个高危漏洞的具体位置
✅ 创建了工作分支
✅ 建立了baseline (测试日志、lint报告)
✅ 编写了Phase1的工作计划
```plaintext

---

## 🔴 Day 3-4: 第一个漏洞修复 (JA3 输入验证)

### 📝 任务说明

**目标**: 修复 [tls/ja3/parse.go](tls/ja3/parse.go) 的输入验证不足问题

**问题**:
```go
// ❌ 当前：无防护
func Parse(ja3 string) (*JA3, error) {
    parts := strings.Split(ja3, ",")  // 无长度检查
    // ...
}
```plaintext

### ✅ 修复步骤

- [ ] **08:00** - 分析现有实现 (30分钟)

```bash
# 1. 查看完整的parse函数
cat tls/ja3/parse.go | grep -A 50 "^func Parse"

# 2. 查看相关的结构定义
grep -A 20 "type JA3 struct" tls/ja3/types.go

# 3. 查看现有测试
cat tls/ja3/*_test.go
```plaintext

**任务**: 
- 理解parse函数的完整流程
- 记下所有需要验证的字段
- 识别潜在的panic点

---

- [ ] **08:30** - 设计修复方案 (30分钟)

在编辑器中创建 `tls/ja3/validator.go`:

```go
// validator.go - 新建文件

package ja3

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// JA3 格式: TLSVersion,Ciphers,Extensions,EllipticCurves,EllipticCurvePointFormats
// 约束:
// - 长度: 1-4096字符
// - 格式: 仅数字、逗号、连字符
// - 部分个数: 恰好5部分

const (
	maxJA3Length = 4096
	maxPartCount = 5
)

var (
	// 允许的字符: 数字、逗号、连字符
	validJA3Pattern = regexp.MustCompile(`^[\d,-]+$`)
	
	// 错误定义
	ErrJA3TooLong   = fmt.Errorf("JA3 string exceeds max length of %d", maxJA3Length)
	ErrJA3Format    = fmt.Errorf("invalid JA3 format: only digits, commas and hyphens allowed")
	ErrJA3Parts     = fmt.Errorf("JA3 must have exactly 5 parts separated by commas")
	ErrJA3Invalid   = fmt.Errorf("invalid JA3 characters or values")
)

// ValidateJA3 执行JA3字符串的完整验证
func ValidateJA3(ja3 string) error {
	// 1. 长度检查
	if len(ja3) == 0 {
		return ErrJA3Parts
	}
	if len(ja3) > maxJA3Length {
		return ErrJA3TooLong
	}

	// 2. 格式检查
	if !validJA3Pattern.MatchString(ja3) {
		return ErrJA3Format
	}

	// 3. 部分检查
	parts := strings.Split(ja3, ",")
	if len(parts) != maxPartCount {
		return fmt.Errorf("%w: got %d", ErrJA3Parts, len(parts))
	}

	// 4. 逐部分验证
	for i, part := range parts {
		if len(part) == 0 && i != 1 && i != 2 && i != 3 && i != 4 {
			// 某些部分可以为空（例如Extensions为空），但需要有逗号
			continue
		}

		// 验证每个元素（数字列表用连字符分隔）
		if part != "" {
			elements := strings.Split(part, "-")
			for _, elem := range elements {
				if _, err := strconv.ParseUint(elem, 10, 64); err != nil {
					return fmt.Errorf("%w: invalid part %d value '%s'", ErrJA3Invalid, i, elem)
				}
			}
		}
	}

	return nil
}

// ParseWithValidation 修复后的解析函数
func ParseWithValidation(ja3 string) (*JA3, error) {
	// 先验证
	if err := ValidateJA3(ja3); err != nil {
		return nil, err
	}

	// 再解析 (已验证，可以简化)
	parts := strings.Split(ja3, ",")
	
	version, _ := strconv.ParseUint(parts[0], 10, 16)
	// ... 继续原有的解析逻辑
	
	return &JA3{
		Version: uint16(version),
		// ... 其他字段
	}, nil
}
```plaintext

**检查清单**:
- [ ] validator.go 创建完成
- [ ] 所有验证函数有文档注释
- [ ] 定义了清晰的error消息

---

- [ ] **09:00** - 更新parse函数 (30分钟)

编辑 [tls/ja3/parse.go](tls/ja3/parse.go)，定位到 `func Parse`:

```go
// 修改前（查找这段）
func Parse(ja3 string) (*JA3, error) {
    parts := strings.Split(ja3, ",")
    // ...

// 修改后（替换为）
func Parse(ja3 string) (*JA3, error) {
    // 添加验证
    if err := ValidateJA3(ja3); err != nil {
        return nil, fmt.Errorf("parse JA3 failed: %w", err)
    }
    
    parts := strings.Split(ja3, ",")
    // ... 保持原有逻辑
}
```plaintext

**检查清单**:
- [ ] 导入了新的validator.go中的函数
- [ ] 调用了ValidateJA3函数
- [ ] 错误信息清晰

---

- [ ] **09:30** - 编写单元测试 (45分钟)

创建 `tls/ja3/validator_test.go`:

```go
package ja3

import (
	"testing"
)

func TestValidateJA3(t *testing.T) {
	tests := []struct {
		name    string
		ja3     string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Valid JA3",
			ja3:     "771,49195-49199,0-23-65-10-11,23-24,0",
			wantErr: false,
		},
		{
			name:    "Empty string",
			ja3:     "",
			wantErr: true,
			errMsg:  "exactly 5 parts",
		},
		{
			name:    "Too long (>4096)",
			ja3:     "771," + strings.Repeat("1-", 2048) + "0",
			wantErr: true,
			errMsg:  "exceeds max length",
		},
		{
			name:    "Invalid characters",
			ja3:     "771;49195,0-23,23,0",  // 分号不允许
			wantErr: true,
			errMsg:  "invalid JA3 format",
		},
		{
			name:    "Wrong part count",
			ja3:     "771,49195,0,23",  // 只有4部分
			wantErr: true,
			errMsg:  "exactly 5 parts",
		},
		{
			name:    "Invalid number",
			ja3:     "abc,49195,0-23,23,0",  // 非数字
			wantErr: true,
			errMsg:  "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateJA3(tt.ja3)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJA3() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

// Fuzzing 测试（Go 1.18+）
func FuzzValidateJA3(f *testing.F) {
	f.Add("771,49195,0-23,23,0")
	f.Add("")
	f.Add(";;;;")
	
	f.Fuzz(func(t *testing.T, ja3 string) {
		// 不应该panic
		_ = ValidateJA3(ja3)
	})
}
```plaintext

**检查清单**:
- [ ] 创建了validator_test.go
- [ ] 包含5+个边界测试用例
- [ ] 包含Fuzzing测试

---

- [ ] **10:15** - 运行测试 (15分钟)

```bash
# 1. 运行新增测试
go test ./tls/ja3 -v

# 2. 查看测试覆盖率
go test ./tls/ja3 -cover

# 3. 运行Fuzzing (可选，耗时)
go test -fuzz=FuzzValidateJA3 -fuzztime=10s ./tls/ja3

# 4. 正向测试（确保现有功能不破损）
go test ./tls -v -count=1
```plaintext

**预期**: 所有测试通过 ✅

---

- [ ] **10:30** - 提交代码 (10分钟)

```bash
# 1. 查看变更
git diff tls/ja3/

# 2. 分段add (可选)
git add tls/ja3/validator.go tls/ja3/validator_test.go
git add tls/ja3/parse.go

# 3. 提交
git commit -m "fix(ja3): add input validation to prevent DoS

- Add ValidateJA3() with length, format, and content checks
- Prevent panic from malformed JA3 strings
- Add comprehensive test coverage (including fuzzing)
- Fixes: HIGH security issue #1

BREAKING: None
Tests: 12 new test cases, fuzzing enabled"

# 4. 本地验证（再跑一遍测试）
go test ./tls/ja3 -v
```plaintext

**检查清单**:
- [ ] Commit message清晰
- [ ] 引用了问题编号
- [ ] 所有测试通过

---

### 📊 Day 3-4 期望产出

```plaintext
完成后:
✅ JA3 parse 函数有完整的输入验证
✅ 编写了12+个单元测试 + Fuzzing
✅ 所有测试通过，无回归
✅ 提交了干净的git commit
✅ 覆盖率从 0% 提升至 70%+
```plaintext

**不到24小时完成第一个关键漏洞修复！** 🎉

---

## 🔴 Day 5-6: 第二个漏洞修复 (Profile 加载)

### 📝 任务说明

**目标**: 修复 [cmd/profilegen/parser.go](cmd/profilegen/parser.go) 的路径遍历和内存炸弹问题

### ✅ 修复步骤 (类似Day 3-4)

- [ ] **08:00** - 分析现有实现 (30分钟)

```bash
cat cmd/profilegen/parser.go | head -100
grep -r "LoadProfile\|ReadFile" cmd/profilegen/
```plaintext

- [ ] **08:30** - 添加安全加载机制 (1小时)

创建 [cmd/profilegen/security.go](cmd/profilegen/security.go):

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	allowedBasePath = "./profiles/specs"
	maxProfileSize  = 10 * 1024 * 1024 // 10MB
)

// ValidateProfilePath 验证配置文件路径
func ValidateProfilePath(path string) (string, error) {
	// 1. 规范化路径
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}

	// 2. 检查白名单
	allowedAbs, _ := filepath.Abs(allowedBasePath)
	if !strings.HasPrefix(absPath, allowedAbs) {
		return "", fmt.Errorf("path outside allowed directory: %s", absPath)
	}

	// 3. 文件大小限制
	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("cannot access file: %w", err)
	}
	if info.Size() > maxProfileSize {
		return "", fmt.Errorf("profile file too large: %d bytes", info.Size())
	}

	return absPath, nil
}
```plaintext

- [ ] **09:30** - 更新LoadProfile函数 (1小时)

修改parser.go的LoadProfile函数使用ValidateProfilePath

- [ ] **10:30** - 编写安全测试 (1小时)

创建 [cmd/profilegen/security_test.go](cmd/profilegen/security_test.go)

- [ ] **11:30** - 运行集成测试 (30分钟)

```bash
go test ./cmd/profilegen -v
```plaintext

---

### 📊 Day 5-6 期望产出

```plaintext
完成后:
✅ Profile 路径验证完整
✅ 文件大小限制实施
✅ 路径遍历测试用例 10+
✅ 集成测试通过
```plaintext

---

## 🟠 Day 7-8: 并发安全修复

### 📝 任务说明

**目标**: 修复全局变量的竞态条件

```bash
# 找出有问题的全局变量
grep -r "^var " internal/cache/ profiles/ | grep -v "=" | head -20
go test -race ./... 2>&1 | grep -A5 "DATA RACE"
```plaintext

### ✅ 修复步骤

- [ ] **08:00** - 识别竞态条件 (1小时)

```bash
# 运行race detector
go test -race ./internal/cache -v

# 记录所有输出的 DATA RACE 问题
```plaintext

- [ ] **09:00** - 添加互斥锁 (2小时)

每个有竞态的全局变量添加 `sync.RWMutex`:

```go
type CacheManager struct {
    mu    sync.RWMutex
    data  map[string]interface{}  // 受保护的数据
}
```plaintext

- [ ] **11:00** - 编写并发测试 (2小时)

```go
func TestConcurrentAccess(t *testing.T) {
    // 启动多个 goroutines 并发读写
    // 验证数据一致性
}
```plaintext

- [ ] **13:00** - 验证修复 (1小时)

```bash
go test -race ./... -count=5  # 循环运行5次确保无race
```plaintext

---

### 📊 Day 7-8 期望产出

```plaintext
完成后:
✅ 所有竞态条件修复
✅ -race 检查通过
✅ 并发单元测试 50+ cases
```plaintext

---

## 🟠 Day 9-10: 文档修复 & 汇总

### ✅ 修复步骤

- [ ] **08:00** - Markdown自动修复 (30分钟)

```bash
./scripts/fix_markdown.sh

# 检查修复效果
npx markdownlint-cli2 '**/*.md' --fix 2>&1 | head -50
```plaintext

- [ ] **08:30** - 手动审查关键文档 (2小时)

```bash
# 检查主要文档是否正确渲染
ls docs/ARCHITECTURE.md README.md
# 逐一检查标题、代码块、链接
```plaintext

- [ ] **10:30** - 整理交付物 (1小时)

```bash
# 1. 创建Week1总结文件
cat > WEEK1_COMPLETION.md << 'EOF'
# Week 1 完成报告

## ✅ 完成的任务

### 高危漏洞修复
- [x] JA3 输入验证加固
  - 12个单元测试用例
  - Fuzzing测试 (100K次)
  - 覆盖率: 70%

- [x] Profile 路径遍历修复
  - 10个集成测试用例
  - 大小限制验证
  - 覆盖率: 65%

### 并发安全
- [x] 竞态条件修复
  - -race 检查通过
  - 50个并发测试

### 文档
- [x] Markdown 修复
  - 342个错误修复
  - 0个遗留错误

## 📊 指标

- 测试覆盖率: 20% → 40%
- 高危问题: 2 → 0
- 中危问题: 4 → 2

## 📅 计划

下周 (Week 2):
- 继续提升 0% 覆盖模块到 50%
- security review
EOF

# 2. 生成报告
git log --oneline --since="$(date -d '1 week ago' '+%Y-%m-%d')" > WEEK1_COMMITS.txt

# 3. 统计
echo "## 代码变更统计" >> WEEK1_COMPLETION.md
git diff --stat HEAD~10..HEAD >> WEEK1_COMPLETION.md
```plaintext

- [ ] **11:30** - 代码审查 & PR准备 (1小时)

```bash
# 1. 最终检查
go test ./... -v -count=1
go vet ./...
golangci-lint run ./...

# 2. 生成diff
git diff main > phase1_security_fixes.patch

# 3. 准备PR
# - 标题: "feat: Phase 1 Security Fixes - Input Validation & Path Traversal"
# - 描述: 包含上面的WEEK1_COMPLETION.md内容
# - 关联Issue
# - 审核者: @tech-lead
```plaintext

### 📊 Day 9-10 期望产出

```plaintext
完成后:
✅ 所有Phase 1任务完成
✅ PR 已提交审核
✅ 周报告已生成
✅ 代码合并前检查通过

关键数字:
📊 新增代码: ~1,500行
📊 新增测试: ~80个单元/集成测试
📊 覆盖率: 15% → 40%
📊 问题修复: 2 高危 + 4 中危 → 修复 2 高危 + 2 中危
```plaintext

---

## 🎉 Week 1 完成检查清单

### 必须完成的 (Done/Not Done)

**代码修复**:
- [ ] JA3 输入验证完整
- [ ] Profile 路径遍历防护完整
- [ ] 并发竞态修复完整

**测试**:
- [ ] 新增 80+ 个单元/集成测试
- [ ] 所有新测试通过
- [ ] -race 检查无警告

**文档**:
- [ ] Markdown lint 错误 = 0
- [ ] Week1完成报告创建
- [ ] commit message规范

**代码审查**:
- [ ] 代码自检 (go vet, golangci-lint) 通过
- [ ] PR 已提交
- [ ] 至少1位reviewer审核

### 可选增强

- [ ] Security review 已完成
- [ ] 性能基准测试已运行
- [ ] 文档已更新

---

## 📞 遇到问题?

### 常见问题

**Q: 测试编译失败**
```bash
# A: 清理缓存，重新下载依赖
go clean -testcache
go mod download
go test ./...
```plaintext

**Q: git merge 冲突**
```bash
# A: 解决冲突，保留两边的修改
git status
# 手动编辑有冲突的文件
git add .
git commit --no-edit
```plaintext

**Q: 某些测试太慢**
```bash
# A: 跳过集成测试，仅运行单元测试
go test ./... -short
```plaintext

### 获取帮助

1. 查看 [COMPREHENSIVE_UPDATE_PLAN.md](COMPREHENSIVE_UPDATE_PLAN.md)
2. 查看 [docs/OPTIMIZATION_ROADMAP.md](docs/OPTIMIZATION_ROADMAP.md)
3. 查看项目的 GitHub Issues

---

## ✅ 周一 (Week 2 开始) 的汇报内容

```plaintext
Good morning! Week 1 完成情况:

✅ 完成:
  - JA3 输入验证修复 (12个测试)
  - Profile 路径遍历修复 (10个测试)
  - 并发竞态修复 (50个测试)
  - Markdown 文档修复 (0个错误)
  
📊 指标:
  - 新增代码: 1,500行
  - 新增测试: 80+个
  - 覆盖率: 15% → 40%
  - 高危问题: 2 → 0

🎯 Week 2 目标:
  - generator/ 模块测试覆盖率 → 60%
  - http/policy/ 模块测试覆盖率 → 60%
  - Security review 完成
  - 文档更新完成

PR: #XXX (正审核中)
```plaintext

---

**预计 Week 1 工作量**: 40-50小时 (5-6 个工作日)  
**预计截止**: 下周五  
**项目负责人**: [您的名字]

Happy coding! 🚀

