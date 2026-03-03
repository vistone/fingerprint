# 灰度推出框架 (Canary Framework) 📊

**生产级灰度推出管理框架 - ProcessWithPipeline** Pipeline 架构的安全推出保障

## 📋 概览

灰度推出框架为 ProcessWithPipeline 流水线架构提供了完整的灰度推出能力，包括：

- ✅ **灰度流量分配** - 基于一致性哈希的流量分配，确保同一请求总是使用同一方式
- ✅ **智能指标收集** - 实时收集业务、性能、系统指标
- ✅ **健康检查** - 自动监控灰度指标，及时发现问题
- ✅ **生成灰度报告** - 数据驱动的灰度评估和决策
- ✅ **生命周期管理** - 四阶段灰度推出的完整支持 (5% → 25% → 50% → 100%)

## 🎯 核心概念

### 灰度阶段 (Canary Stage)

```go
const (
    CanaryStage5Percent   CanaryStage = "5%"
    CanaryStage25Percent  CanaryStage = "25%"
    CanaryStage50Percent  CanaryStage = "50%"
    CanaryStage100Percent CanaryStage = "100%"
)
```

四阶段灰度策略：
1. **5% 灰度** (Day 1-2) - 基础功能验证，确保没有明显问题
2. **25% 灰度** (Day 3-4) - 小范围 A/B 测试，验证性能对标
3. **50% 灰度** (Day 5-6) - 大范围对称性验证，确保结果一致性
4. **100% 灰度** (Day 7+) - 全量切换，完整用户流量

### 一致性流量分配

关键特性：**同一请求 ID 总是被路由到同一处理方式**

```go
// 基于 requestID 的一致性哈希
hash := hashRequestID(requestID)
threshold := uint32(float64(^uint32(0)) * targetPercentage)
useNewMethod := hash < threshold
```

优势：
- 用户体验一致（不会随机切换处理方式）
- 易于诊断问题（能清晰区分新旧处理的请求）
- 支持渐进式的流量增加

## 🚀 快速开始

### 1. 初始化灰度系统

```go
import (
    "github.com/vistone/fingerprint/internal/extension"
)

// 创建灰度配置
config := &extension.CanaryConfig{
    Stage:            extension.CanaryStage5Percent,
    TargetPercentage: 0.05,  // 5% 流量
    Enabled:          false, // 初始禁用
}

// 创建灰度指标收集器
collector := extension.NewCanaryMetricsCollector(config)

// 创建灰度健康检查
healthCheck := extension.NewCanaryHealthCheck(collector)
```

### 2. 启用灰度和路由流量

```go
// 启用 5% 灰度
duration := 48 * time.Hour
collector.EnableCanary(extension.CanaryStage5Percent, 0.05, duration)

// 在请求处理中决定使用哪个方式
requestID := "user-123"
if collector.router.ShouldUseNewMethod(requestID) {
    // 使用新方式 (ProcessWithPipeline)
    result = engine.ProcessWithPipeline(ctx, request)
} else {
    // 使用旧方式 (ProcessWithoutPipeline)
    result = engine.Process(ctx, request)
}

// 记录请求指标
success := result.Error == nil
collector.RecordRequest(requestID, useNewMethod, latency, success)

// 记录缓存
if hitCache {
    collector.RecordCacheHit(true)
}
```

### 3. 监控和评估

```go
// 定期检查健康状况 (每 1 小时)
healthy, alerts, critique := healthCheck.CheckHealth(context.Background())
if !healthy {
    fmt.Printf("⚠️  灰度警告:\n")
    for _, alert := range alerts {
        fmt.Printf("  %s\n", alert)
    }
}

// 如果发现严重问题，立即回滚
if critique != "" && strings.Contains(critique, "❌") {
    collector.DisableCanary()
    // 触发告警和日志
}

// 定期保存指标快照 (每 15 分钟)
collector.SnapshotMetrics()
```

### 4. 生成灰度报告和决策

```go
// 在阶段结束时生成报告
report := collector.GenerateCanaryReport(startTime)

// 打印报告
report.Print()

// 根据建议决定升级或回滚
// report.Recommendation 可能是:
// "✅ 可以升级到下一阶段"
// "⚠️  建议延缓灰度"
// "❌ 立即回滚"
```

## 📊 灰度指标体系

### Tier 1: 业务关键指标 (必须监控)

```go
type CanaryMetrics struct {
    // 请求统计
    TotalRequests      int64  // 总请求数
    NewMethodRequests  int64  // 新方式处理的请求数
    OldMethodRequests  int64  // 旧方式处理的请求数
    
    // 成功率
    SuccessfulRequests int64
    FailedRequests     int64
    ErrorRate          float64  // = FailedRequests / TotalRequests
    
    // 一致性 (新旧方式结果是否相同)
    Consistency        float64  // 0-1
    
    // 缓存效果
    CacheHits          int64
    CacheMisses        int64
    CacheHitRate       float64  // = CacheHits / (CacheHits + CacheMisses)
}
```

