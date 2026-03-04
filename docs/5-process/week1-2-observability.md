# Week 1-2: 可观测性集成实施指南

## 📌 阶段目标

✅ 为核心路径添加追踪和指标  
✅ 集成 OpenTelemetry + Prometheus + 结构化日志  
✅ 为 3 个关键组件添加可观测支持（BehaviorAnalyzer、ProcessingEngine、Pipeline）  
✅ 验证可观测能力工作正常

---

## 实施步骤

### Step 1: 依赖安装 ✅ 已完成

```bash
# go.mod 已更新，包含以下依赖：
- github.com/prometheus/client_golang v1.19.0
- go.opentelemetry.io/otel v1.24.0
- go.opentelemetry.io/otel/exporters/jaeger v1.24.0
- go.opentelemetry.io/otel/sdk/trace v1.24.0
- go.uber.org/zap v1.26.0

# 更新依赖
go mod tidy
```plaintext

### Step 2: 已创建的可观测组件 ✅

#### A. 核心可观测性库：`internal/observability/observability.go`
```plaintext
✅ FingerprintMetrics (指纹生成指标)
✅ BehaviorAnalysisMetrics (行为分析指标)
✅ PipelineMetrics (管道指标)
✅ TracingContext (OpenTelemetry 追踪)
✅ Logger (结构化日志)
✅ ObservabilityMiddleware (自动指标收集)
```plaintext

#### B. Pipeline 框架：`internal/pipeline/pipeline.go`
```plaintext
✅ Pipeline 主类
✅ 5 个内置中间件
✅ 完整的依赖管理
```plaintext

#### C. BehaviorAnalyzer 可观测包装：`security/behavior/instrumented.go`
```plaintext
✅ InstrumentedBehaviorAnalyzer
✅ 自动追踪和指标收集
✅ 结构化日志支持
```plaintext

#### D. ProcessingEngine 可观测包装：`internal/extension/instrumented.go`
```plaintext
✅ InstrumentedProcessingEngine
✅ 处理步骤级别的追踪
✅ 请求处理指标
```plaintext

---

## 使用指南

### 示例 1：为 BehaviorAnalyzer 启用可观测性

```go
package main

import (
	"context"
	"github.com/vistone/fingerprint/security/behavior"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

func main() {
	// 初始化日志和追踪
	cfg := zap.NewProductionConfig()
	logger, _ := cfg.Build()
	defer logger.Sync()
	sugar := logger.Sugar()

	tracer := otel.Tracer("example")

	// 创建原始分析器
	config := &behavior.BehaviorAnalysisConfig{
		TimeWindowSize:         60,
		MinRequestsForAnalysis: 5,
		RegularityThreshold:    0.3,
	}
	analyzer := behavior.NewBehaviorAnalyzer(config)

	// 包装成可观测版本
	observedAnalyzer := behavior.NewInstrumentedBehaviorAnalyzer(
		analyzer,
		tracer,
		sugar,
	)

	// 使用示例
	requestBehavior := &behavior.RequestBehavior{
		SourceIP:    "192.168.1.100",
		SNI:         "example.com",
		TLSVersion:  "1.3",
		HTTPVersion: "2",
	}

	// 记录行为（自动添加追踪和日志）
	observedAnalyzer.RecordBehavior(context.Background(), requestBehavior)

	// 分析（自动添加追踪、指标和日志）
	result, err := observedAnalyzer.AnalyzeReturnBehavior(
		context.Background(),
		"192.168.1.100",
		60,
	)

	if err != nil {
		sugar.Errorw("analysis failed", "error", err)
		return
	}

	if result != nil {
		sugar.Infow("analysis complete",
			"risk_score", result.RiskScore,
			"anomaly_count", len(result.DetectedSignals),
		)
	}
}
```plaintext

### 示例 2：为 ProcessingEngine 启用可观测性

```go
package main

import (
	"context"
	"github.com/vistone/fingerprint/internal/extension"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

func main() {
	cfg := zap.NewProductionConfig()
	logger, _ := cfg.Build()
	defer logger.Sync()
	sugar := logger.Sugar()

	tracer := otel.Tracer("example")

	// 创建原始处理引擎
	engineConfig := &extension.EngineConfig{
		ConcurrentProcessing: true,
		MaxConcurrency:       10,
		EnableCaching:        true,
		CacheSize:            1000,
		TimeoutMs:            5000,
	}
	engine := extension.NewProcessingEngine(engineConfig, registry)

	// 包装成可观测版本
	observedEngine := extension.NewInstrumentedProcessingEngine(
		engine,
		tracer,
		sugar,
	)

	// 使用示例
	request := &extension.ProcessingRequest{
		ExtensionType: 10,
		RawData:       []byte{0x16, 0x03, 0x01, 0x00, 0x05},
		Steps:         []string{"parse", "analyze", "transform"},
		Context:       context.Background(),
	}

	// 处理请求（自动添加追踪、指标和日志）
	result := observedEngine.Process(request)

	if result.Success {
		sugar.Infow("processing complete",
			"extension_type", request.ExtensionType,
			"duration_ms", result.ElapsedMs,
		)
	} else {
		sugar.Errorw("processing failed", "error", result.Error)
	}

	// 获取指标快照
	metrics := observedEngine.GetMetricsSnapshot()
	sugar.Infow("engine metrics", "metrics", metrics)
}
```plaintext

