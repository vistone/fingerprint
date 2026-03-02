# fingerprint SDK 开发者指南

## 概述

fingerprint SDK 提供了灵活的扩展架构，允许开发者：

1. **创建自定义扩展** - 支持新的 TLS 扩展类型
2. **开发解析器** - 将原始字节数据转换为结构化对象
3. **实现分析器** - 分析扩展数据并生成风险评分
4. **构建处理器** - 事件驱动的流式处理
5. **开发第三方插件** - 完整的生态系统集成

## 快速开始

### 1. 注册新扩展类型

```go
package myext

import (
	"github.com/vistone/fingerprint/internal/extension"
)

// 定义新扩展常量
const MyExtensionType extension.ExtensionType = 0x1234

// 注册元数据
func RegisterMyExtension() error {
	metadata := &extension.ExtensionMetadata{
		Type:                  MyExtensionType,
		Name:                  "My Custom Extension",
		Description:           "Custom extension for fingerprinting",
		RFC:                   "RFC XXXX",
		IANANumber:            0x1234,
		Category:              extension.CategoryCustom,
		IsExperimental:        true,
		CompatibleTLSVersions: []uint16{0x0304}, // TLS 1.3
	}

	return extension.RegisterExtension(metadata)
}
```

### 2. 实现解析器

```go
package myext

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/vistone/fingerprint/internal/extension"
)

// MyExtensionData 自定义扩展数据
type MyExtensionData struct {
	Type    extension.ExtensionType
	ID      uint8
	Payload []byte
	RawData []byte
}

func (e *MyExtensionData) GetType() extension.ExtensionType {
	return e.Type
}

func (e *MyExtensionData) GetRawData() []byte {
	return e.RawData
}

func (e *MyExtensionData) GetName() string {
	return "My Custom Extension"
}

func (e *MyExtensionData) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"type":     fmt.Sprintf("0x%04x", e.Type),
		"id":       e.ID,
		"payload":  fmt.Sprintf("%x", e.Payload),
	}
}

// MyParser 自定义解析器
type MyParser struct {
	version string
}

func NewMyParser() *MyParser {
	return &MyParser{version: "1.0.0"}
}

func (p *MyParser) Parse(data []byte, parentContext context.Context) (
	extension.ExtensionData, error) {
	if len(data) < 1 {
		return nil, fmt.Errorf("insufficient data")
	}

	extData := &MyExtensionData{
		Type:    MyExtensionType,
		ID:      data[0],
		RawData: make([]byte, len(data)),
	}
	copy(extData.RawData, data)

	if len(data) > 1 {
		extData.Payload = make([]byte, len(data)-1)
		copy(extData.Payload, data[1:])
	}

	return extData, nil
}

func (p *MyParser) GetType() extension.ExtensionType {
	return MyExtensionType
}

func (p *MyParser) GetVersion() string {
	return p.version
}

// 注册解析器
func RegisterMyParser() error {
	return extension.RegisterParserBuilder(MyExtensionType, func() (
		extension.Parser, error) {
		return NewMyParser(), nil
	})
}
```

### 3. 实现分析器

```go
package myext

import (
	"fmt"
	"github.com/vistone/fingerprint/internal/extension"
)

// MyAnalysisResult 分析结果
type MyAnalysisResult struct {
	ExtType      extension.ExtensionType
	RiskScore    float64
	Anomalies    []string
	Observations map[string]interface{}
}

func (r *MyAnalysisResult) GetExtensionType() extension.ExtensionType {
	return r.ExtType
}

func (r *MyAnalysisResult) HasAnomalies() bool {
	return len(r.Anomalies) > 0
}

func (r *MyAnalysisResult) GetAnomalies() []string {
	result := make([]string, len(r.Anomalies))
	copy(result, r.Anomalies)
	return result
}

func (r *MyAnalysisResult) GetRiskScore() float64 {
	return r.RiskScore
}

func (r *MyAnalysisResult) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"extension_type": fmt.Sprintf("0x%04x", r.ExtType),
		"risk_score":     r.RiskScore,
		"anomalies":      r.Anomalies,
		"observations":   r.Observations,
	}
}

// MyAnalyzer 自定义分析器
type MyAnalyzer struct {
	version string
}

func NewMyAnalyzer() *MyAnalyzer {
	return &MyAnalyzer{version: "1.0.0"}
}

func (a *MyAnalyzer) Analyze(data extension.ExtensionData,
	config map[string]interface{}) (extension.AnalysisResult, error) {
	myData, ok := data.(*MyExtensionData)
	if !ok {
		return nil, fmt.Errorf("invalid data type")
	}

	result := &MyAnalysisResult{
		ExtType:      MyExtensionType,
		RiskScore:    0.0,
		Anomalies:    []string{},
		Observations: make(map[string]interface{}),
	}

	// 执行分析逻辑
	if len(myData.Payload) == 0 {
		result.Anomalies = append(result.Anomalies, "empty_payload")
		result.RiskScore = 0.3
	}

	result.Observations["payload_size"] = len(myData.Payload)
	result.Observations["id"] = myData.ID

	return result, nil
}

func (a *MyAnalyzer) GetType() extension.ExtensionType {
	return MyExtensionType
}

func (a *MyAnalyzer) GetVersion() string {
	return a.version
}

func (a *MyAnalyzer) SupportsConfig() []string {
	return []string{"strict_mode", "custom_threshold"}
}

// 注册分析器
func RegisterMyAnalyzer() error {
	return extension.RegisterAnalyzerBuilder(MyExtensionType, func() (
		extension.Analyzer, error) {
		return NewMyAnalyzer(), nil
	})
}
```

