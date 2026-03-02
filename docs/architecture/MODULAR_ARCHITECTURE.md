# 模块化扩展架构设计文档

## 概述

相比原始的单体式设计，新的模块化扩展架构采用以下原则：

- **解耦** - 核心逻辑与扩展逻辑完全分离
- **可扩展** - 用户可轻松添加新的扩展类型
- **可插拔** - 支持动态加载/卸载组件
- **灵活** - 支持多种处理方式（解析、分析、转换、处理）
- **开放** - 提供清晰的扩展接口，鼓励社区贡献

## 架构详解

### 1. 扩展注册表（Extension Registry）

**问题**：原始代码中扩展类型被硬编码为常量（`0xfe0d`、`0xfd00` 等）

**解决方案**：集中式注册表管理所有扩展元信息

```
┌─────────────────────────────┐
│  Extension Registry(全局)   │
├─────────────────────────────┤
│  metadata: 扩展元数据        │
│  parsers: 解析器映射         │
│  analyzers: 分析器映射       │
│  handlers: 处理器映射        │
│  plugins: 第三方插件         │
└─────────────────────────────┘
```

**优点**：
- 新扩展注册时不需要修改核心代码
- 支持运行时动态注册扩展
- 便于扩展发现和管理

### 2. 解析器工厂（Parser Factory）

**问题**：每种扩展都需要一个专属的解析器，但代码中没有统一的生成机制

**解决方案**：工厂模式创建解析器实例

```
解析流程：
RawData → [Parser] → ExtensionData (结构化)
```

**实现**：
```go
type Parser interface {
    Parse(data []byte, ctx context.Context) (ExtensionData, error)
    GetType() ExtensionType
    GetVersion() string
}
```

**好处**：
- 每个扩展有独立的解析逻辑
- 支持版本管理（不同版本的 RFC）
- 易于测试和维护

### 3. 分析器工厂（Analyzer Factory）

**问题**：分析逻辑耦合在各种方法中，难以独立升级

**解决方案**：独立的分析引擎

```
分析流程：
ExtensionData + Config → [Analyzer] → AnalysisResult
```

**实现**：
```go
type Analyzer interface {
    Analyze(data ExtensionData, config map[string]interface{}) (AnalysisResult, error)
    GetType() ExtensionType
    GetVersion() string
    SupportsConfig() []string
}
```

**优点**：
- 风险评分算法可独立优化
- 支持支持多个分析器链式处理
- 异常检测逻辑独立，易于扩展

### 4. 处理器工厂（Handler Factory）

**问题**：事件驱动的处理流程需要灵活的中间件支持

**解决方案**：事件处理器模式

```
处理流程：
Event → [Handler1] → [Handler2] → ... → Result
         (priority 100)  (priority 50)
```

**实现**：
```go
type Handler interface {
    Handle(event *ExtensionEvent) (*EventResult, error)
    GetType() ExtensionType
    GetPriority() int
    GetName() string
}
```

**使用场景**：
- 日志记录
- 性能监控
- 异常上报
- 数据转换

### 5. 处理引擎（Processing Engine）

**核心协调器**：
- 管理完整的处理流程
- 支持多种处理步骤（parse → analyze → handle → transform）
- 拦截器支持
- 异常处理和恢复

```
Processing Flow:
Request → [Pre-Interceptor] → Parse → Analyze → Handle → Transform → 
          [Post-Interceptor] → Result
```

## 迁移指南

### 从硬编码扩展到注册表

**之前**：
```go
const (
    ExtensionECHOuterExtensions = 0xfd00
    ExtensionEncryptedClientHello = 0xfe0d
)

if extensions[i].Type == 0xfe0d || extensions[i].Type == 0xfd00 {
    // 处理逻辑
}
```

**之后**：
```go
// 一次性注册
RegisterExtension(&ExtensionMetadata{
    Type: 0xfe0d,
    Name: "Encrypted Client Hello",
    // ...
})

// 然后使用
if echExt := findExtensionByType(0xfe0d); echExt != nil {
    // 处理逻辑
}
```

### 从单体解析器到模块化解析器

**之前**：
```go
func (a *ECHAnalyzer) findECHExtension(extensions []ExtensionData) *ExtensionData {
    for i := range extensions {
        if extensions[i].Type == 0xfe0d || extensions[i].Type == 0xfd00 {
            return &extensions[i]
        }
    }
    return nil
}
```

**之后**：
```go
// 实现 Parser 接口
func (p *ECHParser) Parse(data []byte, ctx context.Context) (ExtensionData, error) {
    // 专注于解析逻辑
}

// 注册
RegisterParserBuilder(0xfe0d, func() (Parser, error) {
    return NewECHParser(), nil
})

// 使用
parser, _ := GetParser(extensionType)
parsedData, _ := parser.Parse(rawData, ctx)
```

## 扩展性设计

### 场景 1：添加新的 TLS 扩展支持

**步骤**：

1. 定义扩展常量
```go
const MyExtensionType ExtensionType = 0x1234
```

