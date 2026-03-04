# Week 3-4: 流水线框架 PoC 执行计划

**开始时间**: 2026年3月2日  
**目标完成**: 2026年3月16日  
**改造阶段**: Week 3-4 (第3-4周)

---

## 📋 执行计划概览

### 目标 1: 建立 Pipeline 核心框架 ✓ Day 1-2

- [ ] 创建 `internal/pipeline/` 目录结构
- [ ] 实现 `Pipeline` 主类
- [ ] 实现 `Stage` 接口
- [ ] 实现 `Middleware` 接口
- [ ] 完成单元测试

**预期成果**: 100 行代码，完全可编译

### 目标 2: 迁移现有 ProcessingEngine 步骤 ✓ Day 3-5

- [ ] 创建 `ParseStage` (解析 TLS 扩展)
- [ ] 创建 `AnalyzeStage` (分析指纹)
- [ ] 创建 `TransformStage` (转换结果)
- [ ] 验证依赖管理和错误处理

**预期成果**: 3 个 Stage 实现

### 目标 3: 实现中间件系统 ✓ Day 6-7

- [ ] 创建 `LoggingMiddleware` (日志记录)
- [ ] 创建 `MetricsMiddleware` (性能指标)
- [ ] 创建 `TimeoutMiddleware` (超时控制)
- [ ] 创建 `ErrorHandlingMiddleware` (错误处理)

**预期成果**: 4 个中间件，完整集成

### 目标 4: PoC 集成测试 ✓ Day 8-9

- [ ] 创建集成示例 (main.go)
- [ ] 对比 switch-case vs Pipeline 性能
- [ ] 验证功能正确性
- [ ] 编写使用文档

**预期成果**: 功能验证通过，性能开销 < 5%

### 目标 5: 完成总结与文档 ✓ Day 10

- [ ] 编写周报
- [ ] 准备下一步计划
- [ ] 代码审查

**预期成果**: 完整文档，为 Week 5-6 做准备

---

## 🎯 Day 1-2: Pipeline 核心框架

### Step 1.1: 创建目录结构

```bash
internal/pipeline/
├── pipeline.go          # 核心 Pipeline 类
├── context.go           # 执行上下文
├── middleware.go        # 中间件接口
├── stages.go            # Stage 实现
├── errors.go            # 错误类型
└── pipeline_test.go     # 单元测试
```plaintext

### Step 1.2: 实现 Pipeline 接口

```go
// internal/pipeline/pipeline.go
package pipeline

import "context"

// Stage 流水线的单个阶段
type Stage interface {
    GetName() string           // 阶段名称
    GetDependencies() []string // 依赖的前置阶段
    Execute(ctx context.Context, data *PipelineData) error
}

// Middleware 中间件接口
type Middleware interface {
    Process(ctx context.Context, stageName string, data *PipelineData, next ExecutionFunc) error
}

type ExecutionFunc func(context.Context, *PipelineData) error

// PipelineData 流水线执行数据
type PipelineData struct {
    Input      interface{}
    Output     interface{}
    Context    map[string]interface{}
    Metadata   map[string]string
    Error      error
    Duration   time.Duration
    ExecutedAt time.Time
}

// Pipeline 主类
type Pipeline struct {
    stages      []Stage
    middlewares []Middleware
    stageIndex  map[string]int
}

// Execute 执行流水线
func (p *Pipeline) Execute(ctx context.Context, input interface{}) (*PipelineData, error) {
    // 实现细节见下面的代码
    return nil, nil
}
```plaintext

### Step 1.3: 单元测试框架

```go
// internal/pipeline/pipeline_test.go
package pipeline

import "testing"

func TestPipelineExecution(t *testing.T) {
    // 测试 Pipeline 基本执行
}

func TestDependencyManagement(t *testing.T) {
    // 测试依赖检查
}

func TestMiddlewareChaining(t *testing.T) {
    // 测试中间件链
}
```plaintext

