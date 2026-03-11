# 贡献指南

欢迎贡献代码！请参考以下指南。

## 必读文档

- **[开发指南](./DEVELOPER_GUIDE.md)** - 版本控制规则、编码规范、模块开发
- **[提交检查清单](../COMMIT_CHECKLIST.md)** - 5 分钟快速参考

## 开发流程

### 1. 克隆项目

```bash
git clone https://github.com/vistone/fingerprint.git
cd fingerprint && go work sync
git checkout -b feature/your-feature
```

### 2. 开发测试

```bash
go test ./modules/... -race
go fmt ./...
golangci-lint run
```

### 3. 提交信息

```
<type>(<scope>): <subject>

<body>
```

**Type:** `feat` / `fix` / `docs` / `refactor` / `test` / `chore`

**示例:** `feat(profiles): add Chrome 140 support`

### 4. 版本控制（强制）

详见 [开发指南 - 版本控制规则](./DEVELOPER_GUIDE.md#-版本控制规则强制执行)

**7 步流程：**
1. 代码修改 + 测试通过
2. 更新 CHANGELOG.md
3. 更新版本号（minor +1）
4. 提交版本更新
5. 创建版本 Tags
6. 推送到 GitHub
7. 验证规范

**必须遵守！** 违反规则导致提交被拒或回滚。

## 代码规范

### 错误处理

```go
if err != nil {
    return fmt.Errorf("failed to get profile %q: %w", id, err)
}

var ErrInvalidProfile = errors.New("invalid profile")
```

### 并发安全

```go
type Registry struct {
    mu       sync.RWMutex
    profiles map[string]Profile
}

func (r *Registry) Get(id string) (Profile, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    p, ok := r.profiles[id]
    return p, ok
}
```

### 日志

```go
import "log/slog"

slog.Debug("request", "profile", name)
```

## 测试

### 单元测试

```go
func TestGet(t *testing.T) {
    tests := []struct {
        name string
        id   string
        ok   bool
    }{
        {"exist", "chrome_140", true},
        {"not", "unknown", false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, ok := Get(tt.id)
            if ok != tt.ok {
                t.Errorf("want %v, got %v", tt.ok, ok)
            }
        })
    }
}
```

### 覆盖率

```bash
go test -coverprofile=coverage.out ./modules/...
go tool cover -html=coverage.out
```

## 模块开发

### 创建新模块

```bash
mkdir modules/newmod
cd modules/newmod

cat > go.mod << 'EOF'
module github.com/vistone/fingerprint/modules/newmod
go 1.25.7
require github.com/vistone/fingerprint/modules/core v1.0.11
replace github.com/vistone/fingerprint/modules/core => ../core
EOF

cd ../..
echo "./modules/newmod" >> go.work
go work sync
```

### 依赖规则

- 禁止循环依赖
- core 是零依赖模块
- 其他模块可依赖 core

## 检查清单

提交前：

- [ ] `go test ./modules/... -race` ✓
- [ ] `go fmt ./...` ✓
- [ ] `golangci-lint run` ✓
- [ ] CHANGELOG.md 已更新
- [ ] 版本号已更新（sed）
- [ ] 版本提交已创建
- [ ] Tags 已创建（主项目 + 17 模块）
- [ ] 版本审计通过

## 文档

- [开发指南](./DEVELOPER_GUIDE.md)
- [版本管理](./VERSION_MANAGEMENT.md)
- [API](./API.md)
- [架构](./ARCHITECTURE.md)
- [变更日志](./CHANGELOG.md)

## 感谢

感谢贡献！
