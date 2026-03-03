package extension

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ========================================================================
// 灰度推出框架：ProcessWithPipeline 生产环境灰度管理
// ========================================================================

// CanaryStage 灰度阶段
type CanaryStage string

const (
	CanaryStage5Percent   CanaryStage = "5%"
	CanaryStage25Percent  CanaryStage = "25%"
	CanaryStage50Percent  CanaryStage = "50%"
	CanaryStage100Percent CanaryStage = "100%"
)

// CanaryConfig 灰度配置
type CanaryConfig struct {
	// 灰度阶段
	Stage CanaryStage

	// 目标流量百分比
	TargetPercentage float64

	// 启用灰度
	Enabled bool

	// 灰度开始时间
	StartTime time.Time

	// 灰度时长 (如果为 0 则不限制)
	Duration time.Duration
}

// CanaryMetrics 灰度指标
type CanaryMetrics struct {
	// 基础指标
	TotalRequests      int64
	NewMethodRequests  int64
	OldMethodRequests  int64
	SuccessfulRequests int64
	FailedRequests     int64

	// 延迟指标
	TotalLatency time.Duration
	MinLatency   time.Duration
	MaxLatency   time.Duration
	AvgLatency   time.Duration
	P50Latency   time.Duration
	P95Latency   time.Duration
	P99Latency   time.Duration

	// 缓存指标
	CacheHits    int64
	CacheMisses  int64
	CacheHitRate float64

	// 错误指标
	ErrorCount int64
	ErrorRate  float64

	// 资源指标
	MemoryUsage int64
	GCTime      time.Duration

	// 时间戳
	CollectedAt time.Time
}

// CanaryMetricsCollector 灰度指标收集器
type CanaryMetricsCollector struct {
	mu sync.RWMutex

	// 当前配置
	config *CanaryConfig

	// 指标数据
	metrics *CanaryMetrics

	// 历史数据 (用于分析趋势)
	history []*CanaryMetrics

	// 流量分配器
	router *CanaryRouter
}

// CanaryRouter 灰度流量路由器
type CanaryRouter struct {
	config *CanaryConfig
}

// ShouldUseNewMethod 判断是否使用新方式
func (r *CanaryRouter) ShouldUseNewMethod(requestID string) bool {
	if !r.config.Enabled {
		return false
	}

	// 基于 requestID 的 hash 进行一致性分配
	hash := hashRequestID(requestID)
	// 计算阈值: 百分比 * 最大 uint32
	threshold := uint32(float64(^uint32(0)) * r.config.TargetPercentage)
	return hash < threshold
}

// NewCanaryMetricsCollector 创建灰度指标收集器
func NewCanaryMetricsCollector(config *CanaryConfig) *CanaryMetricsCollector {
	if config == nil {
		config = &CanaryConfig{
			Stage:            CanaryStage5Percent,
			TargetPercentage: 0.05,
			Enabled:          false,
		}
	}

	return &CanaryMetricsCollector{
		config: config,
		metrics: &CanaryMetrics{
			CollectedAt: time.Now(),
		},
		history: make([]*CanaryMetrics, 0, 1000),
		router:  &CanaryRouter{config: config},
	}
}

// EnableCanary 启用灰度
func (c *CanaryMetricsCollector) EnableCanary(stage CanaryStage, percentage float64, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.config.Stage = stage
	c.config.TargetPercentage = percentage
	c.config.Enabled = true
	c.config.StartTime = time.Now()
	c.config.Duration = duration

	fmt.Printf("✅ 灰度已启用: %s (%.0f%% 流量)\n", stage, percentage*100)
}

// DisableCanary 禁用灰度
func (c *CanaryMetricsCollector) DisableCanary() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.config.Enabled = false
	fmt.Println("⚠️  灰度已禁用")
}

// RecordRequest 记录请求
func (c *CanaryMetricsCollector) RecordRequest(requestID string, useNewMethod bool, duration time.Duration, success bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	atomic.AddInt64(&c.metrics.TotalRequests, 1)

	if useNewMethod {
		atomic.AddInt64(&c.metrics.NewMethodRequests, 1)
	} else {
		atomic.AddInt64(&c.metrics.OldMethodRequests, 1)
	}

	if success {
		atomic.AddInt64(&c.metrics.SuccessfulRequests, 1)
	} else {
		atomic.AddInt64(&c.metrics.FailedRequests, 1)
	}

	// 更新延迟指标
	c.metrics.TotalLatency += duration
	if duration < c.metrics.MinLatency || c.metrics.MinLatency == 0 {
		c.metrics.MinLatency = duration
	}
	if duration > c.metrics.MaxLatency {
		c.metrics.MaxLatency = duration
	}
}

