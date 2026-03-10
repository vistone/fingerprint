# 开发者指南

本文档面向希望为 fingerprint 库贡献代码或深入理解其内部实现的开发者。

## 开发环境设置

### 前置要求

- Go 1.25+
- Make
- Git
- golangci-lint (可选，用于本地 lint)

### 克隆和构建

```bash
# 克隆仓库
git clone https://github.com/vistone/fingerprint.git
cd fingerprint

# 同步 workspace
go work sync

# 构建特定模块
go build ./modules/core
go build ./modules/profiles
go build ./modules/fingerprint

# 运行测试
go test ./modules/...

# 运行特定模块测试
go test ./modules/profiles/... -v

# 运行带 race detector 的测试
go test -race ./modules/core/...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./modules/...
go tool cover -html=coverage.out -o coverage.html
```

### Makefile 命令

```bash
make help          # 显示所有可用命令
make test          # 运行所有测试
make test-race     # 运行带 race detector 的测试
make coverage      # 生成覆盖率报告
make lint          # 运行 linter
make fmt           # 格式化代码
make clean         # 清理构建产物
```

## Go Workspace 结构

```plaintext
github.com/vistone/fingerprint/
├── go.work                     # Workspace 定义
├── modules/                    # 所有模块
│   ├── core/                   # 核心类型（零依赖）
│   ├── profiles/               # 指纹配置
│   ├── tls/                    # TLS 分析
│   ├── http/                   # HTTP 分析
│   ├── ml/                     # ML 分类器
│   ├── defense/                # 安全防护
│   ├── frontend/               # 前端 SDK
│   ├── gateway/                # API 网关
│   ├── generator/              # 指纹生成
│   ├── network/                # 网络层
│   ├── internal/               # 内部工具
│   ├── config/                 # 配置管理
│   ├── plugin/                 # 插件系统
│   └── fingerprint/            # Facade 模块
├── cmd/                        # 应用程序入口
├── examples/                   # 示例代码
└── test/                       # 集成测试
```

## 模块开发规范

### 目录约定

每个模块遵循以下结构：

```plaintext
modules/<module>/
├── go.mod                  # 模块定义
├── *.go                    # 公共 API
├── *_test.go               # 单元测试
└── legacy/                 # 遗留兼容代码（可选）
    └── *.go
```

### 模块依赖原则

```plaintext
# 允许的依赖方向
core (零依赖)
    ▲
    ├── profiles ──▶ core
    ├── tls ───────▶ core
    ├── http ──────▶ core
    ├── ml ────────▶ core
    └── ...

# 禁止循环依赖
A ──▶ B ──▶ A  ❌
```

### 创建新模块

```bash
# 1. 创建目录
mkdir modules/mynewmodule
cd modules/mynewmodule

# 2. 初始化 go.mod
cat > go.mod << 'EOF'
module github.com/vistone/fingerprint/modules/mynewmodule

go 1.25.7

require github.com/vistone/fingerprint/modules/core v0.0.0

replace github.com/vistone/fingerprint/modules/core => ../core
EOF

# 3. 创建 main.go
cat > mynewmodule.go << 'EOF'
package mynewmodule

import "github.com/vistone/fingerprint/modules/core"

// 公共 API
EOF

# 4. 添加到 workspace
cd ../..
echo "./modules/mynewmodule" >> go.work

# 5. 同步
go work sync
```

## 编码规范

### 错误处理

```go
// 错误包装：添加上下文
if err != nil {
    return fmt.Errorf("failed to get profile %q: %w", id, err)
}

// 预定义错误
var (
    ErrInvalidProfile = errors.New("invalid client profile")
    ErrNotFound       = errors.New("profile not found")
)

// 错误检查
if errors.Is(err, profiles.ErrNotFound) {
    // 处理未找到
}
```

### 日志记录

```go
// 使用标准库日志或内部日志
import "github.com/vistone/fingerprint/modules/internal/logger"

logger.Debug("processing request", 
    "profile", profileName,
    "browser", browserType,
)
```

### 并发安全

```go
// 使用 sync.RWMutex 保护共享状态
type Registry struct {
    mu       sync.RWMutex
    profiles map[string]ClientProfile
}

func (r *Registry) Get(id string) (ClientProfile, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    p, ok := r.profiles[id]
    return p, ok
}

func (r *Registry) Register(p ClientProfile) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.profiles[p.ID] = p
}
```

## 测试规范

### 单元测试

```go
func TestGet(t *testing.T) {
    tests := []struct {
        name     string
        id       string
        wantOK   bool
    }{
        {"existing", "chrome_133", true},
        {"non-existing", "unknown", false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, ok := Get(tt.id)
            if ok != tt.wantOK {
                t.Errorf("Get() ok = %v, want %v", ok, tt.wantOK)
            }
        })
    }
}
```

