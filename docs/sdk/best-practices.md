# fingerprint SDK 最佳实践指南

## 1. 扩展设计原则

### 单一职责原则（Single Responsibility Principle）

每个组件应该只负责一个功能：

```go
// ✅ 好的做法：分离关注点
type ECHParser struct{}         // 只负责解析
type ECHAnalyzer struct{}       // 只负责分析
type ECHHandler struct{}        // 只负责事件处理

// ❌ 不好的做法：一个组件做所有事情
type ECHExtensionProcessor struct {
    Parse()
    Analyze()
    Handle()
    // ...
}
```plaintext

### 开闭原则（Open/Closed Principle）

对扩展开放，对修改关闭：

```go
// ✅ 好的做法：通过扩展添加新行为
type CustomAnalyzer struct{}
func (a *CustomAnalyzer) Analyze(...) { /* 自定义逻辑 */ }
RegisterAnalyzerBuilder(MyType, func() (Parser, error) {
    return NewCustomAnalyzer(), nil
})

// ❌ 不好的做法：修改现有代码
// 在 ECHAnalyzerImpl.Analyze() 中添加新逻辑
```plaintext

### 依赖倒置原则（Dependency Inversion Principle）

依赖抽象接口，而不是具体实现：

```go
// ✅ 好的做法：依赖接口
func Process(parser extension.Parser, data []byte) {
    result, _ := parser.Parse(data, ctx)
    // ...
}

// ❌ 不好的做法：依赖具体类型
func Process(parser *ECHParser, data []byte) {
    // ...
}
```plaintext

## 2. 代码组织

### 目录结构

```plaintext
your-extension/
├── go.mod
├── go.sum
├── internal/
│   ├── parser.go        # 解析器实现
│   ├── analyzer.go      # 分析器实现
│   ├── handler.go       # 处理器实现
│   └── types.go         # 数据类型定义
├── examples/
│   └── main.go          # 使用示例
├── tests/
│   ├── parser_test.go
│   ├── analyzer_test.go
│   └── integration_test.go
├── docs/
│   ├── API.md
│   └── USAGE.md
└── readme.md
```plaintext

### 模块初始化

```go
package myext

var (
    // 定义扩展类型
    ExtensionType extension.ExtensionType = 0x1234
    
    // 版本信息
    Version = "1.0.0"
    Name    = "My Extension"
)

// Init 初始化所有组件
func Init() error {
    if err := RegisterMetadata(); err != nil {
        return fmt.Errorf("failed to register metadata: %w", err)
    }
    if err := RegisterParser(); err != nil {
        return fmt.Errorf("failed to register parser: %w", err)
    }
    if err := RegisterAnalyzer(); err != nil {
        return fmt.Errorf("failed to register analyzer: %w", err)
    }
    return nil
}
```plaintext

## 3. 错误处理

### 定义明确的错误类型

```go
package myext

var (
    ErrInvalidData      = errors.New("invalid extension data")
    ErrParseFailure     = errors.New("failed to parse extension")
    ErrAnalysisFailed   = errors.New("analysis failed")
    ErrInvalidConfig    = errors.New("invalid configuration")
)

// 包装上游错误
func Parse(data []byte) error {
    if len(data) == 0 {
        return ErrInvalidData
    }
    
    if err := someOperation(data); err != nil {
        return fmt.Errorf("%w: %v", ErrParseFailure, err)
    }
    
    return nil
}
```plaintext

### 优雅的错误恢复

```go
type MyAnalyzer struct{}

func (a *MyAnalyzer) Analyze(data extension.ExtensionData,
    config map[string]interface{}) (extension.AnalysisResult, error) {
    
    result := &MyAnalysisResult{
        ExtType:      a.GetType(),
        Anomalies:    []string{},
    }
    
    // 防御式编程
    defer func() {
        if r := recover(); r != nil {
            result.Anomalies = append(result.Anomalies, "analysis_panic")
            result.RiskScore = 1.0 // 最大风险
        }
    }()
    
    // 分析逻辑
    a.performAnalysis(data, result)
    
    return result, nil
}
```plaintext

## 4. 性能最佳实践

### 4.1 缓存策略

