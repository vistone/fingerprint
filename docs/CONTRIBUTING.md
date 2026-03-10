# 贡献指南

感谢您对 Fingerprint 项目的兴趣和贡献！本文档指导您如何有效地贡献代码和改进。

## 📋 开始前的必读

在开始任何工作之前，**请务必阅读以下文档**：

1. **[版本控制规则](./DEVELOPER_GUIDE.md#-版本控制规则强制执行)** - ⚠️ **强制执行，不能乱来**
   - 每次提交必须遵守 7 步工作流
   - 违反规则导致提交被拒或回滚

2. **[提交检查清单](../COMMIT_CHECKLIST.md)** - 快速参考卡片
   - 5 分钟提交前检查
   - 包含常见错误修正方法

3. **[完整开发指南](./DEVELOPER_GUIDE.md)** - 全面的开发规范
   - 模块开发规范
   - 编码规范和测试规范
   - 提交信息格式

## 🔄 版本控制工作流 (强制)

### Step 1: Fork 和 Clone

```bash
git clone https://github.com/vistone/fingerprint.git
cd fingerprint
go work sync
```

### Step 2: 创建功能分支

```bash
# 功能分支命名: feature/xxxx 或 fix/xxxx
git checkout -b feature/your-feature-name
```

### Step 3: 开发和测试

```bash
# 确保所有测试通过
go test ./modules/... -race
go test -coverprofile=coverage.out ./modules/...

# 代码格式化
go fmt ./...

# Lint 检查
golangci-lint run
```

### Step 4: 提交代码

遵循提交信息格式：

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Type 列表:**
- `feat`: 新功能
- `fix`: 修复
- `docs`: 文档
- `refactor`: 重构
- `test`: 测试
- `chore`: 构建/工具

**示例:**

```
feat(profiles): add Chrome 140 support

Add Chrome 140 fingerprint configuration with updated TLS settings.

Closes #123
```

### Step 5: ⚠️ **强制：版本控制规则**

在推送到 GitHub 前，**必须**执行以下步骤：

#### 5a. 更新 CHANGELOG.md

```markdown
在 docs/CHANGELOG.md 顶部将 [Unreleased] 改为：

## [v1.0.X] - YYYY-MM-DD

### Added
- **Feature Name**: 功能描述

### Fixed
- **Bug Name**: Bug 修复说明

### Changed
- 改进说明
```

#### 5b. 更新所有版本号

```bash
# 例如从 v1.0.6 → v1.0.7
sed -i 's/v1\.0\.6/v1.0.7/g' go.mod modules/*/go.mod

# 验证
grep "v1.0.7" go.mod modules/*/go.mod | wc -l
# 应该显示 ~70 行
```

#### 5c. 提交版本更新

```bash
git add docs/CHANGELOG.md go.mod modules/*/go.mod
git commit -m "chore: Release v1.0.7"
```

#### 5d. 创建版本 Tags

```bash
# 主项目
git tag -a v1.0.7 -m "Release v1.0.7"

# 所有模块（19 个）
for module in agent client config core defense errors fingerprint frontend gateway generator http internal kit metrics ml network plugin profiles tls; do
    git tag -a modules/$module/v1.0.7 -m "Release modules/$module v1.0.7"
done
```

#### 5e. 推送到 GitHub

```bash
git push origin main
git push origin --tags
```

#### 5f. 验证规范

```bash
bash /tmp/version_audit.sh
# 应该显示：✅ 最新提交已有版本 tag
```

## 📝 代码规范

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
```

### 日志记录

```go
import "github.com/vistone/fingerprint/modules/internal/logger"

logger.Debug("processing request", 
    "profile", profileName,
    "browser", browserType,
)
```

## 🧪 测试规范

### 单元测试

```go
func TestGet(t *testing.T) {
    tests := []struct {
        name     string
        id       string
        wantOK   bool
    }{
        {"existing", "chrome_140", true},
        {"non-existing", "unknown", false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, ok := Get(tt.id)
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
        Get("chrome_140")
    }
}
```

### 测试覆盖率

```bash
go test -coverprofile=coverage.out ./modules/...
go tool cover -html=coverage.out
```

## 📦 模块开发

### 创建新模块

```bash
# 1. 创建目录
mkdir modules/newmodule
cd modules/newmodule

# 2. 初始化 go.mod
cat > go.mod << 'EOF'
module github.com/vistone/fingerprint/modules/newmodule

go 1.25.7

require github.com/vistone/fingerprint/modules/core v1.0.7

replace github.com/vistone/fingerprint/modules/core => ../core
EOF

# 3. 添加到 workspace
cd ../..
echo "./modules/newmodule" >> go.work
go work sync
```

### 模块依赖原则

```
禁止循环依赖：
A ──▶ B ──▶ A  ❌

允许的依赖方向：
core（零依赖）
  ▲
  ├── profiles ──▶ core
  ├── tls ───────▶ core
  └── 其他 ──────▶ core
```

## 🔍 提交前检查清单

在推送前检查：

- [ ] `go test ./modules/... -race` 全部通过
- [ ] `go fmt ./...` 代码格式化
- [ ] `golangci-lint run` lint 检查通过
- [ ] CHANGELOG.md 已更新（[v1.0.X] - YYYY-MM-DD）
- [ ] 版本号已更新（sed 批量更新）
- [ ] 版本提交已创建（git commit -m "chore: Release..."）
- [ ] Tags 已创建（v1.0.X + modules/*/v1.0.X）
- [ ] 版本审计通过（bash /tmp/version_audit.sh）

详见 [COMMIT_CHECKLIST.md](../COMMIT_CHECKLIST.md)

## 🚫 严禁的行为

| ❌ 违规行为 | 后果 |
|----------|-----|
| 提交代码但不更新版本号 | 拒绝 |
| 版本号不一致（某些 go.mod 旧版) | 强制回滚 |
| CHANGELOG 未更新 | 强制回滚 |
| 创建提交但无 tag | 不认为是发布 |
| 推送前未验证规范 | 拒绝 push |

## 📖 文档

- [开发者指南](./DEVELOPER_GUIDE.md) - 完整的开发和版本控制规范
- [版本管理](./VERSION_MANAGEMENT.md) - 版本号策略详解
- [API 文档](./API.md) - 完整 API 参考
- [架构说明](./ARCHITECTURE.md) - Go Workspace 架构

## 🆘 获得帮助

- 查看 [COMMIT_CHECKLIST.md](../COMMIT_CHECKLIST.md) - 快速参考卡片
- 查看 [DEVELOPER_GUIDE.md](./DEVELOPER_GUIDE.md) - 完整指南
- 查看 [CHANGELOG.md](./CHANGELOG.md) - 版本历史示例

## 📋 Pull Request 流程

1. Fork 项目并创建功能分支
2. 完成开发和测试
3. 按照【强制：版本控制规则】更新版本
4. 提交 PR，说明改动内容
5. 等待审查和合并

## 感谢

感谢您对 Fingerprint 项目的贡献！

---

**记住：严格遵守版本控制规则，不能乱来！** 🔒
