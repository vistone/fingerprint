# 开发者指南

本文档面向希望为 fingerprint 库贡献代码或深入理解其内部实现的开发者。

## 开发环境设置

### 前置要求

- Go 1.23+
- Make
- Git
- golangci-lint (可选，用于本地 lint)

### 克隆和构建

```bash
# 克隆仓库
git clone https://github.com/vistone/fingerprint.git
cd fingerprint

# 下载依赖
go mod download

# 运行测试
go test ./...

# 运行带 race detector 的测试
go test -race ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```plaintext

### Makefile 命令

```bash
make help          # 显示所有可用命令
make test          # 运行所有测试
make test-race     # 运行带 race detector 的测试
make coverage      # 生成覆盖率报告
make lint          # 运行 linter
make fmt           # 格式化代码
make generate      # 运行代码生成
make clean         # 清理构建产物
```plaintext

## 代码结构

### 目录约定

```plaintext
package/
├── package.go          # 公共 API
├── package_test.go     # 单元测试
├── internal.go         # 内部实现
├── internal_test.go    # 内部测试（测试私有函数）
└── example_test.go     # 示例代码（会出现在文档中）
```plaintext

### 命名约定

| 类型 | 命名规范 | 示例 |
| ------ | ---------- | ------ |
| 接口 | 后缀 `er` 或描述性名词 | `Reader`, `Generator` |
| 结构体 | 驼峰命名 | `ClientProfile`, `HTTP2Spec` |
| 私有函数 | 小写开头 | `parseExtensions` |
| 常量 | 驼峰或大写下划线 | `MaxHeaderSize`, `TLS_RSA_WITH_AES_128_GCM_SHA256` |
| 包名 | 全小写，简短 | `ja3`, `http2` |

## 编码规范

### 错误处理

```go
// 错误包装：添加上下文
if err != nil {
    return fmt.Errorf("failed to parse JA3 %q: %w", ja3String, err)
}

// 预定义错误
var (
    ErrInvalidProfile = errors.New("invalid client profile")
    ErrNotFound       = errors.New("profile not found")
)

// 错误检查
if errors.Is(err, ErrNotFound) {
    // 处理未找到
}
```plaintext

### 日志记录

```go
// 使用结构化日志
import "github.com/vistone/fingerprint/internal/logger"

logger.Debug("processing request", 
    "profile", profileName,
    "duration_ms", duration.Milliseconds(),
)

logger.Info("profile loaded",
    "name", profile.Name,
    "version", profile.Version,
)

logger.Error("failed to generate fingerprint",
    "error", err,
    "request_id", reqID,
)
```plaintext

### 配置使用

```go
// 从全局配置获取
cfg := config.Get()

// 使用配置值
if cfg.FeatureExtraction.Enabled {
    // 启用特征提取
}

// 配置克隆（线程安全）
localCfg := cfg.Clone()
```plaintext

## 添加新的 Profile

### 方法一：代码方式（传统）

```go
// profiles/custom_profiles.go
func MyCustomProfile() ClientProfile {
    return ClientProfile{
        ClientHelloSpec: tls.ClientHelloSpec{
            TLSVersion:       tls.VersionTLS13,
            CipherSuites:     []uint16{0x1301, 0x1302},
            CompressionMethods: []uint8{0},
            Extensions: []tls.TLSExtension{
                &tls.SNIExtension{},
                &tls.SupportedCurvesExtension{Curves: []tls.CurveID{0x001d}},
            },
        },
        HTTP2Settings: map[http2.SettingID]uint32{
            http2.SettingInitialWindowSize: 131072,
        },
        // ... 其他配置
    }
}
```plaintext

### 方法二：YAML 方式（推荐）

```yaml
# profiles/specs/my_custom.yaml
name: MyCustom
browser: CustomApp
os: Linux
version: "1.0"

client_hello:
  version: 0x0303
  cipher_suites:
    - TLS_AES_128_GCM_SHA256
    - TLS_AES_256_GCM_SHA384
  extensions:
    - type: server_name
    - type: supported_groups
      data:
        curves:
          - X25519
          - P-256

http2:
  settings:
    initial_window_size: 131072
    max_frame_size: 16384
  priority:
    exclusive: true
    stream_dep: 0
    weight: 256
```plaintext

然后运行代码生成：

```bash
go generate ./profiles/...
```plaintext

## 测试指南

### 单元测试

```go
func TestParseJA3(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    *JA3
        wantErr bool
    }{
        {
            name:  "valid",
            input: "769,47-53-5-10,0-10-11,23-24-25,0",
            want: &JA3{
                Version: 769,
                // ...
            },
        },
        {
            name:    "invalid",
            input:   "invalid",
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Parse(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if diff := cmp.Diff(tt.want, got); diff != "" {
                t.Errorf("Parse() mismatch (-want +got):\n%s", diff)
            }
        })
    }
}
```plaintext

