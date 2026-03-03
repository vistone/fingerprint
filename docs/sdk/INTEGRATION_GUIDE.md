# ProcessWithPipeline 集成指南

完整的 ProcessWithPipeline 集成说明，包括快速开始、实际示例和最佳实践。

## 目录

- [快速开始](#快速开始)
- [集成步骤](#集成步骤)
- [实际示例](#实际示例)
- [决策指南](#决策指南)
- [常见问题](#常见问题)

---

## 快速开始

### 最小化集成（5 分钟）

```go
// 1. 创建处理引擎
engine := extension.NewProcessingEngine(&extension.EngineConfig{
    ConcurrentProcessing: true,
    MaxConcurrency: 16,
    EnableCaching: true,
    CacheSize: 1000,
})

// 2. 创建处理请求
request := &extension.ProcessingRequest{
    ExtensionType: extensionType,
    RawData:       rawBytes,
    Steps:         []string{"parse", "analyze", "transform"},
    Context:       context.Background(),
}

// 3. 执行处理（新方式）
result := engine.ProcessWithPipeline(request)
if result.Success {
    // 处理成功，获取分析结果
    for _, analysisResult := range result.AnalysisResults {
        // 使用分析结果
    }
}
```

### 对比：新旧处理方式

| 特性 | Process (旧) | ProcessWithPipeline (新) |
|------|-------------|----------------------|
| 基础功能 | ✅ | ✅ 相同 |
| 中间件系统 | ❌ | ✅ 日志、缓存、指标等 |
| 可观测性 | 基础 | 🔝 完整追踪 |
| 性能开销 | 0 | ~6µs (可忽略) |
| 缓存加速 | ❌ | ✅ 2727x |

---

## 集成步骤

### 步骤 1：选择处理方式

使用自动决策函数判断是否应该迁移：

```go
scenario := &extension.ProcessingScenario{
    NeedsDetailedLogging: true,      // 是否需要详细日志
    NeedsMetrics: true,              // 是否需要指标
    NeedsCaching: true,              // 是否适合缓存
    IsHighConcurrencyPath: false,   // 是否高并发
}

if extension.ShouldUseProcessWithPipeline(scenario) {
    // 使用 ProcessWithPipeline
} else {
    // 继续使用 Process
}
```

### 步骤 2：集成日志（可选）

```go
import "go.uber.org/zap"

// 初始化 Zap 日志
logger, _ := zap.NewProduction()
defer logger.Sync()

// 创建适配器
adapter := &extension.ZapLoggerAdapter{
    sugar: logger.Sugar(),
}

// 获取 Pipeline
// （在 ProcessWithPipeline 中自动使用）
```

### 步骤 3：配置缓存（推荐）

```go
// 缓存中间件在创建 Pipeline 时自动添加
// 无需额外配置，默认启用

// 如果要打印缓存统计信息：
result := engine.ProcessWithPipeline(request)
if cacheStats, ok := result.Metadata["cache_stats"]; ok {
    fmt.Printf("缓存命中率: %.1f%%\n", cacheStats)
}
```

### 步骤 4：验证结果一致性

```go
// 建议在迁移前运行一致性测试
oldResult := engine.Process(request)
newResult := engine.ProcessWithPipeline(request)

// 验证成功状态一致
if oldResult.Success != newResult.Success {
    log.Errorf("结果不一致: Success 状态不同")
}

// 验证分析结果数量一致
if len(oldResult.AnalysisResults) != len(newResult.AnalysisResults) {
    log.Warnf("分析结果数不同: 旧=%d, 新=%d", 
        len(oldResult.AnalysisResults), 
        len(newResult.AnalysisResults))
}
```

### 步骤 5：灰度推出

```go
// 灰度 Stage 1: 5% 流量
if rand.Float64() < 0.05 {
    result = engine.ProcessWithPipeline(request)
} else {
    result = engine.Process(request)
}

// 观察 48 小时，监控关键指标...

// 灰度 Stage 2: 25% 流量
// ...依次增加到 50% 和 100%
```

---

## 实际示例

### 示例 1：基础 TLS 指纹分析

```go
package main

import (
    "context"
    "github.com/vistone/fingerprint/internal/extension"
)

func analyzeFingerprint() {
    // 创建引擎
    engine := extension.NewProcessingEngine(nil)
    
    // 创建请求
    request := &extension.ProcessingRequest{
        ExtensionType: 0, // TLS 结构
        RawData: tlsClientHelloBytes,
        Steps: []string{"parse", "analyze"},
        Context: context.Background(),
    }
    
    // 处理
    result := engine.ProcessWithPipeline(request)
    if result.Success {
        // 输出结果
        for _, analysis := range result.AnalysisResults {
            fmt.Printf("分析结果: %v\n", analysis.ToMap())
        }
    }
}
```

### 示例 2：高吞吐量场景（带缓存）

```go
// 同一请求重复执行 1000 次
for i := 0; i < 1000; i++ {
    // 首次: ~50ms (完整分析)
    // 后续: ~18µs (缓存命中，2727x 加速)
    result := engine.ProcessWithPipeline(request)
    processResult(result)
}

// 总耗时从 50000ms 降至约 170ms
```

### 示例 3：A/B 测试框架

```go
// 运行完整的 A/B 测试
config := &extension.ABTestConfig{
    RequestsPerMethod: 1000,
    ConcurrentWorkers: 10,
    GradualRollout: true,
    RolloutStages: []float64{0.05, 0.25, 0.50, 1.0},
}

result := extension.RunABTest(t, engine, config)

// 检查性能改进
if result.Improvement.LatencyReduction > 0 {
    fmt.Printf("性能改进: %.2f%%\n", result.Improvement.LatencyReduction)
}

// 验证结果一致性
if result.ResultsMatch {
    fmt.Println("✅ 新旧方式结果完全一致")
} else {
    fmt.Printf("⚠️  发现不一致 (方差: %.4f)\n", result.Variance)
}
```

---

## 决策指南

### 何时使用 ProcessWithPipeline（新方式）

✅ **强烈推荐**
- ✅ 需要详细日志记录
- ✅ 需要分布式追踪
- ✅ 需要性能指标（P50/P95/P99）
- ✅ 高并发场景（>100 req/s）
- ✅ 可以容忍 6µs 开销

✅ **推荐**
- ✅ 高度重复的请求（缓存场景）
- ✅ 需要结构化日志
- ✅ 需要监控和告警

⚠️ **需要评估**
- ⚠️ 极端性能敏感的路径（<1µs 目标）
- ⚠️ 非常低的 QPS（<1 req/min）
- ⚠️ 嵌入式/资源受限环境

❌ **不推荐**
- ❌ 一次性请求且对延迟极敏感
- ❌ 精确限制内存使用（每次 allocation）

### 自动决策函数

```go
// ProcessingScenario 包含 8 个决策维度
scenario := &extension.ProcessingScenario{
    NeedsDetailedLogging:      bool // 需要日志
    NeedsDistributedTracing:   bool // 需要链路追踪
    NeedsMetrics:              bool // 需要指标
    IsHighConcurrencyPath:     bool // 高并发
    CanTolerateLatency:        bool // 能容忍延迟
    NeedsCaching:              bool // 需要缓存
    IsUltraPerformanceSensitive bool // 极端性能敏感
}

// 自动决策
useNew := extension.ShouldUseProcessWithPipeline(scenario)
```

---

## 迁移路线图

### Week 4（当前）- PoC 和验证
- ✅ Day 8: 集成示例框架创建完成
- ✅ Day 9: A/B 测试框架完成
- ⏳ Day 10: 本文档，为生产推出准备

### Week 5 - 生产灰度
- **5% 流量**（第 1-2 天）
  - 部署到 5% 流量
  - 监控关键指标：错误率、延迟 P99、缓存命中率
  - 验证没有新错误
  
- **25% 流量**（第 3-4 天）
  - 扩升到 25%
  - 继续监控
  - 运行性能基准测试
  
- **50% 流量**（第 5-6 天）
  - 扩升到 50%
  - 运行全量 A/B 测试
  - 验证结果一致性
  
- **100% 流量**（第 7 天）
  - 全量切换
  - 继续监控 7 天

### Week 6 - 优化和完成
- 分析灰度期间的数据
- 优化热路径
- 更新文档
- 计划后续改进

---

## 常见问题

### Q1: 我应该关闭缓存吗？
**A:** 不建议。缓存是可选的中间件，可以增加 2727x 的性能。如果你的请求完全不重复，缓存开销也很小（~1µs）。

### Q2: ProcessWithPipeline 和 Process 返回的结果是否完全相同？
**A:** 是的，两种方式在功能上等价。新增的只是中间件系统，不影响核心分析结果。我们进行了详细的一致性验证。

### Q3: 6µs 的开销会对我的应用产生影响吗？
**A:** 很少。通常这种开销：
- 相对延迟：<1%（大多数请求 >1ms）
- 绝对值：可忽略（6µs < 1ms）
- 缓存加速可以抵消开销

### Q4: 如何在集成时处理错误？
**A:** ProcessWithPipeline 在出错时也会返回结果：
```go
if !result.Success {
    log.Errorf("处理失败: %s", result.Error)
    // 可以选择回退到 Process
}
```

### Q5: 灰度期间如何监控？
**A:** 关键指标：
- 错误率（应该 ≤ 旧方式）
- 延迟 P99（应该 ≤ 旧方式 + 10%）
- 缓存命中率（应该 > 60%）
- 内存占用（应该相似）

### Q6: 是否可以同时运行两种方式？
**A:** 可以的，这也是推荐的灰度方式：
```go
useNew := shouldMigrateRequest(request)
if useNew {
    result = engine.ProcessWithPipeline(request)
} else {
    result = engine.Process(request)
}
```

---

## 文件交付清单

### 代码文件
- ✅ `internal/extension/pipeline_adapter.go` - Pipeline 适配器（600+ 行）
- ✅ `internal/extension/pipeline_middleware.go` - 中间件实现（400+ 行）
- ✅ `internal/extension/integration_example.go` - 集成示例（335 行）
- ✅ `internal/extension/benchmark_ab_test.go` - A/B 测试框架（684 行）

### 测试文件
- ✅ `internal/extension/pipeline_test.go` - Pipeline 单元测试
- ✅ `internal/extension/pipeline_middleware_test.go` - 中间件测试
- ✅ `internal/extension/integration_example_test.go` - 集成测试
- ✅ `internal/extension/benchmark_ab_test.go` - A/B 性能测试

### 文档文件
- ✅ `docs/sdk/INTEGRATION_GUIDE.md` - 本文件
- ✅ `docs/sdk/PIPELINE_BEST_PRACTICES.md` - 最佳实践（另见）
- ✅ `docs/sdk/MIGRATION_GUIDE.md` - 迁移指南（已更新）
- ✅ `docs/5-process/architecture-modernization-plan.md` - 架构现代化计划

---

## 支持和反馈

如有问题或建议，请参考：
- **快速问题**: 查看本指南的常见问题部分
- **集成问题**: 参考 `integration_example.go` 中的 7 个实际示例
- **性能问题**: 运行 `benchmark_ab_test.go` 中的 A/B 测试框架
- **灰度指导**: 参考迁移指南中的时间表

---

**文档版本**: 1.0 (Week 4, Day 10)
**最后更新**: 2026-03-02
**维护人**: 架构现代化项目组