### 4. 实现处理器（可选）

```go
package myext

import (
	"fmt"
	"github.com/vistone/fingerprint/internal/extension"
)

// MyHandler 事件处理器
type MyHandler struct {
	name     string
	priority int
}

func NewMyHandler() *MyHandler {
	return &MyHandler{
		name:     "MyCustomHandler",
		priority: 50,
	}
}

func (h *MyHandler) Handle(event *extension.ExtensionEvent) (
	*extension.EventResult, error) {
	result := &extension.EventResult{
		Success:            true,
		ContinueProcessing: true,
		HandlerName:        h.GetName(),
	}

	// 处理事件
	if data, ok := event.Data.(*MyExtensionData); ok {
		result.Result = map[string]interface{}{
			"handler": h.name,
			"id":      data.ID,
		}
	}

	return result, nil
}

func (h *MyHandler) GetType() extension.ExtensionType {
	return MyExtensionType
}

func (h *MyHandler) GetPriority() int {
	return h.priority
}

func (h *MyHandler) GetName() string {
	return h.name
}

// 注册处理器
func RegisterMyHandler() error {
	handler := NewMyHandler()
	return extension.RegisterHandler(MyExtensionType, handler)
}
```

### 5. 整合所有组件

```go
package myext

// Setup 初始化所有组件
func Setup() error {
	// 注册扩展
	if err := RegisterMyExtension(); err != nil {
		return fmt.Errorf("failed to register extension: %w", err)
	}

	// 注册解析器
	if err := RegisterMyParser(); err != nil {
		return fmt.Errorf("failed to register parser: %w", err)
	}

	// 注册分析器
	if err := RegisterMyAnalyzer(); err != nil {
		return fmt.Errorf("failed to register analyzer: %w", err)
	}

	// 注册处理器（可选）
	if err := RegisterMyHandler(); err != nil {
		return fmt.Errorf("failed to register handler: %w", err)
	}

	return nil
}
```

### 6. 使用处理引擎

```go
package main

import (
	"context"
	"fmt"
	"myext"
	"github.com/vistone/fingerprint/internal/extension"
)

func main() {
	// 初始化扩展
	myext.Setup()

	// 创建处理引擎
	config := &extension.EngineConfig{
		ConcurrentProcessing: false,
		EnableCaching:        true,
		StrictValidation:     true,
		VerboseLogging:       true,
	}
	engine := extension.NewProcessingEngine(config)

	// 准备请求
	request := &extension.ProcessingRequest{
		ExtensionType:  myext.MyExtensionType,
		RawData:        []byte{0x01, 0x02, 0x03},
		Steps:          []string{"parse", "analyze"},
		AnalysisConfig: nil,
		Context:        context.Background(),
		Metadata:       make(map[string]interface{}),
	}

	// 处理请求
	result := engine.Process(request)

	// 输出结果
	if result.Success {
		for _, analysis := range result.AnalysisResults {
			fmt.Printf("Risk Score: %.2f\n", analysis.GetRiskScore())
			fmt.Printf("Anomalies: %v\n", analysis.GetAnomalies())
		}
	} else {
		fmt.Printf("Error: %s\n", result.Error)
	}
}
```