### 基准测试

```go
func BenchmarkGet(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Get("chrome_133")
    }
}
```

### 集成测试

```go
// test/integration_test.go
package test

import (
    "testing"
    "github.com/vistone/fingerprint/modules/profiles"
    "github.com/vistone/fingerprint/modules/tls"
)

func TestIntegration(t *testing.T) {
    profile, ok := profiles.Get("chrome_133")
    if !ok {
        t.Fatal("profile not found")
    }
    
    // 测试 TLS 指纹计算
    ja3 := tls.CalculateJA3(clientHello)
    if ja3 == "" {
        t.Error("failed to calculate JA3")
    }
}
```

## 提交规范

### Commit Message 格式

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Type:**
- `feat`: 新功能
- `fix`: 修复
- `docs`: 文档
- `refactor`: 重构
- `test`: 测试
- `chore`: 构建/工具

**Scope:**
- `core`, `profiles`, `tls`, `http`, `ml`, `defense`, `gateway`, etc.

**示例:**
```
feat(profiles): add Chrome 140 support

Add Chrome 140 fingerprint with updated TLS settings.

Closes #123
```

## ⚠️ 版本控制规则（强制执行）

**本部分规则是强制性的，任何开发者都必须遵守。违反规则将导致提交被拒绝或回滚。**

### 第一法则：不能乱来

本项目采用严格的语义化版本控制，每一次代码提交都必须有清晰的版本记录和追溯。这是保证项目稳定性和可维护性的基础。

> "每次修改完后在写好 CHANGELOG 后，提交到 GitHub 上，最小版本号增加 1"

### 版本号规范

采用 **Semantic Versioning**: `MAJOR.MINOR.PATCH` (e.g., v1.0.5)

| 组件 | 含义 | 变化规则 |
|------|------|---------|
| MAJOR | 主版本号 | 重大架构变化、不兼容更新（极少改变） |
| MINOR | **次版本号** | **每次有效提交必须 +1** |
| PATCH | 修补版本号 | 仅用于紧急热修复 |

### 工作流（必须按顺序执行）

#### Step 1: 完成代码修改

```bash
# 开发完功能，确保所有测试通过
go test ./modules/... -race
go build ./modules/...

# 提交代码
git add .
git commit -m "feat(module): description"
```

#### Step 2: 更新 CHANGELOG.md ✅ 必须

在 `docs/CHANGELOG.md` 顶部将 `[Unreleased]` 改为具体版本号和日期：

```markdown
## [v1.0.6] - 2026-03-10    ← 替换 [Unreleased] 为这样的格式

### Added
- **Feature 1**: 功能描述
- **Feature 2**: 功能描述

### Fixed
- **Bug 1**: Bug 修复说明

### Changed
- 改进说明
```

**规则**：
- 版本号必须与 go.mod 中的版本一致
- 日期必须是当前日期（YYYY-MM-DD 格式）
- 至少列出一个 Added/Fixed/Changed 项

#### Step 3: 更新所有版本号 ✅ 必须

更新项目中所有 `go.mod` 文件的版本号：

```bash
# 当前版本: v1.0.5
# 新版本: v1.0.6

# 批量更新（从 v1.0.5 → v1.0.6）
sed -i 's/v1\.0\.5/v1.0.6/g' go.mod modules/*/go.mod

# 验证更新
grep "v1.0.6" go.mod modules/*/go.mod | wc -l
# 应该显示 ~20 行（1 主 + 19 模块）
```

#### Step 4: 提交版本更新 ✅ 必须

```bash
git add docs/CHANGELOG.md go.mod modules/*/go.mod

# 提交信息固定格式
git commit -m "chore: Release v1.0.6"

# 查看提交 hash（后续需要用）
git log -1 --oneline
# 示例: f0b1d12 (HEAD -> main) chore: Release v1.0.6
```

#### Step 5: 创建版本 Tags ✅ 必须

```bash
# 主项目 tag
git tag -a v1.0.6 -m "Release v1.0.6"

# 所有受影响的模块 tag（通常是全部）
for module in agent client core defense frontend gateway ml profiles; do
    git tag -a modules/$module/v1.0.6 -m "Release modules/$module v1.0.6"
done

# 验证
git tag -l | grep v1.0.6
```

#### Step 6: 推送到 GitHub ✅ 必须

```bash
# 推送主分支
git push origin main

# 推送所有新 tags
git push origin --tags

# 验证推送成功
git push origin --tags --force-with-lease
```

#### Step 7: 验证规范 ✅ 必须

```bash
# 运行规范检查脚本
bash /tmp/version_audit.sh

# 预期输出:
# ✅ 最新提交已有版本 tag
# 检查 CHANGELOG 版本声明: ## [v1.0.6] - YYYY-MM-DD
# v1.0.6 → [commit hash]
```