### 基准测试

```go
func BenchmarkParseJA3(b *testing.B) {
    input := "769,47-53-5-10-49161-49162-49171-49172-50-56-19-4,0-10-11,23-24-25,0"
    
    b.ReportAllocs()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        _, err := Parse(input)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```plaintext

### 模糊测试

```go
func FuzzParseJA3(f *testing.F) {
    f.Add("769,47-53-5-10,0-10-11,23-24-25,0")
    
    f.Fuzz(func(t *testing.T, input string) {
        // 不应 panic
        _, _ = Parse(input)
    })
}
```plaintext

### 测试最佳实践

1. **表驱动测试**: 使用结构体切片定义测试用例
2. **子测试**: 使用 `t.Run()` 为每个用例创建子测试
3. **比较工具**: 使用 `github.com/google/go-cmp/cmp` 进行深度比较
4. **并发测试**: 使用 `-race` 标志检测竞态条件
5. **覆盖率**: 目标 > 80% 的关键路径

## 调试技巧

### 启用调试日志

```go
import "github.com/vistone/fingerprint/internal/logger"

func init() {
    logger.SetLevel(logger.DebugLevel)
}
```plaintext

### 分析 Profile

```go
// 打印 Profile 详情
profile := profiles.Chrome_120
spew.Dump(profile)

// 或者使用 JSON
jsonData, _ := json.MarshalIndent(profile, "", "  ")
fmt.Println(string(jsonData))
```plaintext

### 性能分析

```bash
# CPU 分析
go test -cpuprofile=cpu.prof -bench=. ./...
go tool pprof cpu.prof

# 内存分析
go test -memprofile=mem.prof -bench=. ./...
go tool pprof mem.prof

# 跟踪
go test -trace=trace.out ./...
go tool trace trace.out
```plaintext

## 性能优化指南

### 1. 减少内存分配

```go
// 不好：每次调用都分配
func Bad() []byte {
    return make([]byte, 1024)
}

// 好：使用 sync.Pool 复用
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 1024)
    },
}

func Good() []byte {
    buf := bufferPool.Get().([]byte)
    defer bufferPool.Put(buf)
    return buf[:n]
}
```plaintext

### 2. 避免字符串转换

```go
// 不好
for i := 0; i < len(data); i++ {
    s := string(data[i])  // 分配
    process(s)
}

// 好：批量处理
process(string(data))  // 一次分配
```plaintext

### 3. 预分配切片

```go
// 不好
ciphers := []uint16{}
for _, c := range input {
    ciphers = append(ciphers, c)  // 可能多次扩容
}

// 好
ciphers := make([]uint16, 0, len(input))
for _, c := range input {
    ciphers = append(ciphers, c)
}
```plaintext

## 发布流程

### 版本号规范

遵循语义化版本 (SemVer): `MAJOR.MINOR.PATCH`

- **MAJOR**: 不兼容的 API 变更
- **MINOR**: 向后兼容的功能添加
- **PATCH**: 向后兼容的问题修复

### 发布检查清单

```markdown
- [ ] 所有测试通过
- [ ] 基准测试无回归
- [ ] 覆盖率未下降
- [ ] CHANGELOG 已更新
- [ ] 版本号已更新
- [ ] Tag 已创建
- [ ] 文档已更新
```plaintext

### 创建 Release

```bash
# 1. 更新 CHANGELOG
# 2. 更新版本号（如有必要）
# 3. 提交更改
git add -A
git commit -m "chore: release v1.x.x"

# 4. 打标签
git tag -a v1.x.x -m "Release v1.x.x"

# 5. 推送
git push origin main --tags
```plaintext

## 故障排查

### 常见问题

#### 1. "undefined: tls.ClientHelloSpec"

确保使用了正确的 utls 版本：

```bash
go get github.com/bogdanfinn/utls@latest
```plaintext

#### 2. Profile 选择不正确

检查 User-Agent 解析逻辑：

```go
ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36..."
browser, version, os := useragent.Parse(ua)
fmt.Printf("Browser: %s, Version: %s, OS: %s\n", browser, version, os)
```plaintext

#### 3. 内存泄漏

使用 pprof 分析：

```go
import _ "net/http/pprof"

func init() {
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
}
```plaintext

访问 `http://localhost:6060/debug/pprof/heap`

## 贡献指南

参见 [CONTRIBUTING.md](../CONTRIBUTING.md)

### 提交信息规范

```plaintext
type(scope): subject

body

footer
```plaintext

类型：
- `feat`: 新功能
- `fix`: 修复
- `docs`: 文档
- `style`: 格式（不影响代码）
- `refactor`: 重构
- `test`: 测试
- `chore`: 构建/工具

示例：
```plaintext
feat(ja4): add JA4 fingerprint calculation

Implement JA4 fingerprint as per the new spec.
Supports both TLS 1.2 and TLS 1.3.

Closes #123
```plaintext