```go
type CachedAnalyzer struct {
    analyzer extension.Analyzer
    cache    map[string]extension.AnalysisResult
    mu       sync.RWMutex
}

func (ca *CachedAnalyzer) Analyze(data extension.ExtensionData,
    config map[string]interface{}) (extension.AnalysisResult, error) {
    
    key := ca.getCacheKey(data)
    
    // 检查缓存
    ca.mu.RLock()
    if result, ok := ca.cache[key]; ok {
        ca.mu.RUnlock()
        return result, nil
    }
    ca.mu.RUnlock()
    
    // 执行分析
    result, err := ca.analyzer.Analyze(data, config)
    if err != nil {
        return nil, err
    }
    
    // 更新缓存
    ca.mu.Lock()
    ca.cache[key] = result
    ca.mu.Unlock()
    
    return result, nil
}
```plaintext

### 4.2 避免内存分配

```go
// ❌ 不好：每次都分配新内存
func (p *MyParser) Parse(data []byte, ctx context.Context) (
    extension.ExtensionData, error) {
    result := make([]byte, len(data)) // 不必要的分配
    copy(result, data)
    return &MyData{Data: result}, nil
}

// ✅ 好：直接引用
func (p *MyParser) Parse(data []byte, ctx context.Context) (
    extension.ExtensionData, error) {
    return &MyData{Data: data}, nil // 直接引用，无额外分配
}
```plaintext

### 4.3 批处理

```go
type BatchAnalyzer struct {
    analyzer extension.Analyzer
    batch    []extension.ExtensionData
    batchSize int
}

func (ba *BatchAnalyzer) AddAndAnalyze(data extension.ExtensionData) (
    []extension.AnalysisResult, error) {
    
    ba.batch = append(ba.batch, data)
    
    if len(ba.batch) >= ba.batchSize {
        return ba.analyzeBatch()
    }
    
    return nil, nil
}

func (ba *BatchAnalyzer) analyzeBatch() (
    []extension.AnalysisResult, error) {
    results := make([]extension.AnalysisResult, 0, len(ba.batch))
    
    for _, data := range ba.batch {
        result, err := ba.analyzer.Analyze(data, nil)
        if err != nil {
            return nil, err
        }
        results = append(results, result)
    }
    
    ba.batch = ba.batch[:0] // 重用切片
    return results, nil
}
```plaintext

## 5. 测试最佳实践

### 5.1 单元测试

```go
func TestMyParser_Parse(t *testing.T) {
    type args struct {
        data []byte
    }
    tests := []struct {
        name    string
        args    args
        wantErr bool
        check   func(*MyExtensionData) bool
    }{
        {
            name:    "valid_data",
            args:    args{data: []byte{0x01, 0x02}},
            wantErr: false,
            check: func(ed *MyExtensionData) bool {
                return ed.ID == 0x01
            },
        },
        {
            name:    "empty_data",
            args:    args{data: []byte{}},
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            p := NewMyParser()
            got, err := p.Parse(tt.args.data, context.Background())
            
            if (err != nil) != tt.wantErr {
                t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if err == nil && !tt.check(got.(*MyExtensionData)) {
                t.Errorf("Parse() result check failed")
            }
        })
    }
}
```plaintext

### 5.2 集成测试

```go
func TestIntegration_ParseAndAnalyze(t *testing.T) {
    // 初始化引擎
    engine := extension.NewProcessingEngine(nil)
    
    // 准备请求
    request := &extension.ProcessingRequest{
        ExtensionType:  MyExtensionType,
        RawData:        []byte{0x01, 0x02, 0x03},
        Steps:          []string{"parse", "analyze"},
        Context:        context.Background(),
    }
    
    // 执行处理
    result := engine.Process(request)
    
    // 验证结果
    if !result.Success {
        t.Fatalf("Processing failed: %v", result.Error)
    }
    
    if len(result.AnalysisResults) == 0 {
        t.Fatal("No analysis results")
    }
    
    analysis := result.AnalysisResults[0]
    if analysis.GetRiskScore() < 0 || analysis.GetRiskScore() > 1 {
        t.Errorf("Invalid risk score: %f", analysis.GetRiskScore())
    }
}
```plaintext

### 5.3 性能测试