### 强制性检查清单

**在执行任何代码修改前，请确保理解这个清单：**

```
修改代码
  ↓
✅ 所有单元测试通过 (go test ./modules/... -race)
✅ 格式化代码 (go fmt ./...)
✅ Lint 通过 (golangci-lint run)
  ↓
✅ 更新 CHANGELOG.md (将 [Unreleased] → [v1.0.X] - YYYY-MM-DD)
✅ 更新所有 go.mod 版本 (Minor +1)
✅ 提交版本更新 (git commit -m "chore: Release v1.0.X")
  ↓
✅ 创建主项目 tag (git tag -a v1.0.X ...)
✅ 创建模块 tags (9 个 modules/*/v1.0.X)
✅ 推送到 GitHub (git push origin main && git push origin --tags)
  ↓
✅ 验证 GitHub 同步成功
✅ 运行版本审计脚本确认合规
```

### 严禁的行为 ❌

以下行为严格禁止，发现将导致提交被拒：

| 违规行为 | 后果 |
|---------|------|
| ❌ 提交代码但不更新版本号 | 提交拒绝，必须修正后重新提交 |
| ❌ 版本号不一致（某些 go.mod 是旧版） | 强制 revert，要求重新执行步骤 3 |
| ❌ 提交了新版本但 CHANGELOG 未更新 | 强制 revert，要求更新 CHANGELOG |
| ❌ 创建了提交但没有标记 tag | 必须补标 tag，否则不认为是有效发布 |
| ❌ 推送到 GitHub 前未验证本地规范 | 拒绝 push，要求通过审计脚本 |
| ❌ 修改版本号但不提交（只在 go.mod 中） | 明确 error，需要 git commit 与 tag |

### 当前版本状态

```
当前版本: v1.0.5 ✅
最新提交: f0b1d12 (chore: Release v1.0.5)
CHANGELOG: ✅ 已更新为 [v1.0.5] - 2026-03-10
所有模块: ✅ 同步至 v1.0.5
GitHub: ✅ 所有changes已推送
```

下次发布时版本应为 **v1.0.6**（Minor +1）

## 发布流程详解

本部分详细说明标准发布流程。**请严格按照【版本控制规则】章节的 Step 1-7 执行。**

### 快速参考脚本

```bash
#!/bin/bash
# 完整发布流程（~10 分钟）
# 前提：所有代码改动已完成并通过测试

set -e  # 任何错误立即停止

echo "📋 Step 1: 验证本地状态..."
git status  # 应该显示 clean working tree

echo "📝 Step 2: 更新 CHANGELOG.md"
echo "   手动编辑：将 [Unreleased] 改为 [v1.0.6] - $(date +%Y-%m-%d)"

echo "📦 Step 3: 批量更新版本号..."
sed -i 's/v1\.0\.5/v1.0.6/g' go.mod modules/*/go.mod

echo "✅ Step 4: 提交版本更新..."
git add docs/CHANGELOG.md go.mod modules/*/go.mod
git commit -m "chore: Release v1.0.6"

echo "🏷️  Step 5: 创建主项目 tag..."
git tag -a v1.0.6 -m "Release v1.0.6"

echo "🏷️  Step 6: 创建模块 tags..."
for module in agent client config core defense errors fingerprint frontend gateway generator http internal kit metrics ml network plugin profiles tls; do
    [ -f "modules/$module/go.mod" ] && \
    git tag -a modules/$module/v1.0.6 -m "Release modules/$module v1.0.6" || true
done

echo "🚀 Step 7: 推送到 GitHub..."
git push origin main
git push origin --tags

echo "✨ 发布成功！"
echo "   versiion: v1.0.6"
echo "   commit: $(git log -1 --oneline)"
echo "   tags: $(git tag -l | grep v1.0.6 | wc -l) 个"
```

### 手动执行步骤

**Step 1: 编辑 CHANGELOG.md**

在 `docs/CHANGELOG.md` 中找到顶部的 `## [Unreleased]` 并替换为：

```markdown
## [v1.0.6] - 2026-03-10

### Added
- **Feature Name**: 功能描述
  - 附加说明
  
### Fixed
- **Bug Name**: Bug 修复说明

### Changed
- 改进说明
```

**Step 2: 更新版本号**

```bash
# 从 v1.0.5 → v1.0.6
sed -i 's/v1\.0\.5/v1.0.6/g' go.mod modules/*/go.mod

# 验证所有版本都更新了
grep -r "v1.0.6" go.mod modules/*/go.mod | wc -l
# 应该显示约 20 行
```

**Step 3: 提交**

```bash
git add .
git commit -m "chore: Release v1.0.6"
```

**Step 4: 创建 Tags**