---

## 🔄 Day 3-5: Stage 实现与迁移

### Step 2.1: 创建抽象 Stage

```go
// internal/pipeline/stages.go

// ParseStage: 从原始 TLS 数据解析指纹
type ParseStage struct {
    parsers map[string]Parsers // ja3, ja4, ja4s 等
}

func (ps *ParseStage) GetName() string { return "parse" }
func (ps *ParseStage) GetDependencies() []string { return []string{} }
func (ps *ParseStage) Execute(ctx context.Context, data *PipelineData) error {
    // 从 data.Input ([]byte) 解析出指纹
    // 存储到 data.Output
    return nil
}

// AnalyzeStage: 分析指纹特征
type AnalyzeStage struct {
    analyzer BehaviorAnalyzer
}

func (as *AnalyzeStage) GetName() string { return "analyze" }
func (as *AnalyzeStage) GetDependencies() []string { return []string{"parse"} }
func (as *AnalyzeStage) Execute(ctx context.Context, data *PipelineData) error {
    // 从 data.Output (解析结果) 进行分析
    // 返回分析结果
    return nil
}

// TransformStage: 转换为最终格式
type TransformStage struct {
    transformer ResultTransformer
}

func (ts *TransformStage) GetName() string { return "transform" }
func (ts *TransformStage) GetDependencies() []string { return []string{"parse", "analyze"} }
func (ts *TransformStage) Execute(ctx context.Context, data *PipelineData) error {
    // 将分析结果转换为最终格式（JSON/Protobuf）
    return nil
}
```plaintext

### Step 2.2: 从 ProcessingEngine 迁移

对比以下两种方式的行为：

**旧方式** (ProcessingEngine.Process switch-case):
```go
for _, step := range steps {
    switch step {
    case "parse": // ... 30 行
    case "analyze": // ... 40 行
    case "transform": // ... 20 行
    }
}
```plaintext

**新方式** (Pipeline):
```go
pipeline.AddStage(&ParseStage{...}).
    AddStage(&AnalyzeStage{...}).
    AddStage(&TransformStage{...}).
    Execute(ctx, rawData)
```plaintext

---

## 🔧 Day 6-7: 中间件实现

### Step 3.1: 日志中间件

```go
// internal/pipeline/middlewares.go

type LoggingMiddleware struct {
    logger *zap.SugaredLogger
}

func (lm *LoggingMiddleware) Process(ctx context.Context, stageName string, 
    data *PipelineData, next ExecutionFunc) error {
    
    lm.logger.Infow("stage started", "stage", stageName)
    
    startTime := time.Now()
    err := next(ctx, data)
    duration := time.Since(startTime)
    
    if err != nil {
        lm.logger.Errorw("stage failed", "stage", stageName, "error", err, "duration_ms", duration.Milliseconds())
    } else {
        lm.logger.Infow("stage completed", "stage", stageName, "duration_ms", duration.Milliseconds())
    }
    
    data.Duration = duration
    return err
}
```plaintext

### Step 3.2: 指标中间件

```go
type MetricsMiddleware struct {
    metrics map[string]prometheus.Histogram
}

func (mm *MetricsMiddleware) Process(ctx context.Context, stageName string,
    data *PipelineData, next ExecutionFunc) error {
    
    duration := mm.recordStart()
    err := next(ctx, data)
    mm.recordEnd(stageName, duration, err == nil)
    
    return err
}
```plaintext

### Step 3.3: 超时中间件

```go
type TimeoutMiddleware struct {
    timeout time.Duration
}

func (tm *TimeoutMiddleware) Process(ctx context.Context, stageName string,
    data *PipelineData, next ExecutionFunc) error {
    
    ctx, cancel := context.WithTimeout(ctx, tm.timeout)
    defer cancel()
    
    return next(ctx, data)
}
```plaintext

---

## ✅ Day 8-9: PoC 集成与性能验证

### Step 4.1: 创建集成示例

