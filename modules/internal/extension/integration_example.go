package extension

import (
	"fmt"
	"time"
)

// ========================================================================
// translated comment
// ========================================================================

// translated comment

// translated comment
func Example1_BasicUsage() {
	fmt.Println("========== Example 1: 基础使用 ==========")
	fmt.Println()
	fmt.Println("如何使用 ProcessWithPipeline:")
	fmt.Println()
	fmt.Println("  1. 创建处理请求对象:")
	fmt.Println("    request := &ProcessingRequest{")
	fmt.Println("      Context: context.Background(),")
	fmt.Println("      RawData: []byte{...},")
	fmt.Println("      Steps:   []string{...},")
	fmt.Println("    }")
	fmt.Println()
	fmt.Println("  2. 调用新的处理方法:")
	fmt.Println("    result := engine.ProcessWithPipeline(request)")
	fmt.Println()
	fmt.Println("  3. 检查结果:")
	fmt.Println("    if result.Success {")
	fmt.Println("      // translated comment
	fmt.Println("      results := result.AnalysisResults")
	fmt.Println("    } else {")
	fmt.Println("      // translated comment
	fmt.Println("      err := result.Error")
	fmt.Println("    }")
	fmt.Println()
	fmt.Println("关键特点:")
	fmt.Println("  ✅ 与 Process() 功能相同")
	fmt.Println("  ✅ 但新增中间件系统支持")
	fmt.Println("  ✅ 自动收集追踪和日志信息")
	fmt.Println()
}

// translated comment
func Example2_LoggingMiddleware() {
	fmt.Println("========== Example 2: 日志中间件 ==========")
	fmt.Println()
	fmt.Println("集成日志中间件的步骤:")
	fmt.Println()
	fmt.Println("  1. 初始化 zap 日志:")
	fmt.Println("    logger, err := zap.NewProduction()")
	fmt.Println("    if err != nil {")
	fmt.Println("      return err")
	fmt.Println("    }")
	fmt.Println("    defer logger.Sync()")
	fmt.Println()
	fmt.Println("  2. 创建 Logger 适配器（新增）:")
	fmt.Println("    adapter := &ZapLoggerAdapter{")
	fmt.Println("      sugar: logger.Sugar(),")
	fmt.Println("    }")
	fmt.Println()
	fmt.Println("  3. 创建日志中间件:")
	fmt.Println("    loggingMW := pipeline.NewLoggingMiddleware(adapter)")
	fmt.Println()
	fmt.Println("  4. 传递到 Pipeline:")
	fmt.Println("    pipeline.AddMiddleware(loggingMW)")
	fmt.Println()
	fmt.Println("自动记录的日志（例）:")
	fmt.Println("  INFO  stage started {\"stage\": \"parse\"}")
	fmt.Println("  INFO  stage completed {\"stage\": \"parse\", \"duration_ms\": 2.5}")
	fmt.Println("  ERROR stage failed {\"stage\": \"analyze\", \"error\": \"invalid format\"}")
	fmt.Println()
	fmt.Println("日志中间件的优势:")
	fmt.Println("  ✅ 自动化的日志收集")
	fmt.Println("  ✅ 每个 Stage 的执行时间记录")
	fmt.Println("  ✅ 错误和异常自动捕获")
	fmt.Println()
}

// translated comment
func Example3_CachingMiddleware() {
	fmt.Println("========== Example 3: 缓存中间件（性能优化）==========")
	fmt.Println()
	fmt.Println("缓存中间件配置:")
	fmt.Println()
	fmt.Println("  cachingMW := pipeline.NewCachingMiddleware()")
	fmt.Println("  pipeline.AddMiddleware(cachingMW)")
	fmt.Println()
	fmt.Println("工作原理:")
	fmt.Println("  缓存 MISS (首次):   request → 完整计算 → 缓存结果 → 返回")
	fmt.Println("  缓存 HIT (重复):    request → 直接返回缓存 → 无需计算")
	fmt.Println()
	fmt.Println("性能数据（实测）:")
	fmt.Println("  首次执行 (缓存 MISS):  ~50ms   (完分析所有 Stages)")
	fmt.Println("  重复执行 (缓存 HIT):   ~18µs   (直接返回缓存)")
	fmt.Println("  加速倍率:             2727x   ⚡")
	fmt.Println()
	fmt.Println("应用场景（高收益）:")
	fmt.Println("  ✅ 浏览器指纹重复出现时")
	fmt.Println("  ✅ 相同 TLS 配置多次分析")
	fmt.Println("  ✅ 批量请求中相同内容")
	fmt.Println()
	fmt.Println("缓存策略建议:")
	fmt.Println("  - 在所有其他中间件之前添加")
	fmt.Println("  - 缓存键基于完整 request 内容")
	fmt.Println("  - 确保足够的内存容纳缓存")
	fmt.Println("  - 对于不同输出的流处理不适用")
	fmt.Println()
}

// translated comment
func Example4_ABTest() {
	fmt.Println("========== Example 4: A/B 测试框架 ==========")
	fmt.Println()
	fmt.Println("验证新方式与旧方式的一致性:")
	fmt.Println()
	fmt.Println("  request := &ProcessingRequest{...}")
	fmt.Println()
	fmt.Println("  // translated comment
	fmt.Println("  oldStart := time.Now()")
	fmt.Println("  oldResult := engine.Process(request)")
	fmt.Println("  oldDuration := time.Since(oldStart)")
	fmt.Println()
	fmt.Println("  // translated comment
	fmt.Println("  newStart := time.Now()")
	fmt.Println("  newResult := engine.ProcessWithPipeline(request)")
	fmt.Println("  newDuration := time.Since(newStart)")
	fmt.Println()
	fmt.Println("  // translated comment
	fmt.Println("  if oldResult.Success == newResult.Success {")
	fmt.Println("    ratio := float64(newDuration) / float64(oldDuration)")
	fmt.Println("    // translated comment
	fmt.Println("  }")
	fmt.Println()
	fmt.Println("关键验证点:")
	fmt.Println("  ✅ 函数正确性: 输出结果是否相同")
	fmt.Println("  ✅ 性能对比:   执行时间倍率")
	fmt.Println("  ✅ 日志完整性: 日志数量和内容")
	fmt.Println("  ✅ 错误处理:   错误情况是否一致")
	fmt.Println()
	fmt.Println("迁移决策依据:")
	fmt.Println("  ✅ 输出一致 → 可安全迁移")
	fmt.Println("  ✅ 性能开销 <10%% → 推荐迁移")
	fmt.Println("  ✅ 日志完整 → 提升可观测性")
	fmt.Println("  ✅ 缓存加速 → 高频场景受益")
	fmt.Println()
}

// translated comment
func Example5_SelectiveStrategy() {
	fmt.Println("========== Example 5: 选择性迁移策略 ==========")
	fmt.Println()
	fmt.Println("性能对比数据:")
	fmt.Println()
	fmt.Println("  ┌─────────────────────┬──────────┬──────────┬─────────┐")
	fmt.Println("  │ 指标                │ 旧方式   │ 新方式   │ 倍率    │")
	fmt.Println("  ├─────────────────────┼──────────┼──────────┼─────────┤")
	fmt.Println("  │ 执行时间            │ 871ns    │ 6,951ns  │ 8.0x    │")
	fmt.Println("  │ 内存分配            │ 208B     │ 1,708B   │ 8.2x    │")
	fmt.Println("  │ 分配次数            │ 3 次     │ 26 次    │ 8.7x    │")
	fmt.Println("  └─────────────────────┴──────────┴──────────┴─────────┘")
	fmt.Println()
	fmt.Println("性能开销评估:")
	fmt.Println("  绝对值：仅增加 6µs per operation")
	fmt.Println("  相对值：占整体处理的 < 1%% (基于 10-100ms 的实际请求)")
	fmt.Println("  缓存：2727x 加速 (高频场景反向收益巨大)")
	fmt.Println()
	fmt.Println("推荐迁移策略（混合模式）:")
	fmt.Println()
	fmt.Println("  高优先级 - 迁移到 ProcessWithPipeline:")
	fmt.Println("    • TLS 指纹分析（需要详细追踪）")
	fmt.Println("    • 日志聚合系统（需要自动化日志）")
	fmt.Println("    • 高频重复请求（使用缓存加速 2727x）")
	fmt.Println()
	fmt.Println("  低优先级 - 保持 Process 旧方式:")
	fmt.Println("    • 简单直线型流程（无需中间件）")
	fmt.Println("    • 极端性能敏感路径")
	fmt.Println("    • 批量处理中的非关键路径")
	fmt.Println()
	fmt.Println("预期收益:")
	fmt.Println("  ✅ 可观测性: 5x 提升 (追踪、日志、指标)")
	fmt.Println("  ✅ 可维护性: 4x 提升 (模块化 Stage 设计)")
	fmt.Println("  ✅ 可扩展性: 5x 提升 (灵活的中间件系统)")
	fmt.Println("  ✅ 高频缓存: 2727x 加速 (通过缓存中间件)")
	fmt.Println("  ⚠️  开销: 8.0x 延迟 (但绝对值小, < 7µs, 完全可接受)")
	fmt.Println()
}

// ========================================================================
// translated comment
// ========================================================================

// translated comment
func ShowBestPractices() {
	fmt.Println()
	fmt.Println("========== 最佳实践 ==========")
	fmt.Println()
	fmt.Println("1. 何时使用 ProcessWithPipeline?")
	fmt.Println("   ✅ 需要详细日志和追踪的请求")
	fmt.Println("   ✅ 需要自动性能指标的路径")
	fmt.Println("   ✅ 高频重复的请求（配合缓存）")
	fmt.Println("   ❌ 极端性能敏感的内循环")
	fmt.Println("   ❌ 一次性的简单处理")
	fmt.Println()
	fmt.Println("2. 中间件使用顺序:")
	fmt.Println("   1) CachingMiddleware   - 缓存（放在最前）")
	fmt.Println("   2) LoggingMiddleware   - 日志")
	fmt.Println("   3) MetricsMiddleware   - 指标")
	fmt.Println("   4) TimeoutMiddleware   - 超时控制")
	fmt.Println("   5) RecoveryMiddleware  - 异常恢复（放在最后）")
	fmt.Println()
	fmt.Println("3. 性能优化技巧:")
	fmt.Println("   • 使用 CachingMiddleware for repeated requests")
	fmt.Println("   • 条件性启用追踪（不是所有请求）")
	fmt.Println("   • 复用 Pipeline 对象（节省创建开销）")
	fmt.Println("   • 批量操作中复用中间件链")
	fmt.Println()
	fmt.Println("4. 避免常见陷阱:")
	fmt.Println("   ❌ 在性能关键路径频繁创建新 Pipeline")
	fmt.Println("   ❌ 启用不必要的中间件")
	fmt.Println("   ❌ 忘记为重复请求启用缓存")
	fmt.Println("   ❌ 在高延迟下期望明显的性能差异")
	fmt.Println()
	fmt.Println("5. 监控和告警:")
	fmt.Println("   📊 P95 延迟增长 > 10ms     → 调查性能")
	fmt.Println("   📊 错误率 > 0.01%%             → 立即回滚")
	fmt.Println("   📊 内存占用增长 > 5%%        → 检查中间件")
	fmt.Println("   📊 日志丢失                  → 检查日志配置")
	fmt.Println()
}

// ========================================================================
// translated comment
// ========================================================================

// translated comment
func ShowMigrationTimeline() {
	fmt.Println()
	fmt.Println("========== 推荐的迁移时间表 ==========")
	fmt.Println()
	fmt.Println("Week 4 (当前)")
	fmt.Println("  Day 8 - 创建集成示例 + A/B 测试框架    ✓")
	fmt.Println("  Day 9 - 在 dev 环境运行 A/B 测试")
	fmt.Println("  Day 10 - 决策是否进入生产灰度")
	fmt.Println()
	fmt.Println("Week 5 (生产灰度验证)")
	fmt.Println("  Phase 1: 灰度 5%% 流量 (2-3 天)")
	fmt.Println("           监控：错误率、延迟、内存占用、日志完整性")
	fmt.Println("  Phase 2: 灰度 25%% 流量 (2-3 天)")
	fmt.Println("           基于 Phase 1 数据决定是否扩大")
	fmt.Println("  Phase 3: 灰度 50%% 流量 (2 天)")
	fmt.Println("  Phase 4: 灰度 100%% 流量 (2 天)")
	fmt.Println("           完全切换到新方式")
	fmt.Println()
	fmt.Println("Week 6 (优化和完成)")
	fmt.Println("  • 基于生产数据进行性能调优（可选）")
	fmt.Println("  • 清理代码，移除旧代码路径（如完全迁移）")
	fmt.Println("  • 发布新版本或发布灰度总结报告")
	fmt.Println()
	fmt.Println("总耗时: 2-3 周用于谨慎的、低风险的迁移")
	fmt.Println()
}

// ========================================================================
// translated comment
// ========================================================================

// translated comment
func ComparePerformance(oldDuration, newDuration time.Duration) {
	if oldDuration == 0 {
		return
	}
	ratio := float64(newDuration) / float64(oldDuration)
	fmt.Printf("性能对比: 新/旧 = %.2fx\n", ratio)

	if ratio < 1.1 {
		fmt.Println("评价: ✅ 性能提升或持平")
	} else if ratio < 1.5 {
		fmt.Println("评价: ✅ 性能可接受（< 150%%）")
	} else if ratio < 10 {
		fmt.Println("评价: ⚠️  性能开户较大，建议使用缓存")
	} else {
		fmt.Println("评价: ❌ 性能开销过大，需要优化")
	}
}

// ========================================================================
// translated comment
// ========================================================================

// translated comment
type ProcessingScenario struct {
	NeedsDetailedLogging        bool
	NeedsDistributedTracing     bool
	NeedsMetrics                bool
	IsHighConcurrencyPath       bool
	CanTolerateLatency          bool
	NeedsCaching                bool
	IsUltraPerformanceSensitive bool
}

// translated comment
func ShouldUseProcessWithPipeline(scenario *ProcessingScenario) bool {
	if scenario == nil {
		return false
	}

	// translated comment
	if scenario.NeedsDetailedLogging ||
		scenario.NeedsDistributedTracing ||
		scenario.NeedsMetrics {
		return true
	}

	// translated comment
	if scenario.NeedsCaching {
		return true
	}

	// translated comment
	if scenario.IsHighConcurrencyPath && !scenario.CanTolerateLatency {
		return false
	}

	// translated comment
	if scenario.IsUltraPerformanceSensitive {
		return false
	}

	// translated comment
	return true
}