2. 注册元数据
```go
RegisterExtension(&ExtensionMetadata{
    Type: MyExtensionType,
    Name: "My Extension",
    // ...
})
```

3. 实现解析器
```go
type MyParser struct{}
func (p *MyParser) Parse(data []byte, ctx context.Context) (ExtensionData, error) {
    // 实现解析逻辑
}
```

4. 实现分析器
```go
type MyAnalyzer struct{}
func (a *MyAnalyzer) Analyze(data ExtensionData, config map[string]interface{}) (AnalysisResult, error) {
    // 实现分析逻辑
}
```

5. 注册组件
```go
RegisterParserBuilder(MyExtensionType, func() (Parser, error) {
    return NewMyParser(), nil
})
RegisterAnalyzerBuilder(MyExtensionType, func() (Analyzer, error) {
    return NewMyAnalyzer(), nil
})
```

### 场景 2：创建自定义处理器

监控、日志记录或数据转换：

```go
type LoggingHandler struct{}

func (h *LoggingHandler) Handle(event *ExtensionEvent) (*EventResult, error) {
    log.Printf("Processing extension: %d", event.ExtensionType)
    return &EventResult{
        Success: true,
        ContinueProcessing: true,
    }, nil
}
```

### 场景 3：开发第三方插件

完整的生态系统集成：

```go
type MyPlugin struct{}

func (p *MyPlugin) Register() error {
    // 注册所有组件
    RegisterExtension(...)
    RegisterParserBuilder(...)
    RegisterAnalyzerBuilder(...)
}

// 在应用启动时加载
RegisterPlugin("my-plugin", &MyPlugin{})
```

## 性能考虑

### 1. 缓存策略

```go
config := &EngineConfig{
    EnableCaching: true,
    CacheSize:    1000,
}
engine := NewProcessingEngine(config)
```

### 2. 并发处理

```go
config := &EngineConfig{
    ConcurrentProcessing: true,
    MaxConcurrency:       16,
}
```

### 3. 超时管理

```go
config := &EngineConfig{
    TimeoutMs: 5000, // 5秒超时
}
```

## 与现有代码的兼容性

新架构与现有的 `ech_analysis.go`、`ch_negotiation.go` 等完全兼容：

- 不强制重写现有分析器
- 可逐步迁移到新架构
- 支持混合使用（新旧组件共存）

### 兼容性适配层示例

```go
// 包装现有的 ECH 分析器
type LegacyECHAnalyzerAdapter struct {
    legacy *ECHAnalyzer
}

func (a *LegacyECHAnalyzerAdapter) Analyze(data ExtensionData, 
    config map[string]interface{}) (AnalysisResult, error) {
    // 转换新格式为旧格式
    // 调用旧分析器
    // 转换结果回新格式
}
```

## 质量保证

### 测试框架

```go
// 单元测试：测试每个组件
func TestMyParser_Parse(t *testing.T) {
    parser := NewMyParser()
    data, err := parser.Parse([]byte{...}, context.Background())
    assert.NotNil(t, data)
    assert.Nil(t, err)
}

// 集成测试：测试完整流程
func TestProcessingEngine_Integration(t *testing.T) {
    engine := NewProcessingEngine(nil)
    result := engine.Process(&ProcessingRequest{...})
    assert.True(t, result.Success)
}
```

### 性能基准

```go
// 基准测试
func BenchmarkParse(b *testing.B) {
    parser := NewMyParser()
    for i := 0; i < b.N; i++ {
        parser.Parse(data, ctx)
    }
}
```

## 文档和示例

- [开发者指南](DEVELOPER_GUIDE.md) - 完整的 SDK 使用说明
- [示例集合](../examples/) - 多个实现示例
- [API 参考](API_REFERENCE.md) - 完整的 API 文档（即将推出）

## 社区贡献

我们欢迎社区成员：

1. 提交 RFC（Request for Comments）讨论新扩展
2. 贡献新的解析器和分析器
3. 分享第三方插件
4. 完善文档和示例

## 常见问题

### Q: 如何禁用某个扩展？
A: 目前架构中已注册的扩展无法禁用。未来版本会支持扩展启用/禁用配置。

### Q: 解析器版本向后兼容吗？
A: 强烈建议保持向后兼容。如需破坏性修改，请通过版本号标示明确。

### Q: 如何自定义风险评分算法？
A: 实现自定义的 Analyzer 接口，通过 RegisterAnalyzerBuilder() 注册覆盖默认实现。

### Q: 支持动态加载 Python/JavaScript 扩展吗？
A: 当前版本仅支持 Go。未来可考虑通过 WASM 支持其他语言。

## 路线图

### v2.1.0（近期）
- [ ] 完整的 API 参考文档
- [ ] 更多示例和教程
- [ ] 性能优化

### v2.2.0（中期）
- [ ] 扩展启用/禁用配置
- [ ] 性能监控面板
- [ ] 插件市场概念设计

### v2.3.0（长期）
- [ ] WASM 支持
- [ ] 图形化配置工具
- [ ] 云端扩展共享平台