```go
// internal/pipeline/examples/main.go

func main() {
    // 初始化
    logger := initLogger()
    tracer := initTracer()
    
    // 创建 Pipeline
    pipeline := pipeline.NewPipeline().
        AddStage(&ParseStage{parsers: initParsers()}).
        AddStage(&AnalyzeStage{analyzer: initAnalyzer()}).
        AddStage(&TransformStage{transformer: initTransformer()}).
        AddMiddleware(&pipeline.LoggingMiddleware{logger: logger}).
        AddMiddleware(&pipeline.MetricsMiddleware{}).
        AddMiddleware(&pipeline.TimeoutMiddleware{timeout: 5 * time.Second})
    
    // 执行
    tlsData := readTLSData()
    result, err := pipeline.Execute(context.Background(), tlsData)
    
    if err != nil {
        log.Fatalf("Pipeline failed: %v", err)
    }
    
    fmt.Printf("Pipeline result: %v\n", result.Output)
}
```plaintext

### Step 4.2: 性能对比测试

```go
// test/pipeline_benchmark_test.go

func BenchmarkOldProcessingEngine(b *testing.B) {
    engine := setupOldEngine()
    for i := 0; i < b.N; i++ {
        engine.Process(createTestRequest())
    }
}

func BenchmarkNewPipeline(b *testing.B) {
    p := setupNewPipeline()
    for i := 0; i < b.N; i++ {
        p.Execute(context.Background(), createTestData())
    }
}

// 预期结果:
// BenchmarkOldProcessingEngine: 100 ns/op
// BenchmarkNewPipeline:         120 ns/op (20% 开销，可接受)
```plaintext

### Step 4.3: 功能验证

- ✅ ParseStage 能正确解析 TLS 扩展
- ✅ AnalyzeStage 能正确检测异常
- ✅ TransformStage 能正确转换格式
- ✅ 中间件能正确执行
- ✅ 依赖管理能检查循环依赖

---

## 📝 Day 10: 完成总结

### 文档输出:

1. [WEEK3-4-IMPLEMENTATION.md](#) - 实现细节
2. [WEEK3-4-REPORT.md](#) - 周报
3. [PIPELINE-USAGE-GUIDE.md](#) - 使用指南
4. [PERFORMANCE-ANALYSIS.md](#) - 性能分析

### 交付物:

```plaintext
新增文件:
├── internal/pipeline/
│   ├── pipeline.go (200 行)
│   ├── middleware.go (150 行)
│   ├── stages.go (200 行)
│   ├── pipeline_test.go (300 行)
│   └── examples/main.go (100 行)
├── test/pipeline_benchmark_test.go (100 行)
└── docs/5-process/PIPELINE-GUIDE.md (200 行)

总计: 1250 行代码, 100% 测试覆盖
```plaintext

---

## 成功标准

| 指标 | 目标 | 通过标准 |
| ------ | ------ | -------- |
| **代码行数** | 1000-1500 | ✓ |
| **测试覆盖** | > 90% | ✓ |
| **编译** | 0 错误 | ✓ |
| **性能开销** | < 10% | ✓ |
| **文档** | 完整 | ✓ |
| **PoC 验证** | 功能正确 | ✓ |

---

## 风险和缓解

| 风险 | 概率 | 缓解方案 |
| ------ | ------ | -------- |
| **依赖管理复杂** | 低 | 预先设计，充分测试 |
| **性能退化** | 低 | 基准测试监控 |
| **中间件链过复杂** | 中 | 限制中间件数量 |

---

## 下一步 (Week 5-6 预告)

Week 5-6 将开始**事件溯源**实现:

- 为 BehaviorAnalyzer 添加事件存储
- 实现事件重放机制
- 创建行为投影
- 完整的审计日志

---

**状态**: 📋 规划完成，准备执行  
**开始时间**: 今天(2026-03-02)  
**预计完成**: 2026-03-16