```go
func BenchmarkParse(b *testing.B) {
    p := NewMyParser()
    data := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
    ctx := context.Background()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        p.Parse(data, ctx)
    }
}

func BenchmarkAnalyze(b *testing.B) {
    a := NewMyAnalyzer()
    extData := &MyExtensionData{ID: 0x01}
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        a.Analyze(extData, nil)
    }
}
```plaintext

## 6. 文档编写

### 6.1 注释规范

```go
// MyParser 解析自定义扩展数据
// 
// 它实现了 extension.Parser 接口，负责将原始字节转换为
// MyExtensionData 结构。
//
// 示例：
//  parser := NewMyParser()
//  data, err := parser.Parse([]byte{0x01, 0x02}, ctx)
type MyParser struct{}

// Parse 解析扩展数据
//
// 参数：
//  - data: 原始扩展字节数据（不包含扩展头）
//  - parentContext: 父级上下文（如 ClientHello）
//
// 返回：
//  - ExtensionData: 解析后的扩展数据
//  - error: 解析失败原因
//
// 错误：
//  - ErrInvalidData: 数据格式无效
//  - ErrParseFailure: 解析失败
func (p *MyParser) Parse(data []byte, parentContext context.Context) (
    extension.ExtensionData, error) {
    // ...
}
```plaintext

### 6.2 README 结构

```markdown
# My Custom Extension

简单的单行描述。

## 功能

- 功能 1
- 功能 2
- 功能 3

## 安装

```go
import "myext"

myext.Init()
```plaintext

## 使用

展示基本使用示例。

## 配置

支持的配置选项。

## 性能

性能指标和基准测试。

## 贡献

如何贡献。

## License

MIT
```plaintext

## 7. 版本管理

### 7.1 语义版本化

遵循 [Semantic Versioning](https://semver.org/):

- **MAJOR**：不兼容的 API 变更
- **MINOR**：向后兼容的功能添加
- **PATCH**：向后兼容的错误修复

```go
const Version = "1.2.3" // Major.Minor.Patch
```plaintext

### 7.2 变更日志

```markdown
## [1.2.0] - 2026-02-28

### Added
- 支持新的扩展类型
- 添加性能优化

### Changed
- 更新 API 文档

### Deprecated
- 标记废弃的 API

### Removed
- 移除旧实现

### Fixed
- 修复 bug #123

### Security
- 修复安全漏洞
```plaintext

## 8. 安全考虑

### 8.1 输入验证

```go
func (p *MyParser) Parse(data []byte, ctx context.Context) (
    extension.ExtensionData, error) {
    
    // 验证数据长度
    if len(data) > 65536 { // 最大扩展大小
        return nil, ErrInvalidData
    }
    
    // 验证数据格式
    if len(data) < 4 {
        return nil, ErrInvalidData
    }
    
    // 验证字段值
    if data[0] == 0 && data[1] == 0 {
        return nil, ErrInvalidData // 无效的版本号
    }
    
    // ...
}
```plaintext

### 8.2 资源限制

```go
type RateLimiter struct {
    maxRequests int
    window      time.Duration
    requests    []time.Time
    mu          sync.Mutex
}

func (rl *RateLimiter) Allow() bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    now := time.Now()
    cutoff := now.Add(-rl.window)
    
    // 清理过期记录
    for len(rl.requests) > 0 && rl.requests[0].Before(cutoff) {
        rl.requests = rl.requests[1:]
    }
    
    if len(rl.requests) >= rl.maxRequests {
        return false
    }
    
    rl.requests = append(rl.requests, now)
    return true
}
```plaintext

## 9. 故障排查

### 常见问题

**Q: 分析器返回 panic**
A: 使用 defer + recover 机制，返回最大风险分数

**Q: 内存泄漏**
A: 清理循环引用，使用 sync.Pool 管理缓存

**Q: 性能下降**
A: 添加基准测试，使用 pprof 分析性能

## 10. 资源

- [官方 SDK 指南](./developer-guide.md)
- [架构文档](../architecture/modular-architecture.md)
- [API 参考](../3-references/00-quick-reference.md)

---

记住：**好代码是可读的、可维护的、可测试的代码。**