**告警阈值**:
- 错误率 > 1% → ⚠️  警告，继续观察
- 错误率 > 3% → 🔴 严重，立即回滚
- 缓存命中率 < 50% → ⚠️  警告，检查缓存策略

### Tier 2: 性能指标

```go
type CanaryMetrics struct {
    // 延迟
    TotalLatency   time.Duration
    MinLatency     time.Duration
    MaxLatency     time.Duration
    AvgLatency     time.Duration
    P50Latency     time.Duration
    P95Latency     time.Duration
    P99Latency     time.Duration
}
```

**基准值** (ProcessWithPipeline vs Process):
- P50 延迟允许 +20%
- P95 延迟允许 +10%
- P99 延迟告警 > 150ms

### Tier 3: 系统指标 (参考)

```go
// 这些指标由监控系统单独采集
MemoryUsage    int64  // 内存使用
GCTime         time.Duration  // GC 耗时
// ... 其他基础设施指标
```

## 🛡️ 健康检查和自动决策

### 健康检查分三个层级

```go
healthCheck := extension.NewCanaryHealthCheck(collector)
healthy, alerts, critique := healthCheck.CheckHealth(ctx)

if !healthy {
    // 说明有问题，建议继续观察或采取行动
    for _, alert := range alerts {
        log.Warnf("%s", alert)  // ⚠️  
    }
}

if strings.Contains(critique, "❌") {
    // 严重问题，需要立即回滚
    log.Errorf("%s", critique)  // ❌
    collector.DisableCanary()
}
```

### 自动决策树

```
灰度开始
  ↓
定期监控 (每 1 小时)
  ├─ 错误率 > 3% ? → [立即回滚] ❌
  ├─ P99 > 500ms ? → [立即回滚] ❌
  ├─ OOM/Panic ? → [立即回滚] ❌
  └─ 全部正常 ?
      └─ 成功率 > 99.5% ?
          ├─ 是 → [保存快照，继续观察]
          └─ 否 → [保存快照，预警继续观察]
```

## 📈 生命周期管理

### 完整的四阶段推出

```go
lifecycle := extension.NewCanaryLifecycle()

// Week 5, Day 1-2: 5% 灰度
lifecycle.StartStage(extension.CanaryStage5Percent, 48*time.Hour)
// ... 监控和等待
report5 := lifecycle.EndStage()
report5.Print()

// Week 5, Day 3-4: 25% 灰度
lifecycle.StartStage(extension.CanaryStage25Percent, 48*time.Hour)
// ... 监控和等待
report25 := lifecycle.EndStage()

// Week 5, Day 5-6: 50% 灰度
lifecycle.StartStage(extension.CanaryStage50Percent, 48*time.Hour)
// ... 监控和等待
report50 := lifecycle.EndStage()

// Week 5, Day 7: 100% 灰度
lifecycle.StartStage(extension.CanaryStage100Percent, 0)  // 无时间限制
// ... 继续监控
// 本阶段成为新的默认处理方式
```

## 🧪 测试和验证

### 单元测试

```bash
# 运行所有灰度框架测试
go test ./internal/extension -v -run "^TestCanary" -timeout 30s

# 输出示例
# === RUN   TestCanaryRouter
# ✅ 流量分配正确: 5.00% (期望: 5.00%)
# === RUN   TestCanaryMetricsCollection
# ✅ 指标收集正确:
#   总请求: 1000
#   新方式: 57 (5.7%)
#   成功率: 99.20%
#   缓存命中: 56.0%
```

### 性能基准测试

```bash
# 灰度路由器性能
go test ./internal/extension -bench "CanaryRouter" -benchmem

# 输出
# BenchmarkCanaryRouter-56: 353.6 ns/op (非常快)
# BenchmarkCanaryMetricsCollection-56: 424.1 ns/op (可接受)
```

### 灰度流量模拟

```go
// 在测试中模拟灰度流量
func TestGradualRollout(t *testing.T) {
    config := &extension.CanaryConfig{
        Stage:            extension.CanaryStage5Percent,
        TargetPercentage: 0.05,
        Enabled:          true,
    }
    collector := extension.NewCanaryMetricsCollector(config)
    
    // 模拟 10000 个请求
    for i := 0; i < 10000; i++ {
        shouldUseNew := collector.router.ShouldUseNewMethod("req-" + strconv.Itoa(i))
        // ... 处理请求
        collector.RecordRequest("req-"+strconv.Itoa(i), shouldUseNew, latency, success)
    }
    
    // 验证流量分配
    metrics := collector.GetCurrentMetrics()
    newRatio := float64(metrics.NewMethodRequests) / float64(metrics.TotalRequests)
    if newRatio < 0.04 || newRatio > 0.06 {
        t.Errorf("流量分配偏离目标 5%%")
    }
}
```

## 📚 完整示例

### 实时灰度监控

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/vistone/fingerprint/internal/extension"
)