## 高级用法

### 使用拦截器插入自定义逻辑

```go
type CustomInterceptor struct{}

func (c *CustomInterceptor) Intercept(phase string,
	request *extension.ProcessingRequest,
	result *extension.ProcessingResult) error {
	if phase == "pre" {
		fmt.Println("Pre-processing started")
	} else if phase == "post" {
		fmt.Println("Post-processing completed")
	}
	return nil
}

// 注册拦截器
engine.RegisterInterceptor("pre", &CustomInterceptor{})
engine.RegisterInterceptor("post", &CustomInterceptor{})
```

### 创建第三方插件

```go
type MyPlugin struct {
	info   *extension.PluginInfo
	config map[string]interface{}
}

func (p *MyPlugin) GetInfo() *extension.PluginInfo {
	return p.info
}

func (p *MyPlugin) Init(config map[string]interface{}) error {
	p.config = config
	return nil
}

func (p *MyPlugin) Register() error {
	// 注册扩展、解析器、分析器等
	return myext.Setup()
}

func (p *MyPlugin) Unload() error {
	// 清理资源
	return nil
}

func (p *MyPlugin) Validate() error {
	// 验证插件有效性
	return nil
}

func (p *MyPlugin) GetDependencies() []string {
	return []string{}
}

func (p *MyPlugin) GetVersion() string {
	return "1.0.0"
}

// 注册插件
func RegisterMyPlugin() error {
	plugin := &MyPlugin{
		info: &extension.PluginInfo{
			ID:      "my-plugin",
			Name:    "My Plugin",
			Version: "1.0.0",
			Author:  "Your Name",
		},
	}
	return extension.RegisterPlugin("my-plugin", plugin)
}
```

## API 参考

### 核心接口

#### Parser 接口
```go
type Parser interface {
    Parse(data []byte, parentContext context.Context) (ExtensionData, error)
    GetType() ExtensionType
    GetVersion() string
}
```

#### Analyzer 接口
```go
type Analyzer interface {
    Analyze(data ExtensionData, config map[string]interface{}) (AnalysisResult, error)
    GetType() ExtensionType
    GetVersion() string
    SupportsConfig() []string
}
```

#### Handler 接口
```go
type Handler interface {
    Handle(event *ExtensionEvent) (*EventResult, error)
    GetType() ExtensionType
    GetPriority() int
    GetName() string
}
```

### 注册表 API

```go
// 扩展注册
RegisterExtension(metadata *ExtensionMetadata) error

// 组件注册
RegisterParser(extType ExtensionType, parser Parser) error
RegisterAnalyzer(extType ExtensionType, analyzer Analyzer) error
RegisterHandler(extType ExtensionType, handler Handler) error

// 组件查询
GetParser(extType ExtensionType) (Parser, error)
GetAnalyzer(extType ExtensionType) (Analyzer, error)
GetHandlers(extType ExtensionType) []Handler

// 列表和查找
ListAllExtensions() []*ExtensionMetadata
FindExtensionByName(name string) (ExtensionType, error)
```

## 最佳实践

### 1. 版本管理
- 为每个组件实现版本号
- 使用语义版本（Semantic Versioning）
- 记录变更日志

### 2. 错误处理
- 总是返回有意义的错误消息
- 使用标准错误定义
- 实现错误恢复逻辑

### 3. 性能优化
- 使用缓存减少重复计算
- 实现并发处理
- 测试性能基准

### 4. 文档
- 编写清晰的文档
- 提供完整的示例代码
- 记录配置选项

### 5. 测试
- 编写单元测试
- 进行集成测试
- 测试边界条件

## 疑难解答

### 问题：注册扩展失败

**原因**：扩展已存在或元数据无效

**解决方案**：
```go
// 检查扩展是否已注册
metadata, err := extension.GetMetadata(MyExtensionType)
if err == nil {
    fmt.Println("Extension already registered")
}
```

### 问题：解析器未找到

**原因**：解析器未注册

**解决方案**：确保在序列化前调用 `RegisterParserBuilder()`

### 问题：分析器返回异常错误

**原因**：数据格式不匹配

**解决方案**：在 Analyze() 中进行类型断言和验证

## 联系方式

- GitHub: https://github.com/vistone/fingerprint
- Issues: https://github.com/vistone/fingerprint/issues
- Discussions: https://github.com/vistone/fingerprint/discussions

## License

MIT License - 详见 LICENSE 文件