### 示例 3：使用 Pipeline 框架和可观测性

```go
package main

import (
	"context"
	"github.com/vistone/fingerprint/internal/pipeline"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type MyParseStage struct{}

func (s *MyParseStage) GetName() string {
	return "custom-parse"
}

func (s *MyParseStage) GetDependencies() []string {
	return []string{}
}

func (s *MyParseStage) Execute(ctx context.Context, data *pipeline.StageData) error {
	// 自定义解析逻辑
	data.Output = map[string]interface{}{
		"parsed": true,
	}
	return nil
}

func main() {
	cfg := zap.NewProductionConfig()
	logger, _ := cfg.Build()
	defer logger.Sync()
	sugar := logger.Sugar()

	tracer := otel.Tracer("example")

	// 创建 Pipeline（自动支持可观测性）
	p := pipeline.NewPipeline(tracer).
		AddStage(&MyParseStage()).
		AddMiddleware(pipeline.NewLoggingMiddleware(mockLogger)).
		AddMiddleware(pipeline.NewMetricsMiddleware(mockMetrics)).
		AddMiddleware(pipeline.NewTimeoutMiddleware(5 * time.Second))

	// 执行（自动添加追踪、日志和指标）
	result, err := p.Execute(context.Background(), []byte{0x16, 0x03, 0x01})
	if err != nil {
		sugar.Errorw("pipeline failed", "error", err)
		return
	}

	sugar.Infow("pipeline executed",
		"duration_ms", result.Duration.Milliseconds(),
		"output", result.Output,
	)
}
```plaintext

---

## 验证检查清单 ✅

### 编译验证
```bash
# 1. 编译新组件
go build ./security/behavior
go build ./internal/extension
go build ./internal/pipeline
go build ./internal/observability

# 2. 运行所有测试
go test ./... -v

# 3. 检查是否有编译错误
go build ./...
```plaintext

### 功能验证
```bash
# 1. 验证 BehaviorAnalyzer 可观测包装
go test -v ./security/behavior/ -run Instrumented

# 2. 验证 ProcessingEngine 可观测包装
go test -v ./internal/extension/ -run Instrumented

# 3. 验证 Pipeline 框架
go test -v ./internal/pipeline/examples_test.go
```plaintext

---

## 下一步：集成到实际代码

### 任务 1：为现有 BehaviorAnalyzer 用户添加可观测性

找到所有使用 BehaviorAnalyzer 的地方，用可观测版本包装：

```bash
# 搜索使用 BehaviorAnalyzer 的代码
grep -r "BehaviorAnalyzer" --include="*.go" | grep -v "instrumented.go"
```plaintext

### 任务 2：为现有 ProcessingEngine 用户添加可观测性

类似地，包装所有 ProcessingEngine 的实例：

```bash
# 搜索使用 ProcessingEngine 的代码
grep -r "ProcessingEngine" --include="*.go" | grep -v "instrumented.go"
```plaintext

### 任务 3：创建可观测性初始化助手

创建一个 `observability/setup.go` 文件来统一初始化：

```go
package observability

import (
	"context"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"net/http"
)

// SetupObservability 一次性初始化所有可观测能力
func SetupObservability(serviceName string) (*zap.SugaredLogger, error) {
	// 初始化日志
	cfg := zap.NewProductionConfig()
	logger, _ := cfg.Build()
	sugar := logger.Sugar()
	
	// 初始化 Jaeger 追踪
	exp, _ := jaeger.New(
		jaeger.WithCollectorEndpoint(
			jaeger.WithEndpoint("http://localhost:14268/api/traces"),
		),
	)
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exp),
	)
	otel.SetTracerProvider(tp)
	
	// 暴露 Prometheus 指标
	go http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":8080", nil)
	
	sugar.Infow("observability setup complete", "service", serviceName)
	return sugar, nil
}
```plaintext

---

## 性能检查

运行基准测试验证开销：

```bash
# 编译时性能开销
time go build ./...

# 运行时性能开销
go test -bench=. -benchmem ./internal/pipeline/
go test -bench=. -benchmem ./security/behavior/
```plaintext

**预期结果**：
- 可观测性开销 < 2%
- Pipeline 中间件开销 < 5%

---

## 完成标准

✅ 所有新文件都能编译
✅ 现有测试全部通过  
✅ 至少为 3 个关键组件添加了可观测支持
✅ 创建了使用示例文档
✅ 性能开销在预期范围内

---

## 常见问题

### Q1: 我的应用如何启用这些可观测能力？

**A**: 使用包装器类：
```go
observedAnalyzer := behavior.NewInstrumentedBehaviorAnalyzer(
    originalAnalyzer, tracer, logger)
```plaintext

### Q2: 追踪数据去哪里了？

**A**: 如果配置了 Jaeger 导出，数据将发送到 Jaeger。否则，它们将被缓冲在内存中。

### Q3: 不想要某个中间件怎么办？

**A**: 不添加即可，中间件是可选的：
```go
pipeline := pipeline.NewPipeline(tracer).
    AddStage(stage1).
    // 只添加需要的中间件
    AddMiddleware(pipeline.NewLoggingMiddleware(logger))
```plaintext

---

## 下周预告

**Week 3-4：流水线框架 PoC**
- 迁移一个现有的处理流程到 Pipeline
- 验证性能和功能
- 对比 vs 当前 switch-case 方案