// RecordCacheHit 记录缓存命中
func (c *CanaryMetricsCollector) RecordCacheHit(hit bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if hit {
		atomic.AddInt64(&c.metrics.CacheHits, 1)
	} else {
		atomic.AddInt64(&c.metrics.CacheMisses, 1)
	}
}

// GetCurrentMetrics 获取当前指标
func (c *CanaryMetricsCollector) GetCurrentMetrics() *CanaryMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	metrics := *c.metrics
	metrics.CollectedAt = time.Now()

	// 计算平均延迟
	if metrics.TotalRequests > 0 {
		metrics.AvgLatency = metrics.TotalLatency / time.Duration(metrics.TotalRequests)
		metrics.ErrorRate = float64(metrics.FailedRequests) / float64(metrics.TotalRequests)
	}

	// 计算缓存命中率
	totalCacheOps := metrics.CacheHits + metrics.CacheMisses
	if totalCacheOps > 0 {
		metrics.CacheHitRate = float64(metrics.CacheHits) / float64(totalCacheOps)
	}

	return &metrics
}

// SnapshotMetrics 保存指标快照 (用于历史分析)
func (c *CanaryMetricsCollector) SnapshotMetrics() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 直接复制当前指标而不调用其他需要锁的函数
	snapshot := &CanaryMetrics{
		TotalRequests:      atomic.LoadInt64(&c.metrics.TotalRequests),
		NewMethodRequests:  atomic.LoadInt64(&c.metrics.NewMethodRequests),
		OldMethodRequests:  atomic.LoadInt64(&c.metrics.OldMethodRequests),
		SuccessfulRequests: atomic.LoadInt64(&c.metrics.SuccessfulRequests),
		FailedRequests:     atomic.LoadInt64(&c.metrics.FailedRequests),
		TotalLatency:       c.metrics.TotalLatency,
		MinLatency:         c.metrics.MinLatency,
		MaxLatency:         c.metrics.MaxLatency,
		CacheHits:          atomic.LoadInt64(&c.metrics.CacheHits),
		CacheMisses:        atomic.LoadInt64(&c.metrics.CacheMisses),
		CollectedAt:        time.Now(),
	}

	// 计算平均延迟
	if snapshot.TotalRequests > 0 {
		snapshot.AvgLatency = snapshot.TotalLatency / time.Duration(snapshot.TotalRequests)
		snapshot.ErrorRate = float64(snapshot.FailedRequests) / float64(snapshot.TotalRequests)
	}

	// 计算缓存命中率
	totalCacheOps := snapshot.CacheHits + snapshot.CacheMisses
	if totalCacheOps > 0 {
		snapshot.CacheHitRate = float64(snapshot.CacheHits) / float64(totalCacheOps)
	}

	c.history = append(c.history, snapshot)

	// 只保留最近 1000 个快照
	if len(c.history) > 1000 {
		c.history = c.history[1:]
	}
}

// GetMetricsHistory 获取指标历史
func (c *CanaryMetricsCollector) GetMetricsHistory(hours int) []*CanaryMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cutoff := time.Now().Add(-time.Duration(hours) * time.Hour)
	result := make([]*CanaryMetrics, 0)

	for _, m := range c.history {
		if m.CollectedAt.After(cutoff) {
			result = append(result, m)
		}
	}

	return result
}

// CanaryHealthCheck 灰度健康检查
type CanaryHealthCheck struct {
	collector  *CanaryMetricsCollector
	thresholds *CanaryThresholds
}

// CanaryThresholds 灰度阈值
type CanaryThresholds struct {
	// 告警阈值
	MaxErrorRate    float64
	MaxP99Latency   time.Duration
	MinCacheHitRate float64
	MaxMemoryGrowth int64 // bytes/minute

	// 回滚触发阈值
	CriticalErrorRate  float64
	CriticalP99Latency time.Duration
}

// DefaultCanaryThresholds 默认的灰度阈值
func DefaultCanaryThresholds() *CanaryThresholds {
	return &CanaryThresholds{
		// 告警阈值 (严格)
		MaxErrorRate:    0.01, // 1%
		MaxP99Latency:   150 * time.Millisecond,
		MinCacheHitRate: 0.5,    // 50%
		MaxMemoryGrowth: 100000, // 100KB

		// 回滚触发阈值 (严格)
		CriticalErrorRate:  0.03, // 3%
		CriticalP99Latency: 500 * time.Millisecond,
	}
}

// NewCanaryHealthCheck 创建灰度健康检查
func NewCanaryHealthCheck(collector *CanaryMetricsCollector) *CanaryHealthCheck {
	return &CanaryHealthCheck{
		collector:  collector,
		thresholds: DefaultCanaryThresholds(),
	}
}

