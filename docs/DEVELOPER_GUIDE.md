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

## 发布流程

### 版本号规则

遵循 [Semantic Versioning](https://semver.org/):
- `MAJOR`: 不兼容的 API 变更
- `MINOR`: 向后兼容的功能添加
- `PATCH`: 向后兼容的问题修复

### 发布检查清单

- [ ] 所有测试通过
- [ ] 代码审查完成
- [ ] 文档已更新
- [ ] CHANGELOG.md 已更新
- [ ] Tag 已创建

```bash
# 创建发布标签
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0
```

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