func main() {
    // 创建灰度系统
    config := &extension.CanaryConfig{
        Stage:            extension.CanaryStage5Percent,
        TargetPercentage: 0.05,
        Enabled:          false,
    }
    collector := extension.NewCanaryMetricsCollector(config)
    healthCheck := extension.NewCanaryHealthCheck(collector)
    
    // 启用 5% 灰度
    collector.EnableCanary(
        extension.CanaryStage5Percent,
        0.05,
        48*time.Hour,
    )
    
    // 模拟请求处理
    go func() {
        for i := 0; i < 1000; i++ {
            requestID := fmt.Sprintf("req-%d", i)
            useNew := collector.router.ShouldUseNewMethod(requestID)
            
            // 模拟处理
            start := time.Now()
            success := true  // 假设都成功
            latency := time.Duration(time.Now().UnixNano() - start.UnixNano())
            
            // 记录指标
            collector.RecordRequest(requestID, useNew, latency, success)
            
            if i%50 == 0 {
                collector.RecordCacheHit(i%2 == 0)
            }
            
            time.Sleep(10 * time.Millisecond)
        }
    }()
    
    // 定期监控
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        // 检查健康状况
        healthy, alerts, critique := healthCheck.CheckHealth(context.Background())
        
        if !healthy {
            fmt.Printf("⚠️  警告: %v\n", alerts)
        }
        
        // 获取当前指标
        metrics := collector.GetCurrentMetrics()
        fmt.Printf("[%s] 请求: %d, 新方式: %.1f%%, 成功率: %.2f%%, 缓存命中: %.1f%%\n",
            time.Now().Format("15:04:05"),
            metrics.TotalRequests,
            float64(metrics.NewMethodRequests)/float64(metrics.TotalRequests)*100,
            float64(metrics.SuccessfulRequests)/float64(metrics.TotalRequests)*100,
            metrics.CacheHitRate*100,
        )
        
        if metrics.TotalRequests >= 500 {
            break
        }
    }
    
    // 生成最终报告
    report := collector.GenerateCanaryReport(time.Now().Add(-48*time.Hour))
    report.Print()
}
```

## ⚙️ 配置参考

### CanaryThresholds - 告警和回滚阈值

```go
type CanaryThresholds struct {
    // 告警阈值 (严格监控)
    MaxErrorRate        float64        // 1% - 警告阈值
    MaxP99Latency       time.Duration  // 150ms - 警告阈值
    MinCacheHitRate     float64        // 50% - 警告阈值
    MaxMemoryGrowth     int64          // 100KB/min - 警告阈值

    // 回滚触发阈值 (紧急停止)
    CriticalErrorRate   float64        // 3% - 立即回滚
    CriticalP99Latency  time.Duration  // 500ms - 立即回滚
}

// 使用默认值
thresholds := extension.DefaultCanaryThresholds()

// 自定义阈值
thresholds.MaxErrorRate = 0.02  // 改为 2%
thresholds.CriticalErrorRate = 0.05  // 改为 5%
```

## 📖 常见问题

### Q1: 灰度框架的性能开销是多少?

灰度路由决策: **353.6 ns/op** (非常轻)
灰度指标收集: **424.1 ns/op** (可接受)

总体开销 < 1μs，在毫秒级别的请求处理中可忽略不计。

### Q2: 如何确保一致性路由?

框架使用一致性哈希 (consistent hashing)：

```go
hash := hashRequestID(requestID)
threshold := uint32(float64(^uint32(0)) * targetPercentage)
useNewMethod := hash < threshold
```

同一 requestID 的哈希值固定，因此路由决定是确定性的。

### Q3: 是否支持动态调整灰度比例?

可以，但需要在监控期间进行：

```go
// 从 5% 升级到 25%
collector.config.TargetPercentage = 0.25
collector.config.Stage = extension.CanaryStage25Percent
```

但建议通过完整的 4 个阶段逐步推出，而不是随意调整。

### Q4: 如何处理灰度期间的问题?

自动回滚触发：
1. 错误率 > 3% → 立即禁用新方式
2. 手动回滚：`collector.DisableCanary()`
3. 恢复时重新启用：`collector.EnableCanary(...)`

### Q5: 灰度报告中的"建议"字段基于什么?

```go
if errorRate > 3% {
    recommendation = "❌ 立即回滚"
} else if errorRate > 1% {
    recommendation = "⚠️  建议延缓灰度"
} else if successRate < 99% {
    recommendation = "✅ 灰度进行良好，继续监控"
} else {
    recommendation = "✅ 可以升级到下一阶段"
}
```

## 🎓 深度学习

### 相关文档
- [Pipeline 最佳实践](./pipeline-best-practices.md)
- [集成指南](./integration-guide.md)
- [Week 5-6 灰度推出实施指南](../5-process/week5-6-execution-guide.md)

### 测试代码
- [灰度框架测试](../../internal/extension/canary_framework_test.go)
- [A/B 测试框架](../../internal/extension/benchmark_ab_test.go)

## 📞 支持和反馈

如果发现问题或有改进建议，请提交 Issue 或联系项目团队。

---

**最后更新**: 2026-03-02  
**版本**: 1.0.0  
**状态**: 生产就绪 ✅