// CheckHealth 检查灰度健康状况
func (c *CanaryHealthCheck) CheckHealth(ctx context.Context) (healthy bool, alerts []string, critique string) {
	metrics := c.collector.GetCurrentMetrics()

	if metrics.TotalRequests == 0 {
		return true, []string{}, "无足够的数据"
	}

	alerts = []string{}
	healthy = true

	// 检查错误率
	if metrics.ErrorRate > c.thresholds.CriticalErrorRate {
		critique = fmt.Sprintf("❌ 严重错误率: %.2f%% (> %.2f%%)",
			metrics.ErrorRate*100, c.thresholds.CriticalErrorRate*100)
		return false, alerts, critique
	}

	if metrics.ErrorRate > c.thresholds.MaxErrorRate {
		alerts = append(alerts, fmt.Sprintf("⚠️  高错误率: %.2f%%", metrics.ErrorRate*100))
		healthy = false
	}

	// 检查延迟
	if metrics.P99Latency > c.thresholds.CriticalP99Latency && metrics.P99Latency > 0 {
		critique = fmt.Sprintf("❌ 严重延迟升高: P99=%.0fms (> %.0fms)",
			metrics.P99Latency.Seconds()*1000, c.thresholds.CriticalP99Latency.Seconds()*1000)
		return false, alerts, critique
	}

	if metrics.P99Latency > c.thresholds.MaxP99Latency && metrics.P99Latency > 0 {
		alerts = append(alerts, fmt.Sprintf("⚠️  P99 延迟升高: %.0fms", metrics.P99Latency.Seconds()*1000))
		healthy = false
	}

	// 检查缓存命中率
	if metrics.CacheHitRate > 0 && metrics.CacheHitRate < c.thresholds.MinCacheHitRate {
		alerts = append(alerts, fmt.Sprintf("⚠️  缓存命中率过低: %.1f%%", metrics.CacheHitRate*100))
		healthy = false
	}

	if healthy {
		critique = "✅ 灰度健康"
	}

	return healthy, alerts, critique
}

// CanaryReport 灰度报告
type CanaryReport struct {
	Stage     CanaryStage
	Duration  time.Duration
	StartTime time.Time
	EndTime   time.Time

	// 流量分布
	TotalRequests     int64
	NewMethodRequests int64
	OldMethodRequests int64
	NewMethodPercent  float64

	// 成功率
	SuccessRate float64
	ErrorRate   float64

	// 延迟对比
	NewMethodAvgLatency time.Duration
	OldMethodAvgLatency time.Duration
	LatencyDifference   float64 // percentage

	// 缓存效果
	CacheHitRate float64

	// 结果一致性
	Consistency float64 // 0-1

	// 建议
	Recommendation string
}

// GenerateCanaryReport 生成灰度报告
func (c *CanaryMetricsCollector) GenerateCanaryReport(startTime time.Time) *CanaryReport {
	c.mu.RLock()
	defer c.mu.RUnlock()

	metrics := c.metrics
	duration := time.Since(startTime)

	report := &CanaryReport{
		Stage:             c.config.Stage,
		Duration:          duration,
		StartTime:         startTime,
		EndTime:           time.Now(),
		TotalRequests:     metrics.TotalRequests,
		NewMethodRequests: metrics.NewMethodRequests,
		OldMethodRequests: metrics.OldMethodRequests,
		SuccessRate:       float64(metrics.SuccessfulRequests) / float64(metrics.TotalRequests),
		ErrorRate:         metrics.ErrorRate,
		CacheHitRate:      metrics.CacheHitRate,
	}

	// 计算新方式的百分比
	if metrics.TotalRequests > 0 {
		report.NewMethodPercent = float64(metrics.NewMethodRequests) / float64(metrics.TotalRequests)
	}

	// 生成建议
	if report.ErrorRate > 0.03 {
		report.Recommendation = "❌ 立即回滚 (错误率过高)"
	} else if report.ErrorRate > 0.01 {
		report.Recommendation = "⚠️  建议延缓灰度 (错误率略高)"
	} else if report.CacheHitRate > 0 && report.CacheHitRate < 0.5 {
		report.Recommendation = "⚠️  建议检查缓存策略"
	} else if report.SuccessRate > 0.98 {
		report.Recommendation = "✅ 可以升级到下一阶段"
	} else {
		report.Recommendation = "✅ 灰度进行良好，继续监控"
	}

	return report
}