```bash
# 主项目
git tag -a v1.0.6 -m "Release v1.0.6"

# 所有模块（可用循环）
for module in agent client config core defense errors fingerprint frontend gateway generator http internal kit metrics ml network plugin profiles tls; do
    git tag -a modules/$module/v1.0.6 -m "Release modules/$module v1.0.6"
done

# 验证
git tag -l | grep v1.0.6 | wc -l  # 应该显示 19
```

**Step 5: 推送**

```bash
git push origin main
git push origin --tags
```

### 模块完整列表

项目包含以下 18 个模块，每个 tag 创建时都要对应创建：

| 模块 | go.mod 路径 | Tag 格式 |
|------|-----------|---------|
| agent | modules/agent/go.mod | modules/agent/v1.0.6 |
| client | modules/client/go.mod | modules/client/v1.0.6 |
| config | modules/config/go.mod | modules/config/v1.0.6 |
| core | modules/core/go.mod | modules/core/v1.0.6 |
| defense | modules/defense/go.mod | modules/defense/v1.0.6 |
| errors | modules/errors/go.mod | modules/errors/v1.0.6 |
| fingerprint | modules/fingerprint/go.mod | modules/fingerprint/v1.0.6 |
| frontend | modules/frontend/go.mod | modules/frontend/v1.0.6 |
| gateway | modules/gateway/go.mod | modules/gateway/v1.0.6 |
| generator | modules/generator/go.mod | modules/generator/v1.0.6 |
| http | modules/http/go.mod | modules/http/v1.0.6 |
| internal | modules/internal/go.mod | modules/internal/v1.0.6 |
| kit | modules/kit/go.mod | modules/kit/v1.0.6 |
| metrics | modules/metrics/go.mod | modules/metrics/v1.0.6 |
| ml | modules/ml/go.mod | modules/ml/v1.0.6 |
| network | modules/network/go.mod | modules/network/v1.0.6 |
| plugin | modules/plugin/go.mod | modules/plugin/v1.0.6 |
| profiles | modules/profiles/go.mod | modules/profiles/v1.0.6 |
| tls | modules/tls/go.mod | modules/tls/v1.0.6 |

### 故障排查

**问题 1: Tag 创建失败**
```bash
# 删除本地 tag
git tag -d v1.0.6 modules/*/v1.0.6

# 修正后重新创建
git tag -a v1.0.6 -m "Release v1.0.6"
```

**问题 2: 版本号不一致**
```bash
# 检查是否所有版本都一致
grep -h "github.com/vistone/fingerprint/modules" go.mod modules/*/go.mod | grep -o "v[0-9.]*" | sort | uniq

# 如果有多个版本，重新运行 sed 更新
```

**问题 3: Push 失败**
```bash
# 拉取最新主分支
git pull origin main --rebase

# 重试推送
git push origin main
git push origin --tags
```

### 发布后验证

推送成功后，在 GitHub 检查：

1. **最新提交**: https://github.com/vistone/fingerprint/commits/main
   - 应接近顶部显示 "chore: Release v1.0.6"

2. **Tags 列表**: https://github.com/vistone/fingerprint/tags
   - 应显示 v1.0.6 及所有 modules/*/v1.0.6 tags

3. **CHANGELOG**: https://github.com/vistone/fingerprint/blob/main/docs/CHANGELOG.md
   - 顶部应显示 `## [v1.0.6] - 2026-03-XX`

## 调试技巧

### 查看模块依赖

```bash
# 查看 workspace 中的所有模块
go work edit -json

# 查看特定模块的依赖
cd modules/profiles && go list -m all

# 可视化依赖图
cd modules/profiles && go mod graph | head -20
```

### 排查导入错误

```bash
# 检查导入路径是否正确
go build ./modules/... 2>&1 | grep "cannot find module"

# 确保 workspace 同步
go work sync

# 清理模块缓存
go clean -modcache
go work sync
```

### 性能分析

```bash
# CPU 分析
go test -cpuprofile=cpu.prof -bench=. ./modules/profiles/...
go tool pprof cpu.prof

# 内存分析
go test -memprofile=mem.prof -bench=. ./modules/profiles/...
go tool pprof mem.prof
```

## 常见问题

### Q: 如何处理循环依赖？

A: 将共享类型移动到 `core` 模块，或使用接口解耦。

### Q: 如何添加新的浏览器指纹？

A: 在 `modules/profiles/` 中添加新的 `.go` 文件，参考 `chrome.go` 的模式。

### Q: 如何测试模块间的集成？

A: 使用 `test/` 目录进行集成测试，或创建 `example_test.go`。

## 参考资源

- [Go Modules Reference](https://go.dev/ref/mod)
- [Go Workspace Tutorial](https://go.dev/doc/tutorial/workspaces)
- [Project Architecture](./ARCHITECTURE.md)
- [API Documentation](./API.md)