// PrintCanaryReport 打印灰度报告
func (report *CanaryReport) Print() {
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════╗")
	fmt.Printf("║  灰度报告: %s\n", report.Stage)
	fmt.Println("╚════════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Printf("📊 流量分布:\n")
	fmt.Printf("  总请求数: %d\n", report.TotalRequests)
	fmt.Printf("  新方式:   %d (%.1f%%)\n", report.NewMethodRequests, report.NewMethodPercent*100)
	fmt.Printf("  旧方式:   %d (%.1f%%)\n", report.OldMethodRequests, (1-report.NewMethodPercent)*100)
	fmt.Println()

	fmt.Printf("✅ 成功率: %.2f%%\n", report.SuccessRate*100)
	fmt.Printf("❌ 错误率: %.2f%%\n", report.ErrorRate*100)
	fmt.Println()

	if report.CacheHitRate > 0 {
		fmt.Printf("💾 缓存命中率: %.1f%%\n", report.CacheHitRate*100)
		fmt.Println()
	}

	fmt.Printf("⏱️  灰度时长: %v\n", report.Duration)
	fmt.Println()

	fmt.Printf("📋 建议:\n  %s\n", report.Recommendation)
	fmt.Println()
}

// ========================================================================
// 辅助函数
// ========================================================================

// hashRequestID 对 requestID 进行 hash (用于一致性流量分配)
func hashRequestID(requestID string) uint32 {
	hash := uint32(0)
	for _, ch := range requestID {
		hash = hash*31 + uint32(ch)
	}
	return hash
}

// WaitForCanaryCompletion 等待灰度完成
func WaitForCanaryCompletion(ctx context.Context, collector *CanaryMetricsCollector, duration time.Duration) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	deadline := time.Now().Add(duration)

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			metrics := collector.GetCurrentMetrics()
			fmt.Printf("[%s] 请求数: %d, 错误率: %.2f%%, 缓存命中: %.1f%%\n",
				time.Now().Format("15:04"), metrics.TotalRequests, metrics.ErrorRate*100, metrics.CacheHitRate*100)

			if time.Now().After(deadline) {
				return
			}

		case <-time.After(duration):
			return
		}
	}
}

// ========================================================================
// 灰度生命周期管理
// ========================================================================

// CanaryLifecycle 灰度生命周期管理器
type CanaryLifecycle struct {
	collector   *CanaryMetricsCollector
	healthCheck *CanaryHealthCheck
	stages      []CanaryStage
	currentIdx  int
}

// NewCanaryLifecycle 创建灰度生命周期管理器
func NewCanaryLifecycle() *CanaryLifecycle {
	config := &CanaryConfig{}
	collector := NewCanaryMetricsCollector(config)

	return &CanaryLifecycle{
		collector:   collector,
		healthCheck: NewCanaryHealthCheck(collector),
		stages: []CanaryStage{
			CanaryStage5Percent,
			CanaryStage25Percent,
			CanaryStage50Percent,
			CanaryStage100Percent,
		},
		currentIdx: 0,
	}
}

// StartStage 启动新的灰度阶段
func (c *CanaryLifecycle) StartStage(stage CanaryStage, duration time.Duration) {
	percentages := map[CanaryStage]float64{
		CanaryStage5Percent:   0.05,
		CanaryStage25Percent:  0.25,
		CanaryStage50Percent:  0.50,
		CanaryStage100Percent: 1.00,
	}

	percentage := percentages[stage]
	c.collector.EnableCanary(stage, percentage, duration)
	fmt.Printf("🚀 开始灰度: %s (%.0f%% 流量, 持续 %v)\n", stage, percentage*100, duration)
}

// EndStage 结束当前灰度阶段
func (c *CanaryLifecycle) EndStage() *CanaryReport {
	startTime := c.collector.config.StartTime
	report := c.collector.GenerateCanaryReport(startTime)
	c.collector.DisableCanary()
	return report
}

// CanUpgrade 检查是否可以升级到下一阶段
func (c *CanaryLifecycle) CanUpgrade() bool {
	healthy, alerts, _ := c.healthCheck.CheckHealth(context.Background())

	if !healthy && len(alerts) > 0 {
		fmt.Printf("⚠️  发现 %d 个警告，继续观察\n", len(alerts))
		return false
	}

	return healthy
}

// PrintLifecycleStatus 打印生命周期状态
func (c *CanaryLifecycle) PrintLifecycleStatus() {
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("灰度推出进度")
	fmt.Println("═══════════════════════════════════════════════════")

	for i, stage := range c.stages {
		status := "⏳"
		if i < c.currentIdx {
			status = "✅"
		} else if i == c.currentIdx {
			status = "🔄"
		}

		fmt.Printf("%s %s\n", status, stage)
	}

	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println()
}
